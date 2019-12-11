package factory

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	taskType     string
	uuid         string
	url          string
	buckerDomain string
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

	for i := 0; i < helper.CONFIG.FactoryWorkersNumber; i++ {
		go slave(taskQ, returnQ, httpClient, i)
	}
	go reportFinish(returnQ)

	//will use signal channel to quit
	for {
		r, err := redis.Strings()
		if err != nil {
			helper.Logger.Info("something bad happend %v", err)
			return
		}
		helper.Logger.Println("Now have", r[1])
		taskQ <- r[1]
	}
}

func slave(taskQ chan string, resultQ chan FinishTask, client *http.Client, slave_num int) {

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

		pType, domain, downloadUrl, convertParams, err := parseUrl(startData)
		if err != nil {
			continue
		}

		taskData.buckerDomain = domain
		var retCode int
		convertParamsSlice := strings.Split(convertParams, "/")

		pipa.OriginFileName, retCode = download(client, downloadUrl, startData.Uuid, "")
		if retCode != 200 {
			pipa.returnError(retCode, taskData)
			os.Remove(pipa.OriginFileName)
			continue
		}
		if pType == 1 || pType == 2 {
			pipa.returnUnchange(200, taskData)
			os.Remove(pipa.OriginFileName)
			continue
		}

		for _, task := range convertParamsSlice {
			var r []string
			var names []string

			r, names, taskData.taskType = selectOperation(task, convertParams)
			if taskData.taskType == "" {
				continue
			}

			for i, name := range names {
				if i == 0 {
					continue
				}
				helper.Logger.Info("name: %s, %s", name, r[i])
				splited := strings.Split(r[i], "_")
				if len(splited) < 2 {
					taskData.captures[name] = ""
				} else {
					taskData.captures[name] = splited[1]
				}
			}

			err := pipa.processImage(taskData)
			if err != nil {
				pipa.returnUnchange(200, taskData)
				os.Remove(pipa.OriginFileName)
				helper.Logger.Error("Image process error: ",err)
				continue
			}
			os.Remove(pipa.OriginFileName)
		}
	}
}

func download(client *http.Client, downloadUrl, uuid, localDir string) (string, int) {
	helper.Logger.Info("Start to download %s\n", downloadUrl)
	resp, err := client.Get(downloadUrl)
	if err != nil {
		helper.Logger.Println("Download failed!")
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
		if ok, _ := regexp.MatchString("(jpeg|jpg|png|gif)", downloadUrl); ok == false {
			helper.Logger.Println("MIME TYPE is %s not an image\n", mimeType)
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
		helper.Logger.Info("can not create temp file %s", uuid)
		return "", 404
	}
	defer tmpfile.Close()

	n, err := io.Copy(tmpfile, resp.Body)
	if err != nil {
		return "", 404
	}

	helper.Logger.Info("download %d bytes from %s OK\n", n, downloadUrl)
	return tmpfile.Name(), 200
}

func combineData(blob []byte, mime string) []byte {
	var a [20]byte
	copy(a[:], mime)
	return append(a[:], blob[:]...)
}

func reportFinish(resultQ chan FinishTask) {
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
		helper.Logger.Println("finishing task [%s] for %s code %d\n", r.uuid, r.url, r.code)
	}
}
