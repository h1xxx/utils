package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	str "strings"
)

type infoT struct {
	Servers []serverT `json:"servers"`
	Result  string    `json:"result"`
}

type serverT struct {
	Name        string `json:"public_name"`
	Country     string `json:"country_name"`
	CountryCode string `json:"country_code"`
	Location    string `json:"location"`
	Continent   string `json:"continent"`
	FreeBW      int
	Bandwith    int    `json:"bw"`
	BandwithMax int    `json:"bw_max"`
	Users       int    `json:"users"`
	Load        int    `json:"currentload"`
	IPv4_1      string `json:"ip_v4_in1"`
	IPv4_2      string `json:"ip_v4_in2"`
	IPv4_3      string `json:"ip_v4_in3"`
	IPv4_4      string `json:"ip_v4_in4"`
	Health      string `json:"health"`
	Warning     string `json:"warning"`
}

func main() {
	url := "https://airvpn.org/api/status/"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "curl/7.74.0")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var info infoT

	err = json.Unmarshal(body, &info)
	if err != nil {
		panic(err)
	}

	var all, eu, america, asia, oceania []serverT

	for _, s := range info.Servers {
		s.Location, _, _ = str.Cut(s.Location, ",")
		s.Location = str.TrimSuffix(s.Location, " City")
		s.FreeBW = s.BandwithMax - s.Bandwith

			all = append(all, s)

		switch s.Continent {
		case "Europe":
			eu = append(eu, s)
		case "America":
			america = append(america, s)
		case "Asia":
			asia = append(asia, s)
		case "Oceania":
			oceania = append(oceania, s)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Load >= all[j].Load
	})

	sort.Slice(eu, func(i, j int) bool {
		return eu[i].Load >= eu[j].Load
	})

	sort.Slice(america, func(i, j int) bool {
		return america[i].Load >= america[j].Load
	})

	sort.Slice(asia, func(i, j int) bool {
		return asia[i].Load >= asia[j].Load
	})

	sort.Slice(oceania, func(i, j int) bool {
		return oceania[i].Load >= oceania[j].Load
	})

	header:="server name     cc  city      bandwidth users load  stat  vpn name"

	fmt.Println("+ america")
	fmt.Println(header)
	for _, s := range america {
		printServer(s)
	}

	fmt.Println("\n+ europe")
	fmt.Println(header)
	for _, s := range eu {
		printServer(s)
	}

	fmt.Println("\n+ asia")
	fmt.Println(header)
	for _, s := range asia {
		printServer(s)
	}

	fmt.Println("\n+ oceania")
	fmt.Println(header)
	for _, s := range oceania {
		printServer(s)
	}

	fmt.Println("\n+ all")
	fmt.Println(header)
	for _, s := range all {
		printServer(s)
	}
}

func printServer(s serverT) {
	fmt.Printf("%-16s%s  %-13s", s.Name, s.CountryCode, s.Location)
	fmt.Printf("%6d %5d %4d", s.FreeBW, s.Users, s.Load)
	fmt.Printf("%6s  ", s.Health)
	fmt.Printf("%s-%s\n", s.CountryCode, str.ToLower(s.Name))
}
