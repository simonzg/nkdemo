package simulator

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	PERCENT_15         = "P15_Latency(msec)"
	PERCENT_25         = "P25_Latency(msec)"
	PERCENT_50         = "P50_Latency(msec)"
	PERCENT_75         = "P75_Latency(msec)"
	PERCENT_90         = "P90_Latency(msec)"
	PERCENT_95         = "P95_Latency(msec)"
	PERCENT_99         = "P99_Latency(msec)"
	AVG                = "Avg_Latency(msec)"
	TOTAL_SENT         = "Total_Sent"
	TOTAL_RECV         = "Total_Recv"
	TOTAL_MISSING      = "Total_Missing"
	TOTAL_ERROR_CLIENT = "Total_Error_Client"
)

type SyncAnalyst struct {
	data       map[string][]*SyncRecord
	latencyMap map[TsSentRecord]TsRecvRecord
	results    map[string]string
	count      int

	totalSent int64
	totalRecv int64
}
type RecvRecord struct {
	id     string
	recvAt int64
}

type TsRecord struct {
	id          string
	sentAt      int64
	recvRecords []*RecvRecord
}
type TsSentRecord struct {
	id     string
	sentAt int64
	hash   string
}

type TsRecvRecord struct {
	recvRecords []*SyncRecord
	avgLatency  float64
}

func NewSyncAnalyst(count, errorCount int) *SyncAnalyst {
	analyst := &SyncAnalyst{
		data:       make(map[string][]*SyncRecord),
		latencyMap: make(map[TsSentRecord]TsRecvRecord), // sent info -
		results:    make(map[string]string),

		count:     count,
		totalSent: 0,
		totalRecv: 0,
	}
	analyst.results[TOTAL_ERROR_CLIENT] = fmt.Sprintf("%d", errorCount)
	return analyst
}

func (a *SyncAnalyst) Add(records []*SyncRecord) {
	for _, r := range records {
		hash := r.hash
		_, ok := a.data[hash]
		if !ok {
			a.data[hash] = make([]*SyncRecord, 0)
		}
		if r.direction == IN {
			a.totalRecv++
			a.data[hash] = append(a.data[hash], r)
		} else if r.direction == OUT {
			// send out
			a.totalSent++
			a.data[hash] = append([]*SyncRecord{r}, a.data[hash]...)
		}
	}
}

func (a *SyncAnalyst) analyze(numOfAccounts int, ids []string) {

	missing := make(map[string]bool)
	for hash, records := range a.data {
		if records == nil {
			continue
		}

		n := len(records)
		if records[0].direction != OUT {
			fmt.Println("Error data!")
			return
		}
		sentRecord := records[0]
		sentAt := sentRecord.timestamp
		recvCount := n - 1
		var sum int64
		for _, recvRecord := range records[1:] {
			sum += recvRecord.timestamp - sentAt
		}
		avg := float64(sum) / float64(recvCount)
		a.latencyMap[TsSentRecord{id: sentRecord.id, sentAt: sentRecord.timestamp, hash: hash}] = TsRecvRecord{recvRecords: records[1:], avgLatency: avg}

		if n < numOfAccounts {
			missing[hash] = true
		}
	}

	fmt.Println("----Missing----")
	for hash, idMap := range missing {
		fmt.Println(hash, idMap)
	}
	fmt.Println("---------------")

	avgs := make([]float64, 0)
	for _, recvRecord := range a.latencyMap {
		avgs = append(avgs, recvRecord.avgLatency)
	}

	sort.Float64s(avgs)

	avg, p15, p25, p50, p75, p90, p95, p99 := calcPxxLatency(avgs)

	a.results[TOTAL_SENT] = fmt.Sprintf("%d", a.totalSent)
	a.results[TOTAL_RECV] = fmt.Sprintf("%d", a.totalRecv)
	a.results[PERCENT_15] = p15
	a.results[PERCENT_25] = p25
	a.results[PERCENT_50] = p50
	a.results[PERCENT_75] = p75
	a.results[PERCENT_90] = p90
	a.results[PERCENT_95] = p95
	a.results[PERCENT_99] = p99
	a.results[TOTAL_MISSING] = fmt.Sprintf("%d", int64(a.count-1)*a.totalSent-a.totalRecv)
	a.results[AVG] = avg

	fmt.Println("--------------------------")
	fmt.Println("----- ANALYZE RESULT -----")
	fmt.Println("--------------------------")
	keys := make([]string, 0)
	for k, _ := range a.results {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := a.results[k]
		fmt.Println(k, "=", v)
	}
}

func (a *SyncAnalyst) saveToFile(outDir string) {
	latencyFilename := path.Join(outDir, "latency.csv")
	latencyFile, err := os.Create(latencyFilename)
	if err != nil {
		panic(err)
	}
	defer latencyFile.Close()

	latencyWriter := csv.NewWriter(latencyFile)
	defer latencyWriter.Flush()

	latencyWriter.Write([]string{"Hash", "Sent_By", "Sent_Timestamp", "Average_Duration", "Recv_By"})
	for sentRecord, recvRecord := range a.latencyMap {

		recvIds := make([]string, 0)
		for _, v := range recvRecord.recvRecords {
			recvIds = append(recvIds, fmt.Sprintf("%s(%d)", v.id, v.timestamp-sentRecord.sentAt))
		}
		recvBy := strings.Join(recvIds, ",")
		latencyWriter.Write([]string{sentRecord.hash, sentRecord.id, fmt.Sprintf("%d", sentRecord.sentAt), fmt.Sprintf("%.2f", recvRecord.avgLatency), recvBy})
	}

	resultFilename := path.Join(outDir, "result.csv")
	resultFile, err := os.Create(resultFilename)
	if err != nil {
		panic(err)
	}
	defer resultFile.Close()

	resultWriter := csv.NewWriter(resultFile)
	defer resultWriter.Flush()

	for k, v := range a.results {
		resultWriter.Write([]string{k, v})
	}
}
