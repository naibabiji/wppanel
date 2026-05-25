## v1.0.0-beta9

**安全修复**
- 文件管理 Decompress/Move/Copy 路径校验统一使用 `isPathWithin`，防止同级前缀目录绕过
- Plugin API Key 从 4 字节增至 16 字节，独立于 FastCGI cache key

**新功能**
- 仪表盘公告横幅：从 GitHub `ANNOUNCEMENT.md` 自动读取，30 分钟缓存，可关闭并在新公告时自动弹出
- 扩展配置页面：管理新建站时的默认主题和插件列表，支持一键恢复默认
- 建站/重装 WordPress 时可选：自动清理默认插件、删除未使用主题、勾选安装主题/插件
- 重装 WordPress 同样支持扩展安装选项
