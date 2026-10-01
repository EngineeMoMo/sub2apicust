import { execFileSync } from 'node:child_process';
import { copyFile, mkdir, readFile, readdir, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const frontend = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..');
const workspace = path.dirname(frontend);
const marker = '.mofa-generated';
const markerValue = 'mofa-family-static-v1';

export async function buildFamilyProducts(output = path.join(frontend, 'public')) {
  const root = path.resolve(output);
  await mkdir(root, { recursive: true });
  for (const product of ['recipes', 'studio']) {
    const target = path.resolve(root, product);
    if (path.dirname(target) !== root) throw new Error('产品输出目录越界');
    let entries;
    try { entries = await readdir(target); }
    catch (error) { if (error.code !== 'ENOENT') throw error; entries = []; }
    if (entries.length) {
      if (await readFile(path.join(target, marker), 'utf8').catch(() => '') !== markerValue) throw new Error('已有目录不是本产品生成目录：' + target);
      await rm(target, { recursive: true });
    }
    await mkdir(target, { recursive: true });
    await writeFile(path.join(target, marker), markerValue, 'utf8');
  }
  const recipes = path.join(root, 'recipes');
  execFileSync(process.execPath, [path.join(workspace, 'product-samples/magic-recipes/site/build.mjs'), '--embedded', '--output', recipes], { stdio: 'pipe' });
  const studio = path.join(root, 'studio');
  const source = path.join(workspace, 'product-samples/magic-studio');
  await copyFile(path.join(workspace, 'product-samples/family-runtime/host-client.js'), path.join(studio, 'host-client.js'));
  for (const file of ['index.html', 'theme.css', 'app.mjs', 'core.mjs', 'icons.mjs', 'catalog.mjs', 'new-gallery.mjs']) {
    if (file === 'index.html') {
      const html = (await readFile(path.join(source, file), 'utf8'))
        .replace('<head>', '<head>\n<meta name="mofa-api-site" content="same-origin">')
        .replace('href="https://ai.mofamilys.com/family"', 'href="/family"');
      const integrated = html.replace('href="https://ai.mofamilys.com/tools/studio/submit"', 'href="/tools/studio/submit"');
      await writeFile(path.join(studio, file), integrated, 'utf8');
    } else await copyFile(path.join(source, file), path.join(studio, file));
  }
  await mkdir(path.join(studio, 'assets'), { recursive: true });
  for (const file of await readdir(path.join(source, 'assets'))) {
    if (/^[a-z0-9-]+\.(webp|mp4)$/.test(file) || file === 'mofa-mark-flat.png') {
      await copyFile(path.join(source, 'assets', file), path.join(studio, 'assets', file));
    }
  }
  return { recipes, studio };
}
