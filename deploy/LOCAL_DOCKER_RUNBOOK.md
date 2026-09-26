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
