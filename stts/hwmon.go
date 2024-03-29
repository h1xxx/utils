package main

import (
	"io/ioutil"
	"os"
	"strconv"

	fp "path/filepath"
	str "strings"
)

func readCpuTemps(st *sttsT, vars *varsT) {
	st.cpu1Temp = ""
	st.cpu2Temp = ""
	if vars.debug {
		st.cpu1Temps = []string{}
		st.cpu2Temps = []string{}
	}

	for _, fd := range vars.cpu1TempFds {
		temp := readTemp(fd, vars)
		if temp > st.cpu1Temp {
			st.cpu1Temp = temp
		}
		if vars.debug {
			st.cpu1Temps = append(st.cpu1Temps, temp)
		}
	}

	for _, fd := range vars.cpu2TempFds {
		temp := readTemp(fd, vars)
		if temp > st.cpu2Temp {
			st.cpu2Temp = temp
		}
		if vars.debug {
			st.cpu2Temps = append(st.cpu2Temps, temp)
		}
	}
}

func readDriveTemps(st *sttsT, vars *varsT) {
	st.driveTemp = ""
	if vars.debug {
		st.driveTemps = []string{}
	}
	for _, fd := range vars.driveTempFds {
		temp := readTemp(fd, vars)
		if temp > st.driveTemp {
			st.driveTemp = temp
		}
		if vars.debug {
			st.driveTemps = append(st.driveTemps, temp)
		}
	}
}

func readMoboTemps(st *sttsT, vars *varsT) {
	st.moboTemp = ""
	if vars.debug {
		st.moboTemps = []string{}
	}

	for _, fd := range vars.moboTempFds {
		temp := readTemp(fd, vars)
		if temp > st.moboTemp {
			st.moboTemp = temp
		}
		if vars.debug {
			st.moboTemps = append(st.moboTemps, temp)
		}
	}
}

func readTemp(fd *os.File, vars *varsT) string {
	tempBin := make([]byte, 2)
	_, err := fd.Read(tempBin)

	// skip for benchmarking as this poses a large i/o bottleneck
	if !vars.bench {
		fd.Seek(0, 0)
	}

	if err != nil {
		return "00"
	}

	temp := string(tempBin[:])
	if temp == "10" || temp == "11" || len(temp) == 1 {
		temp = "99"
	}

	return temp
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
		case "drivetemp":
			vars.driveTempHwmons = append(vars.driveTempHwmons,
				hwmonName)
		case "acpitz":
			vars.moboTempHwmons = append(vars.moboTempHwmons,
				hwmonName)
		default:
			vars.miscHwmonNames = append(vars.miscHwmonNames,
				dir.Name()+":"+name)
		}
	}

	sortCpuTempHwmon(vars)
}

func i2cDetect(vars *varsT) {
	i2cDirs, err := os.ReadDir("/sys/bus/i2c/devices")
	if err != nil {
		return
	}

	for _, dir := range i2cDirs {
		file := fp.Join("/sys/bus/i2c/devices", dir.Name(), "name")

		nameBin, err := ioutil.ReadFile(file)
		name, _, _ := str.Cut(string(nameBin), "\n")
		if err != nil {
			continue
		}

		i2cName := fp.Join("/sys/bus/i2c/devices", dir.Name())

		switch name {
		case "w83795g":
			vars.i2cMoboTemps = append(
				vars.i2cMoboTemps, i2cName+"/temp1_input")
		default:
			vars.miscI2cNames = append(vars.miscI2cNames,
				dir.Name()+":"+name)

		}
	}
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
