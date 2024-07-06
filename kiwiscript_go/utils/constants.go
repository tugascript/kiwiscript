// Copyright (C) 2024 Afonso Barracha
// 
// This file is part of KiwiScript.
// 
// KiwiScript is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// KiwiScript is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.

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
