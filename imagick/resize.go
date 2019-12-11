package imagick

import (
	"math"
	"pipa/helper"
)

type Resize struct {
	Width      int
	Height     int
	Zoom       float64
	Force      bool
	Crop       bool
	Pad        bool
	Limit      bool
	Background string
}

func (img *ImageWand) resize(o Resize) (err error) {
	originWidth := int(img.MagickWand.GetImageWidth())
	originHeight := int(img.MagickWand.GetImageHeight())

	// image calculations
	factor := imageCalculations(&o, originWidth, originHeight)

	if o.Limit && !o.Force {
		if originWidth < o.Width && originHeight < o.Height {
			factor = 1.0
			o.Width = originWidth
			o.Height = originHeight
		}
	}
	switch {
	case o.Zoom != Zoom:
		err = img.MagickWand.ResizeImage(uint(float64(originWidth)*o.Zoom), uint(float64(originHeight)*o.Zoom), Method)
		if err != nil {
			helper.Logger.Error("MagickWand resize image failed... err:", err)
			return err
		}
		break
	case o.Crop == true:
		err = img.MagickWand.ResizeImage(uint(float64(originWidth)*factor), uint(float64(originHeight)*factor), Method)
		if err != nil {
			helper.Logger.Error("MagickWand resize image failed... err:", err)
			return err
		}
		err = img.cropImage(o, originWidth, originHeight)
		if err != nil {
			return err
		}
		break
	case o.Pad == true:
		err = img.MagickWand.ResizeImage(uint(float64(originWidth)*factor), uint(float64(originHeight)*factor), Method)
		if err != nil {
			helper.Logger.Error("MagickWand resize image failed... err:", err)
			return err
		}
		err = img.extentImage(o, originWidth, originHeight)
		if err != nil {
			return err
		}
		break
	default:
		err = img.MagickWand.ResizeImage(uint(float64(originWidth)*factor), uint(float64(originHeight)*factor), Method)
		if err != nil {
			helper.Logger.Error("MagickWand resize image failed... err:", err)
			return err
		}
	}
	return nil
}

func newResize() Resize {
	return Resize{0, 0, Zoom, Force, Crop, Pad, Limit, Background}
}

func imageCalculations(o *Resize, inWidth, inHeight int) float64 {
	factor := 1.0
	hFactor := float64(o.Width) / float64(inWidth)
	wFactor := float64(o.Height) / float64(inHeight)

	switch {
	case o.Width > 0 && o.Height > 0:
		if o.Crop {
			factor = math.Max(hFactor, wFactor)
		} else {
			factor = math.Min(hFactor, wFactor)
		}
	case o.Width > 0:
		factor = wFactor
	case o.Height > 0:
		factor = hFactor
	// Identity transform
	default:
		o.Width = inWidth
		o.Height = inHeight
		break
	}

	return factor
}

func (img *ImageWand) cropImage(o Resize, originWidth, originHeight int) error {
	offsetWidth := math.Abs(float64(originWidth-o.Width) / 2)
	offsetHeight := math.Abs(float64(originHeight-o.Height) / 2)
	err := img.MagickWand.CropImage(uint(o.Width), uint(o.Height), int(offsetWidth), int(offsetHeight))
	if err != nil {
		helper.Logger.Error("MagickWand resize crop image... err:", err)
		return err
	}
	return nil
}

func (img *ImageWand) extentImage(o Resize, originWidth, originHeight int) error {
	img.PixelWand.SetColor(o.Background)
	err := img.MagickWand.SetImageBackgroundColor(img.PixelWand)
	if err != nil {
		helper.Logger.Error("MagickWand set image background failed... err:", err)
		return err
	}
	offsetWidth := math.Abs(float64(originWidth-o.Width) / 2)
	offsetHeight := math.Abs(float64(originHeight-o.Height) / 2)
	err = img.MagickWand.ExtentImage(uint(o.Width), uint(o.Height), -int(offsetWidth), -int(offsetHeight))
	if err != nil {
		helper.Logger.Error("MagickWand extent image failed... err:", err)
		return err
	}
	return nil
}
