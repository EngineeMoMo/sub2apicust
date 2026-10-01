# 包号专项审查（2026-09-30）

## 后续修复状态（优先于下方审查时结论）

用户授权修复后，工作区已修四项，尚未提交／推送／部署：请求校验与保存校验分离，当前成员健康不再取决于其他成员；移除Key统一inactive、兼容旧disabled，APIKeyService.Update有效状态写入前统一重新验包号资格（含隐式恢复）；编辑必须提供预期updated_at，恢复撤销需明确确认；管理列表区分配置结构异常与成员警告，前端可刷新重开。严格账号独占及历史隔离标记不变，无迁移。

修复回归位于正式custom_dedicated_*测试与dedicated-view.spec.ts，后续证据在output/dedicated-fix-20260930。下方output/dedicated-audit-20260930仍是旧版坏行为复现，不再代表当前源码。测试／构建终态见HANDOFF四；生产历史403/503仍无现场证据，不作已解决保证。

最终验证：40项前端定向及3项国际化、类型／相关lint／构建通过；Docker定向Go service／handler测试和embed编译通过，临时PostgreSQL执行65项service顶层测试通过。新测试断言的是修复后期望行为；非全量Go或生产验收。无部署、无新迁移。

## 结论与范围

审查当前工作区（含未发布的成员移除修复），发现三类经隔离数据库复现的问题，以及一项源码确认的诊断缺口。上一轮修复不宜单独发布。本轮未新增业务修改、未提交推送部署；不能据此认定生产历史403/503请求的根因。

检查范围：成员保存／移除、授权和Key状态、改绑／撤销／删除、缓存接缝、HTTP选择／转发与WebSocket逐轮校验、用户视图白名单及管理页状态。不是全量安全审计或高并发压力测试。

## 1. P1：退出成员修改旧Key状态可阻断剩余成员

- 位置：`backend/internal/service/api_key_service.go:802`、`:822`；`backend/internal/service/custom_dedicated_accounts.go:142`。
- 未发布修复移除成员后撤销组授权，将本组Key设为disabled，但保留用户对Key记录的所有权。Key Update仅在提交group_id时检查组权限，仅提交status仍可成功。
- 复现：移除用户12后用户11校验正常；调用实际APIKeyService.Update，以用户12身份只提交status=active，更新成功；11的专属校验随即ErrDedicatedAccess。存储适配器只执行测试库Key UPDATE，未配置用户／组仓库，证明此路径未重新验证组权限。KeysView启用按钮也只传状态。
- 这不是已证明的上游访问越权：退出者仍受成员校验限制；已证明的是他能影响剩余成员可用性。
- 状态不一致扩大问题：再改为常规停用值inactive，11仍被拒绝，因为完整性SQL只排除disabled。两个停用状态不能混用。
- 建议：区分访问资格和独占结构；恢复Key有效状态必须验证当前组权限／包号成员资格，覆盖配额／到期修改等隐式恢复路径；统一停用状态。不能仅放行inactive就认为修完，也不能删除历史隔离记录。

## 2. P1：停用一个成员连带阻断其他正常成员

- 位置：`backend/internal/service/custom_dedicated_accounts.go:138`、`:316`。
- 请求时复用保存完整性SQL，要求名单内每位用户均active且有授权。
- 复现：两成员正常，用户11校验成功；仅将12设disabled，11即ErrDedicatedAccess；恢复12后11恢复。账号、分组、11的授权均未修改。
- 删除成员账号／撤掉其授权也在同一条件内，但动态测试本轮只复现停用。
- 建议：保存合法性与请求者资格分离，其他成员停用不应扩大成全组停用；仍保留账号独占、无备用路由、禁止影子账号。

## 3. P2：旧编辑表单无冲突提示恢复刚撤销的包号

- 位置：`backend/internal/service/custom_dedicated_accounts.go:213`；`frontend/src/custom/views/AdminDedicatedAccountsView.vue:67`。
- 复现：保留旧input，Revoke后拒绝访问，再用旧input调用Save，revoked_at变NULL、访问恢复。事务行锁不能识别事务开始前已经陈旧的客户端表单。
- 界面已说明保存会重新激活，主动编辑撤销记录的恢复行为本身不是缺陷；缺口是多标签／多管理员旧表单无版本校验，覆盖后来撤销。
- 建议：更新携带并校验预期版本／updated_at，冲突要求刷新；恢复撤销另行明确确认。

## 4. P2：管理页缺少配置健康状态

- 源码证据：`frontend/src/custom/views/AdminDedicatedAccountsView.vue:28`只按撤销时间／到期时间计算状态；AdminList无配置健康字段。用户页可返回configuration_error，而管理页仍显示“生效中”。
- 该文字只代表绑定有效期，不能当可调用证明；容易误导403排查。本轮没有做浏览器动态验收。
- 建议：区分绑定状态与配置健康状态，给管理员具体失败项；不向普通用户泄露内部账号／其他成员资料。

## 证据与边界

- `output/dedicated-audit-20260930/audit_test.go`是临时审计探针，刻意断言当前坏行为，不应作为期望行为合入正式回归。
- `audit.log`：19个既有service顶层测试加1个审计测试全部执行通过；审计有4个子场景（同伴停用、旧表单恢复、退出者active、退出者inactive）。这里“通过”表示成功复现缺陷，不是修复完成。
- `build.log`：Docker内编译审计二进制成功。临时PostgreSQL18无宿主端口、无业务卷、网络none；测试容器共享临时PG网络命名空间，结束已清理PG。
- 镜像sub2apicust-dedicated-audit:20260930；被测custom_dedicated_accounts、custom_dedicated_members、api_key_service三文件SHA256逐一与工作区一致。
- 未运行全量Go／前端测试，未操作业务库、8080和生产；product-samples未改。
- 删除后换人仍可能需清理旧授权／Key；规则已要求提前准备独立授权，暂不把此配置要求当新缺陷。源码中的防越权接缝存在不等于所有路径均已通过动态安全验证。
