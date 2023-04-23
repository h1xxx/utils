package main

import (
	"fmt"
	"math"
	"time"
)

func doBench(vars *varsT) {
	var results []float64

	for i := 0; i < 30; i++ {
		var runCount float64
		start := time.Now()

		for time.Since(start).Seconds() < 1 {
			var st sttsT
			getAllInfo(&st, vars)
			runCount++
		}

		results = append(results, runCount)
		fmt.Printf("number of runs per 1s: %8.0f\n", runCount)
	}

	mean, stdev := getMeanAndStdev(results)
	fmt.Printf("\nbenchmark result: mean %.0f (stdev %.2f%%)\n",
		mean, stdev)
}

func getMeanAndStdev(s []float64) (float64, float64) {
	var sum, sumDiffSq float64
	size := float64(len(s))

	for _, val := range s {
		sum += val
	}

	mean := sum / size

	for i := 0; i < len(s); i++ {
		sumDiffSq += math.Pow(s[i]-mean, 2)
	}

	variance := sumDiffSq / size
	stdev := math.Sqrt(variance)

	return mean, stdev / mean * 100
}
