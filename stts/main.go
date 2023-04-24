package main

import (
	"flag"
	"os"
	"regexp"
	"time"

	fp "path/filepath"

	"github.com/mdlayher/wifi"
)

type sttsT struct {
	uptime       time.Duration
	loads        [3]float64
	procs        int
	mem          memT
	rootDiskFree float64

	cpu1Temp  string
	cpu2Temp  string
	driveTemp string
	moboTemp  string

	cpu1Temps  []string
	cpu2Temps  []string
	driveTemps []string
	moboTemps  []string

	wifiBss  *wifi.BSS
	wifiInfo *wifi.StationInfo

	batLevel    string
	batTimeLeft string
}

type memT struct {
	total  int
	used   int
	free   int
	shared int
	buffer int
	cache  int
	avail  int
	huge   int
}

type varsT struct {
	bench bool
	debug bool

	has hasT

	meminfoFd *os.File

	cpu1TempHwmon string
	cpu2TempHwmon string
	cpu1TempFds   []*os.File
	cpu2TempFds   []*os.File

	driveTempHwmons []string
	driveTempFds    []*os.File

	moboTempHwmons []string
	i2cMoboTemps   []string
	moboTempFds    []*os.File

	miscHwmonNames []string
	miscI2cNames   []string

	wifiClient *wifi.Client
	wifiIface  *wifi.Interface

	batCapacityFd   *os.File
	batEnergyFd     *os.File
	batEnergyFullFd *os.File
	batPowerFd      *os.File
	batStatusFd     *os.File
}

type hasT struct {
	cpu1Temp  bool
	cpu2Temp  bool
	driveTemp bool
	moboTemp  bool
	wifi      bool
	bat       bool
}

func main() {
	var oneLine, oneLineOnce, bench, debug bool

	flag.BoolVar(&oneLine, "o", false, "print info in one line repeatedly")
	flag.BoolVar(&oneLineOnce, "1", false, "print info in one line once")
	flag.BoolVar(&bench, "b", false, "perform a benchmark")
	flag.BoolVar(&debug, "d", false, "add debugging info")

	flag.Parse()

	var st sttsT
	var vars varsT

	vars.debug = debug
	vars.bench = bench
	getVars(&vars)

	switch {
	case oneLine:
		printOneLine(&st, &vars)
	case oneLineOnce:
		getAllInfo(&st, &vars)
		printOneLineOnce(&st, &vars)
	default:
		getAllInfo(&st, &vars)
		printAll(&st, &vars)
		if debug {
			printDebug(&st, &vars)
		}
	}

	if bench {
		doBench(&st, &vars)
	}

	closeFiles(&vars)
}

func getAllInfo(st *sttsT, vars *varsT) {
	getSysinfo(st, vars)
	getDiskInfo(st, vars)
	readCpuTemps(st, vars)
	readDriveTemps(st, vars)
	readMoboTemps(st, vars)
	getWifiInfo(st, vars)
	getBatInfo(st, vars)
}

func getVars(vars *varsT) {
	var err error
	vars.meminfoFd, err = os.Open("/proc/meminfo")
	errExit(err)

	hwmonDetect(vars)
	i2cDetect(vars)
	detectWlan(vars)
	detectBat(vars)

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

	if len(vars.cpu1TempFds) > 0 {
		vars.has.cpu1Temp = true
	}

	if len(vars.cpu2TempFds) > 0 {
		vars.has.cpu2Temp = true
	}

	if len(vars.driveTempFds) > 0 {
		vars.has.driveTemp = true
	}

	if len(vars.moboTempFds) > 0 {
		vars.has.moboTemp = true
	}

	if vars.wifiIface != nil {
		vars.has.wifi = true
	}

	if vars.batCapacityFd != nil {
		vars.has.bat = true
	}
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

func closeFiles(vars *varsT) {
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

	if vars.batCapacityFd != nil {
		vars.batCapacityFd.Close()
	}

	if vars.batEnergyFd != nil {
		vars.batEnergyFd.Close()
	}

	if vars.batEnergyFullFd != nil {
		vars.batEnergyFullFd.Close()
	}

	if vars.batPowerFd != nil {
		vars.batPowerFd.Close()
	}

	if vars.batStatusFd != nil {
		vars.batStatusFd.Close()
	}
}
