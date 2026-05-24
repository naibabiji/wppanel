### 新增
- 侧边栏新增退出登录按钮
- 新增安全响应头中间件（X-Content-Type-Options / X-Frame-Options / Referrer-Policy / HSTS）
- README 新增安全性章节，含三道锁通俗说明和详细安全机制
- 部署流程加入 `go vet` 静态检查（编译前自动执行，拦截常见代码错误）

### Bug 修复
- 修复面板访问偶发性一直转圈加载（SQLite 连接池限制导致数据库查询无限阻塞）
- 修复数据库自动备份/下载/删除三处路径未包含 db/ 子目录
- 修复 SSL 证书自动续签调度缺失（续签函数已实现但从未被定时触发，现每日凌晨 3 点自动执行）
- 修复构建时间显示 unknown（bootstrap.sh 和部署命令注入 BuildTime）
- 数据库密码从命令行参数改为 MYSQL_PWD 环境变量，避免进程列表泄露
- 远程备份 SSH 主机验证从 no 改为 accept-new（首次自动接受指纹，后续变更则拒绝，防御中间人）

### 移除
- 删除 wp status 命令（wp 无参数已包含运行状态）
