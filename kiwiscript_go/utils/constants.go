package utils

const (
	ProviderGoogle   string = "google"
	ProviderFacebook string = "facebook"
	ProviderEmail    string = "email"

	LocationNZL string = "NZL"
	LocationAUS string = "AUS"
	LocationNAM string = "NAM" // North America
	LocationEUR string = "EUR"
	LocationOTH string = "OTH" // Other
)

var Location = map[string]bool{
	LocationNZL: true,
	LocationAUS: true,
	LocationNAM: true,
	LocationEUR: true,
	LocationOTH: true,
}
