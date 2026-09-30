# 本机 Docker 构建网络故障与防复发清单

记录日期：2026-09-26。仅适用于本仓在当前 Windows 本机的 sub2apicust-local 环境；不是生产部署配置。执行构建、重启或部署前仍须获得对应授权。

## 1. 已核实的直接原因

构建在取得基础镜像元数据时，访问 auth.docker.io/token 连接超时；并非本次品牌代码的编译错误。失败涉及 node:24-alpine、golang:1.27.0-alpine 等，尚未进入业务编译。

- 仅保存 Docker Desktop 代理后仍失败：output/brand/docker-brand-build-proxy.log。
- 现有 Clash 监听127.0.0.1:7897；通过该代理请求 Docker Hub 授权接口返回HTTP200。
- 重启后 Docker httpproxy.log 记录镜像仓库连接走该代理。
- 给构建命令进程同时设置HTTP_PROXY、HTTPS_PROXY、NO_PROXY后，基础镜像元数据约1.4～1.5秒取得，完整构建成功：output/brand/docker-brand-build-client-proxy.log。

**推断**：Docker Desktop与构建客户端的代理路径未一致生效，部分授权请求可能直连。**尚未确定**：重启与客户端代理设置各自贡献、底层DNS／线路的唯一根因，未做独立对照实验。不得把DNS污染或“只重启就能修好”写成已证实结论。

## 2. 每次本机重建前先检查

1. 核对工作树、原数据卷和现有备份；不要覆盖未提交文档。核对docker context目标是本机，不能直接复用到远程环境。
2. 确认Clash仍在运行且端口未改。7897是本次实测值，不保证其他机器相同。
3. Docker Desktop → Settings → Resources → Proxies：HTTP／HTTPS均为http://127.0.0.1:7897。排除localhost、127.0.0.1、::1、postgres、redis、host.docker.internal。
4. 配置保存后若疑似未加载，先检查运行容器和用户工作，再经授权重启Docker；不要无条件重启。
5. **不要仅以浏览器可上网或Docker设置已保存为依据**。先测试代理授权接口，再在实际构建进程设置代理。
6. 只需启动已构建容器时，不必为了启动而重建或重新拉镜像。

## 3. 本机网络预检（PowerShell）

```powershell
$docker = "$env:LOCALAPPDATA\Programs\DockerDesktop\resources\bin\docker.exe"
& $docker context show
& $docker context inspect desktop-linux --format "{{.Endpoints.docker.Host}}"
& $docker info --format "{{.OSType}}"
curl.exe --proxy http://127.0.0.1:7897 --noproxy "" --connect-timeout 10 --max-time 25 -sS -o NUL -w "HTTP %{http_code}\n" "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/node:pull"
```

当前主机应使用desktop-linux及本地npipe端点；授权接口本次成功响应为200。输出正文丢弃，不把令牌记录进日志。若超时，先修代理可用性，不要反复跑完整构建。

## 4. 构建进程代理（仅本次会话）

以下只构建，不自动替换容器；在D:\ccfile\sub2apicust\deploy目录执行。若Docker未启动，先启动并确认引擎可用。不要用setx写全局环境变量，不把本机代理加入Dockerfile、生产.env或仓库业务代码。

```powershell
$docker = "$env:LOCALAPPDATA\Programs\DockerDesktop\resources\bin\docker.exe"
$previousProxy = @{}
foreach ($name in @("HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY")) {
    $previousProxy[$name] = [Environment]::GetEnvironmentVariable($name, "Process")
}
try {
    $env:HTTP_PROXY = "http://127.0.0.1:7897"
    $env:HTTPS_PROXY = "http://127.0.0.1:7897"
    $env:NO_PROXY = "localhost,127.0.0.1,::1,postgres,redis,host.docker.internal"
    & $docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml build sub2api
    if ($LASTEXITCODE -ne 0) { throw "构建失败：不要执行容器更新，先检查失败阶段与日志。" }
} finally {
    foreach ($name in $previousProxy.Keys) {
        [Environment]::SetEnvironmentVariable($name, $previousProxy[$name], "Process")
    }
}
```

## 5. 不能省略的部署验收

- 构建退出码为0且出现Built才算构建成功；代理接口200不等于构建成功。
- 经授权且备份完成后，使用原项目执行up -d --no-build sub2api；不删除卷，不使用down -v。
- 比较image inspect的镜像Id与container inspect的Image：两者必须一致。网页能打开不代表运行的是新镜像。
- compose ps确认应用、PostgreSQL、Redis健康；/health返回200；打开/brand和/login确认新布局、真实品牌设置及正常登录。
- 空态截图不代表带数据业务验收，不制造API调用或费用来证明部署成功。

## 6. 备份与恢复边界

- Docker设置备份：output/brand/docker-settings-before-proxy.json。恢复代理时只核对代理相关项，不整文件覆盖后续配置。
- 数据库备份：output/brand/before-brand-final-20260926.dump；早期备份before-brand-20260926.dump仍保留。
- 旧镜像标签：sub2apicust:before-brand-20260926；镜像回退和数据库回退是不同操作，不自动恢复数据库。
- 上述本地备份与配置已Git忽略，不提交凭据或数据库。
- 本次成功部署镜像前缀9c8c1dfb87bc，属于历史证据，不应当作下一次升级的固定目标。

## 7. 本轮智力效率发布记录

- **2026-09-30 已部署本机Docker（优先于本轮未部署记录）**：用户授权本机测试。原三容器处于停止状态，保留原卷启动Postgres／Redis，先完成597905字节数据库备份与pg_restore目录校验、旧镜像标签sub2apicust:before-console-20260930，再代理预检200并构建。前端类型／国际化3测／构建及Go embed成功；仅应用容器重建，运行镜像与标签一致be2f26f98e8c3dee62d765a8b78e6cd33021dc3b6f618c76085e310180b234b8，三服务healthy，health=ok。迁移242两列和三个索引确认存在；数据库／Redis容器ID及app_data卷不变，用户2／密钥0／组3／套餐1／订单0／包号0／设置280计数保持。备份逐项比较只有Claude版本2.1.283→2.1.285与Codex版本0.158.0→0.159.2两键自动同步，09:44启动日志证实，其余278键不变。HTTP入口index-Ckx_MqIV.js及三个新页面chunk验证通过，账户与管理员接口匿名401；浏览器8080官网正常打开并留给用户测试，尚未做带数据管理员交互验收。证据output/console-deploy-20260930，备份SHA256=1D900EC14FB9915B5E912F0A84778271CEB149178B7CC526E3A9F55411EE037E。未提交推送或操作生产；迁移后的回退需兼顾数据库，不能直接跑旧单用户版本。

**2026-09-28游客适配修正（最新）**：代理授权200、备份校验后build退出0，前端／Go embed成功；仅应用容器更新，镜像与容器一致ea6e3a5ecc3c021f3f128725320fcee293751c69dc795cc6ac8fb678b65e1b2d，三服务healthy、health=ok。数据库／Redis ID、用户／密钥／套餐／订单计数及设置摘要不变。8080浏览器2031px左右留白均40.609px；证据output/preview-responsive-20260928含before.json、database.dump（595135字节，pg_restore目录校验）、build.log、verification.json、responsive-checks.json、docker-wide-fixed.png；回退sub2apicust:before-preview-responsive-20260928。未Go单测、未付款／提交推送／生产部署。

**2026-09-28游客控制台预览部署（最新）**：用户授权本机部署；代理授权预检200，构建进程临时代理、首次build退出0，前端及Go embed成功；仅up --no-deps --no-build --wait应用，三服务healthy。image inspect与容器Image一致ecb3961845e8580395b31170a8a31bfa533e92710671beddf6c1249ed0594c2e；数据库／Redis容器ID及app_data卷保持，2用户／0密钥／3分组／1套餐／0订单／280设置计数和设置摘要一致。4个真实个人接口匿名401，HTTP入口index-C4UWqsIx.js包含GuestPreviewView-D-VurzQx.js；浏览器8080验证首页入口→预览、登录提示取消和真实套餐。备份output/preview-20260928-175007/database.dump已pg_restore目录验证，SHA256=6762F0F26B1ED26CA99867F9F67DCBB828730BC4DDB5D095393C4105742C0FA2；同目录含before.json、build.log、verification.json、public-catalog.json、docker-preview.png。回退sub2apicust:before-preview-20260928-175007；未Go单测、未下单扣款、未提交推送／生产部署。

**2026-09-28游客官网部署（优先于本节历史状态）**：代理授权预检200，实际compose build进程设置HTTP_PROXY／HTTPS_PROXY／NO_PROXY，首次构建退出0，含前端类型检查／国际化3测及Go embed编译；仅应用执行up --no-deps --no-build --wait。运行镜像与容器一致`e7a582285cbdaeb78412a0414ac4280322c1654c05614fec869cced2c2e13b11`，三服务healthy，health=ok；数据库／Redis ID与app_data卷保持。2用户／0密钥／3分组／1套餐／0订单／280设置计数不变；设置摘要变化已追溯到唯一`openai_codex_client_version_synced`自动更新，17:14:59日志确认，其余279键与备份逐项一致。公开套餐真实返回1条，4个个人接口匿名401；HTTP新版入口与套餐／FAQ chunk校验通过，浏览器真实套餐→登录保留plan=1、FAQ搜索、手机390px无横向溢出。备份output/guest-20260928-171257/database.dump共594458字节，pg_restore目录校验通过；该目录另含before.json、build.log、settings-diff.json、verification.json和plans-mobile.png。旧镜像`sub2apicust:before-guest-20260928-171257`。未跑Go单测／完整登录支付流程、未下单扣款、未提交推送或操作生产。恢复仍先核实现场，不自动恢复数据库或删卷。

**2026-09-28最新部署（包号与推广清理，优先于历史记录）**：用户授权后启动Docker，代理授权200，compose build退出0，含前端及Go embed编译。仅重建应用容器，原数据库／Redis保留，三服务healthy，health=ok，镜像与容器一致d0c2c1d46026d0699ce3e153c4aba9c102e707275e84320990549512265bf5a8；包号表存在，接口未登录401，HTTP入口index-B18iYC10.js含包号页面。旧镜像sub2apicust:before-dedicated-20260928-112340；备份output/brand/before-dedicated-20260928-112340.dump已校验；构建日志output/brand/docker-dedicated-build.log。未跑Go单测／真实账号联调，待用户浏览器验收，生产未操作。

**最新可读性修正版（优先于本节初版记录）**：当前8080镜像`cfa0a12c8bf6`，仅应用容器重建，三服务healthy；数据库／Redis容器ID、用户1／分组2／套餐0、280项设置摘要保持。前端构建／Go embed编译通过，新HTTP入口`index-BW6NgO3M.js`引用`IntelligenceView-C0f1Aymd.js`，条形对比、矩阵排序、移除源站链接／频率元素及充值width规则均验证。回退tag`sub2apicust:before-intelligence-ux-20260927-022446`，dump `output/brand/before-intelligence-ux-20260927-022446.dump`（583628字节，pg_restore目录校验通过），构建日志与部署记录分别为同目录`docker-intelligence-ux-20260927-022446.log`／`intelligence-ux-deploy-record.json`。76项定向回归通过，未跑Go单测；浏览器连接仍不可用，未声明完成真实视觉验收。未创建套餐或订单，未推送或生产部署。

- 用户要求「重新发docker」，范围按本机解释并执行，未动生产。代理授权预检HTTP200，实际compose build进程设置HTTP_PROXY／HTTPS_PROXY／NO_PROXY；首轮构建成功，无盲目重试。
- 镜像从`b4fa2d8e97a1`切到`1a7a31d0e424`；使用原compose项目和配置执行`up -d --no-deps --no-build --wait --wait-timeout 120 sub2api`，只替换应用容器。三服务healthy，数据库／Redis容器ID及app_data挂载不变。
- Docker内前端构建／类型检查／国际化3测与Go embed编译通过，未执行Go单测。`/health`=200/ok；HTTP入口`index-CYZhTaMW.js`确实引用`IntelligenceView-Cob7bqKS.js`，新chunk包含Codex Radar源地址，CSS包含矩阵样式。
- 更新前后users=1、api_keys=0、groups=1、subscription_plans=0、settings=280，各计数与全部设置聚合摘要一致；不展示具体设置值，不制造测试订单。
- 回退镜像`sub2apicust:before-intelligence-20260927-015820`；数据库备份`output/brand/before-intelligence-20260927-015820.dump`，582060字节，`pg_restore --list`校验通过。部署记录`output/brand/intelligence-deploy-record.json`，基线`intelligence-before.json`，构建日志`docker-intelligence-20260927-015820.log`，均在output/brand且Git忽略。
- 若用户要求镜像回退，先将上述旧镜像重新tag为`sub2apicust:local-theme`，再使用原compose仅重建sub2api；不要自动恢复数据库，不删除卷。上述hash和备份名是本次实证，下次部署仍须重新取证。
- 浏览器控制工具仍不可用，本轮未做真实浏览器登录与视觉验收；HTTP状态和资源校验不是交互验收。
