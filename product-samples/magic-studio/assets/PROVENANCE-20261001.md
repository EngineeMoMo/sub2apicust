# 2026-10-01 新增视觉与授权影片

> 撤下记录（2026-10-01）：用户否决下方4段影片，已从目录及assets运行文件删除，以下仅保留历史取证，不再作为当前作品或打包输入。最新六张原创图精确提示词见PROVENANCE-VIBRANT-20261001.json；当前视频为十套原创分镜，无对应成片。

## 原创图片

工具：内置 image_gen。没有调用用户自己的模型 API；没有上传参考图给生成器。全部人物明确为25岁以上的虚构成年人，不复制用户参考图的脸、服装或构图。PNG保留，WebP仅缩放和编码，缩略图未裁切。可替换的中文／英文模板与生成图不是同一次测试，模板待目标模型实测。

- fashion-editorial：27岁虚构成年女性，海岸屋顶，陶土色丝质吊带、敞开象牙亚麻衬衫、高腰牛仔裤；双手自然倚靠石栏，无摸头、裸露、内衣或夸张比例。午后地中海光线、海景，生活化杂志摄影。
- cyber-cosplay：26岁虚构成年女性飞行员，原创墨绿炭灰装甲飞行服、银色短发；雨后高架车站，头盔拿在腰侧；不透明实用衣料，不使用既有IP标识。
- anime-pilot：25岁原创成年女性飞行员，蓝色外套、奶油衬衫、棕色长裤与折叠航海图；未来海滨城旧阳台，手绘二维动画背景，不复刻既有角色。
- orbital-mecha：原创石墨灰／象牙白轨道维护机甲，地球背景开放机库，技师比例参照、可信机械关节与自然斜光；不是特定高达或既有IP。
- snow-leopard：成年雪豹在喜马拉雅雪岭花岗岩间行走，四肢与长尾自然，冬日侧光、长焦纪实质感。
- headphone-ad：无品牌钛金属头戴耳机，钴蓝石台、象牙布料与橙色亚克力，硬光、真实金属与织物，不增加标志或文案。
- sweet-woman：25岁虚构成年女性，棕色微卷发与温柔微笑；书店窗边，鼠尾草绿针织衫、浅蓝条纹衬衫、牛仔裤；手拿杯与扶书，真实肤质、暖窗光与冷街景反射，独立电影氛围。
- sweet-man：26岁虚构成年男性，榛色眼睛与深色卷发、亲和微笑；夏雨后海岸街道，奶油T恤、海军蓝衬衫与棕色长裤；手扶蓝色自行车并提花束，傍晚暖光与蓝绿反射，不用参考图的西装或脸。

以上是实际生成要求的中文摘要，不是精确逐字prompt。原始生成PNG来自本线程 generated_images，未把未经授权的抖音画面存入产品资源。

## 授权影片（本地播放，不是模型生成）

四条均裁短、缩放到960宽、H.264/yuv420p、移除音频并优化起播；封面从对应片段第2秒截取。总视频约2.62MB。署名、授权链接和改动说明显示在右侧与播放弹窗。

| 文件 | 原作署名 | 许可 | 原素材／取段 |
| --- | --- | --- | --- |
| film-elephants.mp4 | Elephants Dream, ©2006 Blender Foundation / Orange Open Movie Project | CC BY 2.5 | https://download.blender.org/demo/movies/elephantsdream_teaser.mp4.zip ，12–19秒 |
| film-sintel-landscape.mp4 | Sintel, ©2010 Blender Foundation / Durian Open Movie Project | CC BY 3.0 | https://download.blender.org/durian/trailer/Sintel_Trailer1.720p.DivX_Plus_HD.mkv ，1–7秒 |
| film-sintel-character.mp4 | 同上 | CC BY 3.0 | 同上，24–31秒 |
| film-bunny.mp4 | Big Buck Bunny, ©2008 Blender Foundation / Peach Open Movie Project | CC BY 3.0 | https://download.blender.org/demo/movies/BBB/bbb_sunflower_1080p_30fps_normal.mp4.zip ，60–68秒 |

授权核实：2026-10-01读取各原作官网 [Elephants Dream](https://orange.blender.org/blog/creative-commons-license-2/)、[Sintel](https://durian.blender.org/sharing/)、[Big Buck Bunny](https://peach.blender.org/about/)。页面明确CC BY，例外商标及额外音轨未使用。授权原文HTML保存在ignored output目录，不能把通用Blender条款当每项素材授权。影片是风格学习参考，绝不标成下方提示词的生成效果或本站生成实测。

## 用户抖音参考

用户提供三条完整链接及人物截图，仅用于理解“年轻成年人物、甜美亲和、自然肤质、克制电影光线”的偏好。不搬运画面、字幕、演员脸或剧情。检索工具无法读取；Chrome连接失败，内置浏览器读到首条页面与登录提示，未完成三条视频的实际完整观看，不宣称已逐镜反推。

## 转码工具

FFmpeg官方页面链接的gyan.dev Windows essentials 9.0.2，临时放ignored output，不进入应用镜像或更改系统PATH。下载包SHA256与发布摘要一致：60F467265B1E312373DBCD92200C2618A74850F98D3D078E94296BB3FA2047BA。
