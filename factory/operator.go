package factory

import (
	"errors"
	"os"
	"pipa/backend"
	"pipa/helper"
	"strings"
)

type PipaProcess struct {
	ImagePrecess backend.ImageProcess
	ResultQ      chan FinishTask
}

func (pipa *PipaProcess) returnError(code int, taskData TaskData) {
	pipa.ResultQ <- FinishTask{code, taskData.uuid, taskData.url, nil, ""}
}

func (pipa *PipaProcess) returnUnchange(originFileName string, code int, taskData TaskData) {
	err := pipa.ImagePrecess.ReadImage(originFileName)
	if err != nil {
		helper.Logger.Error("open temp file failed")
	}
	pipa.ResultQ <- FinishTask{code, taskData.uuid, taskData.uuid, pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}
}

func (pipa *PipaProcess) preprocessImage(taskData TaskData) (tempTask *backend.TempTask, err error) {

	tempTask = &backend.TempTask{}
	switch taskData.taskType {
	case RESIZE:
		tempTask.TaskType = RESIZE
		tempTask.ResizeTask, err = pipa.ImagePrecess.ResizePreprocess(taskData.captures)
		if err != nil {
			return tempTask, err
		}
		return tempTask, nil
	case WATERMARK:
		tempTask.TaskType = WATERMARK
		tempTask.WatermarkTask, err = pipa.ImagePrecess.WatermarkPreprocess(taskData.captures)
		if err != nil {
			return tempTask, err
		}

		if tempTask.WatermarkTask.PictureMask.Image != "" {
			watermarkStartData := StartTask{taskData.uuid + "watermark", taskData.bucketDomain + tempTask.WatermarkTask.PictureMask.Image}
			domain, downloadUrl, convertParams, err := parseUrl(watermarkStartData)
			if err != nil {
				return tempTask, err
			}

			watermarkTaskData := TaskData{taskData.uuid, taskData.bucketDomain + tempTask.WatermarkTask.PictureMask.Image,
				"", domain, make(map[string]string), taskData.client}

			convertParamsSlice := strings.Split(convertParams, "/")

			var watermarkResizeTasks []*backend.TempTask
			for _, task := range convertParamsSlice {
				taskData.captures, taskData.taskType = watermarkPictureOperation(task)
				if taskData.captures == nil {
					helper.Logger.Error("some param wrong")
					break
				}

				tempTask, err := pipa.preprocessImage(taskData)
				if err != nil {
					helper.Logger.Error("wrong params: ", err)
					break
				}
				watermarkResizeTasks = append(watermarkResizeTasks, tempTask)
			}

			var retCode int
			tempTask.WatermarkTask.PictureMask.FileName, retCode = download(watermarkTaskData.client, downloadUrl, watermarkTaskData.uuid, "")
			if retCode != 200 {
				os.Remove(tempTask.WatermarkTask.PictureMask.FileName)
				return tempTask, errors.New("watermark picture download failed")
			}
			tempTask.WatermarkTask.PictureMask.WatermarkPictureTasks = watermarkResizeTasks

		} else if tempTask.WatermarkTask.TextMask.Text != "" {
			return tempTask, err
		} else {
			return tempTask, errors.New("watermark wrong params")
		}
		return tempTask, nil
	default:
		return tempTask, errors.New("image is not processed ")
	}
}

func (pipa *PipaProcess) processImage(originFileName string, tempTasks []*backend.TempTask) error {

	for _, tempTask := range tempTasks {
		switch tempTask.TaskType {
		case RESIZE:
			err := pipa.ImagePrecess.ResizeImage(originFileName, tempTask.ResizeTask)
			if err != nil {
				return err
			}
			err = pipa.ImagePrecess.WriteImage(originFileName)
			if err != nil {
				return err
			}
			break
		case WATERMARK:
			if tempTask.WatermarkTask.PictureMask.Image != "" {
				for _, watermarkPictureTask := range tempTask.WatermarkTask.PictureMask.WatermarkPictureTasks {
					watermarkPictureTask.ResizeTask.FileName = originFileName
					switch watermarkPictureTask.TaskType {
					case RESIZE:
						err := pipa.ImagePrecess.ResizeImage(tempTask.WatermarkTask.PictureMask.FileName, watermarkPictureTask.ResizeTask)
						if err != nil {
							return err
						}
						err = pipa.ImagePrecess.WriteImage(tempTask.WatermarkTask.PictureMask.FileName)
						if err != nil {
							return err
						}
						break
					case CROP:
					case ROTATE:
					default:
						helper.Logger.Info("watermark picture is not processed")
					}
				}
			}
			err := pipa.ImagePrecess.ImageWatermark(originFileName, tempTask.WatermarkTask)
			if err != nil {
				return err
			}
			err = pipa.ImagePrecess.WriteImage(originFileName)
			if err != nil {
				return err
			}
			break
		default:
			helper.Logger.Info("image is not processed")
			//return errors.New("image is not processed ")
		}
	}
	return nil
}

func (pipa *PipaProcess) processDone() {
	pipa.ImagePrecess.Terminate()
}
