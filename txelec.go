package txelec

import (
	"github.com/spf13/viper"
)

// LatestPrices returns the latest prices with adders
func LatestPrices() ([]Price, error) {
	url, err := GetLatestDLURL(viper.GetString("sources.lmp_index"))
	if err != nil {
		return nil, err
	}

	prices, err := GetDataFromZippedCSV(url.URL)
	if err != nil {
		return nil, err
	}

	url, err = GetLatestDLURL(viper.GetString("sources.adders_index"))
	if err != nil {
		return nil, err
	}

	adders, err := GetDataFromZippedCSV(url.URL)
	if err != nil {
		return nil, err
	}

	return PricesWithAdders(prices, adders), nil
}
