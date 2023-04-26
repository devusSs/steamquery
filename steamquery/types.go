package main

import "time"

type responseBody struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume      string `json:"volume"`
	MedianPrice string `json:"median_price"`
}

type steamAPIResponse struct {
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
			EUWest struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU West"`
			EUEast struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU East"`
			Poland struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Poland"`
			IndiaEast struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"India East"`
			HongKong struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Hong Kong"`
			Spain struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Spain"`
			Chile struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Chile"`
			USSouthwest struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Southwest"`
			USSoutheast struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Southeast"`
			India struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"India"`
			EUNorth struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"EU North"`
			Emirates struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Emirates"`
			USNorthwest struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Northwest"`
			SouthAfrica struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"South Africa"`
			Brazil struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Brazil"`
			USNortheast struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Northeast"`
			USNorthcentral struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"US Northcentral"`
			Japan struct {
				Capacity string `json:"capacity"`
				Load     string `json:"load"`
			} `json:"Japan"`
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

type lastQueryRunFormat struct {
	FirstRun time.Time `json:"first_run"`
	LastRun  time.Time `json:"last_run"`
}
