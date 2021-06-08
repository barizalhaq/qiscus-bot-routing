package logger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func InitLogger(path, prefix string) *log.Logger {
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s/%s.log", path, "%Y-%m-%d"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithClock(rotatelogs.Local),
	)
	if err != nil {
		panic(err)
	}

	return log.New(writer, prefix, log.LstdFlags)
}

func WriteOutbondLog(logger *log.Logger, resp *http.Response, respBody, reqBody string) {
	logger.Printf("REQUEST: %s %s Header: %s Payload: %s\n", resp.Request.Method, resp.Request.URL, resp.Request.Header, reqBody)
	logger.Printf("RESPONSE: %d %s", resp.StatusCode, respBody)
}