package imagick

import (
	"pipa/backend"
	"strconv"
	"strings"
)

func adjustCropTask(plan *backend.CropTask, width, height uint) {
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
		return "#FFFFFF"
	}
}
