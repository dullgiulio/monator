package monator

import (
    "fmt"
)

type LoadAverage struct {
    average       CheckDuration
    totalLoadTime CheckDuration
    timesChecked  int64
}

// We have a real load average after at least two checks.
func (l *LoadAverage) IsReady() bool {
    return l.timesChecked > 1
}

func (avg *LoadAverage) String() string {
    return fmt.Sprintf("avg %v chk %d", avg.average, avg.timesChecked)
}

func (avg *LoadAverage) Add(t CheckDuration) {
    avg.totalLoadTime += t
    avg.timesChecked++

    // Calculate cached average
    avg.average = CheckDuration(avg.totalLoadTime.Nanoseconds() / avg.timesChecked)
}


