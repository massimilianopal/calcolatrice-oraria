#!/bin/sh
# Compila CalcolatriceOre.exe (Windows 64 bit). Funziona da Linux, macOS o Windows con Go installato.
set -e
go run github.com/akavel/rsrc@v0.10.2 -manifest app.manifest -ico icona.ico -arch amd64 -o rsrc_windows_amd64.syso
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-H windowsgui -s -w" -o CalcolatriceOre.exe .
