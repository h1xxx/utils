package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	fp "path/filepath"
	str "strings"

	"github.com/mdlayher/wifi"
)

type sttsT struct {
	uptime time.Duration
	loads  [3]float64
	procs  int
	mem    memT

	cpu1Temps  []string
	cpu2Temps  []string
	driveTemps []string
	moboTemps  []string

	ssid string
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
	meminfoFd *os.File

	cpu1TempHwmon string
	cpu2TempHwmon string
	cpu1TempFds   []*os.File
	cpu2TempFds   []*os.File

	driveTempHwmons []string
	driveTempFds    []*os.File

	moboTempHwmons []string
	moboTempFds    []*os.File

	i2cMoboTemps []string

	miscHwmonNames []string
	miscI2cNames   []string

	wifiClient *wifi.Client
	wifiIface  *wifi.Interface
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

	var vars varsT
	getVars(&vars)

	var st sttsT
	getSysinfo(&st, &vars)
	readCpuTemps(&st, &vars)
	readDriveTemps(&st, &vars)
	readMoboTemps(&st, &vars)
	getWifiInfo(&st, &vars)

	prettyPrint(st, vars)

	if *args.bench {
		doBench(&st, &vars)
	}

	vars.meminfoFd.Close()
	for _, fd := range vars.cpu1TempFds {
		fd.Close()
	}
	for _, fd := range vars.cpu2TempFds {
		fd.Close()
	}
	for _, fd := range vars.driveTempFds {
		fd.Close()
	}
	for _, fd := range vars.moboTempFds {
		fd.Close()
	}

	if vars.wifiClient != nil {
		vars.wifiClient.Close()
	}
}

func getVars(vars *varsT) {
	var err error
	vars.meminfoFd, err = os.Open("/proc/meminfo")
	errExit(err)

	hwmonDetect(vars)
	i2cDetect(vars)
	detectWlan(vars)

	vars.cpu1TempFds = openHwmon(vars.cpu1TempHwmon, "temp.*_input")
	vars.cpu2TempFds = openHwmon(vars.cpu2TempHwmon, "temp.*_input")

	for _, hwmon := range vars.driveTempHwmons {
		vars.driveTempFds = append(vars.driveTempFds,
			openHwmon(hwmon, "temp.*_input")...)
	}

	for _, hwmon := range vars.moboTempHwmons {
		vars.moboTempFds = append(vars.moboTempFds,
			openHwmon(hwmon, "temp.*_input")...)
	}

	vars.moboTempFds = append(vars.moboTempFds,
		openFiles(vars.i2cMoboTemps)...)
}

func openHwmon(hwmonDir string, ex string) []*os.File {
	var fds []*os.File

	re := regexp.MustCompile(ex)

	hwmonFiles, err := os.ReadDir(hwmonDir)
	if err != nil {
		return fds
	}

	for _, hwmonFile := range hwmonFiles {
		if !re.MatchString(hwmonFile.Name()) {
			continue
		}

		file := fp.Join(hwmonDir, hwmonFile.Name())

		fd, err := os.Open(file)
		if err != nil {
			continue
		}

		fds = append(fds, fd)
	}

	return fds
}

func openFiles(files []string) []*os.File {
	var fds []*os.File

	for _, file := range files {
		fd, err := os.Open(file)
		if err != nil {
			continue
		}

		fds = append(fds, fd)
	}

	return fds
}

func prettyPrint(st sttsT, vars varsT) {
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

	fmt.Printf("%-16s %s\n", "cpu1 temps", st.cpu1Temps)
	fmt.Printf("%-16s %s\n", "cpu2 temps", st.cpu2Temps)
	fmt.Printf("%-16s %s\n", "drive temps", st.driveTemps)
	fmt.Printf("%-16s %s\n\n", "mobo temps", st.moboTemps)

	fmt.Printf("%-16s %s\n", "cpu1 temp hwmon", vars.cpu1TempHwmon)
	fmt.Printf("%-16s %s\n", "cpu2 temp hwmon", vars.cpu2TempHwmon)
	fmt.Printf("%-16s %s\n", "drive temp hwmons", vars.driveTempHwmons)
	fmt.Printf("%-16s %s\n", "mobo temp hwmons", vars.moboTempHwmons)

	fmt.Printf("%-16s %s\n", "i2c mobo temp sensors:", vars.i2cMoboTemps)

	fmt.Printf("misc hwmon names:\n    %s\n",
		str.Join(vars.miscHwmonNames, "\n    "))
	fmt.Printf("misc i2c names:\n    %s\n",
		str.Join(vars.miscI2cNames, "\n    "))

	if vars.wifiClient != nil {
		fmt.Printf("%-16s %s\n", "wifi iface:", vars.wifiIface.Name)
		fmt.Printf("%-16s %s\n", "ssid:", st.ssid)
	}
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
