# 交接文档（HANDOFF）— sub2api 定制 fork

> **本文件是本项目的交接中枢。** 每个接手的智能体（含 Claude）：
> - **开工前必读**：本文件 + [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) + [SYNC.md](SYNC.md)。
> - **收工前必更**：更新下方「四、当前状态」「五、待办」；任何新定制同时登记 CUSTOMIZATIONS.md。
> - 交接一律走本文件——**别把状态只留在会话里**（会话会丢，文件不会）。
>
> 最后更新：2026-10-01（Asia/Shanghai）。

---

## 一、用户是谁 / 怎么协作
- 语言：**全程中文**（回复、文档、注释）。
- 自部署运维者，目标：把开源 sub2api 定制成自有品牌的商用「GPT中转站」，并长期跟随上游。
- 用户**手动部署**；智能体负责改代码、验证、提交推送，**不代替用户上生产**。

## 二、用户要求（项目意图）
1. **核心**：定制 sub2api 的同时能持续同步上游更新、且定制不丢。
   路线 = **方案 A：配置 + 新增功能**，尽量不动上游核心 → 私有 fork + override 构建。
2. **镜像**：GHCR + GitHub Actions CI（push main 自动构建）。
3. **已交付定制**（细节见 CUSTOMIZATIONS.md）：
   - SynaRoute 深链接一键导入（`synaroute://`，镜像上游 CCS 导入，带管理员开关 `hide_synaroute_import_button`）。
   - 关闭 App 内「在线更新」（`DISABLE_ONLINE_UPDATE=true`；可选把更新源指向自己仓库 `UPDATE_GITHUB_REPO`）。
   - 前端定制叠加层 `frontend/src/custom/`（换肤 + 新增/影子替换页面，低冲突叠加层）。
   - **当前主题**：雾钛青（primary-500=#096B68），Logo/站点名仍走**管理员后台设置**。2026-10-01 用户新要求覆盖历史默认深色：无保存偏好默认浅色，已有偏好保留，子产品跟随 API 当前主题。源码与实际部署版本以第四节最新记录为准。
4. **品牌**：魔法家族 / GPT中转站；域名 https://ai.mofamilys.com ；保留核心「M」设计，允许标志重着色。**2026-09-26 最新确认：雾钛青 + 日夜配套；SynaRoute 保留紫色产品识别，统一基础规范，不全换青色。** 此前深蓝发光风为历史方向。
5. **升级方式**：**手动 Docker 更新**（`deploy/update.sh`），不用 App 内按钮。

## 三、用户准则（红线，务必遵守）
1. 🔴 **只说有证据的话；没证据就说「我不知道」**（2026-09-07 用户原话：「以后都拿证据事实说话，不要理论和以为，不确定的就不回答，找到实质性证据再回答」）。
   - 回答分栏：**已核实**（附取证方式：grep 命中数 / 读了哪个文件哪几行 / 命令完整输出）/ **推断**（附「要什么证据才能确认」）/ **不知道**。
   - 「代码里存在能产生该症状的路径」≠「用户遇到的就是它」；用户报的现场要在**那台机器**上取证。
   - 写「已核对」必须同时写核对方式，否则别写那三个字。
2. **定制维护**：优先新增文件；不得不改上游文件时改动尽量小并打 `[CUSTOM]` 注释，且**逐处登记 CUSTOMIZATIONS.md**。配置类定制（.env / override）走 gitignore，不进仓。
3. **本机无 Go 工具链** → 后端改动**未编译**，靠 CI 验证（`gofmt -w ./...` + `go build -tags embed ./cmd/server` + `go test -tags=unit ./...`）；提交时如实标注「未编译」。
4. **同步上游**：先 `git merge-tree --write-tree main upstream/main` 演算冲突，再真合并；rerere 已开；合并后过 CUSTOMIZATIONS.md 末尾自检清单 + 登录发一个请求冒烟。
5. **换肤**：品牌主色只改 `frontend/src/custom/theme.css`；上游组件里的 `teal-*` 若是**语义分类色**（deepseek、模型徽标、渠道状态）**刻意不改**（改了会撞色/丢语义）。
6. 会话上下文超 **200k** 先压缩再继续，压缩后先核对工作树再接着做。

## 四、当前状态（每次收工更新）

- **2026-10-01 用户授权本轮提交推送（发布准备）**：用户要求修好后推送以便手动更新。已限定暂存65个界面／素材文件，导出实际index源码快照；Node106＋Vue538共644项、类型／Vite构建通过。完整lint首次扫到构建后的gitignored运行副本7项旧静态格式错误，随后按干净CI源码范围排除public/recipes、public/studio重查，结果以`output/studio-refinement-release-20261001/lint-source.log`终态为准。包号与更新脚本继续排除，共享theme只暂存本轮产品段落。旧765e5a9的单测／集成现已由GitHub job API核实success，Go lint及安全扫描failure仍有效。
- **发布仍有明确阻塞**：官方公告／npm与Go模块目录确认Axios修复版1.20.0、x/image0.45.0；投稿handler八处未检查清理错误也已由失败日志定位。尝试安装Axios修复版时自动审批拒绝，理由是安全依赖升级超出当前授权；安装未执行、依赖及锁文件未改。已异步请求明确授权这两项依赖升级及清理检查修复，答复前不处理依赖，不把界面成功与可构建镜像当安全发布通过。取证在`output/studio-refinement-release-20261001`及旧SHA真实任务；本轮功能提交／推送及新CI状态待下续记录，不代部署生产。

- **2026-10-01 用户截图修正与图库收口（最新源码，未发布）**：修复配方／工坊内嵌吸顶栏透明背景及层级造成的滚动正文叠字；配方栏全宽，目录避开顶栏，模型设置锚点保留88px空间。SynaRoute入口使用`/keys?product=synaroute`，该上下文选中紫色SynaRoute，普通`/keys`仍为API；不改桌面客户端或导入功能。首页移除“悬停或点按，置顶欣赏”，三图交互保留。
- **本轮图库最新数量覆盖下方历史38图**：按用户要求删除“夜幕之后，一点绯红”，时尚肖像只保留“海风里的丝绸与牛仔”；删除“跑进珊瑚色的夏天”，两项全尺寸／缩略WebP从公开打包撤下，旧PNG及生成记录仅作历史。Cosplay／动物自然／海报社媒／动漫二次元／空间设计各新增2张真实独立生成图，另补首页窗边人像的同一素材，共47图＋4工具，10套分镜继续封存。39图保留实际prompt，8旧图缺原始记录。十张本轮新图未输入参考图；首页人像复用既有参考图生成结果，其参考权属未核实，不能标全新无参考原创。精确prompt／原PNG／WebP／缩略图尺寸与SHA256见`product-samples/magic-studio/assets/PROVENANCE-CURATED-20261001.json`。后续素材由用户自行补充与人工审核，本轮不再自动扩充。
- **线上模型空框的事实边界**：用户确认现场是`https://ai.mofamilys.com`，本轮浏览器只取得该站登录页，未取得真实账号密钥列表／分组／`/v1/models`响应，**不知道线上根因**。已在源码补充空列表／接口失败／有效密钥均无生图权限的具体提示并禁用无选项控件；切换用途或刷新先清旧密钥／目录，避免沿用文字配置。明确标记的宿主夹具验证“无生图权限”提示、“有权限”可选模拟密钥与模型／应用按钮启用；这不是生产认证或真实生图成功，未自动开权限、创建密钥或调用收费模型。
- **本轮取证与限制**：Node106＋Vue80共186项相关测试、类型／限定lint／Vite构建通过。真实浏览器检查内嵌滚动栏不透字、配方设置标题可见、SynaRoute路由与紫色当前态、五个分类新增标题、时尚仅保留一图；1280桌面与390窄屏深浅代表视图、完整竖图contain及无横溢出。模拟宿主／密钥／模型范围明确写入`output/studio-refinement-20261001/browser-evidence.json`与截图，夹具源仅留output，不进入产品源码。十图接触表完成视觉检查。设计技能扫描引擎不可用，没有自动设计扫描通过结论；公开源码预览投稿目录不可用，与旧后端一致。包号／更新脚本并行文件SHA256比对结果见该output目录，主题仅增加本轮产品规则并保留原并行规则。没有本轮提交推送、8080重建、Go／Docker验证或生产部署，下方765e5a9发布门禁仍未修复。

- **2026-10-01 家族版本已推送但发布门禁未全通过（最新终态）**：功能提交765e5a9516eaecf23d94202b4afc48f9bf723339已推送origin/main，165文件；actual暂存快照333项、类型／定向前端lint／Vite构建通过，包号24文件及更新脚本并行实现保留未提交。新SHA的GHCR任务36825037358已成功，真实日志确认ghcr.io/engineemomo/sub2apicust:sha-765e5a9与latest，摘要sha256:341886529ae61b817a24da3a2a40dbb7062aabbc2148ecc9b7956153cb422e9d；这是镜像构建成功，不代表安全／全部CI通过，**暂不建议更新公开服务器**。
- **该SHA实际失败／待完成项（不得沿用旧全绿记录）**：Security Scan36825037421失败。govulncheck发现golang.org/x/image v0.41.0的GO-2026-6222／5061／4961，修复版本至少v0.45.0；本轮StudioHandler.Submit的image.DecodeConfig→webp／vp8l为可达路径，用户头像旧路径也命中。前端audit发现Axios七个high缺少例外，不能自动补例外掩盖。CI36825037384的Go lint因custom_studio_handler.go八处Close／Remove／RemoveAll错误未检查失败；原站前端／shell／release-helpers成功，Go Unit完成、Integration仍运行时取证，不标全部成功。已异步询问是否继续修复依赖再发布，当前尚无本轮后续修复授权。证据output/family-release-20261001/ci-status.json、GitHub本SHA任务与日志；8080及生产未更新，收费调用未执行。

- **2026-10-01 用户授权家族产品提交推送（发布验证）**：本次限定配方／工坊同镜像、控制台同页与主题跟随、配方同页配置、原创素材／完整prompt、投稿审核和视频／分镜待开放；不提交包号专项、更新脚本及其CI接线，部分共享theme／UPDATE_GUIDE只暂存本产品段落。已从实际暂存区导出独立源码快照（不是整个混合工作树），Node104与前端229项共333项通过，类型／定向lint／Vite构建通过；首次Node因漏构建gitignored配方HTML而失败，先执行权威builder后全部通过，不以旧预览产物补进提交。两忽略文件混合Windows换行曾造成全文件diff，已仅在暂存区归一化，限定diff检查通过。证据output/family-release-20261001。推送和该SHA的新CI／GHCR待本轮后续取证，当前不使用旧25071d4结果；未更新8080、运行收费模型或部署生产。

- **2026-10-01 用户要求暂停视频与分镜（最新有效状态）**：工坊前台现仅展示38张原创图与4份工具，视频筛选显示“视频与分镜 待开放”并真实disabled；列表友好提示“你可以先探索图片，或使用 Skills 和工作流”。原10套分镜及提示词／收藏ID留档不删除，发现／搜索／收藏／历史、玩法分镜步骤、旧链接及审核后的公开视频均不能绕过前台关闭状态；旧链接提示并保留当前图片材料。图片完整预览／对应原始提示词、复用与原画幅瀑布流保持。
- **投稿与验证边界**：视频新投稿选项待开放，提交与旧视频重新投稿路径前端阻断；图片可投、旧视频私有查看／撤回及管理员管理保留，服务端MP4协议未停用。Node104（工坊44＋配方56＋宿主1＋打包3）与Vue11（投稿9＋API2）通过，类型／限定lint／Vite构建通过，已有500kB分块及Browserslist旧数据警告未处理。4185新建tab25验证38图、4工具、4个分镜步骤disabled、搜索及旧链接、完整图与原始prompt、1280与390深浅横溢出0、warn/error空；已恢复浅色与默认视口，未刷新用户材料页。证据output/studio-video-paused-20261001与studio-video-paused-*截图；24个包号并行文件SHA256保持。仅源码／4185预览，未更新8080、Docker／Go构建、提交推送、CI或生产部署；真实登录投稿／视频任务／收费调用未验。

- **更新脚本服务器安装命令交付（2026-10-01）**：用户要求完整操作命令，已生成gitignored的output/install-update-cleanup-20261001.sh并通过bash -n。流程读取服务器当前唯一sub2api容器的固定sha引用，备份旧update.sh，写临时文件／语法检查／chmod后替换，使用同一个当前sha运行更新与镜像清理，最后查Compose状态／镜像列表／df；不是猜latest或切到本机未发布版本。最终提供Windows PowerShell输入SSH地址／端口的scp上传命令和服务器bash执行命令，不依赖尚未发布的私有仓下载链接。尚未执行服务器上传、安装或清理、尚未提交推送；结果待用户输出核验。

- **本轮镜像清理最终验证补充**：增加无sha别名的latest升级及重复运行、更新中切换到其他仓库场景后，假Docker回归为22场景全部通过；三文件bash -n及限定git diff --check通过。CI接线已登记，远端CI尚未运行，服务器脚本尚未替换。下条20场景为中途记录，以本条22为准。

- **2026-10-01 用户授权新增镜像清理（本轮已实施，未发布）**：deploy/update.sh改为健康检查成功后仅本仓sha-*／latest和本仓来源标签悬空镜像清理，保留当前／实际更新前镜像的完整ID和其他运行或停止容器引用；latest通过本地update-before／rollback-previous标签保留旧版，同版更新保留已有回退版，首次同版无记录时选创建时间更早的最近sha候选。健康超时退出非零，拉取／重建失败不清理，检查失败保留镜像并显示原因；同部署目录锁防并发。不自动清理数据库、备份、日志或卷。新增20场景假Docker回归及CI shell接线；Git Bash三文件语法检查与20场景全部通过。服务器需单独替换宿主update.sh；未运行真实Docker删除、更新8080、提交推送或生产部署。并行工坊／包号工作全部保留。

- **2026-10-01 工坊空白修复与视频接口查证（最新布局）**：按impeccable布局与craft-floor保留既有品牌／功能，原齐行Grid造成1280同列最大空白165.69px，改为1px Grid最短列瀑布流；完整原画幅与contain、不改DOM精选顺序／键盘阅读顺序。38图源PNG宽高预留值逐一核对，媒体加载／错误／元数据、ResizeObserver与resize重排；筛选断开旧观察、玩法隐藏及pagehide清理、pageshow恢复。首轮4列缩到2列出现隐式空列已修为CSS列数变量、降列先归位，并加入回归；没有新增依赖、后端／认证／宿主修改。
- **本轮证据／发布边界**：Node101（工坊41＋配方56＋宿主1＋打包3）、语法／限定diff检查、Vite生产构建退出0，原500kB分块警告保留。4185新建tab24，320／390／768／1280／2200真实布局纵向间距手机12–13px、桌面18–19px，顶部DOM顺序不递减、外层横溢出0；完整图与原始提示词详情、关闭焦点、分镜三段／无伪视频、空态恢复实测，warn/error空。截图studio-masonry-*在本线程visualizations，数据output/studio-masonry-20261001/browser-geometry.json；视口已恢复、浅色、新预览保留，没刷新用户草稿页。24个包号并行文件SHA256与原基准一致。技能扫描引擎exit127不可用，不声称设计扫描／独立审查通过；未更新8080、Go／Docker构建、提交推送、CI或生产部署。
- **OpenAI视频事实（2026-10-01官方只读查证）**：当前 https://developers.openai.com/api/reference/resources/videos 顶部与 https://developers.openai.com/api/docs/deprecations 的Sora段明确：Videos API与Sora 2系列2026-09-24关闭，暂无一对一替代API；保留接口页是历史参考。不能说OpenAI从未有视频API，也不能把历史接口当现可用服务；本会话无视频工具是另一事实。本站是否另有第三方视频供应商未知，未检查真实账号权限或收费调用；当前10套分镜仍0段对应成片，投稿人工审核链与上一轮未部署边界保持。

- **2026-10-01 生产磁盘截图取证（覆盖下方待取证状态）**：用户四张服务器终端截图确认/dev/vda1 20G、已用5.5G、可用14G、33%；当前sub2apicust sha-25071d4唯一被容器引用，另六个sha版本均CONTAINERS=0，独占大小合计约1012.8MB。上一镜像sha-321fe99可作为保留回退候选；拟仅删除更旧sha-43b715a、sha-fa45b20、sha-a1b9137、sha-d592f94、sha-58b154f，不用force，五版独占空间合计约843.7MB，实际回收需df复核。部署目录278M，数据库备份backup-0.2.7-2026-09-26.sql.gz为5.0M、应用日志48M、Postgres数据168M、Redis数据58M；无构建缓存，不能把这些目录视为可删备份。旧镜像积累已证实，但无更新前后对比，不断言全部磁盘增长都由镜像造成。仅给用户定向清理命令，尚未执行服务器清理；原上游镜像与chatgpt-next-web虽0容器，用途未确认不列入删除。

- **2026-10-01 生产升级后磁盘减少（取证待办）**：已读deploy/update.sh，指定标签时创建docker-compose.override.yml.bak.<时间戳>但无轮转；不自动执行pg_dump，手动数据库备份不会由该脚本删除。末尾仅docker image prune -f且忽略失败，按Docker文档只清悬空镜像，带sha标签的旧镜像仍可能积累。仓内Compose未声明日志轮转，服务器Docker日志默认值／实际配置未知。尚无生产df／docker system df／镜像清单／部署目录大小证据，不能断言现场主因是备份或旧镜像。提供只读检查，不修改更新脚本、不清理服务器或本机Docker；并行未提交工作保留。

- **2026-10-01 工坊风格／投稿审核版（最新，覆盖下方旧状态）**：已明确本会话只有生图工具，没有可调用的视频生成工具；用户未提供视频接口／模型／预算，不调用收费服务或把分镜冒充成片。新增六张真原创素材：像素港口、水墨巨鲤、美漫信使、水彩海湾、剪纸小虎、复古未来火星站，不输入用户截图、不复制人物。当前38图＋10套待样片分镜＋4工具，共52条、发现48条；30图有准确prompt，8张旧图仍缺失。PROVENANCE-STYLES-20261001.json记录工具、完整prompt、无参考图、三种资产尺寸／字节数／SHA256；WebP等比编码无裁切。新增题材／视觉风格两套独立筛选及可辨识的投稿CTA，不改母品牌首屏或SynaRoute。
- **投稿是真后端，不只是占位按钮**：新增handler/custom_studio_handler.go及custom/studio/api.ts、StudioSubmissionsView.vue。工坊进入同源 `/tools/studio/submit`，复用API登录／正文与主题；管理员 `/admin/studio/submissions`。PNG／JPEG／WebP最大12MiB、32px以上／2400万像素以内，MP4最大30MiB；签名／体积检查不等于完整解码、病毒或自动内容审核。上传必填模型／工具、完整prompt、题材／风格及两项授权，标明actual或reference。初始pending仅作者／管理员可读；明确人工确认才能published进入匿名公开目录／Range媒体接口；支持带原因退回／下架、作者撤回及旧状态409。修改重新投稿，不能直接恢复发布；图片和成片可点开并查看对应prompt，反推模板不冒充原始记录。审核用现有认证、限流、后台模式与合规保护，不给静态子页传JWT；上游router新增[CUSTOM]接线，audit_log精确省略投稿／审核请求正文但保留元审计，工坊CSP只放同源目录读取，未放开外部模型连接。
- **存储／隐私边界**：`pricing.data_dir/studio-submissions`，默认Docker `/app/data/studio-submissions`，不在静态目录，JSON＋media私有持久化；每人5待审核／20记录，全站1000记录／512MiB媒体，撤回等仍占额度、无永久清理UI。只支持单应用实例，未来多副本需DB元数据／对象存储。备份要含整个目录，Postgres备份不含文件；反代体积上限需容纳30MiB加表单。当前页替换材料仍仅内存，用户点击投稿后文件与prompt会上传本站并持久保存，不能继续笼统说所有材料不上传。原始来源、许可、人工审核与模型验证分开；未接自动NSFW／版权服务，不保证完全防滥用。细节product-samples/magic-studio/COMMUNITY.md及deploy/FAMILY_INTEGRATED.md。
- **验证与发布边界**：Node96项（工坊36＋配方56＋宿主1＋打包3）与Vue64项定向回归通过，类型／限定lint／Vite生产构建退出0；日志output/studio-community-20261001。Docker代理先验200，在隔离镜像 `sub2apicust:studio-check-20261001` 中5个投稿／审计顶层测试、相关静态产品回归与Go embed编译通过；MP4测试仅签名／Range夹具，不是播放／生成样片。首次编译响应函数名错误已修复；尝试web全量时两个既有favicon.png用例返回HTML失败，未修无关测试，不标全量Go通过。4185另开tab22，六图真实加载、12种风格及组合筛选、完整原prompt、水墨桌面详情、390深色剪纸详情均验；无外层或详情横溢出、warn/error空。匿名投稿入口真实回跳 `/login?redirect=/tools/studio/submit`，未借模拟账号宣称真实投稿／审核闭环。截图studio-style-*在本线程visualizations，恢复浅色／正常视口，没刷新用户草稿页。24个包号并行文件SHA256保持，preserved-check.json；未替换8080、提交推送或部署生产，8080原服务及DB／Redis仍healthy。旧后端的4185会显示“投稿目录暂不可用”，这是未发布状态；不是已上线可用。真实账号、可播放视频、管理员视觉及后续CI／正式镜像更新仍待。

- **2026-10-01 用户再增十图（最新目录与证据）**：用户确认询问视频是否仅提示词，答复明确当前只有10套原创分镜、0段新成片。另按要求实际使用内置image_gen生成10张独立原创图：街角成年女性、唱片店成年男性、夜色时尚、赤红信使Cosplay、莲池剑修、海岸救援机甲、海獭、跑鞋、极光湖景、狸猫食堂；没有输入用户参考截图，人物明确25岁以上。当前32张图＋10套分镜＋4份Skills，共46条、发现区42条；最新十图前置，此前22图保留。
- **准确记录／资源**：十张PNG已复制入工坊assets，网页使用等比缩小的完整WebP和缩略图，无裁切或手工像素编辑；assets/PROVENANCE-EXPANDED-20261001.json保存精确实际prompt、工具、无参考图、生成文件名及三种资产尺寸／字节数／SHA256。共24张图有原始实际记录、8张旧图仍明确缺失；复用模板另写未重跑，不将原图生成成功扩大成任意模型模板实测。new-gallery.mjs共用记录映射、catalog前置新图，index数量和README／PRODUCT／DESIGN同步；没有新运行module、后端／安全白名单／宿主／认证／主题修改或新增依赖。
- **本轮验证**：工坊30＋配方56＋宿主1＋打包3，共90项通过，Vite生产构建退出0（既有500kB分块警告不在本轮修），日志output/studio-images-more-20261001/{node,build}.log。打包十张full／thumb与源WebP逐字节一致、不公开PNG和PROVENANCE JSON。4185另开tab21验证十张完整图全部真实加载／contain、实际prompt与来源逐字一致、1280桌面浅色和390手机深色无外层横溢出，手机横幅详情无内部横溢出，warn/error为空；证据browser-prompt-check.json、studio-more-browser-20261001.json及studio-more-{gallery-desktop,prompt-desktop,mobile}-20261001.png。截图在本线程visualizations，最终恢复浅色／正常视口并保留新预览，不刷新用户材料页。24个并行包号文件SHA256保持原基准，output/studio-images-more-20261001/preserved-check.json。未调用视频或付费API、编译Go、重建Docker、更新8080、提交推送或生产部署；真实剪贴板／TXT落地／账号模型授权仍未验，视频生成服务／预算或授权素材仍待确认。

- **本轮最终范围核对**：24个包号并行文件SHA256与原基准全部一致，证据output/studio-vibrant-20261001/preserved-check.json；限定工坊／打包测试／交接文档git diff --check退出0，app与new-gallery语法检查退出0。最终浏览器实证studio-vibrant-browser-20261001.json及最新prompt-desktop截图保存在本线程visualizations，视口已恢复正常、主题恢复浅色，另开交付tab20保留，未刷新用户原材料页。视频样片问题尚待回复，十套分镜不是已完成成片。

- **2026-10-01 工坊素材再次否决与作品／提示词改造（最新源码）**：用户要求撤下4段旧视频、鲜活图像与9–12段年轻人偏好的短剧／修仙视频。已物理撤下4段Blender原片及封面，目录／玩法／打包回归同步；新增6张原创图（音乐节成年女性、成年男性滑板、紫色Cosplay、赤金天龙、夏日动漫、软糖广告），未使用参考截图生成。当前22张图＋10套原创视频分镜＋4份Skills／工作流，共36条；十套分镜为美剧感4、韩剧感3、修仙动漫3，每套三段12秒计划及中英文完整模板。**没有这10套分镜的成片，本请求的视频部分未完成，不称10段视频已交付。** 可调用工具只有生图；用户视频服务／预算或授权现成素材待确认。搜索所得社区集合第三方媒体权利未核实，未下载或擅用抖音原片。
- **作品与提示词对应**：左侧查看完整媒体与提示词双栏详情，手机上下；卡片直接复制模板／改成我的，右侧材料、收藏、语言保留。14张图真实原始提示词（原8＋本轮6）与来源JSON逐字一致，此前8张只有摘要的图明确原始记录缺失；原始提示词、可替换模板与生成状态分开。TXT增加实际记录／当前模板标记；弹窗复制失败文本留在弹窗内，重开归零并聚焦关闭，关闭恢复焦点。默认浅色／宿主主题与控制台嵌入架构不改。
- **本轮证据与边界**：工坊28＋配方56＋宿主1＋打包3，共88项Node回归通过，日志output/studio-vibrant-20261001/node.log；Vite生产构建退出0，build.log。首次直接构建因node_modules/.bin未进PATH缺vue-tsc，补PATH后完成，没有安装依赖。4185真实浏览器六新图加载、对应实际prompt／当前材料、复制成功提示、Esc／焦点、390手机深色与1280桌面浅色无外层横溢出，warn/error日志为空。截图studio-vibrant-{gallery-desktop,prompt-desktop,image-mobile,video-script-mobile}-20261001.png在本线程visualizations；不把复制成功提示当系统剪贴板或真实TXT落地验收。未调用视频模型、收费API、Skills、定时任务；未编译Go、重建Docker、提交推送或更新8080／生产。8080仍按上轮镜像记录，不将4185源码预览当新版部署。

- **收工取证**：最终重跑仍为Node83＋前端86共169通过，限定范围`git diff --check`退出0；24个并行文件摘要一致，记录output/family-workspace-20261001/preserved-check.json。新测试镜像144c2190d8c0，运行容器仍c599729e28df且应用／原PostgreSQL／Redis均healthy（只读inspect／ps）。预览服务4185为本轮自建Vite session89719；实际最终工坊1280×720浅色、20份媒体索引、右侧contain与无溢出，工坊／配方warn和error日志均为空，证据studio-polish-browser-20261001.json及desktop-light截图在本线程visualizations。已清理本轮失败类型构建产生的两个多余dev-products.js／d.ts，仅保留源码mjs与声明d.mts；未动用户旧预览或8080。

- **最终测试镜像（未替换8080）**：2026-10-01本轮最终 `sub2apicust:family-workspace-20261001` 构建退出0，image inspect为sha256:144c2190d8c0096d9890d2c2401477770371b690a09eddab23c2e537a4ab1fee，日志output/family-workspace-20261001/docker-build-final.log；代理预检200且只给构建进程设置代理。包含最新工作区／媒体／样式与Go embed编译，TestFamily两函数通过。未tag覆盖local-theme、未up服务、未改DB或生产；用户确认本轮更新后才能按原runbook部署，前次c599729e28df仍是8080实际版本。

- **2026-10-01 子产品工作区与作品扩充（最新源码，覆盖下方旧新窗口／二维运镜／默认深色描述）**：登录后配方／工坊默认进入 `/tools/recipes`、`/tools/studio`，在 API 控制台右侧正文嵌入同源静态模块，不开新窗口；显式外部独立URL兼容保留。默认浅色、已有深色偏好保留，子页由宿主观察主站主题同步。原模块源码／共享 family-runtime 适配器预留独立模式，不新增账号或通用跨站SSO。配方直接同页下拉本人可用Key／模型／接口格式或手填，目录读取与权限复查由宿主处理，只有显式应用的连接进入子页内存；来源／窗口／nonce配对，退出取消并清空，不自动运行模型或下发JWT。
- **本轮内容与浏览器证据**：工坊16张原创图片＋4段真实授权电影节选＋4份Skills／工作流，共24条。8张新原创图涵盖年轻成年男女、时尚穿搭、Cosplay、动漫、原创机甲、动物、电商，未将用户参考截图作为生成输入；人物明确25岁以上。影片是Blender开放电影的CC BY节选，裁短／缩放／静音，逐条署名、链接与来源记录在assets/PROVENANCE-20261001.md，不是模型生成或模板实测。左侧点击直达大图／影片，右侧与弹窗完整contain预览、互斥暂停；统一网格／横向分类。4185真实浏览器已验完整大图、4个不同影片实际解码播放、1280／390无外层横溢出、浅深主题与手机弹窗。截图studio-polish-*在本线程visualizations目录，临时视口已恢复浅色／正常尺寸；未刷新用户8080材料页。
- **本轮自动验证与边界**：Node／打包83项＋前端／国际化86项，共169项通过；类型、相关ESLint、Vite生产构建通过，日志output/family-workspace-20261001/{node-final,frontend-final,types-final,lint-final,build-final}.log。Docker内TestFamily两函数定向回归和Go embed编译通过，最终镜像状态另记；不标全量Go单测或新CI通过。开发目录入口曾误落登录页，已新增精确dev-products映射，4185用`--config vite.config.ts`避免旧gitignored编译配置优先加载。24个包号并行文件SHA256逐项一致未改，不提交推送／生产发布。
- **参考与未验**：Chrome扩展连接连续返回失败，不能读取用户Chrome已打开的3条抖音；IAB仅确认首条标题／播放控件，登录遮挡，未完整观看三条／逐镜反推。用户截图只作为年轻甜美人物方向，不复制脸／衣服。真实账号下同页Key／模型应用与付费生成未验；模拟宿主／组件证明交互不能替代真实登录。此轮已异步询问是否继续更新本机8080，尚未收到回复，仍保留原8080镜像c599729e28df及数据／容器，不把4185新源码或新测试镜像当8080已更新。

- **2026-10-01 一体化已部署本机 8080（最新终态，优先于下方待构建记录）**：用户明确选择更新本机 8080，现由同一应用镜像提供主站、/recipes/、/studio/，首页标为“本站产品”，不要求另行托管或设置外部地址。运行镜像与构建镜像一致 sha256:c599729e28dfe8572d14c0577e93f4f5818904e489e918a085d6fd46f7bc994b，主站入口 /assets/index-CYEQGOb2.js；原应用单独替换，应用／PostgreSQL／Redis均healthy，health=ok，原DB／Redis容器ID及sub2apicust-local_app_data未变。25项真实HTTP产品边界／资源摘要检查通过、4个个人接口仍匿名401，原精确CORS许可与未知来源403保持；业务计数和全部280键设置摘要前后无变化。证据 output/family-integrated-20261001/{docker-deploy.log,verification.json,verify.ps1}。
- **本轮编译／测试及失败边界**：177项自动测试、类型／相关lint／Vite构建通过；Docker内新增2个Go测试函数（包含14个路径子用例及两种embed入口）通过，Go embed二进制编译完成，日志docker-build-targeted.log。首次尝试完整web测试失败，原因是旧用例请求已不存在的/logo.png（HEAD的public仅logo.svg），不是本轮新增路由断言；旧测试未修改，不恢复旧品牌图，构建仅执行新增TestFamily回归，测试源码不复制到最终运行镜像。原失败日志docker-build.log保留，不宣称全量后端测试或新CI通过。回退标签与606446字节数据库备份沿用下方本轮已核验记录，不删旧镜像／旧产品容器或恢复数据库。
- **实际页面取证与未验项**：8080浏览器验证配方8份、创作筛选2份、会议及生图示例整理、当前API来源只读、工坊12份创作灵感＋4份Skills／工作流、图片加载、复制成功提示和同源家族链接；1280桌面、390手机无横向溢出，浅／深主题可切换，最终恢复深色／临时视口。两页warn/error日志为空。截图family-integrated-{recipes,studio}-{desktop,mobile}-20261001.png及browser JSON在本线程visualizations目录。家族匿名入口和另开/connect/recipes均返回8080原登录页且保留返回参数；IAB点击配置按钮进入等待但未列出弹窗，未完成真实账号登录／已有Key授权／付费模型或系统剪贴板验收，不能说完整接入通过。仅本机部署，本轮未提交推送／生产发布，24个包号并行文件SHA256保持未动。

- **2026-10-01 配方／工坊一体化开发与本机部署授权（最新决定）**：用户要求两个产品开发后一起部署，并明确选择更新本机 8080；生产仍由用户手动更新。原静态源码保留，新增 Vite 打包器默认生成同源 /recipes/、/studio/，主站首页／控制台空地址直接使用真实内置产品，不再要求三个 MAGIC_* 变量或额外产品服务；显式独立 HTTPS 覆盖仍可用，错误地址仍禁用。配方外置 JS／CSS、内置模式固定当前站点配置选择并忽略 api_site 提示，工坊新增同站点家族首页链接；没有新生成模型、定时任务、自动安装 Skills、账号系统或 API Key 单点撤销。
- **实现与初步取证**：新增 custom_family_products.go／test 及 build-products.mjs／d.mts／test，Go embed 两条入口先处理独立命名空间，白名单 GET／HEAD、404／405、禁止目录和原始源图／文档，专属 CSP 不启用脚本 unsafe-inline，不改变全站安全策略或 API 认证。177 项自动测试通过（打包3、配方55、工坊21、原站98），类型／相关 lint／Vite 构建通过，日志 output/family-integrated-20261001/{bundle,recipes,studio,frontend,types,lint,build}.log；新 Go 文件由缓存 Go 容器仅定向 gofmt，镜像编译与后端测试仍等待本轮构建。不沿用旧 25071d4 CI；该远端旧功能不含一体化。
- **本机备份与范围**：Docker 最初未启动，启动本机引擎后发现原三服务停止，保留原容器／卷启动旧应用及数据库／Redis，代理授权200后完成606446字节备份与 pg_restore 目录／副本 SHA256 校验，摘要 0EEED4AF32A57EB5344D82321AD7660A1C75EE42E8C02ABC25AB133958DDBEBC；回退标签 sub2apicust:before-family-integrated-20261001-001108，旧镜像94dba4064ba6。原基线 users3／keys0／groups3／plans1／orders0／dedicated0／settings280，证据 before.json。本轮本机镜像将来自当前工作树，含此前存在的未提交包号代码；不修改、提交或推送其文件，不把本机完整工作树镜像当限定产品发布快照。只更新原应用、保留原精确 CORS 叠加层，不删除旧产品容器或数据；部署终态以后续记录为准。

- **2026-09-30 家族视觉功能已提交推送（最新发布状态）**：功能提交 25071d4d547c69d6c933bbcb0c276585b23b4b68 已推 origin/main，git push 与 ls-remote 核对远端 SHA 一致。16 个文件包含工坊等宽三图／置顶／最新人像、皇冠图标、SynaRoute 公开官网／下载链接及文档／回归；实际暂存快照 72 项及类型／lint／构建通过，秘密模式扫描无命中，git diff --cached --check 通过。24 个包号并行文件和三条专用主题规则保留未提交，不代用户部署生产或重建 8080；本机 4184 源码预览不等于新镜像已发布。
- **本轮远端构建取证（收工时未完成）**：仅核对功能 SHA 25071d4 的工作流，GHCR 36736764218、CI 36736764231 仍 in_progress；安全扫描 36736764322 已 success，CI 的 frontend、golangci-lint、shell、release-helpers 已 success，test 仍在 Unit tests，镜像处于 Build and push，不标全绿或镜像已可拉取。证据 output/family-polish-release-20260930/{runs,jobs-36736764218,jobs-36736764231,jobs-36736764322}.json；三项 MAGIC_* Repository Variables 名称只读查询仍为空，生产配方／工坊入口仍待地址配置和重新构建。预期功能镜像标签为 sha-25071d4，必须从本轮成功任务日志确认实际标签／摘要后再更新，不沿用 321fe99 发布结果；后续交接 skip-ci 文档提交不生成替代功能镜像。

- **2026-09-30 家族视觉更新再次获准提交推送（最新授权）**：用户要求“提交并推送代码，我更新下”。本次仅发布工坊三张等宽图片及置顶交互、最新深发参考人像、家族首页皇冠图标、SynaRoute 官网／下载链接和相关配置说明／回归；工作区包号专项前后端及三条专用主题规则继续隔离，不混入本轮提交。不新增跨站 SSO、统一退出或配方／工坊一体化托管；主站家族首页与 API 控制台共用当前来源登录，8080 与 4184 是不同来源的本地运行／源码预览。发布验证与远端 SHA／CI／镜像以本轮后续记录为准，不沿用 321fe99 的成功结果，不代用户部署生产。
- **实际暂存快照验证**：本次限定 16 个文件，使用 git checkout-index 导出独立暂存源码，复用现有前端依赖；72 项／8 文件（家族相关 69 项＋国际化 3 项）、vue-tsc、相关 ESLint、Vite 生产构建全部通过。日志 output/family-polish-release-20260930/{tests,types,lint,build}.log；24 个并行文件 SHA256 前后一致，专用包号配置新文件未进入快照，主题仅剩三条包号规则未暂存。未重建本机 8080、部署生产、登录或调用收费模型；原有 Browserslist／chunk 警告不在本轮处理范围，完整后端 CI 仍需核对本轮 SHA。

- **2026-09-30 SynaRoute 免登录官网／下载入口（本地源码，未发布）**：用户要求在家族展台增加下载或官网链接，也可提供未安装引导。本轮选择确定性的公开链接，不新增安装检测或改动原密钥导入：FamilyHomeView的SynaRoute面板在原配置动作下增加“下载客户端”／“访问官网”，匿名与登录用户均可点击，新窗口且noopener noreferrer、不携带Key／JWT、不自动下载或安装。新增family/synarouteLinks.ts集中官方站点https://synaroute.mofamilys.com与/zh/download；来源为真实工程D:/ccfile/SynaRoute的README、site/src/config/site.ts及App.tsx下载路由，不采用搜索结果中的同名第三方。KeysView原导入和100ms焦点提示保持未改，不能把其启发式当真实安装检测；此轮无新上游接缝。样式仅theme.css，原品牌紫色／配置登录／SynaRoute关系图和工坊三图保留。
- **SynaRoute 入口取证及未知边界**：69项／7文件回归通过（此前57项＋home新增1项＋synaRouteImport11项），相关ESLint、vue-tsc与Vite构建通过，证据output/synaroute-links-20260930/{tests,lint,types,build}.log。4184实际匿名1280桌面／390手机两条链接可见、均44px高、无横向溢出，键盘Tab到官网链接focus-visible及2pxoutline；截图synaroute-links-{desktop,mobile}-20260930.png与browser-evidence.json，临时viewport reset，交付tab9保留。网站在线访问本轮未成功：web工具不可访问／GitHub cache miss、隐藏官网页超时导致工具重置、两条HEAD各12秒超时；来源已确认不等于官网此刻可达，不能说安装包下载或真实客户端导入已验。本轮未登录、调用收费模型、修改SynaRoute工程、提交推送、重建8080或部署生产，保留包号及此前并行修改。

- **2026-09-30 工坊三图与参考人像终态（本地源码，未发布；覆盖下方雨夜两图及金发版本）**：用户要求保留原玻璃瓶和狐狸、新增第三张人物，继而否决金发欧美人物，最新以用户提供的深发色室内写真为视觉参考。内置 image_gen 生成虚构28岁成年女性：深棕长发、安静表情、米白吊带／淡蓝灰细条纹衬衫／牛仔裤、自然坐姿，手放膝上；不使用裙装、摸头、截图黑框或尺寸徽标。新增studio-reference-portrait.png／webp／provenance.json，原PNG保留，网页WebP1122×1402、113416字节，精确prompt与参考边界已登记；不代表真人身份、代言或模板实测。此前被否决资产移入ignored output/studio-floral-20260930/iterations，不混入发布；原瓶／狐副本与magic-studio原thumb的SHA256逐一一致。
- **最新尺寸纠正与交互**：用户指出另外两张过小、不协调，theme.css中仅工坊列宽改为40%文案／60%图区，三图统一4:5等宽正常流Grid，轻微±2°与±8px错落，默认互不遮挡；手机去除图区多余最小高度。三张button保留悬停轻抬／回正／1.1倍并置顶、点按单选／恢复，键盘焦点层级高于悬停，reduced-motion关闭过渡。FamilyProductVisual仅custom组件、无新上游业务接缝；caption为“AI 视觉示例 · 非模板实测”。原首屏、品牌、登录／入口、模型功能与包号并行修改保留，没有触发网站模型请求。
- **本轮取证**：57项／6文件通过（home11、products17、switcher8、header3、upgrade8、AppSidebar10），相关ESLint、vue-tsc与最新Vite生产构建通过，日志output/studio-reference-20260930/{tests-final,lint-final,types,build-final}.log。4184实际源码浏览器1280桌面三图约178.46×223.07px、390手机约84.66×105.82px，三对默认相交面积均0、三图均加载、页面无横向溢出；每张实际鼠标悬停z20且pressed=false，玻璃瓶Tab焦点focus-visible/z30，手机尺寸点按true→false。手机尺寸不是实机触摸验证，减少动画仅源码契约测试。证据browser-evidence-final.json及本线程studio-balanced-{desktop,mobile}-20260930.png；临时viewport已reset，交付预览tab7保留。此源码预览设置请求仍回退Sub2API，不当原品牌配置、真实登录或生产验收；未提交推送、重建8080或部署生产。

- **2026-09-30 家族首页侧栏图标（本地源码，未发布）**：用户指出家族首页与仪表盘四宫格重复，已新增custom/components/FamilyHomeIcon.vue的1.5px皇冠轮廓（currentColor、aria-hidden），AppSidebar仅新增带[CUSTOM]的import并将共享buildSelfNavItems内`/family`条目改为该组件，用户端与管理员个人区均复用；两类仪表盘仍DashboardIcon，菜单、路由、登录及主题未改。upgrade-contract新增图标区分回归，55项／6文件通过，相关ESLint与类型通过，证据output/family-icon-20260930/{tests,lint,types}.log；UPGRADE与CUSTOMIZATIONS登记接缝。未提交推送、重建8080或部署生产，没有真实登录侧栏视觉验收；此前工坊人像本地改动与包号并行工作保留。

- **2026-09-30 工坊人像与地址配置说明（本地源码，未发布）**：按用户要求仅替换首页工坊左侧玻璃瓶为既有虚构成年女性雨夜写真，右侧纸艺狐狸及非模板实测说明保留。新增custom/assets/studio-rain-portrait.webp为magic-studio/assets/rain-portrait-thumb.webp原样副本，SHA256均B030EEF9067933F921A8580053620610D7C5C46F31167487544D98DF826D5B40；原生成来源PROVENANCE.json，未新增模型调用、图片编辑或真人代言。FamilyProductVisual仅改import／src／alt，family-home加素材与懒加载断言，SURFACE登记来源，没有改主题或认证业务。FAMILY_PORTAL新增GitHub Settings → Secrets and variables → Actions → Variables的具体操作、三项Repository Variables示例、精确CORS与手动重建／拉取边界；未操作实际变量、DNS或服务器。44项／5文件定向回归、相关ESLint、类型与Vite构建通过，日志output/studio-portrait-20260930。4184为源码预览（自启session42274），1280桌面／390手机浅色无横溢出、人像完整可见、两图加载；截图studio-portrait-{desktop,mobile}-20260930.png与browser-evidence.json留证。该预览公开设置请求失败，首屏回退默认站名，不能当现有品牌配置、真实登录或生产验收；未重建8080或提交推送，保留包号并行工作。此前321fe99镜像不含本轮人像与说明。

- **本轮镜像已发布／集成待完成（覆盖下方构建中状态）**：GHCR36712632118现success，已从成功任务109877820940日志核验真实标签ghcr.io/engineemomo/sub2apicust:sha-321fe99及摘要sha256:c8dd5b34edacf29eec9e30f53f35f5de8c5898706b86715b81d70b8a73db4147，证据output/family-release-20260930/image-publication.json。本轮CI36712632123的Go Unit已success，Integration tests仍in_progress；frontend／Go lint／shell／release-helpers及安全扫描成功，不标全部CI通过。功能提交仍321fe99；随后仅交接文档skip-ci提交，不产生替代镜像。生产仍由用户备份后更新，不建议在集成未完成时直接更新latest。包号专项继续保留未提交，三个MAGIC_*变量未设、独立产品托管／入口仍待。

- **2026-09-30 产品代码已提交推送（发布状态优先）**：功能提交321fe9943b700384b10c2fc1368ed24c6b90eb97已推origin/main，git push与ls-remote核对远端SHA一致。内容为家族展台／控制台切换、配方模型设置／登录配置桥接、工坊独立源码与接入教程，共141文件；未确认合并的包号专项前后端、专用测试／文档及3条主题规则仍保留未提交。实际发布快照285项通过及类型／相关lint／Vite见下一段；本机94dba406镜像包含此前并行工作，不是本次限定范围的GHCR镜像。新提交GHCR36712632118在Build and push，CI36712632123在Go Unit；远端frontend／Go lint／shell／release-helpers及Security Scan36712632091已成功，未标全绿或镜像可用。记录output/family-release-20260930/{runs,jobs-36712632118,jobs-36712632123,jobs-36712632091}.json，不沿用旧43b715a结果。仓库MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN／MAGIC_STUDIO_URL均未设置（只读API核对名称），生产入口将禁用；独立页面仍需单独托管。未操作生产、域名／Variables或用户数据。

- **2026-09-30 本轮产品提交授权**：用户要求“提交并推送代码，我先更新一版看看”。按当前会话默认整理家族首页／产品切换、魔法配方、魔法工坊及接入教程的完整依赖；已询问是否同时合并包号专项，未确认前不混入其前后端运行时修改。原站镜像不自动托管独立产品，生产公开URL／精确origin仍须配置；不把本机4178／4179注入生产。推送与本轮CI状态以随后取证记录为准，不沿用旧镜像结果，生产仍由用户操作。
- **实际暂存版本验证**：141个产品相关文件，主题仅暂存产品／教程部分，包号3条专用规则及其前后端仍保留未提交。git checkout-index导出独立暂存快照，复用现有frontend依赖，无新安装；210项原站／19文件、54项配方、21项工坊通过，共285项，类型／相关lint／Vite生产构建通过。日志output/family-release-20260930/staged/{frontend-tests,frontend-types,frontend-lint,frontend-build,recipes-tests,studio-tests}.log，不能沿用整个工作树252项作为该发布版本计数；基础镜像验证仍等待本轮CI，未重新部署本机或生产。秘密模式扫描未命中，git diff --cached --check通过；3个配方源仅去除末尾空行。

- **互动展台最终修复终态（2026-09-30，覆盖紧随段落的初次a514镜像／251项／审查待完成状态）**：本轮用户再次否决旧卡片，已重做产品展台而非沿用旧ship。三项修复为配方浮签正常流不遮动作、工坊图区与caption分离、页头完整工具区保宽且空间不足产品坞第二行。family-header改为真实公告／语言／订阅组件，仅账户／数据／国际化模拟，不再以文本替身验收。252项／22文件回归通过（家族44项、i18n3项已包含，夹具另跑3项不重复计数），类型／相关lint／Vite通过；Docker内前端及Go embed成功，未全量Go单测／本轮CI。最终8080镜像与容器一致94dba4064ba6d382772e6cc90630cca85a56d315e253715f05142bf098a0c278，入口index-BSFHhB79.js；四容器healthy、health=ok。原数据库／Redis ID、app_data、users3／keys0／groups3／plans1／orders0／dedicated0／280项设置与摘要均不变；CORS及匿名保护通过，配方4178仍f74233cb565c，工坊4179原服务未动。证据output/family-deck-20260930/{regression-final,types-final,lint-final,build-final,header-fixture-final,docker-build-final,docker-deploy-final,verification-final}.log及verification.json；原604527字节备份／SHA256 A0A683397FF0DD821ECE58F6F0D65FFD623C91BEBC0A362A18C18E5F25804444与before-family-deck-20260930-175621回退标签保留。
- **最终视觉／审查范围**：同路径覆盖七张family-deck-*-20260930.png均打开检查。实际8080首页1440／1280深色、390浅色无外层溢出，方法单／浮签不遮字、叠图加载并与caption分离；完整组件夹具1600／1280／1024／390实测外层无横溢出，1600页头141px产品坞第二行、390坞client364／scroll496。实际点击配方CTA到8080/login?redirect=/dashboard?product=recipes后返回，未输入凭据；API与SynaRoute预览运行，紫色191/169/243，手机ArrowLeft／focus-visible与详情展开验证。browser-evidence-final.json含几何及空warn／error日志。首位审查长时间未完成已向用户披露、关闭并换人，中途记录review-first-interim.md不算完整报告；Mendel完整初审fix，两项与首位手机遮字合并三项，最终Verdict三项均resolved、disposition: ship，仅限该修复清单，非全产品／真实登录／用户视觉或生产批准。报告review-{initial,verdict}.md。Godel文档继承检查No changes，documentation-check.md明确仅源码与既有规则匹配，未读取DESIGN／sidecar，不扩大其范围；母品牌未重写，检测器仍不可用。最终用户页tab22保留8080/home配方／深色、默认1280尺寸；只清理本轮17–21测试页与4182／4183两自启进程，用户配方材料页不刷新。未提交推送、付费调用或生产部署，保留并行工作；真实账号与模型授权待用户验收。

- **2026-09-30 产品入口再次优化（优先于下方旧卡片版）**：用户明确“效果太丑了、入口也不新颖”，上一轮修复评分不等于用户批准。保留原BrandPanel首屏／狮冠M／雾钛青／默认深色，下面由长篇卡片改为四产品互动展台；默认配方，产品标签只预览、不登录或请求，当前面板才提供真实进入动作，已知锚点优先，左右／Home／End及循环焦点、手机两列。新增FamilyProductMark，控制台继续四个直接入口并用相同标识，SynaRoute保留紫色；登录先进入API、显式授权及独立产品边界不改。长篇登录说明移入可展开详情，结构示意与非实测图例说明仍可读屏。
- **本机迭代证据**：251项原站回归通过（含家族43项、i18n3项，导出夹具2项与Docker3项不重复计数），类型／相关ESLint／Vite生产构建通过。首次pnpm因短路径临时目录权限失败，改为进程内工作区TEMP与直接Node；初次直接Vite缺node_modules/.bin路径，修实际进程PATH后构建通过，没有安装依赖。Docker代理授权200、原库备份604527字节／SHA256 A0A683397FF0DD821ECE58F6F0D65FFD623C91BEBC0A362A18C18E5F25804444，回退sub2apicust:before-family-deck-20260930-175621；仅应用重建，8080镜像与容器一致a514f4628b29275bf761a9e8ab152e806dfe40a91427a91b18b5ed69f668caef，入口index-rQ0oFKaS.js。四容器healthy、health=ok；原依赖容器ID／app_data／业务计数／280项设置摘要不变，配方仍f74233cb565c，CORS与匿名边界通过。证据output/family-deck-20260930；新的完整独立审查及Docker最终视觉尚在收尾，不能沿用旧ship或宣称用户视觉批准。检测器本次独立尝试退出127，engine未安装且缓存不可写，没有安装或宣称通过；未真实账号／付费模型、提交推送或生产部署，保留并行业务改动。

- **首页修复复核终态（2026-09-30）**：独立审查者Kuhn对原五项P2逐项评分均resolved，disposition: ship，范围仅该修复清单，不是完整产品／真实登录闭环／用户视觉认可或生产批准。报告output/family-layout-20260930/review-{initial,verdict}.md；最终Docker与371项定向证据见紧随段落。Chandrasekhar文档继承检查为No changes，报告output/family-layout-20260930/documentation-check.md；既有母品牌规范未重写，检测引擎不可用，不宣称检测通过。最终用户页tab15保留8080/home深色，恢复默认浏览器尺寸；该页warn／error日志为空。只关闭本轮有效空白测试页10／12／13／14与停止4181布局夹具服务，没有刷新配方材料页；临时错误页11因工具URL限制未手动关闭，交由默认临时页清理。

- **2026-09-30 用户确认两位成员都选全仍保存失败（最新现场反馈）**：不再把本次问题归因为成员漏选，不要让用户重复勾选。只读取线上/login及其公开JS：HTTP200，入口index-BM8dZCAf.js引用AdminDedicatedAccountsView-gZZSIQhx.js；管理页包含user_ids，但没有expected_updated_at、config_issue或本轮多组诊断键，说明该前端尚未包含当前未发布修正。Git HEAD 70271f0aa所含包号SQL明确拒绝账号关联其他分组，即使全部成员选中也会拒绝多组账号；当前源码已修且真实PG双成员多组场景通过。保存校验直接读事务内PostgreSQL，不读Redis。没有管理员认证或生产数据库访问，不能凭前端资源确定线上后端SHA／实际失败项。公开资源证据output/dedicated-config-20260930/online-frontend.json；本轮未修改业务、未提交推送或操作生产，保留并行改动，需前后端同版本发布后再验。

- **2026-09-30 品牌首页纠正与控制台顶部产品切换（最新有效方向）**：用户否决账户摘要＋目录首页，要求原品牌首屏保留、下方介绍附属产品、登录默认先进入魔法API控制台并在顶部切换。FamilyHomeView现原样复用BrandPanel标志／文案／三CTA，新增API通栏及配方／工坊／SynaRoute介绍和真实图例；普通登录／注册恢复原默认落点，匿名产品仅携带站内/dashboard?product=id，控制台显式继续打开，不自动跨域或弹窗。AppHeader仅在user存在时装配FamilyProductSwitcher，原账户／公告／移动菜单保留；所有原站视觉仍仅custom/theme.css。工坊独立内容首版已完成，4179原Node预览可打开，不再说仅规划；SynaRoute仍只是密钥页主动配置，不等于桌面SSO。
- **本机部署与审查终态**：已保留备份output/family-layout-docker-20260930/database.dump（603403字节，SHA256 B082162B87EC2C7804B444186746EC475C49888CE711BC0B06BA1C0929A35657，pg_restore目录及副本哈希验证），回退sub2apicust:before-family-20260930-164415；两次代理授权200，实际构建进程临时代理、Docker内前端及Go embed成功，只更新应用。最终8080镜像／容器一致d02c6ac5689c263ca970768a7e461eb378d905c712b339251bb4e738617d2484，入口index-B2Aw5OM6.js；配方4178仍f74233cb565c，四容器healthy、health=ok。库／Redis ID与app_data保持，users3／keys0／groups3／plans1／orders0／dedicated0／settings280及设置摘要无变化；4178两来源CORS204、未知来源403、真实个人接口匿名401。独立完整UI审查判fix的五项P2已修：keys保持API当前态、原M按日夜accent作mask、交接登记、figcaption读屏说明、brand-focus浅色键盘焦点；修复已部署，最终复核五项resolved，ship仅限修复清单。源码镜像包含工作树已有并行包号／教程代码，未回滚其他任务；构建成功不等于这些业务已真实验收。
- **最终验证及边界**：317项原站定向＋54项配方回归通过，共371项（Docker另跑3项i18n已包含，不重复计数），vue-tsc／相关ESLint／Vite生产构建通过。最终8080实际浏览器1280深色／390浅色无外层横溢出，原创图例加载且真实性说明进入AX树；API节点日夜颜色实测161/217/206与9/107/104，浅色导航键盘focus-visible实测9/107/104、2pxoutline（登录按钮继承原深色outline）。顶部切换为真实组件静态夹具1280／1600／390，1600同排高度75，手机胶囊client364／scroll453、页面390；明确模拟账户，非真实登录。最终五张截图为本线程visualizations的family-restored-*-final-20260930.png及family-switcher-fixture-*-final-20260930.png。先前4175验证匿名产品入口登录目标，后续该独立Vite端口停止／拒绝连接，未重启或杀其未知进程，最终改用实际Docker验收。无真实账号／Key输入或收费模型调用、无通用跨产品SSO／单点退出、无全量Go单测或本轮CI；生产未动、未提交推送。旧备份保留，下方账户摘要首页、登录返回/family、工坊仅规划以及8080未更新均为已被本轮覆盖的历史状态。工程日志output/family-layout-20260930-{regression,types,lint,build,recipes}.log，备份／build-final／deploy-final／verification在output/family-layout-docker-20260930。

- **2026-09-30 多专属分组规则修正（最新用户要求，源码完成未发布）**：用户明确账号关联多个专属组正常，并确认用于同一批包号成员；此要求覆盖此前“一账号只能关联一个组”的历史规则。保持每账号一条有效包号及主要绑定组，允许关联多个同平台标准专属组；各组仅含此账号、无备用路由／订阅、无名单外授权，保存检查所有关联组的非成员未停用Key。成员需授权主要绑定组，其他关联组按需授权同批成员，用户仍可使用无关公开组，此包号账号不进入公开池。新增custom_dedicated_groups.go，保存／配置健康／额度视图统一新结构校验；Key认证、调度／转发／WS复用Check按实际Key组识别关联绑定并检查授权，不只移除保存条件。移除成员事务撤销原账号所有关联专属组授权并停用对应Key，保留无关组；撤销／到期／删除仍阻断全部关联组，旧隔离标记保留，无迁移。
- **截图与证据边界**：用户截图显示两位用户都勾选同一标准专属组且公开组默认可用，截图不能证明弹窗保存成功、包号实际提交user_ids或账号关联数据。已在隔离PG复现：两位已授权但包号只选一位时，另一位被判名单外授权；同一条包号同时选择两位可保存，无关公开组授权不影响。新失败文案提示“同批共用者请全部加入同一条包号”，仍不擅自自动增加成员。待用户确认所指是用户授权弹窗保存还是包号绑定保存，线上根因尚无现场数据，不宣称已修线上。
- **最终验证**：46项前端定向（含原用户授权弹窗4项）、相关ESLint、vue-tsc通过；独立Docker包号service／handler测试及Go embed编译通过。临时PG18真实执行30个配置场景及原多人回归全部通过；多组场景额外验证两成员两组Key认证／选号、错账号／非成员／未授权组拒绝、改公开组与名单外授权拒绝、移除成员清两组且无关组Key保持、换主要组历史标记、撤销／删除／重分配。日志output/dedicated-config-20260930/{multi-build,multi-postgres}.log，临时容器清理；未部署8080／生产、未提交推送，保留并行家族等改动。

- **2026-09-30 线上包号保存400诊断（本轮源码，未发布）**：用户报告已配置专属分组与授权仍返回DEDICATED_ACCOUNT_CONFIG，确认发生在线上，尚未提供account_id／group_id／完整user_ids或服务器只读访问。源码核实Save在串行化事务内直接查询PostgreSQL，完整性校验不读取Redis或调度缓存；不能仅凭通用错误断定线上具体失败项。新增custom_dedicated_config.go：失败后在同事务读取19类条件中的首个失败项，返回白名单config_issue及最多10个资源编号；保留原错误reason与独占判定。handler保留具体错误元数据，管理页按白名单诊断显示中英文失败原因／编号，旧响应回退原文案，不直接展示原始错误或凭据。未调整授权或放宽隔离规则，无迁移。
- **本轮验证**：42项包号前端回归、相关ESLint、vue-tsc通过；独立Docker镜像包号service／handler定向测试及Go embed编译通过。临时PostgreSQL18（无宿主端口、无业务卷）实际执行23个新增保存配置场景及原多人回归通过：完整配置可保存，分组名称不替代专属开关、关联其他组／账号、备用路由、非成员授权／未停用Key、缺成员授权等正确拒绝并给出具体编号，inactive／disabled旧Key场景正确通过。证据output/dedicated-config-20260930/{build,postgres}.log；测试容器已清理。未连接生产数据库、未部署、未提交推送，保留其他线程现有改动；线上根因仍待编号与现场配置证据。

- **2026-09-30 魔法家族统一首页（最新入口范围，源码与本机预览完成未发布）**：用户要求首页登录后跳转／打开家族产品且避免重复登录。按现有魔法 API 账号复用实现/family，原默认/home（无home_content时）与/brand装配新首页；原API介绍迁至/api。官网导航与控制台个人区新增家族首页。首页为账户摘要与四产品目录：API直接本站控制台，配方新窗口打开，SynaRoute进入原密钥页主动导入，工坊标为暂名规划、禁用入口。沿用狮冠M、雾钛青及默认深色，原站样式仅custom/theme.css，不改认证页、计费或任何后端权限。
- **本轮登录与发布边界**：站内沿用authStore有效会话，配方选择配置复用原站/connect/recipes；没有新账号系统，不向产品链接发送密码、JWT、刷新令牌或Key，不自动创建密钥／授权／调用模型。SynaRoute并未实现桌面账号SSO，全产品单点退出也未实现；原站退出后配方页已授权Key须清除或刷新。生产VITE_MAGIC_RECIPES_URL未配置时禁用入口，只有本机DEV默认明确4178预览；api_site仅公开origin提示，配方用原官方／本机白名单预填，本页底部新增新窗口返回家族首页，保留材料。Dockerfile已新增URL／ORIGIN两个ARG及前端构建注入，custom-image.yml从仓库Variables接线；覆盖下方咨询“尚无构建接线”的历史结论，但尚未设置实际变量、构建Docker镜像或运行本轮CI，配方静态产物仍需独立发布。
- **本轮验证证据**：237项原站定向回归（含家族20项、路由45项、原定制及国际化3项）与54项配方Node回归通过，共291项；vue-tsc、相关ESLint、Vite生产构建及workflow YAML解析通过，日志output/family-20260930-{focused,regression,recipes,types,lint,build,workflow}.log。额外来源测试确认公开配方链接不能自动预填本机登录目标，本机来源提示仅限本机HTTP预览。浏览器4175验证/family与默认/home、匿名登录回跳/family、/api原介绍保留；首页1280深色／390浅色实测无外层横向溢出。4178新空白测试页预填本机4175、返回链接安全属性、接入等待／取消、390新增登录设置布局通过，两页错误／警告日志为空。截图family-home-desktop-dark-20260930.png、family-home-mobile-light-20260930.png、family-recipes-connect-mobile-20260930.png在本线程visualizations，首页tab6已留给用户；临时视口复原、主题恢复深色。配方最终构建136797字节并重启自己服务session86473，未刷新用户旧材料页。登录态切换由模拟组件／守卫测试证明，未输入真实账号／Key，真实授权跨窗口闭环、独立产品账号SSO、CORS及收费调用仍未实测。Vite4175原服务保留，8080镜像／生产未更新，无提交推送；其他并行工作未回滚。

- **2026-09-30 魔法工坊本地首版（后续授权，定时关闭）**：用户从「先规划」推进为「先做产品，定时任务先不启用」。独立product-samples/magic-studio已实现8图片／4视频提示词／2 Remotion Skill／2 ComfyUI模板指引，共16条及4个组合玩法；搜索／媒介／用途／资源类型、排序、材料替换／中英模板／复制／TXT／ID分享、收藏／最近浏览／日夜主题／大图／视频运镜播放暂停。8张内置image_gen原创视觉素材及原狮冠M，25个PNG／WebP来源已只读核对通过，精确prompt见assets/PROVENANCE.json。视频是二维示意、模板未目标模型实测、Skill未安装运行，不以视觉图代替模板效果证据。无模型API、自动搜索或调度；页面CSP connect-src none、仅白名单静态文件。
  - 本地4179运行（serve.mjs，session77912），最终21项Node／jsdom／HTTP定向通过，实际浏览器复制替换材料／ID分享、TXT落地内容、收藏重开、空搜索恢复、大图Esc、视频播放暂停已验证；320／390／768／1280／1440／2560宽度无外层横向溢出，截图／测试／素材核对／下载证据在该目录.impeccable/review。部分全页导出及快速跨尺寸截图有工具失真，Skills／320／视频以补录实际视口及DOM几何为准；设计扫描引擎不可用exit127，不报告扫描通过。独立审查首轮fix：资源类型跨区匹配、手机返回列表保留材料、dialog名称三项P2已修；同审查者Verdict Pass为ship、三项均resolved，仅覆盖这三项，不扩大为全产品或生产批准，记录finish-review.md／finish-verdict.md。DESIGN.md与.impeccable/design.json已归档实际token、组件、断点及来源，不重定母品牌；最终用户窗口1280×720恢复深色首页、图片全加载、无横溢出，preview.png为实际视口。未改中转站／配方／SynaRoute运行时代码或配置，未Docker重建／提交推送／生产部署，保留所有并行未提交工作。

- **2026-09-30 魔法 API 到配方入口核查（只读源码，未新增入口）**：用户询问魔法 API 是否已有跳转到配方的入口。rg检查frontend/src及官网PublicLayout、BrandHomeView、控制台AppSidebar，配方相关命中只有登录选择页、辅助与测试，没有本轮新增的官网／侧栏产品跳转入口。现有/connect/recipes是从配方发起的登录后配置选择页，不是配方首页。线上管理员自定义菜单配置未读取，不能宣称线上配置一定不存在。建议部署地址确认后在官网导航与控制台个人区各加“魔法配方”入口、新窗口打开；本轮仅回答现状，没有修改导航、站点配置或部署。

- **2026-09-30 魔法配方部署咨询（仅方案，未执行）**：用户询问发布部署方式。源码确认配方由build.mjs生成单一index.html，serve.mjs仅本机预览；现有Dockerfile构建原站，不包含配方静态发布。建议独立静态入口复用现有魔法 API 登录与服务，可选子域名（recipes.mofamilys.com仅候选）或同源/recipes/路径，未替用户选择／配置域名。发布须分别托管静态文件及更新原站/connect/recipes；独立来源须构建时VITE_MAGIC_RECIPES_ORIGIN与运行时API CORS双许可。查证Dockerfile无该ARG、custom-image.yml未传该build-arg、.dockerignore排除.env.*，因此子域名方案正式发布前尚需接构建链路，不能给运行容器加变量冒充生效。已查官方Nginx、Vite与MDN来源并更新site/README部署说明；没有改构建链路、域名、服务器、安全配置或生产，也没有提交推送／模型请求。站点管理工具及反向代理配置未知，待用户提供。

- **2026-09-30 魔法配方登录魔法 API 选择配置（最新范围，源码完成未发布）**：用户要求在自带模型设置之外，登录魔法 API 直接选择配置。独立4178页新增登录入口和手动连接并列；原站custom新增受保护/connect/recipes，复用既有登录／2FA／passkey成功后的站内回跳，不在配方页收集账号密码。登录后按本人keysAPI分页读取已有密钥，过滤停用／过期／额度耗尽／无效分组，生图排除明确禁用分组；用所选Key只读GET模型目录，用户选模型与文字格式、核对接收来源并显式授权后回传所选API连接。账号密码、登录JWT不出原站，不自动创建／改绑密钥，不修改权限或计费。
- **本轮安全与发布边界**：生产接收来源默认同源，独立配方需原站构建环境VITE_MAGIC_RECIPES_ORIGIN精确允许；本机源额外支持4178。配方登录目标仅官方ai.mofamilys.com或本机HTTP，其他服务仍手动接。回传严格匹配窗口／origin／32字节随机nonce／用途／五分钟有效期，单次接受且有确认消息；不使用通配targetOrigin，密钥不入URL或持久化。登录和导入不自动调用生成模型；视频仍复制使用。窗口被COOP隔离时提示重试或手动配置，不绕过安全头。新增原站样式仅custom/theme.css专属作用域，无新增上游业务接缝或依赖。
- **本轮验证证据**：51项独立Node测试及99项原站定向测试通过（新增选择页／辅助25项，99项含国际化3项），类型／相关ESLint／Vite生产构建通过，日志output/recipes-login-20260930-{node,vitest,regression,types,lint,build}.log。自动测试覆盖允许来源、错误窗口／nonce／用途／过期／重放、显式发送且无登录令牌、文字／生图导入、分页过滤、迟到目录与接收确认；浏览器4178新增入口和1280桌面无外层溢出，4175匿名选择路径进入原登录并保留完整站内目标，配方页警告／错误日志为空。390视口设置工具没有改变新测试页实际DOM宽度（仍1280），因此新增手机控件未实测；不沿用上轮手机证据。未用真实账号密码／Key，登录后真实选择及跨窗口导入闭环、实际CORS／权限待用户验收；没有付费生成请求。本机4178重启到新构建（session18607），另开空白页保留给用户；用户旧页未刷新。只在Vite4175验证原站新源码，8080镜像与生产未更新，未提交推送，其他并行改动保留。

- **2026-09-30 第四产品规划（本轮仅方案）**：用户提供创作工坊截图，要求第四款产品增加图片／视频分类及定期发现提示词、Skills和新玩法，并明确「先规划」。新增product-samples/magic-studio/PLAN.md，暂名「魔法工坊」，建议以内容创作者优先，独立媒介与资源类型分类，效果预览→改变量→复制／复用；16条候选＋交互原型、约40条精选首版、采集审核、站内图片、异步视频分阶段。已网页读取核实Google图片／视频、Runway指南、ComfyUI模板及Remotion Skills来源存在；未收集完首批条目或实测。每日09:00发现／周五18:00精选仅为Asia/Shanghai频率建议，未创建自动化；优先人群与首版生成深度已询问，暂按建议假设，名称／托管／接口／预算待定。未改业务、主题、模型配置、Docker或生产，未安装执行外部Skill、调用模型或提交推送；保留既有未提交任务。

- **2026-09-30 魔法配方自带模型接入（覆盖历史“不调用模型”产品边界，独立本地原型）**：按用户要求新增顶部页内“模型设置”，文字／生图分别配置API地址、Key及模型名。文字兼容Chat Completions／Responses并支持多轮追问，生图兼容Images API返回base64／URL图片；视频仍复制使用。整理及应用配置不发送请求，运行／追问才直接向所填接口发送提示词与当前对话；密钥与配置仅当前页内存，不存储、不写URL、不读取中转站业务配置。CSP开放HTTPS及指定本机HTTP，API请求不带Cookie、不跟随重定向、不自动重试；修改材料／切换配方／更改当前连接清除对话并中止等待，迟到结果隔离。文字120秒／生图180秒，停止等待不能保证服务端不计费。
- **本轮验证**：语法检查、构建及44项core／model／DOM测试通过（日志output/recipes-model-20260930-tests.log，含真实本机HTTP往返）；浏览器用本机4180模拟接口验证Chat Completions、Responses、追问及512px图片显示，生图返回既有Logo，不是AI效果实测。1280深色、390浅色、320窄屏无外层横向溢出，浏览器错误／警告日志为空；截图在本线程visualizations的recipes-model-settings-20260930.png、recipes-model-settings-mobile-20260930.png、recipes-model-run-20260930.png。模拟服务已停止，4178服务运行最终128364字节页面（session64858）；新用户预览留空配置，未刷新用户旧页。未知用户真实服务CORS与兼容性，未调用外部真实或付费模型、未提交推送或生产部署；未修改并行控制台／包号／接入教程业务。

- **2026-09-30 接入教程（本地已完成，未发布）**：新增免登录 `/guide`、登录后右侧正文 `/help/guide`、游客预览 `/preview/guide`，官网导航、首页 CTA 与 FAQ 可直达；覆盖 Claude Code CLI／Local 桌面、Codex CLI／本机桌面、Cherry Studio、Cline、Continue、OpenCode v1、Aider，以及 Cursor BYO Key 限制。补安装入口、终端密钥变量、Codex 用户配置及桌面“使用密钥”／API Key Mode／模型目录步骤，Continue／OpenCode／Aider可折叠示例；地址基于公开 apiBaseUrl，示例仅占位符。上下文链接保留控制台／预览；游客密钥入口支持登录提示和取消。router/index.ts新增限定教程锚点滚动接缝，其他页面行为保留。官方配置文档已在页面列出，特定客户端版本及桌面云端会话需用户实测。与现有未提交包号修复及 product-samples 并存，未提交、推送、重建8080或发布生产。
- **接入教程验证**：80项定向前端测试、国际化3项、typecheck、相关ESLint及生产构建通过，证据output/client-guide-20260930。4175真实浏览器验证游客密钥提示／取消、键盘折叠示例、Codex锚点滚动到章节、预览教程↔FAQ保留一个main与侧栏；1280深色和390浅色无外层横向溢出。截图client-guide-desktop-dark.png／client-guide-mobile-light.png在本线程visualizations目录；Vite仍监听127.0.0.1:4175，可浏览 `/guide`。未使用真实密钥请求模型，未完成所有客户端及登录后控制台带数据验收；本轮不涉及后端，包号验证沿用该专项记录。

- **2026-09-30 包号专项修复已完成源码及定向验证（未发布，覆盖下方待修结论）**：用户授权修复四项。Save仍校验完整配置，但运行时只验证当前成员健康及授权，其他成员停用／软删除／撤权不再拖累全组；退出者旧Key状态不参与运行时全组拒绝，非成员自身仍拒绝。成员移除事务停Key统一inactive（保存兼容旧disabled），APIKeyService.Update新增[CUSTOM]接缝，显式启用及扩额／重置额度／清除或延长到期等隐式激活写入前统一验包号。编辑必传原始expected_updated_at，陈旧／缺失返回409；已撤销恢复必明确reactivate确认。管理页分开绑定状态与配置诊断，提供刷新重开和恢复checkbox；新增样式仅theme.css。严格一账号一组、无备用路由／影子及历史隔离标记均保留，无迁移。
- **本轮证据**：40项前端定向／定制契约＋3项国际化、typecheck、相关ESLint及生产构建通过；Docker最终Go定向service／handler测试及embed编译通过；最终隔离PostgreSQL真实执行65项service顶层测试（含版本冲突、确认恢复、移除事务回滚、同伴停用／删除／撤权、6类Key激活拦截、合法成员Key更新及改绑／删除隔离）通过。证据output/dedicated-fix-20260930（frontend-test.log、frontend-build.log、build-final.log、postgres-final.log），临时PG已清理。无宿主Go，使用Docker编译；未全量Go／本轮CI、未做带真实管理员数据浏览器验收。未提交推送、未更新8080或生产，product-samples不变；生产历史403/503仍不知实际根因。

- **2026-09-30 包号专项审查（优先于下方移除修复已验证状态）**：隔离PostgreSQL复现：退出成员仅改旧Key状态即可让剩余者拒绝访问（active及inactive均复现，未发布修复有此缺口）；停用一个成员连带阻断其他成员；旧编辑payload覆盖后来的撤销并恢复绑定。源码另确认管理页缺配置健康状态。19个既有service测试＋1个四子场景审计测试完成，审计通过指坏行为复现而非修好。报告deploy/DEDICATED_AUDIT.md，证据output/dedicated-audit-20260930；三份被测业务源码哈希与工作区一致。此次仅审查及文档，不新增业务修改，不提交推送部署，不碰product-samples，临时PG已清理。

- **2026-09-30 移除包号成员400修复（未提交／未部署）**：用户报告从两个成员中移除一个无法保存。源码及隔离PostgreSQL证实：旧授权／密钥仍存在时新名单完整性校验返回DEDICATED_ACCOUNT_CONFIG。现Save在原串行化事务内撤销退出成员原组授权、停用该组Key但保留记录和组归属，再校验及保存；失败全部回滚，提交后清组和退出用户认证缓存。已停用的非成员Key不再阻断独占检查，非成员授权及未停用Key仍拒绝，不放宽一账号一组约束；重新加入不会自动启用Key。不触碰余额或其他组，无新迁移。前端提示同步。
- **移除成员验证**：31项前端包号测试、相关ESLint及typecheck通过；Docker Go service/handler定向测试、embed编译通过；19项service顶层测试含真实PostgreSQL成员移除／旧校验复现／回滚／其他组保留／重新加入Key不启用均通过。证据output/member-removal-20260930；仅临时测试库，无业务库操作，8080和线上仍为旧版。未全量测试或本轮CI、未提交推送部署，product-samples不动。

- **2026-09-30 已推送43b715a15**：完整代码SHA `43b715a15f599f5ae2125bc923849d57e4b3522c`，push及ls-remote核实origin/main一致。GHCR36657101655与安全扫描36657101610成功；镜像 `ghcr.io/engineemomo/sub2apicust:sha-43b715a` 已发布，摘要 `sha256:dc36b117e524262faddf0c13f0dad5c9ff510e385e69ddc2860145a7ea6566af`（构建日志确认）。CI36657101544的前端、Go lint、shell、release-helpers和Go Unit通过；8分钟等待窗口结束时Integration仍在运行，不能宣称全部CI通过。证据output/console-release-43b715a；用户发布前应等待该run成功，备份数据库／配置，迁移242后不要直接回滚旧单用户版。生产未操作，product-samples保留未提交。后续交接仅文档[skip ci]提交不产生新镜像。

- **2026-09-30 提交发布收口**：用户确认不放宽包号独占规则，并授权提交推送以便自行发布。保留“一账号一专属组、多成员”以及改绑／删除的旧Key隔离保护。本次提交仅控制台、多人包号、费用显示、测试及维护文档；product-samples与本机配置／备份不纳入。既有139前端定向与Docker Go／PostgreSQL验证及本机部署证据见下文；远端CI和GHCR须查询本轮提交，不沿用旧记录。生产由用户备份后操作。

- **2026-09-30 已部署本机Docker（优先于本轮未部署记录）**：用户授权本机测试。原三容器处于停止状态，保留原卷启动Postgres／Redis，先完成597905字节数据库备份与pg_restore目录校验、旧镜像标签sub2apicust:before-console-20260930，再代理预检200并构建。前端类型／国际化3测／构建及Go embed成功；仅应用容器重建，运行镜像与标签一致be2f26f98e8c3dee62d765a8b78e6cd33021dc3b6f618c76085e310180b234b8，三服务healthy，health=ok。迁移242两列和三个索引确认存在；数据库／Redis容器ID及app_data卷不变，用户2／密钥0／组3／套餐1／订单0／包号0／设置280计数保持。备份逐项比较只有Claude版本2.1.283→2.1.285与Codex版本0.158.0→0.159.2两键自动同步，09:44启动日志证实，其余278键不变。HTTP入口index-Ckx_MqIV.js及三个新页面chunk验证通过，账户与管理员接口匿名401；浏览器8080官网正常打开并留给用户测试，尚未做带数据管理员交互验收。证据output/console-deploy-20260930，备份SHA256=1D900EC14FB9915B5E912F0A84778271CEB149178B7CC526E3A9F55411EE037E。未提交推送或操作生产；迁移后的回退需兼顾数据库，不能直接跑旧单用户版本。

- **2026-09-30 控制台修复与多人包号（未发布）**：完成登录后隐藏官网两处游客入口、控制台FAQ右侧内嵌 `/help/faq`（公共 `/faq` 保留）、管理员用户／账号／组名称与用户端组名；智力效率默认矩阵新增平均任务费用及聚合口径，明确不是本站／Token单价。用户确认多人共享一个账号和专属组、独立Key及余额；包号可选1–100成员、编辑账号／组、撤销后确认删除。242迁移兼容旧单用户，逻辑删除与改绑旧资源保留隔离标记，重新分配仍严格检查授权及非成员Key，不自动修改授权／计费。
- **本轮验证与边界**：139项前端定向测试、国际化3项、类型检查、lint及构建通过；Docker隔离Go定向service/handler测试与embed编译通过；额外在临时PostgreSQL18真实执行18个service测试（含旧数据迁移、多人、移除成员、改绑、删除及新分配SQL），全部通过。证据 `output/console-fixes-20260930/`。未跑全量Go测试／本轮CI，管理页浏览器带数据视觉验收仍待；本机8080业务容器未替换，生产未操作，未提交推送，保留product-samples。
- **线上403/503**：用户无法提供日志；已定位源码触发条件，但两个历史请求根因仍不知道。403由专属校验失败返回，旧版单用户限制与多人授权是可能条件；503既可能来自无可用账号兜底，也有Responses依赖缺失路径。不得宣称此次修改已经解决生产错误。配置规则／具体请求ID／迁移回退注意事项见 `deploy/DEDICATED_TROUBLESHOOTING.md`。

- **2026-09-28 游客发布最终核验（优先于下方待CI状态）**：功能9a31787de及测试初始化修正fa45b2023均已推送origin/main，ls-remote确认完整代码SHA为fa45b2023c01e5c84084dcbcb32d5a7c5bbe213a。该SHA的CI36424788919全部成功（含Go Unit与Integration、前端、Go lint、shell、release-helpers），安全扫描36424789053成功，GHCR36424788938成功；镜像日志确认latest及sha-fa45b20已发布，digest=sha256:69083ada3ce6432eaee2357e57fa384aa0804d125444cc33998e742c601d1d44。证据output/guest-release-fa45b20。新增SQLite测试按现有用例开启外键并限制单连接，已由本次CI验证；本机此前ea6e3a5ecc3c仍含相同业务源码，后续仅测试及文档变化，不为此重复重建本机。用户备份服务器后可在/sub2api-deploy运行./update.sh sha-fa45b20；生产未操作。179项本机定向回归和全量lint通过，product-samples未跟踪文件保留且未提交。本次收口文档使用[skip ci]，不将其当成另一个镜像版本。

- **游客发布CI修正**：代码9a31787de已推送origin/main，CI36423807298前端／Go lint／shell／release-helpers通过，Go单测仅新增TestPublicPlansAnonymousCatalog失败：SQLite外键开关未启用、ent建表报错；Integration skipped。已参照现有PaymentConfig测试开启PRAGMA并将测试连接池限定单连接，未改运行时代码。安全扫描36423807397成功，首轮镜像仍构建中；需要查修正提交对应CI后再给更新版本，不将首轮描述为全绿。

- **2026-09-28 游客功能提交收口**：用户授权提交推送并更新。本轮范围为游客官网／公开套餐／FAQ／游客控制台预览和响应式修复；product-samples保持未跟踪，不混入本次业务发布，共享交接历史保留。代理fetch确认提交前HEAD与origin/main一致；全量前端lint和179项定向回归通过，新增Go测试仅修正gofmt空格。本机8080已运行本轮响应式修复；Go embed此前Docker编译通过，Go单测与远端镜像发布结果必须查本次代码提交CI，当前不宣称通过。生产仍由用户手动更新。

- **2026-09-28 游客预览响应式修复已部署（最新）**：用户截图否决宽屏适配。2031px浏览器复现main限制1280px导致卡片右留白551px；仅theme.css去除工作区宽度上限、设置弹性对称边距、字段等宽列／手机双列、窄屏顶栏隐藏重复徽标以保持操作同排。320／390／768／1024／1440／2031／2560七档实测左右留白对称、无溢出、顶栏操作可见；65项定向回归通过。代理预检与数据库备份后重建本机Docker，前端类型／国际化3测／生产构建及Go embed通过；运行镜像ea6e3a5ecc3c021f3f128725320fcee293751c69dc795cc6ac8fb678b65e1b2d与标签一致、三服务healthy，数据库／Redis容器ID、用户／密钥／套餐／订单计数和设置摘要保持。8080浏览器2031px实测卡片左右边距均40.609px。证据output/preview-responsive-20260928，595135字节备份及回退sub2apicust:before-preview-responsive-20260928保留。未提交推送／生产部署，未下单；视觉待用户再次确认。

- 2026-09-28创作配方更新：新增07-image-prompt.md与08-video-prompt.md，现为工作3份、大学3份、创作2份。主按钮改为“整理并生成提示词”，明确本地模板组装、不调用模型；结果与教程按文字／生图／视频分别引导。构建与语法检查完成，30项core／DOM测试全通过；真实浏览器验证创作筛选2项、两份虚构示例及正确目标工具提示，桌面和390px无外层横向溢出。本机4178服务已重启（session75587），用户旧页未刷新。未调用模型、未验证媒体效果、未提交推送或部署生产；真实系统剪贴板与下载落地仍待验证。

- **2026-09-28 17:52 游客控制台预览已部署本机Docker（优先于下方未部署记录）**：用户明确授权部署查看。代理预检200，先备份数据库并保留回退镜像，再构建；前端类型／国际化3测／构建及Go embed编译成功，仅应用容器重建。运行镜像与本地标签一致`ecb3961845e8580395b31170a8a31bfa533e92710671beddf6c1249ed0594c2e`，三服务healthy、health=ok；数据库与Redis容器ID、app_data卷、2用户／0密钥／3分组／1套餐／0订单／280设置计数及设置摘要全部保持。HTTP入口index-C4UWqsIx.js引用GuestPreviewView-D-VurzQx.js；4个真实个人接口匿名401。浏览器8080实测首页游客按钮→预览、密钥操作登录提示／取消、真实套餐加载。备份和证据output/preview-20260928-175007，回退sub2apicust:before-preview-20260928-175007；数据库备份SHA256为6762F0F26B1ED26CA99867F9F67DCBB828730BC4DDB5D095393C4105742C0FA2。用户可访问http://127.0.0.1:8080自行验收；未下单扣款、未提交推送或部署生产，Go单测／完整登录支付仍待。

- **2026-09-28 新增游客控制台预览（源码完成，尚未进入Docker）**：用户确认首页增加「游客预览控制台」，现主CTA与公共导航均通往 `/preview`；独立只读侧栏包含概览、真实公开套餐、FAQ及五项账户功能说明，不挂载实际账户布局、不读取个人数据。账户操作弹可取消登录提示并保留目标；手机折叠菜单导航后收起。179项定向前端测试、3项国际化测试、类型检查、相关ESLint、生产构建通过；浏览器4175实测首页入口、密钥操作提示／取消、真实套餐、FAQ搜索空态、390手机菜单及1280桌面深浅主题，检查宽度无横向溢出。截图在本线程visualizations目录preview-desktop-dark.png／preview-mobile-light.png。8080仍为上一版e7a582285cbd，不含本轮新预览；本轮未重建Docker、未提交推送或部署生产，保留并行product-samples改动。
- **2026-09-28 配方使用教程**：用户要求补教程，独立原型新增顶部“使用教程”、首页折叠说明及结果区“复制后怎么用？”入口。四步涵盖选配方／填材料／生成复制／到自己的AI工具粘贴发送，另说明核对追问、手动复制、虚构示例、隐私与刷新丢失；明确本站不运行模型。返回按钮保留原配方、材料与结果。构建及27项core／DOM测试通过；浏览器验证大学论文草稿往返保留、焦点与hash不丢，390px无外层横向溢出且无错误日志；截图recipes-tutorial-desktop.png／recipes-tutorial-mobile.png在本线程visualizations目录。仅改site/{template.html,app.js,theme.css,ui.test.cjs,README.md,DESIGN.md}及交接／登记，未触碰其他并行业务改动，未提交推送或部署生产。127.0.0.1:4178已重启到本轮页面（会话54984）；用户原页未刷新，旧页刷新前需保留材料。
- **2026-09-28 游客官网已部署本机Docker（优先于下方源码阶段未编译／404记录）**：用户授权本机部署；desktop-linux本地npipe、代理预检200，实际构建进程配置临时代理，首次构建退出0，前端／类型／国际化3测及Go embed编译通过。仅替换应用容器，新镜像与容器Image一致`e7a582285cbdaeb78412a0414ac4280322c1654c05614fec869cced2c2e13b11`；三服务healthy，health=ok，Postgres／Redis容器ID及app_data卷不变。users=2、api_keys=0、groups=3、subscription_plans=1、payment_orders=0、settings=280计数未变；设置仅`openai_codex_client_version_synced`变化，逐项与备份比对及17:14:59自动同步日志确认，其余279键不变，不宣称全部设置摘要相同。公开目录返回既有1个在售套餐，字段白名单检查通过，4个账户接口匿名401；HTTP入口index-C6kd8wBQ.js引用PublicPlansView-DAcrv8XQ.js与PublicFaqView-Beye-j_3.js。浏览器在8080实测首页、真实套餐、FAQ搜索、购买→登录并保留plan=1；390px套餐无横向溢出。未登录账户执行购买、未下单／扣款，Go单测与完整支付联调仍待。旧镜像`sub2apicust:before-guest-20260928-171257`；备份output/guest-20260928-171257/database.dump共594458字节，pg_restore目录验证通过；同目录含before.json、build.log、settings-diff.json、verification.json、plans-mobile.png。未提交推送或操作生产；保留并行配方与交接改动。
- **2026-09-28 配方自定义确认框**：按用户截图要求，独立原型两处原生confirm已替换为主题内dialog，区分示例覆盖／清空；取消优先、关闭／Esc、Tab双向循环与焦点恢复、异步跨配方保护，旧浏览器不支持时保留数据。node build.mjs及24项core／DOM测试通过；真实浏览器验证模态、取消保留、Esc及焦点循环，桌面深色和390px浅色截图存recipes-confirm-dark.png／recipes-confirm-mobile-light.png，手机弹框边界19至371px，无页面错误日志。浏览器仅使用新测试页和虚构示例，未刷新用户原页或清除其材料；确认执行由DOM测试覆盖。127.0.0.1:4178已重启到新版（会话53940），旧页刷新前需保留输入。仅改site/{app.js,template.html,theme.css,ui.test.cjs,README.md,DESIGN.md}及交接／登记，其他并行的中转站业务改动不碰；未提交推送或生产部署。
- **2026-09-28 游客官网与FAQ实现**：按用户确认方案新增公开官网／套餐目录／FAQ（四类十二题），账户功能弹出可取消的登录提示；套餐跟随后台在售数据，登录／注册／邮箱验证保留选择，支付恢复及原续费优先。后端新增只读字段白名单 `/api/v1/payment/public/plans`，个人接口认证不变；首页自定义内容优先，正常模式默认游客官网。163项定向前端测试通过，类型检查、相关ESLint及前端构建通过；本机无Go，新增后端单测未运行、后端未编译，不沿用旧CI。浏览器实测官网、FAQ搜索、登录提示／取消与1280桌面／390手机深浅色，无外层横向溢出；截图为本线程visualizations的guest-home-desktop-dark.png和guest-faq-mobile-light.png。Vite4175预览进程本轮session73339；8080后端未更新，新目录实测404并就地提示，真实套餐／登录购买联调仍待。详情frontend/src/custom/GUEST_PORTAL.md，接缝见CUSTOMIZATIONS。保留原交接／product-samples改动；未提交推送、未改Docker或生产、未下单扣款。
- **2026-09-28 魔法配方独立原型**：用户要求开始做配方，本轮将已确认的工作3份＋大学3份落成独立本地静态原型product-samples/magic-recipes/site；提供分类搜索、材料表单、示例／方法、完整提示词生成、复制回退、TXT下载入口及双主题。通过构建从现有Markdown抽取提示词，嵌入既有狮冠M；无账号、模型调用、网络接口或持久化输入，无现有业务接缝。node build.mjs成功，node --test core.test.cjs ui.test.cjs共19项通过（DOM测试复用frontend已安装jsdom）；浏览器实测会议／论文生成、必填、大学筛选，1280px桌面与390px手机未见外层横向溢出，日志无错误，深色桌面／浅色手机截图存本线程visualizations的recipes-desktop-dark.png与recipes-mobile-light.png。系统剪贴板落地未由工具证实（虚拟剪贴板为空），TXT下载事件等待超时，保留用户真机验收；不说完整端到端已通过。生成index.html被本目录忽略，可直接打开；127.0.0.1:4178预览进程仍运行（本轮会话74036）。未改中转站、SynaRoute、Docker或生产，未提交推送、未付费调用或创建定时任务；保留并行海报及原有交接改动。最终栈与正式托管未定，本轮静态原型为可逆选择。
- **2026-09-28 大学配方定位修正（优先于此前学生扩展）**：用户明确学生系列只做大学，不做其他学段。04／05／06原文件已整体替换为大学论文精读、编程实验复盘、大学期末复习，版本0.2；README同步，不保留原小学示例作为现行样品。工作系列3份SHA256修改前后一致。Node检查六份章节／围栏、6个链接、大学定位及均值／代码表达式／时间等示例通过；没有执行Python或外部模型试跑，不冒称实测有效。未改业务、创建自动化、调用付费API、提交推送或部署。
- **2026-09-28 海报客服文案更新**：用户明确要求将「微信扫码」改为「客服咨询或相关名词」，双二维码日间／深色海报统一采用「客服咨询」，左侧说明同步为「访问官网 · 客服咨询」。输出仍在with-wechat目录，二维码图案、位置与其余设计不变；此文案决定优先于下方初版标注。不改网站、不发布。
- **2026-09-28 海报双二维码补充**：按用户要求将上传的微信二维码放在官网二维码右侧，仅标「微信扫码」，不推断身份。日间与深色新版均输出到本线程mofa-poster-20260928/with-wechat/{day,dark}/，各含1800×2400 PNG及1080×1440 JPG；原版保留，生成源为add_wechat_qr.py。原二维码字节副本一致、高清微信码区域与合成素材逐像素一致、y<1990区域与旧版一致（脚本断言）；已打开两版分享图目视检查并排标签及排版。尚未手机实扫或验证微信目标。仅更新会话海报及本交接，不改业务、配置、Docker，不提交推送或发布，保留其他任务工作区变更。
- **2026-09-28 魔法配方学生扩展**：用户认为首批内容有用并要求学生场景。新增04知识点理解、05错题复盘、06复习安排三份样品，具体示例为小学高年级数学；强调分轮提示、真实学生回答、来源／冲突、隐私及时间预算，不承诺提分。README更新工作／学习两组及实测门槛。Node检查六份各7项必要章节、围栏、6个本地链接通过；学生例子的平均数、分数及90分钟计划核算通过（分数使用浮点容差）。共24个设计样例，未做外部模型试跑或学生验证；未付费调用、自动发布、开发平台、提交推送或部署。保留其他任务的海报与交接记录。
- **2026-09-28 日间宣传海报**：按用户要求新增雾白底／深钛青M的日间版，保留夜间版。输出在本线程mofa-poster-20260928/day/（绝对父目录见下方海报初稿记录），含1800×2400 PNG、1080×1440 JPG及规格；生成源为同级上层build_day_poster.py。已目视检查、确认与深色版26组文案及坐标一致、两种尺寸二维码模块对比零差异。未手机实扫、未发布、未改业务/配置/Docker/其他任务文件；保留已有CUSTOMIZATIONS、HANDOFF及product-samples工作区变更。
- **2026-09-28 魔法配方首批样品（优先于实验室建议）**：用户选择将AI工作方法包装成模板，并同意先做3份样品。新增product-samples/magic-recipes/说明及会议转待办、需求转验收、报错转排查三份内容，含完整提示词、虚构输入、人工参考答案、边界样例、验收和修订提示词。Node文件检查确认三份各7项必要章节、围栏配对及3个目录链接有效；共12个设计样例，未做外部模型试跑，不能标已验证。未调用付费API、开发平台、创建定时任务、提交推送或部署；仅新增样品及更新交接／定制登记，保留原有HANDOFF修改。
- **2026-09-28 产品线咨询**：用户提出魔法家族母品牌、狮子＋皇冠M共同标识及两产品分工，询问第三款获客产品。建议魔法实验室（暂名）：模型实测对比与自愿分享入口；属于待验证假设，未确认立项。本轮仅补交接，未改业务、提交推送或部署。
- **2026-09-28 发布最终核验（优先于此前状态）**：功能9edc7a7bb及CI修正a1b91372e均已推送origin/main；ls-remote核实修正完整SHA a1b91372e55f23f4592ab1f95b9e043c57e4b498。本地全量前端lint、推广5测通过，此前124项关联回归通过。修正SHA的GHCR运行36375537015成功，日志确认sha-a1b9137及latest已发布，manifest digest 076905e07d5966cfc9c4f2cabdce4c1c0b3d8d4adee7cfed4368c78d0f2f10e2；安全扫描36375537029成功。CI36375537042前端、golangci-lint、shell和release-helpers成功，最后查询Go Unit tests仍in_progress、Integration tests pending，不能宣称全绿。建议待本次CI成功后，用户备份并在/sub2api-deploy执行 ./update.sh sha-a1b9137；生产未操作。
- **发布CI修正**：功能提交9edc7a7bb已推送并核实远端；镜像运行36375045453成功（sha-9edc7a7），安全扫描成功，Go lint成功，但前端CI因ProxyAdBanner空模板触发vue/valid-template-root失败。已改为render返回null的兼容组件并补推；升级应使用修正提交镜像，不能把首轮CI描述为全绿。
- **2026-09-28 包号与推广清理提交发布**：用户授权提交推送origin/main供手动升级；fetch确认提交前HEAD与origin/main一致。已用Go 1.27.0 Docker对本轮Go文件执行gofmt，124项前端定制关联回归全部通过，diff空白检查通过；此前本机Docker前端及Go embed构建通过。此次推送须核对对应SHA的CI／GHCR结果，不沿用旧CI；Go单测与lint以本次CI为准。生产不操作。
- **2026-09-28 本机Docker部署完成（优先于下方未部署记录）**：用户授权本机验证；启动原Docker Desktop，代理授权接口200，实际构建进程设置代理，前端构建／3项i18n／类型检查及Go embed编译成功。原项目仅替换应用容器，PostgreSQL和Redis复用，三服务healthy，health=ok；镜像与运行容器一致为d0c2c1d46026d0699ce3e153c4aba9c102e707275e84320990549512265bf5a8。数据库custom_dedicated_accounts表存在；HTTP入口index-B18iYC10.js引用两张包号页面，新首页及用量页资源不含上游仓库推广链接；两类包号接口未登录均401。地址http://127.0.0.1:8080，待用户真实登录／带数据验证，未执行Go单测、不制造调用费用。回退镜像sub2apicust:before-dedicated-20260928-112340，数据库备份output/brand/before-dedicated-20260928-112340.dump已pg_restore目录校验，构建日志output/brand/docker-dedicated-build.log。未提交推送、未操作生产。
- **2026-09-28 上游推广清理（源码，未部署）**：移除AppHeader管理员GitHub菜单及HomeView／KeyUsageView页脚仓库链接；ProxyAdBanner改为空兼容组件，禁用代理管理／创建账号／编辑账号三处代理购买广告。保留客服配置、OAuth、合规和运维帮助及LICENSE，不动包号未提交工作。新增5项推广回归测试和vue-tsc -b通过；未浏览器验收、未重建镜像、未提交推送或部署。审计边界见CUSTOMIZATIONS。
- **2026-09-28 Claude／ChatGPT-Codex包号首版（源码，未部署）**：用户确认首期平台后，新增管理员「包号管理」与用户「我的专属账号」、独立绑定迁移241、到期／撤销／续期、用户白名单响应及上游剩余百分比快照。首期仅标准模式，需先配置一账号／一标准专属分组／仅授权目标用户；不自动搬号、不改变余额计费。API Key原始身份、两平台选号／转发前及WS每轮均接入实时隔离检查；失败关闭，撤销保留记录不回公共池。额度只读被动快照，无模型探测或强制OAuth刷新；缺失、过期、重置后不冒充0%／100%。使用与回滚边界见DEDICATED_ACCOUNTS.md，所有上游接缝已登记CUSTOMIZATIONS及custom/UPGRADE。已执行119项前端定制关联回归、3项i18n检查、vue-tsc、定向ESLint、Vite生产构建；最终复测同样119项通过，构建退出0，完整构建日志在output/dedicated-frontend-build.log（Git忽略）；构建仅既有Browserslist／chunk提示。新增Go/HTTP单测已写但本机无Go，尚未gofmt／编译／执行后端测试；Docker仅只读探测，desktop-linux引擎管道不存在，未启动Docker或重建服务。数据库迁移、真实账号／双用户隔离、负载与浏览器视觉仍待，不宣称已上线或CI通过。未提交推送／修改生产，保留开工前HANDOFF未提交记录。
- **2026-09-28 包号需求初查**：用户希望单个上游账号指定给用户，并提供菜单查看状态及剩余额度。本轮仅源码初查，未实现功能。已找到User.CanBindGroup专属分组授权、Account.GroupIDs多分组归属、Group备用分组及管理员GetUsage的用量百分比与重置时间能力。专属分组不等于账号独占，本站消费不等于上游余额。待确认首期平台，再设计用户归属校验、脱敏展示及独占调度约束。未改Docker或生产，未提交推送，保留此前交接改动。
- **2026-09-28 宣传海报初稿**：用户要求制作海报，已沿用「魔法家族／多模型 API 服务」、核心M／狮冠与雾钛青，制作1800×2400 PNG及1080×1440 JPG。素材为Pictures/mofamily/mofa-mark-flat.png，保留轮廓后重着色排版；不重画M、不修改原图。输出位于C:/Users/Administrator/.codex/visualizations/2026/09/26/01a0dd78-fd99-7e82-bbcc-5d095ed76b1c/mofa-poster-20260928/，含build_poster.py与poster-spec.json可复现源。采用现有Logo合成与确定性文字排版，未调用AI生图API（内置工具不可用，未擅自调用需要密钥的替代接口）。文案不含价格、稳定率、延迟或官方授权等未经验证承诺；二维码编码为https://ai.mofamilys.com。已目视检查成品、验证尺寸/文字边界无重叠，两种输出的二维码模块采样与生成矩阵一致；尚未手机实扫或验证生产站点可用性。本轮仅海报及交接记录，无业务代码、Docker、配置、提交推送或生产变更，保留原有未提交交接内容。
- **2026-09-27 SynaRoute官网与客户端只读评估**：用户确认两者都看。定位真实工程D:/ccfile/SynaRoute（HEAD 317ce19，package 0.1.69，git status检查前后干净）；线上官网中文页桌面/窄屏及本地Vite纯mock客户端的Key列表、大脑聚合、设置日夜主题、用量页已查看，未启动真实代理、未读取真实密钥、未调用付费模型。已核实官网steps.s4.title/desc原样显示（site/src/data/features.ts:107与语言表缺项）；客户端深色说明色#52525B在实测卡片底色上对比度约2.30:1（src/styles.css:118），设置页9个switch缺可访问名称（ToggleRow.tsx:59）；Linux默认加密分支仍用编译期固定密钥（secret.rs:861、888），而官网Linux标available及统一DPAPI文案不匹配，需优先处理，但未验发布二进制或证明用户密钥泄露。官网卸载清理说明漏掉可能位于exe同级的logs（store.rs:290）。建议保留紫色，精简首屏/分层高级配置属设计判断。SynaRoute产品文件未改、未运行完整测试/构建、未提交推送部署；本轮仅本仓交接补记，不另建SynaRoute真机验收编号清单。
- **2026-09-27 SynaRoute调整咨询**：已读品牌计划及本站导入生成器与测试源码；既定方向保留紫色。未检查客户端工程或当前官网、未运行测试，不能断言产品缺陷。本轮仅补交接，保留原有未提交修改，未改业务或部署。
- **2026-09-27 安全初查**：用户反馈异常登录；静态审查本地c11990cb6密码登录/JWT/管理员鉴权/限流/IP解析/审计/会话绑定及部分定制外联，未在已读路径发现万能密码或无凭据管理员放行，不等于全仓无后门。auth_service.go:535验证密码；admin_auth.go:120存在独立管理员API Key入口，不经过密码/TOTP，泄露后仅改密码不能撤销；config.go:2083默认信任转发IP，结合pkg/ip/ip.go:55、rate_limiter.go:121，若源站直达或反代不清洗伪造头，可污染审计IP和分散限流桶（未证明生产满足条件）；auth_handler.go:264登录200可能只是进入2FA，需关联最终认证；session_binding.go:61的IP/UA不匹配不能单独证明入侵。git diff upstream/main...HEAD核对上述核心认证文件无fork差异，仅为本地已有基线。未取得生产日志/配置/镜像；本机Docker只读ps失败（引擎管道不存在），未编译测试、未改业务代码/配置、未部署/提交推送。在线公告查询未取得可引用结果，不断言无已知漏洞。
- **CI最终结果补充（优先于下方中途状态）**：GitHub API最后核验功能提交d592f9481的CI运行36264121775已completed/success；结合镜像36264121792和安全扫描36264121782，三项全部成功，集成测试不再待办。镜像标签仍为sha-d592f94。纯文档补记b631b82ae已收到git push成功回执，但其后ls-remote连续网络连接失败，不把失败查询描述为远端核验通过。
- **发布已推送／镜像已生成（2026-09-27，本轮最终记录）**：功能提交 `d592f9481c2006993b9825ef830766f81431cf18` 已推送origin/main，并用git ls-remote核实一致。GitHub Actions镜像运行36264121792成功，日志确认已推送 `ghcr.io/engineemomo/sub2apicust:sha-d592f94`（同时latest），manifest digest `sha256:627aebd01fbf7d403b66b45e3e533180d1269b87a5427d348ac9dc6c9e456735`，linux/amd64，版本 `0.2.8-custom.d592f9481`。对应安全扫描36264121782成功；CI 36264121775的前端、golangci-lint、shell、release-helpers和Go单测已通过，但最后查询集成测试仍运行，不能宣称CI全绿。该发布结果以随后纯文档提交补记，不另造镜像。生产未操作；建议待该CI完成后用户备份并执行 `cd /sub2api-deploy && ./update.sh sha-d592f94`。
- **提交前最终检查**：本轮76项定向测试、3项i18n完整性测试、定向ESLint、vue-tsc与Vite生产构建及git diff --check均通过。构建有既有Browserslist／chunk体积提示，未阻断；本机未运行Go单测，CI结果待本次推送后核验。
- **本轮发布收口（用户授权提交推送）**：提交范围为智力效率页面、可读效率对比、排序／两段说明精简、充值空态宽度及维护文档。已fetch确认提交前HEAD与origin/main一致；76项定向回归、3项i18n检查及定向ESLint通过。推送后须核验本次SHA对应CI与GHCR镜像，不能沿用旧版本成功结果。生产由用户按UPDATE_GUIDE备份并手动更新，代理不操作生产、不创建售卖套餐；本机8080尚未包含最新两段说明精简。
- **订单菜单开关取证**：用户新截图已显示展开的订单管理及订阅套餐。源码确认该组为上游原生，受adminSettingsStore.paymentEnabled控制（支付配置enabled），系统设置→支付设置→启用支付对应form.payment_enabled；简单模式隐藏，路由要求管理员且公开payment_enabled明确false时重定向。当前侧栏diff仅增加智力效率，不改订单开关。本轮未改配置，无法从截图确定此前不可见的实际原因。
- **说明区再次精简（最新）**：按用户指定只保留IQ可比性及低样本／缺失数据两段，中英文同步，移除标题和其余说明。intelligence-view.spec.ts 8项组件测试及git diff --check通过；本轮未重建Docker、未提交推送或部署生产，8080仍是上一版。
- **套餐管理入口与上游来源核验（本轮仅取证）**：用户指出订单管理下看不到套餐，所附截图当前选中的是「订阅管理」，截图中的「订单管理」仍折叠。不能据此断言用户展开后的界面一定正常。已用git show／blob哈希确认：本机upstream/main=`a3eb7ef302961cba716dc78b39b93b60c467db0e`，上游AppSidebar原本包含订单管理→`/admin/orders/plans`；AdminPaymentPlansView文件hash `93a5303a76218e6f34fec8f185fd5ccc17d7d35d`、payment_config_plans.go文件hash `095730c1e8a4350a77749687af72807a162d7be4` 均与当前工作树相同。实际按钮中文是「创建套餐」，之前答复称“新建”不够准确。运行8080的AdminPaymentPlansView-COR3nusR.js HTTP200，含payment.admin.createPlan动作；已请求Codex面板打开直达地址，工具仅返回queued，不代表用户已看见或实际点击已验证。本轮不改业务、不建套餐、不重建Docker，当前运行仍cfa0a12c8bf6。
- **当前本机部署：可读性修正版已运行8080**：应用镜像`cfa0a12c8bf6`，与本轮构建ID一致；三服务healthy，数据库／Redis容器ID未变。HTTP实取入口`index-BW6NgO3M.js`引用`IntelligenceView-C0f1Aymd.js`，新条形对比、矩阵排序控件、移除旧频率元素／外链以及支付容器width:100%规则均存在；/health、订阅页和套餐管理入口HTTP200。Docker内前端构建与Go embed编译通过，未跑Go单测；用户=1、分组=2、套餐=0以及280项设置摘要与更新前一致。本轮回退镜像`sub2apicust:before-intelligence-ux-20260927-022446`和583628字节dump已保留，部署证据`output/brand/intelligence-ux-deploy-record.json`，均Git忽略。真实浏览器连接仍不可用，视觉／交互验收仍待；未提交推送、未部署生产。
- **最新用户反馈：智力页面可读性与订阅释疑**：已移除查看源站按钮／JSON外链／刷新频率文字，保留来源归属和实际30分钟轮询；效率分布改为带模型名的分数／耗时／费用三列条形对比，默认high档、可筛选及排序、12条分页和手机堆叠；不合并环境，不造综合排名。矩阵默认源顺序，新增名称／指定档位分数降序，缺失最后、同分保留原序。支付空态窄卡对应定制flex正文与上游mx-auto缺宽度的适配，已仅在theme.css补width:100%，真实浏览器效果待用户验收。76项回归、ESLint和前端构建通过；浏览器工具仍因apikey认证失败不可用。部署结果以本轮后续记录为准。
- **订阅现场再次只读取证**：本机「测试分组1」确有记录（id=2，anthropic／subscription／active，日100、周500、月2000额度），但subscription_plans为0条。代码`ListPlansForSale`只读上架套餐，故当前空态与数据一致，不能将“只建分组不显示商品”称为上游bug。应到订单管理→订阅套餐创建商品、关联该分组、设置售价／周期并上架；未擅自定价、未创建套餐／订单、未改用户分组。
- **上轮初版Docker发布（历史）**：用户要求「重新发docker」后曾部署`1a7a31d0e424`，三服务healthy，构建与HTTP资源验证通过；证据见output/brand/intelligence-deploy-record.json，旧镜像和备份保留。该镜像已由上方可读性修正版替代，不再作为当前运行版本。
- **2026-09-27 智力效率新功能（本轮）**：用户提供 Codex Radar JSON，新增认证 `/intelligence` 和侧栏入口：雷达图标／模型品牌图标、分数条矩阵、耗时／费用散点图、搜索／环境／样本过滤、中英文、来源更新时间与本页获取时间、低样本／过期／失败提示。页面打开期间每30分钟拉取，支持手动刷新；不属于服务端常驻采集。真实源数据24模型／89组合解析通过；70项定制与订阅测试、ESLint及前端构建通过。浏览器检查工具返回 `unsupported Codex auth method: apikey`，真实浏览器视觉验收未完成，不以组件测试代替视觉验收。本轮未提交推送、未重建本机Docker、未部署生产，历史提交授权不自动扩展到本轮。
- **充值／订阅核验（本轮）**：原生 `PaymentView.vue` 已有两Tab与套餐购买／续费；本机公开设置订阅／支付开启且未禁用余额充值。本轮只读SQL `SELECT count(*), count(*) FILTER (WHERE for_sale) FROM subscription_plans` 返回0／0，尚未配置可售套餐；不自行套用参考截图的价格、额度或周期，未改数据库或支付业务。
- **本轮开发预览**：此前启动只监听127.0.0.1:3001的Vite开发服务，`/intelligence` 使用本机8080后端代理，需登录；曾向Codex浏览器面板请求打开（工具返回queued），不等同浏览器验证。开发会话22909；本轮Docker更新后应优先使用8080正式本机入口查看新页面。没有写入登录凭据或创建订单。
- **本轮用户收口决定（最高优先）**：用户表示「暂时先这样」，明确授权提交并推送代码、更新待办、复制新Logo。当前视觉与多模型名称按阶段性接受保留，不再作为必须立即重做的任务；下面关于“未确认／未提交”的条目是历史过程。生产仍未授权部署，CI结果必须按本轮提交单独核验。新Logo副本已校验，路径见第五节。
- **本机2026-09-27 00:58 最新品牌状态**：用户要求去除GPT限定并统一顶部Logo后，本机三项品牌配置已改为「魔法家族」／「多模型 API 服务」／扁平雾钛青M。镜像仍b4fa2d8e97a1，无需重建；只新增素材mofa-mark-flat.png。完整管理员设置前后仅这三项变化，其他277条原始设置事务核验未变，实际登录、侧栏、标题和双主题手机／桌面通过。数据库备份607430字节保留，操作风险与证据见BRAND_IMPLEMENTATION第17节。名称与视觉仍待用户认可，生产未修改。
- **2026-09-27 00:43 最新有效状态**：用户再次否决登录页排版后，已改为统一1120px双栏面板，缩小标题和M、表单左对齐、手机上下堆叠。只改认证展示容器与theme.css，业务页／认证逻辑／SynaRoute不改。镜像b4fa2d8e97a1已部署本机8080，三服务healthy；61项测试、构建、真实30组认证布局与实际登录通过。回退镜像和数据库备份已保留。详见BRAND_IMPLEMENTATION第16节；用户视觉确认仍待，上一版不是已认可设计。
- **2026-09-27 当前有效状态**：已针对上一轮视觉否决完成结构调整并重建本机Docker，不再只换色。认证顶部品牌栏＋大屏双栏；控制台路径顶栏＋独立正文标题／操作＋统一筛选表格分页面板；核心M、完整业务字段、默认深色及SynaRoute保持。61项测试、定向ESLint、前端构建通过；Docker内Go embed编译通过，未跑Go单测。镜像／容器一致30804c47efc0，三个服务healthy，入口127.0.0.1:8080。真实64组双主题／四宽布局检查通过；创建弹窗取消、列设置及移动菜单实测。证据和局限见 BRAND_IMPLEMENTATION 第15节，用户视觉确认及真实带数据验收仍待，不自称设计完全交付。旧镜像／数据库备份保留，未提交／推送／生产部署。
- 以下2026-09-26条目为历史过程，不代表当前运行状态；旧9c8c版本的用户否决仍保留作为改进依据。
- **用户最新视觉反馈：未通过验收**。用户提供2560×1317登录截图，指出与概念稿偏差过大、进入控制台后仍像原版换肤。已对照原稿mofa-aurora-system.html与源码：认证布局限制1160px内容区／440px表单，大屏留白明显；控制台主要新增类名和公共样式，尚未落地原稿的正文页标题、筛选与表格一体化布局。部署及功能冒烟成功不等于视觉达标，不得把当前版本表述为设计完整交付。保持雾钛青／核心M／上游业务契约／默认深色，SynaRoute不动；后续应先对齐登录及真实API密钥页的结构和比例，再扩展共用布局，不能再仅换色或只做独立品牌页。此次仅核查并记录反馈，未改产品代码或重新部署。
- **Docker网络故障防复发记录**：已补 deploy/LOCAL_DOCKER_RUNBOOK.md（直接原因／推断边界、代理预检、构建进程环境变量、失败即停、镜像一致性验收）。后续本机重建先读此文件，不盲目重复构建；本次记录操作未重启服务或修改代理。
- **2026-09-26 23:50 本机新版部署成功（以本条为准）**：用户授权代理调整并重启Docker后，给构建命令进程额外设置HTTP_PROXY/HTTPS_PROXY=http://127.0.0.1:7897，构建成功。镜像与运行容器ID一致，前缀9c8c1dfb87bc，三个容器均healthy；入口 http://127.0.0.1:8080/brand 与 /login。真实设置保留GPT中转站／M标志，健康接口200，正常登录成功。五页面双主题×360/1440共20组无横向溢出，业务API观察无>=400。仍为空库验收，未真实调用上游，不代表带数据全流程完成。
- 原数据卷和旧镜像回退标签保留；新增即时备份 output/brand/before-brand-final-20260926.dump（596221字节）。未提交、推送或部署生产。成功日志 output/brand/docker-brand-build-client-proxy.log，详见 BRAND_IMPLEMENTATION 第14节。此前“未部署／容器停止”描述均为历史状态。
- **2026-09-26 用户要求再次重试**：第三次本地Docker构建仍在 auth.docker.io/token 连接157.240.20.8:443时超时，无法取得 node:24-alpine 元数据。日志 output/brand/docker-brand-build-retry2.log；未进入编译、未启动容器、未改代理或DNS，新版仍未部署。下一步须先解决Docker Hub网络访问，单纯重试尚无改善。
- **2026-09-26 23:28 本机Docker部署尝试（最新状态）**：用户已重新授权本机Docker部署，取代此前暂不运行限制；不授权生产、提交或推送。核对 desktop-linux 本地命名管道与 sub2apicust-local 项目后，保留旧镜像标签 sub2apicust:before-brand-20260926，备份数据库到 output/brand/before-brand-20260926.dump（595798字节，Git忽略）。两次构建均在拉取基础镜像元数据时因 auth.docker.io:443 超时失败，新版未构建／未部署。构建未到编译步骤；未改网络配置、未改设置API、未删除数据卷。最终已停止本次临时启动的postgres/redis，三个本地容器均回到原停止状态。须恢复Docker Hub访问后再继续，日志 output/brand/docker-brand-build.log 与 docker-brand-build-retry.log。
- **2026-09-26 本轮收尾**：独立 /brand、认证品牌双栏和工作台展示钩子源码已落地；长站名手机溢出已修复。20项回归、定向ESLint、i18n 3项及前端构建通过；公开页面24组浏览器检查无溢出，详见 BRAND_IMPLEMENTATION 第12节。
- **部署状态以本条为准**：本轮没有运行Docker、提交、推送或部署；旧8080镜像不含本轮布局。前端503隔离预览不代表后端验收，带数据控制台／真实调用／用户视觉确认待继续。三个新接缝已登记，维护步骤见 frontend/src/custom/UPGRADE.md；不保证未来零冲突。
- **2026-09-26 用户最新限制：先别跑 Docker**。后续不启动 Docker Desktop、不执行容器启动/镜像构建/重建；仅继续代码及前端验证，恢复 Docker 操作须用户重新授权。中断前曾发出 Docker Desktop 后台启动命令；当时项目 compose up 因引擎未运行而失败，本轮尚未重建镜像。当前新增独立 /brand 页面与品牌展示组件，AuthLayout / AppLayout / AppHeader 仅加入展示接缝/样式钩子，视觉仍集中 theme.css；新增品牌/升级契约 9 项测试及 SynaRoute 11 项回归通过，前端 pnpm run build（含 i18n 3 项）通过。工作尚未完成：真实业务视觉复验、升级维护说明与完整变更登记待补；不能宣称本机 8080 已包含本轮改动。原有未提交内容保留，未提交/推送/生产部署。
- **2026-09-26 截图问题已修复并本地重建**：仅 theme.css 修复深色标题优先级、输入错误态红边与主按钮加载图标前景。容器实测标题 #A1D9CE 且 background-image=none；登录、密钥、用量、管理员页深浅主题及 360/768/1440px 检查无横向溢出。真实创建密钥弹窗可打开/取消；当前为空库，未验证带数据图表或上游请求。用户首次合规确认已不再出现，本轮未代确认。通过已有后台设置将本地实例名称/副标题设为 GPT中转站 / 魔法家族，并接入保留 M/狮冠的银色位图加炭青底板（非矢量重绘）。构建、验证证据、局限及一次已恢复的设置接口连带默认值变化见 BRAND_IMPLEMENTATION.md 第 11 节。未改 SynaRoute、业务结构或生产；未提交/推送。
- **2026-09-26 用户截图复核 / 完成口径修正**：用户质疑实际效果是否为改造版。直接读取本地容器登录页 computedStyle：primary-500=9 107 104，dark-950=22 33 36，主按钮底/字=161 217 206 / 23 59 56，证明基础主题已加载；但并非完整品牌改造成品。仍是默认 Sub2API + /logo.svg，核心 M 仅未改原资产，尚未接入本地实例。发现真实缺陷：登录 h1 颜色实际透明，仍用 #085E5B→#096B68 深色渐变，主题层预期浅青标题覆盖未生效，暗底标题偏暗。此前无后端截图未显示品牌区，漏检此项；不得再称主题验收完成。此次只核查和记录，未改业务代码/重建镜像。
- **2026-09-26 本地后端已就绪**：Docker 从当前源码完成前端 + Go embed 构建，sub2apicust-local 三容器 healthy，http://127.0.0.1:8080/health 和公开设置接口均 200；真实管理员登录成功。独立测试数据卷、仅本机端口、在线更新关闭，未碰生产。配置与随机凭据只在被忽略的 deploy/.env 和 deploy/docker-compose.override.yml；启动方法见 BRAND_IMPLEMENTATION.md 第 10 节。首次登录合规确认未代用户接受，业务页完整验收待用户自行阅读确认后继续。没有跑 Go 单测，没有提交/推送，SynaRoute 未改。
- **2026-09-26 阶段一实施**：产品代码只改 frontend/src/custom/theme.css，原有未提交文档保留并续写。雾白/炭青色阶、日夜成对主按钮、输入焦点和卡片已落地；不动核心 M、业务结构、默认深色、分类色或 SynaRoute。真实登录页双主题、360/768/1440px 无横向溢出；主按钮对比度 6.34:1 / 7.76:1。公开设置接口本地返回 500，品牌区未加载；密钥/用量/管理员页均跳登录，业务验收未完成。验证与截图见 BRAND_IMPLEMENTATION.md 第 9 节。未提交/推送/部署，未重新查询 CI。以下设计记录为历史。
- **最新定案 / 下个对话直接执行（2026-09-26）**：用户已确认「就按照你这个设计执行」，并要求先安排工作、换新对话实施。已新增 **BRAND_IMPLEMENTATION.md**，含最终稿绝对路径、色表、核心 M 保留原则、SynaRoute 保留紫色、阶段划分、代码证据、验证与维护边界。**不再重选风格，从阶段一真实业务页换肤开始。** 本轮仅修改交接文档与任务入口；主题、后端、默认深色策略均未改变，未构建/提交/推送/部署。下列「待确认」均为历史讨论，设计方向以本条和计划为准。
- **2026-09-26 品牌第四轮（保留核心 M，颜色可重做）**：用户明确允许更换 Logo 色彩，核心 M 设计必须保留，要求重新思考设计。新增线程预览 mofa-aurora-system.html「雾钛青」：以现有标志透明遮罩保留 M/狮冠轮廓，重着钛青/浅银绿；雾白日间与炭青夜间组成同一套主题；首页与控制台分屏预览，增大正文字号、去掉轨道/网格堆叠、减少装饰。已浏览器验证页面/深浅切换、搜索、空状态、1024px 桌面及 360px 视口外层无溢出，并修正搜索图标定位。仅概念验证，非真实 Vue 页面实装，非矢量重绘，舒适度/转化效果未做用户测试。产品默认深色和业务代码均未变。
- **2026-09-26 用户补充品牌验收要求**：设计必须同时考虑用户舒适度与吸引力，不能仅追求科技感。后续建议区分品牌首屏与高频操作区：首屏用克制的 B+ 品牌表达，控制台减少装饰、保证信息层级与可读性；深浅模式保留选择。舒适度和转化效果尚无用户测试证据，不能宣称已验证。当前默认深色策略不擅自更改，下一步应在真实页面检验，而非继续单纯追加概念稿。
- **2026-09-26 品牌第三轮（B 与 C 对比、增强科技感）**：用户询问 B/C 哪个更好并要求更有科技感。新增线程预览 mofa-tech-directions.html：B+「深蓝精密科技」与 C+「银白精密科技」，采用第二轮克制排版、细网格/轨道/局部冷光与可选接入路径示意，业务表格区不添加装饰。建议 B+ 作主视觉、C+ 作浅色配套，属于设计判断，尚未获用户确认。已检查浏览器方案切换、标志加载及桌面/360px 视口外层无横向溢出。未改产品主题、业务代码或生产环境。
- **2026-09-26 品牌第二轮（高级感 + fork 维护优先）**：用户认为 A/B 不够高级，强调能长期同步开源更新。新增线程预览 mofa-refined-directions.html：C「铂银纸白」、D「石墨冷银」，减少渐变/卡片/发光，控制台示意保留侧栏、筛选与表格结构。已通过浏览器检查外观、切换、搜索示例及 360px 视口外层无横向溢出（表格内部允许横滚）。代码取证：primary/dark 已接入 theme.css，但 AppLayout 仍含 bg-gray-50 与独立 bg-mesh-gradient 覆盖层；不能承诺仅改变量就完整复刻概念稿。建议品牌展示页与业务控制台分开：展示页独立新增，控制台只做受控主题覆盖，不影子替换业务页、不改后端。CSS 覆盖仍需上游升级后检查，不能承诺零维护。尚未实装、未部署。
- **2026-09-26 品牌视觉对照稿已制作（未实装）**：用户确认先看两套稿。在线程可视化目录保存 mofa-brand-directions.html，含 A「银白清透蓝」与 B「墨蓝冰银」，各展示品牌官网首屏和 API 密钥控制台，内容一致、示例数据明确标注。沿用用户提供的朋友设计标志，仅做去背景预览处理，未完成矢量重绘、未改原图。浏览器已检查两套外观、标志加载、方案切换及导入按钮提示；360px 视口下未发现横向溢出。未修改产品主题或业务代码、未部署。
- **2026-09-26 生态品牌讨论（方案待确认）**：用户提供 SynaRoute 官网及银色狮子/皇冠 M 标志参考，希望统筹生态品牌，允许调整色调。本次通过浏览器读取官网文字及首屏截图：当前首屏为浅底紫色强调，定位为本地 API 路由与多模型协同；仅验证页面展示，未验证软件实际行为。讨论方向为母品牌统一背书、SynaRoute 与 API 服务各自定位、银白与蓝色为主的视觉家族；尚未获用户确认，未修改主题或业务代码。
- **2026-09-26 Codex 接手核查（仅结构与待办）**：已读取交接文件、目录与依赖清单；核查开始时本地 main 与本地 origin/main 引用一致，工作树干净，HEAD 为 0d1e3e7b8，VERSION 为 0.2.8。本次未联网核验远端/CI、未运行构建、未访问生产环境；仅更新本交接文件，不改业务代码。下列 CI/部署历史记录不代表本次重新验证。
- 本地分支 `main`，HEAD=`0d1e3e7b8`，与本地 `origin/main` 引用相同；VERSION=0.2.8。**当前工作树有未提交交接文档改动，不是干净工作树**；本轮未 fetch，远端现状未重新验证。
- fork 接线：`origin`=https://github.com/EngineeMoMo/sub2apicust.git ；`upstream`=Wei-Shaw/sub2api（**push 已禁用**，只读）。
- 上游同步：**2026-09-26 已合并上游 0.2.8 快照（a3eb7ef30，207 提交）→ 零冲突、11 处关键定制全存活、VERSION 0.2.8**。合并后 `pnpm run build` 通过；后端 `api_contract_test.go` 新增的 CCS 夹具已镜像 SynaRoute（全后端 hide_ccs = hide_synaroute = 20）。
  - ⚠️ 该快照是**上次成功拉到的** 0.2.8；当日 github 从本机拉不动，快照之后若有更新的上游提交未纳入 —— 下次网络好时 `git fetch upstream` 增量再同步即可。
- 前端：`pnpm run build` 通过（typecheck + i18n + vite）。本机预览：`pnpm -C frontend run dev` → :3000（无后端时仅外观预览，数据页会跳登录）。
- ✅ **CI 已验证（2026-09-26）**：0.2.8 合并提交 `58b154f53` 的三条 workflow 全绿 —— Build custom image(GHCR) / CI(单测，含 `api_contract_test`) / Security Scan。即后端已编译通过、单测通过、**0.2.8-custom 镜像已推到 GHCR**。
- **部署/升级方式（2026-09-26 用户定案：Docker + `deploy/update.sh`）**：日常升级 = 部署目录跑 `./update.sh`（拉 GHCR 定制镜像 `ghcr.io/engineemomo/sub2apicust:latest` → 重建 → 启动自动迁移 → `/health` 自检 → 清旧镜像）；指定版/回滚 = `./update.sh sha-<提交>`（自动切 override 的 image 标签并备份）。首次从现有 0.2.7 平移见 `deploy/DEPLOY_CUSTOM.md` 第七节。**不用 App 内按钮**（原因见「五」）。

## 五、待办 / 下一步

- [x] **本轮截图所指源码修正**：滚动吸顶栏、SynaRoute选中态、首页提示文案、两图撤下、五类各两张新图与首页人像入库已完成；186项回归、类型／lint／构建与浏览器证据在`output/studio-refinement-20261001`。最新47图优先于下方历史数量，剩余素材由用户自行添加与审核。
- [ ] **线上生图目录真实取证**：同一登录账号确认API密钥页是否有启用、未过期、额度可用且绑定启用分组的密钥；若文字有而生图无，核对分组生图权限；若有可用密钥却无模型，取所选密钥`/v1/models`响应／状态码（不发送密钥明文）。当前仅模拟目录流程通过，不能标线上问题已解决。本轮新源码尚未提交发布，需后续明确授权，并处理既有765e5a9门禁后再核对新SHA的CI／镜像；不代部署生产。

- [x] **家族产品限定提交与推送**：765e5a9已在origin/main，GHCR固定标签与摘要由成功任务日志确认；排除包号与新版更新脚本。交接文档后续skip-ci提交不生成替代应用镜像。
- [ ] **765e5a9暂不部署：修复新SHA门禁**：检查并处理工坊handler八处资源清理错误返回；核实x/image至少v0.45.0修复上述WebP／VP8L可达漏洞，升级后重新验证投稿尺寸／类型／恶意素材与头像路径；处理Axios七个high，不自动扩大安全例外。依赖修复已向用户询问，后续应另发新功能SHA并核对安全扫描、Go lint、单测／集成与GHCR成功，再允许用户更新测试。不得把333项前端／静态回归或已构建镜像当安全发布通过。

- [ ] **本轮家族发布与用户更新测试**：按用户授权提交／推送限定家族产品版本，核对新功能SHA对应的CI／镜像构建实际状态和固定sha标签；镜像成功后用户备份Postgres、部署配置及包含studio-submissions的app_data，再手动更新。真实登录下主题同步、配方Key／模型同页选择、图片投稿待审／管理员发布／撤回仍需用户实测，视频与分镜保持待开放；不混入并行包号与更新脚本，不代部署生产。

- [x] **暂不开放视频与分镜**：已关闭工坊展示、玩法／旧链接与新视频投稿前端入口，显示友好“待开放”，原始资料与历史记录保留。服务端协议和管理员历史管理不属于本次停用范围。
- [ ] **视频与分镜重新开放条件**：先核实实际可用视频供应商、任务提交／查询接口、账号权限与费用预算；准备9–12段题材／风格不重复的原创或明确获授权成片，逐条附对应完整实际提示词、参数及来源／授权。经人工审核和桌面／手机播放、完整画幅验收后，再同时恢复core.mjs的VIDEO_CONTENT_ENABLED、投稿选项与提交／重投路径，并重跑相关回归。不凭历史Sora文档接入、不自动收费、不搬运未授权作品；当前保持待开放。

- **瀑布流待用户视觉确认**：查看4185工坊与studio-masonry-desktop-final／mobile截图，确认原画幅紧凑布局；本轮仅预览与静态构建，8080仍旧版。后续提交／镜像替换需明确授权，包号并行修改继续隔离。
- **视频接口不能沿用旧Sora文档接入**：OpenAI官方Videos API2026-09-24已关闭且未提供一对一替代。若用户希望自动生成样片，先确认实际可用供应商／任务接口文档与预算，不发送密钥明文；本站第三方视频能力与实际账号权限尚未验证。也可沿已有投稿审核链展示有许可的成片与完整实际prompt；十套分镜仍不能说已生成。

- [x] **生产磁盘占用定位（截图）**：已取得df／Docker镜像独占大小／部署目录输出，证实带sha标签旧镜像约1GB，数据库备份5MB；已给保留当前及上一版、删除五个更旧镜像的具体命令。
- [x] **更新脚本镜像保留策略（源码）**：用户已要求增加，当前＋实际更新前版本＋其他容器引用镜像保护，失败不清理，22场景假Docker回归与语法通过；同步UPDATE_GUIDE与定制清单。
- [ ] **新版更新脚本交付与生产验收**：尚未提交推送，需把新版deploy/update.sh单独替换到服务器/sub2api-deploy/update.sh并chmod；镜像pull不更新宿主脚本。用户执行更新后复核df／镜像清单／回退版。备份／日志轮转未在此轮实现，不做全局prune或数据卷删除。

- **最新风格／投稿版验收与发布**：看4185风格预览；投稿／审核已实施并通过隔离测试，但旧8080尚无新接口。获得本轮明确更新授权后按runbook先备份DB＋app_data投稿目录，前后端同版本重建原应用，不改数据库／Redis、不删卷；用真实用户完成图片／真实MP4上传→pending私有→管理员确认发布→公开作品和对应prompt→下架／撤回不可公开，以及深浅／手机／反代上传体积验收。不可把静态预览、mock用户或验证镜像当部署；多副本上线前先更换存储层，清理额度与自动内容安全待另议。
- **视频能力仍待条件**：当前无视频生成工具，十套分镜仍0段对应样片。已向用户询问自己的魔法API视频模型／任务接口与单条预算（不用发送密钥明文），等待补充；未确认前不收费调用、不搬运未经授权的第三方原片或用静态动画冒充。用户可先投稿自己有展示授权的成片及真实prompt／明确反推模板，人工审核通过才公开。

- **最新十图验收**：本轮十张真图已接入，32张图库与90项回归／构建、实际桌面／手机检查完成，请用户看4185最新预览；没有替换8080或发布生产。确认后再按runbook独立授权更新／提交，不将当前源码预览当上线。
- **视频仍未完成**：10套分镜没有对应生成视频，本轮明确答复，不再把“视频条目”误称成片。需可用视频生成服务及预算，或可核实展示／再分发许可的素材；每段保存真实生成记录或明确反推模板再上播放器，不用海报动画、无关旧电影或编造原prompt填补。

- **本轮视频缺口（最高优先）**：用户要9–12段有吸引力的实际视频，现在只完成10套分镜，0段新成片。等待用户在已发问题中选择自己的视频服务（须确认接口和预算）、上传可展示片段或可核实再分发授权的现成素材；逐条拿到真实样片／真实对应prompt／来源与许可后，再上播放器与卡片，不用海报、二维运镜、无关影片或后写模板冒充。现成短剧若未公开原始提示词只能标“反推参考模板”，不能声称作者原prompt。
- **本轮验收／发布**：请用户查看4185最新作品与提示词布局及六张新图；确认后才补媒体并决定是否更新8080。当前新版未Docker构建或部署，无新提交／远端CI；保留原8080、数据库与包号并行工作。系统剪贴板、TXT落地、真实账号模型配置和收费生成仍不能沿用模拟测试；不增加自动采集、调用或定时任务。

- **最新本轮收口优先**：169项测试、类型／lint／构建及新工作区已完成。用户确认后才替换本机8080（重新备份、保留旧镜像和精确CORS层、仅更新应用、比对镜像与运行容器以及业务／设置摘要），不沿用前次更新授权作为本轮已部署证据。生产与提交推送须另行明确授权，不能把构建新测试镜像当发布完成。
- **真实工作区验收**：用户在同一站点真实登录→顶部配方／工坊→检查右侧展示、主站主题同步、已有Key／模型下拉、显式应用、取消与退出清理。没有Key时原密钥管理创建，不替用户创建测试Key或发收费生成；模型目录可查询，实际文字／生图／计费需要用户授权。外部独立URL、SynaRoute桌面账号SSO与跨站统一退出不在本轮。
- **视觉与素材后续**：请用户确认原创成年男女与新网格／完整预览；若必须按三条抖音逐镜分析，先修Chrome连接或用户上传可播放视频，当前未知完整视频内容。现有4段授权参考是开放动画电影，不称真人模型演示，待用户提供许可素材或授权模型／预算后可增原创真人电影短片。工坊生成API、Skills实际运行和定时任务仍未启用。

- **一体化最新待办（覆盖旧“必须独立托管”）**：本机8080构建／部署／数据保护／HTTP／桌面及手机页面验证已完成，用户可直接打开/recipes/和/studio/。下一步由用户在8080真实登录后核验控制台切换、配方配置弹窗及已有Key／模型授权，再决定收费调用；本轮keys基线0，不创建Key或发收费请求冒充验收。IAB未显示选择弹窗，需在用户正常浏览器确认，系统剪贴板／TXT实际落地仍待。正式生产需另行授权提交推送并核对新SHA的CI／GHCR，旧sha-25071d4不含本次集成，不拿旧CI替代；完整web测试的旧logo.png夹具失败单列处理，不在本轮修无关用例。统一清除已授权Key、视频API仍未完成，工坊定时任务保持关闭，不混入包号发布。

- **最新版本更新门槛**：25071d4 功能已推送，核实 GHCR 36736764218、CI 36736764231 和 Security Scan 36736764322 成功终态，再由用户备份部署配置／数据库并在 /sub2api-deploy 执行 ./update.sh sha-25071d4；标签仍须核对本轮成功任务日志，不在构建中拉取或直接使用 latest。发布后验工坊三图／鼠标与键盘置顶、用户及管理员皇冠、免登录官网下载入口、真实登录返回家族首页。生产、真实官网下载安装／协议导入和模型授权未验；配方／工坊仍须托管并设置公开地址，完整跨站 SSO 未新增，包号专项不在此功能提交中。

- **本轮家族视觉发布（优先于下方待授权历史）**：已获提交推送授权，按限定暂存版本验证后推送 origin/main；核对本轮 SHA 的 CI、安全扫描与 GHCR，再提供实际镜像标签供用户备份后更新。包号专项仍另行处理，独立配方／工坊地址和精确来源仍须配置；不把重发主站镜像当全部产品已托管或跨站 SSO 已实现。皇冠真实侧栏、生产登录回跳、官网实际下载及真实模型授权仍待用户验收。

- **SynaRoute 官网／下载入口待发布**：用户在4184/home#synaroute确认新增两个公开入口；链接与布局已在源码和浏览器验证，官网实时可达性／安装包下载／真实协议导入仍待用户网络下验证。本轮未改原导入失败提示或增加可靠安装检测，不能仅凭焦点断言未安装。确认后另行授权提交／发布，与未确认包号专项隔离；网站换域名时更新family/synarouteLinks.ts，不把第三方同名产品地址替入。

- **皇冠图标待发布验收**：本轮家族首页侧栏皇冠仅在源码，尚未进入8080或远端镜像；用户确认后再授权提交／发布。更新时验证用户及管理员个人区、展开／收起和双主题，仪表盘四宫格保持不变，不沿用旧镜像截图当新图标验收。

- **工坊三图待用户视觉确认与发布**：用户查看4184/home#studio或最新studio-balanced桌面／手机截图，确认深发色参考写真与三张等宽放大的布局；不能再沿用被否决的雨夜两图、金发人物或1.3:1:1卡片。默认无遮挡、鼠标悬停／键盘置顶及手机尺寸点按已取证，不等于用户视觉认可或实机触摸验收。本轮仅本地源码，321fe99不含这些改动；提交推送／Docker部署仍须新的用户授权，保留并行包号改动。独立产品实际发布地址仍未知；先托管，再按FAMILY_PORTAL“在哪里配置”创建MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN／MAGIC_STUDIO_URL仓库Variables并重建镜像，不把本机预览或回退默认站名当线上配置。

- **本轮发布检查**：功能321fe99及其镜像已发布，Go Unit通过，CI36712632123集成仍运行；先核验集成终态再由用户备份并执行./update.sh sha-321fe99，不标全绿或直接更新latest。此版仅产品相关，包号专项仍在工作区、须另行确认提交。三个MAGIC_*仓库Variables均未配置；独立配方／工坊静态页面分别托管并设置公开构建变量后需重建镜像，当前生产入口保持禁用，不宣称推原站镜像即可运行全部产品。

  - **产品展台视觉验收**：本轮替代旧产品卡片，最终94dba406镜像已部署8080，三项修复评分ship不等于用户视觉批准。用户在8080/home切换四个预览并确认入口表达，再用自己的账号验API控制台／顶部产品坞及配置授权；完整真实组件的模拟账户夹具只证明布局，不能当真实登录或SSO证据。生产仍不部署。

- **首页最新验收方向**：打开8080/home确认原品牌首屏及下方子产品介绍；用用户自己的账号验证默认进API控制台、顶部胶囊和显式打开配方／工坊，SynaRoute只进入密钥配置。不再要求登录返回产品目录。生产地址／域名仍待确认，未配置地址应禁用；后台home_content是否覆盖默认首页需发布时核实。真实配置授权、模型权限／计费、跨窗口闭环仍待用户测试；不代建Key或收费调用，工坊定时任务继续关闭。

- [ ] 多专属组修正与具体400诊断需同版本前后端发布／本轮CI；用户已确认两位成员都选全仍失败，不再要求重复勾选。线上公开管理页仍无新诊断；需核验实际后端版本与失败请求。此前严格单组规则已被本轮同批成员多专属组要求替代。

- [ ] 线上包号400：需本次保存account_id、group_id、user_ids及现场数据库配置来确认具体失败项；新诊断补丁尚未发布。源码／隔离库可排除Save使用缓存及完整标准独占配置必然失败，不能替代线上取证。

- **独立产品登录的剩余边界**：最新首页验收按本节首条；若需要独立母品牌域名或SynaRoute桌面账号SSO，先确认授权协议、产品账号能力与退出策略，再另行实施。不跨域共享JWT或宣称现有密钥导入就是SSO；工坊已是公开内容首版，正式发布地址仍待确认。

- **家族与配方发布配置**：确认同源/recipes/或独立子域名，以及服务器宝塔／1Panel／原生Nginx等管理方式；按FAMILY_PORTAL.md设置仓库MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN公开Variables（已接Docker构建ARG），独立来源还需精确API CORS。两端新版本验收后再由用户授权提交发布、查本次CI镜像、备份并部署；配方单文件仍单独托管，不能覆盖原站server块，不用运行容器变量冒充前端构建生效，也不把本机4178／候选域名写进生产入口。当前未操作仓库变量、域名或服务器，生产部署未授权。

- **魔法配方登录选择验收与发布**：从8080家族首页／控制台打开配方后，4178模型设置使用本机8080；用用户自己的账号选择已有有效Key／分组／模型并明确授权，再自行运行，验证真实目录、跨窗口、权限与计费。精确4178 CORS已HTTP验证，但没有真实账号闭环证据。线上需确认来源、配置VITE_MAGIC_RECIPES_ORIGIN并授权两端发布；生产仍未更新。手动连接与复制继续保留，视频接入待另行适配，不代建Key或收费请求。

- **第四产品本地验收**：魔法工坊16条资源／4个玩法已实施，4179仍独立Node预览，已接入本机家族介绍及控制台打开入口。待确认名称、界面、人群并在目标工具实测；约40条内容、站内生成、正式托管与Docker仍待决定。定时任务暂不启用，不自动创建调度；说明见product-samples/magic-studio/README.md，不把入口集成当模型／Skill效果通过。

- **魔法配方模型接入验收**：用户在4178顶部“模型设置”填写自己的文字或生图连接，应用后整理提示词并运行，验证实际Key／分组模型权限、CORS与计费；当前只证明模拟兼容接口可用，不代表真实服务接通。视频需要提供服务商任务提交／查询格式后另行适配；正式托管未决定。历史“是否接模型”已由本轮要求更新为自带接口，复制流程继续保留。真实剪贴板／TXT下载落地与配方模型效果仍待，不用接口模拟测试替代。

- **接入教程剩余验收**：前端检查与公开／预览验收已完成，8080已随本机源码镜像包含教程；生产未更新。后续用实际分组密钥分别验证Anthropic Messages、Responses与至少一个Chat Completions客户端及/help/guide正文；客户端版本变化重新核对官方文档，不把HTTP入口存在当真实调用成功。

- **专项修复下一步（覆盖下方“先修复”待办）**：四项代码及定向回归已完成；待用户授权提交推送或本机Docker测试。发布时查本轮SHA的CI／GHCR并备份，前后端同版更新，旧缓存页缺expected_updated_at会安全拒绝，需刷新。管理员实际验收两个成员移除一人、剩余者调用、旧表单409及明确恢复；不要把隔离测试镜像当8080部署。全量Go／线上请求根因未验证。

- **包号发布前阻断项**：不要单独发布当前成员移除补丁。先修复deploy/DEDICATED_AUDIT.md中的Key恢复／停用状态不一致、全员状态耦合及旧表单覆盖撤销，再补期望行为回归；管理页补配置诊断。当前仅完成审查，待实施；严格独占及历史隔离保留，生产历史403/503仍无现场证据。不自动提交推送部署。

- **当前优先**：移除包号成员400的修复待用户授权提交／推送或本机部署；不要将sha-43b715a当作已含此修复的版本。发布后验两个用户直接移除一个、剩余者请求可用及被移除者Key停用；旧版不要仅停用Key就误称可以保存（旧校验仍计入停用Key）。

- **发布最后门槛**：代码43b715a15及镜像sha-43b715a已推送发布；仍需确认CI36657101544的集成测试终态。本轮不能标全绿；成功后用户可按UPDATE_GUIDE备份并运行 `./update.sh sha-43b715a`。生产403/503仍无现场日志，不宣称本次解决历史故障。

- **本轮提交推送已获授权**：推送后核实对应代码SHA的CI／安全扫描／GHCR，再提供固定镜像标签；迁移242发布前备份数据库，保持独占规则，不直接回滚旧单用户镜像。

- **2026-09-30后续授权部署完成**：8080现已更新，本轮此前“需授权本机部署”待办关闭。用户可登录测试游客入口隐藏、FAQ右侧、包号成员／改绑／删除与费用显示；提交推送及生产仍未执行，不能拿生产站点验本机新版。

- **2026-09-30 本轮待验收**：管理员真机确认多人选择、名称、重选账号／分组和删除流程，深浅主题与手机；用户确认FAQ留在右侧及智力效率费用展示。当前源码未部署到8080或生产，下一步需用户授权提交／推送或本机部署；发布时查本轮SHA的CI，不沿用fa45b20。242上线先备份，不直接回退旧单用户镜像。两个生产错误缺日志，保持原因未定，不绕过安全校验。
- [x] **游客功能发布验证**：代码fa45b2023的CI（含Go单测／集成）、安全扫描、GHCR全绿；使用镜像sha-fa45b20。旧记录「Go单测仍待」已由本条取代。
- [ ] **游客功能生产更新**：由用户备份后在/sub2api-deploy执行./update.sh sha-fa45b20；生产首页配置、完整登录选购与支付恢复仍需用户验收，不自动下单或部署。
- [ ] **游客适配复核**：宽屏偏左问题已修并部署8080（ea6e3a5ecc3c），请用户以截图对应窗口刷新确认；不把旧版无溢出检查等同完整视觉验收。

- 魔法配方创作扩展：待用户体验生图／视频配方并确认视觉；需另行授权模型实测（8份共32个设计样例，建议至少64次，尚未执行）。仍需验证系统剪贴板与TXT下载落地；本轮不部署。
- [x] **游客控制台预览入口**：首页主CTA与公共导航已添加，独立只读预览完成并通过179项定向回归与浏览器验收。
- [x] **游客控制台预览本机发布**：用户授权后已部署8080，镜像ecb3961845e8；新版入口、个人接口匿名认证、数据保留及浏览器交互通过。待用户操作与视觉确认。
- [x] **配方教程**：已将使用流程、外部AI运行步骤和常见问题加入页面，27项回归及真实浏览器往返／手机检查通过；不代表新增模型执行能力。
- [x] **游客官网本机Docker部署**：8080已运行`e7a582285cbd`，Go embed构建成功，三服务健康；公开真实套餐／FAQ／登录提示验收通过，备份和旧镜像保留。
- [ ] **游客官网剩余验收**：Go unit测试仍待本轮未来提交CI；登录后选中套餐、注册回跳和支付恢复完整联调尚未实测，不做未经确认的下单扣款。用户确认公开套餐文案／权益及最终视觉；生产仍由用户授权后手动升级。
- [x] **配方确认框**：已用主题内弹框替代覆盖／清空的原生confirm，24项回归及真实浏览器取消／键盘／手机检查通过；视觉由用户确认，刷新旧页前先保存输入。
- [x] **游客官网源码**：公开首页／套餐／FAQ及可取消登录提示已完成，163项定向测试通过，保持当前品牌，不开放个人数据。
- [ ] **游客官网后端与业务验收**：本机8080新后端已部署，真实目录和匿名认证边界通过，替代此前仅4175预览阶段。剩余关闭开关、登录／注册后的完整回跳、确认购买和支付恢复联调见上方待办；未来提交必须检查对应CI，不沿用旧结果。
- [ ] **游客官网发布确认**：用户确认首页／FAQ文案与布局；目标站点检查home_content是否覆盖官网。无自动生产部署，未确认运营政策不发布承诺。
- [x] **配方交互原型**：独立页面已生成并本地启动，入口product-samples/magic-recipes/site/index.html，维护说明见site/README.md；19项核心／DOM回归通过，不等同模型效果或生产验收。
- [ ] **配方原型验收**：用户确认界面及6份流程，在实际目标AI工具检查粘贴与TXT文件下载；浏览器自动化尚不能证实系统剪贴板／文件落地。最终托管与是否接中转站另行决定，未经授权不部署生产或开公共模型执行。
- [x] **大学配方样品**：学生定位限定大学；已替换为论文精读、编程实验复盘、期末复习，工作系列不变。入口product-samples/magic-recipes/README.md，六份内容草稿，非已上线产品。
- [ ] **大学配方评审**：只在大学范围确认优先专业／课程，不再扩展中小学。先核对内容与难度，再按README执行授权模型测试；多轮配方需验证理解反馈、修正和真实实验记录边界。六份共24个样例、至少48次首轮抽样，付费调用须另获授权，不用文档检查代替学习效果。
- [ ] **日夜海报定稿**：雾钛青深色与雾白日间两版已新增官网／微信并排双二维码（with-wechat目录），微信码标签按用户要求更新为「客服咨询」，待用户确认；正式发布前手机实扫两个二维码，确认目标与有效性。原图与此前单二维码成品均保留。
- [x] **魔法配方首批内容样品**：3份已完成，入口product-samples/magic-recipes/README.md，当前为待评审草稿，不是自动化程序。
- [ ] **魔法配方评审与实测**：用户已认可内容方向；现为工作3份＋大学3份，按README的24个样例至少各运行两次，须先授权模型及预算，留存输出与失败记录；验证通过后才讨论制作自动化及发布。未经授权不开放免费执行或托管用户密钥。
- [x] **第三产品方向初选**：用户选择魔法配方并授权先做3份样品；此前魔法实验室仅为历史建议，不继续实施。获客人群、实际转化与平台开发仍未确认，不承诺知名度增长。
- [ ] 修正版镜像sha-a1b9137已发布，等待CI36375537042后端单测／集成测试完成；仅此项仍未确认。生产升级仍由用户手动备份执行，不沿用首轮失败CI或旧镜像标签。
- [ ] 本次包号发布：核对功能提交对应CI和GHCR成功后提供固定sha镜像标签；生产由用户备份并手动升级，真实账号联调仍待。
- [ ] 本机新版已部署：用户验证包号管理／我的专属账号及推广清理；此前“授权部署后验收”现进入验收阶段。真实账号隔离、额度采样、续期撤销仍待联调；Go单测仍待执行。
- [ ] 推广清理：授权部署后验收用户菜单、首页／用量页脚及三处代理选择区域；当前仅源码变化，运行站点未更新。
- [x] **包号首期源码**：用户已选Claude与ChatGPT/Codex；管理员绑定／续期／撤销、用户状态与额度、调度隔离和回归代码已加入，文档见DEDICATED_ACCOUNTS.md。仅标准模式，复用手工预配置的独占分组，不改变计费。
- [ ] **包号上线前验证**：先在Go1.27环境gofmt本轮变更并核对Wire，再执行Go单测／集成／lint／embed构建和真实数据库迁移；取得授权后本机新镜像验收两名用户隔离、旧Key／备用路由／影子账号拒绝、HTTP／WS到期撤销、Claude／Codex真实额度、深浅色与手机。当前未提交推送／未部署，不沿用旧CI；启用绑定后旧镜像没有隔离保护，回滚先暂停相关账号与Key。
- [ ] **宣传海报用户确认**：2026-09-28雾钛青竖版初稿已交付，待确认文案/版式；对外发布前手机实扫并确认目标站点。未自动上传、发布或发送给第三方。
- [x] **SynaRoute咨询范围与初查**：官网和客户端均已检查，结果见第四节2026-09-27只读评估；是否实施及改动范围待用户决定，真实后端/安装包验收未执行。
- [ ] **异常登录现场取证**：取得告警所属系统（本站/SSH/其他）、时间时区、3–5条脱敏记录；保留并关联生产审计与反代日志的action/path/status_code/actor/auth_method/IP/UA/request_id，区分失败尝试、2FA中间状态及非本人成功操作。核对生产镜像、代理头清洗与源站暴露、管理员API Key是否启用及使用记录；不要发送密码、完整Token、Cookie、密钥或.env。证据不足不得宣称安全或已被入侵；未核对代理拓扑前不直接切换转发IP开关。
- [x] **说明订单菜单显示条件**：已对照上游菜单、管理员设置store、设置表单及路由守卫；无需为解释开关改业务或重建镜像。
- [ ] **两段说明精简上线**：源代码与回归断言已修改，待本机镜像重建及视觉验收；不要将当前8080页面当成此轮新版。
- [ ] **核验用户实际菜单展开／套餐直达页**：先区分「订阅管理」（用户订阅分配）与折叠的「订单管理」→「订阅套餐」（商品）。直达`http://127.0.0.1:8080/admin/orders/plans`。若展开后仍缺子项或直达页无「创建套餐」，须取实际页面／控制台证据再定位；不能用源码存在或HTTP200代替交互验证。
- [x] **智力页面反馈修正（代码）**：移除外链与频率提示、明确矩阵排序并增加选项、改为可读三指标对比；修复支付居中容器宽度适配。76项回归通过，真实视觉仍待。
- [x] **反馈修正版本机更新**：8080运行`cfa0a12c8bf6`；镜像一致、三服务健康、HTTP新资源、订阅新分组与全部设置保持检查通过。无收费套餐创建，无生产操作。
- [x] **新增智力效率源代码**：接入用户指定JSON，提供矩阵／效率图、图标、30分钟浏览器同步与失败保留；真实数据解析及70项定向回归通过，详见CUSTOMIZATIONS。
- [x] **智力效率本机Docker发布**：用户已授权且更新完成，8080运行`1a7a31d0e424`；构建、镜像一致性、三服务健康、HTTP新资源与数据／设置保持检查通过，旧镜像及数据库备份保留。
- [ ] **智力效率视觉验收**：浏览器控制工具认证错误，仍需用户在真实浏览器检查登录、深浅色／手机／筛选／对比分页／网络错误及精简说明。用户已授权本轮提交推送，生产由用户更新，不以旧本机镜像验收最新代码。
- [ ] **配置订阅套餐**：本机原生订阅功能已开启但套餐0个；用户需给出分组、售价、币种、周期、日／周／月额度及续费规则后再配置。不要照搬参考截图，不创建测试售卖套餐或实际订单。
- [x] **本轮视觉收口**：用户明确「暂时先这样」，保留统一认证双栏、雾钛青、核心M与扁平Logo、魔法家族／多模型 API 服务；不继续自动改风格。这是阶段性接受，不等于全部业务验收完成。
- [x] **代码及交接整理**：本轮用户授权提交并推送到自己的origin/main；不推upstream、不发布生产。已有文档纳入版本管理，本机.env／override／数据库备份／设置快照／日志不入仓。推送结果以Git远端HEAD核验为准。
- [x] **新Logo复制**：C:/Users/Administrator/Pictures/mofamily/mofa-mark-flat.png，16313字节；源文件与副本SHA256一致，旧mofa-logo-local.png保留。
- [x] **本机部署与已测范围**：镜像b4fa2d8e97a1仍运行8080；真实登录、空态业务页、认证多尺寸及日夜主题已测。品牌名称／副标题／Logo来自本机设置，非代码硬编码。
- [x] **维护边界与故障记录**：展示接缝已登记CUSTOMIZATIONS和custom/UPGRADE；Docker代理防复发见deploy/LOCAL_DOCKER_RUNBOOK.md。上游仍需合并核对，不能承诺零冲突。
- [x] **本轮提交推送及镜像发布**：功能提交d592f9481已推送，GHCR镜像sha-d592f94构建推送成功；只用此标签升级，不使用随后纯文档提交的SHA猜镜像标签。
- [x] **本轮CI最后结果**：d592f9481对应CI运行36264121775最终success（含Go单测／集成测试），安全扫描36264121782及镜像36264121792均success；通过GitHub API针对同一功能SHA核验。本机未跑Go单测，不混淆本地与CI结果。
- [ ] **带数据业务验收**：在获授权环境检验真实密钥／用量表格、分页、图表、错误／警告路径和上游请求；空库截图与临时DOM探针不能代替。
- [ ] **首次引导与完整交互回归**：完整走查引导、键盘／焦点、长文案、加载与错误状态；目前仅验证关闭引导后关键操作。
- [ ] **生产升级（用户操作）**：本轮用户确认自行Docker更新；待本次GHCR构建成功后，按deploy/DEPLOY_CUSTOM.md及deploy/UPDATE_GUIDE.md备份数据库、记录旧镜像、指定新sha标签升级、健康检查并准备回退。检查智力效率新入口／排序／对比／两段说明，以及支付开启后的订阅套餐配置。站点名／Logo需在生产独立配置；推送不等于部署。
- [ ] **后续独立任务**：SynaRoute保留紫色，留用户新会话；可选Logo矢量重绘及进一步品牌体验打磨，不自行开启。
- [x] **升级方式保持Docker + update.sh**：不采用App内在线更新，上游预编译二进制会覆盖定制；普通重启不会还原可写层，镜像重建才会，历史澄清保持有效。

## 六、文档地图
- [BRAND_IMPLEMENTATION.md](BRAND_IMPLEMENTATION.md) — **最新已批准设计的执行计划**：素材路径、色表、阶段顺序、代码接线、防坑、验收与新对话提示。
- [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) — **定制唯一权威清单**（新增文件 / 接线改动 / 行为修改 / 自检清单）。
- [SYNC.md](SYNC.md) — 同步上游 + 构建镜像 + 部署 runbook + 4 条红线。
- [deploy/DEPLOY_CUSTOM.md](deploy/DEPLOY_CUSTOM.md) — 定制镜像部署清单（登录/起服/升级/回滚/迁移）。
- [deploy/UPDATE_GUIDE.md](deploy/UPDATE_GUIDE.md) — 日常更新 / 回滚操作速查卡（模式 A latest / B 钉 sha、常见坑）。
- [frontend/src/custom/README.md](frontend/src/custom/README.md) — 前端叠加层三种用法 + 影子替换代价。
