package txelec

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
)

type API struct {
	lastPrice  map[string]float64
	lastUpdate map[string]string
	location   *time.Location
	region     string
	client     influxdb2.Client
}

type DispPrice struct {
	Region  string  `json:"region"`
	Price   float64 `json:"price_kwh"`
	Updated string  `json:"update_dts"`
}

func (a *API) updatePrice() error {
	prices, err := LatestPrices()
	if err != nil {
		return err
	}

	pricemap := make(map[string]Price)

	for _, price := range prices {
		pricemap[price.Region] = price
	}

	for k, v := range pricemap {
		a.lastPrice[k] = (v.PriceMWh + v.Adder.All) / 10
		a.lastUpdate[k] = v.Timestamp
		logrus.Infof("region: %s, price: %f, updated: %s", k, a.lastPrice[k], a.lastUpdate[k])
	}

	return nil
}

func (a *API) Price(rgn string) float64 {
	if price, ok := a.lastPrice[rgn]; ok {
		return price
	}
	return -1
}

func (a *API) Start() {
	logrus.Info("starting API")
	go func() {
	sync_loop:
		for {
			logrus.Info("starting sync...")
			a.updatePrice()
			a.storePrice()
			pr := a.lastUpdate[DefaultRegion()]
			for {
				logrus.Info("waiting for first price change...")
				time.Sleep(30 * time.Second)
				a.updatePrice()
				if a.lastUpdate[DefaultRegion()] != pr {
					break
				}
			}
			logrus.Info("sync achieved")
			st := time.Now()
			ti := time.NewTicker(5 * time.Minute)
			for range ti.C {
				a.updatePrice()
				a.storePrice()
				if time.Now().After(st.Add(time.Hour)) {
					logrus.Info("resyncing...")
					ti.Stop()
					continue sync_loop
				}
			}
		}
	}()

	http.HandleFunc("/elec", func(w http.ResponseWriter, r *http.Request) {
		pRgn := r.URL.Query().Get("region")
		pFmt := r.URL.Query().Get("format")

		rgn := DefaultRegion()
		formt := "csv"

		if _, ok := a.lastPrice[pRgn]; pRgn == "all" || (len(pRgn) < 32 && ok) {
			rgn = pRgn
		}

		if strings.ToLower(pFmt) == "json" {
			formt = "json"
		}

		prices := make([]DispPrice, 0)
		if pRgn == "all" {
			for k, v := range a.lastPrice {
				prices = append(prices, DispPrice{
					Region:  k,
					Price:   v,
					Updated: a.lastUpdate[k],
				})
			}
		} else {
			prices = append(prices, DispPrice{
				Region:  rgn,
				Price:   a.lastPrice[rgn],
				Updated: a.lastUpdate[rgn],
			})
		}

		if formt == "csv" {
			for _, price := range prices {
				if rgn == "all" {
					fmt.Fprintf(w, "%.2f,%s,%s\n", price.Price, price.Region, price.Updated)
				} else {
					fmt.Fprintf(w, "%.2f,%s\n", price.Price, price.Updated)
				}
			}
			return
		} else if formt == "json" {
			out, err := json.MarshalIndent(prices, "", "  ")
			if err != nil {
				fmt.Fprint(w, "error writing json response")
			}
			fmt.Fprint(w, string(out))
			return
		}

		fmt.Fprint(w, "invalid format")

	})

	logrus.Info("listening for requests")
	log.Fatal(http.ListenAndServe(":8001", nil))
}

func NewAPI() (*API, error) {
	l, err := time.LoadLocation("America/Chicago")
	if err != nil {
		return nil, err
	}

	a := &API{
		region:   DefaultRegion(),
		location: l,
	}
	a.lastPrice = make(map[string]float64)
	a.lastUpdate = make(map[string]string)
	a.client = influxdb2.NewClient("http://localhost:8086", "")

	err = a.updatePrice()

	return a, err
}

func (a *API) storePrice() {
	if !StorageEnabled() {
		return
	}

	writeAPI := a.client.WriteAPIBlocking("", "tx_electricity/autogen")

	for k, v := range a.lastPrice {
		p := influxdb2.NewPoint("price",
			map[string]string{"region": k},
			map[string]interface{}{"price": v},
			time.Now())
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			logrus.Error(err)
		}
	}

	return
}
