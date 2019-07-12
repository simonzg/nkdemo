package simulator

import (
	"os"

	"github.com/satori/go.uuid"
	"math"
	// "reflect"
	// "strconv"
	"fmt"
)

const (
	TS_FORMAT = "2006-01-02 15:04:05"
)

func GetUUID() string {
	return uuid.NewV4().String()
}

func GetCurDir() string {
	curDir, _ := os.Getwd()
	return curDir
}

type Defaults struct {
	Verbose   bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Server    string `short:"s" long:"server" description:"server host address/domain" default:"localhost"`
	Port      int    `short:"p" long:"port" description:"server port number" default:"8888"`
	ServerKey string `short:"k" long:"serverkey" description:"server key" default:"defaultkey"`
}

func calcPxxLatency(latencies []float64) (string, string, string, string, string, string, string, string) {
	n := float64(len(latencies))
	var sum float64
	for _, l := range latencies {
		sum += l
	}
	if len(latencies) == 0 {
		return "0", "0", "0", "0", "0", "0", "0", "0"
	}

	avg := fmt.Sprintf("%.2f", sum/float64(n))
	i15 := int(math.Floor(n * float64(0.15)))
	p15 := fmt.Sprintf("%.2f", latencies[i15])
	i25 := int(math.Floor(n * float64(0.25)))
	p25 := fmt.Sprintf("%.2f", latencies[i25])
	i50 := int(math.Floor(n * float64(0.50)))
	p50 := fmt.Sprintf("%.2f", latencies[i50])
	i75 := int(math.Floor(n * float64(0.75)))
	p75 := fmt.Sprintf("%.2f", latencies[i75])
	i90 := int(math.Floor(n * float64(0.90)))
	p90 := fmt.Sprintf("%.2f", latencies[i90])
	i95 := int(math.Floor(n * float64(0.95)))
	p95 := fmt.Sprintf("%.2f", latencies[i95])
	i99 := int(math.Floor(n * float64(0.99)))
	p99 := fmt.Sprintf("%.2f", latencies[i99])
	return avg, p15, p25, p50, p75, p90, p95, p99
}
