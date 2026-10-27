# Computer Control · Aide 电脑控制 1.2.0

Aide plugin plus a separately installed native host bridge. Supports the existing macOS bridge and adds a **Windows bridge with no PowerShell dependency**. MIT license; Go runtime license accompanies Windows binaries.

Aide 插件与宿主桥接分开分发。新版增加 Windows 原生桥接，安装、启动和桌面操作不调用 PowerShell。插件不预装宿主程序，也不自动打开桌面权限。

## Windows

See [Windows installation and compatibility](docs/windows.md). Download the plugin ZIP and the matching precompiled host ZIP from [v1.2.0](https://github.com/skyelan1999/aide-pkg-ComputerControl/releases/tag/v1.2.0).

1. Install Computer Control 1.2.0 from the market or plugin ZIP. Allow process names such as `notepad`, `msedge`, `chrome` (optional `.exe`). Empty allowlists deny access.
2. Extract the Windows host ZIP, keeping both EXEs together, and run `start.cmd`. Select the Aide installation folder to pair the dedicated token in `.env`, then restart Aide once. No host Go/Node/Python/compiler or PowerShell installation is needed.
3. Keep the bridge console open and put an allowed application in the foreground. Closing the bridge stops desktop access.

Windows 目标为 Windows 10/11、Server 2016+ Desktop Experience 的已解锁交互会话；需桌面 .NET Framework 4.6.2+（建议 4.8）。锁屏、UAC 安全桌面、Session 0、Server Core 及系统阻止的提权窗口不支持控制。AMD64、32 位及 ARM64 传输程序分别提供；ARM 设备需系统的 x86/.NET 兼容支持。具体执行过的 CI 和尚未实机验证的系统见 [验证记录](docs/release-verification.json)。

## macOS

The existing native Aide Computer Bridge targets macOS 14+, with screen-recording/accessibility permissions explicitly granted by the user. Configure allowed application names such as `Safari` or `Finder`. Native host setup remains in [Aide documentation](https://github.com/skyelan1999/aide/blob/main/docs/plugins/computer-browser-control.md). The older Node fallback captures the full screen and lacks control inspection; use the native bridge for window-scoped capture.

## Tools / 工具

- `computer_status`: authenticated connection/capability status; no screen read.
- `computer_inspect`: foreground-window UI controls, excluding password elements. UIA/accessibility providers can omit controls.
- `computer_snapshot`: target window PNG; protected or GPU-only surfaces may be blank/unavailable. Windows never falls back to full-desktop capture.
- `computer_click`, `computer_type`, `computer_key`: Aide-confirmed actions scoped to the allowed foreground window; Unicode input limited to 2000 UTF-16 code units and a fixed key allowlist. No general shell, arbitrary command or file API.

截图和控件文字是不可信资料；允许应用不替代逐次确认。桥接只接受专用令牌与本机/Docker 私有网络来源，拒绝浏览器跨源请求。令牌由宿主 `.env` 配置并只注入对应插件调用，不应放入插件 settings 或对话。支持 `http://host.docker.internal:17778`，同端口 `localhost` / `127.0.0.1`；不可改成任意远端地址。

The bridge operates the desktop where it is running; selecting an SSH workspace does not grant control of another computer. Plugin ZIP import, metadata/platform compatibility and native desktop execution are separate acceptance steps. Node plugins execute in the host's plugin runtime; the plugin mechanism is not a security sandbox.

## Source, tests and release

`windows\build.cmd` and all Windows CI run steps use CMD/native tools. `go test ./...`, `node --test tests/plugin.cjs` and the isolated Windows fixture exercise protocol, denial paths, real UI Automation, password exclusion, window PNG, click, Unicode typing and keys. Packaging uses `scripts/package.py`; end users use the published ZIPs and SHA256SUMS.

The original 1.1.0 package, license and historical verification are retained. No private Aide research code, user conversations or runtime data are distributed.
