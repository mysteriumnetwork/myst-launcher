go run cmd/resource/resource.go
go build -v -trimpath -ldflags "-s -w -H windowsgui" -o  bin/myst-launcher-amd64.exe github.com/mysteriumnetwork/myst-launcher/cmd/app
go build -v -trimpath -ldflags "-s -w" -o  bin/myst-launcher-dbg-amd64.exe github.com/mysteriumnetwork/myst-launcher/cmd/app