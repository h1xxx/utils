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
	rd := bufio.NewReaderSize(vars.meminfoFd, 32)
	lineId := 1
	for {
		lineBin, _, err := rd.ReadLine()
		if err != nil {
			break
		}

		switch lineId {
		case 3:
			key, val := parseMeminfoLine(string(lineBin))
			if key != "MemAvailable:" {
				msg := "wrong /proc/meminfo line number %d: %s"
				errExit(fmt.Errorf(msg, lineId, key))
			}
			st.mem.avail = val * 1024
		case 5:
			key, val := parseMeminfoLine(string(lineBin))
			if key != "Cached:" {
				msg := "wrong /proc/meminfo line number %d: %s"
				errExit(fmt.Errorf(msg, lineId, key))
			}
			st.mem.cache = val * 1024
		case 24:
			key, val := parseMeminfoLine(string(lineBin))
			if key != "SReclaimable:" {
				msg := "wrong /proc/meminfo line number %d: %s"
				errExit(fmt.Errorf(msg, lineId, key))
			}
			st.mem.cache += val * 1024
		case 25:
			break
		}
		lineId++
	}
	// skip for benchmarking as this poses a large i/o bottleneck
	if !vars.bench {
		vars.meminfoFd.Seek(0, 0)
	}
}

func parseMeminfoLine(line string) (string, int) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return line, 0
	}
	key := fields[0]
	val, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		msg := "can't parse /proc/meminfo line: %s\nerror: %h"
		errExit(fmt.Errorf(msg, line, err))
	}

	return key, int(val)
}
