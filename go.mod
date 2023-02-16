module github.com/mysteriumnetwork/myst-launcher

go 1.16

require (
	code.cloudfoundry.org/archiver v0.0.0-20220328120804-99329f9bbb8b
	github.com/Microsoft/go-winio v0.5.2
	github.com/artdarek/go-unzip v1.0.0
	github.com/asaskevich/EventBus v0.0.0-20200907212545-49d423059eef
	github.com/blang/semver/v4 v4.0.0
	github.com/codingsince1985/checksum v1.2.4
	github.com/containerd/containerd v1.5.18 // indirect
	github.com/docker/docker v20.10.10+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/gabriel-samfira/go-wmi v0.0.0-20200311221200-7c023ba1e6b4
	github.com/go-ole/go-ole v1.2.6
	github.com/gonutz/w32 v1.0.0
	github.com/google/cabbie v1.0.3 // indirect
	github.com/google/glazier v0.0.0-20220830195856-e8bb738d64fd
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/iamacarpet/go-win64api v0.0.0-20220720120512-241a9064deec
	github.com/lxn/walk v0.0.0-20210112085537-c389da54e794
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mysteriumnetwork/go-fileversion v1.0.0-fix1
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pkg/errors v0.9.1
	github.com/scjalliance/comshim v0.0.0-20190308082608-cf06d2532c4e
	github.com/stretchr/testify v1.7.0
	github.com/tc-hib/winres v0.1.5
	github.com/tryor/gdiplus v0.0.0-20200830101413-c570de9579b3
	github.com/tryor/winapi v0.0.0-20200525040926-cd87d62e2f9b
	github.com/winlabs/gowin32 v0.0.0-20210302152218-c9e40aa88058
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/toast.v1 v1.0.0-20180812000517-0a84660828b2 // indirect
)

replace (
	//tag:patch-v1
	github.com/gabriel-samfira/go-wmi => github.com/mysteriumnetwork/go-wmi v0.0.0-20211216181752-dbce75057213

	//tag:shlwapi-r4
	github.com/gonutz/w32 => github.com/mysteriumnetwork/w32 v1.0.1-0.20211216070125-4741b8b8111b

	//tag:patch-v1
	github.com/iamacarpet/go-win64api => github.com/mysteriumnetwork/go-win64api v0.0.0-20220913095126-48aad5daeb64

	//branch:linklabel-color
	github.com/lxn/walk => github.com/mysteriumnetwork/walk v0.0.0-20220201093859-484feb886dfe

	//tag:stream-r1
	github.com/tryor/gdiplus => github.com/mysteriumnetwork/gdiplus v0.0.0-20211020173905-2bd21ea15fae
)
