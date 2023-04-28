package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	appID     = 730
	contextID = 2
)

func fetchCSGOInventory(userID64 string) (map[string]int, error) {
	url := fmt.Sprintf("http://steamcommunity.com/inventory/%s/%d/%d", userID64, appID, contextID)

	client := http.Client{}
	client.Timeout = 5 * time.Second

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("got unwanted Steam response status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var steamReturn steamInventoryReturn

	if err := json.Unmarshal(body, &steamReturn); err != nil {
		return nil, err
	}

	itemCountMap := make(map[string]int)

	for _, item := range steamReturn.Descriptions {
		itemCountMap[item.Name] = itemCountMap[item.Name] + 1
	}

	return itemCountMap, nil
}
