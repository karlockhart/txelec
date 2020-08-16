package txelec

import (
	"github.com/spf13/viper"
)

// LoadConfiguration loads the configuration
func LoadConfiguration() error {
	viper.SetConfigName("txelec")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/txelec/")
	viper.SetDefault("region.default", "LZ_NORTH")
	viper.SetDefault("storage.enabled", false)

	return viper.ReadInConfig()
}

// LMPIndexURL returns the LMP Report Index
func LMPIndexURL() string {
	return viper.GetString("sources.lmp_index")
}

// PriceAddersIndexURL returns the LMP Report Index
func PriceAddersIndexURL() string {
	return viper.GetString("sources.adders_index")
}

// SettlementIndexURL returns the Settlement Report Index
func SettlementIndexURL() string {
	return viper.GetString("sources.settlement_index")
}

// DefaultRegion returns the default region for which to show prices
func DefaultRegion() string {
	return viper.GetString("region.default")
}

// DefaultRegion returns the default region for which to show prices
func StorageEnabled() bool {
	return viper.GetBool("storage.enabled")
}
