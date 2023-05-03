package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"encoding/json"

	probing "github.com/prometheus-community/pro-bing"
)

const (
	testInfoFileName = "./test_info.json"
)

type testInfo struct {
	BuildInfo struct {
		BuildVersion string `json:"build_version"`
		BuildDate    string `json:"build_date"`
		BuildOS      string `json:"build_os"`
		BuildArch    string `json:"build_arch"`
		GoVersion    string `json:"go_version"`
	} `json:"build_info"`
	Systeminfo struct {
		CPUCount        int    `json:"cpu_count"`
		CGOCalls        int64  `json:"cgo_calls"`
		GoRoutinesCount int    `json:"goroutines_count"`
		Pagesize        int    `json:"pagesize"`
		ProcessID       int    `json:"process_id"`
		PathInfo        string `json:"path_info"`
		HostInfo        string `json:"host_info"`
		AvgRTT          string `json:"avg_rtt"`
	} `json:"system_info"`
	AppInfo struct {
		LogsExist         bool   `json:"default_logs_dir_exists"`
		FilesExist        bool   `json:"default_files_dir_exists"`
		UsingBeta         bool   `json:"using_beta"`
		UsingConfig       string `json:"using_config"`
		UsingGCloudConfig string `json:"using_gcloud_config"`
	} `json:"app_info"`
}

func dirExists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}
	return true
}

func pingForNetworkTest() (int64, error) {
	pinger, err := probing.NewPinger("steamcommunity.com")
	if err != nil {
		return 0, err
	}

	if buildOS == "windows" {
		pinger.SetPrivileged(true)
	}

	if buildOS == "linux" {
		pinger.SetPrivileged(true)
		log.Printf("[%s] Please set following command:\n", warnSign)
		exePath, err := os.Executable()
		if err != nil {
			return 0, err
		}
		log.Printf("[%s] %s\n", warnSign, fmt.Sprintf("setcap cap_net_raw=+ep %s", exePath))

		log.Printf("Did you enter that command? (y/n): ")
		inputReader := bufio.NewReader(os.Stdin)
		userInput, _ := inputReader.ReadString('\n')

		if userInput != "y" {
			return 0, errors.New("cannot execute ping command without privileges")
		}
	}

	pinger.Count = 3

	log.Printf("[%s] Test pinging %s\n", infSign, pinger.Addr())

	err = pinger.Run()
	if err != nil {
		return 0, err
	}

	stats := pinger.Statistics()

	return stats.AvgRtt.Milliseconds(), nil
}

func printTestInfo(usingBeta bool, cfgPath, gCloudPath string, avgRTT int64) {
	log.Printf("[%s] CPU Cores (available): \t%d\n", infSign, runtime.NumCPU())
	log.Printf("[%s] CGO calls: \t\t%d\n", infSign, runtime.NumCgoCall())
	log.Printf("[%s] Goroutines: \t\t%d\n", infSign, runtime.NumGoroutine())
	log.Printf("[%s] Pagesize: \t\t%d\n", infSign, os.Getpagesize())
	log.Printf("[%s] Process ID: \t\t%d\n", infSign, os.Getpid())
	log.Printf("[%s] AVG RTT to Steam: \t%d ms\n", infSign, avgRTT)

	pPath, err := os.Getwd()
	if err != nil {
		log.Printf("[%s] Error querying path info: %s\n", errSign, err.Error())
	} else {
		log.Printf("[%s] Path info: \t\t%s\n", infSign, pPath)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("[%s] Error querying host info: %s\n", errSign, err.Error())
	} else {
		log.Printf("[%s] Host info: \t\t%s\n", infSign, hostname)
	}

	fmt.Println()

	log.Printf("[%s] Logs dir exists: \t%t\n", infSign, dirExists("./logs"))
	log.Printf("[%s] Files dir exists: \t%t\n", infSign, dirExists("./files"))
	log.Printf("[%s] Using beta mode: \t%t\n", infSign, usingBeta)

	fmt.Println()

	log.Printf("[%s] Using config: \t\t%s\n", infSign, cfgPath)
	log.Printf("[%s] Using gcloud config: \t%s\n", infSign, gCloudPath)
}

func saveTestInfoToFile(usingBeta bool, cfgPath, gCloudPath, avgRTT string) error {
	f, err := os.Create(testInfoFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	pPath, err := os.Getwd()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	var info testInfo

	info.BuildInfo.BuildVersion = buildVersion
	info.BuildInfo.BuildDate = buildDate
	info.BuildInfo.BuildOS = buildOS
	info.BuildInfo.BuildArch = buildArch
	info.BuildInfo.GoVersion = goVersion

	info.Systeminfo.CPUCount = runtime.NumCPU()
	info.Systeminfo.CGOCalls = runtime.NumCgoCall()
	info.Systeminfo.GoRoutinesCount = runtime.NumGoroutine()
	info.Systeminfo.Pagesize = os.Getpagesize()
	info.Systeminfo.ProcessID = os.Getpid()
	info.Systeminfo.PathInfo = pPath
	info.Systeminfo.HostInfo = hostname
	info.Systeminfo.AvgRTT = avgRTT

	info.AppInfo.LogsExist = dirExists("./logs")
	info.AppInfo.FilesExist = dirExists("./files")
	info.AppInfo.UsingBeta = usingBeta
	info.AppInfo.UsingConfig = cfgPath
	info.AppInfo.UsingGCloudConfig = gCloudPath

	return json.NewEncoder(f).Encode(&info)
}
