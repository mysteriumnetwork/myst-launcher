go run cmd/resource/resource.go

set GOBIN=%cd%\bin
go install -v -trimpath -ldflags "-s -w -H windowsgui" github.com/mysteriumnetwork/myst-launcher/cmd/app
move bin\app.exe bin\myst-launcher-amd64-dbg.exe