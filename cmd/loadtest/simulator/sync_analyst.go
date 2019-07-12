package simulator

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"sort"
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
	data       map[string]*TsRecord
	latencyMap map[int64]float64
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
	sentAt      int64
	recvRecords []*RecvRecord
}

func NewSyncAnalyst(count, errorCount int) *SyncAnalyst {
	analyst := &SyncAnalyst{
		data:       make(map[string]*TsRecord),
		latencyMap: make(map[int64]float64),
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
		tss, ok := a.data[hash]
		if r.direction == IN {
			// recv in
			a.totalRecv += 1
			if ok {
				a.data[hash].recvRecords = append(tss.recvRecords, &RecvRecord{recvAt: r.timestamp, id: r.id})
			} else {
				a.data[hash] = &TsRecord{recvRecords: []*RecvRecord{&RecvRecord{recvAt: r.timestamp, id: r.id}}}
			}
		} else if r.direction == OUT {
			// send out
			a.totalSent += 1
			if ok {
				a.data[hash].sentAt = r.timestamp
			} else {
				a.data[hash] = &TsRecord{recvRecords: make([]*RecvRecord, 0), sentAt: r.timestamp}
			}
		}
	}
}

func (a *SyncAnalyst) analyze(numOfAccounts int, ids []string) {

	missing := make(map[string]map[string]bool)
	for hash, tsRecord := range a.data {
		if tsRecord == nil || tsRecord.sentAt == 0 {
			continue
		}
		if len(tsRecord.recvRecords) == 0 {
			continue
		}

		n := len(tsRecord.recvRecords)
		sentAt := tsRecord.sentAt
		var sum int64
		for _, r := range tsRecord.recvRecords {
			sum += r.recvAt - sentAt
		}
		avg := float64(sum) / float64(n)
		a.latencyMap[sentAt] = avg

		if len(tsRecord.recvRecords) < numOfAccounts {
			idMap := make(map[string]bool)
			for _, id := range ids {
				idMap[id] = true
			}
			for _, rr := range tsRecord.recvRecords {
				delete(idMap, rr.id)
			}
			missing[hash] = idMap
		}
	}

	fmt.Println("----Missing----")
	for hash, idMap := range missing {
		fmt.Println(hash, idMap)
	}
	fmt.Println("---------------")

	avgs := make([]float64, 0)
	for _, avg := range a.latencyMap {
		avgs = append(avgs, avg)
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

	latencyWriter.Write([]string{"Sent_Timestamp", "Average_Duration"})
	for ts, avgDuration := range a.latencyMap {
		latencyWriter.Write([]string{fmt.Sprintf("%d", ts), fmt.Sprintf("%.2f", avgDuration)})
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
