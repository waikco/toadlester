package app

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

// RunTest executes a given test using the vegeta library and returns the
// related metrics once complete.
func (a *App) RunTest(name string, duration time.Duration, tps int, target string) (*vegeta.Metrics, error) {
	// set up test
	b, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	}
	src := bytes.NewBuffer(b)
	targeter := vegeta.NewHTTPTargeter(src, nil, nil)

	rate := vegeta.Rate{Freq: tps, Per: time.Second}
	attacker := vegeta.NewAttacker()
	defer attacker.Stop()

	// run test
	var metrics vegeta.Metrics

	for res := range attacker.Attack(targeter, rate, duration, name) {
		metrics.Add(res)
	}
	metrics.Close()

	r := vegeta.NewTextReporter(&metrics)
	a.Logger.Info().Msgf("%v", r.Report(os.Stdout))
	return &metrics, nil
}
