package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func getSysinfo(st *sttsT, vars *varsT) {
	si := &syscall.Sysinfo_t{}
	_ = syscall.Sysinfo(si)

	uptimeTime := time.Unix(si.Uptime, 0)
	st.uptime = uptimeTime.Sub(time.Unix(0, 0))

	const SI_LOAD_SHIFT = 16
	st.loads[0] = float64(si.Loads[0]) / (1 << SI_LOAD_SHIFT)
	st.loads[1] = float64(si.Loads[1]) / (1 << SI_LOAD_SHIFT)
	st.loads[2] = float64(si.Loads[2]) / (1 << SI_LOAD_SHIFT)

	st.procs = int(si.Procs)

	st.mem.total = int(si.Totalram)
	st.mem.free = int(si.Freeram)
	st.mem.shared = int(si.Sharedram)
	st.mem.buffer = int(si.Bufferram)

	// add missing info from /proc/meminfo
	readMeminfo(st, vars)
	st.mem.used = st.mem.total - st.mem.free - st.mem.buffer - st.mem.cache
}

func readMeminfo(st *sttsT, vars *varsT) {
	scanner := bufio.NewScanner(vars.meminfoFd)
	lineID := 1
	for scanner.Scan() {
		switch lineID {
		case 3:
			key, val := parseMeminfoLine(scanner.Text())
			if key != "MemAvailable" {
				msg := "wrong /proc/meminfo line number: %s"
				errExit(fmt.Errorf(msg, key))
			}
			st.mem.avail = val * 1024
		case 5:
			key, val := parseMeminfoLine(scanner.Text())
			if key != "Cached" {
				msg := "wrong /proc/meminfo line number: %s"
				errExit(fmt.Errorf(msg, key))
			}
			st.mem.cache = val * 1024
		case 24:
			key, val := parseMeminfoLine(scanner.Text())
			if key != "SReclaimable" {
				msg := "wrong /proc/meminfo line number: %s"
				errExit(fmt.Errorf(msg, key))
			}
			st.mem.cache += val * 1024
		case 25:
			break
		}
		lineID++
	}
}

func parseMeminfoLine(line string) (string, int) {
	fields := strings.Split(line, ":")
	if len(fields) != 2 {
		errExit(fmt.Errorf("can't parse /proc/meminfo line: %s", line))
	}
	key := strings.TrimSpace(fields[0])
	valStr := strings.TrimSpace(fields[1])
	valStr = strings.TrimSuffix(valStr, " kB")
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		msg := "can't parse /proc/meminfo line: %s\nerror: %h"
		errExit(fmt.Errorf(msg, line, err))
	}

	return key, int(val)
}
