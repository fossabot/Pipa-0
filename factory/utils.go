package factory

import (
	"errors"
	"fmt"
	"pipa/helper"
	"strings"
)

func parseUrl(startData StartTask) (buckerDomain, downloadUrl, convertParams string, err error) {
	//download content from data stripe all the query string and add "http://"
	uuid := startData.Uuid
	url := startData.Url
	helper.Logger.Info(fmt.Sprintf("I got task %s %s\n", uuid, url))

	var pos int

	if pos0 := strings.Index(url, "?x-oss-process=image/"); pos0 != -1 {
		pos = pos0
	}  else if pos1 := strings.Index(url, "?"); pos1 != -1 {
		pos = pos1
		return "", "", "", errors.New("can not found convert parameters")
	} else {
		helper.Logger.Info("can not found convert parameters")
		return  "", "", "", errors.New("can not found convert parameters")
	}

	//if remove any slash at the start
	var startPos int
	var v rune
	for startPos, v = range url[0:pos] {
		if string(v) != "/" {
			break
		}
	}

	buckerDomain = "http://" + strings.Split(url, "/")[2] + "/"
	downloadUrl = url[startPos:pos]
	convertParams = url[pos+len("?x-oss-process=image/"):]

	return buckerDomain, downloadUrl, convertParams, nil
}
