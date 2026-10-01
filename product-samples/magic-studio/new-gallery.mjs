const field = (key, label, value) => ({ key, label, value });
const art = (id, title, category, ratio, subject, scene, light, tags) => ({
  id, title, category, ratio, art: id, media: 'image', type: 'prompt', tool: '生图工具', tags,
  summary: scene, status: '模板待实测', previewLabel: '原创 AI 视觉示例',
  fields: [field('subject', '主体', subject), field('scene', '场景与构图', scene), field('light', '光线与质感', light)],
  zh: '创作一张原创画面：{{subject}}。{{scene}}。{{light}}。保持人物为明确成年人，比例和手部自然，保留真实细节；产品结构连贯，机械与动物符合主体逻辑。不要照搬真人、现有角色、品牌标志；不添加文字、水印或界面。画幅在目标工具中设置。',
  en: 'Create an original image of {{subject}}. {{scene}}. {{light}}. Any person must be clearly adult, with natural anatomy and believable hands. Preserve coherent product, mechanical and animal structure. Do not copy real people, existing characters or brand marks. No text, watermark or UI. Set aspect ratio in the target tool.',
  steps: ['在支持生图的工具中选择模型与画幅。', '替换主体与场景，先生成小样。', '检查手部、结构、身份与材质，再优化。'],
  requirements: '展示图由内置 image_gen 原创生成；可替换模板没有逐条在目标模型实测，不保证复制后产生相同结果。',
  tip: '人物均为虚构成年人。风格可以借鉴，脸、服装和故事应自己创作。',
  source: { label: '图片提示方法 · 官方指南', url: 'https://ai.google.dev/gemini-api/docs/image-generation' }
});
export const newImages = [
  art('sweet-woman', '书店窗边，刚好的微笑', '年轻人像', 'portrait', '25岁的虚构成年女性，棕色微卷发，温柔自然的微笑', '书店咖啡角，鼠尾草绿针织上衣、浅蓝条纹衬衫与牛仔裤，双手自然拿杯和扶书', '暖色窗光与冷色街景反射平衡，真实肤质，独立电影质感', ['女性', '甜美', '电影人像']),
  art('sweet-man', '夏雨之后的海岸街角', '年轻人像', 'portrait', '26岁的虚构成年男性，深色卷发，亲和清爽的微笑', '海岸街道，奶油色T恤与海军蓝衬衫，一手扶蓝色自行车，一手提花束', '傍晚暖光与蓝绿路面反射，自然肤质，生活化抓拍', ['男性', '清新', '电影人像']),
  art('fashion-editorial', '海风里的丝绸与牛仔', '时尚肖像', 'portrait', '27岁的虚构成年女性，松弛自信的气质', '海岸屋顶，陶土色丝质吊带、敞开的亚麻衬衫与高腰牛仔裤，手臂自然倚靠石栏', '午后自然光，细微性感但不裸露，克制杂志摄影', ['女性', '时尚', '穿搭']),
  art('cyber-cosplay', '末班车之前 · 星际飞行员', 'Cosplay', 'portrait', '26岁的虚构成年女性飞行员，短银发，原创科幻装甲服', '雨后的高架车站，墨绿与炭灰不透明战术服，头盔拿在腰侧', '车站柔光和雨地倒影，真实衣料与肤质，电影级景深', ['cosplay', '科技', '女性']),
  art('anime-pilot', '云层下的航海图', '动漫二次元', 'portrait', '25岁的原创成年动漫女性飞行员', '未来海滨城的旧阳台，蓝色飞行外套与棕色长裤，手握折叠地图', '手绘二维动画线条、暖色云层、丰富背景，不用塑料3D质感', ['动漫', '二次元', '手绘']),
  art('orbital-mecha', '轨道机库 · 地球在身后', '科技机甲', 'wide', '原创石墨灰与象牙白轨道维护机甲，不使用现有高达角色', '面向地球的大型机库，技师作为尺度参照，机械关节结构清晰', '斜向太阳光、磨损金属与空间纵深，工业概念艺术', ['机甲', '科技', '太空']),
  art('snow-leopard', '雪线之上，一次安静的经过', '动物自然', 'wide', '自然比例的成年雪豹', '喜马拉雅雪岭与浅色花岗岩，全身四肢和长尾自然可见', '侧向冬日晨光，远处蓝灰山峰，纪实长焦摄影', ['动物', '风景', '雪豹']),
  art('headphone-ad', '钛金属与一抹橙', '电商产品', 'square', '无品牌钛金属无线头戴耳机', '钴蓝石台，象牙色布料与半透明橙色亚克力作为道具，耳机完整可见', '硬光与精确阴影，可信金属及织物，留白克制', ['电商', '产品摄影', '耳机'])
];


export const originalImagePrompts = {
  "jade-bottle": "Use case: product-mockup. Asset type: original AI demonstration image for a visual prompt library, portrait 4:5. A single unbranded frosted jade-green glass skincare bottle with a brushed metal cap, standing on a pale limestone plinth. One fine leaf and soft water reflection, calm contemporary product photography, large soft side light, gentle natural shadow, accurate glass and condensation, warm gray studio background, restrained sage and ivory palette, bottle fully visible. No typography, logo, watermark, interface or collage.",
  "paper-fox": "Use case: stylized-concept. Asset type: original AI demonstration image for a visual prompt library, square. A charming small fox constructed from folded burnt-orange paper, precise creases, white paper chest, expressive clean silhouette, sitting on a warm pale mint studio floor. Tactile paper, refined three-dimensional craft, soft side light and delicate shadow. A single fox, generous space, no typography, logo, watermark, UI or collage.",
  "rain-portrait": "Use case: photorealistic-natural. Asset type: original AI demonstration image for a visual prompt library, portrait 4:5. Cinematic close portrait of a fictional adult East Asian woman adventurer wearing a simple charcoal waterproof coat, standing on a quiet rainy city street at blue hour. Short dark hair, natural face texture, dignified expression, tiny rain drops, cyan reflections and distant warm amber lights, shallow depth of field, believable photographic grain, no celebrity likeness, no text, no watermark, no UI.",
  "floating-island": "Use case: illustration-story. Asset type: original AI demonstration image for a visual prompt library, landscape 3:2. An expansive dreamlike landscape of floating green islands above an ocean of cloud, one tiny stone observatory and a quiet waterfall on the nearest island, warm dawn sun against pale blue and apricot sky, finely painted atmospheric editorial illustration, rich natural foliage, contemplative rather than epic, crisp foreground and soft distant mountains, no lettering, logo, watermark, interface or collage.",
  "citrus-poster": "Use case: ads-marketing. Asset type: original AI demonstration image for a visual prompt library, portrait 4:5. Refined editorial still life of a sliced blood orange, whole mandarin, curving green leaf and a translucent pale peach glass dish, bold composition near the lower half with generous clean orange-red background at top for future typography. Hard afternoon sun and crisp diagonal shadow, vivid color, tactile photographic surfaces. No actual text, logo, watermark, UI, frames or collage.",
  "coffee-still": "Use case: product-mockup. Asset type: original AI demonstration image for a visual prompt library, square. A handmade celadon ceramic cup filled with dark coffee on a folded oatmeal linen napkin, one croissant edge, a small silver spoon, top-down contemporary café product photograph, natural morning window light, subtle ceramic glaze texture, balanced airy composition with warm neutral background and tiny muted teal accent. No text, logos, watermark, interface or collage.",
  "mini-city": "Use case: stylized-concept. Asset type: original AI demonstration image for a visual prompt library, square. An exquisitely detailed miniature isometric coastal city block on a clean pale blue background, white curved buildings, terracotta roofs, tiny palm trees, one waterfront path, turquoise sea edge, little yellow tram, tasteful contemporary 3D editorial illustration, no labels, physically soft shadows, charming hand-crafted scale, architecture clearly readable. No typography, logos, watermark, UI or collage.",
  "calm-room": "Use case: photorealistic-natural. Asset type: original AI demonstration image for a visual prompt library, landscape 3:2. An uncluttered architect-designed living room, low pale sage fabric sofa, warm oak shelving, curved floor lamp, cream plaster walls, one large leafy plant, huge window showing soft woodland, subtle morning light, natural material texture, editorial interior photography, gentle depth and calm elegant composition, no people, no typography, logo, watermark, UI or collage."
};

const freshSpecs = [
  {
    "id": "festival-friends",
    "title": "汽水色的夏日",
    "category": "年轻人像",
    "ratio": "portrait",
    "prompt": "Create an original photorealistic fashion editorial photograph, vertical 4:5. Two fictional East Asian adult women aged 25 and 27 laughing mid-conversation at an outdoor summer music festival. One wears a cobalt-blue cropped ribbed T-shirt and loose white jeans, the other a coral-red sporty tank top and denim shorts. No nudity. Hair caught naturally in a breeze, one holding a translucent lime-green soda cup. Shot from waist up, candid slightly off-center framing, real pores and tiny flyaway hairs, believable hands, unforced joy rather than a posed AI smile. Out-of-focus cherry-red stage canopy, bright blue sky, subtle sun flare and energetic afternoon daylight. Rich restrained color, independent youth magazine photography, medium-format film texture. Entire subject comfortably in frame. No lettering, no watermark, no famous person, no copied reference face.",
    "subject": "两位25岁以上的原创成年女性，钴蓝短T与珊瑚红背心、牛仔穿搭，自然交谈大笑",
    "scene": "夏日露天音乐节，腰部以上抓拍，红色舞台篷与蓝天，一杯青柠汽水",
    "light": "下午逆光、真实肤质与碎发、独立青年杂志的胶片质感"
  },
  {
    "id": "city-skater",
    "title": "落日之前，再滑一圈",
    "category": "年轻人像",
    "ratio": "portrait",
    "prompt": "Create a completely original photorealistic vertical 4:5 street-fashion photograph. A fictional handsome 26-year-old East Asian adult man, short textured dark hair, relaxed genuine smile, carrying an orange skateboard under one arm while turning towards a friend outside the frame. Oversized cream tee with a small abstract blue wave graphic but no text, faded cobalt utility pants, silver necklace. Waist-up candid composition, sunset on a seaside skatepark, saturated terracotta ramp in the foreground and turquoise sea behind, a little motion in the hair but crisp eyes. Youthful effortless energy, attractive natural features, no plastic skin, believable fingers and anatomy. Editorial film photography with confident sunlight, rich color separation, realistic grain. No celebrities, no brands, no watermark, no lettering.",
    "subject": "26岁的原创成年男生，奶油白T恤与钴蓝工装裤，单手夹着橙色滑板",
    "scene": "海边滑板公园，回头看向画外朋友，腰部以上的随性抓拍",
    "light": "落日侧逆光，青蓝海水与陶土色坡道，自然笑意、真实手部"
  },
  {
    "id": "lilac-cosplay",
    "title": "月兔信使 · 紫色频道",
    "category": "Cosplay",
    "ratio": "portrait",
    "prompt": "Create an original vertical 4:5 premium cosplay fashion editorial photograph of a fictional 25-year-old adult East Asian woman. A playful, confident adult character with a short lilac bob and two small curved pearlescent horn accessories, not an existing anime character. She wears a sophisticated opaque lavender and deep-plum cropped utility jacket, a high-neck fitted white top and high-waisted graphite cargo trousers, fingerless gloves; tasteful fashionable styling, no nudity or lingerie. She reaches one gloved hand towards a small floating translucent pink message orb near her face, with a warm teasing smile. Three-quarter waist-up framing showing her entire hand. Futuristic violet transit lounge, pale mint signage shapes without letters, glowing pink reflections, real photographic fabric texture and convincing soft light on skin. Distinct energetic color, editorial composition, not a generic game render. No watermark, no text, no celebrity likeness.",
    "subject": "25岁的原创成年女性，淡紫短发与珍珠小角，紫色工装短夹克、白色高领内搭、灰黑长裤",
    "scene": "未来交通站，三分之二身像，伸出戴手套的手靠近粉色传信光球",
    "light": "紫与薄荷绿空间、真实织物与柔光，时尚Cosplay摄影，不用游戏塑料质感"
  },
  {
    "id": "dragon-cloud",
    "title": "云海不眠 · 赤金巡游",
    "category": "奇幻风景",
    "ratio": "wide",
    "prompt": "Create an original breathtaking wide 3:2 fantasy illustration, one complete artwork, no panels. A monumental red-and-gold Chinese celestial dragon swimming in the sky through luminous indigo storm clouds above an emerald floating island city. A tiny traveler in an orange cloak stands on a stone observation bridge at lower right for scale. Sweeping diagonal composition, dragon body anatomically coherent, airy negative space, beautiful warm sunrise splitting the cloud bank, luminous sapphire rivers winding through architecture, controlled intricate detail, painterly cinematic key art with vibrant pigments and elegant organic shapes. Original fantasy world, no existing game or animation characters. No text, no lettering, no logos, no watermark.",
    "subject": "原创赤金天龙在云海中巡游，一位橙色披风旅行者作为尺度参照",
    "scene": "青翠浮岛与石桥、远山城池，龙体横贯画面，斜向构图",
    "light": "靛蓝云层与暖日出、蓝宝石河流，精细但留白的电影概念插画"
  },
  {
    "id": "anime-summer",
    "title": "蓝色冰沙与海风",
    "category": "动漫二次元",
    "ratio": "portrait",
    "prompt": "Create an original hand-painted 2D anime illustration, vertical 4:5. A clearly adult 25-year-old female travel illustrator with mint-green hair in a loose ponytail, cheerful bright eyes, wearing an opaque cobalt windbreaker over a coral shirt, sitting beside a seaside fruit-ice stall holding a transparent cup of vivid blue shaved ice. A small curious orange cat perched on the bench beside her, warm summer sunlight, turquoise sea, red parasol, scattered wind-blown postcards with abstract marks not readable letters. Dynamic candid composition, expressive but elegant face, natural hands, brilliant fresh color, visible painterly brush texture and crisp selective linework, premium contemporary illustration. Fully original character, no resemblance to existing anime characters, no school uniform, no sexualization, no nudity. No caption, no watermark, no logo.",
    "subject": "明确25岁的原创成年动漫女旅行插画师，薄荷绿马尾与蓝色风衣，一只橙猫",
    "scene": "海滨冰沙摊，手拿蓝色冰沙，红伞、海风与散落明信片",
    "light": "二维手绘、鲜活蓝红色与温暖阳光、可见笔触和局部清晰线条"
  },
  {
    "id": "candy-orbit",
    "title": "软糖星球 · 一口跳跃",
    "category": "电商产品",
    "ratio": "square",
    "prompt": "Create an original square 1:1 premium playful product advertising still. A clear unbranded resealable pouch filled with glossy jewel-colored fruit gummies floats above a cobalt-blue curved platform. Orange slices, strawberries, violet gummy rings and lime-green translucent candy pieces arc around the pouch in an organized dynamic spiral, not chaotic clutter. Hot coral seamless background with a soft directional shadow, crisp studio hard light, convincing transparent material and appetizing subtle sugar texture, vivid young consumer-brand art direction, bold minimal composition with generous breathing room. Packaging intentionally has no lettering or logo. High-end commercial photography mixed with believable tabletop styling. No text, no watermark, no existing brand.",
    "subject": "无品牌透明软糖包装，橙片、草莓、紫色糖圈和青柠软糖",
    "scene": "钴蓝弧形台座上方悬浮，糖果沿有序螺旋分布，珊瑚红背景",
    "light": "硬光棚拍、真实糖晶与透明塑料，年轻品牌大胆鲜亮配色、留白"
  }
];

const expansionSpecs = [
  {
    "id": "cherry-street",
    "title": "转过街角，夏天刚好",
    "category": "年轻人像",
    "ratio": "portrait",
    "subject": "25岁的原创成年女性，深棕长发、樱桃红针织背心、敞开的白衬衫与浅蓝牛仔裤",
    "scene": "明亮的老城街角，倚着蓝色自行车，抬眼朝画外朋友轻笑；三分之二身抓拍",
    "light": "透明日光、红蓝撞色、真实皮肤与碎发、活泼但不僵硬的青年杂志摄影",
    "prompt": "Create an original photorealistic youth fashion editorial, vertical 4:5. A fictional 25-year-old adult East Asian woman with dark brown softly wavy hair, a warm playful smile and natural skin texture. Cherry-red ribbed tank top, open opaque white cotton shirt, loose light-blue jeans; tasteful fully clothed summer styling. She leans casually beside a powder-blue bicycle at a sunlit old-city corner and turns to look at a friend just off camera, one hand loosely holding the bicycle handle, the other resting naturally at her waist. Three-quarter-length framing, relaxed asymmetric pose, believable fingers, breezy hair and candid energy. Warm yellow stucco, a red-orange café awning without lettering and crisp turquoise sky. Sunlit color separation, slight film grain, authentic editorial photography, not glossy plastic skin or a stock pose. One person only, no text, logos, watermark, UI, celebrity likeness or copied reference face."
  },
  {
    "id": "vinyl-afternoon",
    "title": "B面，还没听完",
    "category": "年轻人像",
    "ratio": "portrait",
    "subject": "26岁的原创成年男性，深色蓬松短发、青绿短袖衬衫与白T恤，自然亲和",
    "scene": "复古唱片店窗边，一只手拿抽象橙色唱片封套，侧身与画外朋友交谈",
    "light": "蜂蜜色窗光、深蓝背景与翠绿衣料，清晰眼神和生活化胶片质感",
    "prompt": "Create a completely original photorealistic vertical 4:5 editorial photograph. A fictional attractive 26-year-old adult East Asian man with tousled dark hair, clear kind eyes and a small spontaneous smile, in a teal-green short-sleeved camp-collar shirt over a white tee and dark denim. Inside a small vintage vinyl record shop, half-turned toward a friend outside the frame, holding one orange record sleeve with purely abstract color shapes and absolutely no lettering. Waist-up framing, casually resting his other hand on a wooden record crate, fingers anatomically believable. Late-afternoon honey sunlight through the window, cobalt-blue shelving softly out of focus, rich teal-orange color harmony, candid sweet approachable energy. Real skin pores, subtle film grain, natural proportions, editorial photography rather than an artificial beauty advertisement. One adult person only. No musician or actor likeness, existing album covers, logos, text, watermark, UI or collage."
  },
  {
    "id": "midnight-editorial",
    "title": "夜幕之后，一点绯红",
    "category": "时尚肖像",
    "ratio": "portrait",
    "subject": "27岁的原创成年女性，黑色丝质吊带、象牙白西装外套与深蓝牛仔裤，优雅自信",
    "scene": "城市屋顶的薄暮时刻，侧身倚靠栏杆，双手自然放低；不是裙装",
    "light": "冷蓝城市天际线与温暖红光，克制性感、完整不透明衣料、真实杂志肖像",
    "prompt": "Create an original premium fashion magazine photograph, vertical 4:5. A fictional 27-year-old adult East Asian woman with long dark hair and a relaxed confident expression, wearing a fully opaque black silk camisole with a modest neckline, an ivory tailored blazer worn casually over the shoulders, and high-waisted indigo jeans. No dress, no skirt, no lingerie, no nudity. Three-quarter portrait on an urban rooftop at blue hour, torso gently angled, forearms resting naturally on a stone railing, both hands low and relaxed, not touching her head. A soft warm scarlet practical light grazes one side of her face against a cool blue city skyline; a hint of wind moves her hair. Sophisticated subtly alluring adult styling, genuine skin texture and believable hands, elegant non-suggestive pose, 85mm editorial photography with shallow depth and fine grain. Keep the whole head and upper hips within frame. No celebrity, copied reference face, lettering, fashion brand, watermark or UI."
  },
  {
    "id": "scarlet-rider",
    "title": "赤色快递员 · 下一站",
    "category": "Cosplay",
    "ratio": "portrait",
    "subject": "25岁的原创成年女性科幻信使，黑色短发，赤红与炭灰骑行夹克及长裤",
    "scene": "未来城市维修站，半身侧坐在停稳的无品牌摩托旁，拿着闭合头盔与银色快递盒",
    "light": "鲜红衣料与青蓝维修灯，电影实景般的磨损、金属与真实肤质",
    "prompt": "Create an original cinematic cosplay fashion photograph, portrait 4:5. A fictional 25-year-old adult East Asian female futuristic motorcycle courier, short black bob, alert friendly eyes and a subtle mischievous smile. Original scarlet-red and charcoal technical motorcycle jacket, opaque fitted black high-neck top, matching full-length armored trousers and boots; no existing game or anime costume, no exposed lingerie. She sits sideways on the edge of a stationary unbranded motorcycle in a futuristic service bay, a closed silver helmet held at her hip and a compact metal delivery case secured behind her. Waist-to-head composition with a little motorcycle visible, hands natural, one foot grounded, no hand near her head. Warm red reflected light contrasts with cyan overhead workshop strips, practical worn metal, small scratches and tactile fabric, shallow cinematic depth. Energetic contemporary character storytelling, realistic youthful adult facial features and skin. No weapon, copied character, celebrity, logos, text, watermark, UI or collage."
  },
  {
    "id": "lotus-sword",
    "title": "莲风起，剑未落",
    "category": "动漫二次元",
    "ratio": "portrait",
    "subject": "25岁的原创成年动漫男性剑修，蓝白修行衣与朱红腰带，长发随风",
    "scene": "夕照莲池石阶，剑收在鞘中，回身望向飞过的灵鸟；有动作但不堆叠特效",
    "light": "鲜明青蓝水面与橙金夕光，精细二维动画线条和丰富环境笔触",
    "prompt": "Create an original high-quality hand-drawn 2D anime key visual, portrait 4:5. A clearly adult 25-year-old male xianxia swordsman with long dark hair loosely tied, gentle confident eyes, wearing flowing sky-blue and ivory cultivation robes with a vermilion sash and practical dark boots. An original character, not from any existing anime or novel. He stands on old stone steps beside a vivid turquoise lotus lake at golden dusk, turning gracefully to watch a small luminous crane fly past; his sword remains sheathed at his waist and both hands are relaxed and anatomically coherent. Robe hems and hair sweep softly in a warm breeze, pink lotus petals drift in the foreground. Three-quarter full-figure composition with ample environment, orange-gold sun behind distant mountain pagodas, confident fine linework, rich painterly background, nuanced cel shading and clear atmospheric perspective. Beautiful youthful adult hero, no childlike body or face. No generic plastic 3D render, giant energy beams, text, logos, watermark, UI or copied characters."
  },
  {
    "id": "tide-mecha",
    "title": "破浪者 · 蓝色航线",
    "category": "科技机甲",
    "ratio": "wide",
    "subject": "原创象牙白与安全橙救援机甲，重型工业关节与透明驾驶舱，不是现有高达",
    "scene": "海岸防波堤旁，机甲跨过浅水区，前景浪花与远处港口形成尺度对比",
    "light": "鲜亮海蓝、白色浪花与夕照橙金，动态但结构清晰的科幻概念画",
    "prompt": "Create an original cinematic science-fiction concept illustration, landscape 3:2. A unique coastal rescue exosuit mech with ivory ceramic armor, safety-orange panels, chunky articulated industrial legs, a small transparent enclosed cockpit and practical hydraulic joints. Not a Gundam, not any existing robot franchise design. The mech strides through shallow azure coastal water beside a long concrete breakwater, water spraying naturally around its grounded feet. Three-quarter low-angle wide view with the entire machine visible and a tiny rescue boat nearby as scale; faraway container-port silhouettes and a brilliant turquoise horizon. Clear balanced mechanical weight, coherent limbs and functional components, no guns or combat. Warm late-afternoon sunlight makes the orange accents glow against saturated cobalt water and white foam. Highly detailed painted concept art, dynamic diagonal composition, tangible material wear, dramatic but readable rather than visually cluttered. No text, serial numbers, logos, watermark, interface or collage."
  },
  {
    "id": "otter-morning",
    "title": "海獭的早餐时间",
    "category": "动物自然",
    "ratio": "square",
    "subject": "一只成年海獭，自然圆润的身体、湿润毛发、真实前爪与小贝壳",
    "scene": "青绿色潮汐池里仰躺漂浮，前爪轻托一枚贝壳，四周水纹与海藻",
    "light": "清透晨光、闪亮水滴与柔和倒影，生动纪实动物摄影而非玩偶",
    "prompt": "Create an original photorealistic wildlife photograph, square 1:1. One adult sea otter floating belly-up in a clear turquoise tidal pool, looking curiously toward the camera, gently holding a small shell between its anatomically correct front paws. Its entire body and hind feet comfortably fit within the frame, dense wet brown fur with silver water droplets, natural rounded face and lively eyes, real animal proportions rather than a cartoon toy. A few olive kelp ribbons curve through the water and subtle ripples spread around the otter. Overhead three-quarter viewpoint, bright transparent morning sunlight, realistic caustics, sparkling highlights, nuanced blue-green water and warm fur contrast. Charming lively documentary animal photography with crisp subject and softly detailed environment. No clothes, human hands, extra animals, anthropomorphic expression, lettering, logo, watermark, UI or collage."
  },
  {
    "id": "coral-sneaker",
    "title": "跑进珊瑚色的夏天",
    "category": "电商产品",
    "ratio": "square",
    "subject": "一只无品牌象牙白跑鞋，青柠色鞋带、浅灰网布与真实鞋底结构",
    "scene": "珊瑚红弧形台阶，跑鞋完整侧向展示，蓝色硬质阴影与少量水滴",
    "light": "鲜明硬光、高级运动广告质感、细致网布和橡胶，不添加品牌标识",
    "prompt": "Create an original high-end sports product advertising photograph, square 1:1. One unbranded contemporary running shoe, ivory-white woven mesh upper, lime-green laces, subtle light-gray structural overlays and a realistic layered off-white rubber sole. No brand logo, no letters, no trademark pattern. Place the single shoe naturally on a sculpted coral-red curved step, full shoe visible in a clean side three-quarter view, believable ground contact and matching shadow. A crisp cobalt-blue shadow cuts diagonally across the background, a few tiny water droplets on the coral surface add freshness without a giant splash. Bold sunlit hard lighting, refined texture detail, confident youth sports color palette, clean editorial advertising composition with breathable negative space. Accurate lacing and shoe structure, not a floating or melted object. No person, second shoe, typography, logo, watermark, UI or collage."
  },
  {
    "id": "aurora-lagoon",
    "title": "极光落在蓝色湖面",
    "category": "奇幻风景",
    "ratio": "wide",
    "subject": "无人的北方冰川湖、黑色玄武岩岸线与通透极光",
    "scene": "低机位广角，前景薄冰与圆石，中景静水倒影，远处积雪山脊",
    "light": "祖母绿极光、靛蓝夜空、柔紫雪线，电影感自然风景与真实尺度",
    "prompt": "Create an original cinematic fine-art landscape photograph, horizontal 3:2. An empty northern glacial lagoon at night, framed by low dark basalt shores and a distant snow-covered mountain ridge. A graceful luminous emerald aurora ribbon flows across the deep indigo sky and reflects naturally in calm icy blue water; subtle violet light touches the snow. Low wide-angle viewpoint with a few dark rounded stones and translucent thin ice in the foreground, strong clear layers of near shore, lake, mountains and expansive sky. Realistically restrained stars, authentic glacial texture and atmospheric depth, rich blue-green-violet color without neon oversaturation. Serene yet visually striking, long-exposure cinematic natural landscape with believable geography, no giant moon, no floating islands, no fantasy buildings or duplicated mountains. No people, text, logos, watermark, UI or collage."
  },
  {
    "id": "lantern-alley",
    "title": "狸猫食堂，今晚营业",
    "category": "场景插画",
    "ratio": "portrait",
    "subject": "原创圆润狸猫厨师，小围裙与真实可爱短肢，迷你拉面屋",
    "scene": "深蓝夜色中的微缩亚洲巷弄，橙色灯笼照亮木窗，蒸汽与小碗形成温暖故事",
    "light": "暖橙与孔雀蓝对比、精细黏土定格材质、可触摸的手作场景",
    "prompt": "Create an original charming miniature stop-motion-inspired 3D illustration, vertical 4:5. A small friendly original tanuki raccoon-dog chef with rounded natural animal features, wearing a plain tiny indigo apron, serving one steaming ramen bowl through the wooden counter of a miniature night-time noodle shop. No existing cartoon character. Full little shop visible in a narrow cozy alley, warm glowing orange paper lanterns without writing, peacock-blue evening shadows, tiny terracotta plant pots and a cream fabric awning, a little drifting noodle steam. Eye-level view with the chef and bowl clearly readable, tactile handmade clay fur and slightly imperfect wood, finely crafted miniature details, warm inviting lively storytelling rather than glossy plastic. Crisp main subject, gentle depth and believable warm light spilling onto blue paving stones. Avoid anthropomorphic human hands, multiple faces, excessive clutter, all lettering, logos, watermark, UI or collage."
  }
];

const recordedImage = spec => ({
  ...art(spec.id, spec.title, spec.category, spec.ratio, spec.subject, spec.scene, spec.light, [spec.category, '鲜活色彩', '原创']),
  rawPrompt: spec.prompt, rawPromptLabel: '这张图实际使用的原始提示词',
  styles: spec.style ? [spec.style] : undefined,
  generation: { tool: '内置 image_gen', date: '2026-10-01', framing: spec.ratio === 'wide' ? '横幅 3:2' : spec.ratio === 'square' ? '正方 1:1' : '竖幅 4:5', notes: '未提供参考图；未设置 seed；复用模板另写，未逐条重跑。' }
});
export const freshImages = freshSpecs.map(recordedImage);
// 用户移除的两张保留历史生成记录，但不再进入目录／旧链接。
export const expansionImages = expansionSpecs.filter(spec => !['midnight-editorial', 'coral-sneaker'].includes(spec.id)).map(recordedImage);
const curatedSpecs = [
  {
    "id": "brass-cartographer",
    "title": "黄铜地图师 · 沙海来信",
    "category": "Cosplay",
    "ratio": "portrait",
    "styles": [
      "写实摄影"
    ],
    "subject": "28岁的原创成年女性地图师，深棕短卷发、护目镜与黄铜罗盘",
    "scene": "沙漠驿站工坊，墨绿不透明工装与米色围巾，手握展开的无字星图，三分之二身像",
    "light": "沙金窗光、真实皮革与金属磨损，叙事型Cosplay摄影",
    "prompt": "Use case: photorealistic-natural. Create one original vertical 4:5 premium cosplay editorial photograph. A fictional 28-year-old adult woman cartographer with short dark curls, brass goggles pushed above her forehead, an opaque deep forest-green tailored utility jacket and a pale linen scarf. She holds an unfolded cream star map showing purely abstract constellation dots and a brass compass, natural anatomically believable hands. Three-quarter-length composition inside a desert waystation workshop: warm sand outside the arched window, wooden drawers and a single globe softly blurred behind. Curious focused expression, distinctive layered travel costume, tactile worn leather and brushed brass, golden late-afternoon window light, photographic skin texture. Original character, no existing franchise costume or celebrity likeness. No lettering, logos, watermark, interface, collage, nudity or lingerie."
  },
  {
    "id": "frost-musician",
    "title": "霜蓝奏者 · 冬夜序曲",
    "category": "Cosplay",
    "ratio": "portrait",
    "styles": [
      "写实摄影",
      "电影感"
    ],
    "subject": "29岁的原创成年男性魔法音乐家，深色发、霜蓝长外套与银色肩饰",
    "scene": "雪夜石造音乐厅台阶，完整提琴和琴弓自然握持，半身至膝的站姿",
    "light": "冰蓝雪光与室内琥珀窗光、真实绣线和羊毛，克制电影摄影",
    "prompt": "Use case: photorealistic-natural. Create an original vertical 4:5 cinematic cosplay photograph of a fictional 29-year-old adult male fantasy musician. Dark wavy hair, calm confident face, an opaque frost-blue long wool coat with restrained silver embroidered shoulder details and dark trousers. He stands naturally on the stone steps of a snowy concert hall holding one complete wooden violin by its neck and a bow lowered in the other hand; all fingers believable, no playing pose. Frame head to knees with comfortable headroom. Light snowfall, antique carved stone, amber windows against blue winter dusk, precise woven wool and wood grain, realistic skin and elegant theatrical costume photography. A wholly original character, no existing game or anime identity. No lettering, watermark, logos, UI, collage, nudity or weapons."
  },
  {
    "id": "red-panda-moss",
    "title": "苔藓树梢，一位小访客",
    "category": "动物自然",
    "ratio": "wide",
    "styles": [
      "写实摄影"
    ],
    "subject": "自然体型与毛发的成年小熊猫",
    "scene": "湿润山地森林，完整小熊猫沿覆苔树枝侧向行走，尾巴清楚可见",
    "light": "雨后漫射晨光、细密红棕毛发与蕨叶，纪实野生动物摄影",
    "prompt": "Use case: photorealistic-natural. Create one original landscape 3:2 wildlife photograph of an adult red panda walking along a moss-covered branch in a humid mountain forest. Full animal clearly visible including all four paws and its long ringed tail, natural anatomy, alert gentle expression, wet fern leaves and layered dark-green forest softly out of focus. Off-center horizontal composition with breathing space in the walking direction. Soft diffuse morning light after rain, extremely tactile reddish fur, fine whiskers, tiny dew beads, authentic long-lens natural-history photography. No human objects, anthropomorphic clothes, extra limbs, text, logo, watermark, interface or collage."
  },
  {
    "id": "turtle-blue",
    "title": "蓝海航行 · 珊瑚之上",
    "category": "动物自然",
    "ratio": "wide",
    "styles": [
      "写实摄影"
    ],
    "subject": "完整成年海龟，清楚的龟甲和自然鳍肢",
    "scene": "海龟在多彩珊瑚礁上方游过，斜向构图，远方鱼群作为尺度",
    "light": "清澈蓝海与真实日光束，水下摄影，不夸张荧光",
    "prompt": "Use case: photorealistic-natural. Create an original landscape 3:2 underwater wildlife photograph. A complete adult sea turtle glides above a colorful living coral reef, its shell and naturally posed flippers clearly visible, anatomically coherent, small distant reef fish for scale. Diagonal composition, clear deep turquoise water fading to cobalt blue, real sunlight beams and caustic light on the turtle's shell, soft suspended particles, rich but believable orange and violet corals. Premium natural-history underwater photography with precise textures and an unhurried mood. No diver, plastic, text, logo, watermark, UI, collage, exaggerated neon or additional limbs."
  },
  {
    "id": "jazz-cutout",
    "title": "午夜爵士 · 蓝黄节拍",
    "category": "海报社媒",
    "ratio": "portrait",
    "styles": [
      "平面海报"
    ],
    "subject": "抽象黄铜小号剪影与太阳黄色圆盘",
    "scene": "钴蓝竖版海报，奶油白剪纸曲线形成节奏，主体落在下半部，上方留排版空位",
    "light": "粗纹纸与丝网油墨、平面拼贴构图、清楚的有限色板",
    "prompt": "Use case: ads-marketing. Create a single original vertical 4:5 art-poster background for a jazz social campaign, WITHOUT any typography. A large elegant abstract brass trumpet silhouette curves through the lower half, a bold warm-yellow circular disk behind it, two cream-white paper-cut rhythm shapes on an intense cobalt-blue field. Asymmetric flat graphic composition, restrained three-color palette, real fibrous cut-paper edges and subtle screen-print ink texture. Strong negative space across the upper third reserved for a future headline, editorial art direction, simple confident shapes rather than 3D objects. No words, letters, numbers, brand, watermark, border, interface or collage of multiple posters."
  },
  {
    "id": "blue-botanical",
    "title": "蓝花实验 · 一页春意",
    "category": "海报社媒",
    "ratio": "portrait",
    "styles": [
      "平面海报",
      "写实摄影"
    ],
    "subject": "一朵蓝紫色鸢尾花，完整茎叶与清晰花瓣",
    "scene": "浅青柠绿平面背景，低位斜向构图，花瓣投出清楚阴影，顶部留白",
    "light": "明亮侧光、细腻花瓣纹理与鲜明蓝绿色分离，实验植物海报摄影",
    "prompt": "Use case: ads-marketing. Create one original vertical 4:5 contemporary botanical campaign image with no text. One sculptural blue-violet iris flower and two clean green leaves, complete natural stem, arranged diagonally across the lower two thirds of a pale lime-green seamless paper background. Strong side sunlight produces a crisp intentional botanical shadow, exquisite real petal folds and translucent edges, vivid controlled color separation, premium experimental editorial still-life photography. Generous untouched negative space at the top for future typography. Minimal prop-free composition. No vase, lettering, logo, watermark, UI, frame, multiple panels or artificial plastic texture."
  },
  {
    "id": "anime-rain-tram",
    "title": "雨停以后，电车还在",
    "category": "动漫二次元",
    "ratio": "wide",
    "styles": [
      "二次元"
    ],
    "subject": "27岁的原创成年女性摄影师，黑色短发与芥末黄雨衣",
    "scene": "雨后的旧街电车站，背影微侧，手拿小相机，绿色电车和湿街暖灯",
    "light": "二维手绘线条、细腻水粉背景、雨夜青绿与琥珀色对比",
    "prompt": "Use case: illustration-story. Create one original landscape 3:2 hand-painted 2D anime artwork. A clearly adult 27-year-old female street photographer with a short black bob, an opaque mustard-yellow raincoat and dark trousers waits beside a small green vintage tram after rain. She stands in three-quarter back view holding a compact camera naturally at chest level. Atmospheric old-city lane, warm amber shop windows with no readable signage, wet blue-green pavement reflecting light, quiet overhead wires and a tiny puddle in the foreground. Carefully drawn perspective, selective crisp linework, textured gouache-painted background and a thoughtful cinematic composition. Wholly original adult character and city, no school uniform, copied anime character, sexualization, lettering, watermark, UI or collage."
  },
  {
    "id": "anime-sky-mechanic",
    "title": "风艇修理铺 · 云上的午后",
    "category": "动漫二次元",
    "ratio": "portrait",
    "styles": [
      "二次元"
    ],
    "subject": "28岁的原创成年男性空艇维修师，栗色卷发、米色工作服与蓝绿背带裤",
    "scene": "悬空木平台上的修理铺，手持扳手，橙色小空艇在背景，云海与风中旗帜",
    "light": "明快二维动画、可见手绘笔触、柔和下午阳光与细致机械结构",
    "prompt": "Use case: illustration-story. Create an original vertical 4:5 hand-painted 2D anime illustration. A clearly adult 28-year-old male airship mechanic with chestnut curls and a warm relaxed grin wears an opaque cream work shirt and teal overalls. He stands on a weathered wooden platform above a sea of clouds, holding a single wrench at his side with believable fingers. Behind him a small original orange sail-powered flying boat rests beside an open repair shed, coherent rigging and simple brass engine details, fabric pennants flutter gently. Bright afternoon sun, painterly cloud atmosphere, selective expressive linework, visible hand-painted texture, lively but uncluttered composition. Entire head and hands comfortably in frame. No existing anime character, child, text, logo, watermark, interface or collage."
  },
  {
    "id": "sunken-lounge",
    "title": "落一阶，坐进暖光里",
    "category": "空间设计",
    "ratio": "wide",
    "styles": [
      "写实摄影"
    ],
    "subject": "下沉式客厅，陶土布艺弧形沙发、浅石灰岩与胡桃木",
    "scene": "沙漠现代住宅，宽横幅室内视角，两级台阶进入会客区，大窗外淡色庭院",
    "light": "斜向午后日光、真实织物与木材、连贯空间透视，建筑杂志摄影",
    "prompt": "Use case: photorealistic-natural. Create one original landscape 3:2 architectural interior editorial photograph of a sophisticated sunken conversation lounge in a contemporary desert house. A low curved terracotta fabric sofa arranged around a small warm limestone table, two clearly usable wide steps down into the seating area, walnut built-in shelving, cream lime-plaster walls, a tall window overlooking a pale gravel courtyard with one olive tree. Coherent perspective and realistic room proportions, comfortable circulation, sunlight slicing diagonally across the floor, believable fabric weave, warm stone and wood grain. Refined lived-in minimalism, no generic clutter or extravagant luxury props. No people, text, logos, watermark, UI, frame or collage."
  },
  {
    "id": "lilac-bookshop",
    "title": "转角书屋 · 浅紫的安静",
    "category": "空间设计",
    "ratio": "wide",
    "styles": [
      "写实摄影"
    ],
    "subject": "独立书店，浅紫拱形阅览区、橡木书架与软木地面",
    "scene": "小型街角书屋，书架转角与圆桌阅读位，完整通道，书脊无可读文字",
    "light": "柔和天窗自然光、可信建筑尺度与材质，空间编辑摄影",
    "prompt": "Use case: photorealistic-natural. Create an original landscape 3:2 architectural editorial photograph of a small distinctive independent bookshop. Pale lilac plaster arches frame a cozy reading alcove, honey-oak floor-to-ceiling bookshelves filled with books whose spines have only abstract color blocks and no readable text, cork flooring, one simple round oak reading table with two cream upholstered chairs. A clear generous walking path from the foreground to a street-corner window, coherent architectural perspective, human-scale dimensions, soft daylight from a narrow skylight and window, tactile plaster, cork and wood. Quiet inviting design, thoughtful color rather than clutter. No people, signage, lettering, logos, watermark, UI, collage or multiple rooms."
  }
];
export const curatedImages = curatedSpecs.map(spec => ({ ...recordedImage(spec), styles: spec.styles }));
export const homepageImage = {
  ...recordedImage({"id":"window-portrait","title":"窗边片刻 · 灰蓝条纹","category":"年轻人像","ratio":"portrait","styles":["写实摄影"],"subject":"28岁的虚构成年女性，深棕长发，白色吊带、灰蓝细条纹衬衫与牛仔裤","scene":"灰米色窗边，端正坐姿，双手自然放在腿上，头顶至大腿中部","light":"柔和窗光与真实肤质、自然灰米色室内摄影","prompt":"Use case: photorealistic-natural. Create one clean 4:5 vertical editorial portrait for a website gallery, using the attached image only as a reference for dark hairstyle, facial aesthetic, quiet expression, soft indoor light and muted neutral palette. A fictional adult 28-year-old East Asian woman with long loosely tousled dark brown hair and delicate natural makeup sits comfortably beside a window, looking calmly at the camera with her head slightly tilted. A beautiful natural face with realistic skin texture, restrained expression, not a large smile. Ordinary fully clothed casual styling: ivory camisole, loose very fine gray-blue striped cotton shirt worn normally over it, and blue denim jeans. The outfit consists of separate top and trousers, never a skirt or dress. Her hands rest naturally in her lap; no hand touching head or hair. Relaxed seated upright pose, both legs down. Frame from just above the whole head to mid thighs; face prominent, comfortable headroom. Quiet gray-beige plaster wall, subtle diffused window light, creamy neutral tones and gentle shadows similar to the reference. Natural magazine photography, not an illustration, not glossy airbrushed skin, no green tint, no garden or blonde hair. Remove all screenshot artifacts: no black frame, rounded corners, UI, dimension badge, lettering, logos or watermark. Single clean photograph, no collage."}),
  styles: ['写实摄影'],
  generation: { tool: '内置 image_gen', date: '2026-09-30', framing: '竖幅 4:5', notes: '复用首页既有图例；原生成使用用户提供的发型、神态与室内光线参考，参考权属未核实。人物为虚构成年人；复用模板未重跑。' },
  requirements: '复用已有首页 AI 图例与真实生成记录。原生成使用用户视觉参考，参考权属未核实；不代表真人身份或代言。可替换模板未实测。'
};
export const diverseImages = [
  {
    "id": "pixel-night",
    "title": "像素港口 · 夜航开始",
    "category": "场景插画",
    "ratio": "wide",
    "style": "像素",
    "subject": "原创像素港口与小型青绿渡船，少量成年旅人",
    "scene": "横向游戏场景，深蓝夜空、亮紫码头、橘黄窗灯和水面倒影，清晰8位像素块",
    "light": "严格低分辨率像素美术，用有限色板与抖动层次，不用写实渲染",
    "prompt": "Create an original landscape 3:2 pixel-art game environment, not a photograph or 3D render. A lively small midnight harbor with a mint-green ferry at the dock, a compact violet ticket hut, warm amber window lights and two tiny clearly adult travelers with luggage. Strict deliberate square pixel clusters, a limited 24-color palette, crisp stair-stepped edges, hand-placed dithering, no smooth gradients, no anti-aliased painting. Deep indigo sky, plum dock structures, cyan water reflections and orange lights with strong readable silhouettes. Side-on scene like a lovingly crafted 1990s adventure-game background, rich atmosphere and readable space without excessive clutter. Original environment, no existing game sprites, no text, UI, logos or watermark."
  },
  {
    "id": "ink-koi",
    "title": "墨色山河 · 一尾朱砂",
    "category": "奇幻风景",
    "ratio": "portrait",
    "style": "水墨",
    "subject": "原创水墨巨鲤与烟云山河，朱砂色鱼尾作为唯一鲜艳重点",
    "scene": "竖向宣纸留白，山水横断面里一尾鲤鱼盘旋，墨色从浓到淡渐隐",
    "light": "真实毛笔飞白、墨晕与宣纸纤维，传统水墨与现代超现实构图",
    "prompt": "Create a completely original vertical 4:5 contemporary Chinese ink-wash artwork on tactile off-white xuan paper. An enormous graceful koi swims through layered misty ink mountains as if the river and sky were one. Charcoal-black and dilute gray brushwork, confident dry-brush streaks, feathered ink bleeding and large intentional negative spaces. A single restrained vermilion-red koi tail is the only saturated color accent. The composition is asymmetrical and poetic, with crisp near rocks dissolving into pale distant peaks. The fish anatomy and flowing movement remain elegant and readable. Authentic hand-painted ink aesthetics, paper fibers, no photorealism, no glossy 3D, no computer neon or generic fantasy effects. No calligraphy, seals, signatures, text, watermark, UI or copied historic painting."
  },
  {
    "id": "comic-hero",
    "title": "美漫封面 · 橙色闪电",
    "category": "Cosplay",
    "ratio": "portrait",
    "style": "美漫",
    "subject": "25岁的原创成年女性城市英雄，青绿短发、橙黑骑行夹克与工装长裤",
    "scene": "都市天台跃步瞬间，半身到膝盖的动态透视，斜向高楼剪影与留白",
    "light": "粗细分明墨线、网点印刷与大胆橙蓝撞色，经典美漫插画非写实人像",
    "prompt": "Create an original vertical 4:5 American-comic cover illustration with no lettering. A clearly adult 25-year-old fictional female urban messenger hero with short teal hair, an orange-and-black fitted motorcycle jacket, opaque black top and full-length graphite cargo trousers. She lands lightly from a short rooftop step, one boot forward, confident delighted expression and natural readable hands, a small delivery satchel swinging at her side. Powerful diagonal city silhouettes behind her, cobalt sky and tangerine highlights, bold varied black ink outlines, graphic spot-black shadows, visible offset-print halftone dots and hand-painted color separations. Energetic foreshortening, a coherent adult body and distinctive original costume; do not imitate any existing superhero or artist. Not a photograph, anime, 3D render or glossy beauty image. No guns, logos, text, speech bubbles, watermark or UI."
  },
  {
    "id": "watercolor-island",
    "title": "水彩旅行 · 海湾慢一天",
    "category": "奇幻风景",
    "ratio": "wide",
    "style": "水彩",
    "subject": "原创蓝白海湾小镇、彩色帆船与岸边花园",
    "scene": "横向旅行画册构图，俯看弯曲海湾，宽阔浅蓝水面、散落房屋与橙色船帆",
    "light": "透明水彩层染、冷压纸纹理和轻巧铅笔细节，清新不写实",
    "prompt": "Create an original horizontal 3:2 watercolor travel-journal illustration, not a photograph. An elevated view over a bright Mediterranean-inspired bay with small white-and-cobalt houses, a curving stone waterfront, bougainvillea gardens and three tiny sailboats with orange and yellow sails. Broad translucent turquoise watercolor washes, pale blue sea, fresh lemon sunlight, soft pigment blooms, visible cold-pressed paper grain and a few delicate graphite architectural lines. Buildings and boats are selectively detailed while large sea areas breathe; charming gentle irregular hand-painted edges rather than precise 3D geometry. Lively summer atmosphere with confident colorful accents and rich but airy composition. No readable signs, text, logos, watermark, UI, collage or imitation of a specific painter."
  },
  {
    "id": "papercut-tiger",
    "title": "剪纸小虎 · 春山漫游",
    "category": "动物自然",
    "ratio": "square",
    "style": "剪纸",
    "subject": "原创橙红小虎穿过青绿色叶片，纯手工剪纸造型",
    "scene": "方形层叠纸艺，老虎完整侧身可见，几层蓝绿山形与花叶形成空间",
    "light": "厚纸边缘、剪刀切口和层间真实阴影，现代民艺色彩，不用毛绒写实",
    "prompt": "Create an original square 1:1 handcrafted layered paper-cut artwork. A charming orange-red tiger with cream paper cheeks and bold charcoal paper stripes walks through teal and jade foliage, its full body, four paws and curved tail clearly readable. The tiger is a stylized animal, not a human or existing mascot. All forms are physically cut from thick matte colored paper, with slightly irregular scissor edges, shallow stacked layers and real directional shadows between layers. A few cobalt mountain shapes and small apricot blossoms form a balanced lively spring composition. Crisp graphic silhouettes, tactile paper fibers, contemporary folk-art color design. No furry photorealism, plastic 3D, gradients, lettering, logo, watermark, UI or frame."
  },
  {
    "id": "retro-observatory",
    "title": "复古未来 · 火星日落站",
    "category": "海报社媒",
    "ratio": "portrait",
    "style": "复古未来",
    "subject": "原创火星观测站、圆形天线与远处橙色星球，纯图形复古科幻海报",
    "scene": "竖向几何平面构成，深蓝剪影、杏黄沙丘、红橙星空与留白",
    "light": "丝网印刷色块、有限套色与细微纸纹，70年代科幻视觉而非电影剧照",
    "prompt": "Create an original vertical 4:5 retro-futurist science-fiction travel poster artwork with absolutely no text. A solitary elegant Mars observatory with a curved antenna and a small dome stands on layered apricot dunes. A giant burnt-orange planet rises beyond the indigo horizon while one tiny ivory rover follows a clean curved track. Flat confident geometric silhouettes, dramatic diagonal framing, a limited screen-print palette of midnight blue, vermilion, mustard and cream, subtle authentic printed-paper texture and slight color registration imperfections. Inspired by the general visual language of 1970s space-age graphic design, without copying any existing poster or artist. Sophisticated editorial composition, not a photorealistic film still, not glossy 3D or a generic futuristic interface. No typography, logos, watermark, UI or collage."
  }
].map(recordedImage);

const videoSpecs = [
  {
    "id": "coast-reunion",
    "title": "海岸重逢 · 迟到的答案",
    "category": "美剧感",
    "tone": "浪漫悬念",
    "cast": "25岁的短栗色头发成年女性，蓝灰开衫；27岁的成年男性，深棕夹克，两人是原创角色",
    "scene": "黄昏后的海边木栈道，风吹动灯串，远处是深蓝海面",
    "shots": [
      [
        "0–3秒",
        "女主独自靠栏望海，听到身后脚步，先抬眼再转身。",
        "50mm中近景，肩侧缓慢推近，冷海光与暖灯串分开。"
      ],
      [
        "3–7秒",
        "男主停在两步之外，递出一张折好的渡船票；女主看到票上手写的空白位置，惊讶逐渐变成微笑。",
        "反打近景接手部特写，保持同一180度轴线，浅景深不遮挡眼神。"
      ],
      [
        "7–12秒",
        "她接过票，轻声问‘你还记得？’，他点头，两人一起转向海面，留下一秒停顿。",
        "双人侧面中景缓慢拉远，风声与脚步自然，避免夸张笑容。"
      ]
    ],
    "en": "An original adult romance on a seaside pier at blue hour. 0–3s: a 25-year-old woman turns on hearing footsteps, a slow 50mm shoulder push. 3–7s: a 27-year-old man offers a folded ferry ticket, reverse close-up then hand insert on one axis. 7–12s: she softly asks whether he remembers; he nods, they face the sea in a gentle pull-back. Natural pauses, warm string lights against cool sea light."
  },
  {
    "id": "elevator-secret",
    "title": "电梯停在十三层",
    "category": "美剧感",
    "tone": "都市悬疑",
    "cast": "26岁的成年女建筑师，黑色短外套；28岁的成年男保安，浅灰制服，无品牌",
    "scene": "深夜现代写字楼，电梯外有一扇亮着暖光但本应空置的办公室",
    "shots": [
      [
        "0–4秒",
        "女主看一眼手机时间，电梯门意外打开，她没有立刻迈出。",
        "从电梯内部固定中景，办公室暖光落在冷色走廊。"
      ],
      [
        "4–8秒",
        "远处男主抬手示意安静，女主听见办公室内玻璃杯被放下的声音。",
        "慢速侧移露出男主，切女主眼神特写，音效只做一个明确线索。"
      ],
      [
        "8–12秒",
        "办公室门柄缓缓转动；她后退半步，电梯门在两人视线之间合拢。",
        "门柄特写接电梯内近景，压住剪辑节奏，不展示怪物或血腥。"
      ]
    ],
    "en": "Original restrained urban mystery with adults. 0–4s: a 26-year-old architect hesitates as an elevator opens onto a deserted office floor. Fixed medium shot, amber office light in a cool hallway. 4–8s: a 28-year-old guard gestures for silence; a single glass clink is heard, lateral reveal then eye close-up. 8–12s: a door handle turns; she retreats and elevator doors close. No gore, no monster reveal, coherent geography."
  },
  {
    "id": "storm-letter",
    "title": "暴风来信 · 没说出口的话",
    "category": "美剧感",
    "tone": "情绪短剧",
    "cast": "25岁的成年红棕发女性，酒红针织上衣；27岁的成年深发男性，米白衬衫",
    "scene": "临海小屋厨房，蓝色暴风天窗光，桌上有一封未寄出的信",
    "shots": [
      [
        "0–3秒",
        "女主将信放到桌上，手停了一瞬，镜头不拍出信中文字。",
        "桌面特写，湿润窗光与一盏小暖灯形成自然冷暖。"
      ],
      [
        "3–8秒",
        "男主看见信，停下倒茶动作；两人互相看着，谁也不先伸手。",
        "双人中景，再切女主眼神，保持呼吸和真实紧张，不急推。"
      ],
      [
        "8–12秒",
        "他把茶杯推到她面前，她慢慢坐下；远处一声闷雷，气氛由僵硬变柔和。",
        "杯子轻响接双人侧面近景，静止收尾，不用拥吻或夸张煽情。"
      ]
    ],
    "en": "An original adult emotional drama in a seaside kitchen before a storm. 0–3s: a woman places an unreadable letter on the table, close-up with cool window light and one warm lamp. 3–8s: a man pauses while pouring tea; a quiet two-shot then her eyes, believable hesitation. 8–12s: he slides her a cup and she sits; a distant low thunder, static side close-up. Subtle performance, no celebrity likeness."
  },
  {
    "id": "coastal-choice",
    "title": "沿海公路 · 留下还是出发",
    "category": "美剧感",
    "tone": "公路青春",
    "cast": "26岁的成年男摄影师，褪色牛仔夹克；25岁的成年女音乐人，橄榄绿短外套",
    "scene": "晴天海岸公路停车点，一辆无品牌旧旅行车，远处海浪",
    "shots": [
      [
        "0–4秒",
        "男主将相机放进车后备箱，女主站在车旁看着一枚拨片。",
        "低位广角建立车、人物与海的关系，逆光保留脸部细节。"
      ],
      [
        "4–8秒",
        "女主喊住他，把拨片抛给他；他自然接住，先意外后笑。",
        "中景动作完整，切接住后的表情，不做危险驾驶。"
      ],
      [
        "8–12秒",
        "两人坐进停着的车，窗外海风吹动地图；车仍未发动，故事停在选择之前。",
        "从后排缓慢靠近两人肩线，明快日光，无车祸或高速追逐。"
      ]
    ],
    "en": "An original sunny road-trip scene with two adults. 0–4s: a 26-year-old photographer stores his camera beside a parked unbranded wagon; a 25-year-old musician holds a guitar pick. Low wide shot with the sea. 4–8s: she tosses him the pick, he catches it and smiles, one clear medium action. 8–12s: both sit in the still parked car, wind lifts a map, a slow back-seat push. No dangerous driving."
  },
  {
    "id": "shared-umbrella",
    "title": "雨停之前 · 同一把伞",
    "category": "韩剧感",
    "tone": "甜感相遇",
    "cast": "25岁的成年女插画师，浅黄针织衫；27岁的成年男花店店主，海军蓝长外套",
    "scene": "雨后的首尔风格原创街角，薄荷色花店，路面积水映出暖光",
    "shots": [
      [
        "0–3秒",
        "女主在门檐下躲雨，怀里护着速写本；男主打开一把透明伞。",
        "眼平中景，真实细雨，不用滤镜磨皮。"
      ],
      [
        "3–7秒",
        "他将伞向她那边倾斜，她发现他肩膀淋湿，伸手把伞扶回两人中间。",
        "手部小特写后接双人近景，视线自然交错，动作克制。"
      ],
      [
        "7–12秒",
        "两人笑了一下并肩迈出门檐，脚踩过浅浅水洼，花店暖光留在身后。",
        "背侧缓慢跟拍，不绕人物大幅旋转，细雨与脚步声。"
      ]
    ],
    "en": "Original sweet adult romance on a fictional Seoul-inspired street. 0–3s: a 25-year-old illustrator shelters a sketchbook as a 27-year-old florist opens a transparent umbrella. Eye-level medium shot, real rain. 3–7s: he tilts the umbrella toward her; she centers it over both, hand insert then two-shot. 7–12s: a shared smile and a gentle walk across a shallow puddle, slow rear-side tracking, warm florist light behind."
  },
  {
    "id": "rooftop-soda",
    "title": "屋顶上的橘子汽水",
    "category": "韩剧感",
    "tone": "明快日常",
    "cast": "25岁的成年女设计师，天蓝卫衣；26岁的成年男厨师，奶油白T恤，原创人物",
    "scene": "傍晚城市屋顶，几盆绿植，一张折叠桌，两瓶橘子汽水",
    "shots": [
      [
        "0–4秒",
        "男主拧开汽水，泡沫突然溢出，他手忙脚乱找纸巾；女主被逗笑。",
        "双人中景，保留动作因果，橙色汽水与蓝天形成鲜活色彩。"
      ],
      [
        "4–8秒",
        "她递纸巾，他故作镇定地抹去手上的泡沫，两人笑着看向城市。",
        "反打近景，控制手部数量，真实自然笑意。"
      ],
      [
        "8–12秒",
        "两瓶汽水轻碰，夕阳映在瓶身，猫从桌子下面走过。",
        "手部特写后拉回两人，轻松氛围，不加商业标志。"
      ]
    ],
    "en": "Original lighthearted adult rooftop scene. 0–4s: a 26-year-old chef opens orange soda which overflows; a 25-year-old designer laughs. Medium two-shot in fresh orange and blue dusk colors. 4–8s: she passes a tissue, he regains composure, natural reverse close-ups. 8–12s: they softly clink the unbranded bottles and a cat walks under the table; insert then wider shot, believable comedy timing."
  },
  {
    "id": "winter-second-cup",
    "title": "冬天的第二杯",
    "category": "韩剧感",
    "tone": "温柔治愈",
    "cast": "26岁的成年女书店主，莓红围巾；27岁的成年男翻译，灰绿毛衣",
    "scene": "飘雪街道旁的原创小咖啡馆，木窗框与暖色台灯，没有真实商标",
    "shots": [
      [
        "0–3秒",
        "女主将一杯热咖啡放在对面空位，看了看门口。",
        "侧面近景，杯口薄蒸汽，窗外蓝白雪色。"
      ],
      [
        "3–7秒",
        "门铃响，男主带着雪花进来，看见第二杯咖啡，放松地笑。",
        "门口中景接眼神反打，不做慢动作拉长每个动作。"
      ],
      [
        "7–12秒",
        "他坐下，把一本包好的书放到桌上；她将咖啡推近，两人安静开始交谈。",
        "稳定双人中近景，门铃与杯声轻响，自然口型，字幕另做。"
      ]
    ],
    "en": "Original cozy winter adult romance. 0–3s: a 26-year-old bookseller puts a second hot coffee opposite her, side close-up with delicate steam and blue snow outside. 3–7s: a 27-year-old translator enters with snow on his coat, notices the cup and smiles, doorway medium then reverse close-up. 7–12s: he offers a wrapped book, she slides the coffee closer and they talk quietly. Stable two-shot, warm wood and practical lamp light."
  },
  {
    "id": "cloud-sword",
    "title": "云上试剑 · 第一声风",
    "category": "修仙动漫",
    "tone": "国风热血",
    "cast": "25岁的原创成年女剑修，青白衣袍与墨色腰带，无现有动漫角色元素",
    "scene": "悬于云海之上的古石平台，远处翠色山峰，金色晨光",
    "shots": [
      [
        "0–3秒",
        "剑修站定，呼气，脚边落叶轻轻升起，剑仍在鞘中。",
        "低位中景，手绘2D国风动画，线条清楚、衣褶有重量。"
      ],
      [
        "3–7秒",
        "她拔剑向前划出一道弧线，风带动一圈落叶，剑光是细薄青光而非满屏爆炸。",
        "动作中景，关键帧明确，单一方向运镜，不切碎剑招。"
      ],
      [
        "7–12秒",
        "叶片落下，远处云海分开一线；她收剑回头，嘴角露出自信笑意。",
        "从手部到半身的慢抬镜，保留云海层次，动作有起承转合。"
      ]
    ],
    "en": "Original hand-drawn 2D cultivation animation with a 25-year-old adult female sword cultivator in white and teal robes. 0–3s: she exhales on a stone platform above clouds; leaves lift while her sword is sheathed, low medium. 3–7s: one clear draw and arc sends a thin teal gust through leaves, coherent key poses and one camera direction. 7–12s: clouds part slightly, she sheathes the blade and turns with a confident smile, slow tilt to half-body. No existing anime character."
  },
  {
    "id": "spirit-fox",
    "title": "灵狐归来 · 灯火认得你",
    "category": "修仙动漫",
    "tone": "奇幻治愈",
    "cast": "26岁的原创成年男药师，靛蓝长袍；一只原创银白灵狐，尾部有温柔蓝光",
    "scene": "雨后山城药铺门前，红色灯笼、湿石阶与青苔",
    "shots": [
      [
        "0–4秒",
        "药师正要关门，看见石阶下有一串微弱蓝光脚印。",
        "远中景推近，2D水彩动画质感，暖灯与蓝光清晰分区。"
      ],
      [
        "4–8秒",
        "灵狐从石阶转角走出，抖落雨水，药师蹲下伸出手掌；狐狸轻碰他的手。",
        "平视中景保持狐狸四肢结构，再给轻触特写，不变成真人。"
      ],
      [
        "8–12秒",
        "蓝光沿灯笼绳索升起，药铺亮起一排灯，药师与狐狸一同抬头。",
        "缓慢仰移到两者与灯笼，光效克制，温暖而不是炫技闪屏。"
      ]
    ],
    "en": "Original watercolor-textured 2D fantasy with a 26-year-old adult herbalist and a silver spirit fox. 0–4s: he sees faint blue pawprints on rain-wet stone steps, medium push, red lanterns against blue magic. 4–8s: the fox rounds a corner, shakes off rain, then gently touches his offered palm, stable medium and one contact insert, coherent paws. 8–12s: blue light climbs a lantern cord and the shop lights up as both look upward, slow tilt, restrained warm magic."
  },
  {
    "id": "celestial-gate",
    "title": "天门开启 · 不退的一步",
    "category": "修仙动漫",
    "tone": "史诗冒险",
    "cast": "27岁的原创成年男修士，赭红衣袍；25岁的原创成年女阵师，深蓝衣袍，两人并肩",
    "scene": "暴风中的山巅，一座古老石门浮现蓝金阵纹",
    "shots": [
      [
        "0–3秒",
        "两人在风中抬头，石门的一道阵纹点亮，脚下碎石轻微震动。",
        "大远景确立巨大尺度，2D动画概念美术，风与衣摆方向一致。"
      ],
      [
        "3–8秒",
        "女阵师展开双手让阵纹连成圆环；男修士挡住迎面碎石，两人动作相互配合。",
        "横向跟拍中景，明确动作顺序，石屑适量，不掩盖人物面部。"
      ],
      [
        "8–12秒",
        "石门开启，暖金光落在他们脸上，两人交换眼神，向前迈出一步。",
        "眼神特写接背面双人远景，门内只见光，留出故事悬念。"
      ]
    ],
    "en": "Original epic hand-drawn 2D cultivation adventure with two adults, a 27-year-old crimson-robed cultivator and a 25-year-old blue-robed formation mage. 0–3s: a towering stone gate lights up above a stormy peak, wide scale shot. 3–8s: she connects glowing sigils while he shields against a few flying stones, sequential actions and lateral medium tracking. 8–12s: the gate opens in warm golden light, they exchange a look and take one step forward, eye insert then rear wide. Original characters, no copied franchise."
  }
];

export const videoPlans = videoSpecs.map(spec => ({
  id: spec.id, title: spec.title, category: spec.category, ratio: 'wide', media: 'video', type: 'prompt',
  tool: '支持多镜头的视频工具', tags: [spec.category, spec.tone, 'AI短剧', '原创分镜', '12秒'],
  summary: spec.shots[0][1], status: '分镜已写 · 样片待补', previewLabel: '视频分镜 · 未生成成片',
  shots: spec.shots, rawPromptLabel: '原创视频分镜提示词（尚未生成）',
  fields: [field('subject', '原创成年角色', spec.cast), field('scene', '场景', spec.scene), field('style', '情绪与风格', spec.tone)],
  zh: '制作一段12秒原创' + spec.category + '短剧，目标画幅16:9。角色：{{subject}}。场景：{{scene}}。情绪：{{style}}。所有人物均为明确25岁以上成年人，不复制参考作品中的相貌、服装、对白或剧情。\n\n' + spec.shots.map(shot => shot[0] + '｜' + shot[1] + '\n镜头：' + shot[2]).join('\n\n') + '\n\n连续性：同一角色脸部、衣着、道具与光向保持一致，动作有准备、执行和停顿，保持轴线和空间关系。禁止肢体变形、闪烁、面部漂移、随机新增人物、字幕和水印。声音：自然环境音，轻音乐可后期添加；如工具不支持对话或音频，将台词与音效另行制作。\n工具设置：在实际支持的工具中设置16:9及12秒；不支持整段时按上述时间段分别生成，再剪辑。先做短段验证，不保证一次生成达到分镜或时长要求。',
  en: 'Create an original 12-second short drama. Adult cast: {{subject}}. Setting: {{scene}}. Mood: {{style}}. ' + spec.en + ' Maintain identity, costume, lighting, props, screen direction and coherent action. No deforming hands, flicker, extra people, subtitles or watermarks. Natural ambience; make dialogue or music separately if unsupported. Set 16:9 and duration in the actual tool; generate shots separately if needed. No copied faces, costumes, dialogue or plot from existing shows. This brief has not been tested.',
  steps: ['先读三段分镜，替换原创角色与场景。', '在自己的视频工具核对画幅、时长、参考图和音频支持；先测试一个镜头。', '检查角色一致性、动作与空间连续，再剪辑三段镜头和声音。'],
  requirements: '目前只有完整视频分镜与复用提示词，没有对应可播放样片。本站未调用视频生成服务；授权来源或用户视频接口与预算确认后再补成片。',
  tip: '美剧感、韩剧感是镜头与氛围方向，不代表官方剧集、The Accident片段或现成演员。参考视频只借鉴方法，不冒充原片作者提示词。',
  source: { label: '视频工具能力说明（非样片出处）', url: 'https://seed.bytedance.com/en/seedance2_0' }
}));
