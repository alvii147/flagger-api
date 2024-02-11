package testkit

import (
	"fmt"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

func MustParseMailMessage(msg string) (string, string) {
	mailMsg, err := mail.ReadMessage(strings.NewReader(msg))
	if err != nil {
		panic(fmt.Sprintf("MustParseMailMessage failed to mail.ReadMessage: %v", err))
	}

	r := regexp.MustCompile(`^multipart/alternative;\s*boundary="(\S+)"$`)
	matches := r.FindStringSubmatch(mailMsg.Header.Get("Content-Type"))
	if len(matches) != 2 {
		panic("MustParseMailMessage failed to find content type")
	}

	boundary := matches[1]

	msgBytes, err := io.ReadAll(mailMsg.Body)
	if err != nil {
		panic(fmt.Sprintf("MustParseMailMessage failed to io.ReadAll: %v", err))
	}

	r = regexp.MustCompile(`--+` + boundary + `-*`)
	mailSections := r.Split(string(msgBytes), -1)

	nonEmptyMailSections := []string{}
	for _, sec := range mailSections {
		if utf8.RuneCountInString(strings.TrimSpace(sec)) != 0 {
			nonEmptyMailSections = append(nonEmptyMailSections, sec)
		}
	}

	if len(nonEmptyMailSections) != 2 {
		panic("MustParseMailMessage failed to parse text and html sections")
	}

	textMsg := nonEmptyMailSections[0]
	htmlMsg := nonEmptyMailSections[1]

	return textMsg, htmlMsg
}
