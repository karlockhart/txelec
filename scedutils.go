package txelec

import (
	"strconv"
	"strings"
)

// Price is a price per MWh
type Price struct {
	PriceMWh     float64
	RepeatedHour bool
	Timestamp    string
	Region       string
	Adder        Adder
}

// Adder is a price adder per MWh
type Adder struct {
	RTORDPA      float64
	RTORPA       float64
	All          float64
	RepeatedHour bool
	Timestamp    string
}

// Prices converts the raw k,v data to a list of prices
func Prices(prices []map[string]string) []Price {
	p := make([]Price, 0)
	for _, line := range prices {
		export := false
		t := Price{}
		for k, v := range line {
			switch strings.ToLower(k) {
			case "lmp":
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					continue
				}
				export = true
				t.PriceMWh = f
			case "repeatedhourflag":
				if strings.ToLower(v) == "n" {
					t.RepeatedHour = false
				} else if strings.ToLower(v) == "y" {
					t.RepeatedHour = true
				}
			case "scedtimestamp":
				t.Timestamp = v
			case "settlementpoint":
				t.Region = v
			}
		}
		if export {
			p = append(p, t)
		}
	}
	return p
}

// Adders converts raw k,v data to a list of adders
func Adders(adders []map[string]string) []Adder {
	ad := make([]Adder, 0)
	for _, line := range adders {
		export := false
		a := Adder{}
		for k, v := range line {
			switch strings.ToLower(k) {
			case "rtorpa":
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					continue
				}
				export = true
				a.RTORPA = f
				a.All += f
			case "rtordpa":
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					continue
				}
				export = true
				a.RTORDPA = f
				a.All += f
			case "repeatedhourflag":
				if strings.ToLower(v) == "n" {
					a.RepeatedHour = false
				} else if strings.ToLower(v) == "y" {
					a.RepeatedHour = true
				}
			case "scedtimestamp":
				a.Timestamp = v
			}
		}
		if export {
			ad = append(ad, a)
		}
	}
	return ad
}

// PricesWithAdders converts a list of k,v prices and adders into prices with matching adders
func PricesWithAdders(prices, adders []map[string]string) []Price {
	a := Adders(adders)
	p := Prices(prices)
	for i, v := range p {
		for _, av := range a {
			if av.Timestamp == v.Timestamp {
				p[i].Adder = av
			}
		}
	}
	return p
}
