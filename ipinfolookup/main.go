package main

import (
	"fmt"
	"os"
	"time"

	"ipinfolookup/ipinfo"
)

var IP_CSV = "/home/user/country_asn.csv.gz"

func main() {
	ipList := []string{
		"174.129.67.35",
		"123.241.231.80",
		"178.164.150.231",
		"148.163.160.5",
		"86.104.23.230",
		"116.21.16.234",
		"211.92.51.13",
		"83.5.12.95",
	}

	compact := true

	if !compact {
		fmt.Println("reading ipinfo csv file...")
		ipInfoSl, err := ipinfo.ReadIpInfoCsvGz(IP_CSV)
		errExit(err)
		fmt.Println("ip range count:", len(ipInfoSl))

		for _, ip := range ipList {
			ipInfo, err := ipinfo.IPLookup(ipInfoSl, ip)
			fmt.Println(ipInfo)
			errExit(err)
		}
	}

	if compact {
		fmt.Println("reading ipinfo csv file...")
		ipInfoSl, err := ipinfo.ReadIpInfoCsvGzCompact(IP_CSV)
		errExit(err)
		fmt.Println("ip range count:", len(ipInfoSl))

		for _, ip := range ipList {
			ipInfo, err := ipinfo.IPLookupCompact(ipInfoSl, ip)
			errExit(err)
			fmt.Println(ipInfo, string(ipInfo.CountryCode[:]))
		}
		time.Sleep(10 * time.Second)
	}
}

func errExit(err error, msg ...string) {
	if err != nil {
		if len(msg) == 1 {
			fmt.Printf("%s error: %s\n", msg[0], err)
		} else {
			fmt.Printf("error: %s\n", err)
		}
		os.Exit(1)
	}
}
