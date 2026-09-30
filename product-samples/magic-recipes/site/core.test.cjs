const test = require('node:test');
const assert = require('node:assert/strict');
const core = require('./core.cjs');
const recipe = { id: 'test', category: 'university', title: '大学论文', description: '证据阅读', tags: ['研究'], prompt: '目标：{{field:goal}}\n材料：{{field:material}}', fields: [{ id: 'goal', label: '目标', required: false }, { id: 'material', label: '材料', required: true }] };
test('必填空白返回字段错误', () => assert.equal(core.validate(recipe, { material: '   ' })[0].id, 'material'));
test('缺失可选字段明确标为未知', () => assert.match(core.compile(recipe, { material: '事实' }).text, /未提供/));
test('字段默认值处理空白，用户输入优先于默认值', () => {
  const custom = { ...recipe, fields: recipe.fields.map(field => field.id === 'goal' ? { ...field, emptyValue: '不添加文字' } : field) };
  assert.match(core.compile(custom, { material: '事实', goal: '   ' }).text, /目标：不添加文字/);
  assert.match(core.compile(custom, { material: '事实', goal: '标题' }).text, /目标：标题/);
});
test('只替换模板占位符，不递归解释用户材料', () => {
  const value = '{{field:goal}} $& <script>alert(1)</script>';
  assert.ok(core.compile(recipe, { goal: '阅读', material: value }).text.endsWith(value));
});
test('保留代码缩进和材料内部换行', () => assert.ok(core.compile(recipe, { material: 'def run():\n    pass' }).text.endsWith('def run():\n    pass')));
test('拒绝超长字段', () => assert.equal(core.validate(recipe, { material: '文'.repeat(12001) }).length, 1));
test('允许边界长度', () => assert.equal(core.validate(recipe, { material: '文'.repeat(12000) }).length, 0));
test('分类筛选不显示工作项', () => assert.equal(core.filter([recipe], 'work', '').length, 0));
test('搜索标签及空格兼容', () => assert.equal(core.filter([recipe], 'all', ' 研究 ').length, 1));
test('无匹配返回空数组', () => assert.equal(core.filter([recipe], 'all', '不存在').length, 0));
test('生成失败不返回可复制文本', () => assert.equal(core.compile(recipe, {}).text, ''));
