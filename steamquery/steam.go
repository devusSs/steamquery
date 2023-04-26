package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "time/tzdata"
)

const (
	statusAPIURL = "https://api.steampowered.com/ICSGOServers_730/GetGameServersStatus/v1/?key="
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

	return resp.Result.Services.SessionsLogon == "normal" && resp.Result.Services.SteamCommunity == "normal", nil
}
