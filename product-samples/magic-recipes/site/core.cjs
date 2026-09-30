const RecipeCore = (() => {
  function validate(recipe, values) {
    return recipe.fields.flatMap(field => {
      const value = String(values[field.id] ?? '');
      if (field.required && !value.trim()) return [{ id: field.id, message: '请填写' + field.label + '。' }];
      if (value.length > 12000) return [{ id: field.id, message: field.label + '超过12000字，请保留必要片段。' }];
      return [];
    });
  }
  function compile(recipe, values) {
    const errors = validate(recipe, values);
    if (errors.length) return { errors, text: '' };
    const text = recipe.prompt.replace(/{{field:([a-z]+)}}/g, (match, fieldId) => {
      const field = recipe.fields.find(item => item.id === fieldId);
      if (!field) throw new Error('未定义字段：' + fieldId);
      return String(values[fieldId] ?? '').trim() || field.emptyValue || '未提供，请按规则处理未知信息';
    });
    return { errors: [], text };
  }
  function filter(recipes, category, query) {
    const term = query.trim().toLocaleLowerCase();
    return recipes.filter(recipe => (category === 'all' || recipe.category === category) &&
      [recipe.title, recipe.description, ...recipe.tags].join(' ').toLocaleLowerCase().includes(term));
  }
  return { validate, compile, filter };
})();
if (typeof module !== 'undefined' && module.exports) module.exports = RecipeCore;
else globalThis.MofaRecipeCore = RecipeCore;
