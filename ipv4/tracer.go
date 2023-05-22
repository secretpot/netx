package ipv4

import (
	"fmt"
	"log"
	"sync"

	"github.com/secretpot/jbutil/errs"
	jmath "github.com/secretpot/jbutil/math"

	"os/exec"
)

func trobeCore(domain string, pCount int, waitTime float64, firstTTL int, maxTTL int, sQueries int) ([]byte, error) {
	// traceroute on MacOS doesnot support non-integer waittime
	// it will cost ((hop count // squeries + 1) * waitTime) seconds
	N := fmt.Sprintf("-N %d", sQueries)
	q := fmt.Sprintf("-q %d", pCount)
	w := fmt.Sprintf("-w %.f", jmath.Round(waitTime, 2))

	param := []string{"-nI", q, w, N}
	if firstTTL > 0 {
		param = append(param, fmt.Sprintf("-f %d", firstTTL))
	}
	if maxTTL > 0 {
		param = append(param, fmt.Sprintf("-m %d", maxTTL))
	}
	param = append(param, domain)
	cmd := exec.Command("traceroute", param...)
	return cmd.Output()
}

func Trobe(domain string, pCount int, waitTime float64, firstTTL int, maxTTL int, sQueries int) []EchoSummary {
	/* traceroute probe abbr. */
	output, err := trobeCore(domain, pCount, waitTime, firstTTL, maxTTL, sQueries)
	errs.ERQ(err)
	return ParseTracerouteOutput(string(output))
}
func SimpleTrobe(domain string, pCount int, waitTime float64) []EchoSummary {
	return Trobe(domain, pCount, waitTime, 0, 0, 16)
}
func TrobePoint(domain string, pCount int, waitTime float64, ttl int) (s EchoSummary) {
	defer func() {
		err := recover()
		if err != nil {
			log.Fatalln(err)
			data := []float64{}
			for i := 0; i < pCount; i++ {
				data = append(data, 0.0)
			}
			s = NewEchoSummary(ttl, "*", data)
		}
	}()
	s = Trobe(domain, pCount, waitTime, ttl, ttl, 1)[0]
	return
}
func SimpleTrobeWithEnsurance(domain string, pCount int, waitTime float64, ensurance int) []EchoSummary {
	data := SimpleTrobe(domain, pCount, waitTime)
	wg := new(sync.WaitGroup)
	for index := range data {
		wg.Add(1)
		go func(summary *EchoSummary, ttl int) {
			defer wg.Done()
			for i := 0; i < ensurance-1 && summary.traceIP == "*"; i++ {
				summary.traceIP = TrobePoint(domain, pCount, waitTime, ttl).traceIP
			}
		}(&data[index], index+1)
	}
	wg.Wait()
	return data
}
