# 1.2.0

- Native Windows host for Windows 10/11 and interactive Server 2016+, with AMD64/386/ARM64 transports. No PowerShell dependency.
- Foreground-window capture, UI Automation inspection excluding passwords, scoped clicks, Unicode input and limited keys; retain Aide confirmation.
- Explicit local pairing, bounded authenticated HTTP, session/integrity restrictions, and isolated native-window CI coverage.
- See docs/windows.md and release-verification.json for exact tested and untested environments.

# Changelog

## 1.1.0

- 独立分发现有 Computer Control，保留原版本、MIT 许可、运行代码和市场 ZIP。
- 补充六项工具、macOS 原生桥接、系统权限、应用范围及操作确认说明。
- 更正市场平台声明为 macOS，区分原生窗口截图与旧 Node 回退全屏截图。
- 记录 Linux 逻辑/协议验证与 macOS 实机未验证范围。
