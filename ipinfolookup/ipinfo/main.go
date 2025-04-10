package ipinfo

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"

	str "strings"
)

type IpInfoT struct {
	StartIp    string
	EndIp      string
	StartIpInt int32
	EndIpInt   int32
	As         AsT
	Country    CountryT
}

type AsT struct {
	Asn      int
	AsName   string
	AsDomain string
}

type CountryT struct {
	CountryCode   string
	Country       string
	ContinentCode string
	Continent     string
}

func ReadIpInfoCsvGz(ipCsvFile string) ([]IpInfoT, error) {
	var ipInfoSl []IpInfoT

	fd, err := os.Open(ipCsvFile)
	if err != nil {
		return ipInfoSl, err
	}
	defer fd.Close()

	gr, err := gzip.NewReader(fd)
	if err != nil {
		return ipInfoSl, err
	}
	defer gr.Close()

	cr := csv.NewReader(gr)

	reIPv6 := regexp.MustCompile("^[0-9a-fA-F]{1,4}:")
	var i int
	for {
		// skip header
		if i == 0 {
			_, _ = cr.Read()
			i++
			continue
		}

		fields, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ipInfoSl, err
		}

		if len(fields) != 9 {
			msg := "incorrect line in country_asn.csv.gz: %s"
			err := fmt.Errorf(msg, str.Join(fields, ","))
			return ipInfoSl, err
		}

		var (
			ipInfo  IpInfoT
			as      AsT
			country CountryT
		)

		// skip ipv6
		if reIPv6.MatchString(fields[0]) {
			continue
		}

		if str.HasPrefix(fields[6], "AS") {
			asn, err := strconv.Atoi(fields[6][2:])
			if err != nil {
				msg := "incorrect line in country_asn.csv.gz: %s"
				fmt.Printf(msg, str.Join(fields, ","))
			}
			as.Asn = asn
			as.AsName = fields[7]
			as.AsDomain = fields[8]
		}

		if as.Asn == 0 {
			as.AsName = "n/a"
			as.AsDomain = "n/a"
		}

		if fields[2] != "" {
			country.CountryCode = fields[2]
			country.Country = fields[3]
			country.ContinentCode = fields[4]
			country.Continent = fields[5]
		} else {
			country.CountryCode = "n/a"
			country.Country = "n/a"
			country.ContinentCode = "n/a"
			country.Continent = "n/a"
		}

		if country.CountryCode == "" && as.Asn == 0 {
			continue
		}

		ipInfo.StartIp = fields[0]
		ipInfo.EndIp = fields[1]
		ipInfo.As = as
		ipInfo.Country = country

		ipInfo.StartIpInt, err = IpStrToInt(ipInfo.StartIp)
		if err != nil {
			return ipInfoSl, err
		}

		ipInfo.EndIpInt, err = IpStrToInt(ipInfo.EndIp)
		if err != nil {
			return ipInfoSl, err
		}

		if ipInfo.StartIpInt > ipInfo.EndIpInt {
			msg := "incorrect line in country_asn.csv.gz: %s"
			err := fmt.Errorf(msg, str.Join(fields, ","))
			return ipInfoSl, err
		}

		ipInfoSl = append(ipInfoSl, ipInfo)
	}

	sort.Slice(ipInfoSl, func(i, j int) bool {
		return ipInfoSl[i].StartIpInt >= ipInfoSl[j].StartIpInt
	})

	return ipInfoSl, nil
}

func IpStrToInt(ipStr string) (int32, error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return 0, fmt.Errorf("incorrect IP string: %s", ipStr)
	}

	ipUint := binary.BigEndian.Uint32(netIP[12:16])

	return int32(int64(ipUint) - 2147483648), nil
}

func IPLookup(ipRanges []IpInfoT, ip string) (IpInfoT, error) {
	ipInt, err := IpStrToInt(ip)
	if err != nil {
		return IpInfoT{}, nil
	}

	idx := sort.Search(len(ipRanges), func(i int) bool {
		return ipInt >= ipRanges[i].StartIpInt
	})

	// nothing was found
	if idx >= len(ipRanges) {
		return IpInfoT{}, nil
	}

	// ip not in range
	if ipInt < ipRanges[idx].StartIpInt || ipInt >= ipRanges[idx].EndIpInt {
		return IpInfoT{}, nil
	}

	return ipRanges[idx], nil
}
