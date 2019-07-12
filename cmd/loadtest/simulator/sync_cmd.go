package simulator

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"
)

const MAX_SYNC_WORKER = 50

type SyncCmd struct {
	Defaults
	NumOfUsers   int   `short:"n" long:"nclients" description:"number of clients running at the same time" default:"10"`
	Duration     int   `short:"d" long:"duration" description:"test duration in seconds" default:"30"`
	SyncInterval int64 `short:"i" long:"interval" description:"sync interval defined with milliseconds" default:"100"`

	logger  *zap.Logger
	workers []*SyncWorker
	ticker  *time.Ticker
	seconds int64
	outDir  string

	wg sync.WaitGroup
}

func (cmd *SyncCmd) Execute(args []string) error {
	err := cmd.init()
	if err != nil {
		return err
	}
	cmd.start()
	return nil
}

func (cmd *SyncCmd) init() error {

	timestamp := time.Now().Format("0102-150405")
	folderName := fmt.Sprintf("n%d-%ds-%s", cmd.NumOfUsers, cmd.Duration, timestamp)
	outDir := path.Join(GetCurDir(), "log", "sync", folderName)
	logDir := path.Join(outDir, "log")
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		fmt.Println("Could not create folder: ", logDir)
	}
	cmd.logger = NewLogger(logDir, "master")
	fmt.Println(cmd.logger)
	cmd.outDir = outDir

	cmd.workers = make([]*SyncWorker, 0)
	i := 0
	for i < cmd.NumOfUsers {
		name := fmt.Sprintf("#%d", i+1)
		customID := fmt.Sprintf("sync-worker-%d", i+1)
		logger := NewLogger(logDir, customID)
		worker := NewSyncWorker(logger, outDir, customID, name, outDir, cmd.Server, cmd.Port, cmd.SyncInterval)
		cmd.workers = append(cmd.workers, worker)
		i++
	}
	return nil
}

func (cmd *SyncCmd) getStats() (int64, int64, int) {
	var totalSend, totalRecv int64
	errorCount := 0

	totalSend, totalRecv = 0, 0

	for _, worker := range cmd.workers {
		sendCount, recvCount := worker.Counts()
		totalSend += sendCount
		totalRecv += recvCount
		if worker.ExitCode() > CodeNormal {
			errorCount += 1
		}
	}

	return totalSend, totalRecv, errorCount
}

func (cmd *SyncCmd) start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	limiter := make(chan bool, MAX_SYNC_WORKER)
	fmt.Println("Initializing workers ...")
	for _, worker := range cmd.workers {
		cmd.wg.Add(1)

		go func(worker *SyncWorker) {
			limiter <- true
			defer func() {
				<-limiter
				cmd.wg.Done()
			}()
			worker.Init()
		}(worker)
	}
	cmd.wg.Wait()
	fmt.Println("Initialized.")

	fmt.Println("Start workers at ", time.Now().UnixNano()/int64(time.Millisecond))
	for _, worker := range cmd.workers {
		cmd.wg.Add(1)
		go worker.Start()
	}
	fmt.Println("Started.")

	time.AfterFunc(time.Second*time.Duration(cmd.Duration), func() {
		close(done)
	})
	cmd.mainLoop(interrupt, done)
}

func (cmd *SyncCmd) stop() {
	fmt.Println("Gracefully Stoping workers ...")
	for _, worker := range cmd.workers {
		go func(worker *SyncWorker) {
			worker.StopSend()
			time.AfterFunc(time.Second*5, func() {
				defer cmd.wg.Done()
				worker.Records()
				worker.Shutdown(CodeNormal)
			})
		}(worker)
	}
	cmd.wg.Wait()
	fmt.Println("Stopped.")

	_, _, errorCount := cmd.getStats()
	analyst := NewSyncAnalyst(cmd.NumOfUsers, errorCount)
	for _, worker := range cmd.workers {
		analyst.Add(worker.Records())
	}

	cmd.saveConfigToFile()
	ids := make([]string, 0)
	for _, worker := range cmd.workers {
		ids = append(ids, worker.customID)
	}
	analyst.analyze(cmd.NumOfUsers, ids)
	analyst.saveToFile(cmd.outDir)

	fmt.Println("Test Ended at ", time.Now().Format(TS_FORMAT))
	fmt.Println("Output directory: ", cmd.outDir)
}

func (cmd *SyncCmd) saveConfigToFile() {
	f, err := os.Create(cmd.outDir + "/" + "config.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	writer.Write([]string{fmt.Sprintf("%d", cmd.NumOfUsers), fmt.Sprintf("%d", cmd.Duration)})
	writer.Write([]string{"workers"})

}

func (cmd *SyncCmd) mainLoop(interrupt chan os.Signal, done chan struct{}) {
	secondTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-interrupt:
			fmt.Println("User Interrupt")
			cmd.stop()
			return
		case <-done:
			fmt.Println("Duration Interrupt")
			cmd.stop()
			return
		case <-secondTicker.C:
			cmd.seconds += 1
			totalSend, totalRecv, errorCount := cmd.getStats()
			cmd.logger.Info("Master Stat: ",
				zap.Int64("sent", totalSend),
				zap.Int64("recv", totalRecv),
				zap.Int("error_count", errorCount),
				zap.Int64("seconds", cmd.seconds))
		}
	}

}
