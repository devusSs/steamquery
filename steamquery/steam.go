package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "time/tzdata"
)

type steamStatus int

const (
	statusAPIURL = "https://api.steampowered.com/ICSGOServers_730/GetGameServersStatus/v1/?key="

	steamNormal  steamStatus = iota
	steamDelayed steamStatus = iota
	steamDown    steamStatus = iota
)

// Actual check on the Steam API for status of CSGO servers.
func isSteamCSGOAPIUp(cfg *config) (bool, error) {
	res, err := http.Get(fmt.Sprintf("%s%s", statusAPIURL, cfg.SteamAPIKey))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	var resp steamAPIResponse

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return false, err
	}

	var steamStatusSessions steamStatus
	var steamStatusCommunity steamStatus

	if resp.Result.Services.SessionsLogon == "normal" {
		steamStatusSessions = steamNormal
	} else if resp.Result.Services.SessionsLogon == "delayed" {
		steamStatusSessions = steamDelayed
	} else {
		steamStatusSessions = steamDown
	}

	if resp.Result.Services.SteamCommunity == "normal" {
		steamStatusCommunity = steamNormal
	} else if resp.Result.Services.SteamCommunity == "delayed" {
		steamStatusCommunity = steamDelayed
	} else {
		steamStatusCommunity = steamDown
	}

	return steamStatusSessions < 3 && steamStatusCommunity < 3, nil
}
