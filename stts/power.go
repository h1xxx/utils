package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	str "strings"
)

func detectBat(vars *varsT) {
	var err error
	batLoc := "/sys/class/power_supply/BAT0/"

	vars.batCapacityFd, err = os.Open(batLoc + "capacity")
	if err != nil {
		vars.batCapacityFd = nil
	}

	vars.batEnergyFd, err = os.Open(batLoc + "energy_now")
	if err != nil {
		vars.batEnergyFd = nil
	}

	vars.batEnergyFullFd, err = os.Open(batLoc + "energy_full")
	if err != nil {
		vars.batEnergyFullFd = nil
	}

	vars.batPowerFd, err = os.Open(batLoc + "power_now")
	if err != nil {
		vars.batPowerFd = nil
	}

	vars.batStatusFd, err = os.Open(batLoc + "status")
	if err != nil {
		vars.batStatusFd = nil
	}
}

func getBatInfo(st *sttsT, vars *varsT) {
	st.batLevel = readShortString(vars.batCapacityFd)

	batEnergy := readPower(vars.batEnergyFd)
	batEnergyFull := readPower(vars.batEnergyFullFd)
	batPower := readPower(vars.batPowerFd)
	batStatus := readShortString(vars.batStatusFd)

	var minLeft int

	if batStatus == "Dis" && batPower != 0 {
		minLeft = int(batEnergy / batPower * 60)
	} else if batStatus == "Cha" && batPower != 0 {
		minLeft = int((batEnergyFull - batEnergy) / batPower * 60)
	}

	st.batTimeLeft = fmt.Sprintf("%d:%2.2d", minLeft/60, minLeft%60)

	// skip for benchmarking as this poses a large i/o bottleneck
	if !vars.bench {
		vars.batCapacityFd.Seek(0, 0)
		vars.batEnergyFd.Seek(0, 0)
		vars.batEnergyFullFd.Seek(0, 0)
		vars.batPowerFd.Seek(0, 0)
		vars.batStatusFd.Seek(0, 0)
	}
}

func readShortString(fd *os.File) string {
	rd := bufio.NewReaderSize(fd, 3)
	var levelBin [3]byte
	_, err := rd.Read(levelBin[:])
	if err != nil {
		return "0"
	}

	levelFields := str.Fields(string(levelBin[:]))

	return levelFields[0]
}

func readPower(fd *os.File) float64 {
	rd := bufio.NewReaderSize(fd, 10)
	var powerBin [10]byte
	_, err := rd.Read(powerBin[:])
	if err != nil {
		return 0
	}

	powerStr := str.Fields(string(powerBin[:]))
	power, err := strconv.ParseFloat(powerStr[0], 64)
	if err != nil {
		return 0
	}

	return power
}
