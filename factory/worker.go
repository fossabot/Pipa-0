package factory

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"pipa/helper"
	"pipa/redis"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TaskData struct {
	Uuid string `json:"uuid"`
	Url  string `json:"url"`
}

type FinishTask struct {
	code int
	uuid string
	url  string
	blob []byte
	mime string
}

var UNKNOWN string = "unknown"

func returnError(code int, uuid string, url string, Q chan FinishTask) {
	Q <- FinishTask{code, uuid, url, nil, ""}
}

func returnUnchange(code int, uuid string, url string, filePath string, Q chan FinishTask) {
	buffer, _ := bimg.Read(filePath)
	img := bimg.NewImage(buffer)
	Q <- FinishTask{code, uuid, url, buffer, img.Type()}
}

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
		helper.Logger.Println("####r= ", r)
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
		task := <-taskQ
		//split url
		var data TaskData
		dec := json.NewDecoder(strings.NewReader(task))
		if err := dec.Decode(&data); err != nil {
			helper.Logger.Println("Decode failed")
			returnError(400, UNKNOWN, "", resultQ)
			continue
		}
		uuid := data.Uuid
		url := data.Url
		helper.Logger.Info("I got task %s %s\n", uuid, url)

		//download content from data stripe all the query string and add "http://"
		taskType := ""
		var pos int
		var pType int

		if pos0 := strings.Index(url, "?x-oss-process=image/"); pos0 != -1 {
			pos = pos0
			pType = 0
		} else if pos1 := strings.Index(url, "?x-oss-process=style/"); pos1 != -1 {
			pos = pos1
			pType = 1
		} else if pos2 := strings.Index(url, "?"); pos2 != -1 {
			pos = pos2
			pType = 2
		} else {
			helper.Logger.Info("can not found convert parameters")
			returnError(400, uuid, url, resultQ)
			continue
		}

		//if remove any slash at the start
		var startPos int
		var v rune
		for startPos, v = range url[0:pos] {
			if string(v) != "/" {
				break
			}
		}

		downloadUrl := "http://" + url[startPos:pos]
		convertParams := url[pos+len("?x-oss-process=image/"):]
		convertParamsSlice := strings.Split(convertParams, "/")

		originFileName, retCode := download(client, downloadUrl, uuid)
		if retCode != 200 {
			returnError(retCode, uuid, url, resultQ)
			os.Remove(originFileName)
			continue
		}

		if pType == 1 || pType == 2 {
			returnUnchange(200, uuid, url, originFileName, resultQ)
			os.Remove(originFileName)
			continue
		}

		for _, task := range convertParamsSlice {
			var r []string
			var names []string

			r, names, taskType = selectOperation(task, convertParams)
			if taskType == "" {
				continue
			}

			captures := make(map[string]string)
			for i, name := range names {
				if i == 0 {
					continue
				}
				helper.Logger.Info("name: %s, %s", name, r[i])
				splited := strings.Split(r[i], "_")
				if len(splited) < 2 {
					captures[name] = ""
				} else {
					captures[name] = splited[1]
				}
			}


		}
	}

}

func download(client *http.Client, downloadUrl string, uuid string) (string, int) {
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
	tmpfile, err := ioutil.TempFile("", uuid)
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
