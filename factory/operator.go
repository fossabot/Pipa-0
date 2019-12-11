package factory

import (
	"errors"
	"os"
	"pipa/backend"
	"pipa/helper"
	"strings"
)

type PipaProcess struct {
	ImagePrecess   backend.ImageProcess
	OriginFileName string
	ResultQ        chan FinishTask
}

func (pipa *PipaProcess) returnError(code int, taskData TaskData) {
	pipa.ResultQ <- FinishTask{code, taskData.uuid, taskData.url, nil, ""}
}

func (pipa *PipaProcess) returnUnchange(code int, taskData TaskData) {
	err := pipa.ImagePrecess.ReadImage(pipa.OriginFileName)
	if err != nil {
		helper.Logger.Error("open temp file failed")
	}
	pipa.ResultQ <- FinishTask{code, taskData.uuid, taskData.uuid, pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}
}

func (pipa *PipaProcess) processImage(taskData TaskData) error {

	switch taskData.taskType {
	case RESIZE:
		plan, err := pipa.ImagePrecess.ResizePreprocess(taskData.captures)
		if err != nil {
			return err
		}

		err = pipa.ImagePrecess.ResizeImage(pipa.OriginFileName, plan)
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
			watermarkData := StartTask{taskData.uuid + "watermark", taskData.buckerDomain + plan.PictureMask.Image}

			pType, domain, downloadUrl, convertParams, err := parseUrl(watermarkData)
			if err != nil {
				return err
			}

			taskData.buckerDomain = domain
			var retCode int
			convertParamsSlice := strings.Split(convertParams, "/")

			plan.PictureMask.Filename, retCode = download(taskData.client, downloadUrl, watermarkData.Uuid, "")
			if retCode != 200 {
				os.Remove(plan.PictureMask.Filename)
				return errors.New("watermark picture download failed")
			}
			if pType == 1 || pType == 2 {
				os.Remove(plan.PictureMask.Filename)
				return errors.New("watermark picture param is wrong")
			}

			for _, task := range convertParamsSlice {
				watermarkPictureOperation(plan, task)

				err := pipa.ImagePrecess.ImageWatermark(pipa.OriginFileName, plan)
				if err != nil {
					continue
				}
			}
		} else if plan.TextMask.Text != "" {
			err := pipa.ImagePrecess.ImageWatermark(pipa.OriginFileName, plan)
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
