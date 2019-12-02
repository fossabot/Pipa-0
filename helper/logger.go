package helper

import (
	"pipa/log"
	"sync"
)
// Global singleton
var Wg sync.WaitGroup
var Logger log.Logger
