package imagick

import (
	"gopkg.in/gographics/imagick.v3/imagick"
	"pipa/backend"
	"strconv"
	"strings"
)

const (
	//Resize default param
	Zoom       = 0.0
	Force      = false
	Crop       = false
	Pad        = false
	Limit      = true
	Background = "#FFFFFF"
	Method     = imagick.FILTER_POINT
	//Watermark default param
	XMargin      = 10
	YMargin      = 10
	GRAVITY      = imagick.GRAVITY_SOUTH_EAST
	Transparency = 100
	FrontSize    = 40.0
)

const (
	NorthWest = "nw"
	North     = "north"
	NorthEast = "ne"
	West      = "west"
	Center    = "center"
	East      = "east"
	SouthWest = "sw"
	South     = "south"
	SouthEast = "se"
)

//Text type
const (
	DefaultTextType   = "wqy-zenhei"
	WQYZhenHei        = "wqy-zenhei"
	WQYMicroHei       = "wqy-microhei"
	FangZhengShuoSong = "fangzhengshusong"
	FangZhengKaiTi    = "fangzhengkaiti"
	FangZhengHeiTi    = "fangzhengheiti"
	FangZhengFangSong = "fangzhengfangsong"
	DroidSansFallBack = "droidsansfallback"
)

func adjustCropTask(plan *backend.ResizeTask, width, height uint) {
	//单宽高缩放
	if plan.Width+plan.Height != 0 && plan.Width*plan.Height == 0 {
		return
	}
	//单长短边缩放
	if plan.Long+plan.Short != 0 && plan.Long*plan.Short == 0 {
		if plan.Long != 0 {
			if width >= height {
				plan.Width = plan.Long
				plan.Height = 0
			} else {
				plan.Height = plan.Long
				plan.Width = 0
			}
		} else {
			if width >= height {
				plan.Height = plan.Short
				plan.Width = 0
			} else {
				plan.Width = plan.Short
				plan.Height = 0
			}
		}
		return
	}

	//同时指定宽高缩放
	if plan.Width > 0 && plan.Height > 0 {
		if plan.Mode == "lfit" { //长边优先
			if width >= height {
				plan.Height = 0
			} else {
				plan.Width = 0
			}
		}

		if plan.Mode == "mfit" { //短边优先
			if width >= height {
				plan.Width = 0
			} else {
				plan.Height = 0
			}
		}
		return
	}

	//同时指定长短边缩放
	if plan.Long > 0 && plan.Short > 0 {
		if plan.Mode == "lfit" { //长边优先
			if width >= height {
				plan.Width = plan.Long
				plan.Height = 0
			} else {
				plan.Height = plan.Long
				plan.Width = 0
			}
		}

		if plan.Mode == "mfit" { //短边优先
			if width >= height {
				plan.Height = plan.Short
				plan.Width = 0
			} else {
				plan.Width = plan.Short
				plan.Height = 0
			}
		}
		return
	}
	return
}

func checkColor(color string) string {
	switch {
	case strings.Contains(color, ","):
		rgb := []int{255, 255, 255}
		colors := strings.Split(color, ",")
		for i, num := range colors {
			n, err := strconv.Atoi(num)
			if err != nil {
				break
			}
			if n > 255 {
				continue
			}
			rgb[i] = n
		}
		return "rgb(" + string(rgb[0]) + "," + string(rgb[1]) + "," + string(rgb[2]) + ")"
	case len(color) == 6:
		return "#" + color
	default:
		return Background
	}
}

func selectTextType(tType string) string {
	switch tType {
	case WQYZhenHei:
		return "WQYZH.ttf"
	case WQYMicroHei:
		return "WQYWMH.ttf"
	case FangZhengShuoSong:
		return "FZSSJW.TTF"
	case FangZhengKaiTi:
		return "FZKTJW.TTF"
	case FangZhengHeiTi:
		return "FZHTJW.TTF"
	case FangZhengFangSong:
		return "FZFSJW.TTF"
	case DroidSansFallBack:
		return "DroidSansFallBack.ttf"
	default:
		return "WQYZH.ttf"
	}
}
