package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type sttsT struct {
	uptime time.Duration
	loads  [3]float64
	procs  int
	mem    memT
}

type memT struct {
	total  int
	used   int
	free   int
	shared int
	buffer int
	cache  int
	avail  int
}

type varsT struct {
	meminfoFD *os.File
}

type argsT struct {
	bench *bool
}

var args argsT

func init() {
	args.bench = flag.Bool("b", false, "perform a benchmark")
}

func main() {
	flag.Parse()

	var st sttsT
	var vars varsT

	getVars(&vars)
	getSysinfo(&st, &vars)
	prettyPrint(st)

	if *args.bench {
		doBench(&st, &vars)
	}

	vars.meminfoFD.Close()
}

func getVars(vars *varsT) {
	var err error
	vars.meminfoFD, err = os.Open("/proc/meminfo")
	errExit(err)
}

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
	scanner := bufio.NewScanner(vars.meminfoFD)
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

func prettyPrint(st sttsT) {
	upDays := int(st.uptime.Hours() / 24)
	upHours := int(st.uptime.Hours()) % 24
	fmt.Printf("%-16s%dd %dh\n\n", "uptime", upDays, upHours)

	fmt.Printf("%-16s%.2f\n", "load 1m", st.loads[0])
	fmt.Printf("%-16s%.2f\n", "load 5m", st.loads[1])
	fmt.Printf("%-16s%.2f\n\n", "load 15m", st.loads[2])

	fmt.Printf("%-16s%d\n\n", "process count", st.procs)

	MB := 1024 * 1024
	fmt.Printf("%-16s%5d\n", "total mem", st.mem.total/MB)
	fmt.Printf("%-16s%5d\n", "used mem", st.mem.used/MB)
	fmt.Printf("%-16s%5d\n", "free mem", st.mem.free/MB)
	fmt.Printf("%-16s%5d\n", "shared mem", st.mem.shared/MB)
	fmt.Printf("%-16s%5d\n", "buffer", st.mem.buffer/MB)
	fmt.Printf("%-16s%5d\n", "cache", st.mem.cache/MB)
	fmt.Printf("%-16s%5d\n", "buff/cache",
		(st.mem.buffer+st.mem.cache)/MB)
	fmt.Printf("%-16s%5d\n\n", "available", st.mem.avail/MB)
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
