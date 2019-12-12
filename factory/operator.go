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

func (pipa *PipaProcess) processImage(originFileName string, taskData TaskData) error {

	switch taskData.taskType {
	case RESIZE:
		plan, err := pipa.ImagePrecess.ResizePreprocess(taskData.captures)
		if err != nil {
			return err
		}

		err = pipa.ImagePrecess.ResizeImage(originFileName, plan)
		if err != nil {
			return err
		}

		pipa.ResultQ <- FinishTask{200, taskData.uuid, taskData.url,
			pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}
		return nil

	case WATERMARK:
		plan, err := pipa.ImagePrecess.WatermarkPreprocess(taskData.captures)
		if err != nil {
			return err
		}
		if plan.PictureMask.Image != "" {
			watermarkStartData := StartTask{taskData.uuid + "watermark", taskData.bucketDomain + plan.PictureMask.Image}
			domain, downloadUrl, convertParams, err := parseUrl(watermarkStartData)
			if err != nil {
				return err
			}

			watermarkTaskData := TaskData{taskData.uuid, taskData.bucketDomain + plan.PictureMask.Image,
				"", domain, make(map[string]string), taskData.client}

			var retCode int
			convertParamsSlice := strings.Split(convertParams, "/")

			plan.PictureMask.Filename, retCode = download(watermarkTaskData.client, downloadUrl, watermarkTaskData.uuid, "")
			if retCode != 200 {
				os.Remove(plan.PictureMask.Filename)
				return errors.New("watermark picture download failed")
			}

			for _, task := range convertParamsSlice {
				watermarkTaskData.captures, watermarkTaskData.taskType = watermarkPictureOperation(originFileName, task)
				if len(watermarkTaskData.captures) == 0 {
					os.Remove(plan.PictureMask.Filename)
					return errors.New("watermark picture's param is wrong")
				}
				err := pipa.processImage(plan.PictureMask.Filename, watermarkTaskData)
				if err != nil {
					os.Remove(plan.PictureMask.Filename)
					return errors.New("watermark picture process failed")
				}
				err = pipa.ImagePrecess.WriteImage(plan.PictureMask.Filename)
			}
			err = pipa.ImagePrecess.ImageWatermark(originFileName, plan)
			if err != nil {
				return err
			}
		} else if plan.TextMask.Text != "" {
			err := pipa.ImagePrecess.ImageWatermark(originFileName, plan)
			if err != nil {
				return err
			}
		} else {
			return errors.New("watermark add failed")
		}
		pipa.ResultQ <- FinishTask{200, taskData.uuid, taskData.url,
			pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}
		return nil
	default:
		return errors.New("image is not processed ")
	}
}

func (pipa *PipaProcess) processDone() {
	pipa.ImagePrecess.Terminate()
}
