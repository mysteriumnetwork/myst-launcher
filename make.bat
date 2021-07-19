go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.2.0
go generate
go build -ldflags="-s -w -H windowsgui" && myst-node-launcher.exe -tray
