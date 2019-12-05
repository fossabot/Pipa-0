package factory

import (
	"regexp"
	"strings"
)

const (
	RESIZE    = "resize"
	WATERMARK = "watermark"
)

var resizePattern = regexp.MustCompile("resize(?P<m>,m_[a-z]+)?(?P<w>,w_[0-9]+)?(?P<h>,h_[0-9]+)?(?P<l>,l_[0-9]+)?(?P<s>,s_[0-9]+)?(?P<limit>,limit_[0-1])?(?P<color>,color_[0-9a-fA-F,]+)?(?P<p>,p_[0-9]+)?")
var watermarkPattern = regexp.MustCompile("watermark(?P<t>,t_[0-9]+)?(?P<g>,g_[a-z]+)?(?P<x>,x_[0-9]+)?(?P<y>,y_[0-9]+)?(?P<voffset>,voffset_[-0-9]+)?(?P<text>,text_[a-zA-Z0-9-_=]+)?(?P<type>,type_[a-zA-Z0-9-_=]+)?(?P<color>,color_[0-9a-fA-F,]+)?(?P<size>,size_[0-9]+)?(?P<shadow>,shadow_[0-9]+)?(?P<rotate>,rotate_[0-9]+)?(?P<fill>,fill_[0-1])?(?P<image>,image_[a-zA-Z0-9-_=]+)?")
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
		return nil,nil,""
	}

	return r,names,taskType
}
