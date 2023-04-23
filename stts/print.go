package main

import (
	"fmt"

	str "strings"
)

func printAll(st *sttsT, vars *varsT) {
	upDays := int(st.uptime.Hours() / 24)
	upHours := int(st.uptime.Hours()) % 24
	up := fmt.Sprintf("%dd%dh", upDays, upHours)
	prStr("uptime", up)
	sep()

	prFloat("load 1m", st.loads[0])
	prFloat("load 5m", st.loads[1])
	prFloat("load 15m", st.loads[2])
	sep()

	prInt("process count", st.procs)
	sep()

	mb := 1024 * 1024
	prInt("total mem", st.mem.total/mb)
	prInt("used mem", st.mem.used/mb)
	prInt("free mem", st.mem.free/mb)
	prInt("shared mem", st.mem.shared/mb)
	prInt("buffer", st.mem.buffer/mb)
	prInt("cache", st.mem.cache/mb)
	prInt("buff/cache", (st.mem.buffer+st.mem.cache)/mb)
	prInt("available", st.mem.avail/mb)
	sep()

	prStr("cpu1 temps", str.Join(st.cpu1Temps, ","))
	prStr("cpu2 temps", str.Join(st.cpu2Temps, ","))
	prStr("drive temps", str.Join(st.driveTemps, ","))
	prStr("mobo temps", str.Join(st.moboTemps, ","))
	sep()

	if vars.wifiClient != nil {
		prStr("wifi iface", vars.wifiIface.Name)
		prStr("ssid", st.wifiBss.SSID)
		prInt("wifi signal", st.wifiInfo.Signal)
		sep()
	}

	if vars.batCapacityFd != nil {
		prStr("bat level", st.batLevel)
	}

	if vars.batPowerFd != nil {
		prStr("bat time left", st.batTimeLeft)
	}
}

func printDebug(st *sttsT, vars *varsT) {
	sep()
	prStr("debug info", "")
	prStr("==========", "")
	sep()

	prStr("cpu1 temp hwmon  ", vars.cpu1TempHwmon)
	prStr("cpu2 temp hwmon  ", vars.cpu2TempHwmon)
	sep()

	prSl("drive temp hwmons", vars.driveTempHwmons)
	prSl("mobo temp hwmons", vars.moboTempHwmons)
	prSl("i2c mobo temp sensors", vars.i2cMoboTemps)
	prSl("misc hwmon names", vars.miscHwmonNames)
	prSl("misc i2c names", vars.miscI2cNames)
}

func prFloat(s string, f float64) {
	fmt.Printf("%-14s%10.2f\n", s, f)
}

func prInt(s string, i int) {
	fmt.Printf("%-14s%10d\n", s, i)
}

func prStr(s, v string) {
	fmt.Printf("%-14s%10s\n", s, v)
}

func prSl(s string, sl []string) {
	if len(sl) == 0 {
		return
	}
	fmt.Printf("%s\n    %s\n", s, str.Join(sl, "\n    "))
	sep()
}

func sep() {
	fmt.Println()
}
