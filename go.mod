module github.com/mysteriumnetwork/myst-launcher

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.17
	github.com/asaskevich/EventBus v0.0.0-20200907212545-49d423059eef
	github.com/buger/jsonparser v0.0.0-20180808090653-f4dd9f5a6b44
	github.com/containerd/containerd v1.5.3 // indirect
	github.com/docker/docker v20.10.7+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/go-ole/go-ole v1.2.5
	github.com/gonutz/w32 v1.0.0
	github.com/lxn/walk v0.0.0-20210112085537-c389da54e794
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mysteriumnetwork/go-fileversion v1.0.0-fix1
	github.com/tc-hib/winres v0.1.5
	github.com/tryor/gdiplus v0.0.0-20200830101413-c570de9579b3
	github.com/tryor/winapi v0.0.0-20200525040926-cd87d62e2f9b
	github.com/winlabs/gowin32 v0.0.0-20210302152218-c9e40aa88058
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	google.golang.org/grpc v1.39.0 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
)

replace (
	//tag:shlwapi-r1
	github.com/gonutz/w32 => github.com/mysteriumnetwork/w32 v1.0.1-0.20211020171222-078e36ca2fb8

	//tag:stream-r1
	github.com/tryor/gdiplus => github.com/mysteriumnetwork/gdiplus v0.0.0-20211020173905-2bd21ea15fae
)
