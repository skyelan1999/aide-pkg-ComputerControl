@echo off
setlocal
cd /d "%~dp0\.."
if not exist dist mkdir dist
set "FRAMEWORK=%WINDIR%\Microsoft.NET\Framework\v4.0.30319"
"%FRAMEWORK%\csc.exe" /nologo /optimize+ /platform:x86 /target:exe /out:dist\AideComputerDesktop.exe /r:System.Drawing.dll /r:System.Windows.Forms.dll /r:System.Web.Extensions.dll /r:"%FRAMEWORK%\WPF\UIAutomationClient.dll" /r:"%FRAMEWORK%\WPF\UIAutomationTypes.dll" /r:"%FRAMEWORK%\WPF\WindowsBase.dll" windows\Desktop.cs
if errorlevel 1 exit /b 1
for %%A in (amd64 386 arm64) do (
 set "GOOS=windows"
 set "GOARCH=%%A"
 call go build -trimpath -ldflags "-s -w" -o dist\AideComputerBridge-windows-%%A.exe ./cmd/bridge
 if errorlevel 1 exit /b 1
)
