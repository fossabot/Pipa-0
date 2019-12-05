package factory

import (
	"os"
	"pipa/backend"
	"pipa/helper"
)

type PipaProcess struct {
	ImagePrecess   backend.ImageProcess
	OriginFileName string
	ResultQ        chan FinishTask
}

func (pipa *PipaProcess) returnError(code int, uuid string, url string) {
	pipa.ResultQ <- FinishTask{code, uuid, url, nil, ""}
}

func (pipa *PipaProcess) returnUnchange(code int, uuid string, url string) {
	err := pipa.ImagePrecess.ReadImage(pipa.OriginFileName)
	if err != nil {
		helper.Logger.Error("open temp file failed")
	}
	pipa.ResultQ <- FinishTask{code, uuid, url, pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}
}

func (pipa *PipaProcess) processImage(taskType string, captures map[string]string, uuid string, url string) error {

	switch taskType {
	case RESIZE:
		plan, err := pipa.ImagePrecess.ResizePreprocess(captures)
		if err != nil {
			pipa.returnUnchange(200, uuid, url)
			os.Remove(pipa.OriginFileName)
			return err
		}

		err = pipa.ImagePrecess.ResizeImage(pipa.OriginFileName, plan)
		if err != nil {
			pipa.returnUnchange(200, uuid, url)
			os.Remove(pipa.OriginFileName)
		}
		os.Remove(pipa.OriginFileName)

		pipa.ResultQ <- FinishTask{200, uuid, url,
			pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}

	case WATERMARK:

	}
}

func (pipa *PipaProcess) processDone() {
	pipa.ImagePrecess.Terminate()
}
