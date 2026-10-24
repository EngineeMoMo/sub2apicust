import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import test from 'node:test';
import {
  checkSources, classifyFiles, compareCommits, githubJson, inspectSource,
  main, parseArguments, renderReport, validateManifest,
} from './check-updates.mjs';

const baseSha = 'a'.repeat(40);
const nextSha = 'b'.repeat(40);
const reviewSha = 'c'.repeat(40);
const checkedAt = '2026-10-23T08:00:00Z';
const sourceFixture = overrides => ({
  id: 'reference', repository: 'example/kiro', branch: 'main',
  observedCommit: baseSha, reviewedCommit: baseSha, portedCommit: null,
  observedAt: '2026-10-20T00:00:00Z', observedReleases: [], watchPaths: ['src/', 'go.mod'], ...overrides,
});
const stateFixture = sources => ({ schemaVersion: 1, sources });
const manifestFixture = sources => ({ schemaVersion: 1, sources });
const commitFixture = number => ({ sha: number.toString(16).padStart(40, '0'), commit: { message: `fix: ${number}`, committer: { date: checkedAt } } });
const comparisonFixture = overrides => ({
  status: 'ahead', total_commits: 1, commits: [commitFixture(1)], files: [{ filename: 'src/auth.rs', status: 'modified' }], ...overrides,
});
const releaseFixture = overrides => ({
  id: 123, tag_name: 'v1.0.0', draft: false, prerelease: false,
  published_at: '2026-10-22T00:00:00Z', updated_at: '2026-10-22T00:00:00Z', body: 'refresh fix', ...overrides,
});

function apiFixture({ source = sourceFixture(), canonical = source.repository, branch = source.branch, head = baseSha, comparison, releases = [], archived = false, disabled = false } = {}) {
  const requests = [];
  return {
    requests,
    get: async path => {
      requests.push(path);
      if (path === `/repos/${source.repository}`) return { full_name: canonical, default_branch: branch, archived, disabled };
      if (path === `/repos/${canonical}/commits/${encodeURIComponent(branch)}`) return { sha: head };
      if (path.includes('/compare/')) return comparison ?? comparisonFixture();
      if (path.includes('/releases?')) return releases;
      throw new Error(`未预期请求：${path}`);
    },
  };
}

test('版本化清单有六个固定来源，尚无业务移植基线', async () => {
  const manifest = validateManifest(JSON.parse(await readFile(new URL('./sources.lock.json', import.meta.url), 'utf8')));
  assert.equal(manifest.sources.length, 6);
  assert.ok(manifest.sources.every(source => source.portedCommit === null));
  assert.equal(manifest.sources.filter(source => source.reviewedCommit !== null).length, 3);
});

test('拒绝重复来源、短 SHA、空关注路径和错误清单版本', () => {
  const source = sourceFixture();
  assert.throws(() => validateManifest(manifestFixture([source, source])), /重复/);
  assert.throws(() => validateManifest(manifestFixture([sourceFixture({ observedCommit: 'abc' })])), /完整 SHA/);
  assert.throws(() => validateManifest(manifestFixture([sourceFixture({ watchPaths: [] })])), /不能为空/);
  assert.throws(() => validateManifest({ schemaVersion: 2, sources: [source] }), /版本/);
});

test('没有更新时不请求 compare，不产生提醒，保留独立审阅及移植字段', async () => {
  const source = sourceFixture({ reviewedCommit: reviewSha });
  const api = apiFixture();
  const original = JSON.stringify(source);
  const result = await inspectSource(source, null, api.get, { checkedAt });
  assert.equal(result.result.actionable, false);
  assert.equal(result.result.relevance, 'none');
  assert.equal(result.result.unreviewed, true);
  assert.equal(result.result.reviewedCommit, reviewSha);
  assert.equal(result.result.portedCommit, null);
  assert.equal(result.state.head, baseSha);
  assert.equal(api.requests.filter(path => path.includes('/compare/')).length, 0);
  assert.equal(JSON.stringify(source), original);
});

test('代码更新需要处理；观察游标推进不代表已审阅或已移植', async () => {
  const source = sourceFixture();
  const api = apiFixture({ head: nextSha });
  const { report, nextState } = await checkSources(manifestFixture([source]), stateFixture({}), api.get, { checkedAt });
  assert.equal(report.failed, 0);
  assert.equal(report.actionable, 1);
  assert.equal(report.results[0].relevance, 'code');
  assert.equal(nextState.sources.reference.head, nextSha);
  assert.equal(source.reviewedCommit, baseSha);
  assert.equal(source.portedCommit, null);
});

test('同一更新第二次观察保持安静，但待审阅差异仍可追溯', async () => {
  const source = sourceFixture();
  const api = apiFixture({ head: nextSha });
  const first = await checkSources(manifestFixture([source]), stateFixture({}), api.get, { checkedAt });
  const second = await checkSources(manifestFixture([source]), first.nextState, api.get, { checkedAt });
  assert.equal(first.report.actionable, 1);
  assert.equal(second.report.actionable, 0);
  assert.equal(second.report.results[0].unreviewed, true);
  assert.ok(second.report.results[0].reviewUrl.endsWith(`${baseSha}...${nextSha}`));
});

test('纯 README 和图片更新留记录，不冒充功能修复', async () => {
  const api = apiFixture({ head: nextSha, comparison: comparisonFixture({ files: [{ filename: 'README.md' }, { filename: 'images/banner.png' }] }) });
  const { result, state } = await inspectSource(sourceFixture(), null, api.get, { checkedAt });
  assert.equal(result.relevance, 'documentation');
  assert.equal(result.actionable, false);
  assert.equal(state.head, nextSha);
});

test('重命名前在关注路径内的文件仍算代码变化；未知文件保守要求审阅', () => {
  assert.equal(classifyFiles([{ filename: 'docs/example.md', previous_filename: 'src/refresh.rs' }], ['src/']), 'code');
  assert.equal(classifyFiles([{ filename: '.github/workflows/build.yml' }], ['src/']), 'manual-review');
  assert.equal(classifyFiles([], ['src/']), 'manual-review');
});

test('compare 超过一页时完整分页并保留第一页文件列表', async () => {
  const requests = [];
  const result = await compareCommits('example/kiro', baseSha, nextSha, async path => {
    requests.push(path);
    if (path.endsWith('page=1')) return comparisonFixture({ total_commits: 101, commits: Array.from({ length: 100 }, (unused, index) => commitFixture(index + 1)) });
    return { status: 'ahead', total_commits: 101, commits: [commitFixture(101)] };
  });
  assert.equal(requests.length, 2);
  assert.equal(result.commits.length, 101);
  assert.equal(result.files[0].filename, 'src/auth.rs');
  assert.equal(result.incomplete, false);
});

test('分页上限或服务端提前空页不能被当作完整无更新', async () => {
  const incomplete = comparisonFixture({ total_commits: 101, commits: Array.from({ length: 100 }, (unused, index) => commitFixture(index + 1)), files: [{ filename: 'README.md' }] });
  const capped = await compareCommits('example/kiro', baseSha, nextSha, async () => incomplete, { maxPages: 1 });
  assert.equal(capped.incomplete, true);
  assert.equal(classifyFiles(capped.files, ['src/'], capped.incomplete), 'manual-review');
  let pages = 0;
  const earlyEmpty = await compareCommits('example/kiro', baseSha, nextSha, async () => {
    pages += 1;
    return pages === 1 ? incomplete : comparisonFixture({ total_commits: 101, commits: [] });
  });
  assert.equal(earlyEmpty.incomplete, true);
});

test('GitHub 300 文件上限保守标记不完整，即使所有可见文件均为文档', async () => {
  const result = await compareCommits('example/kiro', baseSha, nextSha, async () => comparisonFixture({ files: Array.from({ length: 300 }, (unused, index) => ({ filename: `docs/${index}.md` })) }));
  assert.equal(result.incomplete, true);
  assert.equal(classifyFiles(result.files, ['src/'], result.incomplete), 'manual-review');
});

test('force push 分叉与回退都要求处理，不当成普通增量', async () => {
  for (const status of ['diverged', 'behind']) {
    const api = apiFixture({ head: nextSha, comparison: comparisonFixture({ status, total_commits: 0, commits: [], files: [] }) });
    const { result } = await inspectSource(sourceFixture(), null, api.get, { checkedAt });
    assert.equal(result.comparison.status, status);
    assert.equal(result.actionable, true);
  }
});

test('仓库迁移及默认分支变化被记录；含斜杠分支正确编码', async () => {
  const api = apiFixture({ canonical: 'newowner/kiro', branch: 'release/stable' });
  const { result } = await inspectSource(sourceFixture(), null, api.get, { checkedAt });
  assert.deepEqual(result.metadataChanges, ['仓库迁移或重命名', '默认分支变化']);
  assert.equal(result.repository, 'newowner/kiro');
  assert.ok(api.requests.includes('/repos/newowner/kiro/commits/release%2Fstable'));
});

test('仓库归档和重新启用都产生提醒', async () => {
  const source = sourceFixture();
  const archivedApi = apiFixture({ archived: true });
  const archived = await inspectSource(source, null, archivedApi.get, { checkedAt });
  assert.deepEqual(archived.result.metadataChanges, ['仓库已归档']);
  const restored = await inspectSource(source, archived.state, apiFixture().get, { checkedAt });
  assert.deepEqual(restored.result.metadataChanges, ['仓库恢复维护']);
});

test('首次观察历史 Release 不误报；观察基线之后的新发布会提醒', async () => {
  const oldRelease = releaseFixture({ id: 100, published_at: '2026-10-01T00:00:00Z', updated_at: '2026-10-01T00:00:00Z' });
  const newRelease = releaseFixture({ prerelease: true });
  const api = apiFixture({ releases: [newRelease, oldRelease] });
  const { result } = await inspectSource(sourceFixture(), null, api.get, { checkedAt });
  assert.equal(result.releaseChanges.length, 1);
  assert.equal(result.releaseChanges[0].id, newRelease.id);
  assert.equal(result.releaseChanges[0].prerelease, true);
  assert.equal(result.actionable, true);
});

test('Release 正文改变即使更新时间相同也可检出', async () => {
  const release = releaseFixture();
  const source = sourceFixture({ observedReleases: [{ id: release.id, updatedAt: release.updated_at, bodySha256: createHash('sha256').update('old').digest('hex') }] });
  const { result } = await inspectSource(source, null, apiFixture({ releases: [release] }).get, { checkedAt });
  assert.equal(result.releaseChanges[0].change, 'edited');
});

test('发布列表分页达到上限时明确待审，不静默截断', async () => {
  const releases = Array.from({ length: 100 }, (unused, index) => releaseFixture({ id: index, published_at: '2026-10-01T00:00:00Z', updated_at: '2026-10-01T00:00:00Z' }));
  const { result } = await inspectSource(sourceFixture(), null, apiFixture({ releases }).get, { checkedAt, maxPages: 1 });
  assert.equal(result.releasesIncomplete, true);
  assert.equal(result.actionable, true);
});

test('一个仓库限流时保留原游标，其他来源仍检查；失败不能表示无更新', async () => {
  const failing = sourceFixture();
  const healthy = sourceFixture({ id: 'healthy', repository: 'example/healthy' });
  const prior = { requestedRepository: failing.repository, head: reviewSha, checkedAt: failing.observedAt, releases: [] };
  const healthyApi = apiFixture({ source: healthy });
  const result = await checkSources(manifestFixture([failing, healthy]), stateFixture({ reference: prior }), async path => {
    if (path === '/repos/example/kiro') throw new Error('GitHub HTTP 429');
    return healthyApi.get(path);
  }, { checkedAt });
  assert.equal(result.report.failed, 1);
  assert.equal(result.report.results.length, 2);
  assert.deepEqual(result.nextState.sources.reference, prior);
  assert.equal(result.nextState.sources.healthy.head, baseSha);
  assert.match(renderReport(result.report), /不能解释为没有更新/);
});

test('Release 检查失败后不提前推进已成功读取的 HEAD', async () => {
  const source = sourceFixture();
  const api = apiFixture({ head: nextSha });
  const checked = await checkSources(manifestFixture([source]), stateFixture({}), async path => {
    if (path.includes('/releases?')) throw new Error('GitHub HTTP 503');
    return api.get(path);
  }, { checkedAt });
  assert.equal(checked.report.failed, 1);
  assert.equal(checked.nextState.sources.reference, undefined);
});

test('同一 id 改指另一仓库时不沿用旧仓的观察 SHA', async () => {
  const source = sourceFixture();
  const prior = { requestedRepository: 'other/repository', head: reviewSha };
  const { result } = await inspectSource(source, prior, apiFixture().get, { checkedAt });
  assert.equal(result.base, baseSha);
  assert.equal(result.actionable, false);
});

test('损坏的观察状态不得被静默重置成一次成功检查', async () => {
  const source = sourceFixture();
  for (const sources of [null, []]) {
    await assert.rejects(checkSources(manifestFixture([source]), stateFixture(sources), apiFixture().get), /状态格式/);
  }
  const prior = { requestedRepository: source.repository, head: 'bad', checkedAt, releases: [] };
  const checked = await checkSources(manifestFixture([source]), stateFixture({ reference: prior }), apiFixture().get, { checkedAt });
  assert.equal(checked.report.failed, 1);
  assert.match(checked.report.results[0].error, /观察游标损坏/);
  assert.deepEqual(checked.nextState.sources.reference, prior);
});

test('无效 HEAD、compare 契约缺失、分页不一致及错误状态明确报错', async () => {
  await assert.rejects(inspectSource(sourceFixture(), null, apiFixture({ head: 'short' }).get, { checkedAt }), /HEAD 格式/);
  await assert.rejects(compareCommits('example/kiro', baseSha, nextSha, async () => ({ commits: [] })), /状态缺失/);
  let page = 0;
  await assert.rejects(compareCommits('example/kiro', baseSha, nextSha, async () => {
    page += 1;
    return comparisonFixture({ total_commits: page === 1 ? 101 : 102 });
  }), /分页期间/);
  await assert.rejects(checkSources(manifestFixture([sourceFixture()]), { schemaVersion: 9, sources: {} }, apiFixture().get), /状态格式/);
});

test('Markdown 报告转义上游文本；不完整报告显式说明', async () => {
  const comparison = comparisonFixture({ commits: [{ sha: nextSha, commit: { message: '<script>bad</script> | [fake](url)' } }], files: Array.from({ length: 300 }, () => ({ filename: 'README.md' })) });
  const { result } = await inspectSource(sourceFixture(), null, apiFixture({ head: nextSha, comparison }).get, { checkedAt });
  const markdown = renderReport({ checkedAt, failed: 0, actionable: 1, results: [result] });
  assert.ok(!markdown.includes('<script>'));
  assert.match(markdown, /列表不完整/);
  assert.match(markdown, /尚未移植/);
});

test('CLI 参数缺值、未知选项及非法网络路径或 Token 不会开始请求', () => {
  assert.throws(() => parseArguments(['--out']), /缺少路径/);
  assert.throws(() => parseArguments(['--unknown']), /未知参数/);
  assert.equal(parseArguments(['--fail-on-update']).failOnUpdate, true);
  assert.throws(() => githubJson('https://example.com'), /只允许访问/);
  assert.throws(() => githubJson('/repos/example/kiro', { token: 'bad\nvalue' }), /Token 格式/);
});

test('完整运行生成报告及游标，退出2/0/1分别表示新变化、无新变化和失败', async context => {
  const originalSummary = process.env.GITHUB_STEP_SUMMARY;
  delete process.env.GITHUB_STEP_SUMMARY;
  const fixtureDirectory = await mkdtemp(join(tmpdir(), 'kiro-reference-test-'));
  context.after(async () => {
    if (originalSummary === undefined) delete process.env.GITHUB_STEP_SUMMARY;
    else process.env.GITHUB_STEP_SUMMARY = originalSummary;
    assert.equal(dirname(resolve(fixtureDirectory)), resolve(tmpdir()));
    assert.ok(fixtureDirectory.startsWith(join(tmpdir(), 'kiro-reference-test-')));
    await rm(fixtureDirectory, { recursive: true, force: true });
  });
  const manifestPath = join(fixtureDirectory, 'sources.json');
  const out = join(fixtureDirectory, 'report');
  const manifestText = JSON.stringify(manifestFixture([sourceFixture()]));
  await writeFile(manifestPath, manifestText);
  const args = ['--manifest', manifestPath, '--out', out, '--fail-on-update'];
  const api = apiFixture({ head: nextSha });
  assert.equal(await main(args, api.get), 2);
  const first = JSON.parse(await readFile(join(out, 'latest.json'), 'utf8'));
  assert.equal(first.actionable, 1);
  assert.equal(first.failed, 0);
  assert.equal(await main(args, api.get), 0);
  const savedState = await readFile(join(out, 'state.json'), 'utf8');
  assert.equal(await main(args, async () => { throw new Error('GitHub HTTP 429'); }), 1);
  assert.equal(await readFile(join(out, 'state.json'), 'utf8'), savedState);
  assert.match(await readFile(join(out, 'latest.md'), 'utf8'), /检查失败/);
  assert.equal(await readFile(manifestPath, 'utf8'), manifestText);
});
