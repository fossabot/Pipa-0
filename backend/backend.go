package backend

type ResizeTask struct {
	Mode       string
	Width      int
	Height     int
	Long       int
	Short      int
	Limit      bool
	Color      string
	Proportion int
}

type WatermarkTask struct {
	Transparency int //透明度
	Position     string
	XMargin      int
	YMargin      int
	Voffset      int
	PictureMask  WatermarkPicture
	TextMask     WatermarkText
	Order        int //default 0 图片水印在前 1 文字水印在前
	Align        int //default 0 图片文字上对齐 1 中对齐 2 下对齐
	Interval     int
}

type WatermarkPicture struct {
	Image      string
	Filename   string
	Proportion int
	Rotate     Rotate
	Crop       Crop
}

type WatermarkText struct {
	Text   string
	Type   string
	Color  string
	Size   int
	Shadow int
	Rotate int
	Fill   bool
}

type Crop struct {
}

type Rotate struct {
}

type ImageProcess interface {
	//get image blob
	GetImageBlob() []byte
	//get image format
	GetImageFormat() string
	//get image
	ReadImage(fileName string) error
	//resize preprocess image
	ResizePreprocess(captures map[string]string) (*ResizeTask, error)
	//resize image
	ResizeImage(fileName string, plan *ResizeTask) error
	//watermark preprocess image
	WatermarkPreprocess(captures map[string]string) (*WatermarkTask, error)
	//watermark process image
	ImageWatermark (fileName string, plan *WatermarkTask) error
	Terminate()
}
