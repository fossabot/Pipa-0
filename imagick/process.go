package imagick

import (
	"encoding/base64"
	"errors"
	"gopkg.in/gographics/imagick.v3/imagick"
	"pipa/backend"
	"pipa/helper"
	"strconv"
)

type ImageWand struct {
	MagickWand *imagick.MagickWand
	PixelWand  *imagick.PixelWand
	DrawWand   *imagick.DrawingWand
}

func Initialize() backend.ImageProcess {
	imagick.Initialize()

	imageProcess := NewImageWand()

	return imageProcess
}

func (img *ImageWand) Terminate() {
	imagick.Terminate()
}

func NewImageWand() (imageProcess backend.ImageProcess) {
	img := ImageWand{
		MagickWand: imagick.NewMagickWand(),
		PixelWand:  imagick.NewPixelWand(),
		DrawWand:   imagick.NewDrawingWand(),
	}
	helper.Logger.Info("MagickWand ready: ", img.MagickWand, "PixelWand ready: ", img.PixelWand, "DrawWand ready: ", img.DrawWand)
	return &img
}

func (img *ImageWand) ResizePreprocess(captures map[string]string) (*backend.ResizeTask, error) {

	n := backend.ResizeTask{}
	var err error

	if captures["P"] != "" {
		if captures["fileName"] == "" {
			return &backend.ResizeTask{}, errors.New("wrong resize fileName detect")
		}
		n.FileName = captures["fileName"]
		n.Proportion, err = strconv.Atoi(captures["P"])
		if err != nil {
			return &backend.ResizeTask{}, errors.New("wrong resize P detect")
		}
		if n.Proportion < 1 || n.Proportion > 100 {
			return &backend.ResizeTask{}, errors.New("wrong resize P detect")
		}
	} else {
		n.FileName = ""
	}

	if captures["p"] == "" {
		n.Proportion = 0
	} else {
		n.Proportion, err = strconv.Atoi(captures["p"])
		if err != nil {
			return &backend.ResizeTask{}, errors.New("wrong resize P detect")
		}
		if n.Proportion < 1 || n.Proportion > 100 {
			return &backend.ResizeTask{}, errors.New("wrong resize p detect")
		}
		return &n, nil
	}

	if captures["w"] == "" {
		n.Width = 0
	} else {
		n.Width, _ = strconv.Atoi(captures["w"])
		if n.Width < 1 || n.Width > 4096 {
			return &backend.ResizeTask{}, errors.New("wrong resize width detect")
		}
	}

	if captures["h"] == "" {
		n.Height = 0
	} else {
		n.Height, _ = strconv.Atoi(captures["h"])
		if n.Height < 1 || n.Height > 4096 {
			return &backend.ResizeTask{}, errors.New("wrong resize height detect")
		}
	}

	if captures["l"] == "" {
		n.Long = 0
	} else {
		n.Long, _ = strconv.Atoi(captures["l"])
		if n.Long < 1 || n.Long > 4096 {
			return &backend.ResizeTask{}, errors.New("wrong resize long detect")
		}
	}

	if captures["s"] == "" {
		n.Short = 0
	} else {
		n.Short, _ = strconv.Atoi(captures["s"])
		if n.Short < 1 || n.Short > 4096 {
			return &backend.ResizeTask{}, errors.New("wrong resize short detect")
		}
	}

	if captures["limit"] == "" {
		n.Limit = true
	} else {
		limit, _ := strconv.Atoi(captures["limit"])
		if limit == 1 {
			n.Limit = true
		} else if limit == 0 {
			n.Limit = false
		} else {
			return &backend.ResizeTask{}, errors.New("wrong resize limit detect")
		}

	}

	n.Color = checkColor(captures["color"])

	switch captures["m"] {
	case "", "lfit":
		n.Mode = "lfit"
		if ((n.Width != 0 || n.Height != 0) && (n.Long != 0 || n.Short != 0)) == true {
			return &backend.ResizeTask{}, errors.New("can not resize in height&width and long&short at the same time")
		}
	case "mfit":
		n.Mode = "mfit"
		if ((n.Width != 0 || n.Height != 0) && (n.Long != 0 || n.Short != 0)) == true {
			return &backend.ResizeTask{}, errors.New("can not resize in height&width and long&short at the same time")
		}
	case "fill", "pad", "fixed":
		n.Mode = captures["m"]
		if n.Width != 0 && n.Height == 0 {
			n.Height = n.Width
		}

		if n.Width == 0 && n.Height != 0 {
			n.Width = n.Height
		}
	default:
		return &backend.ResizeTask{}, errors.New("wrong resize mode detect")
	}
	return &n, nil
}

func (img *ImageWand) ResizeImage(fileName string, plan *backend.ResizeTask) error {
	helper.Logger.Println("start resize image, plan: ", plan)
	err := img.ReadImage(fileName)
	if err != nil {
		helper.Logger.Error("open temp file failed")
		return err
	}
	o := newResize()
	o.Limit = plan.Limit
	o.Background = plan.Color

	if plan.FileName != "" {
		picture := imagick.NewMagickWand()
		err = picture.ReadImage(plan.FileName)
		if err != nil {
			helper.Logger.Error("open origin picture file failed")
			return err
		}
		originWidth := int(picture.GetImageWidth())
		originHeight := int(picture.GetImageHeight())
		factor := float64(plan.Proportion) / 100.0
		widthFactor := float64(originWidth / plan.Width)
		heightFactor := float64(originHeight / plan.Height)
		if widthFactor*factor < heightFactor {
			factor = widthFactor * factor
		} else {
			factor = heightFactor * factor
		}
		o.Zoom = factor
		err = img.resize(o)
		if err != nil {
			return err
		}
		return nil
	}

	//proportion zoom
	if plan.Proportion != 0 {
		factor := float64(plan.Proportion) / 100.0
		helper.Logger.Info("scaling factor: ", factor)
		o.Zoom = factor
		err = img.resize(o)
		if err != nil {
			return err
		}
		return nil
	}

	switch plan.Mode {
	//长边优先
	case "lfit":
		adjustCropTask(plan, img.MagickWand.GetImageWidth(), img.MagickWand.GetImageHeight())
		o := Resize{Width: plan.Width, Height: plan.Height}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	//短边优先
	case "mfit":
		adjustCropTask(plan, img.MagickWand.GetImageWidth(), img.MagickWand.GetImageHeight())
		o := Resize{Width: plan.Width, Height: plan.Height}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	case "pad":
		o := Resize{Width: plan.Width, Height: plan.Height, Pad: true}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	case "fixed":
		o := Resize{Width: plan.Width, Height: plan.Height, Force: true}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	case "fill":
		o := Resize{Width: plan.Width, Height: plan.Height, Crop: true}
		err = img.resize(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (img *ImageWand) WatermarkPreprocess(captures map[string]string) (*backend.WatermarkTask, error) {

	n := backend.WatermarkTask{}

	if captures["t"] == "" {
		n.Transparency = Transparency
	} else {
		n.Transparency, _ = strconv.Atoi(captures["t"])
		if n.Transparency < 0 || n.Transparency > 100 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark t detect")
		}
	}

	switch captures["g"] {
	case "nw":
		n.Position = NorthWest
	case "north":
		n.Position = North
	case "ne":
		n.Position = NorthEast
	case "west":
		n.Position = West
	case "center":
		n.Position = Center
	case "east":
		n.Position = East
	case "sw":
		n.Position = SouthWest
	case "south":
		n.Position = South
	case "se":
		n.Position = SouthEast
	default:
		if captures["g"] != "" {
			return &backend.WatermarkTask{}, errors.New("wrong watermark g detect")
		}
		n.Position = SouthEast
	}

	if captures["x"] == "" {
		n.XMargin = 10
	} else {
		n.XMargin, _ = strconv.Atoi(captures["x"])
		if n.XMargin < 0 || n.XMargin > 4096 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark x detect")
		}
	}

	if captures["y"] == "" {
		n.YMargin = 10
	} else {
		n.YMargin, _ = strconv.Atoi(captures["y"])
		if n.YMargin < 0 || n.YMargin > 4096 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark y detect")
		}
	}

	if captures["voffset"] == "" {
		n.Voffset = 0
	} else {
		n.Voffset, _ = strconv.Atoi(captures["voffset"])
		if n.YMargin < -1000 || n.YMargin > 1000 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark y detect")
		}
	}

	if captures["image"] == "" {
		n.PictureMask.Image = ""
	} else {
		imageUrl, err := base64.StdEncoding.DecodeString(captures["image"])
		if err != nil {
			helper.Logger.Warn("Image base64 code is invalid!")
			return &backend.WatermarkTask{}, errors.New("Image base64 code is invalid ")
		}
		n.PictureMask.Image = string(imageUrl)
	}

	if captures["text"] == "" {
		n.TextMask.Text = ""
	} else {
		if len(captures["text"]) > 64 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark text to long")
		}
		text, err := base64.StdEncoding.DecodeString(captures["text"])
		if err != nil {
			helper.Logger.Warn("Text base64 code is invalid!")
			return &backend.WatermarkTask{}, errors.New("Text base64 code is invalid ")
		}
		n.TextMask.Text = string(text)
	}

	if captures["type"] == "" {
		n.TextMask.Type = DefaultTextType
	} else {
		textType, err := base64.StdEncoding.DecodeString(captures["type"])
		if err != nil {
			helper.Logger.Warn("Type base64 code is invalid!")
			return &backend.WatermarkTask{}, errors.New("Type base64 code is invalid ")
		}
		n.TextMask.Type = string(textType)
	}

	n.TextMask.Color = checkColor(captures["color"])

	if captures["size"] == "" {
		n.TextMask.Size = FrontSize
	} else {
		n.TextMask.Size, _ = strconv.Atoi(captures["size"])
		if n.YMargin < 0 || n.YMargin > 1000 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark text size detect")
		}
	}

	if captures["shadow"] == "" {
		n.TextMask.Shadow = 0
	} else {
		n.TextMask.Shadow, _ = strconv.Atoi(captures["shadow"])
		if n.YMargin < 0 || n.YMargin > 100 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark text shadow detect")
		}
	}

	if captures["rotate"] == "" {
		n.TextMask.Rotate = 0
	} else {
		n.TextMask.Rotate, _ = strconv.Atoi(captures["rotate"])
		if n.TextMask.Rotate < 0 || n.TextMask.Rotate > 360 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark text rotate detect")
		}
	}

	if captures["fill"] == "" {
		n.TextMask.Fill = false
	} else {
		fill, _ := strconv.Atoi(captures["limit"])
		if fill == 1 {
			n.TextMask.Fill = true
		} else if fill == 0 {
			n.TextMask.Fill = false
		} else {
			return &backend.WatermarkTask{}, errors.New("wrong watermark text fill detect")
		}
	}

	if captures["order"] == "" {
		n.Order = 0
	} else {
		order, _ := strconv.Atoi(captures["order"])
		if order == 1 {
			n.Order = 1
		} else if order == 0 {
			n.Order = 0
		} else {
			return &backend.WatermarkTask{}, errors.New("wrong watermark order detect")
		}
	}

	if captures["align"] == "" {
		n.Align = 0
	} else {
		align, _ := strconv.Atoi(captures["align"])
		if align == 0 {
			n.Align = 0
		} else if align == 1 {
			n.Align = 1
		} else if align == 2 {
			n.Align = 2
		} else {
			return &backend.WatermarkTask{}, errors.New("wrong watermark align detect")
		}
	}

	if captures["interval"] == "" {
		n.Interval = 0
	} else {
		n.Interval, _ = strconv.Atoi(captures["interval"])
		if n.Interval < 0 || n.Interval > 1000 {
			return &backend.WatermarkTask{}, errors.New("wrong watermark text rotate detect")
		}
	}

	return &n, nil
}

func (img *ImageWand) ImageWatermark(fileName string, plan *backend.WatermarkTask) error {
	helper.Logger.Println("start resize image, plan: ", plan)
	err := img.ReadImage(fileName)
	if err != nil {
		helper.Logger.Error("open temp file failed")
		return err
	}
	w := newWatermark()
	originWidth := int(img.MagickWand.GetImageWidth())
	originHeight := int(img.MagickWand.GetImageHeight())
	if plan.PictureMask.Image != "" {
		picture := imagick.NewMagickWand()
		err = picture.ReadImage(plan.PictureMask.Filename)
		if err != nil {
			helper.Logger.Error("open watermark picture file failed")
			return err
		}
		wmWidth := int(picture.GetImageWidth())
		wmHeight := int(picture.GetImageHeight())

		w.Picture = picture
		w.Transparency = plan.Transparency
		switch plan.Position {
		case NorthWest:
			w.XMargin = plan.XMargin
			w.YMargin = plan.YMargin
			break
		case North:
			w.XMargin = (originWidth - wmWidth) / 2
			w.YMargin = plan.YMargin
			break
		case NorthEast:
			w.XMargin = originWidth - plan.XMargin - wmWidth
			w.YMargin = plan.YMargin
			break
		case West:
			w.XMargin = plan.XMargin
			w.YMargin = (originHeight-wmHeight)/2 - plan.Voffset
			break
		case Center:
			w.XMargin = (originWidth - wmWidth) / 2
			w.YMargin = (originHeight-wmHeight)/2 - plan.Voffset
			break
		case East:
			w.XMargin = originWidth - plan.XMargin - wmWidth
			w.YMargin = (originHeight-wmHeight)/2 - plan.Voffset
			break
		case SouthWest:
			w.XMargin = plan.XMargin
			w.YMargin = originHeight - plan.YMargin - wmHeight
			break
		case South:
			w.XMargin = (originWidth - wmWidth) / 2
			w.YMargin = originHeight - plan.YMargin - wmHeight
			break
		case SouthEast:
			w.XMargin = originWidth - plan.XMargin - wmWidth
			w.YMargin = originHeight - plan.YMargin - wmHeight
			break
		default:
			w.XMargin = originWidth - plan.XMargin - wmWidth
			w.YMargin = originHeight - plan.YMargin - wmHeight
		}
		err = img.watermark(w)
		if err != nil {
			return err
		}
		return nil
	} else if plan.TextMask.Text != "" {
		w.Transparency = plan.Transparency
		w.Text.color = plan.TextMask.Color
		w.Text.textType = selectTextType(plan.TextMask.Type)
		w.Text.fontSize = plan.TextMask.Size
		w.Text.shadow = plan.TextMask.Shadow
		w.Text.rotate = plan.TextMask.Rotate
		w.Text.fill = plan.TextMask.Fill
		switch plan.Position {
		case NorthWest:
			w.Gravity = imagick.GRAVITY_NORTH_WEST
			w.XMargin = plan.XMargin
			w.YMargin = plan.YMargin
			break
		case North:
			w.Gravity = imagick.GRAVITY_NORTH
			w.XMargin = 0
			w.YMargin = plan.YMargin
			break
		case NorthEast:
			w.Gravity = imagick.GRAVITY_NORTH_EAST
			w.XMargin = plan.XMargin
			w.YMargin = plan.YMargin
			break
		case West:
			w.Gravity = imagick.GRAVITY_WEST
			w.XMargin = plan.XMargin
			w.YMargin = -plan.Voffset
			break
		case Center:
			w.Gravity = imagick.GRAVITY_CENTER
			w.XMargin = 0
			w.YMargin = -plan.Voffset
			break
		case East:
			w.Gravity = imagick.GRAVITY_EAST
			w.XMargin = plan.XMargin
			w.YMargin = -plan.Voffset
			break
		case SouthWest:
			w.Gravity = imagick.GRAVITY_SOUTH_WEST
			w.XMargin = plan.XMargin
			w.YMargin = plan.YMargin
			break
		case South:
			w.Gravity = imagick.GRAVITY_SOUTH
			w.XMargin = 0
			w.YMargin = plan.YMargin
			break
		case SouthEast:
			w.Gravity = imagick.GRAVITY_SOUTH_EAST
			w.XMargin = plan.XMargin
			w.YMargin = plan.YMargin
			break
		default:
			w.Gravity = imagick.GRAVITY_SOUTH_EAST
			w.XMargin = plan.XMargin
			w.YMargin = plan.YMargin
		}
		err = img.watermark(w)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("have not watermark")
	}
}

func (img *ImageWand) GetImageBlob() []byte {
	return img.MagickWand.GetImageBlob()
}

func (img *ImageWand) GetImageFormat() string {
	return img.MagickWand.GetImageFormat()
}

func (img *ImageWand) ReadImage(fileName string) error {
	return img.MagickWand.ReadImage(fileName)
}

func (img *ImageWand) WriteImage(fileName string) error {
	return img.MagickWand.WriteImage(fileName)
}
