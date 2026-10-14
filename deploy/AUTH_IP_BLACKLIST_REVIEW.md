# IP认证防护审查与本地验收（2026-10-13）

2026-10-14 上游0.2.16发布终态：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。

2026-10-14 同步上游0.2.16：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

2026-10-14 网段防护发布终态：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

## 2026-10-14 网段防护本轮

已核实：仅新增/扩展自有代码和254迁移，手动通配符/CIDR规范化，自动按配置网段共享Redis计数并持久封整段。自审覆盖真实单IP与封禁目标分离、INET包含/重叠和规范边界、配置version/目标generation/重叠规则状态摘要、策略事务锁顺序、管理员鉴权/审计与前端确认，未发现阻塞问题。旧在途失败不能覆盖解封或范围变更；其他覆盖规则仍独立有效。只保护已有敏感认证，成功不计数；无法从同网段推定全部IP均恶意，生产误拦率未知。

- Docker Go1.27.2：18项后端专项含真实PG18/Redis8及7组网络回归、全量unit、Go embed和嵌入产品边界通过。截图10时间点以不同IP回放，在20:32:32第10次封61.120.9.0/24；实际分类仍要求明确失败，截图400不能证明全部计入。新测试覆盖不同IP/邻网段、成功不计数、24并发只封一次、重启/Redis清空、IPv6、到期、单IP切换及旧请求竞争。
- 前端16项管理页回归、类型/定向ESLint/Vite通过，通配符确认展示CIDR/256地址范围、范围/前缀保存和非法值拦截。
- 独立8081/auth_ip_smoke的44条真实HTTP断言通过：6次成功登录名单0，两个IP密码失败加第三IP注册格式异常封/24，网段其他IP登录/验证码403，相邻网段200；模型入口401、已登录管理员me/管理/解封200，旧解封409。重启/Redis DB1清空仍封；关闭自动后手动通配符仍生效，重复覆盖409，无效中间星号400，临时到期后首次失败重新累计。8条相关审计含规范网段和范围配置；转发头模拟仅用于隔离测试，已对照真实审计client_ip验证。
- 8080新镜像b33cb472be2a、容器0fa7d82eda3d，三服务healthy；原业务计数/281设置摘要/卷及PG/Redis/recipes身份保持，version3→4和network/24/64为迁移预期；自动仍false、名单0。真实管理员登录/新接口/新chunk BznUWgMC通过。备份665933字节、pg_restore目录1282行、SHA256ff83db7ad955edffcc47e6cffc904c1eafa5a115d26a36592e6d085916e141af；旧镜像保留。

证据output/auth-ip-network-20261014：backend-check-final.log、frontend-check.log、build-release/local.log、smoke.json、local-before/after/verified.json和数据库备份。首次启动隔离应用失败为测试TOTP参数非HEX，已修正；HTTP脚本错用不存在的send-email-code路由已改为send-verify-code后重跑通过；后端初次测试绕过service直接传星号给repo导致失败，改用guard.List后专项/全量通过。正式本轮SHA/CI/安全/GHCR须另行核验，生产未代部署。


2026-10-14 手动封禁发布终态：功能065eee935a5eea3fb023c5faf469b988feb0fa73已推origin/main；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38497226115)（unit/integration/PG18/Redis8/lint0/前端38文件502项含IP管理14项）、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38497226138)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38497226114)全部成功。固定镜像sha-065eee9（linux/amd64），摘要sha256:8da3604012c2a2c121dd47a7fa22542c785b5c77c60335177498a40058807d38，registry/镜像revision/source/构建日志及23文件快照哈希交叉验证通过，verified-release.json记录终态。8080随后已备份更新至e9fd3b9c7ea2/112691d6fc2b，三服务健康，10条HTTP及真实管理员登录/new POST400/新chunk CoRkhMZ9通过；原PG/Redis/挂载、业务计数/防护配置相同，保留原本地智力效率源码，正式镜像排除其他会话改动。8081独立测试容器/网络/匿名卷已清理，local-verified.json/smoke.json及本SHAlogs留存；生产未部署，以下隔离阶段“未操作8080/待新SHA”表述为历史，不据此推断当前状态。

## 2026-10-14 管理员手动封禁（本轮）

已核实：Docker Go1.27.2中15项认证专项（含真实PG18/Redis8）和全量unit通过；前端14项管理交互、类型、定向ESLint及Vite构建通过。实际镜像Go embed及产品路由通过；8081独立auth_ip_smoke库32条HTTP断言通过，关闭自动开关仍手动拦截注册/登录/验证码/找回/Passkey/2FA/OAuth，重启和Redis清空后仍有效，管理员JWT管理和auth/me仍200，模型入口保持原401，手动审计3条均有管理员身份/原因/来源。永久、45分钟、到期、IPv6、非法输入、重复409和旧解封版本均实测。静态审查未发现阻塞问题；本轮功能不操作原8080数据或配置，生产未部署，新SHA CI/安全/GHCR仍须验证。

增加完整IPv4/IPv6手动新增、永久/1–1440分钟临时期限、必填原因、确认和来源展示。独立于自动开关，不扩大模型/用户/管理员接口范围。迁移253保留旧auto数据并约束manual无失败统计；旧250–252未改。重复有效封禁409，不静默缩短原期限；共享IP事务锁和generation，覆盖并发添加、旧自动请求、旧解封状态、到期恢复与重新自动封禁清理旧原因的验证。

验证与发布证据目录 `output/auth-ip-manual-20261014`。本轮隔离快照基于06873209c，只叠加手动封禁文件；付款/库存248/249、智力效率、投稿补修和前端依赖升级均排除。下方07bcef204发布证据只适用于原自动封禁版本，不能据此宣称手动功能已发布。生产部署及生产真实IP取证仍由用户完成。

## 2026-10-14发布验真

功能`07bcef204d09349e0574b51a5b8cfdd6266288a3`已独立推送origin/main，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252759)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252800)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252762)全部成功；对应SHA／所有job conclusion和日志已保存并断言。

- CI unit／integration通过；Unit tests传专用AUTH_IP_TEST_POSTGRES_DSN（日志凭据自动脱敏）／AUTH_IP_TEST_REDIS_ADDR，专用PG18／Redis8健康，真实存储测试不会因DSN缺失skip；包号测试仍使用另一服务。
- golangci-lint2.14.0报告0 issues，Go1.27.2；前端类型及38文件497项（含IP管理9项）通过。govulncheck零可达漏洞，另有2项未调用模块漏洞，前端审计按已登记例外通过，不宣称所有依赖零漏洞。
- 发布镜像`ghcr.io/engineemomo/sub2apicust:sha-07bcef2`，linux/amd64；摘要`sha256:81bcf46cd8e1008f058c296f09d32af059d0e7fadefa66cf355b522885d5bc13`。buildx imagetools registry输出、镜像revision=完整功能SHA／source=本仓与构建日志containerimage.digest一致；未启动本机Docker引擎或部署生产。镜像版本0.2.13-custom.07bcef204。
- 发布快照只含本功能、文档和CI真实存储接线；18个功能源文件与10-13已测试v3哈希一致。router只暂存认证防护接缝，Makefile只加入IP spec；未发布其他付款／库存248/249、投稿、依赖升级或智力效率，工作区内容保留。

证据`output/auth-ip-release-20261014`的verified-release.json、同SHA runs／jobs、最终test／lint／frontend／安全／build日志、image-inspect.txt／image-config.json及manifest。用户备份数据库、部署配置和app_data后执行`./update.sh sha-07bcef2`；默认自动拉黑关闭，生产真实IP／共享出口失败分布和带数据交互待验。下方10-13为本地历史取证。

## 已核实

用户最新要求允许永久封禁／找管理员解除，或配置封禁期限。保留明确失败计数，成功登录／注册／正常多步骤不计数；验证码／TOTP、业务校验、令牌失效、429和系统故障排除。新增block_mode=timed/permanent，默认10次失败／120分钟、timed15分钟；临时到期恢复，永久须管理员移除。规则切换不改变已有记录期限，开关仍关闭。此授权覆盖上一轮只新增临时封禁的决定。规则、API和分类边界见[功能说明](AUTH_IP_BLACKLIST.md)。

审查涵盖认证路由范围、响应分类与缓冲、Redis原子窗口／入场时间／唯一事件、数据库规则与状态版本竞争、管理员权限／审计、模式校验／期限保存、前端表单／409／时区及迁移。未发现阻塞性问题。PUT遗漏模式400以避免旧客户端静默降级；旧version不能覆盖模式、旧generation不能重封。已有模型接口和管理会话不接入名单，永久记录不自动解除。

- Docker Go1.27.2：13项专项、真实PostgreSQL18／Redis8和全量`go test -tags=unit ./...`通过；运行镜像Go embed及嵌入产品路由回归通过。
- 连续32次成功请求不计数；明确失败分类、响应替换、256KiB以上成功响应透传、入场时间／窗口边界、80路并发、管理员解封／到期恢复／旧请求版本竞争均通过回归。
- 实际镜像前端依赖：5文件55项通过（IP管理9项），类型检查／定向ESLint／Vite构建通过；永久输入及重新加载、切回临时保留时长有回归。
- 专用guard_smoke库8081实际镜像：阈值3，9次成功登录均200、名单0条；第3次失败403永久封禁、blocked_until=NULL。实际应用重启和临时Redis FLUSHALL后仍拒绝；持JWT的管理员解除200、旧状态409。切换临时45分钟不解除原永久记录；解封后第3次失败生成约2700秒期限，切回永久不延长旧临时期限。只在测试库将到期移到过去，下一次正确登录200。真实审计2条自动封禁、1条成功解除、2条永久模式规则更新；测试账户合规确认未改变原账户。
- 原8080真实管理员登录及config/entries200，管理接口匿名401，模型入口仍401；新chunk为AuthIPBlacklistView-DotqpFXJ.js，入口index-Cgh4f6-U.js，HTTP校验新版两种方式文案通过。
- 仅替换应用容器589ddc6d3ae7；镜像与标签sub2apicust:auth-ip-mode-20261013一致：sha256:7f423b3890c9a070657c0097adf7b486499ed042cc8e3a350a2266caf1b2bad5。三服务healthy，原PG/Redis/recipes容器及全部原挂载保持，recipes停止。
- 250/251/252迁移生效，总迁移301，248/249未带入；用户3／密钥0／分组3／套餐1／订单0，281项设置完整摘要保持。名单0条，enabled=false、threshold=10、window_seconds=7200、block_seconds=900、block_mode=timed、version=3。
- 构建快照为HEAD a6cdbd870加本功能及原有智力效率，不带其他会话未提交业务／依赖改动。3只测试容器、其匿名卷及专用网络已清理。

证据目录：output/auth-ip-local-20261013。backend-mode-tests.log、frontend-mode-tests.log、build-mode.log、mode-isolated-smoke.json、mode-audit.json、mode-http-smoke.json、mode-before/after.json及cleanup-mode.txt对应上述结论。构建源docker-src-v3及manifest记录31个叠加文件，使用已发布HEAD的前端依赖，未部署其他会话的Vite/Vitest升级；此前cooldown证据保留作为历史。

部署前两份pg_dump均通过pg_restore目录核验：database.dump 655269字节，SHA256 BF50B77C63EC935F7565D55AFE28FEF9B6B06135344AF6F1D2B568FCFD9A958F；database-before-cooldown.dump 661963字节，SHA256 464A3F7A1A3AA3EECC920006BDD172A58230E6C0D3E60B264DE40A3A2431CBBD。初始旧镜像sub2apicust:before-auth-ip-20261013-221424不含本防护。251前的阶段旧镜像sub2apicust:before-auth-ip-cooldown-20261013-225403是旧永久封禁实现，不能把它当作支持临时名单到期的回退版本。回退先核实现场，不自动恢复数据库或删除卷。

## 252迁移前备份

252前新备份database-before-mode.dump为662966字节，SHA256 DFF741233A5D9FA6C649479F45594C27682EF014A25BE0DBD08B9B5E303A22CA，pg_restore目录检查通过。保留33e2ba735e66为sub2apicust:before-auth-ip-mode-20261013；此旧版可读取永久名单但无法设置模式，回退后修改规则可能继续生成临时封禁，需先核对现场。迁移不回退、不删数据卷。

## 推断与未知

失败分类减少正常重复操作、多步骤认证和验证码错误造成的误拦，不等于共享IP零误拦。多人在同一出口输错密码仍共同累计；永久模式下误拦须管理员解除。生产合理阈值和误拦率未知，开启前按审计与真实流量验证。启动日志的独立定价更新源连接0.0.0.0失败未在本轮修复，不宣称远程定价刷新成功。

生产Cloudflare→Nginx真实IP、转发头清洗／源站限制未取证。截图400缺少响应reason／正文，新规则不能承诺10条全部符合计数条件。原管理员页面带真实记录的浏览器交互仍待用户验收；HTTP接口与前端单测不等同完整人工验收。宿主无Go，10-13使用Docker工具链；10-14发布和CI／安全／GHCR／lint已通过，生产未代部署。
