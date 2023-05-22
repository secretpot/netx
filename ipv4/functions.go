package ipv4

import (
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/secretpot/jbutil/errs"
)

func Localhost() (ip string) {
	conn, err := net.Dial("udp", "88.88.88.88:8888")
	errs.ERQ(err)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func LocalhostName() (hostname string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return
}

func ParseTracerouteOutput(output string) []EchoSummary {
	parseResult := []EchoSummary{}
	outputLines := strings.Split(output, "\n")

	for _, line := range outputLines[1 : len(outputLines)-1] {
		data := strings.Fields(line)

		ttl, err := strconv.Atoi(data[0])
		if err != nil {
			ttl = -1
		}
		traceIP := data[1]
		rtt := []float64{}
		if traceIP == "*" {
			for i := 0; i < len(data)-1; i++ {
				rtt = append(rtt, 0.0)
			}
		} else {
			for i := 2; i < len(data)-1; i++ {
				if t, err := strconv.ParseFloat(data[i], 64); err != nil {
					rtt = append(rtt, 0.0)
				} else {
					rtt = append(rtt, t)
				}
			}
		}
		parseResult = append(parseResult, NewEchoSummary(ttl, traceIP, rtt))
		sort.Slice(parseResult, func(i, j int) bool {
			return parseResult[i].ttl < parseResult[j].ttl
		})
	}
	return parseResult
}
