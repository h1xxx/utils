package main

import (
	"bufio"
	"fmt"
	"strconv"
	"syscall"
	"time"

	str "strings"
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
	st.mem.used -= st.mem.huge
}

func readMeminfo(st *sttsT, vars *varsT) {
	rd := bufio.NewReaderSize(vars.meminfoFd, 32)
	for {
		lineBin, _, err := rd.ReadLine()
		if err != nil {
			break
		}
		fields := str.Fields(string(lineBin))

		switch fields[0] {
		case "MemAvailable:":
			val := parseMeminfoLine(fields)
			st.mem.avail = val * 1024
		case "Cached:":
			val := parseMeminfoLine(fields)
			st.mem.cache = val * 1024
		case "SReclaimable:":
			val := parseMeminfoLine(fields)
			st.mem.cache += val * 1024
		case "Hugetlb:":
			val := parseMeminfoLine(fields)
			st.mem.huge = val * 1024
		}
	}

	// skip for benchmarking as this poses a large i/o bottleneck
	if !vars.bench {
		vars.meminfoFd.Seek(0, 0)
	}
}

func parseMeminfoLine(fields []string) int {
	if len(fields) < 2 {
		return 0
	}
	val, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		msg := "can't parse /proc/meminfo line: %s\nerror: %h"
		errExit(fmt.Errorf(msg, str.Join(fields, " "), err))
	}

	return int(val)
}
