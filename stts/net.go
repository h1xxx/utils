package main

import (
	"net"

	"github.com/mdlayher/wifi"
)

func detectWlan(vars *varsT) {
	ifaces, err := net.Interfaces()
	errExit(err)
	for _, iface := range ifaces {
		_ = iface.Name
	}

	vars.wifiClient, _ = wifi.New()

	wifis, _ := vars.wifiClient.Interfaces()
	for _, wifi := range wifis {
		vars.wifiIface = wifi
		// todo: add some filtering in case of multiple wifi devices
		return
	}
	vars.wifiClient = nil
	vars.wifiIface = nil
}

func getWifiInfo(st *sttsT, vars *varsT) {
	if vars.wifiClient == nil {
		return
	}

	bss, _ := vars.wifiClient.BSS(vars.wifiIface)
	st.ssid = bss.SSID
}
