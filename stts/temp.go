package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	fp "path/filepath"
	str "strings"
)

func readHwmon(st *sttsT, vars *varsT) {
	for _, fd := range vars.cpu1TempFds {
		st.cpu1Temps = append(st.cpu1Temps, readTemp(fd))
	}

	for _, fd := range vars.cpu2TempFds {
		st.cpu2Temps = append(st.cpu2Temps, readTemp(fd))
	}
}

func readTemp(fd *os.File) string {
	var tempBin [2]byte
	_, err := io.ReadFull(fd, tempBin[:])
	fmt.Println(tempBin)
	if err != nil {
		fd.Seek(0, 0)
		return "00"
	}

	fd.Seek(0, 0)

	return string(tempBin[:])
}

func hwmonDetect(vars *varsT) {
	hwmonDirs, err := os.ReadDir("/sys/class/hwmon")
	if err != nil {
		return
	}

	for _, dir := range hwmonDirs {
		file := fp.Join("/sys/class/hwmon", dir.Name(), "name")

		nameBin, err := ioutil.ReadFile(file)
		name, _, _ := str.Cut(string(nameBin), "\n")
		if err != nil {
			continue
		}

		hwmonName := fp.Join("/sys/class/hwmon/", dir.Name())

		switch name {
		case "k10temp", "coretemp":
			if vars.cpu1TempHwmon == "" {
				vars.cpu1TempHwmon = hwmonName
			} else {
				vars.cpu2TempHwmon = hwmonName
			}
		}
		fmt.Println(file, name)
	}

	sortCpuTempHwmon(vars)
}

func sortCpuTempHwmon(vars *varsT) {
	if vars.cpu2TempHwmon == "" {
		return
	}
	id1Str := str.TrimPrefix(vars.cpu1TempHwmon, "/sys/class/hwmon/hwmon")
	id2Str := str.TrimPrefix(vars.cpu2TempHwmon, "/sys/class/hwmon/hwmon")

	id1, err := strconv.Atoi(id1Str)
	errExit(err)
	id2, err := strconv.Atoi(id2Str)
	errExit(err)

	if id1 > id2 {
		tmp := vars.cpu1TempHwmon
		vars.cpu1TempHwmon = vars.cpu2TempHwmon
		vars.cpu2TempHwmon = tmp
	}
}
