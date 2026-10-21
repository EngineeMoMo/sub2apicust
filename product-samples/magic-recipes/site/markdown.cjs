(function (root, factory) {
  const api = factory();
  if (typeof module === 'object' && module.exports) module.exports = api;
  else root.MofaMarkdown = api;
})(globalThis, () => {
  'use strict';
  // 只把模型回答解析成固定的几种节点，再用 createElement/textContent 渲染；
  // 不接受原始 HTML，不加载远程图片，链接仅允许 http(s)/mailto。
  const MAX_SOURCE = 200000, MAX_DEPTH = 6, MAX_LINE = 4000, MAX_COLUMNS = 30, MAX_ROWS = 500;
  const FENCE = /^( {0,3})(`{3,}|~{3,})\s*([^`\s]*)[^`]*$/;
  const HEADING = /^ {0,3}(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$/;
  const RULE = /^ {0,3}([-*_])(?:[ \t]*\1){2,}[ \t]*$/;
  const QUOTE = /^ {0,3}> ?/;
  const BULLET = /^( {0,3})([-*+])[ \t]+(.*)$/;
  const ORDERED = /^( {0,3})(\d{1,9})([.)])[ \t]+(.*)$/;
  const DELIMITER = /^ {0,3}\|?[ \t]*:?-+:?[ \t]*(\|[ \t]*:?-+:?[ \t]*)*\|?[ \t]*$/;
  const ESCAPABLE = '\\`*_{}[]()#+-.!|>~';
  const INLINE = {
    code: /(`+)([^`\n]+?)\1(?!`)/y,
    strong: /(?:\*\*([^\n]+?)\*\*|__([^\n]+?)__)/y,
    strike: /~~([^~\n]+?)~~/y,
    em: /(?:\*([^\s*][^*\n]*?)\*|_([^\s_][^_\n]*?)_(?![A-Za-z0-9]))/y,
    link: /(!?)\[([^\]\n]{0,500})\]\(([^\s()]{1,2000})\)/y,
    auto: /https?:\/\/[A-Za-z0-9\-._~:/?#@!$&'*+,;=%]+/y
  };
  function at(pattern, text, index) {
    pattern.lastIndex = index;
    return pattern.exec(text);
  }
  function safeHref(value) {
    try {
      const url = new URL(value);
      if (!['http:', 'https:', 'mailto:'].includes(url.protocol) || url.username || url.password) return '';
      return url.href;
    } catch { return ''; }
  }
  function inline(text, insideLink = false) {
    const nodes = [];
    let buffer = '';
    const flush = () => { if (buffer) { nodes.push({ type: 'text', text: buffer }); buffer = ''; } };
    if (text.length > MAX_LINE) return [{ type: 'text', text }];
    let index = 0;
    while (index < text.length) {
      const char = text[index];
      let match;
      if (char === '\\' && ESCAPABLE.includes(text[index + 1] || '')) { buffer += text[index + 1]; index += 2; continue; }
      if (char === '`' && (match = at(INLINE.code, text, index))) {
        flush(); nodes.push({ type: 'code', text: match[2].trim() || match[2] }); index += match[0].length; continue;
      }
      if (!insideLink && (char === '[' || (char === '!' && text[index + 1] === '[')) && (match = at(INLINE.link, text, index))) {
        const href = safeHref(match[3]);
        flush();
        if (!href) nodes.push({ type: 'text', text: match[0] });
        else if (match[1]) nodes.push({ type: 'link', href, image: true, children: [{ type: 'text', text: match[2] || href }] });
        else nodes.push({ type: 'link', href, children: inline(match[2] || href, true) });
        index += match[0].length; continue;
      }
      if ((char === '*' || char === '_') && text[index + 1] === char && (match = at(INLINE.strong, text, index))) {
        flush(); nodes.push({ type: 'strong', children: inline(match[1] ?? match[2], insideLink) }); index += match[0].length; continue;
      }
      if (char === '~' && text[index + 1] === '~' && (match = at(INLINE.strike, text, index))) {
        flush(); nodes.push({ type: 'del', children: inline(match[1], insideLink) }); index += match[0].length; continue;
      }
      if ((char === '*' || (char === '_' && !/[A-Za-z0-9]/.test(text[index - 1] || ''))) && (match = at(INLINE.em, text, index))) {
        flush(); nodes.push({ type: 'em', children: inline(match[1] ?? match[2], insideLink) }); index += match[0].length; continue;
      }
      if (!insideLink && char === 'h' && (match = at(INLINE.auto, text, index))) {
        let url = match[0];
        while (/[.,;:!?'")\]]$/.test(url)) url = url.slice(0, -1);
        const href = url.length > 'https://'.length ? safeHref(url) : '';
        if (href) { flush(); nodes.push({ type: 'link', href, children: [{ type: 'text', text: url }] }); index += url.length; continue; }
      }
      buffer += char; index++;
    }
    flush();
    return nodes;
  }
  function paragraphInline(lines) {
    const nodes = [];
    lines.forEach((line, index) => {
      if (index) nodes.push({ type: 'br' });
      nodes.push(...inline(line.trim()));
    });
    return nodes;
  }
  function splitRow(line) {
    let row = line.trim();
    if (row.startsWith('|')) row = row.slice(1);
    if (row.endsWith('|') && !row.endsWith('\\|')) row = row.slice(0, -1);
    const cells = [];
    let cell = '';
    for (let index = 0; index < row.length; index++) {
      if (row[index] === '\\' && row[index + 1] === '|') { cell += '|'; index++; }
      else if (row[index] === '|') { cells.push(cell.trim()); cell = ''; }
      else cell += row[index];
    }
    cells.push(cell.trim());
    return cells.slice(0, MAX_COLUMNS);
  }
  function tableStart(lines, index) {
    const header = lines[index], delimiter = lines[index + 1];
    if (!header || !delimiter || !header.includes('|') || !DELIMITER.test(delimiter)) return false;
    const width = splitRow(header).length;
    return width > 0 && width === splitRow(delimiter).length;
  }
  function interrupts(lines, index) {
    const line = lines[index];
    return !line.trim() || FENCE.test(line) || HEADING.test(line) || RULE.test(line) || QUOTE.test(line)
      || BULLET.test(line) || ORDERED.test(line) || tableStart(lines, index);
  }
  function blocks(lines, depth) {
    const result = [];
    let index = 0;
    while (index < lines.length) {
      const line = lines[index];
      if (!line.trim()) { index++; continue; }
      if (depth >= MAX_DEPTH) {
        const rest = [];
        while (index < lines.length && lines[index].trim()) rest.push(lines[index++]);
        result.push({ type: 'paragraph', children: [{ type: 'text', text: rest.join('\n') }] });
        continue;
      }
      let match = FENCE.exec(line);
      if (match) {
        const indent = match[1].length, marker = match[2], body = [];
        index++;
        while (index < lines.length) {
          const closing = lines[index].match(/^ {0,3}(`{3,}|~{3,})[ \t]*$/);
          if (closing && closing[1][0] === marker[0] && closing[1].length >= marker.length) { index++; break; }
          body.push(lines[index].replace(new RegExp('^ {0,' + indent + '}'), ''));
          index++;
        }
        result.push({ type: 'code', language: match[3].slice(0, 30), text: body.join('\n') });
        continue;
      }
      if ((match = HEADING.exec(line))) { result.push({ type: 'heading', level: match[1].length, children: inline(match[2]) }); index++; continue; }
      if (RULE.test(line)) { result.push({ type: 'rule' }); index++; continue; }
      if (QUOTE.test(line)) {
        const quoted = [];
        while (index < lines.length && QUOTE.test(lines[index])) quoted.push(lines[index++].replace(QUOTE, ''));
        result.push({ type: 'quote', children: blocks(quoted, depth + 1) });
        continue;
      }
      if (tableStart(lines, index)) {
        const header = splitRow(lines[index]);
        const align = splitRow(lines[index + 1]).map(cell => cell.startsWith(':') && cell.endsWith(':') ? 'center' : cell.endsWith(':') ? 'right' : cell.startsWith(':') ? 'left' : '');
        const rows = [];
        index += 2;
        while (index < lines.length && lines[index].trim() && lines[index].includes('|') && rows.length < MAX_ROWS) {
          const cells = splitRow(lines[index++]);
          rows.push(header.map((_cell, column) => inline(cells[column] || '')));
        }
        result.push({ type: 'table', align, header: header.map(cell => inline(cell)), rows });
        continue;
      }
      const bullet = BULLET.exec(line), ordered = bullet ? null : ORDERED.exec(line);
      if (bullet || ordered) {
        const isOrdered = Boolean(ordered), list = { type: 'list', ordered: isOrdered, start: isOrdered ? Number(ordered[2]) : 1, loose: false, items: [] };
        while (index < lines.length) {
          const item = isOrdered ? ORDERED.exec(lines[index]) : BULLET.exec(lines[index]);
          if (!item) break;
          const lead = item[1].length, width = lines[index].length - item[isOrdered ? 4 : 3].length;
          const content = [item[isOrdered ? 4 : 3]];
          index++;
          while (index < lines.length) {
            const next = lines[index];
            if (!next.trim()) {
              let ahead = index + 1;
              while (ahead < lines.length && !lines[ahead].trim()) ahead++;
              if (ahead < lines.length && lines[ahead].length - lines[ahead].trimStart().length > lead) {
                list.loose = true;
                for (; index < ahead; index++) content.push('');
                continue;
              }
              break;
            }
            const spaces = next.length - next.trimStart().length;
            if (spaces > lead) { content.push(next.slice(Math.min(spaces, width))); index++; continue; }
            if (!interrupts(lines, index)) { content.push(next); index++; continue; }
            break;
          }
          let task;
          const checkbox = /^\[([ xX])\][ \t]+/.exec(content[0]);
          if (checkbox) { task = checkbox[1] !== ' '; content[0] = content[0].slice(checkbox[0].length); }
          list.items.push({ task, children: blocks(content, depth + 1) });
          while (index < lines.length && !lines[index].trim()) {
            let ahead = index + 1;
            while (ahead < lines.length && !lines[ahead].trim()) ahead++;
            const following = lines[ahead] && (isOrdered ? ORDERED.exec(lines[ahead]) : BULLET.exec(lines[ahead]));
            if (!following) break;
            list.loose = true;
            index = ahead;
          }
        }
        result.push(list);
        continue;
      }
      const paragraph = [line];
      index++;
      while (index < lines.length && !interrupts(lines, index)) paragraph.push(lines[index++]);
      result.push({ type: 'paragraph', children: paragraphInline(paragraph) });
    }
    return result;
  }
  function parse(source) {
    const text = String(source ?? '').slice(0, MAX_SOURCE).replace(/\r\n?/g, '\n');
    return blocks(text.split('\n'), 0);
  }
  function render(source, document, options = {}) {
    const fragment = document.createDocumentFragment();
    const element = (tag, className) => { const node = document.createElement(tag); if (className) node.className = className; return node; };
    function inlineNodes(parent, nodes) {
      for (const node of nodes) {
        if (node.type === 'text') parent.append(document.createTextNode(node.text));
        else if (node.type === 'br') parent.append(element('br'));
        else if (node.type === 'code') { const code = element('code'); code.textContent = node.text; parent.append(code); }
        else if (node.type === 'link') {
          const link = element('a', node.image ? 'md-image-link' : '');
          link.href = node.href; link.target = '_blank'; link.rel = 'noopener noreferrer'; link.referrerPolicy = 'no-referrer'; link.title = node.href;
          if (node.image) link.append(document.createTextNode('图片：'));
          inlineNodes(link, node.children);
          parent.append(link);
        } else {
          const tag = { strong: 'strong', em: 'em', del: 'del' }[node.type];
          const wrapper = element(tag);
          inlineNodes(wrapper, node.children);
          parent.append(wrapper);
        }
      }
    }
    function blockNodes(parent, nodes, tight = false) {
      for (const node of nodes) {
        if (node.type === 'paragraph') {
          if (tight) inlineNodes(parent, node.children);
          else { const paragraph = element('p'); inlineNodes(paragraph, node.children); parent.append(paragraph); }
        } else if (node.type === 'heading') {
          const heading = element(node.level <= 2 ? 'h5' : 'h6', 'md-h' + Math.min(node.level, 4));
          inlineNodes(heading, node.children); parent.append(heading);
        } else if (node.type === 'code') {
          const box = element('div', 'md-code'), head = element('div', 'md-code-head'), label = element('span');
          label.textContent = node.language && node.language !== 'text' ? node.language : '文本';
          head.append(label);
          if (options.onCopy) {
            const copy = element('button', 'md-code-copy');
            copy.type = 'button'; copy.textContent = '复制';
            copy.setAttribute('aria-label', '复制这段' + (node.language || '文本'));
            copy.addEventListener('click', () => options.onCopy(node.text, copy));
            head.append(copy);
          }
          const pre = element('pre'), code = element('code');
          code.textContent = node.text; pre.append(code); box.append(head, pre); parent.append(box);
        } else if (node.type === 'quote') {
          const quote = element('blockquote'); blockNodes(quote, node.children); parent.append(quote);
        } else if (node.type === 'rule') parent.append(element('hr'));
        else if (node.type === 'table') {
          const wrap = element('div', 'md-table'), table = element('table'), head = element('thead'), body = element('tbody'), headRow = element('tr');
          wrap.tabIndex = 0; wrap.setAttribute('role', 'region'); wrap.setAttribute('aria-label', '表格，可横向滚动');
          node.header.forEach((cell, column) => {
            const th = element('th', node.align[column] ? 'align-' + node.align[column] : '');
            inlineNodes(th, cell); headRow.append(th);
          });
          head.append(headRow);
          for (const row of node.rows) {
            const tr = element('tr');
            row.forEach((cell, column) => {
              const td = element('td', node.align[column] ? 'align-' + node.align[column] : '');
              inlineNodes(td, cell); tr.append(td);
            });
            body.append(tr);
          }
          table.append(head, body); wrap.append(table); parent.append(wrap);
        } else if (node.type === 'list') {
          const list = element(node.ordered ? 'ol' : 'ul');
          if (node.ordered && node.start !== 1) list.start = node.start;
          for (const item of node.items) {
            const li = element('li', item.task === undefined ? '' : 'md-task');
            if (item.task !== undefined) {
              const box = element('span', 'md-check' + (item.task ? ' done' : ''));
              box.setAttribute('aria-hidden', 'true');
              const state = element('span', 'sr-only');
              state.textContent = item.task ? '已完成：' : '未完成：';
              li.append(box, state);
            }
            blockNodes(li, item.children, !node.loose);
            list.append(li);
          }
          parent.append(list);
        }
      }
    }
    blockNodes(fragment, parse(source));
    return fragment;
  }
  return { parse, inline, render, safeHref };
});
