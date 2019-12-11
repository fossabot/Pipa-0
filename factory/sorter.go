package factory

import (
	"pipa/backend"
	"regexp"
	"strconv"
	"strings"
)

const (
	RESIZE    = "resize"
	WATERMARK = "watermark"
	CROP      = "crop"
	ROTATE    = "rotate"
)

var resizePattern = regexp.MustCompile("resize(?P<m>,m_[a-z]+)?(?P<w>,w_[0-9]+)?(?P<h>,h_[0-9]+)?(?P<l>,l_[0-9]+)?(?P<s>,s_[0-9]+)?(?P<limit>,limit_[0-1])?(?P<color>,color_[0-9a-fA-F,]+)?(?P<p>,p_[0-9]+)?")
var watermarkPattern = regexp.MustCompile("watermark(?P<t>,t_[0-9]+)?(?P<g>,g_[a-z]+)?(?P<x>,x_[0-9]+)?(?P<y>,y_[0-9]+)?(?P<voffset>,voffset_[-0-9]+)?(?P<text>,text_[a-zA-Z0-9-_=]+)?(?P<type>,type_[a-zA-Z0-9-_=]+)?(?P<color>,color_[0-9a-fA-F,]+)?(?P<size>,size_[0-9]+)?(?P<shadow>,shadow_[0-9]+)?(?P<rotate>,rotate_[0-9]+)?(?P<fill>,fill_[0-1])?(?P<image>,image_[a-zA-Z0-9-_=]+)?(?P<order>,order_[0-1])?(?P<align>,align_[0-2])?(?P<interval>,interval_[1-9]+)?")
var resizeNames = resizePattern.SubexpNames()
var watermarkNames = watermarkPattern.SubexpNames()

func selectOperation(task, convertParams string) (r []string, names []string, taskType string) {
	taskKeys := strings.Split(task, ",")
	switch taskKeys[0] {
	case RESIZE:
		r = resizePattern.FindStringSubmatch(convertParams)
		names = resizeNames
		taskType = RESIZE
	case WATERMARK:
		r = watermarkPattern.FindStringSubmatch(convertParams)
		names = watermarkNames
		taskType = WATERMARK
	default:
		return nil, nil, ""
	}

	return r, names, taskType
}

func watermarkPictureOperation(plan *backend.WatermarkTask ,task string) {

	taskKeys := strings.Split(task, ",")
	switch taskKeys[0] {
	case RESIZE:
		splited := strings.Split(taskKeys[1], "_")
		if splited[0] == "P" {
			num, _ := strconv.Atoi(splited[1])
			if num > 0 && num < 101 {
				plan.PictureMask.Proportion = num
			} else {
				return
			}
		} else {
			return
		}
	case CROP:

	case ROTATE:

	default:

	}
}
