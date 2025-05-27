package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"ipinfolookup/ipinfo"
	str "strings"
)

type cfg struct {
	ipCsv   *string
	sep     *string
	field   *int
	test    *bool
	compact *bool
}

func main() {
	var cfg cfg

	cfg.ipCsv = flag.String("i", "", "ipinfo.io country_asn.csv.gz file")
	cfg.sep = flag.String("s", "\t", "field separator")
	cfg.field = flag.Int("f", 1, "field for ip address string")
	cfg.test = flag.Bool("test", false, "run a test")
	cfg.compact = flag.Bool("compact", false, "use compact data in test")

	flag.Parse()

	if *cfg.ipCsv == "" {
		*cfg.ipCsv = os.Getenv("IPINFO_CSV")
	}

	if *cfg.ipCsv == "" {
		home, _ := os.UserHomeDir()
		if fileExists(home + "/country_asn.csv.gz") {
			*cfg.ipCsv = home + "/country_asn.csv.gz"
		} else {
			fmt.Println("please specify ipinfo.io csv input file")
			os.Exit(1)
		}
	}

	if len(*cfg.sep) != 1 {
		fmt.Println("please specify a single separator")
		os.Exit(1)
	}

	if *cfg.test {
		testLib(cfg)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "reading ipinfo csv file...")
	ipInfoSl, err := ipinfo.ReadIpInfoCsvGz(*cfg.ipCsv)
	errExit(err)

	input := bufio.NewScanner(os.Stdin)
	var sb str.Builder

	sepStr := *cfg.sep
	sepRune := rune(sepStr[0])

	for input.Scan() {
		line := input.Text()
		fields := str.Split(line, *cfg.sep)

		sb.WriteString(line)

		if len(fields) < *cfg.field {
			sb.WriteRune('\n')
			continue
		}

		ip := fields[*cfg.field-1]
		ipInfo, err := ipinfo.IPLookup(ipInfoSl, ip)
		if err != nil {
			sb.WriteRune('\n')
			continue
		}

		sb.WriteRune(sepRune)
		sb.WriteString(ipInfo.Country.CountryCode)

		sb.WriteRune(sepRune)
		sb.WriteString("AS")
		sb.WriteString(strconv.Itoa(ipInfo.As.Asn))

		sb.WriteRune(sepRune)
		sb.WriteString(ipInfo.As.AsDomain)

		sb.WriteRune(sepRune)
		sb.WriteString(ipInfo.As.AsName)
		sb.WriteRune('\n')

		if sb.Len() > 65536 {
			fmt.Print(sb.String())
			sb.Reset()
		}
	}

	err = input.Err()
	errExit(err)

	fmt.Print(sb.String())
}

func testLib(cfg cfg) {
	// some real-world, non-generated ip addresses
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

	ipRandomList := generateRandomPublicIPv4s(1000)
	ipList = append(ipRandomList, ipList...)

	if !*cfg.compact {
		fmt.Println("reading ipinfo csv file to non-compact data struct...")
		ipInfoSl, err := ipinfo.ReadIpInfoCsvGz(*cfg.ipCsv)
		errExit(err)
		fmt.Println("ip range count:", len(ipInfoSl))

		for _, ip := range ipList {
			ipInfo, err := ipinfo.IPLookup(ipInfoSl, ip)
			errExit(err)
			if ipInfo.As.Asn != 0 {
				fmt.Println(ip, ipInfo)
			}
		}
	}

	if *cfg.compact {
		fmt.Println("reading ipinfo csv file to compact data struct...")
		ipInfoSl, err := ipinfo.ReadIpInfoCsvGzCompact(*cfg.ipCsv)
		errExit(err)
		fmt.Println("ip range count:", len(ipInfoSl))

		for _, ip := range ipList {
			ipInfo, err := ipinfo.IPLookupCompact(ipInfoSl, ip)
			errExit(err)
			if ipInfo.Asn != 0 {
				fmt.Printf("%-19s AS%d\t%-40s %3s\n",
					ip, ipInfo.Asn, ipInfo.AsDomain,
					string(ipInfo.CountryCode[:]))
			}
		}
	}
}

func isPublicIPv4(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	// check private and reserved ranges
	_, private24, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16, _ := net.ParseCIDR("192.168.0.0/16")
	_, reserved, _ := net.ParseCIDR("100.64.0.0/10")

	return !private24.Contains(ip) &&
		!private20.Contains(ip) &&
		!private16.Contains(ip) &&
		!reserved.Contains(ip)
}

func generateRandomPublicIPv4s(count int) []string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	ips := make([]string, 0, count)

	for len(ips) < count {
		// Generate random IP
		ip := net.IPv4(
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
		).To4()

		if ip != nil && isPublicIPv4(ip) {
			ips = append(ips, ip.String())
		}
	}

	return ips
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
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
