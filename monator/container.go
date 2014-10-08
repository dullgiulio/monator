package monator

import (
    "fmt"
    "time"
    "io/ioutil"
    "encoding/json"
)


type CheckContainer struct {
    checks []*Check
}

func NewCheckContainer() *CheckContainer {
    return &CheckContainer{
        checks: make([]*Check, 0),
    }
}

func (c *CheckContainer) All() []*Check {
    checks := make([]*Check, len(c.checks))

    for i := 0; i < len(c.checks); i++ {
        checks[i] = c.checks[i]
    }

    return checks
}

func (c *CheckContainer) LoadFromJson(filename string) error {
    if contents, err := ioutil.ReadFile(filename); err != nil {
        return err
    } else {
        var checks []Check

        if err = json.Unmarshal(contents, &checks); err != nil {
            return err
        }

        return c.Add(&checks)
    }

    return nil
}

func (c *CheckContainer) SetHeaders(headers map[string]string) {
    for i := range c.checks {
        c.checks[i].SetHeaders(headers)
    }
}

func (c *CheckContainer) Add(checks *[]Check) error {
    // TODO: Error and fail if Name is not unique.
    for _, check := range *checks {
        check.Init()

        // Just a trick to be sure to have the right pointers.
        func(c *CheckContainer, check Check) {
            c.checks = append(c.checks, &check)
        }(c, check)
    }

    return nil
}

func (checks *CheckContainer) StartChecks(ch CheckResultChan) {
    for _, check := range checks.All() {
        go check.PerformCheck(ch)
    }
}

func (c *CheckContainer) ReadChannel(ch CheckResultChan, verbose bool) {
    checkFrequency := 100 * time.Millisecond

    for {
        select {
        case result := <-ch:
            // Channel has been closed.
            if result == nil {
                return
            }

            check := result.Check

            result.AfterCheckFixes(check)

            if result.Status != CheckStatusAverage {
                // Calculate average for this check
                check.Average.Add(result.Duration)
                // Augment this last check with the per-Check averages.
                result.Average = check.Average
            }

            // Show status update only if there is a status change
            // or the status is a persisting error.
            if verbose ||
                result.Status != check.LastStatus ||
                !result.Status.In(CheckStatusSuccess, CheckStatusFailure) {
                if result.Status == CheckStatusAverage {
                    if result.Average.IsReady() || verbose {
                        fmt.Println(result)
                    }
                } else {
                    // TODO: Output should be configurable.
                    fmt.Println(result)
                }
            }

            check.LastStatus = result.Status
        case <-time.After(checkFrequency):
            continue
        }
    }
}

func (c *CheckContainer) EmitAverages(ch CheckResultChan, frequency time.Duration) {
    if frequency.Nanoseconds() == 0 {
        return
    }

    go func(c *CheckContainer, ch CheckResultChan) {
        for {
            time.Sleep(frequency)

            for _, c := range c.checks {
                result := NewErrorCheckResult(CheckStatusAverage, c,
                    CheckDuration(0), "Load averages")
                result.Average = c.Average

                ch <- result
            }
        }
    }(c, ch)
}

