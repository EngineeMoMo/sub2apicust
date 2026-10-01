import test from 'node:test';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { mkdir, mkdtemp, readFile, readdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { buildFamilyProducts } from './build-products.mjs';

const workspace = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../../..');
const evidence = path.join(workspace, 'output/family-integrated-20261001/bundle-tests');
const require = createRequire(import.meta.url);
const { JSDOM } = require('jsdom');
await mkdir(evidence, { recursive: true });

test('打包仅发布运行资源，配方脚本和样式外置，工坊原图 WebP 完整保留', async () => {
  const output = await mkdtemp(path.join(evidence, 'pack-'));
  const products = await buildFamilyProducts(output);
  assert.deepEqual((await readdir(products.recipes)).sort(), ['.mofa-generated', 'bundle.js', 'index.html', 'theme.css']);
  const html = await readFile(path.join(products.recipes, 'index.html'), 'utf8');
  assert.ok(html.includes('name="mofa-api-site" content="same-origin"'));
  assert.ok(html.includes('src="./bundle.js"'));
  assert.ok(!html.includes('<script>') && !html.includes('<style>') && !html.includes('unsafe-inline'));
  const assets = await readdir(path.join(products.studio, 'assets'));
  assert.ok(assets.includes('jade-bottle.webp') && assets.includes('paper-fox.webp'));
  assert.ok(assets.every(file => /\.(webp|mp4)$/.test(file) || file === 'mofa-mark-flat.png'));
  assert.equal(assets.filter(file => file.endsWith('.mp4')).length, 0);
  assert.ok(!assets.some(file => file.startsWith('film-')));
  assert.ok(assets.includes('festival-friends.webp') && assets.includes('lilac-cosplay-thumb.webp'));
  for (const id of ["cherry-street","vinyl-afternoon","scarlet-rider","lotus-sword","tide-mecha","otter-morning","aurora-lagoon","lantern-alley"]) {
    for (const suffix of ['.webp', '-thumb.webp']) {
      assert.deepEqual(await readFile(path.join(products.studio, 'assets', id + suffix)), await readFile(path.join(workspace, 'product-samples/magic-studio/assets', id + suffix)));
    }
  }
  assert.ok(!assets.includes('PROVENANCE-EXPANDED-20261001.json'));
  for (const id of ['midnight-editorial','coral-sneaker']) for (const suffix of ['.webp','-thumb.webp']) assert.ok(!assets.includes(id+suffix));
  for (const id of ['brass-cartographer','frost-musician','red-panda-moss','turtle-blue','jazz-cutout','blue-botanical','anime-rain-tram','anime-sky-mechanic','sunken-lounge','lilac-bookshop','window-portrait']) for (const suffix of ['.webp','-thumb.webp']) assert.deepEqual(await readFile(path.join(products.studio,'assets',id+suffix)),await readFile(path.join(workspace,'product-samples/magic-studio/assets',id+suffix)));
  assert.ok(!assets.includes('PROVENANCE-CURATED-20261001.json'));
  assert.ok((await readdir(products.studio)).includes('host-client.js'));
  assert.ok(!assets.includes('PROVENANCE.json'));
  const studio = await readFile(path.join(products.studio, 'index.html'), 'utf8');
  assert.ok(studio.includes('href="/family"'));
  assert.ok(studio.includes('href="/tools/studio/submit"'));
  assert.ok(studio.includes('id="style-filters"'));
  for(const id of ['pixel-night','ink-koi','comic-hero','watercolor-island','papercut-tiger','retro-observatory']) assert.ok(assets.includes(id+'.webp')&&assets.includes(id+'-thumb.webp'));
  assert.ok(!assets.includes('PROVENANCE-STYLES-20261001.json'));
  const source = await readFile(path.join(workspace, 'product-samples/magic-studio/assets/jade-bottle.webp'));
  assert.deepEqual(await readFile(path.join(products.studio, 'assets/jade-bottle.webp')), source);
});

test('内置配方使用当前来源选择配置，不受 api_site 参数影响或隐式调用模型', async () => {
  const output = await mkdtemp(path.join(evidence, 'connect-'));
  const { recipes } = await buildFamilyProducts(output);
  const html = await readFile(path.join(recipes, 'index.html'), 'utf8');
  const script = await readFile(path.join(recipes, 'bundle.js'), 'utf8');
  const requests = [], opened = [];
  const dom = new JSDOM(html, { url: 'https://family.example.test/recipes/?api_site=https://evil.test', runScripts: 'outside-only' });
  const window = dom.window;
  window.matchMedia = () => ({ matches: false });
  window.HTMLElement.prototype.scrollIntoView = () => {};
  window.fetch = (...args) => { requests.push(args); throw new Error('禁止隐式调用'); };
  window.open = url => { opened.push(url); return null; };
  try {
    window.eval(script);
    const site = window.document.getElementById('magic-site');
    assert.equal(site.value, window.location.origin);
    assert.equal(site.readOnly, true);
    assert.equal(window.document.getElementById('magic-family-home').href, 'https://family.example.test/family');
    assert.equal(window.document.getElementById('connect-magic-api').textContent, '选择魔法 API 配置');
    assert.equal(window.localStorage.length, 0);
    assert.equal(requests.length, 0);
    window.document.getElementById('connect-magic-api').click();
    assert.equal(new URL(opened[0]).origin, window.location.origin);
    assert.equal(new URL(opened[0]).pathname, '/connect/recipes');
    assert.deepEqual([...new URL(opened[0]).searchParams.keys()], ['origin', 'nonce', 'kind']);
    assert.equal(requests.length, 0);
  } finally { window.close(); }
});

test('重建仅清理有生成标记的产品目录，不覆盖用户文件', async () => {
  const output = await mkdtemp(path.join(evidence, 'rebuild-'));
  const { recipes } = await buildFamilyProducts(output);
  await writeFile(path.join(recipes, 'obsolete.js'), 'obsolete');
  await buildFamilyProducts(output);
  assert.ok(!(await readdir(recipes)).includes('obsolete.js'));
  const protectedOutput = await mkdtemp(path.join(evidence, 'protected-'));
  await mkdir(path.join(protectedOutput, 'recipes'));
  await writeFile(path.join(protectedOutput, 'recipes/private.txt'), 'keep');
  await assert.rejects(buildFamilyProducts(protectedOutput), /不是本产品生成目录/);
  assert.equal(await readFile(path.join(protectedOutput, 'recipes/private.txt'), 'utf8'), 'keep');
});
