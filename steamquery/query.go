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

	firstQueryRun time.Time
	lastQueryRun  time.Time
)

func main() {
	cfgPath := flag.String("c", "./files/config.json", "sets the config path")
	gCloudPath := flag.String("g", "./files/gcloud.json", "sets the google cloud config path")
	ignoreChecks := flag.Bool("ic", false, "[DEV] ignores checks (update, config, sheets conn, steam conn)")
	useBeta := flag.Bool("b", false, "opts into beta features")
	flag.Parse()

	if *ignoreChecks {
		log.Printf("[%s] Ignoring checks, NOT RECOMMENDED!\n", warnSign)
	}

	log.Printf("[%s] Currently running app version %s\n", infSign, version)

	// Check for updates.
	if !*ignoreChecks {
		log.Printf("[%s] Checking for updates...\n", infSign)

		latestRelease, err := checkLatestReleaseGithub()
		if err != nil {
			log.Printf("[%s] Checking latest Github release failed: %s\n", errSign, err.Error())
			return
		}

		isUpdated, err := checkVersionMatch(latestRelease)
		if err != nil {
			log.Printf("[%s] Checking latest release version failed: %s\n", errSign, err.Error())
			return
		}

		if !isUpdated {
			log.Printf("[%s] App is outdated, updating now...\n", warnSign)

			updateURL, err := findMatchingOSAndPlatform(latestRelease)
			if err != nil {
				log.Printf("[%s] Error finding release files: %s\n", errSign, err.Error())
				return
			}

			// Windows update files will be .zip.
			if strings.Contains(updateURL, "Windows") {
				if err := handlePatchDownloadAndUnzipWindows(updateURL); err != nil {
					log.Printf("[%s] Error downloading or unzipping patch: %s\n", errSign, err.Error())
					return
				}

				if err := os.RemoveAll("./tmp"); err != nil {
					log.Printf("[%s] Error removing temp directory: %s\n", errSign, err.Error())
				}
			} else {
				// Linux and MacOS update files will be .tar.gz.
				if err := handlePatchDownloadAndUnzip(updateURL); err != nil {
					log.Printf("[%s] Error downloading or unzipping patch: %s\n", errSign, err.Error())
					return
				}

				if err := os.RemoveAll("./tmp"); err != nil {
					log.Printf("[%s] Error removing temp directory: %s\n", errSign, err.Error())
				}
			}

			log.Printf("[%s] Please rerun the app to use the latest version\n", infSign)

			return
		} else {
			log.Printf("[%s] App is already running newest version\n", sucSign)
		}
	}

	if err := createDefaultLogDirectory(); err != nil {
		log.Printf("[%s] Creating logs directory failed: %s\n", errSign, err.Error())
		return
	}

	if err := createLogFile(); err != nil {
		log.Printf("[%s] Creating log file failed: %s\n", errSign, err.Error())
		return
	}

	if err := createLastQueryRunFile(); err != nil {
		log.Printf("[%s] Creating query log file failed: %s\n", errSign, err.Error())
		return
	}

	// It is safe to use the WriteX methods from here.

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		writeError(fmt.Sprintf("Error loading config: %s", err.Error()))
		return
	}

	if !*ignoreChecks {
		if err := cfg.checkConfig(); err != nil {
			writeError(fmt.Sprintf("Error checking config: %s", err.Error()))
			return
		}
	}

	// ! BETA FEATURE
	// Query the status of CSGO via Steam API and query user's inventory for cases & capsules.
	if *useBeta {
		steamUp, err := isSteamCSGOAPIUp(cfg)
		if err != nil {
			writeError(fmt.Sprintf("Querying Steam API failed: %s", err.Error()))
			return
		}

		if !steamUp {
			writeWarning("Steam might be down or delayed")
			return
		}

		itemCountMap, err := fetchCSGOInventory(cfg.SteamUserID64)
		if err != nil {
			writeError(fmt.Sprintf("Error fetching CSGO inventory: %s", err.Error()))
			return
		}

		// Stickers contain "Sticker" in item.name
		// Capsules contain "Base Grade ContaineR" in item.name
		for name, am := range itemCountMap {
			log.Printf("[%s] Name: %s ; Amount: %d\n", infSign, name, am)
		}
	}

	spreadsheetID = cfg.SpreadSheetID

	svc, err := newSpreadsheetService(*gCloudPath)
	if err != nil {
		writeError(fmt.Sprintf("Error creating spreadsheet service: %s", err.Error()))
		return
	}

	if !*ignoreChecks {
		if err := svc.testConnection(); err != nil {
			writeError(fmt.Sprintf("Error getting spreadsheet test info: %s", err.Error()))
		}
	}

	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			writeError(fmt.Sprintf("Error running clear screen func: %s", err.Error()))
			return
		}
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			writeError(fmt.Sprintf("Error running clear screen func: %s", err.Error()))
			return
		}
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			writeError(fmt.Sprintf("Error running clear screen func: %s", err.Error()))
			return
		}
	}

	runQuery(cfg, svc, *ignoreChecks)

	listenForCTRLC()

	writeInfo("Cleaning up, please wait...")

	if err := logFile.Close(); err != nil {
		log.Printf("[%s] Error closing log file: %s\n", errSign, err.Error())
		return
	}

	if err := lastQueryRunFile.Close(); err != nil {
		log.Printf("[%s] Error closing query log file: %s\n", errSign, err.Error())
		return
	}

	log.Printf("[%s] Done cleaning up, exiting...\n", sucSign)
}

func runQuery(cfg *config, svc *spreadsheetService, ignoreChecks bool) {
	callClear()

	if !ignoreChecks {
		steamUp, err := isSteamCSGOAPIUp(cfg)
		if err != nil {
			writeError(fmt.Sprintf("Querying Steam API failed: %s", err.Error()))

			writeInfo("Rerunning query in 30 mins...")

			time.AfterFunc(30*time.Minute, func() {
				runQuery(cfg, svc, ignoreChecks)
			})

			return
		}

		if !steamUp {
			writeWarning("Steam might be down or delayed")
			writeInfo("Rerunning query in 30 mins...")

			time.AfterFunc(30*time.Minute, func() {
				runQuery(cfg, svc, ignoreChecks)
			})

			return
		}

		jsonQuery, err := readFromQueryLogFile()
		if err != nil {
			writeError(fmt.Sprintf("Error reading last query log file: %s", err.Error()))
			return
		}

		lastQueryRun = jsonQuery.LastRun
		firstQueryRun = jsonQuery.FirstRun

		if lastQueryRun.IsZero() {
			writeInfo("Running query for 1st time...")
		} else {
			if time.Since(lastQueryRun) < 1*time.Minute {
				writeWarning(fmt.Sprintf("Time since last query run: %.0f second(s)", time.Since(lastQueryRun).Seconds()))
				writeWarning(fmt.Sprintf("Please manually run this app again in %.0f second(s)", time.Until(lastQueryRun.Add(1*time.Minute)).Seconds()))
				return
			}
			writeInfo(fmt.Sprintf("Last query run: %v", lastQueryRun))
		}
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
				writeError(fmt.Sprintf("Error on request: %s", err.Error()))

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					writeError(fmt.Sprintf("Error on updating error cell value: %s", err.Error()))
					return
				}

				return
			}

			req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")

			res, err := client.Do(req)
			if err != nil {
				writeError(fmt.Sprintf("Error on response: %s", err.Error()))

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					writeError(fmt.Sprintf("Error on updating error cell value: %s", err.Error()))
					return
				}

				return
			}
			defer res.Body.Close()

			if res.StatusCode != 200 {
				writeError(fmt.Sprintf("Error on response: %s", err.Error()))

				var errorInterface []interface{}

				errorInterface = append(errorInterface, fmt.Sprintf("Got unwanted status code on response: %d (reason: %s)", res.StatusCode, res.Status))

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					writeError(fmt.Sprintf("Error on updating error cell value: %s", err.Error()))
					return
				}

				return
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				writeError(fmt.Sprintf("Error converting response: %s", err.Error()))

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					writeError(fmt.Sprintf("Error on updating error cell value: %s", err.Error()))
					return
				}

				return
			}

			var resp responseBody

			if err := json.Unmarshal(body, &resp); err != nil {
				writeError(fmt.Sprintf("Error on json response: %s", err.Error()))

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					writeError(fmt.Sprintf("Error on updating error cell value: %s", err.Error()))
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
				writeError(fmt.Sprintf("Error on updating value on google sheet: %s", err.Error()))

				var errorInterface []interface{}

				errorInterface = append(errorInterface, err.Error())

				if err := updateEntryOnSheet(cfg.ErrorCell, errorInterface, svc); err != nil {
					writeError(fmt.Sprintf("Error on updating error cell value: %s", err.Error()))
					return
				}

				return
			}

			writeSuccess(fmt.Sprintf("UPDATED ITEM: %s ; LOWEST PRICE: %s", itemName, lowestPrice))

			// We gotta prevent spamming Steam or else we get a 429.
			if itemsCounted == 20 {
				itemsCounted = 0
				writeInfo("Sleeping for 1 minute to prevent spam...")
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
		writeError(fmt.Sprintf("Error updating last updated cell: %s", err.Error()))
		writeInfo("Rerunning Google sheets entry in 1 minute...")

		time.AfterFunc(1*time.Minute, func() {
			if err := svc.writeSingleEntryToTable(cfg.LastUpdatedCell, lastUpdateInterface); err != nil {
				writeError(fmt.Sprintf("Error updating last updated cell on 2nd try: %s", err.Error()))
				writeWarning("There might be something wrong with Google or your connection, exiting...")
				return
			}
		})
	}

	if totalItems != len(cfg.ItemList) {
		writeWarning(fmt.Sprintf("Items counted vs items list in config mismatch: %d vs. %d", totalItems, len(cfg.ItemList)))

		var errorInterface []interface{}

		errorInterface = append(errorInterface, "Not all items have been updated.")

		if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
			writeError(fmt.Sprintf("Error updating error cell: %s", err.Error()))
			writeInfo("Rerunning Google sheets entry in 1 minute...")

			time.AfterFunc(1*time.Minute, func() {
				if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
					writeError(fmt.Sprintf("Error updating error cell on 2nd try: %s", err.Error()))
					writeWarning("There might be something wrong with Google or your connection, exiting...")
					return
				}
			})
		}
	}

	// Say that no error occured.
	var errorInterface []interface{}

	errorInterface = append(errorInterface, "No error(s) occured.")

	if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
		writeError(fmt.Sprintf("Error updating error cell: %s", err.Error()))
		writeInfo("Rerunning Google sheets entry in 1 minute...")

		time.AfterFunc(1*time.Minute, func() {
			if err := svc.writeSingleEntryToTable(cfg.ErrorCell, errorInterface); err != nil {
				writeError(fmt.Sprintf("Error updating error cell on 2nd try: %s", err.Error()))
				writeWarning("There might be something wrong with Google or your connection, exiting...")
				return
			}
		})
	}

	// Function calls itself again after 12 hours.
	writeSuccess(fmt.Sprintf("Done, rerunning query again in %d hours...", cfg.UpdateInterval))

	lastQueryData := lastQueryRunFormat{}
	lastQueryData.LastRun = time.Now()

	if firstQueryRun.IsZero() {
		log.Println("GOT ZERO DATE FOR FIRST QUERY RUN")
		lastQueryData.FirstRun = time.Now()
	}

	jsonMarshal, err := json.Marshal(lastQueryData)
	if err != nil {
		writeError(fmt.Sprintf("Error marshaling to json: %s", err.Error()))
		return
	}

	if err := writeToQueryLogFile(string(jsonMarshal)); err != nil {
		writeError(fmt.Sprintf("Error writing to last query log file: %s", err.Error()))
		return
	}

	time.AfterFunc(time.Duration(cfg.UpdateInterval)*time.Hour, func() {
		runQuery(cfg, svc, ignoreChecks)
	})
}

func callClear() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		log.Fatalf("[%s] Unsupported platform\n", errSign)
	}
}

func updateEntryOnSheet(cell string, values []interface{}, svc *spreadsheetService) error {
	return svc.writeSingleEntryToTable(cell, values)
}
