package ipinfo

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"

	str "strings"
)

type IpInfoCompactT struct {
	StartIpInt  int32
	EndIpInt    int32
	Asn         int32
	AsDomain    string
	CountryCode [2]byte
}

func ReadIpInfoCsvGzCompact(ipCsvFile string) ([]IpInfoCompactT, error) {
	var ipInfoSl []IpInfoCompactT

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

		var ipInfo IpInfoCompactT

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
			ipInfo.Asn = int32(asn)

			// too large for compact list - disabled
			//ipInfo.AsName = fields[7]

			ipInfo.AsDomain = fields[8]
		}

		if ipInfo.Asn == 0 {
			// too large for compact list - disabled
			//ipInfo.AsName = "n/a"

			ipInfo.AsDomain = "n/a"

			// AS info should be always available, if it's not just
			// skip it
			continue
		}

		countryCode := fields[2]
		if countryCode == "" && ipInfo.Asn == 0 {
			continue
		}

		if countryCode == "" || len(countryCode) != 2 {
			countryCode = "__"
		}

		ipInfo.CountryCode[0] = countryCode[0]
		ipInfo.CountryCode[1] = countryCode[1]

		startIp := fields[0]
		endIp := fields[1]

		ipInfo.StartIpInt, err = IpStrToInt(startIp)
		if err != nil {
			return ipInfoSl, err
		}

		ipInfo.EndIpInt, err = IpStrToInt(endIp)
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

func IPLookupCompact(ipRanges []IpInfoCompactT, ip string) (IpInfoCompactT, error) {
	ipInt, err := IpStrToInt(ip)
	if err != nil {
		return IpInfoCompactT{}, nil
	}

	idx := sort.Search(len(ipRanges), func(i int) bool {
		return ipInt >= ipRanges[i].StartIpInt
	})

	// nothing was found
	if idx >= len(ipRanges) {
		return IpInfoCompactT{}, nil
	}

	// ip not in range
	if ipInt < ipRanges[idx].StartIpInt || ipInt >= ipRanges[idx].EndIpInt {
		return IpInfoCompactT{}, nil
	}

	return ipRanges[idx], nil
}
