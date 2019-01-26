package main

import (
	"flag"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	log "github.com/sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
	"os"
	"time"
)

// Program information to print on each invocation.
const version string = "Bell Scheduler 3.0.0"

// Schedule that defines when the cron file should be checked for changes.
const updateScheduleSeconds int64 = 60

func main() {
	log.Info(version)

	logLevel := flag.String("log", "info", "logging level. Can be one of [trace, debug, info, warn, error, fatal, panic]")
	soundPath := flag.String("bell", "", "path to bell sound file")
	cronPath := flag.String("cron", "", "path to bell cron file")
	loopCount := flag.Int64("loops", 1, "number of times to play the bell sound file")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Warnf("Invalid input log level - %v", *logLevel)
		log.Warnf("Setting log level to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)

	if *soundPath == "" {
		log.Error("Path to bell sound file required")
		flag.Usage()
		return
	} else if !fileExists(*soundPath) {
		log.Errorf("Bell sound file does not exist - %v", *soundPath)
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

	// Create the cron scheduler
	c := cron.New()

	c.AddFunc("0/5 * * * * *", func() { fmt.Println(time.Now()) })
	//c.AddFunc("TZ=Asia/Tokyo 30 04 * * * *", func() { fmt.Println("Runs at 04:30 Tokyo time every day") })
	//c.AddFunc("@hourly",      func() { fmt.Println("Every hour") })
	//c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
	c.Start()

	select {}

	///* Create the cron scheduler */
	//Scheduler scheduler = StdSchedulerFactory.getDefaultScheduler();
	//
	///* Create the job that schedules tasks based on the cron file */
	//JobDetail reschedulerJob = newJob(BellRescheduler.class)
	//.withIdentity(RESCHEDULER_JOB_KEY)
	//.build();
	//reschedulerJob.getJobDataMap().put(CRON_FILE, bellCronFile);
	//reschedulerJob.getJobDataMap().put(SOUND_FILE, bellFile);
	//reschedulerJob.getJobDataMap().put(NUM_LOOPS, loops);
	//
	///* Get the number of seconds between schedule updates */
	//Integer scheduleUpdateSeconds =
	//	Ints.tryParse(System.getProperty("bellScheduler.scheduleUpdateSeconds", String.valueOf(UPDATE_SCHEDULE_SECONDS)));
	//if (scheduleUpdateSeconds == null) {
	//	scheduleUpdateSeconds = UPDATE_SCHEDULE_SECONDS;
	//}
	//
	///* Create the trigger to check for schedule updates */
	//Trigger reschedulerTrigger = newTrigger()
	//.withIdentity(RESCHEDULER_TRIGGER_KEY)
	//.startNow()
	//.withSchedule(simpleSchedule()
	//.withIntervalInSeconds(scheduleUpdateSeconds)
	//.repeatForever())
	//.build();
	//
	///* Add the reschedule job/trigger */
	//scheduler.scheduleJob(reschedulerJob, reschedulerTrigger);
	//
	///* Start the scheduler */
	//scheduler.start();

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func playBell(path string) error {
	f, err := os.Open(path)
	if err != nil {
		log.Errorf("File not found - %v", err)
		return err
	}

	s, format, err := wav.Decode(f)
	if err != nil {
		log.Errorf("Cannot decode sound file - %v", err)
		return err
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Errorf("Cannot initialize speaker - %v", err)
		return err
	}

	done := make(chan struct{})

	t := beep.Loop(5, s)

	speaker.Play(beep.Seq(t, beep.Callback(func() {
		close(done)
	})))

	<-done

	return nil
}
