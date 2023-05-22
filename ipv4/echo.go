package ipv4

import (
	"fmt"
	"math"
	"sort"

	jmath "github.com/secretpot/jbutil/math"
)

type EchoSummary struct {
	ttl            int
	traceIP        string
	receivedEchoes []float64
	lostCount      int
}

func NewEchoSummary(ttl int, traceIP string, data []float64) EchoSummary {
	r, l := []float64{}, 0
	for _, point := range data {
		if point > 0 {
			r = append(r, jmath.Round(point, 6))
		} else {
			l++
		}
	}
	sort.Float64s(r)
	return EchoSummary{ttl, traceIP, r, l}
}
func (this *EchoSummary) TTL() int {
	return this.ttl
}
func (this *EchoSummary) TraceIP() string {
	return this.traceIP
}
func (this *EchoSummary) Received() int {
	return len(this.receivedEchoes)
}
func (this *EchoSummary) Loss() int {
	return this.lostCount
}
func (this *EchoSummary) Send() int {
	return this.Received() + this.Loss()
}
func (this *EchoSummary) LossRate() float64 {
	rate := float64(this.Loss()) / float64(this.Send())
	return jmath.Round(rate, 2)
}
func (this *EchoSummary) Min() float64 {
	if len(this.receivedEchoes) > 0 {
		return this.receivedEchoes[0]
	} else {
		return 0.0
	}

}
func (this *EchoSummary) Max() float64 {
	if len(this.receivedEchoes) > 0 {
		return this.receivedEchoes[len(this.receivedEchoes)-1]
	} else {
		return 0.0
	}
}
func (this *EchoSummary) Mean() float64 {
	if len(this.receivedEchoes) > 0 {
		sum := 0.0
		for _, pt := range this.receivedEchoes {
			sum += pt
		}
		return jmath.Round(sum/float64(len(this.receivedEchoes)), 6)
	} else {
		return 0.0
	}
}
func (this *EchoSummary) Mdev() float64 {
	if len(this.receivedEchoes) > 0 {
		dsum, mean := 0.0, this.Mean()
		for _, point := range this.receivedEchoes {
			diff := point - mean
			dsum += diff * diff
		}
		return math.Sqrt(dsum) / float64(this.Received())
	} else {
		return 0.0
	}
}
func (this *EchoSummary) ToFlux(measurement string, timestamp string, domain string, who string) string {
	return fmt.Sprintf(
		"%v,hostname=%v,host=%v,domain=%v,who=%v,trace_ip=%v,ttl=%v min=%v,max=%v,avg=%v,mdev=%v,transmit=%v,receive=%v,loss=%v,loss_rate=%v %v",
		measurement,
		LocalhostName(), Localhost(), domain, who, this.TraceIP(), this.TTL(),
		this.Min(), this.Max(), this.Mean(), this.Mdev(), this.Send(), this.Received(), this.Loss(), this.LossRate(),
		timestamp,
	)
}
func (this EchoSummary) String() string {
	return fmt.Sprintf("%d %s send: %d, received: %v, loss: %d, min/max/mean/medev: %.3f/%.3f/%.3f/%.3f ms", this.ttl, this.traceIP, this.Send(), this.Received(), this.lostCount, this.Min(), this.Max(), this.Mean(), this.Mdev())
}
