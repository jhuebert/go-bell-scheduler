package main

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
	"os"
	"strings"
)

type BellSchedule map[string]cron.EntryID

type scheduleDifference struct {
	commonKeys  []string
	onlyOnLeft  []string
	onlyOnRight []string
}

func GetUpdateScheduleFunc(c *cron.Cron, cronPath string, scheduleMap BellSchedule, bellFunc func()) func() {
	return func() {
		log.Info("Updating schedule")
		updateSchedule(c, cronPath, scheduleMap, bellFunc)
	}
}

func updateSchedule(c *cron.Cron, cronPath string, currentSchedule BellSchedule, bellFunc func()) {

	// Read the updated patterns from the cron file
	fileSchedule, err := readSchedule(cronPath)
	if err != nil {
		log.Errorf("Unable to read %q - %v", cronPath, err)
		return
	}

	// Calculate the difference in patterns
	difference := getDifference(currentSchedule, fileSchedule)

	// Remove any scheduled patterns that no longer exist in the cron file
	for _, key := range difference.onlyOnLeft {
		log.Infof("Removing %q", key)
		c.Remove(currentSchedule[key])
	}

	// Schedule new patterns in the cron file
	for _, key := range difference.onlyOnRight {

		id, err := c.AddFunc(key, bellFunc)
		if err != nil {
			log.Errorf("Unable to add %q", key)
			continue
		}

		log.Infof("Added %q", key)
		log.Infof("Next execution is %v", c.Entry(id).Next)
		currentSchedule[key] = id
	}
}

func readSchedule(cronPath string) (BellSchedule, error) {

	lines, err := readLines(cronPath)
	if err != nil {
		log.Errorf("Error while reading cron file - %v", err)
		return nil, err
	}

	schedule := BellSchedule{}
	for _, line := range lines {
		item := strings.TrimSpace(line)
		if (len(item) < 1) || strings.HasPrefix(item, "#") {
			continue
		}

		fields := strings.Fields(item)
		if len(fields) > 6 {
			log.Debugf("Format contains more than 6 fields. Retaining only first 6.")
			item = strings.Join(fields[0:6], " ")
		}

		schedule[item] = 1
	}

	return schedule, nil
}

func getDifference(left BellSchedule, right BellSchedule) *scheduleDifference {

	difference := scheduleDifference{}

	for key := range left {
		_, ok := right[key]
		if !ok {
			difference.onlyOnLeft = append(difference.onlyOnLeft, key)
		} else {
			difference.commonKeys = append(difference.commonKeys, key)
		}
	}

	for key := range right {
		_, ok := left[key]
		if !ok {
			difference.onlyOnRight = append(difference.onlyOnRight, key)
		}
	}

	return &difference
}

func readLines(path string) ([]string, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Warnf("Error closing file - %v", err)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
