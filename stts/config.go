package main

import (
	"bufio"
	"fmt"
	"os"

	str "strings"
)

func parseConfig(file string, vars *varsT) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	input := bufio.NewScanner(fd)
	for input.Scan() {
		line := str.TrimSpace(input.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parseLine(line, vars)
	}

	return nil
}

func parseLine(line string, vars *varsT) {
	fields := str.Split(line, "=")
	errMsg := "incorrect config line: %s"

	if len(fields) != 2 {
		errExit(fmt.Errorf(errMsg, line))
	}

	key := fields[0]
	val := fields[1]

	switch key {
	case "cpu_temp":
		vars.show.cpuTemp = getBoolVal(val, line)
	case "mobo_temp":
		vars.show.moboTemp = getBoolVal(val, line)
	case "drive_temp":
		vars.show.driveTemp = getBoolVal(val, line)
	case "wifi":
		vars.show.wifi = getBoolVal(val, line)
	case "battery":
		vars.show.bat = getBoolVal(val, line)
	case "add_info":
		if val == "" {
			return
		}
		fd, err := os.Open(val)
		errExit(err)
		vars.addInfoFd = append(vars.addInfoFd, fd)
	case "vpn_route":
		if val == "" {
			return
		}
		vars.vpnRoute = val
		vars.show.vpn = true
	case "vpn_pid":
		if val == "" {
			return
		}
		vars.vpnPidFile = val
		vars.show.vpn = true
	default:
		errExit(fmt.Errorf("incorrect config line: %s", line))
	}
}

func getBoolVal(val, line string) bool {
	switch val {
	case "true":
		return true
	case "false":
		return false
	default:
		errExit(fmt.Errorf("incorrect config line: %s", line))
	}
	return true
}
