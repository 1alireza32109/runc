package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/logs"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/sirupsen/logrus"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		// This is the golang entry point for runc init, executed
		// before main() but after libcontainer/nsenter's nsexec().
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()

		level := os.Getenv("_LIBCONTAINER_LOGLEVEL")
		logLevel, err := logrus.ParseLevel(level)
		if err != nil {
			panic(fmt.Sprintf("libcontainer: failed to parse log level: %q: %v", level, err))
		}

		logPipeFdStr := os.Getenv("_LIBCONTAINER_LOGPIPE")
		logPipeFd, err := strconv.Atoi(logPipeFdStr)
		if err != nil {
			panic(fmt.Sprintf("libcontainer: failed to convert environment variable _LIBCONTAINER_LOGPIPE=%s to int: %s", logPipeFdStr, err))
		}
		err = logs.ConfigureLogging(logs.Config{
			LogPipeFd: logPipeFd,
			LogFormat: "json",
			LogLevel:  logLevel,
		})
		if err != nil {
			panic(fmt.Sprintf("libcontainer: failed to configure logging: %v", err))
		}
		logrus.Debug("child process in init()")

		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			// as the error is sent back to the parent there is no need to log
			// or write it to stderr because the parent process will handle this
			os.Exit(1)
		}
		panic("libcontainer: container init failed to exec")
	}
}
