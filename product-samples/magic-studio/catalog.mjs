import { newImages, curatedImages, homepageImage, diverseImages, expansionImages, freshImages, videoPlans, originalImagePrompts } from './new-gallery.mjs';
const google = { label: 'Google 图片提示指南', url: 'https://ai.google.dev/gemini-api/docs/image-generation' };
const comfy = { label: 'ComfyUI 官方工作流', url: 'https://github.com/Comfy-Org/workflow_templates' };
const remotion = { label: 'Remotion 官方 Skills', url: 'https://www.remotion.dev/docs/ai/skills' };
const field = (key, label, value, hint = '') => ({ key, label, value, hint });
const image = (id, title, category, art, ratio, tags, summary, fields, zh, en, tip) => ({
 id, title, category, art, ratio, tags, summary, fields, zh, en, tip,
 media: 'image', type: 'prompt', tool: '生图工具', source: google, status: '未实测', previewLabel: '原创 AI 视觉示例',
 steps: ['选择支持对应能力的生图工具；整理好参考素材。', '替换下方主体、场景等材料，检查完整提示词。', '先出一张小样，检查构图、文字和主体细节，再调整。'],
 requirements: '示例图为原创 AI 视觉素材；下方可替换模板尚未在目标模型逐条验证。比例请在目标工具中设置。'
});
const originalItems = [
 image('jade-product', '一瓶清透的夏天', '电商产品', 'jade-bottle', 'portrait', ['产品摄影', '玻璃质感'], '把普通产品图整理成干净、有材质感的广告画面。',
 [field('subject', '产品主体', '无品牌玉绿色磨砂护肤瓶'), field('scene', '背景与道具', '浅色石灰石台座、一片细叶和柔和水面反光'), field('light', '光线', '大型柔光从侧面照射')],
 '为{{subject}}创作一张克制的产品广告摄影。场景是{{scene}}。{{light}}，保留自然阴影，真实呈现玻璃、金属与少量凝露的材质。产品完整可见，主体突出，背景简洁。画面不添加未经提供的标志、文案或额外产品。建议竖幅4:5。',
 'Create a refined product advertising photograph of {{subject}}. Setting: {{scene}}. Lighting: {{light}}. Show natural shadows, accurate glass and metal materials, subtle condensation, a fully visible product, and a quiet clean background. Keep the hero product distinct. Do not invent labels, copy, logos, or additional products. Suggested framing: portrait 4:5.',
 '需要准确品牌文字时，先提供真实包装参考图；不要依赖模型凭空重绘商标。'),
 image('paper-mascot', '纸做的小狐狸', '角色头像', 'paper-fox', 'square', ['纸艺', '3D', '吉祥物'], '用折纸质感做一个轮廓清楚、适合头像的小角色。',
 [field('subject', '角色', '一只坐着的小狐狸'), field('material', '材质与配色', '橙色折纸、白色纸胸口、明确折痕'), field('background', '背景', '浅薄荷色摄影棚地面')],
 '创作{{subject}}的精致三维纸艺形象。材质为{{material}}，背景为{{background}}。角色轮廓简单清楚，表情亲切，纸张折痕与厚度可信。柔和侧光，细腻接触阴影，构图留出呼吸空间。只出现一个角色，不加文字。建议正方形。',
 'Create a refined three-dimensional papercraft image of {{subject}}. Materials and colors: {{material}}. Background: {{background}}. Use a clean friendly silhouette, believable paper folds and thickness, soft side lighting, subtle contact shadows, and generous space. Show a single character without lettering. Suggested framing: square.',
 '同一角色需要多张图时，保留相同参考图、配色与轮廓描述，逐张检查。'),
 image('rain-character', '雨夜电影肖像', '角色头像', 'rain-portrait', 'portrait', ['电影感', '写实', '雨夜'], '以光线、景深和环境建立角色氛围。',
 [field('subject', '人物', '一位虚构的成年女性探险者'), field('scene', '场景', '蓝调时刻的安静雨夜街道'), field('palette', '光线色彩', '青色反光与远处琥珀色灯光')],
 '拍摄{{subject}}的电影近景肖像，人物处于{{scene}}。使用{{palette}}，浅景深，自然皮肤纹理，细小雨滴，真实防水外套材质。人物神态坚定而自然，保留人脸结构，不做夸张磨皮。背景服务于人物，不加文字。建议竖幅4:5。',
 'Create a cinematic close portrait of {{subject}} in {{scene}}. Use {{palette}}, shallow depth of field, natural skin texture, fine raindrops, and realistic waterproof fabric. Keep a dignified natural expression and believable facial anatomy. Let the background support the subject. No lettering. Suggested framing: portrait 4:5.',
 '这是一位虚构人物。使用真人参考时，应自行确认素材使用权和目标工具限制。'),
 image('cloud-world', '云海上的浮岛', '场景插画', 'floating-island', 'wide', ['幻想', '插画', '风景'], '把世界观变成有前景、中景和远景的画面。',
 [field('subject', '主要场景', '漂浮在云海上的绿色岛屿'), field('detail', '视觉焦点', '近处岛屿上的小型石制天文台与瀑布'), field('time', '时间与氛围', '温暖黎明、淡蓝与杏色天空')],
 '绘制{{subject}}的精细编辑插画。视觉焦点为{{detail}}。氛围是{{time}}。通过清晰前景、较淡中景和柔和远景建立空间，植物细节丰富，光线可信，画面安静而富有想象。不要让所有区域一样锐利。不加文字。建议横幅3:2。',
 'Paint a detailed editorial illustration of {{subject}}, focusing on {{detail}}. Atmosphere: {{time}}. Build depth with a crisp foreground, softer middle distance, and gentle far mountains. Use rich foliage, believable light, and a contemplative sense of wonder. Avoid equal sharpness everywhere. No lettering. Suggested framing: landscape 3:2.',
 '指定一个主要焦点；建筑、瀑布和岛屿数量都过多时，构图容易失去层次。'),
 image('citrus-cover', '柑橘色的留白海报', '海报社媒', 'citrus-poster', 'portrait', ['静物', '海报', '留白'], '先做好背景与主视觉，为自己的标题留出位置。',
 [field('subject', '静物主体', '切开的血橙、完整橘子与弯曲绿叶'), field('color', '背景主色', '明亮的橙红色'), field('space', '留白位置', '画面上半部')],
 '创作{{subject}}的编辑风格摄影海报底图。背景为{{color}}，主体主要分布在下半部，{{space}}保留干净空间供后期排字。硬朗午后阳光与明确斜向阴影，果皮、玻璃和叶片质感真实。画面不生成文字、标志或水印。建议竖幅4:5。',
 'Create an editorial photographic poster background featuring {{subject}} against {{color}}. Place the still life mainly in the lower half and leave {{space}} clean for typography added later. Use hard afternoon sunlight, crisp diagonal shadows, and tactile fruit, glass, and leaf textures. Generate no lettering, logos, or watermarks. Suggested framing: portrait 4:5.',
 '标题留给设计工具排版，可以避免生成文字出错，也更方便统一品牌字体。'),
 image('coffee-story', '咖啡馆的一方晨光', '电商产品', 'coffee-still', 'square', ['生活方式', '俯拍', '陶瓷'], '适合咖啡、家居器物和社媒内容的自然静物。',
 [field('subject', '器物', '装着黑咖啡的手工青瓷杯'), field('props', '搭配材料', '燕麦色亚麻布、银色小勺与一角可颂'), field('light', '光线', '自然清晨窗光')],
 '用俯拍视角拍摄{{subject}}，搭配{{props}}。光线为{{light}}。保持空气感与平衡，真实表现陶瓷釉面、亚麻纤维和食物纹理。暖中性色背景，少量柔和青色点缀。道具不抢主体，不加文字。建议正方形。',
 'Photograph {{subject}} from above with {{props}}. Use {{light}}, an airy balanced composition, believable ceramic glaze, linen fibers, and food texture. Work with warm neutral surfaces and a small muted teal accent. Props support the hero object. No lettering. Suggested framing: square.',
 '若需要与实物一致，上传准确器物照片，并检查杯柄、比例和颜色。'),
 image('mini-city', '掌心里的海滨城市', '场景插画', 'mini-city', 'square', ['等距', '3D', '微缩'], '把建筑、街道和地标整理成可读的微缩城市。',
 [field('subject', '城市主题', '海滨城市街区'), field('features', '关键元素', '白色弧形建筑、赤陶屋顶、棕榈树、黄色小电车'), field('background', '背景', '干净浅蓝色背景')],
 '制作{{subject}}的精细等距微缩三维插画。包含{{features}}，置于{{background}}。明确街道和海岸关系，体积与比例统一，柔和物理阴影，精巧手工感。只选择少量关键地标，不堆满全部建筑。不加地名或文字。建议正方形。',
 'Create a refined isometric miniature 3D illustration of {{subject}}, including {{features}}, against {{background}}. Keep streets and coast spatially readable, with consistent scale, soft physical shadows, and a finely crafted miniature feel. Select a few defining landmarks rather than crowding the scene. No labels or lettering. Suggested framing: square.',
 '这类画面表达城市主题，不代表真实地图；若需要真实地标，提供准确参考。'),
 image('calm-interior', '松弛的窗边客厅', '空间设计', 'calm-room', 'wide', ['室内', '写实', '自然光'], '先确定家具与材料，再用自然光营造空间。',
 [field('subject', '空间', '安静现代客厅'), field('materials', '家具与材料', '浅鼠尾草色沙发、橡木书架、米色灰泥墙'), field('view', '窗景', '柔和林地')],
 '拍摄{{subject}}的建筑杂志风格室内照片。使用{{materials}}，大窗外是{{view}}。自然晨光，细腻材料纹理，真实空间比例，克制装饰。家具之间有合理通道，透视连贯，画面不出现人物和文字。建议横幅3:2。',
 'Create an architectural editorial photograph of {{subject}} featuring {{materials}}, with {{view}} outside a large window. Use natural morning light, tactile materials, realistic proportions, and restrained decoration. Preserve usable circulation and coherent perspective. No people or lettering. Suggested framing: landscape 3:2.',
 '概念图不能代替施工设计；核对真实尺寸、通道和家具摆放。'),
 {
 id: 'remotion-caption', title: '给短片加上清晰字幕', media: 'video', type: 'skill', category: '字幕剪辑', tags: ['Remotion', '字幕', 'Agent Skill'], tool: 'Remotion', source: remotion,
 status: '未运行', summary: '使用官方字幕技能，让助手在项目里处理字幕节奏与排版。', fields: [field('subject', '视频内容', '30秒产品介绍短片'), field('style', '字幕风格', '两行以内、重点词突出、适合手机观看')],
 zh: '使用 Remotion 的 /remotion-captions 技能，为{{subject}}制定并实现字幕方案。字幕要求：{{style}}。先检查现有项目、实际字幕数据和素材，不编造对白与时间戳。给出可编辑字幕组件，检查安全边距、断句、字号、时间同步和字体可用性。缺少音频或转写时先列出所需材料，预览检查后再渲染。',
 en: 'Use Remotion /remotion-captions for {{subject}}. Caption style: {{style}}. Inspect the existing project, real caption data, and media first. Do not invent dialogue or timestamps. Create editable captions and check safe margins, line breaks, font availability, readability, and synchronization. Request missing audio or transcripts before rendering.',
 steps: ['打开官方Skills说明，检查当前版本与项目依赖。', '在自己的Remotion项目中选择字幕技能并提供真实字幕／音频。', '先预览断句与同步，再按项目说明渲染。'], requirements: '需要Remotion项目、对应运行环境与真实字幕数据。本站未安装运行此技能。', tip: '安装技能不会自动提供音频转写、素材或渲染资源。', command: 'npx skills add remotion-dev/skills'
 },
 {
 id: 'remotion-render', title: '把品牌故事做成动效', media: 'video', type: 'skill', category: '动效设计', tags: ['Remotion', 'React', 'Agent Skill'], tool: 'Remotion', source: remotion,
 status: '未运行', summary: '借助官方技能组织可编辑的标题、时间线和视频渲染。', fields: [field('subject', '内容', '三段品牌故事'), field('style', '风格要求', '克制青色、清楚排版、平滑转场')],
 zh: '使用 Remotion 官方最佳实践与创建／渲染技能，为{{subject}}制作可编辑视频项目。风格要求：{{style}}。先核对项目、实际文案、品牌素材和目标画幅，再说明组成与时间线。动画用可编辑组件实现，不将正文烘焙进截图。先预览，再按真实环境渲染；缺失素材及未运行步骤如实列出。',
 en: 'Use official Remotion best practices and creation/rendering skills to build an editable video for {{subject}}. Style: {{style}}. Check the project, actual copy, brand assets, and target aspect ratio first. Explain the compositions and timeline. Use editable components rather than baking body text into screenshots. Preview before rendering and list missing assets or unexecuted steps honestly.',
 steps: ['查阅官方技能目录和适用环境。', '让助手检查现有项目、整理实际文案及时间线。', '预览构图与动画，再运行项目渲染流程。'], requirements: '需要Remotion项目与本地运行环境，具体许可证及渲染条件以官方当前说明为准。本站未运行。', tip: '程序视频擅长精确字幕和动效，生成镜头素材仍需另备。', command: 'npx skills add remotion-dev/skills'
 },
 {
 id: 'comfy-image', title: '从官方模板开始生图', media: 'image', type: 'workflow', category: '本地工作流', tags: ['ComfyUI', '模板', '节点'], tool: 'ComfyUI', source: comfy,
 status: '未运行', summary: '先找与你的环境匹配的官方模板，再替换提示词和参数。', fields: [field('subject', '创作目标', '自然光产品摄影'), field('environment', '运行环境', '先填写ComfyUI版本、显存和已有模型')],
 zh: '我想用 ComfyUI 官方模板制作{{subject}}。我的环境：{{environment}}。请先根据官方当前模板确认模型文件、节点、硬件及版本要求，列出缺失项，再给出替换提示词和参数的位置。不自动下载或执行未知节点，不编造不存在的节点与模型。先保存原模板副本，再生成一张小样并记录使用的版本及参数。',
 en: 'Help me adapt an official ComfyUI template for {{subject}}. My environment: {{environment}}. Check current model files, nodes, hardware, and version requirements against the official template. List missing dependencies, then explain where to change prompts and parameters. Do not invent nodes or models. Preserve the original template and record the actual versions and settings for a first small test.',
 steps: ['从下方官方仓库或ComfyUI模板选择器找对应模板。', '核对模型、节点与硬件，保留原模板副本。', '替换提示词，先生成小样并记录参数。'], requirements: '需要可用ComfyUI环境与模板要求的模型／节点；此条是操作指引，不是已运行的工作流文件。', tip: '模板预览有结果，不等于你的机器已具备全部依赖。'
 },
 {
 id: 'comfy-video', title: '图片接进视频工作流', media: 'video', type: 'workflow', category: '本地工作流', tags: ['ComfyUI', '图生视频', '节点'], tool: 'ComfyUI', source: comfy,
 status: '未运行', summary: '核对起始图片、模型和节点条件，逐步搭起图生视频。', fields: [field('subject', '起始素材', '一张产品图'), field('goal', '期望动作', '主体保持稳定，镜头缓慢推进')],
 zh: '我想用官方 ComfyUI 图生视频模板处理{{subject}}，目标为{{goal}}。请先确认当前模板实际支持的输入、模型、节点版本、显存、画幅和时长。给出起始图导入、提示词、参数与结果预览的位置，缺失依赖单独列出。不要把静态图动画演示当成模型生成结果；先试短段，检查主体变形和帧间连续性。',
 en: 'Help me use an official ComfyUI image-to-video template with {{subject}}, aiming for {{goal}}. Verify supported inputs, models, node versions, memory requirements, aspect ratios, and duration. Explain the image input, prompts, parameters, and result preview. List missing dependencies. Test a short segment first and inspect deformation and temporal consistency.',
 steps: ['从官方模板中选实际支持图生视频的条目。', '核对环境及输入图尺寸，设置短段参数。', '检查连续性和变形，再增加时长或后期剪辑。'], requirements: '视频模板各自有不同模型、节点及硬件要求。本站没有运行或提供预装环境。', tip: '工作流与提示词分开保存版本，便于排查是素材、参数还是依赖变化。'
 }
];
export const items = [
 ...diverseImages,
 ...curatedImages,
 homepageImage,
 ...expansionImages,
 ...freshImages,
 ...newImages,
 ...originalItems.filter(item => item.media === 'image' && item.type === 'prompt').map(item => ({ ...item, rawPrompt: originalImagePrompts[item.art], rawPromptLabel: '这张图实际使用的原始提示词' })),
 ...videoPlans,
 ...originalItems.filter(item => item.type !== 'prompt')
].map(item => ({ ...item, styles: item.styles || (item.type !== 'prompt' ? [] :
 item.category === '动漫二次元' || item.category === '修仙动漫' ? ['二次元'] :
 item.category === '科技机甲' || ['cloud-world','dragon-cloud'].includes(item.id) ? ['概念设计'] :
 ['paper-mascot', 'mini-city', 'lantern-alley', 'candy-orbit'].includes(item.id) ? ['3D手作'] :
 item.media === 'video' ? ['电影感'] :
 item.category === '海报社媒' ? ['平面海报', '写实摄影'] : ['写实摄影']) }));
for(const item of items){if(['rain-character','sweet-woman','sweet-man','cyber-cosplay','midnight-editorial','fashion-editorial'].includes(item.id))item.styles=['写实摄影','电影感'];}
export const plays = [
 { id: 'product-story', title: '从一张产品图，到一段广告', summary: '先准备产品画面，研究都市分镜的光线和景别，再为自己的素材补字幕。视频条目目前仅分镜，尚无样片。', art: 'jade-bottle', steps: [{ id: 'jade-product', label: '准备产品画面' }, { id: 'elevator-secret', label: '研究都市镜头与光线' }, { id: 'remotion-caption', label: '为自己的短片加字幕' }] },
 { id: 'character-story', title: '让角色，进入自己的故事', summary: '先固定原创形象，再写出自然动作与情绪，不复刻现有剧集角色。', art: 'sweet-woman', steps: [{ id: 'sweet-woman', label: '建立原创人物肖像' }, { id: 'coast-reunion', label: '写出重逢与情绪特写' }, { id: 'comfy-video', label: '核对自己的视频工作流' }] },
 { id: 'little-city', title: '给小世界，一个大景别', summary: '从微缩城市画面出发，研究公路分镜的空间层次，整理成自己的短片。', art: 'mini-city', steps: [{ id: 'mini-city', label: '设计微缩城市' }, { id: 'coastal-choice', label: '研究公路大景别' }, { id: 'remotion-render', label: '组织可编辑视频' }] },
 { id: 'quiet-world', title: '做一个有生命的幻想世界', summary: '先设计世界，再写出两人冒险的动作节奏；两个条目仅作方法参考，不是同一生成过程。', art: 'floating-island', steps: [{ id: 'cloud-world', label: '绘制世界主视觉' }, { id: 'celestial-gate', label: '设计人物与环境互动' }, { id: 'comfy-video', label: '检查生成环境' }] }
];
export const typeLabels = { prompt: '提示词', skill: 'Skill', workflow: '工作流' };
export const artSizes = Object.fromEntries([
 [1122, 1402, ['brass-cartographer','frost-musician','jazz-cutout','blue-botanical','anime-sky-mechanic','window-portrait','anime-pilot','anime-summer','cherry-street','citrus-poster','city-skater','comic-hero','cyber-cosplay','fashion-editorial','festival-friends','ink-koi','jade-bottle','lantern-alley','lilac-cosplay','lotus-sword','rain-portrait','retro-observatory','scarlet-rider','sweet-man','sweet-woman','vinyl-afternoon']],
 [1536, 1024, ['red-panda-moss','turtle-blue','anime-rain-tram','sunken-lounge','lilac-bookshop','aurora-lagoon','calm-room','dragon-cloud','floating-island','pixel-night','snow-leopard','tide-mecha','watercolor-island']],
 [1254, 1254, ['candy-orbit','coffee-still','headphone-ad','mini-city','otter-morning','paper-fox','papercut-tiger']],
 [1672, 941, ['orbital-mecha']]
].flatMap(([width, height, names]) => names.map(name => [name, { width, height }])));
