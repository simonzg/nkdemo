package simulator

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"nakama/server"
)

type SyncRecord struct {
	id        string
	timestamp int64
	direction Direction
	hash      string
}

type ExitCode int

const (
	CodeNone ExitCode = iota
	CodeNormal
	CodeSendError
	CodeRecvError
)

func (ec ExitCode) String() string {
	return [...]string{"CodeNone", "CodeNormal", "CodeSendError", "CodeRecvError"}[ec]
}

type Direction int

const (
	IN Direction = iota
	OUT
)

func (d Direction) String() string {
	return [...]string{"IN", "OUT"}[d]
}

type SyncWorker struct {
	logger     *zap.Logger
	client     *NKClient
	stopped    bool
	exitCode   ExitCode
	recvCh     chan *server.Envelope
	doneCh     chan struct{}
	sendTicker *time.Ticker
	interval   int64
	outDir     string

	name     string
	customID string

	sendCount int64
	recvCount int64
	records   []*SyncRecord
}

func NewSyncWorker(logger *zap.Logger, logDir, customID, name string, outDir string, serverHost string, port int, interval int64) *SyncWorker {
	client := NewNKClient(logger, serverHost, port)
	l := logger.With(zap.String("customID", customID), zap.String("component", "worker"))
	offset := rand.Intn(5)
	sign := rand.Intn(1)
	if sign == 0 {
		offset = -offset
	}
	interval = interval + int64(offset)
	return &SyncWorker{
		logger:   l,
		client:   client,
		stopped:  false,
		recvCh:   make(chan *server.Envelope, 1000),
		doneCh:   make(chan struct{}),
		outDir:   outDir,
		interval: interval,

		name:     name,
		customID: customID,

		sendCount: 0,
		recvCount: 0,
		records:   make([]*SyncRecord, 0),
	}
}

func (w *SyncWorker) Name() string {
	return w.name
}

func (w *SyncWorker) ExitCode() ExitCode {
	return w.exitCode
}

func (w *SyncWorker) Records() []*SyncRecord {
	return w.records
}

func (w *SyncWorker) Counts() (int64, int64) {
	return w.sendCount, w.recvCount
}

func (w *SyncWorker) recvPump() {
	for {
		_, data, err := w.client.conn.ReadMessage()
		if err != nil {
			if w.stopped {
				return
			}
			fmt.Println("Read Error: ", err)
			w.Shutdown(CodeRecvError)
			return
		}

		if w.stopped {
			return
		}

		e := &server.Envelope{}
		// fmt.Println("Recved: ", )
		err = proto.Unmarshal([]byte(data), e)
		if err != nil {
			fmt.Println("Unmarshal error", err)
		}

		w.recvCh <- e
		w.logger.Info("Recved: ", zap.String("data", string(data)))
	}
}

// send out message on ticker
func (w *SyncWorker) mainLoop() {
	id := 0
	for {
		select {
		case <-w.doneCh:
			return
		case <-w.sendTicker.C:
			hash := fmt.Sprintf("msg-%s-%d", w.name, id+1)
			id += 1
			e := &server.Envelope{CollationId: hash}
			err := w.client.Send(e)

			if err != nil {
				if w.stopped {
					return
				}
				fmt.Println("Send Error:", err)
				w.Shutdown(CodeSendError)
				return
			} else {
				// monitor the outgoing request
				w.processSent(e)
			}
		case e := <-w.recvCh:
			w.processRecv(e)
		}

	}
}

func (w *SyncWorker) Init() error {
	// fmt.Println(w.name + " => Init")
	err := w.client.Connect(w.customID)
	if err != nil {
		fmt.Println("Connect Error:", err)
		return err
	}

	return nil
}

func (w *SyncWorker) Start() {
	w.sendTicker = time.NewTicker(time.Millisecond * time.Duration(w.interval))

	// sending coroutine
	go w.recvPump()
	// fmt.Println(w.name + " => Started")
	w.mainLoop()
}

func (w *SyncWorker) StopSend() {
	if w.sendTicker != nil {
		w.sendTicker.Stop()
	}
}

func (w *SyncWorker) Shutdown(code ExitCode) {
	if w.stopped {
		return
	}
	if code > CodeNormal {
		fmt.Println("Shutdown early with", code)
	}
	w.stopped = true
	w.exitCode = code
	w.saveToFile()
	w.client.Stop()

	close(w.doneCh)
	// fmt.Println(w.name + " => Stopped")
}

// receive message
func (w *SyncWorker) processRecv(e *server.Envelope) {
	hash := e.CollationId
	if hash == "" {
		// fmt.Println("empty envelope, skip")
		return
	}
	w.logger.Info("Record Recv: ", zap.String("hash", hash))

	// only update records when receive SpacePresence message
	w.recvCount += 1
	w.appendRecord(&SyncRecord{
		id:        w.customID,
		direction: IN,
		hash:      e.CollationId,
	})
}

func (w *SyncWorker) processSent(e *server.Envelope) {
	hash := e.CollationId
	w.logger.Info("Record Sent: ", zap.String("hash", hash))

	w.sendCount += 1
	w.appendRecord(&SyncRecord{
		id:        w.customID,
		direction: OUT,
		hash:      hash,
	})
}

func (w *SyncWorker) appendRecord(record *SyncRecord) {
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	record.timestamp = millis
	w.records = append(w.records, record)
}

func (w *SyncWorker) saveToFile() {
	f, err := os.Create(fmt.Sprintf("%s/%s.csv", w.outDir, w.customID))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	writer.Write([]string{"Direction", "Timestamp", "Hash"})
	for _, r := range w.records {
		writer.Write([]string{r.direction.String(), fmt.Sprintf("%d", r.timestamp), r.hash})
	}
}
