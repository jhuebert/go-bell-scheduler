package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
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
	loopCount := flag.Int64("loops", 1, "number of times to play the bell sound in succession")
	updateScheduleSeconds := flag.Int64("update", 60, "number of seconds delay between checking the cron schedule for updates")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Warnf("Invalid input log level - %v", *logLevel)
		log.Warnf("Setting log level to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)

	if *bellPath == "" {
		log.Error("Path to bell sound file required")
		flag.Usage()
		return
	} else if !fileExists(*bellPath) {
		log.Errorf("Bell sound file does not exist - %v", *bellPath)
		flag.Usage()
		return
	}

	if *cronPath == "" {
		log.Error("Path to cron file required")
		flag.Usage()
		return
	} else if !fileExists(*cronPath) {
		log.Errorf("Cron file does not exist - %v", *cronPath)
		flag.Usage()
		return
	}

	if *loopCount < 1 {
		log.Error("Number of loops must be at least 1")
		flag.Usage()
		return
	}

	if *updateScheduleSeconds < 1 {
		log.Error("Number of seconds between sechdule updates must be at least 1")
		flag.Usage()
		return
	}

	// Create the cron scheduler
	scheduleMap := make(BellSchedule)
	c := cron.New()

	// Create the function that plays the bell
	bellFunc := GetPlayBellFunc(*bellPath)

	// Create the function that updates the schedule
	updateFunc := GetUpdateScheduleFunc(c, *cronPath, scheduleMap, bellFunc)

	// Add the update function to be executed periodically
	_, err = c.AddFunc("@every "+strconv.FormatInt(*updateScheduleSeconds, 10)+"s", updateFunc)
	if err != nil {
		log.Error("Error adding schedule updater - %v", err)
		return
	}

	// Start the scheduler
	c.Start()

	// Wait indefinitely
	select {}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
