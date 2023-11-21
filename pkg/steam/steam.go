// Provides basic access to Steam API endpoints for CS2 (app id 730)
package steam

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SteamStatus represents the status of the Steam services
type SteamStatus struct {
	Sessions  status
	Community status
}

// String returns a string representation of the SteamStatus
func (s SteamStatus) String() string {
	return fmt.Sprintf("sessions: %v, community: %v", s.Sessions, s.Community)
}

// IsOnline returns true if both sessions and community are online
func (s SteamStatus) IsOnline() bool {
	return s.Sessions < offline && s.Community < offline
}

// IsDelayed returns true if either sessions or community is delayed
func (s SteamStatus) IsDelayed() bool {
	return s.Sessions == delayed || s.Community == delayed
}

// GetSteamStatus returns the current status of the Steam services
func GetSteamStatus(apiKey string) (SteamStatus, error) {
	if apiKey == "" {
		return SteamStatus{}, fmt.Errorf("api key is empty")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(statusURL, apiKey), nil)
	if err != nil {
		return SteamStatus{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SteamStatus{}, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SteamStatus{}, fmt.Errorf("error getting status: %v", resp.Status)
	}

	var status statusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return SteamStatus{}, fmt.Errorf("error decoding response: %v", err)
	}

	return SteamStatus{
		Sessions:  parseStatus(status.Result.Services.SessionsLogon),
		Community: parseStatus(status.Result.Services.SteamCommunity),
	}, nil
}

// SteamInventoryResponse represents the response from the Steam inventory "API"
type SteamInventoryResponse struct {
	Assets []struct {
		Appid      int    `json:"appid"`
		Contextid  string `json:"contextid"`
		Assetid    string `json:"assetid"`
		Classid    string `json:"classid"`
		Instanceid string `json:"instanceid"`
		Amount     string `json:"amount"`
	} `json:"assets"`
	Descriptions []struct {
		Appid           int    `json:"appid"`
		Classid         string `json:"classid"`
		Instanceid      string `json:"instanceid"`
		Currency        int    `json:"currency"`
		BackgroundColor string `json:"background_color"`
		IconURL         string `json:"icon_url"`
		IconURLLarge    string `json:"icon_url_large,omitempty"`
		Descriptions    []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
			Color string `json:"color,omitempty"`
		} `json:"descriptions"`
		Tradable int `json:"tradable"`
		Actions  []struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"actions,omitempty"`
		Name           string `json:"name"`
		NameColor      string `json:"name_color"`
		Type           string `json:"type"`
		MarketName     string `json:"market_name"`
		MarketHashName string `json:"market_hash_name"`
		MarketActions  []struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"market_actions,omitempty"`
		Commodity                 int `json:"commodity"`
		MarketTradableRestriction int `json:"market_tradable_restriction"`
		Marketable                int `json:"marketable"`
		Tags                      []struct {
			Category              string `json:"category"`
			InternalName          string `json:"internal_name"`
			LocalizedCategoryName string `json:"localized_category_name"`
			LocalizedTagName      string `json:"localized_tag_name"`
			Color                 string `json:"color,omitempty"`
		} `json:"tags"`
		Fraudwarnings []string `json:"fraudwarnings,omitempty"`
	} `json:"descriptions"`
	TotalInventoryCount int `json:"total_inventory_count"`
	Success             int `json:"success"`
	Rwgrsn              int `json:"rwgrsn"`
}

// GetInventory returns the inventory of the given Steam user
func GetInventory(steamID uint64) (SteamInventoryResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(inventoryURL, steamID), nil)
	if err != nil {
		return SteamInventoryResponse{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SteamInventoryResponse{}, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SteamInventoryResponse{}, fmt.Errorf("error getting inventory: %v", resp.Status)
	}

	var inventory SteamInventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&inventory); err != nil {
		return SteamInventoryResponse{}, fmt.Errorf("error decoding response: %v", err)
	}

	return inventory, nil
}

// Gets a supported currency by it's ISO4217 code
func GetCurrencySignByCode(isoCode string) (string, error) {
	for code, sign := range supportedCurrenties {
		if isoCode == code {
			return sign, nil
		}
	}
	return "", fmt.Errorf("currency %s not supported", isoCode)
}

var (
	supportedCurrenties = map[string]string{
		"USD": "$",
		"GBP": "£",
		"EUR": "€",
		"CHF": "CHF",
		"RUB": "₽",
		"PLN": "zł",
		"BRL": "R$",
		"JPY": "¥",
		"NOK": "kr",
		"IDR": "Rp",
		"MYR": "RM",
		"PHP": "₱",
		"SGF": "S$",
		"THB": "฿",
		"VND": "₫",
		"KRW": "₩",
		"TRY": "₺",
		"UAH": "₴",
		"MXN": "$",
		"CAD": "CA$",
		"AUD": "A$",
		"NZD": "NZ$",
		"CNY": "¥",
		"INR": "₹",
		"CLP": "$",
		"PEN": "S/",
		"COP": "$",
		"ZAR": "R",
		"HKD": "HK$",
		"TWD": "NT$",
		"SAR": "﷼",
		"AED": "د.إ",
		"ARS": "$",
		"ILS": "₪",
		"KZT": "₸",
		"KWD": "د.ك",
		"QAR": "ر.ق",
		"CRC": "₡",
		"UYU": "$",
	}
)

type status int

func (s status) String() string {
	switch s {
	case online:
		return "online"
	case delayed:
		return "delayed"
	case offline:
		return "offline"
	default:
		return "unknown"
	}
}

const (
	online  status = iota
	delayed status = iota
	offline status = iota
)

const (
	statusURL    = "https://api.steampowered.com/ICSGOServers_730/GetGameServersStatus/v1/?key=%s"
	inventoryURL = "https://steamcommunity.com/inventory/%d/730/2"
)

type statusResponse struct {
	Result struct {
		App struct {
			Version   int    `json:"version"`
			Timestamp int    `json:"timestamp"`
			Time      string `json:"time"`
		} `json:"app"`
		Services struct {
			SessionsLogon  string `json:"SessionsLogon"`
			SteamCommunity string `json:"SteamCommunity"`
			IEconItems     string `json:"IEconItems"`
			Leaderboards   string `json:"Leaderboards"`
		} `json:"services"`
		Datacenters struct {
			Peru struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Peru"`
			EUGermany struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Germany"`
			EUAustria struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Austria"`
			EUPoland struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Poland"`
			IndiaChennai struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"India Chennai"`
			HongKong struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Hong Kong"`
			EUSpain struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Spain"`
			Chile struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Chile"`
			USCalifornia struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US California"`
			USAtlanta struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Atlanta"`
			IndiaBombay struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"India Bombay"`
			EUSweden struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Sweden"`
			Emirates struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Emirates"`
			USSeattle struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Seattle"`
			SouthAfrica struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"South Africa"`
			Brazil struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Brazil"`
			USVirginia struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Virginia"`
			USChicago struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Chicago"`
			Japan struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Japan"`
			EUHolland struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Holland"`
			USNewYork struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US NewYork"`
			EUFinland struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU Finland"`
			IndiaMumbai struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"India Mumbai"`
			Argentina struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Argentina"`
			SouthKorea struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"South Korea"`
			Singapore struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Singapore"`
			Australia struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Australia"`
			ChinaChengdu struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"China Chengdu"`
			ChinaShanghai struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"China Shanghai"`
			ChinaTianjin struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"China Tianjin"`
			ChinaGuangzhou struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"China Guangzhou"`
		} `json:"datacenters"`
		Matchmaking struct {
			Scheduler        string `json:"scheduler"`
			OnlineServers    int    `json:"online_servers"`
			OnlinePlayers    int    `json:"online_players"`
			SearchingPlayers int    `json:"searching_players"`
			SearchSecondsAvg int    `json:"search_seconds_avg"`
		} `json:"matchmaking"`
		Perfectworld struct {
			Logon struct {
				Availability string `json:"availability"`
				Latency      string `json:"latency"`
			} `json:"logon"`
			Purchase struct {
				Availability string `json:"availability"`
				Latency      string `json:"latency"`
			} `json:"purchase"`
		} `json:"perfectworld"`
	} `json:"result"`
}

func parseStatus(s string) status {
	switch s {
	case "normal":
		return online
	case "delayed":
		return delayed
	case "offline":
		return offline
	default:
		return offline
	}
}
