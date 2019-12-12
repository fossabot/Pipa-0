package factory

import (
	"strings"
)

const (
	RESIZE    = "resize"
	WATERMARK = "watermark"
	CROP      = "crop"
	ROTATE    = "rotate"
)

var resizeNames = []string{"m", "w", "h", "l", "s", "limit", "color", "p"}
var watermarkNames = []string{"t", "g", "x", "y", "voffset", "text", "limit", "color", "size", "shadow", "rotate", "fill", "image", "order", "align", "interval"}

func selectOperation(task string) (captures map[string]string, taskType string) {
	taskKeys := strings.Split(task, ",")
	switch taskKeys[0] {
	case RESIZE:
		captures = splitTaskKeys(resizeNames, taskKeys)
		taskType = RESIZE
		break
	case WATERMARK:
		captures = splitTaskKeys(watermarkNames, taskKeys)
		taskType = WATERMARK
		break
	default:
		return nil, ""
	}
	return captures, taskType
}

func watermarkPictureOperation(originFileName, task string) (captures map[string]string, taskType string) {
	captures = make(map[string]string)
	taskKeys := strings.Split(task, ",")
	switch taskKeys[0] {
	case RESIZE:
		for i := 1; i < len(taskKeys); i++ {
			params := strings.Split(taskKeys[i], "_")
			for _, name := range resizeNames {
				if params[0] == name {
					captures[name] = params[1]
				} else if params[0] == "P" {
					captures["P"] = params[1]
					captures["fileName"] = originFileName
				}
			}
			if len(captures) < i {
				return nil, ""
			}
		}
	case CROP:
	case ROTATE:
	default:
		return nil, ""
	}
	return captures, taskType
}

func splitTaskKeys(operationName []string, taskKeys []string) map[string]string {
	captures := make(map[string]string)
	for i := 1; i < len(taskKeys); i++ {
		params := strings.Split(taskKeys[i], "_")
		for _, name := range operationName {
			if params[0] == name {
				captures[name] = params[1]
			}
		}
		if len(captures) < i {
			return nil
		}
	}
	return captures
}
