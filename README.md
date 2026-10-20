# Computer Control · Aide 电脑控制

Aide `computer-control` 插件的独立分发仓库，版本 **1.1.0**，MIT 许可。运行代码和安装 ZIP 与[原市场来源](https://github.com/skyelan1999/aide/tree/4dc9a7b0b01fef6ac3550b0b0fcf97db42d45f94/plugins/computer-control)一致。

## 能力

提供六项工具：

- `computer_status`：查询桥接状态，不读取屏幕。
- `computer_inspect`：读取允许的前台应用的辅助功能控件文字及坐标，排除安全密码文本控件；需要原生桥接与辅助功能权限。
- `computer_snapshot`：截取允许的前台应用窗口；需要原生桥接与屏幕录制权限。
- `computer_click`：确认后点击坐标。
- `computer_type`：确认后向当前焦点输入文字，最多 2000 字符。
- `computer_key`：确认后发送有限按键：ENTER、TAB、ESCAPE、BACKSPACE、DELETE、UP、DOWN、LEFT、RIGHT、SPACE。

## 平台与宿主依赖

本版桌面桥接面向 **macOS**；原生 Aide Computer Bridge 构建目标为 **macOS 14+**。Windows/Linux 没有随本包提供的本机桌面桥接实现。Node 插件可在 Aide 运行环境中注册，不代表该环境能够控制本机桌面。

ZIP 不包含原生 `.app` 或桥接服务。需要在目标 Mac 配置 Aide Computer Bridge，并由用户在系统设置授予屏幕录制和辅助功能权限；插件安装不会自动授予权限。宿主配置与构建见 [Aide 文档](https://github.com/skyelan1999/aide/blob/main/docs/plugins/computer-browser-control.md)和 [构建脚本](https://github.com/skyelan1999/aide/blob/main/scripts/build-native-computer-bridge.sh)。应用启动沿用 `start.command`。

由 Aide 启动器向插件进程提供 `AIDE_COMPUTER_BRIDGE_TOKEN`，至少 32 字符。桥接默认 `http://host.docker.internal:17778`，也接受相同端口的 localhost 或 127.0.0.1；不能设置任意远程桥接地址。不要在对话或仓库中粘贴令牌。

宿主还保留 `scripts/computer-bridge.js` 的旧 Node 回退实现：它使用 macOS 命令，**截图为全屏**，且不提供 `computer_inspect`。需要窗口限定截图和控件读取时应使用原生桥接。健康检查成功不代表系统权限已授权或桌面操作已验证。

## 配置与确认

在插件设置中填写允许的应用名称，例如：

```json
{"allowedApps":["Safari","Finder"]}
```

空列表拒绝截图、控件读取和输入操作；应用名称按桥接识别的名称精确匹配，目标必须处于前台。点击、输入和按键沿用 Aide 逐次确认流程，允许应用列表不替代操作确认。确认说明仅记录输入字符数等摘要，不保存完整输入文本。页面、控件和截图内容作为不可信资料处理。

操作针对**桥接所在 Mac**，不会因选择 SSH 工作区自动控制远程设备桌面。工具返回截图数据，不代表模型已具备或使用视觉输入；文本模型定位可使用 `computer_inspect`。

## 安装

市场版本基线为 `0.1.17.0 RC2`，同时需要上述宿主桥接和系统权限。从 [v1.1.0 Release](https://github.com/skyelan1999/aide-pkg-ComputerControl/releases/tag/v1.1.0) 的固定提交链接下载 ZIP 和 SHA256SUMS，校验后通过 Aide 插件管理导入，或在[插件市场](https://skyelan1999.github.io/aide-pkg-shop/)搜索 `computer-control`。

```sh
sha256sum -c SHA256SUMS
```

安装包位于 `packages/v1.1.0/`，保留原 LICENSE、index.js 和 manifest.json，未重新打包；已安装相同 ID/版本无需重复安装。Release 使用固定提交仓库下载链接，没有另行上传二进制附件。原 manifest 默认停用；市场安装会注册并启用，请检查设置。

## 验证与限制

本次在 Linux 验证插件注册、应用范围与操作确认、输入限制、桥接 HTTP 鉴权和 Aide 确认逻辑。**macOS 原生桥接构建、系统授权及真实截图/控件读取/点击/输入/按键未运行**；Linux 测试通过不代表本机桌面控制可用。精确结果见 `docs/release-verification.json`。本次无宿主代码、镜像或产品 UI 变更。

SHA-256 用于完整性校验，不是发布者签名；Node 插件在宿主执行，插件机制不是权限沙箱。
