package factory

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"pipa/backend"
	"pipa/helper"
	"pipa/imagick"
	"pipa/redis"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type StartTask struct {
	Uuid string `json:"uuid"`
	Url  string `json:"url"`
}

type TaskData struct {
	uuid         string
	url          string
	taskType     string
	bucketDomain string
	captures     map[string]string
	client       *http.Client
}

type FinishTask struct {
	code int
	uuid string
	url  string
	blob []byte
	mime string
}

var UNKNOWN string = "unknown"

func StartWork() {
	taskQ := make(chan string, 10)
	returnQ := make(chan FinishTask)

	httpClient := &http.Client{Timeout: time.Second * 5}

	helper.Wg.Add(helper.CONFIG.FactoryWorkersNumber + 1)
	for i := 0; i < helper.CONFIG.FactoryWorkersNumber; i++ {
		go slave(taskQ, returnQ, httpClient, i)
	}
	go reportFinish(returnQ)

	//will use signal channel to quit
	for {
		r, err := redis.Strings()
		if err != nil {
			helper.Logger.Info("something bad happened", err)
			time.Sleep(6 * time.Second)
			continue
		}
		helper.Logger.Println("Now have", r[1])
		taskQ <- r[1]
	}
}

func slave(taskQ chan string, resultQ chan FinishTask, client *http.Client, slave_num int) {

	defer helper.Wg.Done()
	for {
		pipa := PipaProcess{
			ImagePrecess: imagick.Initialize(),
			ResultQ:      resultQ,
		}
		defer pipa.processDone()
		task := <-taskQ
		//split url
		var startData StartTask
		taskData := TaskData{"", UNKNOWN, "", "", make(map[string]string), nil}
		dec := json.NewDecoder(strings.NewReader(task))
		if err := dec.Decode(&startData); err != nil {
			helper.Logger.Println("Decode failed")
			pipa.returnError(400, taskData)
			continue
		}

		domain, downloadUrl, convertParams, err := parseUrl(startData)
		if err != nil {
			continue
		}
		taskData.uuid = startData.Uuid
		taskData.url = startData.Url
		taskData.bucketDomain = domain
		taskData.client = client

		convertParamsSlice := strings.Split(convertParams, "/")

		var tempTasks []*backend.TempTask
		for _, task := range convertParamsSlice {
			taskData.captures, taskData.taskType = selectOperation(task)
			if taskData.captures == nil {
				helper.Logger.Error("some param wrong")
				break
			}

			tempTask, err := pipa.preprocessImage(taskData)
			if err != nil {
				helper.Logger.Error("wrong params: ", err)
				break
			}
			tempTasks = append(tempTasks, tempTask)
		}

		var retCode int
		originFileName, retCode := download(taskData.client, downloadUrl, taskData.uuid, "")
		if retCode != 200 {
			pipa.returnError(retCode, taskData)
			os.Remove(originFileName)
			continue
		}

		err = pipa.processImage(originFileName, tempTasks)
		if err != nil {
			pipa.returnUnchange(originFileName, 200, taskData)
			os.Remove(originFileName)
			helper.Logger.Error("Image process error: ", err)
			break
		}
		pipa.ResultQ <- FinishTask{200, startData.Uuid, startData.Url,
			pipa.ImagePrecess.GetImageBlob(), pipa.ImagePrecess.GetImageFormat()}
		os.Remove(originFileName)
	}
}

func download(client *http.Client, downloadUrl, uuid, localDir string) (string, int) {
	helper.Logger.Println(fmt.Sprintf("Start to download %s\n", downloadUrl))
	resp, err := client.Get(downloadUrl)
	if err != nil {
		helper.Logger.Println("Download failed!", err)
		return "", 404
	}
	defer resp.Body.Close()

	//check header
	if resp.StatusCode != 200 {
		helper.Logger.Println("Request is not 200")
		return "", resp.StatusCode
	}

	mimeType := resp.Header.Get("Content-Type")

	if strings.Contains(mimeType, "image") == false {
		if ok, _ := regexp.MatchString("(jpeg|jpg|png|gif|bmp|webp|tiff)", downloadUrl); ok == false {
			helper.Logger.Println(fmt.Sprintf("MIME TYPE is %s not an image\n", mimeType))
			return "", http.StatusUnsupportedMediaType //415
		}
	}

	contentLength := resp.Header.Get("Content-Length")

	if len, _ := strconv.Atoi(contentLength); len > (20 << 20) {
		return "", http.StatusRequestEntityTooLarge
	}

	/* open temp file */
	tmpfile, err := ioutil.TempFile(localDir, uuid)
	if err != nil {
		helper.Logger.Error("can not create temp file", uuid)
		return "", 404
	}
	defer tmpfile.Close()

	n, err := io.Copy(tmpfile, resp.Body)
	if err != nil {
		return "", 404
	}

	helper.Logger.Info(fmt.Sprintf("download %d bytes from %s OK\n", n, downloadUrl))
	return tmpfile.Name(), 200
}

func combineData(blob []byte, mime string) []byte {
	var a [20]byte
	copy(a[:], mime)
	return append(a[:], blob[:]...)
}

func reportFinish(resultQ chan FinishTask) {
	defer helper.Wg.Done()
	redisConn := redis.Pool.Get()
	defer redisConn.Close()
	for r := range resultQ {
		//put data back to redis
		if r.code == 200 {
			combined := combineData(r.blob, r.mime)
			redisConn.Do("MULTI")
			redisConn.Do("SET", r.url, combined)
			redisConn.Do("LPUSH", r.uuid, r.code)
			redisConn.Do("EXEC")
			r.blob = nil
		} else {
			redisConn.Do("LPUSH", r.uuid, r.code)
		}
		helper.Logger.Info(fmt.Sprintf("finishing task [%s] for %s code %d\n", r.uuid, r.url, r.code))
	}
}
