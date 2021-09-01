package _const

var ImageTag = "latest"

const (
	ImageName = "mysteriumnetwork/myst"
)

func GetImageName() string {
	return ImageName + ":" + ImageTag
}
