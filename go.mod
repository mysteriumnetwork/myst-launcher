module github.com/mysteriumnetwork/myst-launcher

go 1.16

require (
	code.cloudfoundry.org/archiver v0.0.0-20220328120804-99329f9bbb8b
	github.com/Microsoft/go-winio v0.5.1
	github.com/artdarek/go-unzip v1.0.0
	github.com/asaskevich/EventBus v0.0.0-20200907212545-49d423059eef
	github.com/blang/semver/v4 v4.0.0
	github.com/codingsince1985/checksum v1.2.4
	github.com/containerd/containerd v1.5.8 // indirect
	github.com/docker/docker v20.10.10+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/gabriel-samfira/go-wmi v0.0.0-20200311221200-7c023ba1e6b4
	github.com/go-ole/go-ole v1.2.6
	github.com/gonutz/w32 v1.0.0
	github.com/google/glazier v0.0.0-20211213200644-0506347f83ee
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/iamacarpet/go-win64api v0.0.0-20211130162011-82e31fe23f80 // indirect
	github.com/lxn/walk v0.0.0-20210112085537-c389da54e794
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mysteriumnetwork/go-fileversion v1.0.0-fix1
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/tc-hib/winres v0.1.5
	github.com/tryor/gdiplus v0.0.0-20200830101413-c570de9579b3
	github.com/tryor/winapi v0.0.0-20200525040926-cd87d62e2f9b
	github.com/winlabs/gowin32 v0.0.0-20210302152218-c9e40aa88058
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	honnef.co/go/tools v0.3.2
)

replace (
	//tag:patch-v1
	github.com/gabriel-samfira/go-wmi => github.com/mysteriumnetwork/go-wmi v0.0.0-20211216181752-dbce75057213

	//tag:shlwapi-r4
	github.com/gonutz/w32 => github.com/mysteriumnetwork/w32 v1.0.1-0.20211216070125-4741b8b8111b

	//branch:linklabel-color
	github.com/lxn/walk => github.com/mysteriumnetwork/walk v0.0.0-20220201093859-484feb886dfe

	//tag:stream-r1
	github.com/tryor/gdiplus => github.com/mysteriumnetwork/gdiplus v0.0.0-20211020173905-2bd21ea15fae
)
