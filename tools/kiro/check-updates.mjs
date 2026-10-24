import { execFile } from 'node:child_process';
import { createHash, randomUUID } from 'node:crypto';
import { appendFile, mkdir, readFile, rename, rm, writeFile } from 'node:fs/promises';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const commitPattern = /^[a-f0-9]{40}$/;
const repositoryPattern = /^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/;

function requireValue(condition, message) {
  if (!condition) throw new Error(message);
}

export function validateManifest(manifest) {
  requireValue(manifest.schemaVersion === 1, '来源清单版本不支持');
  requireValue(Array.isArray(manifest.sources) && manifest.sources.length > 0, '来源清单不能为空');
  const identifiers = new Set();
  for (const source of manifest.sources) {
    requireValue(typeof source.id === 'string' && !identifiers.has(source.id), '来源标识缺失或重复');
    identifiers.add(source.id);
    requireValue(repositoryPattern.test(source.repository), '仓库名格式错误');
    requireValue(typeof source.branch === 'string' && source.branch.length > 0, '默认分支缺失');
    requireValue(commitPattern.test(source.observedCommit), '观察提交必须是完整 SHA');
    requireValue(source.reviewedCommit === null || commitPattern.test(source.reviewedCommit), '审阅提交格式错误');
    requireValue(source.portedCommit === null || commitPattern.test(source.portedCommit), '移植提交格式错误');
    requireValue(Number.isFinite(Date.parse(source.observedAt)), '观察时间格式错误');
    requireValue(Array.isArray(source.watchPaths) && source.watchPaths.length > 0, '关注路径不能为空');
    requireValue(source.watchPaths.every(path => typeof path === 'string' && path.length > 0), '关注路径格式错误');
    requireValue(Array.isArray(source.observedReleases), '发布基线缺失');
  }
  return manifest;
}

export function githubJson(path, { token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN } = {}) {
  requireValue(path.startsWith('/repos/'), '只允许访问 GitHub 仓库 API');
  requireValue(!token || !/[\r\n\x00]/.test(token), 'GitHub Token 格式错误');
  const escapeConfig = value => value.replaceAll('\\', '\\\\').replaceAll('"', '\\"');
  const config = token ? `header = "Authorization: Bearer ${escapeConfig(token)}"\n` : '';
  return new Promise((resolveRequest, rejectRequest) => {
    const child = execFile(process.platform === 'win32' ? 'curl.exe' : 'curl', [
      '--silent', '--show-error', '--location', '--max-redirs', '3',
      '--proto', '=https', '--proto-redir', '=https',
      '--connect-timeout', '15', '--max-time', '45',
      '--header', 'Accept: application/vnd.github+json',
      '--header', 'X-GitHub-Api-Version: 2022-11-28',
      '--header', 'User-Agent: sub2apicust-kiro-reference-watch',
      '--config', '-', '--write-out', '\n%{http_code}',
      `https://api.github.com${path}`,
    ], { encoding: 'utf8', windowsHide: true, timeout: 50_000, maxBuffer: 32 * 1024 * 1024 }, (error, stdout) => {
      if (error) {
        rejectRequest(new Error(`GitHub 网络请求失败（curl ${error.code ?? 'timeout'}）；请检查网络或代理`));
        return;
      }
      const separator = stdout.lastIndexOf('\n');
      const status = Number(stdout.slice(separator + 1));
      if (status !== 200) {
        rejectRequest(new Error(`GitHub HTTP ${status}：${path}；403/429 可能为限流，404 可能为删除或权限变化`));
        return;
      }
      try {
        resolveRequest(JSON.parse(stdout.slice(0, separator)));
      } catch {
        rejectRequest(new Error(`GitHub 返回非 JSON 响应：${path}`));
      }
    });
    child.stdin.on('error', () => {});
    child.stdin.end(config);
  });
}

export function classifyFiles(files, watchPaths, incomplete = false) {
  if (incomplete) return 'manual-review';
  const matches = path => watchPaths.some(prefix => prefix.endsWith('/') ? path.startsWith(prefix) : path === prefix);
  if (files.some(file => matches(file.filename) || (file.previous_filename && matches(file.previous_filename)))) return 'code';
  if (files.length === 0) return 'manual-review';
  const isDocumentation = path => /(^|\/)(docs?|screenshots?|images?)\//i.test(path)
    || /\.(md|rst|txt|png|jpe?g|gif|webp|svg)$/i.test(path);
  return files.every(file => isDocumentation(file.filename) && (!file.previous_filename || isDocumentation(file.previous_filename)))
    ? 'documentation' : 'manual-review';
}

export async function compareCommits(repository, base, head, get, { maxPages = 10 } = {}) {
  const url = `https://github.com/${repository}/compare/${base}...${head}`;
  if (base === head) return { status: 'identical', commits: [], files: [], totalCommits: 0, incomplete: false, url };
  const commits = new Map();
  let comparison;
  let expectedCommits;
  for (let page = 1; page <= maxPages; page += 1) {
    const data = await get(`/repos/${repository}/compare/${base}...${head}?per_page=100&page=${page}`);
    requireValue(['ahead', 'behind', 'diverged', 'identical'].includes(data.status), 'GitHub compare 状态缺失');
    requireValue(Number.isInteger(data.total_commits) && data.total_commits >= 0, 'GitHub compare 提交总数缺失');
    requireValue(Array.isArray(data.commits), 'GitHub compare 提交列表缺失');
    if (!comparison) {
      requireValue(Array.isArray(data.files), 'GitHub compare 文件列表缺失');
      comparison = { status: data.status, files: data.files, url };
      expectedCommits = data.total_commits;
    }
    requireValue(data.status === comparison.status && data.total_commits === expectedCommits, '比较结果在分页期间发生变化');
    for (const commit of data.commits) {
      requireValue(commitPattern.test(commit.sha), 'GitHub compare 提交 SHA 格式错误');
      commits.set(commit.sha, {
        sha: commit.sha,
        subject: String(commit.commit?.message ?? '').split('\n')[0],
        date: commit.commit?.committer?.date ?? null,
      });
    }
    if (commits.size >= expectedCommits || data.commits.length === 0) break;
  }
  const files = comparison.files.map(file => {
    requireValue(typeof file.filename === 'string', 'GitHub compare 文件名缺失');
    return { filename: file.filename, previous_filename: file.previous_filename, status: file.status };
  });
  return {
    ...comparison, files, commits: [...commits.values()], totalCommits: expectedCommits,
    incomplete: commits.size < expectedCommits || files.length >= 300,
  };
}

async function readReleases(repository, get, maxPages) {
  const releases = new Map();
  for (let page = 1; page <= maxPages; page += 1) {
    const batch = await get(`/repos/${repository}/releases?per_page=100&page=${page}`);
    requireValue(Array.isArray(batch), 'GitHub Release 列表缺失');
    for (const release of batch.filter(item => !item.draft)) {
      requireValue(Number.isSafeInteger(release.id) && typeof release.tag_name === 'string', 'GitHub Release 标识缺失');
      requireValue(Number.isFinite(Date.parse(release.published_at)), 'GitHub Release 发布时间缺失');
      requireValue(Number.isFinite(Date.parse(release.updated_at)), 'GitHub Release 更新时间缺失');
      releases.set(release.id, {
        id: release.id, tag: release.tag_name, publishedAt: release.published_at,
        updatedAt: release.updated_at, prerelease: Boolean(release.prerelease),
        bodySha256: createHash('sha256').update(release.body ?? '').digest('hex'),
      });
    }
    if (batch.length < 100) return { releases: [...releases.values()], incomplete: false };
  }
  return { releases: [...releases.values()], incomplete: true };
}

export async function inspectSource(source, previous, get, { checkedAt, maxPages = 10 } = {}) {
  const metadata = await get(`/repos/${source.repository}`);
  requireValue(repositoryPattern.test(metadata.full_name), 'GitHub 规范仓库名缺失');
  requireValue(typeof metadata.default_branch === 'string' && metadata.default_branch.length > 0, 'GitHub 默认分支缺失');
  requireValue(typeof metadata.archived === 'boolean' && typeof metadata.disabled === 'boolean', 'GitHub 仓库状态缺失');
  const repository = metadata.full_name;
  const branch = metadata.default_branch;
  const head = await get(`/repos/${repository}/commits/${encodeURIComponent(branch)}`);
  requireValue(commitPattern.test(head.sha), 'GitHub 默认分支 HEAD 格式错误');
  const prior = previous?.requestedRepository === source.repository ? previous : null;
  if (prior) {
    requireValue(commitPattern.test(prior.head) && Number.isFinite(Date.parse(prior.checkedAt))
      && Array.isArray(prior.releases), '观察游标损坏；请保留原文件并人工核查基线');
  }
  const base = prior?.head ?? source.observedCommit;
  const comparison = await compareCommits(repository, base, head.sha, get, { maxPages });
  const relevance = comparison.status === 'identical' ? 'none' : classifyFiles(comparison.files, source.watchPaths, comparison.incomplete);
  const releaseResult = await readReleases(repository, get, maxPages);
  const baselineReleases = new Map((prior?.releases ?? source.observedReleases).map(release => [release.id, release]));
  const cutoff = Date.parse(prior?.checkedAt ?? source.observedAt);
  const releaseChanges = releaseResult.releases.flatMap(release => {
    const known = baselineReleases.get(release.id);
    const changed = known && (known.updatedAt !== release.updatedAt || (known.bodySha256 && known.bodySha256 !== release.bodySha256));
    const added = !known && (Date.parse(release.publishedAt) > cutoff || Date.parse(release.updatedAt) > cutoff);
    return changed || added ? [{ ...release, change: known ? 'edited' : 'new' }] : [];
  });
  const metadataChanges = [];
  if ((prior?.repository ?? source.repository).toLowerCase() !== repository.toLowerCase()) metadataChanges.push('仓库迁移或重命名');
  if ((prior?.branch ?? source.branch) !== branch) metadataChanges.push('默认分支变化');
  if ((prior?.archived ?? false) !== metadata.archived) metadataChanges.push(metadata.archived ? '仓库已归档' : '仓库恢复维护');
  if ((prior?.disabled ?? false) !== metadata.disabled) metadataChanges.push(metadata.disabled ? '仓库已禁用' : '仓库恢复可用');
  const historyChanged = comparison.status !== 'identical';
  const actionable = (historyChanged && (comparison.status !== 'ahead' || relevance !== 'documentation'))
    || metadataChanges.length > 0 || releaseChanges.length > 0 || releaseResult.incomplete;
  return {
    result: {
      id: source.id, repository, branch, head: head.sha, base,
      reviewedCommit: source.reviewedCommit, portedCommit: source.portedCommit,
      unreviewed: source.reviewedCommit !== head.sha,
      reviewUrl: source.reviewedCommit ? `https://github.com/${repository}/compare/${source.reviewedCommit}...${head.sha}` : null,
      comparison, relevance, metadataChanges, releaseChanges,
      releasesIncomplete: releaseResult.incomplete, actionable, error: null,
    },
    state: {
      requestedRepository: source.repository, repository, branch, head: head.sha,
      archived: metadata.archived, disabled: metadata.disabled,
      checkedAt, releases: releaseResult.releases,
    },
  };
}

export async function checkSources(manifest, previousState, get, { checkedAt = new Date().toISOString(), maxPages = 10 } = {}) {
  validateManifest(manifest);
  requireValue(previousState?.schemaVersion === 1 && previousState.sources !== null
    && typeof previousState.sources === 'object' && !Array.isArray(previousState.sources), '巡检状态格式错误');
  const nextState = { schemaVersion: 1, sources: { ...previousState.sources } };
  const results = [];
  for (const source of manifest.sources) {
    try {
      const inspected = await inspectSource(source, previousState.sources[source.id], get, { checkedAt, maxPages });
      results.push(inspected.result);
      nextState.sources[source.id] = inspected.state;
    } catch (error) {
      results.push({ id: source.id, repository: source.repository, error: error instanceof Error ? error.message : '检查出现未知错误', actionable: true });
    }
  }
  return {
    report: {
      schemaVersion: 1, checkedAt,
      failed: results.filter(result => result.error).length,
      actionable: results.filter(result => result.actionable).length,
      results,
    },
    nextState,
  };
}

function markdownText(value) {
  return String(value).replace(/[\r\n]/g, ' ').replace(/[\\`*_[\]<>|]/g, '\\$&');
}

export function renderReport(report) {
  const lines = [
    '# Kiro 参考项目更新巡检', '', `检查时间：${report.checkedAt}；需处理 ${report.actionable} 项；失败 ${report.failed} 项。`,
    '', '审阅和移植基线只从 sources.lock.json 读取，巡检不会推进它们。观察到更新不代表已采用修复。', '',
  ];
  for (const result of report.results) {
    lines.push(`## ${markdownText(result.repository)}`, '');
    if (result.error) {
      lines.push(`**检查失败**：${markdownText(result.error)}。本仓观察游标保留，不能解释为没有更新。`, '');
      continue;
    }
    lines.push(
      `- 默认分支：${markdownText(result.branch)}；HEAD：\`${result.head}\`。`,
      `- 已审阅：${result.reviewedCommit ? `\`${result.reviewedCommit}\`` : '尚未审阅'}；已移植：${result.portedCommit ? `\`${result.portedCommit}\`` : '尚未移植'}。`,
      `- 本次变化：${result.comparison.status}；分类：${result.relevance}；待审阅：${result.unreviewed ? '是' : '否'}。`,
      `- [本次差异](${result.comparison.url})${result.reviewUrl ? `；[相对审阅基线的全部差异](${result.reviewUrl})` : ''}。`,
    );
    for (const change of result.metadataChanges) lines.push(`- **${change}**。`);
    if (result.comparison.incomplete) lines.push('- **提交或文件列表不完整，必须人工核查完整差异；不能判定为只有文档变化。**');
    if (result.releasesIncomplete) lines.push('- **Release 分页达到上限，发布列表不完整，须人工核查。**');
    for (const release of result.releaseChanges) {
      lines.push(`- Release ${release.change}：[${markdownText(release.tag)}](https://github.com/${result.repository}/releases/tag/${encodeURIComponent(release.tag)})${release.prerelease ? '（预发布）' : ''}。`);
    }
    for (const commit of result.comparison.commits.slice(0, 30)) {
      lines.push(`- [${commit.sha.slice(0, 12)}](https://github.com/${result.repository}/commit/${commit.sha}) ${markdownText(commit.subject)}`);
    }
    if (result.comparison.commits.length > 30) lines.push(`- 报告正文仅显示前 30 条，完整采集 ${result.comparison.commits.length}/${result.comparison.totalCommits} 条见同名 JSON。`);
    const files = result.comparison.files.slice(0, 40).map(file => `\`${markdownText(file.filename)}\``);
    if (files.length) lines.push('', `变化文件（最多显示 40 个）：${files.join('、')}。`);
    lines.push('');
  }
  return `${lines.join('\n')}\n`;
}

async function readJson(path, fallback) {
  try {
    return JSON.parse(await readFile(path, 'utf8'));
  } catch (error) {
    if (error.code === 'ENOENT' && fallback) return fallback;
    throw new Error(`无法读取 JSON：${path}；${error.code ?? '格式错误'}`);
  }
}

async function writeAtomic(path, contents) {
  const temporary = `${path}.${randomUUID()}.tmp`;
  try {
    await writeFile(temporary, contents, 'utf8');
    await rename(temporary, path);
  } finally {
    await rm(temporary, { force: true });
  }
}

export function parseArguments(args) {
  const options = { out: resolve(scriptDirectory, '../../output/kiro-reference-watch'), manifest: join(scriptDirectory, 'sources.lock.json'), failOnUpdate: false };
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === '--fail-on-update') options.failOnUpdate = true;
    else if (argument === '--out' || argument === '--manifest') {
      requireValue(args[index + 1] && !args[index + 1].startsWith('--'), `${argument} 缺少路径`);
      options[argument === '--out' ? 'out' : 'manifest'] = resolve(args[++index]);
    } else throw new Error(`未知参数：${argument}`);
  }
  return options;
}

export async function main(args = process.argv.slice(2), get = githubJson) {
  const options = parseArguments(args);
  const manifest = await readJson(options.manifest);
  const statePath = join(options.out, 'state.json');
  const previousState = await readJson(statePath, { schemaVersion: 1, sources: {} });
  const { report, nextState } = await checkSources(manifest, previousState, get);
  const markdown = renderReport(report);
  const reportJson = `${JSON.stringify(report, null, 2)}\n`;
  await mkdir(options.out, { recursive: true });
  const timestamp = report.checkedAt.replaceAll(/[:.]/g, '-');
  await writeAtomic(join(options.out, `${timestamp}.json`), reportJson);
  await writeAtomic(join(options.out, `${timestamp}.md`), markdown);
  await writeAtomic(join(options.out, 'latest.json'), reportJson);
  await writeAtomic(join(options.out, 'latest.md'), markdown);
  if (process.env.GITHUB_STEP_SUMMARY) await appendFile(process.env.GITHUB_STEP_SUMMARY, markdown, 'utf8');
  await writeAtomic(statePath, `${JSON.stringify(nextState, null, 2)}\n`);
  console.log(`已检查 ${report.results.length} 个来源；需处理 ${report.actionable} 项；失败 ${report.failed} 项。报告：${join(options.out, 'latest.md')}`);
  return report.failed ? 1 : options.failOnUpdate && report.actionable ? 2 : 0;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().then(code => { process.exitCode = code; }).catch(error => {
    console.error(error.message);
    process.exitCode = 1;
  });
}
