# 日常更新操作手册（定制版 sub2api）

## 2026-10-19 上游v0.2.17同步（当前可更新版）

2026-10-19上游v0.2.17同步发布终态：业务5233e9090698a9bd02f645340cd19e85e7e27743已推origin/main，包含upstream/main=b0a02a3304a0a78cdd003ddda1dd23ba02c03084及新增47条上游提交；同SHA CI 39196526010（Linux unit/integration、专用PG18/Redis8、Go lint 0 issues、前端53文件831项）、安全39196525919、GHCR 39196526008全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-5233e90（linux/amd64），index digest sha256:08fe2367c8241ca9b43fe741b1ccd03db65ba5a8a5579f8bbc4a8f4ce6d2c550；registry revision/source/摘要与正式embed构建日志交叉验真，正式镜像隔离28项HTTP检查含真实登录、角色权限及IP防护通过。无新增迁移或依赖；312个自有新增文件内容不变，支付两处冲突保留WAP期限/快捷金额定制，原未提交产品工作不混入。生产由用户备份后手动./update.sh sha-5233e90，未代部署或付费生图。PR8028仍未合并，本轮不包含，不能保证解决截图无图。证据output/upstream-sync-20261019/verified-release.json、release-source-proof.json、formal-smoke.json及job日志；Windows全量5项失败如实保留在backend-unit-summary.json，正式Linux CI通过覆盖本轮发布门禁。后续纯文档skip-ci提交不改变此业务镜像。

备份数据库、配置与app_data后，在服务器部署目录执行：

```bash
./update.sh sha-5233e90
```

更新后验收登录、业务接口与支付/退款；本轮没有新增迁移，现有分组到期平移及收款保护规则沿用。截图无图问题尚未取得现场响应，本版不包含未合并的PR8028。

2026-10-18认证IP缺凭证计数发布终态：业务`117e9a1a08f1a225dbe5b041ae243be29251770b`已推origin/main；同SHA CI `39138612694`（unit／integration、专用PG18／Redis8、Go lint 0 issues、前端53文件830项）、安全`39138612602`、GHCR `39138612579`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-117e9a1`（linux/amd64），index digest `sha256:166ede3f3a5aa77028b4479f754e353442285fa28b466d426a0194c5eb8e6896`；registry revision／source／摘要与正式embed日志交叉验真，IP管理chunk `AuthIPBlacklistView-b89c0411.js`。缺失必要人机凭证的400及包含captcha／TOTP字段名的格式错误现按规则累计，正常挑战失败／腾讯灾备票据／服务故障／成功操作仍不计数；原自动开关／阈值／范围／期限不变，无新迁移、不回灌历史。用户备份后手动`./update.sh sha-117e9a1`并确认自动规则。未代部署、未实际封禁189.24.123.4；生产私有规则／实际版本／历史响应原因仍未核实，复现路径不等于现场唯一根因。证据`output/auth-login-probe-20261018/verified-release.json`、`release-source-proof.json`、`local-verification.json`及job日志。本轮便携Go本地embed通过，覆盖历史无Go未编译判断；后续纯文档skip-ci不改变业务镜像，其他会话工作保留。

## 2026-10-18 独立订单核账菜单（当前可更新版）

2026-10-18订单核账菜单发布终态优先：业务提交`57aad622fa4cdeea90df1d22866b6b514a92e94b`已推送origin/main；同SHA CI `39135534896`（unit／integration、Go lint 0 issues、前端53文件830项）、安全`39135534923`、GHCR `39135534858`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-57aad62`，平台`linux/amd64`，index digest `sha256:f1e11005ea3fbd2f7beeeeff44ad94b2e8794568c8c71be64f6f2ee86515120c`；registry revision／source／平台／摘要与正式构建日志交叉核验，新页面chunk `AdminOrderReconciliationView-T4m7REbz.js`及同提交embed后端均已打包。新入口为管理员“订单管理 → 订单核账”，设置→支付保留限额／自动清理规则／联系方式及跳转；无新增迁移，原付款／等待期／审计保护不变。未代部署生产或处理真实订单，真实管理员／手机视觉／商户核账仍待用户更新验收。证据`output/order-reconciliation-release-20261018/verified-release.json`、`release-final-verification.json`、`release-source-proof.json`及job日志。后续纯文档`[skip ci]`提交不改变此业务镜像，下方未提交／待发布为历史阶段。

备份数据库、配置与app_data后，在服务器部署目录执行：

```bash
./update.sh sha-57aad62
```

更新后进入“订单管理 → 订单核账”进行批量复查、渠道关单、人工释放或登记线下收款；自动清理开关与等待时长仍在“设置 → 支付”独立保存，等待期与商户账单确认要求保持原规则。

## 2026-10-17 支付通道删除保护与历史占用自动清理（当前可更新版）

最终业务提交`16ef9172cf666e6361f17161aa83a53178ee0f3f`已推送origin/main；同SHA CI `39029645317`（unit／integration、专用PG18、Go lint 0 issues、前端53文件824项）、安全`39029645211`、GHCR `39029645261`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-16ef917`，平台`linux/amd64`，index digest `sha256:49b4285ec235151d36459e3cdd49320103fffa0864d1c552351378a43ef64d07`；registry revision／source／平台／摘要及正式embed构建日志已交叉核验。

按现有流程备份数据库、部署配置和app_data后，在实际部署目录执行：

```bash
./update.sh sha-16ef917
```

更新后在设置→支付开启“原支付通道已删除时，超期占用自动清理”并保存（默认关闭）；已取消／过期／失败、未记录付款且原实例确实不存在的历史占用，在期限＋配置等待（默认24小时）＋10分钟后按每60秒最多20笔处理，失败仍有30分钟退避，可点一次“立即批量复查”。只释放占用，不充值或发订阅，实际收款须商户账单核实。

支付补丁没有新增迁移；通道仍存在但配置／查询失败、已付款、引用无效／冲突及等待未满的订单继续保留，不换绑历史订单。更新后刷新管理员页面，检查清理审计和占用下降。证据`output/payment-auto-release-20261017/verified-release.json`、`release-final-verification.json`及同目录job日志。首轮2af3c432c的unused已补修并保留失败记录；不使用首轮候选交付。未代部署生产或处理真实订单，其他任务未提交工作保留。后续纯文档`[skip ci]`提交不改变该业务镜像。以下旧标签及未发布状态为历史阶段。

## 2026-10-15 分组活动显示、到期平移删除及R1–R4补修（当前可更新版）

最终业务提交`0e99748ba82575442aae95bda1f6358a43c46ddf`已推送origin/main；同SHA CI `38750224990`（unit／integration、专用PG18、Go lint 0 issues、前端53文件821项）、安全`38750225143`、GHCR `38750224942`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-0e99748`，平台`linux/amd64`，index digest `sha256:9b3b82a8daeb04fa7d0816825f2401e9732c027746ec0ea62d25e2f876525251`；registry revision／source／平台／摘要与构建日志交叉验真。

含R1–R4、管理员当前活动倍率及基础说明、分组到期事务平移后软删除与迁移258；福利时段结束仍只恢复基础倍率，沿用原计费逻辑。**新版启动会补处理已有到期规则，用户须先备份数据库／配置／app_data并审阅源组及平移目标，再执行`./update.sh sha-0e99748`。**生产由用户手动更新，真实收费与到期平移现场验收仍待；20:30历史故障根因未知。

```bash
./update.sh sha-0e99748
```

证据在`output/group-access-fix-20261015/verified-release.json`、`release-source-proof.json`及同目录job日志；首轮c4fa7fd的ineffassign失败已补修，旧候选不作交付。后续`[skip ci]`纯文档提交不改变镜像业务版本。以下旧版标签与未发布状态是历史阶段。

分组到期行为（本版已发布）：福利时段结束只恢复基础倍率；整个分组到期由后台事务迁移未删除Key与对应用户／显式授权，再软删原组，保留Key属性、余额及历史账单。只支持同平台按量组，订阅与包号拒绝自动合并权益。258防并发绑回已删组，257保持原校验和；前后端须同版。**新版启动即补处理已有到期规则；不需要自动删组的先关闭分组到期设置，可保留福利时段。**失败保留原组并30秒兜底重试，日志含源／目标ID和迁移数量。

列表与现场边界：管理列表显示当前活动倍率并注明基础倍率，新增管理员批量时段接口，R1–R4已随本版修复。test线上下拉只读实测活动0.8x、21:17:27后刷新为1x；真实账单未验证，20:30故障不能由用户事后删除的规则倒推。本轮未代部署或改真实规则。

2026-10-15 20:26历史审查：`sha-7944797`对应源码的降级绕过目标门槛、通知心跳互等、RPM逐次查询和已付费图片GET误拦均已在`sha-0e99748`补修并回归，详情见[GROUP_ACCESS_REVIEW.md](GROUP_ACCESS_REVIEW.md)。原审查发现和首轮失败保留用于追溯，不作为当前发布状态。

## 2026-10-15 分组门槛、到期、缓存与分时福利（已发布）

最终发布提交`794479760593961c08ea842095543ec1d0bd58c5`，同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38727724771)（unit／integration、专用PG18、Go lint 0、前端52文件815项）／[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38727724898)／[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38727724764)全部成功。固定`ghcr.io/engineemomo/sub2apicust:sha-7944797`（linux/amd64），摘要`sha256:34cf7d44c87608d92964329571250f65127b20dc3ca5cc99f7bf582533e4b0c5`，registry的revision／source与正式embed构建日志已验真。下方未提交／待发布是历史阶段；纯文档skip-ci提交不生成替代镜像。

先备份数据库、部署配置和app_data，再由用户在实际部署目录执行：

```bash
./update.sh sha-7944797
```

升级自动执行最终257迁移，默认不开启已有分组规则。刷新后台进入「分组管理 → 开放规则」，两个USD额度门槛可分别开关，同时开启为AND；指定日期的分时福利可填三段示例，23点换组须选兜底。先用测试分组把时段改成未来几分钟，验证保存／单独关闭／原Key新请求倍率和到期切换；再验真实退款资格与实际账单。退款或余额失去资格仍拒绝，只有到期触发兜底；WS跨分组／基础倍率变化需重连。详见[操作说明](GROUP_ACCESS_RULES.md)，未代部署8080或生产、未配置真实活动；生产吞吐和商户退款待验。

2026-10-15发布授权补充：用户已要求更新测试，分组门槛／缓存／分时福利进入本轮提交与镜像门禁，最终固定标签及同SHA CI／安全／GHCR以HANDOFF最新状态为准。用户在实际部署目录备份数据库／配置／app_data后执行本轮固定标签更新；未代部署本机8080或生产。下方尚未授权／未提交为历史阶段，不能据此使用旧标签。

2026-10-15后续补充：待发布分组规则现包含资格缓存优化及一次性分时福利倍率。最终257还包含事务通知触发器和活动JSON，应用每实例增加一个LISTEN连接；前后端必须同版。入口、三段福利示例、到期兜底、原有倍率优先级及通知延迟边界见[分组开放规则](GROUP_ACCESS_RULES.md)。最终验证看output/group-access-cache-20261015/verification.json及HANDOFF，当前未提交／推送／发布镜像或部署；旧标签不包含本功能，下方初版测试数属历史。

## 2026-10-15 分组开放门槛与到期切换（源码完成，尚未发布）

新增257迁移及管理员「分组管理 → 开放规则」：月净充值／账户余额门槛独立开关，同时开启须同时满足；退款后原Key也会重查资格。分组可设统一到期时间和同平台目标，已有Key下一请求按目标权限及价格自动路由，无需手动改绑。金额为USD账户额度，月份按北京时间自然月；操作与WS边界见[分组开放规则](GROUP_ACCESS_RULES.md)。

本机40项后端专项含PG18、Linux58包unit、前端42项／类型／lint／构建及Linux embed编译通过，证据output/group-access-20261015/verification.json。当前未提交、未推送、无本轮CI或镜像标签；下方所有旧固定标签均不包含本功能。待发布后核验同SHA CI／安全／GHCR，备份数据库／配置／app_data后升级同版前后端及257迁移，再验管理员保存、真实退款资格和原Key到期后实际账单。未代部署本机8080或生产。

## 2026-10-15 搜索流式初始0漏计修复（当前已核验）

功能提交`fb5916a4fe1853995b18f016d0806a2990a4c73b`已推送origin/main，同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38665635231)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38665635268)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38665635087)全部成功，含后端unit/integration、lint、前端及脚本检查。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-fb5916a`（linux/amd64），摘要`sha256:0c33d76be7e126af2e171d7d9086b1e84abcbe38cfde9c28507a1aa982154570`；registry/revision/source及正式embed构建日志已交叉核验。

修复上游`message_start`先报搜索次数0、随后有成功结果、终态缺次数时仅收Token的问题。沿用分组搜索基础价和有效倍率，无新增配置或迁移；旧版升级仍需执行既有256迁移。先备份数据库、配置和app_data并记录旧镜像，由用户在部署目录执行：

```bash
./update.sh sha-fb5916a
```

升级后确认实际镜像、health和前后端同版，检查请求实际所在分组的搜索基础单价。以新请求核验账单中的`web_search`次数和工具附加费，再与上游同一请求对账；基础价0.05、有效倍率0.15、一次成功搜索应加0.0075，输入127/输出529且价卡3/M与15/M时总实扣0.00874740。历史账单不补扣。该金额已由源码回归验证，尚无该笔生产原始响应，不把源码复现当现场根因已确认；若仍无搜索快照，需检查脱敏原始JSON/SSE是否含实际计量或成功结果。

本轮没有代部署生产或替换本机8080，隔离测试镜像/导出容器已清理，原业务镜像和卷保留。证据`output/newapi-ledger-20261014/search-zero-verified-release.json`；后续纯文档skip-ci提交不生成新镜像。

## 2026-10-14 Claude网页搜索工具附加费（当前可更新版）

功能提交`d3a2c346352e7eca5aacc197785fd366f90f83b6`已推送origin/main，同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38589112494)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38589112074)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38589112344)全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-d3a2c34`（linux/amd64），摘要`sha256:5baca94dfd6ad73ea03117406400253c083ff9886addca0e4b80528bb6149ba5`；registry manifest、revision/source标签及构建日志交叉核验。首个候选d83900dc2因DTO测试errcheck未通过，更新使用本段最终固定版本。

新增256迁移，旧账单保持NULL且不补扣。先备份PostgreSQL、部署配置和app_data，记录旧镜像；由用户在实际部署目录执行：

```bash
./update.sh sha-d3a2c34
```

更新后核验实际镜像/health、迁移256和前后端同版。在分组管理的Anthropic/composite分组核对“Claude网页搜索工具附加费”：未设基础单价默认0.01 USD/次、0免费；实际工具次数乘基础单价，再乘Token同一有效倍率，叠加到Token费。确认新请求账单/导出的工具调用数、原始价和实扣价，与上游相同请求对账。配置及覆盖边界见[CLAUDE_TOOL_BILLING.md](CLAUDE_TOOL_BILLING.md)；代码不会自动同步New API价卡，阶梯错价或充值价值差异仍需单独核验。生产与本机8080未代部署，交接纯文档skip-ci不产生另一个功能镜像。

2026-10-14全工作区最新可更新版本：59项改动已推送6406511c687e59b2afd37cf24da3db0d6011bea5，包含上游0.2.16、IP网段防护、248/249收款/库存迁移、工坊配额与清理、智力效率雷达Bench及依赖/脚本升级。同SHA CI38522649457、安全38522649572、GHCR38522649396全绿；正式镜像ghcr.io/engineemomo/sub2apicust:sha-6406511（linux/amd64），摘要sha256:9608581fa3c7e8c9b36e368e7815f58376c9c3ee9d2c48f449e0412f6b2244bb已与registry/源码标签/构建日志交叉核验。先备份数据库/配置/app_data（含studio-submissions），再执行`./update.sh sha-6406511`，前后端同版；真实付款/过期入口/人工释放核验见PAYMENT_COLLECTION_LIMITS，智力效率首轮读取与出网按HANDOFF验收。生产与本机8080未部署本镜像，新版生产业务待用户验收；纯文档skip-ci提交不生成新镜像，下方旧镜像及未发布说明为历史。

2026-10-14 上游0.2.16发布终态：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。

2026-10-14 同步上游0.2.16：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

2026-10-14 网段防护发布终态：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

2026-10-14 网段防护本轮：新版支持管理员通配符210.230.206.*（/24）、CIDR及自动网段累计封禁。正式固定标签及本轮CI/GHCR验真以HANDOFF最新状态为准，旧sha-065eee9只支持完整IP。升级前备份数据库/配置/app_data，保留旧镜像；254保留开关及已有名单、version递增，默认network/IPv4 /24/IPv6 /64，可切单IP或调整前缀。刷新管理页并选择临时/永久及时长，成功不计数；同网段正常用户也可能受封禁影响。生产由用户部署和核验真实客户端IP，本机8080已验证，未代部署生产。


2026-10-14 手动封禁发布终态：新增管理员手动IP封禁、永久/临时、原因和来源；功能065eee935a5eea3fb023c5faf469b988feb0fa73已推，同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38497226115)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38497226138)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38497226114)成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-065eee9`（linux/amd64），摘要`sha256:8da3604012c2a2c121dd47a7fa22542c785b5c77c60335177498a40058807d38`；registry/源码revision/构建日志已验真。备份PostgreSQL、部署配置和app_data，保留旧镜像，然后在实际部署目录运行`./update.sh sha-065eee9`；刷新后台“IP 黑名单 → 手动新增 IP”，升级含253迁移，前后端同版。重复有效封禁409不覆盖，自动开关关闭仍拦截名单，只针对敏感认证。8080已备份更新且三服务健康；生产未代部署，旧sha-07bcef2不含手动入口，以下待发布段已属历史。

2026-10-14 管理员手动封禁正在验证发布：新版本新增253迁移和后台“IP 黑名单 → 手动新增 IP”，支持永久/临时期限和原因；自动开关关闭仍检查名单，已有有效封禁不覆盖。备份后按HANDOFF本轮固定镜像更新，前后端同版；旧sha-07bcef2不含此功能。生产未代部署，真实客户端IP与管理员已知攻击IP仍需现场确认。

## 2026-10-14 敏感认证IP防护（当前可更新版）

功能提交`07bcef204d09349e0574b51a5b8cfdd6266288a3`已推送origin/main。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252759)（真实PG18／Redis8、unit／integration、lint0、前端38文件497项）、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252800)、[镜像](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252762)全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-07bcef2`，linux/amd64，摘要`sha256:81bcf46cd8e1008f058c296f09d32af059d0e7fadefa66cf355b522885d5bc13`，经registry manifest、revision／source标签与构建日志核验。本条覆盖下方官网及本功能未发布历史。

先备份PostgreSQL、部署配置和app_data，记录旧镜像；由用户在部署目录执行：

```bash
./update.sh sha-07bcef2
```

启动时自动应用250/251/252，前后端同版。默认自动拉黑关闭、10次可识别失败／120分钟、临时15分钟；核验真实客户端IP后在管理员“IP黑名单”选择永久（管理员解除）或临时时长并开启。名单只检查注册／登录／2FA／Passkey／验证码／密码找回和重置／OAuth等认证流程，成功不计数，OAuth GET只检查；模型调用、管理和当前用户接口、Token刷新／登出／微信支付OAuth不挂本名单。共享出口失败累计的误拦率未知，历史审计不回灌。其他会话付款／库存248/249、投稿、依赖升级和智力效率改接不在本镜像；生产未代部署。详情见[功能说明](AUTH_IP_BLACKLIST.md)和[验证记录](AUTH_IP_BLACKLIST_REVIEW.md)。

> 2026-10-13 23:31 IP防护已部署本机8080（7f423b3890c9），尚未发布生产镜像。新增250/251/252迁移；成功操作不计数，仅明确失败累计，默认10次/120分钟、临时15分钟，开关关闭。可选择永久封禁／管理员解除，或临时时长1–1440分钟；规则切换不修改已有名单期限。前后端须同版，PUT必须提交block_mode；旧页面返回400需刷新。备份后按[认证IP防护说明](AUTH_IP_BLACKLIST.md)核验真实IP、转发头及源站限制，观察共享出口失败分布后选择模式并开启。新SHA CI/安全/GHCR待验，下方官网SHA不含本功能；本地证据与回退边界见HANDOFF和[审查记录](AUTH_IP_BLACKLIST_REVIEW.md)，生产由用户部署。

> 2026-10-10最终发布：包号职责收敛及安全补修已推送origin/main，功能提交`026d9ab2e78e244858e9a57fafe3297e193434ac`。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165503)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165599)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165523)全部成功，含Go unit（独立PostgreSQL）、integration、lint、前端及脚本检查。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-026d9ab`（linux/amd64），摘要`sha256:93d974774096864fcf60a67050586c01fbf7bab27f6f6eb7f349e0e6aa96adeb`。Go1.27.2、x/net0.60.0与x/tools0.50.0已生效，govulncheck零可达漏洞（模块中2项未调用漏洞，不宣称全部依赖零漏洞）；golangci-lint2.14.0通过，旧HTTP/2接口仅有五文件/明确符号限定的弃用兼容规则，迁移另列待办。证据output/security-release-20261010与output/dedicated-release-20261010。用户备份后执行`./update.sh sha-026d9ab`；生产未代部署，真实请求/到期/跨组计费待用户验收。本条覆盖下方历史未发布或失败候选记录，后续纯文档提交不产生新镜像。


> 2026-10-10包号职责收敛仅源码完成，尚无本轮已核验镜像。发布后同版更新前后端，旧billing-policy接口移除、旧策略值不再生效；数据库不删表不清零，回退旧镜像会恢复旧限额。上线验收见[包号当前规则](DEDICATED_ENTITLEMENT_ONLY.md)，不沿用旧SHA的CI结论。

## 2026-10-08 收款限制与人工开通（最新可更新版本）

功能`575467a5f4d194ebd01cf771edacef9f0075d395`已推送；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962687)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962637)、[镜像](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962677)全部成功。

固定镜像`ghcr.io/engineemomo/sub2apicust:sha-575467a`（linux/amd64），摘要`sha256:5539364a4f724569a4f0ef22bf12db74357fce5b6e3669327be619137d5e408e`。备份数据库、配置和app_data后，在部署目录执行：

```bash
./update.sh sha-575467a
```

新增247迁移及共享收款限额、快捷充值配置、套餐人工开通模式。限额默认关闭；先按[收款说明](PAYMENT_COLLECTION_LIMITS.md)核对旧付款链接及线下到账，再配置单笔50（严格少于50设49.99）／每日1000并启用。支付宝交易不存在不释放预占，人工续费在关闭在线支付时仍可联系管理员。真实商户和生产交互待用户验收，未代部署。下方待发布提示为历史，不影响此终态；纯文档提交不生成新镜像。

2026-10-08审查补修未发布：支付宝交易不存在或SDK业务失败继续保留额度，人工续费使用不受在线支付开关限制的公开套餐联系入口。待新版本发布后验收，不将既有镜像当作包含修复。

2026-10-08最新源码：收款限制与人工开通已完成但未发布，新增247迁移及默认关闭的50／1000人民币共享限额、可配置快捷充值／自定义开关、套餐contact_admin与联系说明，人工发放继续扣库存。迁移、接缝、证据与待办以HANDOFF最新状态、CUSTOMIZATIONS及deploy/PAYMENT_COLLECTION_LIMITS.md为准；旧sha-70e66cf不含此功能，不能沿用旧CI结论。发布前须核对历史付款链接及线下到账，前后端同版，生产由用户部署。

## 2026-10-08 库存与帮助改进（最新可更新版本）

最终功能`70e66cf7305e0cb4610671b5120e3612f3478fb6`已推送；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127131)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127128)、[镜像](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127116)均成功。

固定镜像`ghcr.io/engineemomo/sub2apicust:sha-70e66cf`（linux/amd64），摘要`sha256:bf94b9dcfca53dd394660904566bef67adc979f8abc721d607d54fc77964440e`。先备份数据库、配置和app_data，再在部署目录执行：

```bash
./update.sh sha-70e66cf
```

包含套餐库存与人工分配扣减、帮助图标区分、六节教程分类、17题FAQ，以及“本站支持生图，需联系管理员开通权限”说明。245／246迁移会建立库存与归属账本，旧套餐默认不限量；设限额前核对历史占用，无法归属的人工记录按[库存说明](SUBSCRIPTION_PLAN_STOCK.md)处理。升级后验管理员分配／购买／售罄／取消释放及帮助页面；真实支付与客户端接入由用户验证。未代用户部署。下方未发布提示为历史记录，纯文档提交不生成新镜像。

2026-10-08帮助页改进尚未发布：更新后应看到常见问题用问号、接入教程用书本；教程六类目录与FAQ分类筛选在控制台／公开页／游客预览可用。本地56项、类型／lint／构建和四种公开夹具布局通过，不代表真实客户端接入成功。当前无本轮新镜像标签，未代用户部署。

2026-10-08同轮库存补充包含246人工库存归属迁移：管理员分配、默认赠送／兑换、延长期限同样受库存约束，我的订阅展示库存并禁用售罄续订。上线前核对无套餐／多套餐分组及历史混合来源，不能假定旧人工分配全部自动还原；操作见[库存说明](SUBSCRIPTION_PLAN_STOCK.md)。本轮尚无发布标签，仍未部署。

2026-10-08套餐库存仅源码完成，当前没有本轮可用镜像标签。待授权发布并核验对应SHA CI／安全／GHCR后，备份数据库再更新同版前后端；245迁移保留旧套餐不限量并回填现存已售／预占。库存总限额包括历史占用，补货方法及晚付人工处理见[SUBSCRIPTION_PLAN_STOCK.md](SUBSCRIPTION_PLAN_STOCK.md)。不把sha-07f0ee6当作已有库存功能，不代用户部署。

## 2026-10-07 版本权限与依赖安全修复（最新可更新版本）

功能`07f0ee69e0b9318edbdea0694e2d5ecd0ae2e24a`已推送；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687152)（含Go unit／integration）、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687062)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687015)全部成功。包含版本仅管理员展示、公开版本值清空、依赖17告警清零；前端本地2928项通过。

实际镜像`ghcr.io/engineemomo/sub2apicust:sha-07f0ee6`（linux/amd64），摘要`sha256:638ef47fc3b3d7f185849f90ef0bd8cd48b38732032c20f6176962b87e14821e`，成功manifest推送与containerimage.digest已核验。

用户先备份数据库、部署配置和app_data并记录旧镜像，再在部署目录执行：

```bash
./update.sh sha-07f0ee6
```

更新后验证健康检查、管理员版本可见、普通用户与游客无徽标，公开`/api/v1/settings/public`及首页注入的version为空，原有图标／公告展示正常。本版包含此前244迁移与支付宝官方直码，仍需按下节新建订单真机验收。未代用户生产部署；下方本轮未发布提示为历史记录，纯文档skip-ci提交不替代本功能镜像。

> [CUSTOM] 2026-10-07版本权限修改尚未发布，下方已发布镜像不含该改动。发布后验证管理员版本可见、普通用户与游客无徽标，公开settings/public及首页注入version为空。详见[安全报告](SECURITY_REVIEW_20261007.md)与HANDOFF最新状态。

> 同轮依赖安全修复也尚未发布：移除未使用图标依赖链、DOMPurify3.4.16，生产依赖复扫全0、前端2928项及类型／构建通过。需以新SHA完成CI／安全／镜像门禁后再更新，不将旧sha-80c30c8视为包含本次修复。

## 2026-10-07 开通周期与支付宝直码（最新可更新版本）

功能提交 `80c30c8250d7de59da55b964251f1c437bcf6ddd` 已推origin/main，包含订阅开通周期、支付宝官方直码及SheetJS安全依赖补修。同SHA的 [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276507)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276518)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276525) 全部completed/success，CI含unit／integration、Go静态检查、前端类型／关键回归、脚本及发布辅助检查。

成功构建日志确认镜像 `ghcr.io/engineemomo/sub2apicust:sha-80c30c8`（linux/amd64），摘要 `sha256:f3ed8f3fa622bc3ed629472c5c2f0d6e343723646ec030a5d50ee99ca59989f3`。前一候选ba405f1a8因SheetJS例外到期导致安全失败，不作为本轮推荐版本；现已升级官方0.20.3并移除两条例外。

先备份Postgres、部署配置及app_data并记录旧镜像，在部署目录执行：

```bash
./update.sh sha-80c30c8
```

本版新增244迁移：周／月按开通时间的7／30天周期计算，旧未对齐窗口回填当前周期，保留已用额度与到期时间；无法按新边界精确拆分的旧聚合用量保守结转到下个周期，不在迁移时赠送新额度。

支付宝必须重新创建支付订单后扫码，新码直接编码官方签名WAP地址；恢复旧订单仍显示旧码。更新后核对实际运行镜像、订阅重置时间、扫码付款页的商户／金额、小额支付后电脑自动成功与仅入账一次。真实App直达和商户权限仍待用户验收，CI成功不能替代真机验证；生产由用户部署。


## 2026-10-07 v0.2.13定制合并版（历史版本）

功能提交 `4bc5abbb53544fe89f9fc3bc001f5a69b7ca07e5` 已推origin/main；同SHA的 [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492464883)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492464870)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492465165) 全部成功。固定镜像 `ghcr.io/engineemomo/sub2apicust:sha-4bc5abb`（linux/amd64），摘要 `sha256:df08bf887996ab87efe2a5f8a4c9958a925723ff9b0fd4764b30e4512dadde52`，由成功构建的manifest推送及digest输出确认；不要用中间失败候选2c282e409或a8baeb515作为更新目标。

本版含上游v0.2.13、订阅到期与重置时间提示、支付宝桌面WAP扫码与安全修复、更新脚本镜像保留，以及已认证包号跳过余额预占的兼容修复。Vue／source-map-js新增高危依赖已更新，前端安全仍按原SheetJS例外通过。更新脚本的BSD sed与Bash 3兼容性已修，macOS CI及Linux22场景通过。

先备份Postgres、部署配置、app_data，记录旧镜像；在实际部署目录执行：

```bash
./update.sh sha-4bc5abb
```

如需同步本轮更新脚本，可从仓库该功能提交的 `deploy/update.sh` 获取并替换部署目录中的脚本，保留原脚本备份。脚本代码推送不等于服务器脚本自动更新。

两项上游241迁移与本仓241/242/243按完整文件名并存，勿重命名已应用文件。升级前未使用的重置密码链接需重新申请；Key创建默认每用户200个有效Key／每小时60次，余额并发预占默认启用，有效包号仍不预占客户余额。更新后核验health、运行镜像版本、订阅时间提示、包号零余额准入，并按ALIPAY_DESKTOP_WAP.md验收实际扫码支付；本地／CI成功不等于真实业务已验收。生产由用户部署。

## 2026-10-06 v0.2.13本地合并候选（历史记录）

已合并稳定上游源码及本轮支付宝／订阅／更新脚本定制，本地检查结果见HANDOFF。尚未推送、没有本次提交对应的CI或新镜像，下面已发布版本记录保持有效。发布前须核验本次SHA；保留包号预付跳过余额预占、支付严格金额与赠金／折扣的实付和到账区分。两项上游241迁移与本仓241/242/243并存，勿重命名已应用文件；升级前备份数据库及配置。上游重置密码令牌改哈希，升级前未使用链接需要重新申请；Key创建默认每用户200个有效Key／每小时60次，余额并发预占默认启用。生产仍由用户部署。

## 2026-10-02 配方打开原图修复（最新可更新版本）

功能4bf1436fb8ecae0c0780c3b741e26cefba9dd170已推origin/main，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283643)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283348)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283175)全部success；Go unit／integration、lint0、前端类型／lint和372项回归通过，镜像embed编译完成。实际固定镜像ghcr.io/engineemomo/sub2apicust:sha-4bf1436（linux/amd64），摘要sha256:038e95a020dc71a5615c02cec490a9a92cf8172242686977051b9686cdeae87e，经manifest推送与containerimage.digest核实。

生成后的“打开图片”改页内原图弹窗，带关闭／Esc和手机完整预览；Base64提供下载入口，远程图可长按或右键保存。仅自有配方静态资源与交接，不新增迁移；含上一版包号免扣和迁移243，旧acd73d7不含本次图片修复。61项配方＋3项打包、类型／生产构建与独立／同源iframe夹具通过；下载落地、真实线上根因与结果仍待验收，内嵌测试的一条MutationObserver错误来源未知。更新脚本专项未纳入，代理未操作8080或生产。

用户先备份Postgres、部署配置和app_data（含studio-submissions），记录旧镜像，再执行：

```bash
cd /sub2api-deploy
./update.sh sha-4bf1436
```

更新后刷新配方，用真实已生成图片检查打开图片1、多图序号、手机关闭与原图保存；无需为了验证打开功能重新请求模型。包号策略与计费验收仍按下节执行。

## 2026-10-01 包号免扣与共享限制（最新可更新版本）

功能acd73d716daa6a8e728a36fca5dce5fd523ae3f6已推origin/main，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854705)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854536)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854627)全部成功。Go unit／integration、PG、lint0 issues及前端类型／372项回归通过，镜像embed编译成功；govulncheck未发现漏洞，前端审计按既有例外通过。本次包括有效包号实扣0／余额0可用、独立参考计量、后台使用限制，以及此前同批成员多专属组、移除成员、Key恢复与旧表单保护；公共组按原规则计费，前后端与迁移243同版。293c091不包含此次包号改动。

实际固定镜像ghcr.io/engineemomo/sub2apicust:sha-acd73d7，linux/amd64，摘要sha256:b41acdce7446e1b7bca618e16e6f6dbdab5b13fd30732760247063deb576bfa7，经成功任务manifest推送及containerimage.digest核实。先完成下面的备份并记录旧镜像，再由用户执行：

```bash
cd /sub2api-deploy
./update.sh sha-acd73d7
```

更新前备份Postgres、部署配置与app_data（含studio-submissions），记录当前镜像。确认新SHA门禁全部成功后在服务器用固定sha标签更新。迁移启动时执行；更新后在包号管理配置使用限制，默认2并发／30次每UTC固定分钟／日上限0不限／2MiB／生图关闭，不是Token或美元预算。验证0余额专属组实扣0、普通组原计费、到期撤销拒新请求、共享限额和实际客户端。需要WS会话更新或实时音频的客户端须先验证兼容性；本版包号WS仅支持生成与取消。回退旧计费版本前暂停相关包号Key／账号。

宿主更新脚本专项仍未提交，本次应用镜像不会替换服务器上的update.sh。本轮未代部署生产或调用收费模型，线上历史扣款仍需流水核实；后续纯文档skip-ci提交不产生替代镜像。下方293c091及更旧版本记录为历史。

## 2026-10-01 截图与发布门禁修复（最新可更新版本）

功能提交`293c091b3ab0c8f187c031bed1955bb843939d40`已推origin/main：[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744777)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744680)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744656)全部成功。Go lint 0 issues、unit／integration与镜像embed编译成功；govulncheck无漏洞，前端审计按既有例外通过（两项SheetJS high例外仍存在，未扩例外）。本地实际提交快照644项、类型／源码lint／Vite构建通过。

固定镜像`ghcr.io/engineemomo/sub2apicust:sha-293c091`，linux/amd64，摘要`sha256:c517187b769f8b49f1fb29b1667277b745bb42deb5be05eabbf3f39e54274a06`，由该SHA成功任务的manifest推送及containerimage.digest核实。包含滚动吸顶栏、SynaRoute选中态、首页文案、47图目录及空模型列表诊断，还有Axios1.20.0、x/image0.45.0与投稿资源清理检查；视频／分镜仍待开放。下方765e5a9和d84d60e的暂缓状态是历史，更新应使用本固定标签；收尾文档skip-ci不生成替代应用镜像。

先备份Postgres、部署配置与app_data（含studio-submissions），记录当前镜像作为回退版本，再在服务器执行：

```bash
cd /sub2api-deploy
./update.sh sha-293c091
```

更新后确认健康检查，登录验收滚动栏、SynaRoute、五类新图与保留的单张时尚肖像、配方密钥／生图模型选择、投稿待审和管理员审核。线上空模型框没有取得真实账号的Key／分组／models响应，不能据模拟选择测试宣称现场根因或生图效果已确认；若仍无选项，按新版具体提示检查配置并提供脱敏状态码。包号与新版更新脚本及CI接线不包含在本次应用提交；宿主update.sh不会随应用镜像自动替换。本轮未代部署生产或调用收费模型。

## 2026-10-01 配方与工坊一体化（已推送，暂缓部署）

功能提交765e5a9516eaecf23d94202b4afc48f9bf723339已推送，[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36825037358)成功，固定标签sha-765e5a9、摘要sha256:341886529ae61b817a24da3a2a40dbb7062aabbc2148ecc9b7956153cb422e9d。但是[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36825037421)发现x/image的WebP／VP8L可达漏洞与Axios七个high，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36825037384)的Go lint也有八处资源清理错误返回未检查。**镜像存在不等于发布门禁通过：暂不建议运行update.sh或拉latest更新公开服务器。** 修复后需新SHA、新安全／CI及镜像成功，再由用户备份更新；本轮未操作生产。

本机8080已使用同一应用镜像提供`/recipes/`与`/studio/`，无需两个额外服务或产品地址变量。原数据库／Redis／数据卷／业务计数及全部设置摘要保持，镜像／HTTP／实际页面证据见HANDOFF最新终态；迁移和可选独立托管详见[FAMILY_INTEGRATED.md](FAMILY_INTEGRATED.md)。

用户已授权本轮限定提交推送，实际暂存源码快照333项测试、类型／定向lint／Vite构建通过；包含最新主题／同页配置、38图提示词、投稿人工审核及视频／分镜待开放，包号和新版更新脚本不包含在本次应用发布中。**新功能SHA与GHCR发布状态以HANDOFF本轮最终记录为准，生产未操作。** 下方`sha-25071d4`和更旧标签不包含一体化；须核对新SHA的CI和镜像构建成功，再由用户备份Postgres、部署配置及app_data（含studio-submissions），更新固定sha标签。不能直接拉旧标签或latest认为已更新；真实账号配置授权／投稿审核及收费调用仍需用户测试，静态与模拟账号回归不等于真实模型效果。

## 2026-09-30 家族视觉与下载入口更新（最新）

功能提交 `25071d4d547c69d6c933bbcb0c276585b23b4b68` 已推送 origin/main：工坊三张等宽图片与鼠标／点按／键盘置顶、最新深发人像、家族首页独立皇冠图标、SynaRoute 官网与下载链接，以及发布地址配置说明。只包含 16 个家族相关文件，不包含工作区包号专项，未新增跨站 SSO 或一体化产品托管。实际暂存快照 72 项／8 文件、类型／相关 lint／Vite 构建通过；真实生产登录与模型授权仍需用户验收。

**记录时发布尚未完成**：[镜像构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764218)在 Build and push、[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764231)的 Go Unit 仍运行；[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764322)以及 CI 的前端／Go lint／shell／release-helpers 已成功，不标全部通过，不把旧提交的成功结果沿用到本轮。待三项成功并核对镜像任务日志中实际标签／摘要，再备份数据库和部署配置、记录旧标签后执行：

```bash
cd /sub2api-deploy
./update.sh sha-25071d4
```

上面是本轮功能提交的预期标签，不代表构建中已能拉取；后续交接文档 `[skip ci]` 提交不生成替代镜像。三项 `MAGIC_RECIPES_URL`／`MAGIC_RECIPES_ORIGIN`／`MAGIC_STUDIO_URL` 仓库变量名称只读查询仍为空，配方／工坊入口保持待配置，需按 `frontend/src/custom/FAMILY_PORTAL.md` 设置实际已托管地址并重建。上线后验证 `/health`、三图及交互、皇冠与仪表盘图标区分、SynaRoute 公开链接、真实登录后的家族首页／API 控制台往返；代理不代操作生产。

## 2026-09-30 家族产品与接入教程版本

功能提交 `321fe9943b700384b10c2fc1368ed24c6b90eb97` 已推送origin/main：保留品牌首屏，新增互动产品展台／控制台产品坞、配方模型设置与显式登录配置、工坊独立源码及接入教程。未混入工作区待确认的包号专项修复；本机8080此前整工作区镜像不等于本次发布范围。

**发布门槛**：[本轮镜像构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36712632118)已成功，实际标签 `ghcr.io/engineemomo/sub2apicust:sha-321fe99`、摘要 `sha256:c8dd5b34edacf29eec9e30f53f35f5de8c5898706b86715b81d70b8a73db4147`，从成功任务日志核验。[本轮CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36712632123)的Go Unit、frontend、Go lint、shell、release-helpers及安全扫描成功，集成仍运行；先确认集成成功，不把镜像发布等同全部CI通过。实际暂存源码已通过210项原站、54项配方及21项工坊测试与类型／lint／Vite构建。

集成成功后，备份数据库与部署配置，记录当前标签，在服务器执行（后续仅文档的skip-ci提交不生成新镜像，仍使用此功能标签）：

```bash
cd /sub2api-deploy
./update.sh sha-321fe99
```

独立配方／工坊不随原站镜像自动托管。仓库 `MAGIC_RECIPES_URL`、`MAGIC_RECIPES_ORIGIN`、`MAGIC_STUDIO_URL` 尚未配置（只读API核对变量名称），本次生产入口将显示待配置并禁用；需先单独托管、确认HTTPS或同源路径、设置公开构建变量后重建。独立配方还需精确API CORS，不放松认证或来源限制，不把本机4178／4179地址用于生产。具体见 `frontend/src/custom/FAMILY_PORTAL.md`。

更新后核查 `/health`、`/home`、`/family`、`/api`、`/guide`及真实登录后的产品切换。真实Key授权／计费及用户视觉批准仍待，不自动调用收费模型；后台自定义 `home_content` 仍可能覆盖默认首页。代理未操作生产。

## 2026-09-30 多人包号与控制台版本

代码 `43b715a15f599f5ae2125bc923849d57e4b3522c` 已推送；GHCR及安全扫描成功，镜像 `ghcr.io/engineemomo/sub2apicust:sha-43b715a` 已生成（摘要 `sha256:dc36b117e524262faddf0c13f0dad5c9ff510e385e69ddc2860145a7ea6566af`）。

**发布门槛**：记录时[本轮CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36657101544)的集成测试仍在运行，Go单测、Go lint、前端及其他任务通过；请先确认该CI全部成功，不把镜像成功当作全量测试通过。

通过后先备份数据库和部署配置，再在服务器执行：

```bash
cd /sub2api-deploy
./update.sh sha-43b715a
```

本次新增迁移242（多成员与删除隔离），保持一账号一专属分组的独占规则。上线后检查登录隐藏游客入口、FAQ右侧正文、包号名称／多人／改绑／撤销后删除及任务费用显示。不要直接回滚旧单用户镜像；详细规则与回退边界见[包号排查说明](DEDICATED_TROUBLESHOOTING.md)。生产尚未由代理操作。

## 2026-09-28 游客功能已验证版本

代码提交 `fa45b2023c01e5c84084dcbcb32d5a7c5bbe213a` 的 CI（含Go单测／集成）、安全扫描及GHCR构建全部通过，发布镜像 `ghcr.io/engineemomo/sub2apicust:sha-fa45b20`。本次包括公开官网／真实套餐／FAQ、游客控制台及宽屏手机适配；不开放个人数据接口。

服务器由用户操作：先备份数据库与部署配置，并记录当前镜像标签，再执行：

```bash
cd /sub2api-deploy
./update.sh sha-fa45b20
```

更新后检查 `/health`、首页游客入口、`/preview`、`/plans`、`/faq` 和登录回跳；如配置了自定义 `home_content`，其优先级仍保留。未执行生产更新或真实支付验收。后续仅文档提交不生成新镜像，不能替换此处已验证标签。

> [CUSTOM] **Windows本机源码构建**另见 [LOCAL_DOCKER_RUNBOOK.md](LOCAL_DOCKER_RUNBOOK.md)：已记录Docker Hub授权超时与代理预检。本文下方仍是服务器镜像更新流程；不要把本机127.0.0.1:7897代理配置带到生产。

> 速查卡：以后升级 / 回滚照这份敲即可。完整部署、首次平移、数据备份/恢复见 [DEPLOY_CUSTOM.md](DEPLOY_CUSTOM.md)。
>
> 你的环境（已确认）：部署目录 `/sub2api-deploy`，compose 用 `docker-compose.local.yml`，
> 定制镜像 `ghcr.io/engineemomo/sub2apicust`（owner 是 EngineeMoMo，中间**两个 e**），DB 用户/库均 `sub2api`。

---

## 更新分两步

### 第一步：出新镜像（代码侧，一般由维护者 / AI 协助）
上游出新版、或要改定制（换色 / 加功能）时 → 在 fork 仓库合并或改代码 → `git push` → CI 自动构建新镜像推到 GHCR。
（需要 Go 编译校验，本机若无 Go 就靠 CI。）构建完成后你会拿到新的镜像标签 `sha-xxxxxxx`。

### 第二步：服务器部署（你操作）
进部署目录：
```bash
cd /sub2api-deploy
```
按 override 里 image 钉的是 `latest` 还是 `sha-xxxx`，选一种模式：

**模式 A ｜ 用 latest（省心，日常一条命令）**
override 的 image 是 `...:latest` 时：
```bash
./update.sh
```
每次自动拉最新定制镜像并重建。

**模式 B ｜ 钉 sha（可控，当前就是这个）**
override 钉了固定 `sha-xxxx` 时，不带参数的 `./update.sh` 只会重拉同一版、**不升级**。
升级到新版本要带新 sha（维护者会给你）：
```bash
./update.sh sha-<新提交>
```
脚本会自动切 override 标签 + 备份原文件，再拉取重建。

**从模式 B 切到模式 A**（以后只敲 `./update.sh`）：把 override 重写为 latest：
```bash
cat > /sub2api-deploy/docker-compose.override.yml <<'EOF'
services:
  sub2api:
    image: ghcr.io/engineemomo/sub2apicust:latest
    environment:
      - DISABLE_ONLINE_UPDATE=true
EOF
```

`update.sh` 每次都会：记录更新前镜像 → 拉镜像 → 重建容器 → **启动时自动跑数据库迁移** → 等 `/health` 就绪 → 定向清理本项目旧镜像。

### 旧镜像保留规则（2026-10-01）

- 仅处理 `ghcr.io/engineemomo/sub2apicust` 的 `sha-*`／`latest` 标签，以及带本仓来源标签的悬空镜像；保留当前镜像、实际更新前镜像和所有其他运行／停止容器引用的镜像。手工 `v*` 等标签保持，不清数据库、Redis、日志、备份、数据卷或其他项目镜像。
- 更新成功后，更新前镜像增加本地 `rollback-previous` 标签，支持只拉 `latest` 时保留旧版；重复更新同版时不覆盖已有回退镜像。首次使用新版脚本却没有发生版本切换时，保留按镜像创建时间排序、比当前更早的最近一个 `sha-*` 镜像作为回退候选，这不等于已验证该候选曾运行成功。
- `/health` 超时返回非零，跳过全部清理；拉取／重建失败同样不清理。检查失败时保留相关镜像并输出原因，删除不使用强制参数。`update-before` 是拉取前的保护标签，更新成功时移除；中断／失败时可能保留用于人工排查。
- 部署目录 `.sub2api-update.lock` 防止同目录并发更新，正常退出自动移除。被强制终止时可能遗留；确认没有更新进程后，使用 `rmdir /sub2api-deploy/.sub2api-update.lock` 移除空锁目录再重试。
- **服务器必须单独替换 `/sub2api-deploy/update.sh` 为本仓新版，并执行 `chmod +x /sub2api-deploy/update.sh`。仅拉取应用镜像不会更新宿主机脚本。** `rollback-previous`／`update-before` 为服务器本地标签，不是GHCR发布标签；指定版回退继续使用原 `sha-*` 标签，不把本地标签传给会执行pull的更新命令。
- 回归：`bash -n deploy/update.sh`、`bash deploy/tests/update-test.sh`。后者仅使用假Docker，不连接daemon；当前22个升级、latest（含无sha别名）、同版、回滚、仓库切换、容器占用、检查／删除失败及并发锁场景通过。真实服务器回收量需升级后用 `df -h` 和 `docker image ls ghcr.io/engineemomo/sub2apicust` 核验。

---

## 回滚
切回上一个正常的版本：
```bash
cd /sub2api-deploy && ./update.sh sha-<上一个正常的提交>
```
迁移是前向增量的，回滚镜像通常安全（旧代码忽略新列）；只有怀疑迁移本身出问题才用数据库备份恢复，见 DEPLOY_CUSTOM.md 第七节。

---

## 每次升级后验证
```bash
docker compose -f docker-compose.local.yml -f docker-compose.override.yml exec -T sub2api /app/sub2api --version
```
- 版本号是你要的 `0.2.8-custom.xxx`；
- 浏览器：主题正常、老数据都在、后台**无「立即更新」按钮**；
- 设过 logo / 站点名的，检查显示正常。

---

## 常见坑
- **`pull` 报 not found**：多半是镜像 owner 拼错——正确是 `engineemomo`（EngineeMoMo，中间**两个 e**），别写成 `enginemomo`。
- **`./update.sh` 跑了但版本没变**：override 钉了固定 sha（模式 B），要 `./update.sh sha-<新>` 或切 latest（模式 A）。
- **`pull` 报 unauthorized / denied**：GHCR 登录过期 → `docker login ghcr.io -u EngineeMoMo`（用有 `read:packages` 的 PAT；**别用 `echo <token> |` 的写法**，token 会进 shell history）。
- **升级前想备份数据库**：
```bash
docker compose -f docker-compose.local.yml exec -T postgres pg_dump -U sub2api sub2api | gzip > backup-$(date +%F).sql.gz
```
