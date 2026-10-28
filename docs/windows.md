# Windows 电脑控制 / Computer Control

Version 1.2.0. The Windows bridge uses native Win32, UI Automation and .NET APIs. Installation, startup, builds and desktop operations do not invoke PowerShell, WSL shell scripts, Node.js or Python on the user's host. Go/.NET compilation is only needed by developers; download the precompiled ZIP as an end user.

## 安装 / Install

1. 在插件市场安装/更新 Computer Control 1.2.0。在允许应用中填写进程名称，例如 `notepad`、`msedge`、`chrome`；可带 `.exe`，不填窗口标题。保留逐次操作确认。
2. 下载与你 Windows CPU 相符的桥接 ZIP：通常选 `windows-amd64`（x64），32 位 Windows 10 选 `windows-386`，ARM 设备选 `windows-arm64`。解压到本机固定目录，两个 EXE 必须放在一起。
3. 双击 `AideComputerBridge.exe`（或 `start.cmd`），在原生目录对话框中选择 Aide 安装目录（包含 `compose.yaml`）。这会复用或生成 `.env` 中的 `AIDE_COMPUTER_BRIDGE_TOKEN`，不改其他配置、镜像或数据。令牌不显示到控制台。只在本机运行可信的官方桥接程序。
4. 重启 Aide 一次，让容器读取令牌；保持桥接窗口打开。如果 Windows 防火墙提示，按本机 Docker 私有网络情况授权，程序不会自动扩大防火墙规则。执行 `computer_status` 检查连接。
5. 将一个允许的应用置于前台，再读取控件、截图或确认后操作。关闭桥接窗口即可撤销桌面访问。

Install the plugin, allow executable process names, download/extract the matching native host ZIP and run `AideComputerBridge.exe` (or `start.cmd`). Choose the Aide installation folder to pair the token, restart Aide once and leave the bridge window open. A foreground allowed application and per-action Aide confirmation remain required. Close the bridge window to stop access. No Go, C# compiler, PowerShell, Node or Python installation is required on the end-user machine.

## 范围 / Compatibility

- Target: Windows 10/11 and Windows Server 2016+ **with Desktop Experience and an unlocked interactive user session**. Session 0, Server Core, UAC/lock/secure desktops and disconnected remote desktops cannot be controlled. This is a desktop helper compatibility target, not a claim that Docker Desktop supports Windows Server.
- Requires the desktop .NET Framework 4.6.2+ (4.8 recommended), including WinForms/UI Automation. The helper is x86 managed code; ARM Windows requires its built-in x86 compatibility and applicable .NET support. The transport executable is built for AMD64, 386 and ARM64; compiled architecture does not by itself establish hardware acceptance.
- Native APIs avoid OS-language-dependent command output. Unicode text, spaces/non-ASCII paths and physical desktop coordinates (including negative monitor positions) are supported. Capture uses PrintWindow for the target window only; protected/GPU-only applications may return unavailable or blank content. No full-desktop fallback occurs.
- Runs at the user's normal integrity level. Windows may block interaction with elevated applications; it does not elevate, disable UAC, change input-desktop security or bypass application protections.
- Listens on port 17778 with a dedicated bearer token, rejects browser-origin requests, and accepts only loopback and the documented Docker private subnets. It provides no shell, arbitrary process launch, file-read or remote tunnelling endpoint. Aide must be able to resolve/reach `host.docker.internal:17778` from its execution container.
- UI Automation providers can omit controls. Password elements are excluded, output is bounded and a stuck provider is timed out. A window/app switch aborts subsequent input rather than continuing in a different application.

## Build / 构建

Run `windows\build.cmd` from a normal CMD prompt with Go 1.26 and the system .NET Framework compiler installed. This produces native Go transports for three CPU targets and the UI Automation helper. `python scripts/package.py` is a developer packaging step only. CI Windows steps explicitly use `cmd`, never PowerShell.

Go's official supported Windows baseline is Windows 10 / Server 2016 or later: https://go.dev/wiki/MinimumRequirements . API availability and build success are distinct from end-user acceptance on each release of Windows. Exact executed checks and untested versions are recorded in `docs/release-verification.json`.
