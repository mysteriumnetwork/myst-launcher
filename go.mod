module github.com/mysteriumnetwork/myst-launcher

go 1.20

require (
	code.cloudfoundry.org/archiver v0.0.0-20230320125732-4c59f8192b7d
	github.com/Microsoft/go-winio v0.6.1
	github.com/advbet/sseclient v0.0.0-20220420050220-2d425222e61b
	github.com/artdarek/go-unzip v1.0.0
	github.com/asaskevich/EventBus v0.0.0-20200907212545-49d423059eef
	github.com/bi-zone/go-fileversion v1.0.0
	github.com/blang/semver/v4 v4.0.0
	github.com/codingsince1985/checksum v1.3.0
	github.com/docker/docker v24.0.9+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/gabriel-samfira/go-wmi v0.0.0-20200311221200-7c023ba1e6b4
	github.com/go-ole/go-ole v1.2.6
	github.com/gonutz/w32 v1.0.0
	github.com/google/glazier v0.0.0-20230310154036-250f5ed41d5b
	github.com/iamacarpet/go-win64api v0.0.0-20221230174906-cb41e6e774e8
	github.com/lxn/walk v0.0.0-20210112085537-c389da54e794
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e
	github.com/mysteriumnetwork/node v0.0.0-20240618102753-adb67589d429
	github.com/otiai10/copy v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	github.com/scjalliance/comshim v0.0.0-20230315213746-5e51f40bd3b9
	github.com/shirou/gopsutil/v3 v3.23.9
	github.com/stretchr/testify v1.8.4
	github.com/tc-hib/winres v0.1.6
	github.com/tryor/gdiplus v0.0.0-20200830101413-c570de9579b3
	github.com/tryor/winapi v0.0.0-20200525040926-cd87d62e2f9b
	github.com/winlabs/gowin32 v0.0.0-20221003142512-0d265587d3c9
	golang.org/x/net v0.24.0
	golang.org/x/sys v0.19.0
)

require (
	github.com/arthurkiller/rollingwriter v1.1.3-0.20220211070658-c19a8e8b35be // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/cabbie v1.0.5 // indirect
	github.com/google/deck v1.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/image v0.10.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/toast.v1 v1.0.0-20180812000517-0a84660828b2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.0.2 // indirect
)

replace (
	github.com/bi-zone/go-fileversion => github.com/mysteriumnetwork/go-fileversion v1.0.0

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
