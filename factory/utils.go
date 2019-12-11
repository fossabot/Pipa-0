package factory

import (
	"errors"
	"pipa/helper"
	"strings"
)

func parseUrl(startData StartTask) (pType int, buckerDomain, downloadUrl, convertParams string, err error) {
	//download content from data stripe all the query string and add "http://"
	uuid := startData.Uuid
	url := startData.Url
	helper.Logger.Info("I got task %s %s\n", uuid, url)

	var pos int

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
		return 3, "", "", "", errors.New("can not found convert parameters")
	}

	//if remove any slash at the start
	var startPos int
	var v rune
	for startPos, v = range url[0:pos] {
		if string(v) != "/" {
			break
		}
	}

	buckerDomain = strings.Split(url, "/")[0]
	downloadUrl = "http://" + url[startPos:pos]
	convertParams = url[pos+len("?x-oss-process=image/"):]

	return pType, buckerDomain, downloadUrl, convertParams, nil
}
