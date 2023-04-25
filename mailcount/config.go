package main

import (
	"bufio"
	"fmt"
	"os"

	str "strings"
)

func (s *sessionT) parseConfig(file string) {
	fd, err := os.Open(file)
	errExit(err)
	defer fd.Close()

	var section string

	input := bufio.NewScanner(fd)
	for input.Scan() {
		line := str.TrimSpace(input.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if str.HasPrefix(line, "[") && str.HasSuffix(line, "]") {
			section = str.TrimPrefix(line, "[")
			section = str.TrimSuffix(section, "]")
			section = str.TrimSpace(section)
			continue
		}
		s.parseLine(section, line)
	}
}

func (s *sessionT) parseLine(section, line string) {
	fields := str.Split(line, "=")
	errMsg := "incorrect config line: %s"

	if len(fields) != 2 {
		errExit(fmt.Errorf(errMsg, line))
	}

	key := str.TrimSpace(fields[0])
	val := str.TrimSpace(str.Join(fields[1:], "="))

	if section == "general" {
		switch key {
		case "user":
			s.user = val
		case "pass":
			s.pass = val
		case "mail_server":
			s.mailServer = val
		default:
			errExit(fmt.Errorf(errMsg, line))
		}
	} else if section == "folders" {
		var f folderT
		f.path = val
		f.mnemo = key
		s.folders = append(s.folders, f)
	} else {
		errExit(fmt.Errorf(errMsg, line))
	}
}
