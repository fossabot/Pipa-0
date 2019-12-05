package imagick

import (
	"errors"
	"gopkg.in/gographics/imagick.v3/imagick"
	"pipa/backend"
	"pipa/helper"
	"strconv"
)

type ImageWand struct {
	MagickWand *imagick.MagickWand
	PixelWand  *imagick.PixelWand
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
	}
	helper.Logger.Info("MagickWand ready: ", img.MagickWand, "PixelWand ready: ", img.PixelWand)
	return &img
}

func (img *ImageWand) ResizePreprocess(captures map[string]string) (*backend.CropTask, error) {

	n := backend.CropTask{}

	if captures["p"] == "" {
		n.Proportion = 0
	} else {
		n.Proportion, _ = strconv.Atoi(captures["p"])
		if n.Proportion < 1 || n.Proportion > 1000 {
			return &backend.CropTask{}, errors.New("wrong resize p detect")
		}
		return &n, nil
	}

	if captures["w"] == "" {
		n.Width = 0
	} else {
		n.Width, _ = strconv.Atoi(captures["w"])
		if n.Width < 1 || n.Width > 4096 {
			return &backend.CropTask{}, errors.New("wrong resize width detect")
		}
	}

	if captures["h"] == "" {
		n.Height = 0
	} else {
		n.Height, _ = strconv.Atoi(captures["h"])
		if n.Height < 1 || n.Height > 4096 {
			return &backend.CropTask{}, errors.New("wrong resize height detect")
		}
	}

	if captures["l"] == "" {
		n.Long = 0
	} else {
		n.Long, _ = strconv.Atoi(captures["l"])
		if n.Long < 1 || n.Long > 4096 {
			return &backend.CropTask{}, errors.New("wrong resize long detect")
		}
	}

	if captures["s"] == "" {
		n.Short = 0
	} else {
		n.Short, _ = strconv.Atoi(captures["s"])
		if n.Short < 1 || n.Short > 4096 {
			return &backend.CropTask{}, errors.New("wrong resize short detect")
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
			return &backend.CropTask{}, errors.New("wrong resize limit detect")
		}

	}

	n.Color = checkColor(captures["color"])

	switch captures["m"] {
	case "", "lfit":
		n.Mode = "lfit"
		if ((n.Width != 0 || n.Height != 0) && (n.Long != 0 || n.Short != 0)) == true {
			return &backend.CropTask{}, errors.New("can not resize in height&width and long&short at the same time")
		}
	case "mfit":
		n.Mode = "mfit"
		if ((n.Width != 0 || n.Height != 0) && (n.Long != 0 || n.Short != 0)) == true {
			return &backend.CropTask{}, errors.New("can not resize in height&width and long&short at the same time")
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
		return &backend.CropTask{}, errors.New("wrong resize mode detect")
	}
	return &n, nil
}

func (img *ImageWand) ResizeImage(fileName string, plan *backend.CropTask) error {
	helper.Logger.Println("start resize image, plan: ", plan)
	err := img.ReadImage(fileName)
	if err != nil {
		helper.Logger.Error("open temp file failed")
	}
	o := newOptions()
	o.Limit = plan.Limit
	o.Background = plan.Color

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
		o := Options{Width: plan.Width, Height: plan.Height}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	//短边优先
	case "mfit":
		adjustCropTask(plan, img.MagickWand.GetImageWidth(), img.MagickWand.GetImageHeight())
		o := Options{Width: plan.Width, Height: plan.Height}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	case "pad":
		o := Options{Width: plan.Width, Height: plan.Height, Pad: true}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	case "fixed":
		o := Options{Width: plan.Width, Height: plan.Height, Force: true}
		err = img.resize(o)
		if err != nil {
			return err
		}
		break
	case "fill":
		o := Options{Width: plan.Width, Height: plan.Height, Crop: true}
		err = img.resize(o)
		if err != nil {
			return err
		}
	}
	return nil
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
