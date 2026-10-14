# 项目规则 — sub2api 定制 fork（本仓专属；通用规则见 ~/.claude/CLAUDE.md）

2026-10-14 上游0.2.16发布终态：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。

2026-10-14 同步上游0.2.16：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

2026-10-14 网段防护发布终态：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

2026-10-14 网段防护本轮：用户要求通配符及自动按网段封禁。新增254，手动210.230.206.*归一为/24；自动不同IP共同累计明确失败，默认IPv4 /24、IPv6 /64，可改单IP、前缀及期限。成功不计数，模型/管理接口不挂名单。8080已备份更新，三服务healthy，开关false、version4、名单0、原数据与设置摘要/卷保持；真实PG/Redis、全量unit、embed、前端16项及44条HTTP断言通过。共享网段正常用户也可能受限，不能承诺零误拦。正式SHA/CI/GHCR状态以HANDOFF最新记录为准，其他会话改动排除，生产未代部署。

2026-10-14 手动封禁发布终态：065eee935已推origin/main，同SHA CI（unit/integration/PG18/Redis8/lint0/前端502项含IP管理14项）、安全/GHCR成功；固定镜像sha-065eee9，摘要8da3604012c2a2c121dd47a7fa22542c785b5c77c60335177498a40058807d38。8080已备份更新至e9fd3b9c7ea2，三服务healthy，原业务计数/防护配置/挂载保留；253生效。永久/临时手动添加、原因、来源、重复409和管理员解封已实测，只保护认证。独立测试已清理，其他会话源码保留且未发布；生产由用户备份后`./update.sh sha-065eee9`，证据及待办见HANDOFF最新终态。

2026-10-14 管理员手动封禁：新增253迁移、POST entries及自有管理页入口，支持永久/临时、必填原因和来源；重复有效封禁409，关闭自动开关也生效，共享IP锁和状态版本。只保护现有认证范围。本轮验证/发布以HANDOFF和AUTH_IP_BLACKLIST_REVIEW最新记录为准；旧07bcef204不含手动功能，其他会话改动排除，生产由用户部署。

> 2026-10-14 IP防护已发布：07bcef204d09349e0574b51a5b8cfdd6266288a3已推origin/main，同SHA CI（真实PG18／Redis8、unit／integration、lint0、前端497项）、安全和GHCR全部成功。固定镜像sha-07bcef2（linux/amd64），摘要sha256:81bcf46cd8e1008f058c296f09d32af059d0e7fadefa66cf355b522885d5bc13，registry标签／摘要已验真。用户备份后执行`./update.sh sha-07bcef2`；迁移250/251/252、默认关闭、成功不计数，只保护认证，永久／临时可配置。模型和管理接口不受名单影响。未带其他会话安全补修／248/249／依赖升级／智力效率，生产未代部署。具体取证与待办见HANDOFF及deploy/AUTH_IP_BLACKLIST_REVIEW.md。本条优先于下方未发布历史。

> 2026-10-13 23:31最新：用户允许永久封禁、联系管理员解除，或配置期限。IP防护新增252迁移和block_mode=timed/permanent；永久记录无到期时间，规则修改不改变已有期限。保留成功／验证码／业务校验／429／系统故障不计数；默认10次失败/120分钟、timed15分钟，自动拉黑关闭。8080运行7f423b3890c9、三服务healthy，原数据／281项设置／卷保持。13项后端专项／全量unit／Go embed、55项前端及真实隔离认证、重启／Redis清空／管理员解封通过。此次授权覆盖上一轮禁止新增永久模式；不恢复全请求计数。共享IP误拦率未知，未提交、未生产部署，新SHA CI/安全/GHCR仍待。证据与回退边界见HANDOFF、deploy/AUTH_IP_BLACKLIST_REVIEW.md；宿主无Go，本轮用Docker工具链，保留其他会话未提交内容。

> 2026-10-13最终发布：官网与登录页动效、交互、首屏加载已推送origin/main，功能提交`f8928e20dac15b341fc8d657e421b6f8ed0a8d71`。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38361578948)（含真实PostgreSQL的unit/integration、lint、前端）、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38361578790)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38361578902)全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-f8928e2`（linux/amd64），摘要`sha256:cd8c1667f943c0a6706742dcc3c0e9a1c203e75ba53dd28ec528da9f920e3d6a`。这一版只有前端改动，没有后端代码和迁移；不含工作区里未提交的10-11安全修复，也不含并行“智力效率”会话的改动。用户备份后执行`./update.sh sha-f8928e2`，生产未代部署。生产JS/CSS仍无gzip，需用户在Nginx补gzip_types。本条覆盖下方10-10的更新命令，详见HANDOFF第四、五节。

> 2026-10-10最终发布：包号职责收敛及安全补修已推送origin/main，功能提交`026d9ab2e78e244858e9a57fafe3297e193434ac`。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165503)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165599)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165523)全部成功，含Go unit（独立PostgreSQL）、integration、lint、前端及脚本检查。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-026d9ab`（linux/amd64），摘要`sha256:93d974774096864fcf60a67050586c01fbf7bab27f6f6eb7f349e0e6aa96adeb`。Go1.27.2、x/net0.60.0与x/tools0.50.0已生效，govulncheck零可达漏洞（模块中2项未调用漏洞，不宣称全部依赖零漏洞）；golangci-lint2.14.0通过，旧HTTP/2接口仅有五文件/明确符号限定的弃用兼容规则，迁移另列待办。证据output/security-release-20261010与output/dedicated-release-20261010。用户备份后执行`./update.sh sha-026d9ab`；生产未代部署，真实请求/到期/跨组计费待用户验收。本条覆盖下方历史未发布或失败候选记录，后续纯文档提交不产生新镜像。


2026-10-10用户最新要求：包号只负责账号归属、有效期和免扣，通用请求限制交sub2api。源码已移除独立并发/RPM/日次数/体积/生图/入口/帧/成员人数上限及后台策略接口；保留归属隔离与免扣凭证，Key限额恢复原生计量。未提交推送部署，真实PG专项待验；以HANDOFF和deploy/DEDICATED_ENTITLEMENT_ONLY.md为准，不恢复历史包号独立限额方案。

2026-10-08发布终态：收款限制与人工开通及两项审查修复已推送575467a5f，同SHA CI（含unit／integration／lint）、安全、GHCR全部成功；固定镜像sha-575467a。更新及证据见HANDOFF和deploy/UPDATE_GUIDE，下方本轮待发布记录为历史；生产由用户备份后更新。

2026-10-08审查补修：支付宝关单须检查SDK业务成功码，交易不存在保留收款预占；人工续费进入公开套餐联系目录，不受在线支付开关限制。定向回归通过，未提交推送部署，详见HANDOFF最新记录。

2026-10-08最新源码：收款限制与人工开通已完成但未发布，新增247迁移及默认关闭的50／1000人民币共享限额、可配置快捷充值／自定义开关、套餐contact_admin与联系说明，人工发放继续扣库存。迁移、接缝、证据与待办以HANDOFF最新状态、CUSTOMIZATIONS及deploy/PAYMENT_COLLECTION_LIMITS.md为准；旧sha-70e66cf不含此功能，不能沿用旧CI结论。发布前须核对历史付款链接及线下到账，前后端同版，生产由用户部署。

2026-10-08最新发布终态：库存与帮助、生图需管理员开通说明已推送，最终功能70e66cf73同SHA CI／安全／GHCR全部成功，固定镜像sha-70e66cf；245／246迁移升级前须备份，生产由用户自行更新。证据output/release-stock-help-20261008，详见HANDOFF和UPDATE_GUIDE。本条覆盖以下未发布状态，纯文档skip-ci提交不替代镜像SHA。

2026-10-08帮助页最新补充：FAQ问号／教程书本图标，教程六节按场景区分，FAQ四类16题可筛选；56项、类型／lint／构建及375／1440浅深色公开夹具浏览器检查通过。未发布，库存工作树保留；接缝见CUSTOMIZATIONS，来源与验收边界见GUEST_PORTAL及HANDOFF。

2026-10-08库存补充最新状态：管理员单次／批量／续期、兑换与默认赠送已统一扣库存；246归属账本及我的订阅库存／售罄禁用完成，已付订单发放不重复扣。旧多来源历史须核对。4包421后端项、21文件216前端项、类型／lint／前后端构建通过，manual-*证据与限制见HANDOFF和库存说明；未提交推送部署。

2026-10-08套餐库存源码已完成未发布：-1默认不限量、0停售，订单事务内原子预占，未付结束释放，已付／退款／到期不补货；晚付重新取库存，不足记审计且不自动开通。184前端／236后端定向、真实PG并发及完整迁移、类型／lint／前后端构建通过，详见HANDOFF与deploy/SUBSCRIPTION_PLAN_STOCK.md。未提交推送部署，本轮已用现有便携Go编译，不沿用旧CI。

2026-10-07最新发布终态：功能07f0ee69e已推origin/main，同SHA CI（含Go unit／integration）、安全与GHCR全部成功，固定镜像sha-07f0ee6；版本仅管理员展示、公开version为空及17条依赖告警清零已发布，生产由用户备份后自行更新。详见HANDOFF与deploy/UPDATE_GUIDE；本条覆盖下方本轮未发布记录。

2026-10-07后续用户要求修复：已移除未使用@lobehub/icons依赖链、DOMPurify3.4.16及锁文件同步，生产依赖复扫所有严重度0；前端全量368文件／2928项、类型／生产构建及相关lint通过。覆盖下条17告警待处理状态，保留版本权限修复；尚未提交推送部署，详见HANDOFF及安全报告。

2026-10-07版本权限与安全抽查已本地完成未发布：VersionBadge仅管理员可见，公开JSON／HTML注入version为空；13项前端、公开设置／权限支付定向Go、类型／lint与前后端构建通过。生产依赖审计仍13中危／4低危、0高危／严重，未证实业务可利用性，详见HANDOFF及deploy/SECURITY_REVIEW_20261007.md。未推送部署，不将旧80c30c825 CI当本轮结果。

2026-10-07最新发布终态：功能80c30c8250d7de59da55b964251f1c437bcf6ddd已推origin/main，同SHA CI（unit／integration）、安全扫描与GHCR均成功，固定镜像sha-80c30c8。包含周／月按开通计时、244迁移、支付宝新订单官方直码及SheetJS0.20.3补修；更新前备份，更新后必须新建支付订单真机验收。生产由用户部署，证据与步骤见HANDOFF四／五和deploy/UPDATE_GUIDE；本条覆盖下方历史未发布状态。

2026-10-07发布终态：用户授权后已推送功能4bc5abbb5，同SHA CI（unit／integration）、安全扫描和GHCR全部成功，固定镜像sha-4bc5abb；发布与更新步骤以HANDOFF第四节和deploy/UPDATE_GUIDE最新节为准，覆盖下方历史未发布状态。生产由用户部署。

- 2026-10-06用户授权合并上游v0.2.13：48个未提交文件先备份并保存快照e2adf6005；处理支付路由冲突，保留扫码／游客目录和上游匿名查询限流，新增已认证包号跳过余额预占接缝。Linux全量unit、真实PG包号／迁移、2921项前端全量、类型／源码lint／构建和登录接口冒烟已核验，最终结果见HANDOFF。完整integration已在用户明确授权socket后补跑通过（52包），另补跑securityaudit的PG／Redis测试214项通过，剩余6项既有／外部依赖跳过已记录；未推送部署，不把本地检查等同新SHA CI或生产验收。

- 2026-10-06支付宝严格安全复查已补强：签名字段／商户／总额与查询订单绑定，固定配置公钥、精确分金额，扫码HTTPS／可信站点、歧义参数拒绝及限流前no-store。111顶层＋103子项、真实仓储16并发重放仅入账一次连续三轮、Host fuzz与embed通过；静态检查最终结果见output/alipay-security-review-20261006/verification.json，报告deploy/ALIPAY_SECURITY_REVIEW.md。未提交推送部署；真实PG并发、新SHA CI及App实付待验，不承诺绝对安全，保留更新脚本和订阅UI专项。

- 2026-10-06我的订阅时间／样式已完成未发布：原时间提示存在，现custom卡片读取原progress接口的服务端resets_at，15秒本地显示更新、实际到期日期及过期／暂停／最后不完整周期提示；续费仍原分组路由，样式只theme.css、平台语义色保持。21新回归／399关键项、类型／相关lint／Vite和1440／900／390深浅浏览器模拟夹具通过。未改后端／业务库／配置、未提交推送／重建8080／生产部署；支付宝及更新脚本专项保留，真实账号仍待验。接缝见CUSTOMIZATIONS与custom/UPGRADE，结果见HANDOFF四／五。

- 2026-10-06电脑端支付宝 WAP 扫码源码完成，尚未提交／推送／部署。默认关闭开关 `payment_alipay_desktop_wap_qrcode`，仅电脑官方支付宝使用 WAP＋本站随机短 QR，免登录 302，分钟期限／状态／官方目标校验；手机与其他通道原流程保持。前端378关键项、支付195顶层＋87子项、API契约、类型／lint／Vite／Go embed通过，Windows全量两项路径／时间精度失败未标全绿，实际新SHA Linux CI及真机实付待验。使用output内已核验便携Go，覆盖下方本轮“无Go所以未编译”历史判断，宿主未安装工具链；更新脚本专项字节保持。接缝与启用验收见CUSTOMIZATIONS、HANDOFF四／五及deploy/ALIPAY_DESKTOP_WAP.md，不能把本地签名或302当真实扫码成功。

- 2026-10-02配方图片打开修复已发布：功能4bf1436fb已推origin/main，同SHA CI（Go unit／integration、lint0、前端372项／类型／lint）、安全及GHCR全成功，固定镜像sha-4bf1436／摘要038e95a020dc71a5615c02cec490a9a92cf8172242686977051b9686cdeae87e，证据见HANDOFF终态。旧data新窗口href改dialog页内原图，Base64下载入口、远程保存提示、Esc／焦点与失效清理；本机61配方＋3打包、类型／构建及独立／生产资源iframe夹具通过。下载落地和线上根因未知，内嵌夹具有一条未知来源MutationObserver错误，不能说控制台全空。无CSP、认证、包号、8080或生产变更，更新脚本专项保持未提交，生产由用户备份后更新。

- 2026-10-01用户授权包号代码已提交推送：功能acd73d716，65文件，包含免扣／共享策略及既有多人多组、成员移除、Key恢复、旧表单保护。该SHA的CI（Go unit／integration、lint、前端372项及类型）、安全和GHCR全部成功，实际镜像sha-acd73d7／摘要b41acdce7446e1b7bca618e16e6f6dbdab5b13fd30732760247063deb576bfa7；完整证据见HANDOFF最新终态。更新脚本专项仍未提交，不沿用293c091，不代部署生产；用户备份后更新并验收真实账号。

- 2026-10-01用户已授权开发包号免扣与防滥用，本地源码完成：有效包号实扣0、余额0可用、公共组原计费；同包号共享并发／RPM／配置的日请求上限，迁移243、独立参考账本及后台“使用限制”，两平台与WS同规则。不是Token／美元预留预算，不自动封IP或退款。Linux全量unit、真实PG专项、372前端关键项、类型／lint／前后端构建已验证，最新进度见HANDOFF四／五及DEDICATED_BILLING_PLAN第六节；未提交推送部署，不能沿用293c091发布结果。本轮使用仓库output内核验过的完整便携Go工具链，覆盖以下“宿主无Go所以本輪未编译”的历史判断，宿主仍未安装Go；所有新增上游接缝已登记CUSTOMIZATIONS。

- 2026-09-30家族首页／配方／工坊／教程已按用户授权推送321fe99，实际发布快照285项与类型／lint／build通过；GHCR已成功发布sha-321fe99，Go Unit通过，CI集成仍运行，不标全绿，详见HANDOFF。此前“不混入product-samples”仅针对旧包号发布。包号专项未确认合并，运行时修改仍未提交；三个MAGIC_*仓库Variables未设置、独立产品需另行托管，生产不代部署。

- 2026-09-30包号最新要求：用户明确允许同一账号关联同一批包号成员的多个标准专属分组，覆盖下方严格单组历史。源码已同时修保存／请求／撤销与移除成员，具体400诊断附失败编号；每账号仍一条包号，同组多人须在该条名单内选全。46项前端与30个真实隔离PG配置场景、Go embed通过，尚未提交推送部署，详见HANDOFF最新状态，保留并行工作。

## 交接协议（每个会话必走）
- **开工先读**：[HANDOFF.md](HANDOFF.md)（交接中枢：用户要求 / 准则 / 当前状态 / 待办）、
  [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)（定制唯一权威清单）、[SYNC.md](SYNC.md)（同步 / 构建 / 部署 runbook）。
- **收工前更新** HANDOFF.md 的「四、当前状态 / 五、待办」；任何新定制同时登记 CUSTOMIZATIONS.md。
- 交接走文件，别把状态只留在会话里。

## 本仓最关键的几条（详见 HANDOFF.md「三、用户准则」）
- 2026-09-30第四产品后续授权已做本地首版：product-samples/magic-studio，4179运行，16条资源／4个玩法、图片视频分类及复用操作；21项定向与真实浏览器复制／分享／TXT／收藏等通过。独立审查三项P2修复复核均resolved，ship仅限该修复清单。用户明确定时任务先不启用：无自动搜索／调度／模型API，不发布生产；视频仅二维运镜、模板／Skill实际效果待验，名称／托管待确认。详见HANDOFF四／五与本产品README；下方「仅规划」属历史。
- 2026-09-30本地接入教程已完成：公开 `/guide`、控制台 `/help/guide` 与游客预览／FAQ往返；80项定向、类型／lint／build通过，4175桌面深色／手机浅色和登录取消／键盘折叠／章节滚动验证通过。尚未提交推送或部署，真实客户端密钥调用待实测；两处上游接缝含router限定锚点滚动，详见HANDOFF与CUSTOMIZATIONS。
- 2026-09-30后续已按用户授权修复专项审查四项，尚未发布：请求只验当前成员、旧Key有效状态写入前验包号、编辑原始版本及撤销恢复确认、管理配置诊断。新增APIKeyService.Update的[CUSTOM]接缝，前后端需同版本，详见HANDOFF和DEDICATED_TROUBLESHOOTING。下方审查“待修”是历史状态，不能据此删保护或直接部署。
- 2026-09-30专项审查发现成员移除补丁仍有退出者改Key状态阻断全组的缺口，以及成员状态耦合、旧表单恢复撤销问题；不要直接发布。报告deploy/DEDICATED_AUDIT.md、证据output/dedicated-audit-20260930，本轮只审查未修改业务或部署。
- 2026-09-30后续未发布修复：包号移除成员在Save事务中撤销原组授权并停用该组Key，失败回滚，保留历史记录／其他组／余额；已验证但尚未提交部署，sha-43b715a不含此修复，详见HANDOFF最新状态。
- 2026-09-30发布进展：功能43b715a15已推送、镜像sha-43b715a与安全扫描成功；CI36657101544集成测试仍待终态，其余步骤已通过。不要把下方授权／未推送历史当现状，也不要把镜像成功等同全部CI通过。生产未部署。
- 2026-09-30最新收口：用户确认包号继续严格独占并授权提交推送origin/main，生产仍由用户自行发布；本轮CI／GHCR需按新提交查证，product-samples不混入提交。
- 2026-09-30用户后续授权已部署本机8080，镜像be2f26f98e8c，三服务健康，迁移242已生效；此前“仅隔离测试”属历史。证据output/console-deploy-20260930，生产未改、未提交推送，详见HANDOFF最新状态。
- 2026-09-30未发布改动：多人包号迁移242保留旧Key隔离标记，不可删标记绕过403或直接回退旧单用户版本；规则与诊断见 `deploy/DEDICATED_TROUBLESHOOTING.md`，进展以HANDOFF四／五为准。本轮Docker仅用于隔离编译与测试，不等于8080或生产已更新。
- 🔴 只说有证据的话，没证据就说「我不知道」；回答分「已核实 / 推断 / 不知道」并附取证方式。
- 全程中文。
- 定制优先新增文件；改上游文件必打 `[CUSTOM]` 注释并登记 CUSTOMIZATIONS.md；配置走 gitignore。
- **本机无 Go**：后端改动未编译，靠 CI 验证（`go build -tags embed` + `go test -tags=unit`）；提交时如实标注。
- 同步上游先 `git merge-tree` 演算再合并（rerere 已开），合并后过自检清单 + 冒烟。
- 换肤只改 `frontend/src/custom/theme.css`；上游 `teal-*` 语义分类色刻意不改。
