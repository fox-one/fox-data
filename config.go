package main

import (
	"encoding/base64"

	"github.com/fox-one/fox-data/session"
)

var (
	// CONFIGDATA config data, base64(config file)
	CONFIGDATA = ""
)

func newSession() (*session.Session, error) {
	data, err := base64.StdEncoding.DecodeString(CONFIGDATA)
	if err != nil {
		return nil, err
	}

	return session.New(data)
}
