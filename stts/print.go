package main

import (
	"fmt"
	"time"

	str "strings"
)

func printOneLine(st *sttsT, vars *varsT) {
	for {
		getAllInfo(st, vars)
		printOneLineOnce(st, vars)
		time.Sleep(5 * time.Second)
	}
}

func printOneLineOnce(st *sttsT, vars *varsT) {
	var out []string

	load := fmt.Sprintf("load %.2f", st.loads[0])

	var mem string
	memUsed := float64(st.mem.used) / (1024 * 1024)
	if memUsed > 1024 {
		mem = fmt.Sprintf("mem %.1fG", memUsed/1024)
	} else {
		mem = fmt.Sprintf("mem %.0fM", memUsed)
	}

	var disk string
	if st.rootDiskFree > 1024 {
		disk = fmt.Sprintf("df %.1fG", st.rootDiskFree/1024)
	} else {
		disk = fmt.Sprintf("df %.0fM", st.rootDiskFree)
	}

	var temps, sep string
	if vars.has.cpu1Temp && !vars.has.cpu2Temp {
		temps += fmt.Sprintf("cpu %s°C", st.cpu1Temp)
		sep = " "
	} else if vars.has.cpu1Temp && vars.has.cpu2Temp {
		temps += fmt.Sprintf("cpu1 %s°C", st.cpu1Temp)
		temps += fmt.Sprintf(" cpu2 %s°C", st.cpu2Temp)
		sep = " "
	}

	if vars.has.moboTemp {
		temps += sep + fmt.Sprintf("mb %s°C", st.moboTemp)
		sep = " "
	}

	if vars.has.driveTemp {
		temps += sep + fmt.Sprintf("d %s°C", st.driveTemp)
		sep = " "
	}

	var wifi string
	if vars.has.wifi {
		wifi = vars.wifiIface.Name
		if st.wifiBss != nil && st.wifiBss.SSID != "" {
			wifi += fmt.Sprintf(" %s %ddBm",
				st.wifiBss.SSID, st.wifiInfo.Signal)
		} else {
			wifi += " no conn"
		}
	}

	var bat string
	if vars.has.bat {
		bat = "bat " + st.batLevel + "%"
		if st.batTimeLeft != "0:00" {
			bat += " " + st.batTimeLeft
		}
	}

	if len(st.addInfo) > 0 {
		out = append(out, st.addInfo...)
	}

	out = append(out, load)
	out = append(out, mem)
	out = append(out, disk)

	if temps != "" {
		out = append(out, temps)
	}

	if wifi != "" {
		out = append(out, wifi)
	}

	if bat != "" {
		out = append(out, bat)
	}

	fmt.Printf("%s\n", str.Join(out, " | "))
}

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
	prFloat("/ disk free", st.rootDiskFree/1024)
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
	prInt("hugepages", st.mem.huge/mb)
	sep()

	prStr("cpu1 temp", st.cpu1Temp)
	prStr("cpu2 temp", st.cpu2Temp)
	prStr("drive temp", st.driveTemp)
	prStr("mobo temp", st.moboTemp)
	sep()

	if vars.wifiClient != nil && st.wifiBss != nil {
		prStr("wifi iface", vars.wifiIface.Name)
		prStr("ssid", st.wifiBss.SSID)
		prInt("wifi signal", st.wifiInfo.Signal)
		sep()
	}

	if vars.batCapacityFd != nil {
		prStr("bat level", st.batLevel+"%")
	}

	if vars.batPowerFd != nil {
		prStr("bat time left", st.batTimeLeft)
	}

	for _, p := range st.ps {
		if !vars.login && p.args[0] == '-' {
			continue
		}

		sep()
		prStrL("pid", p.pid)
		prStrL("comm", p.stat.comm)
		prStrL("bin", p.bin)
		prStrL("args", p.args)
		prStrL("pwd", p.pwd)
		prIntL("MB read", p.readBytes/mb)
		prIntL("MB written", p.writeBytes/mb)
		prIntL("fd count", p.fdCount)
		prIntL("ppid", p.stat.ppid)

		if vars.files {
			header := "files"
			for _, file := range p.files {
				prStrL(header, file)
				header = ""
			}
		}

		if vars.env {
			header := "env vars"
			for _, env := range p.env {
				prStrL(header, env)
				header = ""
			}
		}
	}
}

func printDebug(st *sttsT, vars *varsT) {
	sep()
	prStr("debug info", "")
	prStr("==========", "")
	sep()

	prStr("cpu1 temps", str.Join(st.cpu1Temps, ","))
	prStr("cpu2 temps", str.Join(st.cpu2Temps, ","))
	prStr("drive temps", str.Join(st.driveTemps, ","))
	prStr("mobo temps", str.Join(st.moboTemps, ","))
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

func prIntL(s string, i int) {
	fmt.Printf("%-14s%d\n", s, i)
}

func prStr(s, v string) {
	fmt.Printf("%-14s%10s\n", s, v)
}

func prStrL(s, v string) {
	fmt.Printf("%-14s%s\n", s, v)
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
