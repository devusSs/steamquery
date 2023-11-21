package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/devusSs/steamquery/internal/backpack"
	bratelimit "github.com/devusSs/steamquery/internal/backpack/ratelimit"
	"github.com/devusSs/steamquery/internal/config"
	"github.com/devusSs/steamquery/internal/format"
	"github.com/devusSs/steamquery/internal/steam/filter"
	"github.com/devusSs/steamquery/internal/steam/ratelimit"
	sratelimit "github.com/devusSs/steamquery/internal/steam/ratelimit"
	"github.com/devusSs/steamquery/internal/tables"
	"github.com/devusSs/steamquery/internal/updater"
	"github.com/devusSs/steamquery/pkg/log"
	"github.com/devusSs/steamquery/pkg/steam"
	"github.com/devusSs/steamquery/pkg/system"
	flag "github.com/spf13/pflag"
)

func main() {
	startTime := time.Now()

	var helpFlag *bool = flag.Bool("help", false, "Print help information and exit")
	var versionFlag *bool = flag.Bool("version", false, "Print version information and exit")
	var noUpdateFlag *bool = flag.Bool("no-update", false, "Disable update check on startup")
	var debugFlag *bool = flag.Bool("debug", false, "Enable debug logging (more and verbose logging, also console output)")
	var consoleFlag *bool = flag.Bool("console", false, "Enable console output (debug does that automatically)")
	var logsDirFlag *string = flag.StringP("logs", "l", "./logs", "Directory to store logs in")
	var configFileFlag *string = flag.StringP("config", "c", "./.config.json", "Path to config file")
	var filterFileFlag *string = flag.StringP("filter", "f", "", "Path to filter file if desired, empty uses default filter")
	var itemsFileFlag *string = flag.StringP("items", "i", "", "Path to additional items file if desired, empty uses raw inventory")
	var gcloudFileFlag *string = flag.StringP("gcloud", "g", ".gcloud.json", "Path to Google credentials file")
	flag.Parse()

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	if err := checkOSAndArchComp(); err != nil {
		fmt.Printf("Error checking OS / arch compatibility: %s\n", err.Error())
		os.Exit(1)
	}

	if !*noUpdateFlag {
		if err := updater.CheckForUpdatesAndApply(buildVersion); err != nil {
			fmt.Printf("Error updating: %s\n", err.Error())
			os.Exit(1)
		}
	}

	log.SetDefaultLogsDirectory(*logsDirFlag)
	log.SetDefaultLogFileName("steamquery.log")

	logger := log.NewLogger(
		log.WithName("main"),
		log.WithConsole(*consoleFlag),
		log.WithDebug(*debugFlag),
	)

	logger.Info("App start")

	sratelimit.SetRateLimitConfig(*logsDirFlag, 15, time.Minute)
	bratelimit.SetRateLimitConfig(*logsDirFlag, 1000, time.Hour)

	state, err := sratelimit.LoadRateLimitState()
	if err != nil {
		logger.Error("Error loading rate limit state: %v", err)
		os.Exit(1)
	}

	logger.Debug("loaded steam rate limit state: %v", state)

	if !sratelimit.WithinRateLimit(state, 2) {
		logger.Error("Rate limit exceeded, retry later")
		os.Exit(1)
	}

	logger.Debug("steam rate limit not exceeded, continuing")

	cfg, err := config.Load(*configFileFlag)
	if err != nil {
		logger.Error("Error loading config: %v", err)
		os.Exit(1)
	}

	logger.Debug("loaded config from %s: %v", *configFileFlag, cfg)
	logger.Info("Successfully loaded config file")

	currencySign, err := steam.GetCurrencySignByCode(cfg.Currency)
	if err != nil {
		logger.Error("Error getting currency sign for %s: %v", cfg.Currency, err)
		os.Exit(1)
	}

	logger.Debug("will use currency sign: %s", currencySign)

	steamStatus, err := steam.GetSteamStatus(cfg.SteamAPIKey)
	if err != nil {
		logger.Error("Error getting Steam status: %v", err)
		os.Exit(1)
	}

	state.LastRequestTime = time.Now()
	state.RequestCount++
	if err := ratelimit.SaveRateLimitState(state); err != nil {
		logger.Error("Error saving rate limit state: %v", err)
		os.Exit(1)
	}

	logger.Debug("got steam status: %v", steamStatus)

	if !steamStatus.IsOnline() {
		logger.Error("Steam services are offline, retry later")
		os.Exit(1)
	}

	if steamStatus.IsDelayed() {
		logger.Warn("Steam services are delayed, expect issues")
	} else {
		logger.Info("Steam services are online")
	}

	logger.Info("Fetching inventory...")

	inv, err := steam.GetInventory(cfg.SteamUserID64)
	if err != nil {
		logger.Error("Error getting inventory: %v", err)
		os.Exit(1)
	}

	state.LastRequestTime = time.Now()
	state.RequestCount++
	if err := ratelimit.SaveRateLimitState(state); err != nil {
		logger.Error("Error saving rate limit state: %v", err)
		os.Exit(1)
	}

	logger.Debug("queried inventory for user %d", cfg.SteamUserID64)
	logger.Info("Successfully got inventory")

	logger.Debug("unfiltered inventory response: %d item(s)", len(inv.Descriptions))

	if err := filter.LoadFilterOptions(*filterFileFlag); err != nil {
		logger.Error("Error loading filter options: %v", err)
		os.Exit(1)
	}

	logger.Debug("loaded filter options: %v", filter.GetFilterSettings())

	inv = filter.FilterInventoryResponse(inv)

	logger.Debug("filtered inventory response: %d item(s)", len(inv.Descriptions))

	amountMap := filter.GetItemAmountMap(inv)

	logger.Debug("got amount map: %d item(s)", len(amountMap))

	itemsAmountMap, err := filter.AddItems(amountMap, *itemsFileFlag)
	if err != nil {
		logger.Error("Adding additional items failed: %v", err)
		os.Exit(1)
	}

	logger.Debug("added items to amount map: total: %d item(s)", len(itemsAmountMap))

	if !backpack.IsAvailable() {
		logger.Error("csgobackpack is currently unavailable, retry later")
		os.Exit(1)
	}

	logger.Info("Fetching item prices...")

	bstate, err := bratelimit.LoadRateLimitState()
	if err != nil {
		logger.Error("Error loading rate limit state: %v", err)
		os.Exit(1)
	}

	logger.Debug("loaded backpack rate limit state: %v", bstate)

	if !bratelimit.WithinRateLimit(bstate, len(amountMap)) {
		logger.Error("Rate limit exceeded, retry later")
		os.Exit(1)
	}

	logger.Debug("backpack rate limit not exceeded, continuing")

	items := make([]inventoryItem, 0, len(itemsAmountMap))
	for marketHashName, amount := range itemsAmountMap {
		price, err := backpack.GetItemPrice(
			marketHashName,
			&backpack.RequestOptions{
				MedianTime: cfg.MedianPriceDays,
				Currency:   cfg.Currency,
			},
		)
		if err != nil {
			if err != backpack.ZeroPriceError {
				logger.Error("Error getting item price: %v", err)
				os.Exit(1)
			}
			price = 0.0
			logger.Warn("Item currently has no price: %s", marketHashName)
		}
		item := inventoryItem{
			MarketHashName: marketHashName,
			Amount:         amount,
			Price:          price,
		}
		items = append(items, item)

		bstate.LastRequestTime = time.Now()
		bstate.RequestCount++
	}

	logger.Debug("got item prices: %d item(s)", len(items))
	logger.Info("Successfully fetched item prices")

	if err := bratelimit.SaveRateLimitState(bstate); err != nil {
		logger.Error("Error saving rate limit state: %v", err)
		os.Exit(1)
	}

	sheetsSvc, err := tables.NewSpreadsheetService(*gcloudFileFlag, cfg.SpreadSheetID)
	if err != nil {
		logger.Error("Error creating spreadsheet service: %v", err)
		os.Exit(1)
	}

	logger.Debug("successfully created spreadsheet service")

	if err := sheetsSvc.Test(); err != nil {
		logger.Error("Error testing spreadsheet connection: %v", err)
		os.Exit(1)
	}

	logger.Debug("successfully tested spreadsheet connection")

	logger.Info("Fetching pre run data from spreadsheet...")

	preRunData, err := fetchPreRunData(sheetsSvc, cfg, currencySign)
	if err != nil {
		logger.Error("Error fetching pre run data: %v", err)
		os.Exit(1)
	}

	logger.Info("Successfully fetched pre run data from spreadsheet")

	if preRunData.Error != lastRunNoError {
		logger.Warn("Last run error: %s", preRunData.Error)
	}

	if time.Since(preRunData.LastUpdated) < lastUpdateCooldown {
		logger.Warn(
			"Last update was less than %s ago, please refrain from spamming",
			lastUpdateCooldown.String(),
		)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].MarketHashName < items[j].MarketHashName
	})

	startRow := cfg.StartingRow
	endRow := startRow + uint(len(items))

	logger.Debug("will write to rows %d-%d", startRow, endRow)

	logger.Info("Writing data to spreadsheet...")

	itemsData := make([][]interface{}, 0, len(items))
	for _, item := range items {
		itemsData = append(itemsData, []interface{}{item.MarketHashName})
	}

	if err := sheetsSvc.Write(
		fmt.Sprintf("%s%d", cfg.ItemColumn, startRow),
		fmt.Sprintf("%s%d", cfg.ItemColumn, endRow),
		itemsData,
	); err != nil {
		logger.Error("Error writing items: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote items")

	amountData := make([][]interface{}, 0, len(items))
	for _, item := range items {
		amountData = append(amountData, []interface{}{item.Amount})
	}

	if err := sheetsSvc.Write(
		fmt.Sprintf("%s%d", cfg.AmountColumn, startRow),
		fmt.Sprintf("%s%d", cfg.AmountColumn, endRow),
		amountData,
	); err != nil {
		logger.Error("Error writing amounts: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote amounts")

	singlePriceData := make([][]interface{}, 0, len(items))
	for _, item := range items {
		singlePriceStr := format.FormatPricePrintable(
			item.Price,
			cfg.DecimalSeparator,
			currencySign,
		)
		singlePriceData = append(singlePriceData, []interface{}{singlePriceStr})
	}

	if err := sheetsSvc.Write(
		fmt.Sprintf("%s%d", cfg.SinglePriceColumn, startRow),
		fmt.Sprintf("%s%d", cfg.SinglePriceColumn, endRow),
		singlePriceData,
	); err != nil {
		logger.Error("Error writing single prices: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote single prices")

	totalPriceData := make([][]interface{}, 0, len(items))
	for _, item := range items {
		totalPriceStr := format.FormatPricePrintable(
			item.Price*float64(item.Amount),
			cfg.DecimalSeparator,
			currencySign,
		)
		totalPriceData = append(totalPriceData, []interface{}{totalPriceStr})
	}

	if err := sheetsSvc.Write(
		fmt.Sprintf("%s%d", cfg.TotalPriceColumn, startRow),
		fmt.Sprintf("%s%d", cfg.TotalPriceColumn, endRow),
		totalPriceData,
	); err != nil {
		logger.Error("Error writing total prices: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote total prices")

	newTotal := 0.0
	for _, item := range items {
		newTotal += item.Price * float64(item.Amount)
	}

	difference := newTotal - preRunData.Total

	totalStr := format.FormatPricePrintable(newTotal, cfg.DecimalSeparator, currencySign)
	differenceStr := format.FormatPricePrintable(difference, cfg.DecimalSeparator, currencySign)

	if err := sheetsSvc.Write(cfg.TotalValueCell, cfg.TotalValueCell, [][]interface{}{{totalStr}}); err != nil {
		logger.Error("Error writing total value cell: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote total value cell")

	if err := sheetsSvc.Write(cfg.DifferenceCell, cfg.DifferenceCell, [][]interface{}{{differenceStr}}); err != nil {
		logger.Error("Error writing difference cell: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote difference cell")

	if err := sheetsSvc.Write(cfg.ErrorCell, cfg.ErrorCell, [][]interface{}{{lastRunNoError}}); err != nil {
		logger.Error("Error writing error cell: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote error cell")

	if err := sheetsSvc.Write(cfg.LastUpdatedCell, cfg.LastUpdatedCell, [][]interface{}{{time.Now().Format(lastUpdatedFormat)}}); err != nil {
		logger.Error("Error writing total value cell: %v", err)
		os.Exit(1)
	}

	logger.Debug("wrote last updated cell")

	logger.Info("Wrote data to spreadsheet")

	logger.Info("App exit")

	logger.Debug("run took %v", time.Since(startTime))
}

const (
	appMessage = "steamquery by devusSs - keep track of your Steam CS2 inventory"
)

var (
	buildVersion   string
	buildDate      string
	buildGitCommit string
)

func init() {
	if buildVersion == "" {
		buildVersion = "dev"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}
	if buildGitCommit == "" {
		buildGitCommit = "unknown"
	}

	flag.CommandLine.SortFlags = true
	flag.Usage = printHelp
}

func printHelp() {
	fmt.Println(appMessage)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ./steamquery [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Println(appMessage)
	fmt.Println()
	fmt.Printf("Build version:\t\t%s\n", buildVersion)
	fmt.Printf("Build date:\t\t%s\n", buildDate)
	fmt.Printf("Build Git commit:\t%s\n", buildGitCommit)
	fmt.Println()
	fmt.Printf("Build Go version:\t%s\n", runtime.Version())
	fmt.Printf("Build Go os:\t\t%s\n", runtime.GOOS)
	fmt.Printf("Build Go arch:\t\t%s\n", runtime.GOARCH)
}

var (
	supportedOS   = []string{"macOS", "Windows", "Linux"}
	supportedArch = []string{"x86_64", "amd64", "aarch64", "arm64"}
)

func checkOSAndArchComp() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	osV, err := system.GetOS(ctx)
	if err != nil {
		return fmt.Errorf("getting os: %w", err)
	}

	archV, err := system.GetArch(ctx)
	if err != nil {
		return fmt.Errorf("getting arch: %w", err)
	}

	if !slices.Contains(supportedOS, osV) {
		return fmt.Errorf("unsupported os: %s", osV)
	}

	if !slices.Contains(supportedArch, archV) {
		return fmt.Errorf("unsupported arch: %s", archV)
	}

	if osV == "Windows" && archV != "x86_64" || osV == "Windows" && archV != "amd64" {
		return fmt.Errorf("unsupported arch: %s", archV)
	}

	return nil
}

type inventoryItem struct {
	MarketHashName string
	Amount         int
	Price          float64
}

type preRunData struct {
	LastUpdated time.Time
	Error       string
	Total       float64
}

const (
	lastUpdatedFormat  = time.RFC3339Nano
	lastUpdateCooldown = 5 * time.Minute
	lastRunNoError     = "No error occured."
)

func fetchPreRunData(
	svc *tables.SpreadsheetService,
	cfg *config.Config,
	currencySign string,
) (*preRunData, error) {
	luRaw, err := svc.Read(cfg.LastUpdatedCell, cfg.LastUpdatedCell)
	if err != nil {
		return nil, fmt.Errorf("getting last updated cell: %w", err)
	}

	var lu time.Time
	if len(luRaw.Values) > 0 {
		lu, err = time.Parse(lastUpdatedFormat, luRaw.Values[0][0].(string))
		if err != nil {
			return nil, fmt.Errorf("parsing last updated cell: %w", err)
		}
	}

	errRaw, err := svc.Read(cfg.ErrorCell, cfg.ErrorCell)
	if err != nil {
		return nil, fmt.Errorf("getting error cell: %w", err)
	}

	var errStr string
	if len(errRaw.Values) > 0 {
		errStr = errRaw.Values[0][0].(string)
	}

	totalRaw, err := svc.Read(cfg.TotalValueCell, cfg.TotalValueCell)
	if err != nil {
		return nil, fmt.Errorf("getting total value cell: %w", err)
	}

	var total float64
	if len(totalRaw.Values) > 0 {
		totalStr := format.FormatPriceCalculatable(
			totalRaw.Values[0][0].(string),
			cfg.DecimalSeparator,
			currencySign,
		)
		total, err = strconv.ParseFloat(totalStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing total value cell: %w", err)
		}
	}

	return &preRunData{
		LastUpdated: lu,
		Error:       errStr,
		Total:       total,
	}, nil
}
