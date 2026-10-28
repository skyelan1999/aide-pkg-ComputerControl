@echo off
setlocal
cd /d "%~dp0"
AideComputerBridge.exe %*
if errorlevel 1 pause
