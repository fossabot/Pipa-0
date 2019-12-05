package backend

type CropTask struct {
	Mode       string
	Width      int
	Height     int
	Long       int
	Short      int
	Limit      bool
	Color      string
	Proportion int
}

type ImageProcess interface {
	//get image blob
	GetImageBlob() []byte
	//get image format
	GetImageFormat() string
	//get image
	ReadImage(fileName string) error
	//resize preprocess image
	ResizePreprocess(captures map[string]string) (*CropTask, error)
	//resize image
	ResizeImage(fileName string, plan *CropTask) (error)

	Terminate()

}