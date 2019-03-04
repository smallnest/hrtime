package hrtime

import (
	"math"
	"time"
)

// MergeBenchmarks merge multiple Benchmark so we can use it in concurrent cases.
// Each goroutine uses its Benchmark and we can merge the results into one Benchmark.
func MergeBenchmarks(benchmarks ...*Benchmark) *Benchmark {
	if len(benchmarks) == 0 {
		return nil
	}

	var start = time.Duration(math.MaxInt64)
	var stop time.Duration
	var laps []time.Duration
	for _, b := range benchmarks {
		b.mustBeCompleted()
		laps = append(laps, b.laps...)
		if b.start < start {
			start = b.start
		}
		if b.stop > stop {
			stop = b.stop
		}
	}

	return &Benchmark{
		step:  len(laps),
		laps:  laps,
		start: start,
		stop:  stop,
	}
}

// Benchmark helps benchmarking using time.
type Benchmark struct {
	step  int
	laps  []time.Duration
	start time.Duration
	stop  time.Duration
}

// NewBenchmark creates a new benchmark using time.
// Count defines the number of samples to measure.
func NewBenchmark(count int) *Benchmark {
	if count <= 0 {
		panic("must have count at least 1")
	}

	return &Benchmark{
		step:  0,
		laps:  make([]time.Duration, count),
		start: 0,
		stop:  0,
	}
}

// mustBeCompleted checks whether measurement has been completed.
func (bench *Benchmark) mustBeCompleted() {
	if bench.stop == 0 {
		panic("benchmarking incomplete")
	}
}

// finalize calculates diffs for each lap.
func (bench *Benchmark) finalize(last time.Duration) {
	if bench.stop != 0 {
		return
	}

	bench.start = bench.laps[0]
	for i := range bench.laps[:len(bench.laps)-1] {
		bench.laps[i] = bench.laps[i+1] - bench.laps[i]
	}
	bench.laps[len(bench.laps)-1] = last - bench.laps[len(bench.laps)-1]
	bench.stop = last
}

// Next starts measuring the next lap.
// It will return false, when all measurements have been made.
func (bench *Benchmark) Next() bool {
	now := Now()
	if bench.step >= len(bench.laps) {
		bench.finalize(now)
		return false
	}
	bench.laps[bench.step] = Now()
	bench.step++
	return true
}

// Laps returns timing for each lap.
func (bench *Benchmark) Laps() []time.Duration {
	bench.mustBeCompleted()
	return append(bench.laps[:0:0], bench.laps...)
}

// Histogram creates an histogram of all the laps.
//
// It creates binCount bins to distribute the data and uses the
// 99.9 percentile as the last bucket range. However, for a nicer output
// it might choose a larger value.
func (bench *Benchmark) Histogram(binCount int) *Histogram {
	bench.mustBeCompleted()

	opts := defaultOptions
	opts.BinCount = binCount

	return NewDurationHistogram(bench.laps, &opts)
}

// HistogramClamp creates an historgram of all the laps clamping minimum and maximum time.
//
// It creates binCount bins to distribute the data and uses the
// maximum as the last bucket.
func (bench *Benchmark) HistogramClamp(binCount int, min, max time.Duration) *Histogram {
	bench.mustBeCompleted()

	laps := make([]time.Duration, 0, len(bench.laps))
	for _, lap := range bench.laps {
		if lap < min {
			laps = append(laps, min)
		} else {
			laps = append(laps, lap)
		}
	}

	opts := defaultOptions
	opts.BinCount = binCount
	opts.ClampMaximum = float64(max.Nanoseconds())
	opts.ClampPercentile = 0

	return NewDurationHistogram(laps, &opts)
}
