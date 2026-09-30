export const catalog = [
  {
    id: 'meeting', file: '01-meeting-actions.md', category: 'work', title: '会议转待办',
    description: '把讨论变成有来源的任务。没说清的负责人和期限，留待确认。',
    outcome: '待办表、已确认决议、待确认问题', tags: ['会议', '任务', '团队'],
    fields: [
      { id: 'date', label: '会议日期与时区', hint: '不确定可以留空，不必猜日期。', type: 'text', required: false, example: '2026-09-28，Asia/Shanghai' },
      { id: 'material', label: '会议记录', hint: '用 [L1]、[L2] 标记每段，便于核对任务来源。', required: true, example: '[L1] 小林：我在2026-09-30提交首页文案初稿。\n[L2] 小周：我负责复核价格页，期限还没定。\n[L3] 小林：要不要加英文版？下次讨论。\n[L4] 小周：本轮先不做英文版。\n[L5] 小林：同意。' }
    ]
  },
  {
    id: 'requirements', file: '02-requirements-acceptance.md', category: 'work', title: '需求转验收清单',
    description: '先把要做什么、怎么验收说清楚，再开始实现。', outcome: '功能范围、验收条件、待确认项', tags: ['产品', '需求', '验收'],
    fields: [
      { id: 'goal', label: '这次想解决什么问题', type: 'text', required: true, example: '让用户更快找到自己的历史订单。' },
      { id: 'context', label: '已有功能与现状', required: false, hint: '只填已知事实，其他可留空。', example: '已有“我的订单”页面，其他实现未知。' },
      { id: 'material', label: '需求与限制', required: true, hint: '要求标为 [R1]，限制标为 [C1]；冲突也原样保留。', example: '[R1] 按订单号精确查询，只能查自己的订单。\n[R2] 搜索词为空时显示自己的全部订单，沿用现有分页。\n[R3] 无匹配结果时显示“未找到订单”。\n[R4] 不改支付流程。\n[C1] 不加导出，不加模糊搜索。' }
    ]
  },
  {
    id: 'triage', file: '03-error-triage.md', category: 'work', title: '报错转排查单',
    description: '先分清事实与猜测，再确定下一步要取什么证据。', outcome: '事实、待验证假设、只读取证顺序', tags: ['报错', 'API', '排查'],
    fields: [
      { id: 'environment', label: '运行环境与版本', type: 'text', required: false, example: '某本地客户端，版本未知。' },
      { id: 'expected', label: '预期行为', type: 'text', required: true, example: '请求返回模型列表。' },
      { id: 'actual', label: '实际发生了什么', type: 'text', required: true, example: '收到401。' },
      { id: 'steps', label: '复现步骤', required: true, example: '点击“测试连接”一次。' },
      { id: 'material', label: '脱敏错误与已有证据', required: true, hint: '按 [E1] 编号。不粘贴密钥、Cookie、密码或完整 .env。', example: '[E1] 客户端显示：GET https://api.example.invalid/v1/models -> 401。\n[E2] 同次响应正文：{"error":"missing authorization header"}。\n[E3] 设置页填过Key，但未检查实际请求。' }
    ]
  },
  {
    id: 'paper', file: '04-student-understand.md', category: 'university', title: '大学论文精读',
    description: '不止读懂摘要，也分清作者的结论究竟由什么证据支持。', outcome: '研究问题、方法、证据、局限与理解题', tags: ['大学', '论文', '文献', '研究'],
    fields: [
      { id: 'course', label: '专业、课程与已有基础', type: 'text', required: true, example: '大学研究方法入门，了解均值，不熟悉研究设计。' },
      { id: 'goal', label: '本次阅读目标', type: 'text', required: true, example: '理解研究结果能否支持因果结论。' },
      { id: 'scope', label: '材料范围与书目信息', required: false, hint: '全文、摘要或部分章节？作者、年份不知道就不填。', example: '以下三个虚构教学片段，不对应真实论文，无书目信息。' },
      { id: 'material', label: '论文片段', required: true, hint: '按 [P1] 编号；只提供获准使用的材料。', example: '[P1] 某课程观察40名学生，20人自行选择使用练习工具，20人未使用；不是随机分组。\n[P2] 期末成绩均值分别为78分和72分。片段未提供标准差、统计检验或入学基础测量。\n[P3] 作者认为工具可能帮助学习，并说明无法排除两组原有差异。' },
      { id: 'question', label: '最不理解的地方', required: false, example: '组间相差6分，为什么不能说工具让人成绩提高6分？' }
    ]
  },
  {
    id: 'lab', file: '05-student-mistake-review.md', category: 'university', title: '大学编程实验复盘',
    description: '从自己的代码出发，定位问题、理解修正，再用真实运行验证。', outcome: '局部提示、验证建议、有依据的复盘', tags: ['大学', '代码', 'Python', '实验'],
    fields: [
      { id: 'course', label: '课程、语言与版本', type: 'text', required: true, example: '大学Python程序设计，具体版本未记录。' },
      { id: 'material', label: '实验要求、代码与真实现象', required: true, hint: '要求按 [R1] 编号；附输入、输出、预期依据，不上传凭证。', example: '[R1] 对非空数值列表计算算术平均值。\n代码：\ndef mean_score(scores):\n    total = 0\n    for score in scores:\n        total += score\n    return total / (len(scores) - 1)\n\n输入：[60, 80, 100]\n学生报告输出：120.0\n预期：80，依据R1。空列表行为未定义。' },
      { id: 'attempt', label: '已经试过的修改', required: false, example: '还没有修改，希望先理解错误。' }
    ]
  },
  {
    id: 'revision', file: '06-student-revision-plan.md', category: 'university', title: '大学期末复习',
    description: '用你的课程材料和实际时间，排出有产物、不超时的复习计划。', outcome: '复习清单、时间核算、自编理解题', tags: ['大学', '期末', '复习', '线性代数'],
    fields: [
      { id: 'course', label: '大学课程与已有基础', type: 'text', required: true, example: '大学线性代数，学过矩阵乘法和消元。' },
      { id: 'exam', label: '考试日期', type: 'text', required: false, hint: '未确定就留空，不需要猜。', example: '未定' },
      { id: 'material', label: '课程范围、时间与限制', required: true, hint: '主题用 [T1]，日期用 [D1]；每天可用分钟包含休息。', example: '[T1] 线性方程组与消元，行变换易错，优先复习。\n[T2] 矩阵乘法及维度，容易搞错能否相乘。\n[D1] 2026-09-29：30分钟，含休息。\n[D2] 2026-09-30：20分钟，含休息。\n[D3] 2026-10-01：40分钟，含休息。\n材料：自己的讲义和已做练习，题号未提供。\n限制：不加时；考试权重与形式未知。' }
    ]
  },
  {
    id: 'image', file: '07-image-prompt.md', category: 'creative', target: 'image', title: '生图提示词',
    description: '把主体、风格和构图说明白，整理成可直接交给生图工具的提示词。',
    outcome: '可复制的生图请求，接入生图模型后可直接生成图片', tags: ['创作', '生图', '图片', '海报', '摄影'],
    fields: [
      { id: 'scene', label: '主体与场景', required: true, hint: '描述要画什么、在哪里；不要只写“高级”“好看”。', example: '雾青森林中的一座小型玻璃温室，内部有暖色灯光，周围是湿润蕨类，没有人物。' },
      { id: 'style', label: '风格、色调与光线', required: false, emptyValue: '不额外指定风格，保持光线与场景协调。', example: '写实建筑摄影，冷青环境与暖黄室内形成对比，清晨柔和光线。' },
      { id: 'layout', label: '构图与视角', required: false, emptyValue: '不额外指定构图，清晰突出所描述的主体。', example: '平视中景，温室位于画面右侧，左侧留出干净空间。' },
      { id: 'format', label: '画幅与用途', type: 'text', required: false, hint: '这里只记录文字要求，比例仍需在目标工具中确认。', emptyValue: '不指定比例，使用目标工具当前设置。', example: '横向16:9，用作网站横幅；不加入网站界面。' },
      { id: 'text', label: '画面文字', required: false, hint: '留空默认不添加文字；需要文字时写明原文和位置。', emptyValue: '不添加文字、标志或水印。', example: '不添加文字、标志或水印。' },
      { id: 'avoid', label: '希望避免的内容', required: false, emptyValue: '不额外指定避免项。', example: '人物、车辆、霓虹招牌、杂乱背景。' }
    ]
  },
  {
    id: 'video', file: '08-video-prompt.md', category: 'creative', target: 'video', title: '视频提示词',
    description: '按主体、动作、镜头和时间整理一个短片段，复制到支持视频生成的工具。',
    outcome: '一段含动作与镜头要求的视频提示词', tags: ['创作', '视频', '镜头', '短片', '分镜'],
    fields: [
      { id: 'scene', label: '主体与环境', required: true, hint: '先做一个片段；没有上传参考图时，用文字说明主体外观。', example: '木桌上的透明玻璃杯，杯中是温热茶水，窗外背景虚化，无人物。' },
      { id: 'action', label: '动作与变化顺序', required: true, hint: '写清先发生什么、再发生什么；分段时长请自行核算。', example: '0—2秒从杯子中景开始；2—5秒镜头缓慢推近，蒸汽持续上升；5—6秒停在杯口细节。' },
      { id: 'camera', label: '镜头与运动', required: false, emptyValue: '不额外指定运镜，保持镜头稳定，不增加切镜。', example: '单镜头缓慢推近，不切镜，不绕拍。' },
      { id: 'style', label: '风格、光线与色调', required: false, emptyValue: '不额外指定风格，保持场景与光线一致。', example: '写实静物摄影，柔和侧光，暖色调。' },
      { id: 'timing', label: '时长与画幅', type: 'text', required: true, hint: '这是文字要求，不会自动设置工具参数；请确认工具支持的时长。', example: '6秒，横向16:9，单镜头；在目标工具中另行确认设置。' },
      { id: 'audio', label: '声音与字幕', required: false, hint: '留空默认不要求音频和字幕，不假定目标工具支持声音。', emptyValue: '不要求音频，不加字幕。', example: '不要求音频，不加字幕。' },
      { id: 'avoid', label: '希望避免的内容', required: false, emptyValue: '不额外指定避免项。', example: '手部入镜、杯子变形、液体溢出、突然切镜、文字或水印。' }
    ]
  }
];
