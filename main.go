package main

import (
	"os"
	"os/signal"
	"pipa/factory"
	"pipa/helper"
	"pipa/log"
	"pipa/redis"
	"runtime"
	"syscall"
)


func DumpStacks() {
	buf := make([]byte, 1<<16)
	stackLen := runtime.Stack(buf, true)
	helper.Logger.Error("Received SIGQUIT, goroutine dump:")
	helper.Logger.Error(buf[:stackLen])
	helper.Logger.Error("*** dump end")
}

func main() {

	helper.SetupConfig()

	//pipa log
	logLevel := log.ParseLevel(helper.CONFIG.LogLevel)
	helper.Logger = log.NewFileLogger(helper.CONFIG.LogPath, logLevel)

	redis.Initialize()
	defer redis.Close()

	factory.StartWork()

	// ignore signal handlers set by Iris
	signal.Ignore()
	signalQueue := make(chan os.Signal)
	signal.Notify(signalQueue, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR1)
	for {
		s := <-signalQueue
		switch s {
		case syscall.SIGHUP:
			// reload config file
			helper.SetupConfig()
		case syscall.SIGUSR1:
			go DumpStacks()
		default:
			// stop pipa server, order matters
			helper.Logger.Info("pipa Server stopped")
			helper.Wg.Wait()
			helper.Logger.Info("done")
			return
		}
	}

}
