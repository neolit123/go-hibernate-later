@echo off

echo calling go-windres...
go-winres simply --icon .\icon.png --manifest gui

echo taskkill /im go-hibernate-later.exe
taskkill /im go-hibernate-later.exe

go build -ldflags "-H=windowsgui"
