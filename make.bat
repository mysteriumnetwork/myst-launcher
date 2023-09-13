go run cmd/resource/resource.go

go build -v -trimpath -ldflags "-s -w -H windowsgui" -o bin/myst-launcher-amd64.exe github.com/mysteriumnetwork/myst-launcher/cmd/app
go build -v -trimpath -ldflags "-s -w -H windowsgui -X github.com/mysteriumnetwork/myst-launcher/const.VendorID=Kryptex_AU" -o bin/myst-launcher-kryptex-amd64.exe github.com/mysteriumnetwork/myst-launcher/cmd/app
go build -v -trimpath -ldflags "-s -w" -o bin/myst-launcher-cli.exe github.com/mysteriumnetwork/myst-launcher/cmd/app-cli