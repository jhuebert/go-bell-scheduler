package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

// Program information to print on each invocation.
const version string = "Bell Scheduler 3.0.0"

func main() {
	log.Info(version)

	bellPath := flag.String("bell", "", "path to bell sound")
	cronPath := flag.String("cron", "", "path to bell cron")
	logLevel := flag.String("log", "info", "logging level. Can be one of [trace, debug, info, warn, error, fatal, panic]")
	loopCount := flag.Int("loops", 1, "number of times to play the bell sound in succession")
	updateScheduleSeconds := flag.Int("update", 60, "number of seconds delay between checking the cron schedule for updates")
	ntpPeriod := flag.Int("ntp-period", 3600, "number of seconds delay between fetching the latest network time")
	ntpUrl := flag.String("ntp-url", "time.nist.gov", "number of seconds delay between fetching the latest network time")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Warnf("Invalid input log level - %q", *logLevel)
		log.Warnf("Setting log level to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)

	if *bellPath == "" {
		log.Error("Path to bell sound file required")
		flag.Usage()
		return
	} else if !fileExists(*bellPath) {
		log.Errorf("Bell sound file does not exist - %q", *bellPath)
		flag.Usage()
		return
	}

	if *cronPath == "" {
		log.Error("Path to cron file required")
		flag.Usage()
		return
	} else if !fileExists(*cronPath) {
		log.Errorf("Cron file does not exist - %q", *cronPath)
		flag.Usage()
		return
	}

	if *loopCount < 1 {
		log.Error("Number of loops must be at least 1")
		flag.Usage()
		return
	}

	if *updateScheduleSeconds < 1 {
		log.Error("Number of seconds between schedule updates must be at least 1")
		flag.Usage()
		return
	}

	if *ntpPeriod < 1 {
		log.Error("Number of seconds between network time updates must be at least 1")
		flag.Usage()
		return
	}

	if *ntpUrl == "" {
		log.Error("NTP URL required")
		flag.Usage()
		return
	}

	// Create the cron scheduler
	scheduleMap := make(BellSchedule)
	c := New()

	// Create the function that plays the bell
	bellFunc := GetPlayBellFunc(*bellPath, *loopCount)

	// Add the time offset update function to be executed periodically
	updateTimeOffsetFunc := GetUpdateTimeOffsetFunc(c, *ntpUrl)
	_, err = c.AddFunc("@every "+strconv.Itoa(*ntpPeriod)+"s", updateTimeOffsetFunc)
	if err != nil {
		log.Error("Error adding time offset updater - %v", err)
		return
	}

	// Add the schedule update function to be executed periodically
	updateScheduleFunc := GetUpdateScheduleFunc(c, *cronPath, scheduleMap, bellFunc)
	_, err = c.AddFunc("@every "+strconv.Itoa(*updateScheduleSeconds)+"s", updateScheduleFunc)
	if err != nil {
		log.Error("Error adding schedule updater - %v", err)
		return
	}

	// Execute the updater functions once immediately
	updateTimeOffsetFunc()
	updateScheduleFunc()

	// Start the scheduler
	c.Start()

	// Wait indefinitely
	select {}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
