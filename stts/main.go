package main

import (
	"flag"
	"fmt"
	"os"
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
