package main

import (
	"bufio"
	"fmt"
	"github.com/beevik/ntp"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

type BellSchedule map[string]EntryID

type scheduleDifference struct {
	commonKeys  []string
	onlyOnLeft  []string
	onlyOnRight []string
}

func GetUpdateScheduleFunc(c *Cron, cronPath string, scheduleMap BellSchedule, bellFunc func()) func() {
	return func() {
		log.Info("Updating schedule")
		updateSchedule(c, cronPath, scheduleMap, bellFunc)
	}
}

func updateSchedule(c *Cron, cronPath string, currentSchedule BellSchedule, bellFunc func()) {

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

func updateEntryOffset(c *Cron) {

	externalTime, err := ntp.Time("time.nist.gov")
	if err != nil {
		log.Warnf("Unable to get network time - %v", err)
		return
	}

	// Must be immediately after return of network time to minimize the amount of difference
	// TODO Could potentially use channel to synchronize
	internalTime := time.Now()

	// Calculate the difference in time between the local machine and network
	difference := internalTime.Sub(externalTime)

	fmt.Println(externalTime)
	fmt.Println(internalTime)
	fmt.Println(difference)

	//entries := c.Entries()

	//for _, e := range entries {
	//e.Next = nil
	//}

	// TODO This will not work as it gives us a copy of the entries

	//TODO Update next scheduled time for each entry

	// Call Entries fetch one more time so that the updated timestamps will be applied
	c.Entries()
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
