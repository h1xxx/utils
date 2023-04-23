package main

import (
	"fmt"

	str "strings"
)

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
		fmt.Printf("%-16s %s\n", "ssid:", st.wifiBss.SSID)
		fmt.Printf("%-16s %d\n", "wifi signal:", st.wifiInfo.Signal)
	}

	if vars.batCapacityFd != nil {
		fmt.Printf("%-16s %s%%\n", "battery level:", st.batLevel)
	}
	if vars.batPowerFd != nil {
		fmt.Printf("%-16s %s\n", "battery time left:", st.batTimeLeft)
	}
}
