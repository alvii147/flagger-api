package testkit

import (
	"fmt"
	"regexp"
	"time"
)

func MustParseLogMessage(msg string) (string, time.Time, string, string) {
	r := regexp.MustCompile(`^\s*\[([DIWE])\]\s+(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(.+)\s*$`)
	matches := r.FindStringSubmatch(msg)

	if len(matches) != 5 {
		panic("MustParseLogMessage failed to get exactly five matches on r.FindStringSubmatch")
	}

	logLevel := matches[1]
	logTime, err := time.ParseInLocation("2006/01/02 15:04:05", matches[2], time.UTC)
	if err != nil {
		panic(fmt.Sprintf("MustParseLogMessage failed to time.ParseInLocation %s", matches[2]))
	}

	logFile := matches[3]
	logMsg := matches[4]

	return logLevel, logTime, logFile, logMsg
}
