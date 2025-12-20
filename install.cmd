@echo off
REM LlamaGate Universal Installer Launcher for Windows
REM This is a convenience launcher that calls the Windows installer

cd /d "%~dp0"
call install\windows\install.cmd

