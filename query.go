package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	_ "time/tzdata"
)

const (
	baseURL = "https://steamcommunity.com/market/priceoverview/?appid=730&currency=3&market_hash_name="
)

var (
	clear map[string]func()

	lastUpdate time.Time
)

func main() {
	cfgPath := flag.String("c", "./files/config.json", "sets the config path")
	gCloudPath := flag.String("g", "./files/gcloud.json", "sets the google cloud config path")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		log.Println("Error loading config: ", err.Error())
		return
	}

	if err := cfg.checkConfig(); err != nil {
		log.Println("Error checking config: ", err.Error())
		return
	}

	spreadsheetID = cfg.SpreadSheetID

	svc, err := newSpreadsheetService(*gCloudPath)
	if err != nil {
		log.Println("Error creating spreadsheet service: ", err.Error())
		return
	}

	if err := svc.testConnection(); err != nil {
		log.Println("Error getting spreadsheet test info: ", err.Error())
	}

	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	runQuery(cfg, svc)

	listenForCTRLC()

	log.Println("[EXIT] Program exit since runQuery() has not been called again")
}

func runQuery(cfg *config, svc *spreadsheetService) {
	callClear()

	if err := pingSteamOnline(); err != nil {
		log.Println("[WARN] There might be an issue with your network or Steam might be down: ", err.Error())
		log.Println("[INFO] Rerunning query in 30 mins...")

		time.AfterFunc(30*time.Minute, func() {
			runQuery(cfg, svc)
		})

		return
	}

	pstTime, err := timeIn(time.Now(), "America/Los_Angeles")
	if err != nil {
		log.Println("Error loading timezone: ", err.Error())
		return
	}

	steamDown, err := checkForSteamUsualDowntime(pstTime)
	if err != nil {
		log.Println("Error checking if Steam might be down: ", err.Error())
		return
	}

	if steamDown {
		log.Println("[WARN] Steam might have issues, trying to query again in 30 mins...")

		time.AfterFunc(30*time.Minute, func() {
			runQuery(cfg, svc)
		})

		return
	}

	log.Println("[START] Running query...")

	if lastUpdate.IsZero() {
		log.Println("[INFO] Running query for 1st time...")
	} else {
		log.Printf("[INFO] Last query run: %v\n", lastUpdate)
	}

	client := http.Client{}
	client.Timeout = 5 * time.Second

	itemsCounted := 0
	totalItems := 0

	for _, item := range cfg.ItemList {
		for itemName, tableCell := range item {
			expNameEsc := url.QueryEscape(itemName)

			req, err := http.NewRequest(http.MethodGet, baseURL+expNameEsc, nil)
			if err != nil {
				log.Println("Error on request: ", err.Error())

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					log.Println("Error on updating error cell value: ", err.Error())
					return
				}

				return
			}

			req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")

			res, err := client.Do(req)
			if err != nil {
				log.Println("Error on response: ", err.Error())

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					log.Println("Error on updating error cell value: ", err.Error())
					return
				}

				return
			}
			defer res.Body.Close()

			if res.StatusCode != 200 {
				log.Println("Error on response: ", res.Status)

				var errorInterface []interface{}

				errorInterface = append(errorInterface, fmt.Sprintf("Got unwanted status code on response: %d (reason: %s)", res.StatusCode, res.Status))

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					log.Println("Error on updating error cell value: ", err.Error())
					return
				}

				return
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Println("Error converting response: ", err.Error())

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					log.Println("Error on updating error cell value: ", err.Error())
					return
				}

				return
			}

			var resp responseBody

			if err := json.Unmarshal(body, &resp); err != nil {
				log.Println("Error on json response: ", err.Error())

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					log.Println("Error on updating error cell value: ", err.Error())
					return
				}

				return
			}

			var priceInterface []interface{}

			// Fixes any errors caused by Steam setting "-" instead of a 0 in price.
			lowestPrice := strings.ReplaceAll(resp.LowestPrice, "-", "0")

			priceInterface = append(priceInterface, lowestPrice)

			itemsCounted++
			totalItems++

			if err := updateEntryOnSheet(tableCell, priceInterface, svc); err != nil {
				log.Println("Error on updating value on google sheet: ", err.Error())

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					log.Println("Error on updating error cell value: ", err.Error())
					return
				}

				return
			}

			log.Printf("[SUCCESS] UPDATED ITEM: %s ; LOWEST PRICE: %s\n", itemName, lowestPrice)

			// We gotta prevent spamming Steam or else we get a 429.
			if itemsCounted == 20 {
				itemsCounted = 0
				log.Println("[INFO] Sleeping for 1 minute to prevent spam...")
				time.Sleep(1 * time.Minute)
			}
		}
	}

	var lastUpdateInterface []interface{}

	timeNow := time.Now()

	hour := timeNow.Hour()
	minute := timeNow.Minute()
	year, month, day := timeNow.Date()

	lastUpdateInterface = append(lastUpdateInterface, fmt.Sprintf("%d.%d.%d - %d:%d", day, int(month), year, hour, minute))

	if err := svc.writeSingleEntryToTable(cfg.LastUpdatedCell, lastUpdateInterface); err != nil {
		log.Println("Error updating last updated cell: ", err.Error())
		log.Println("[INFO] Rerunning Google sheets entry in 1 minute...")

		time.AfterFunc(1*time.Minute, func() {
			if err := svc.writeSingleEntryToTable(cfg.LastUpdatedCell, lastUpdateInterface); err != nil {
				log.Println("Error updating last updated cell on 2nd try: ", err.Error())
				log.Println("[WARN] There might be something wrong with Google or your connection, exiting...")
				return
			}
		})
	}

	if totalItems != len(cfg.ItemList) {
		log.Printf("Items counted vs items list in config mismatch: %d vs. %d\n", totalItems, len(cfg.ItemList))

		var errorInterface []interface{}

		errorInterface = append(errorInterface, "Not all items have been updated.")

		if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
			log.Println("Error updating last updated cell: ", err.Error())
			log.Println("[INFO] Rerunning Google sheets entry in 1 minute...")

			time.AfterFunc(1*time.Minute, func() {
				if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
					log.Println("Error updating error cell on 2nd try: ", err.Error())
					log.Println("[WARN] There might be something wrong with Google or your connection, exiting...")
					return
				}
			})
		}
	}

	// Say that no error occured.
	var errorInterface []interface{}

	errorInterface = append(errorInterface, "No error(s) occured.")

	if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
		log.Println("Error updating error cell: ", err.Error())
		log.Println("[INFO] Rerunning Google sheets entry in 1 minute...")

		time.AfterFunc(1*time.Minute, func() {
			if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
				log.Println("Error updating error cell on 2nd try: ", err.Error())
				log.Println("[WARN] There might be something wrong with Google or your connection, exiting...")
				return
			}
		})
	}

	// Function calls itself again after 12 hours.
	log.Printf("[DONE] Running query again in %d hours...\n", cfg.UpdateInterval)

	lastUpdate = time.Now()

	time.AfterFunc(time.Duration(cfg.UpdateInterval)*time.Hour, func() {
		runQuery(cfg, svc)
	})
}

func callClear() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		panic("Your platform is unsupported! I can't clear the terminal screen :(")
	}
}

func updateEntryOnSheet(cell string, values []interface{}, svc *spreadsheetService) error {
	return svc.writeSingleEntryToTable(cell, values)
}
