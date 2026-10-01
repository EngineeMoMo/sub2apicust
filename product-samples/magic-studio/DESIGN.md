---
name: 魔法工坊
description: 魔法家族第四产品的本地视觉实现归档
colors:
  bg: "#101d20"
  surface: "#17282b"
  surface-2: "#203438"
  text: "#e8f2ef"
  muted: "#adc3bf"
  accent: "#a1d9ce"
  on-accent: "#173b38"
  line: "#33494c"
  field: "#122225"
  hover: "#243d40"
  bg-light: "#f3f7f5"
  surface-light: "#fff"
  surface-2-light: "#e9f0ed"
  text-light: "#19352f"
  muted-light: "#536d66"
  accent-light: "#096b68"
  on-accent-light: "#fff"
  line-light: "#d3dfda"
  field-light: "#f7faf8"
  hover-light: "#e5eeea"
  media-label: "#102327df"
  media-label-text: "#eef5f3"
  preview-overlay: "#132327dd"
  image-text: "#fff"
  dialog-backdrop: "#071518bb"
typography:
  display:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "clamp(26px,2.35vw,38px)"
    fontWeight: 650
    lineHeight: 1.35
    letterSpacing: "-.025em"
  headline:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "22px"
    lineHeight: 1.4
    letterSpacing: "-.01em"
  title:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "17px"
    fontWeight: 600
  card-title:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "15px"
    fontWeight: 600
    lineHeight: 1.55
  body:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "16px"
    lineHeight: 1.6
  label:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "13px"
  control:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "14px"
    fontWeight: 600
  prompt:
    fontFamily: "'Segoe UI','Microsoft YaHei','PingFang SC',sans-serif"
    fontSize: "13px"
    lineHeight: 1.85
  command:
    fontFamily: "Consolas,monospace"
    fontSize: "12px"
    lineHeight: 1.7
rounded:
  field: "7px"
  control: "8px"
  nav: "9px"
  search: "10px"
  compact-card: "12px"
  panel: "14px"
  pill: "999px"
spacing:
  tight: "8px"
  small: "12px"
  regular: "16px"
  gallery: "18px"
  panel: "20px"
  spacious: "24px"
  workspace: "32px"
  wide-workspace: "40px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    typography: "{typography.control}"
    rounded: "{rounded.control}"
    padding: "10px 16px"
  button-secondary:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.text}"
    typography: "{typography.control}"
    rounded: "{rounded.control}"
    padding: "10px 16px"
  button-icon:
    backgroundColor: "transparent"
    textColor: "{colors.text}"
    rounded: "{rounded.nav}"
    height: "44px"
    width: "44px"
  input-material:
    backgroundColor: "{colors.field}"
    textColor: "{colors.text}"
    rounded: "{rounded.field}"
    padding: "9px 10px"
  chip-category:
    backgroundColor: "transparent"
    textColor: "{colors.muted}"
    rounded: "{rounded.pill}"
    padding: "8px 12px"
  chip-category-selected:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.accent}"
    rounded: "{rounded.pill}"
    padding: "8px 12px"
  nav-current:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.accent}"
    rounded: "{rounded.nav}"
    padding: "10px 16px"
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.panel}"
  workbench:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.panel}"
    padding: "20px"
---

# 魔法工坊设计规范

## 概述

**继承的品牌方向：“魔法家族 · 雾钛青”**

这是既有母品牌的实现归档，没有另选母品牌世界或批准新的概念稿。继承依据为根目录 `BRAND_IMPLEMENTATION.md` 第16–18节、现有狮冠M和本目录 `PRODUCT.md`：保留M／狮子／皇冠、雾钛青、默认深色及浅色配套、中文系统字体。具体工作区策略留在 `SURFACE.md`；本文件的布局数值仅约束魔法工坊当前表面，不推广到中转站、配方或SynaRoute。

当前视觉由低彩度界面与原创图片共同组成。界面用背景、面板、字段的明度差和细线区分层级；雾钛青用于选择、复制、来源链接和焦点。图片承担主要色彩，Skill／工作流以工具名和线性图标表达，避免伪造生成结果。

**视觉特征：**

- 继承狮冠M与雾钛青日夜主题。
- 图片先呈现效果，文字始终标明媒介与验证状态。
- 细线、适度圆角、可见焦点与明确中文操作。
- 桌面保留复制操作，手机显式返回灵感与筛选。

**已核实的归档依据：**读取 `theme.css`、`index.html`、`app.mjs`、`core.mjs`、`PRODUCT.md`、`SURFACE.md`、`README.md` 和 `assets/PROVENANCE.json`，并目视检查有效截图。token来自实际CSS；以样式表最后定义的规则为准。`.impeccable/design.json`记录阴影、动画、断点和组件片段，不修改运行时样式，也不补造未使用的色阶。

**验证口径：**`.impeccable/review/tests.log`为本轮21项通过记录；`QA.md`保留真实浏览器、素材与截图证据限制。`finish-review.md`首轮为fix，`finish-verdict.md`由同一审查者将三项修复评为resolved／ship，范围仅限类型切换与导航同步、手机返回及材料保留、两个dialog名称。此结论不等于用户视觉确认、全产品生产批准、真实模型或Skill效果验证。扫描引擎不可用（exit127），没有自动扫描通过结论。

## 颜色

前言token记录深色默认值及浅色对应值；运行时仍由同一组CSS变量随 `data-theme`切换。颜色按用途取用，不按图片色彩临时更换界面主色。

### 主色

- **雾钛青：**`accent`／`accent-light`用于主按钮、选中状态、收藏、官方来源、文本选区和焦点；`on-accent`／`on-accent-light`配成按钮前景。浅色沿用母品牌主色，深色采用实际实现的浅雾青。

### 中性色

- **炭青／雾白背景：**`bg`及其浅色对应值是页面底，`surface`是卡片、复用区与弹框，`surface-2`是更明显的内层或选中导航。
- **正文与说明：**`text`承载标题和正文，`muted`承载帮助、分类、次级操作和来源状态。浅色始终切换整组token。
- **字段与细线：**`field`提供输入／提示词背景，`line`分隔区域和界定输入，`hover`用于透明操作的指针悬停。
- **图片上的标签：**`media-label`／`media-label-text`用于卡片媒介标签；`preview-overlay`／`image-text`用于预览说明与运镜控制。弹框遮罩使用 `dialog-backdrop`。

**品牌继承规则。** 复用既有雾钛青与狮冠M，不把本产品素材色彩转换为新的家族主色；SynaRoute既有紫色不在本归档范围内。

## 字体

**标题／正文字体：**统一继承前言中的中文系统字体栈，没有外部字体请求。**标签／等宽字体：**普通标签沿用正文；官方安装命令才用Consolas及等宽回退。

### 层级

- **Display：**页面唯一h1使用 `display`；手机在600px以下为28px，359px以下为25px，延续相同字重与行高。
- **Headline：**复用条目标题使用 `headline`；玩法标题同为22px但行高1.5，手机为20px。小屏复用条目标题降到21px。
- **Title：**复用区标题使用 `title`；工具名在画廊为20px、预览为19px，手机画廊为17px。
- **Card title：**卡片标题使用 `card-title`，600px以下为13px；卡片分类是12px，手机为11px。
- **Body：**整体基准为 `body`。条目摘要与控件多为14px；材料标签为13px；验证状态和帮助为12px。这里归档实际密度，不把所有局部文字误写成16px。
- **Prompt / Command：**提示词为 `prompt`，安装命令为 `command`。保留提示词换行和命令任意位置折行；计数用tabular-nums。

## 布局

结构为家族顶栏、左侧发现内容与右侧复用区。发现区按标题／搜索与排序／媒介／用途／结果状态／画廊／页脚排列；Skills复用同一检索结构；新玩法隐藏检索控件，呈现图片加三步操作。

桌面顶栏高76px，横向边距随视口在20px与64px间变化。工作区初始上边距36px、底边距48px；左列自适应，右列360px，列间32px。没有固定居中的最大宽度：超宽显示器通过增加媒体墙列数使用空间。

| CSS范围 | 画廊列数 | 右侧／正文行为 |
| --- | --- | --- |
| 1201–1649px | 3 | 360px复用区，32px列间距 |
| 1650px起 | 4 | 380px复用区，40px列间距 |
| 2150px起 | 5 | 延续380px复用区 |
| 1101–1200px | 3 | 330px复用区，24px列间距、18px内边距 |
| 901–1100px | 2 | 330px复用区，24px列间距、18px内边距 |
| 601–900px | 2 | 单列正文，复用区在画廊之后；材料字段两列 |
| 600px及以下 | 2 | 左右18px，卡片间12px；材料字段单列 |
| 359px及以下 | 2 | 左右12px，标题与计数缩减，卡片收藏按钮避让正文 |

2026-10-01最新画廊采用1px细行CSS Grid，卡片依次放到当前最短列，不用CSS columns、dense补洞或统一裁切。列表DOM仍按精选／排序结果排列；放置顶部随DOM不递减，同高时从左向右，键盘与读屏不改为逐列顺序。纵向间距桌面18px、手机12px，取整最多多出不到1px。原图自然比例、contain与完整提示词保留，38图宽高预留值与原始PNG逐一验证。ResizeObserver跟踪卡片内容与容器宽度，加载／错误／视频元数据／窗口变化事件回退，筛选后断开旧卡观察；隐藏玩法时停止观察，pagehide清理、pageshow恢复。响应式列数由CSS变量指定，降列前归位卡片，避免旧4／5列位置生成隐式空列。没有观察器或无法测高时保持普通Grid可读，不隐藏内容。玩法桌面图片占36%，内容占余宽，600px以下上下排列。

实测旧1280布局最大同列空白165.69px；新版320／390手机12–13px、768／1280／2200桌面18–19px，外层横溢出0，原始提示词详情、关闭焦点、视频分镜、空态恢复保留。源码预览截图studio-masonry-*与几何／日志在本线程visualizations及output/studio-masonry-20261001；未更新8080，不把模拟观察器测试当浏览器排版证据。设计扫描器exit127不可用，依照技能布局／craft-floor进行一次桌面手机检查与一次修复确认，不宣称自动扫描或独立审查通过。

复用区在桌面sticky距顶24px，最终最大高度为 `calc(100dvh - 136px)`。区内内容独立滚动；标题、复制／下载／分享和动作说明留在容器外层，长提示词或Skill说明不会把复制操作挤出面板。900px以下取消sticky和高度限制，内容随正文滚动；选择条目跳到复用区并聚焦aside，显式返回按钮滚回发现区并聚焦搜索（玩法中聚焦最近浏览），不清筛选或本页材料。

900px以下顶栏换行，导航独立占一行且可横向滚动。600px以下搜索占一行，排序与媒介／更多筛选顺序仍可达；归档以样式表末尾覆盖后的静态更多筛选布局为准。

有效截图位于 `.impeccable/review/`：`desktop.png`、`mobile.png`、`user-1280.png`、`width-320-live.png`、`width-768.png`、`width-2560.png`、`skills-live.png`、`skills-bottom.png`、`plays.png`、`video.png`、`mobile-edit.png`、`mobile-return.png`。320／390／768／1280／1440／2560无外层横向溢出来自 `QA.md`实测；没有把缩放展示后的截图像素当CSS几何。`user.png`为历史，`skills.png`及 `width-320.png`导出失真，不作为当前验收依据。

## 层次与投影

静止界面以色层与细线组织深度，媒体卡片不靠投影抬起。选中卡片使用2px雾钛青环；透明按钮通过hover色层反馈。阴影只用于toast与dialog，准确值保存在sidecar：深色和浅色各自有一组。dialog另有半透明遮罩，遮罩不是页面的常驻装饰。

卡片图像hover在450ms内小幅放大至1.045，采用实际的缓出曲线。二维运镜默认暂停，按播放按钮才开始12秒线性循环；后台页自动暂停。减少动画偏好关闭过渡与运镜动画，滚动跳转也改为即时。没有加载骨架、视差或自动播放的视觉承诺。

## 形状

框架用轻弧角而非一律药丸：输入 `field`、按钮与提示词框 `control`、导航与预览 `nav`、搜索 `search`、卡片／复用区／dialog `panel`。手机卡片改为 `compact-card`；只有用途筛选是 `pill`。边框以1px `line`为主，图像容器裁切溢出。

狮冠M使用 `assets/mofa-mark-flat.png`既有位图，顶栏36×36px显示、10px圆角；保留原轮廓和品牌配色，不用界面代码重绘标志。线性图标来自 `icons.mjs`，默认20px、1.7描边、圆端／圆连接，装饰图标设 `aria-hidden`。

## 组件

### 按钮

主按钮用于复制与空态恢复，前言记录日常尺寸与形态，最低高度46px。透明图标按钮常规44×44px；卡片收藏最终宽度44px、最低高度44px。主按钮hover亮度0.94；图标按钮hover使用 `hover`；全局键盘焦点为2px `accent`、4px偏移。复制等待临时禁用按钮，降至0.6透明度。

### 标签与筛选

媒介采用分段按钮，选中为 `accent`／`on-accent`填充；用途采用细线药丸，选中为 `surface-2`加雾钛青文字与边框。都以 `aria-pressed`表达选择。媒介是图片／视频，资源类型是提示词／Skill／工作流，二者保持独立含义。选择Skill或工作流会进入Skills区；Skills区选择提示词会返回发现区，结果与导航同步。

### 卡片与容器

媒体卡片有图像、媒介标签、标题、用途／资源类型和独立收藏操作；选中环与 `aria-pressed`指向当前条目。Skill卡片显示真实工具名、线性图标与摘要。图像懒加载和异步解码；加载失败保留文字“预览暂时不可用，提示词仍可使用”。本地渲染不虚构热度、点赞或成功率。

### 输入与字段

搜索最少48px高，用图标、字段与 `/`快捷键提示组成，焦点时边框变为雾钛青。材料字段最少44px高、每字段最多300字符，文本替换为纯文本，空白使用示例值。完整提示词为只读textarea，最少144px、最多360px高，可纵向调整。恢复示例需内联确认，先聚焦“保留我的内容”；取消与确认后返回恢复按钮。

### 导航

顶栏是发现／Skills·工作流／新玩法／我的收藏；当前项用 `aria-current="page"`、内层色底与雾钛青文字。最近浏览是标题旁的次级操作。切换区域清筛选但本页材料按条目保留；浏览器后退／条目hash可恢复有效选项。跳过链接直达发现内容； `/`快捷键只在非输入区且没有打开dialog时工作。

### 复用工作区

复用区依次呈现媒体／工具预览、标题与验证状态、材料、中文／英文提示词、可展开步骤、官方出处以及固定操作行。图片预览采用contain显示完整图；二维视频示意使用独立播放／暂停控件且明确标为示意。方法来源链接保留新窗口安全属性。

自动复制失败时插入并选中手动复制textarea，给出Ctrl+C或长按提示，不能仅显示成功toast。TXT导出包含材料、条件、来源和验证状态；分享仅编码条目ID。收藏、最近浏览与主题只持久化白名单字段；存储不可用时保留本页状态并准确提示。私人材料仅本页Map，不进入存储或链接。

### 空态、反馈与弹窗

搜索无结果提供清除筛选；空收藏／历史说明如何产生记录并提供去发现入口。toast为 `role="status"`和礼貌live region、4200ms显示；结果数也为礼貌live region。大图与使用说明使用原生dialog，均用有效非空标题的 `aria-labelledby`；关闭按钮、Esc与遮罩外点击可关闭。原生模态键盘焦点返回由本轮真实浏览器确认，详见QA，不宣称完整读屏审查。

### 素材与扩展边界

素材来源由 `assets/PROVENANCE.json`管理：8张内置image_gen原创PNG、对应WebP与缩略图，以及既有母品牌PNG。页面使用WebP，源PNG保留；25个PNG／WebP的Description／XMP与来源记录一致，证据为 `asset-check.log`。新增素材必须记录精确生成prompt或实际出处，保留同等metadata；不能用原创视觉图证明可替换模板已在目标工具实测。

当前仅本地静态产品：8图片模板、4视频提示词、2 Remotion Skill、2 ComfyUI官方模板指引及4个组合玩法。视频是静态图二维运镜；Skill未安装运行、模板待实测。扩大目录、实测模型、接入生成API、账号、自动搜索、调度、正式托管和生产发布须按产品事实与后续授权推进，不从本设计文档推导启用许可。用户已明确“先做产品，定时任务先不启用”。

## 使用准则

### 保留：

- 保留既有狮冠M与雾钛青；2026-10-01 用户要求无偏好默认浅色、内嵌跟随主站，切换时替换整组界面token。
- 用图片呈现效果，同时持续显示原创示例、二维示意、未实测或未运行等真实状态。
- 维持桌面固定复制操作与手机显式返回，返回保留筛选和当前页材料。
- 保留可见焦点、有效dialog名称、原生键盘退出、减少动画偏好及准确失败回退。
- 在新增资源和素材时补验证状态、真实官方出处与来源metadata。

### 避免：

- 不以本产品归档重定母品牌、改SynaRoute紫色或重设计其他家族产品。
- 不把原创图、二维视频示意或复制操作成功写成目标模型／Skill效果已验证。
- 不用 `user.png`、`skills.png`或 `width-320.png`替代本轮有效截图。
- 不把三项修复的ship评分扩展成全产品生产批准或自动扫描通过。
- 不持久化私人材料、不把材料放进分享链接、不创建未授权定时任务或在此文档中虚构统计。

## 2026-10-01 工作区与作品展示更新

- 内置控制台正文使用宿主适配器同步主题，隐藏重复品牌导航和独立主题按钮，原静态产品仍可独立浏览。
- 作品列表改为桌面3列／较窄2列网格、统一4:5缩略画幅；用途单行横向滚动，不让大量类别压住作品。缩略图只是索引，点击完整大图／视频查看原画幅，右侧与弹窗均为 `object-fit:contain`。
- 左侧图片直接打开大图、影片直接打开原生播放器；完整影片、控件、署名和许可显示在弹窗，关闭恢复焦点并暂停播放。右侧保留材料替换与复制；手机使用显式返回。
- 当前16张图片、4段授权影片、4份Skills／工作流。新增8张原创图不使用参考截图作为输入；人物为25岁以上虚构成年人，借鉴清新情绪而不复刻相貌。素材来源与处理见 `assets/PROVENANCE-20261001.md`。
- 视频改为 CC BY 原片节选，旧二维运镜标签和生成结果暗示不再使用；不能把浏览器播放通过扩大成提示词、真实模型、付费视频或Skill实测。

## 2026-10-01 视频与分镜待开放（当前优先）

- 最新素材收口覆盖下方历史数量：47图、10套封存分镜、4工具，共61条源码记录，前台发现47图；39图有实际prompt、8图缺原始记录。五个指定题材各新增两图，另补首页同一人像；撤下夜幕时尚与珊瑚色球鞋，时尚保留海风图。真实素材与尺寸／提示词检查见PROVENANCE-CURATED-20261001及定向测试，不用改标签冒充新图。后续素材与人工审核由用户继续。
- 内嵌topbar保持原布局，补不透明`var(--bg)`与z-index:5，修复滚动时正文透出叠字；日夜主题变量保持。浏览器代表视图与完整画幅检查见output/studio-refinement-20261001，未部署生产；没有自动设计扫描通过结论。

- 保留既有原画幅瀑布流，不用删素材或改裁切来关闭功能。视频分类使用真实disabled按钮与“视频与分镜 待开放”文字，列表上方常驻友好提示；图片、Skills与工作流正常使用。
- 分镜从发现／搜索／收藏／历史和公开投稿展示隐藏，旧链接回退图片并提示，组合玩法的分镜步骤disabled且不显示进入箭头。原始目录及收藏记录保留，以便后续恢复。
- 投稿类型显示“视频成片（待开放）”，提交及旧视频重投有前端阻断；保留旧记录查看／撤回与管理员管理。禁用颜色沿用muted，不通过过低透明度损害可读性，不新增组件或依赖。

## 2026-10-01 用户否决旧素材后的当前实现（历史）

- 本节覆盖下方旧数量、统一裁切和影片播放记录。当前22张原创图像、10套视频分镜、4份Skills／工作流；原4段影片与海报已撤下，十套分镜没有对应可播放样片。
- 六张新作品优先：青年音乐节／滑板人像、紫色Cosplay、赤金天龙、夏日动漫、软糖电商；色彩来自作品，不另改雾钛青界面或母品牌。成年人物明确25岁以上，未将参考截图传给生图工具。
- 卡片图片保留原画幅，不用统一4:5裁切；下方独立44px复制模板与“改成我的”，右上收藏有清晰对比。桌面3列、窄屏2列，超宽4／5列；保持DOM阅读顺序。
- 作品详情为桌面媒体／提示词双栏、手机上下排列；完整图像contain，原始实际提示词与可替换模板分开。14张有逐字对应的原始记录，8张旧图明确缺失，仅提供模板；不编造热度或已实测状态。
- 视频条目显示首段分镜摘要与“样片待补”，详情可读完整三段及复制提示词；没有静态运镜冒充视频或伪播放入口。美剧感4、韩剧感3、修仙动漫3。
- 弹窗关闭恢复焦点，重开scrollTop归零并聚焦关闭；复制失败文本留在当前弹窗，避免被模态遮住。材料与语言沿用右侧当前值，原始记录不随替换变化。
- 默认浅色／跟随宿主规则保持。新截图studio-vibrant-*来自4185本轮预览，不代表8080或生产已更新；验证与未验边界看README与根HANDOFF。

## 2026-10-01 再增十张原创图（最新目录）

后续风格／投稿版覆盖本节旧数量：38图、10分镜、4工具，52条、发现48条，准确原始记录30图。先展示像素／水墨／美漫／水彩／剪纸／复古未来六图，明显区分媒材而非摄影重复。视觉风格独立于题材筛选；手机横向滚动不挤压外层。投稿CTA为轻描边入口，复用控制台登录和正文；投稿与审核继承主站雾青主题，完整素材contain、不伪造自动审核或成片。审核状态、作者隐私、手工确认与部署限制见COMMUNITY。

- 不重设计界面，本轮只补内容与对应记录。32张原创图＋10套视频分镜＋4份Skills，共46条，发现区42条；最新10张前置，保留此前22张。
- 题材涵盖成年男女街拍、夜色时尚、原创Cosplay、莲池剑修、救援机甲、海獭、跑鞋、极光与狸猫食堂，共9个用途类别，画幅有竖／横／方，不统一裁切。
- 每张真生成原图和缩略WebP一一对应，保存实际完整prompt，详情独立显示原始记录与复用模板；共24张有准确记录，8张旧图仍明确缺失。工具是内置image_gen，不传用户参考图、不复刻脸或已有角色。
- 视频仍只有10套分镜，0段新成片。未启用视频模型、不以静态图动画冒充视频，不自动修改8080或生产。
