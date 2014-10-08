package monator

import (
    "fmt"
    "time"
    "strings"
)

type CheckResult struct {
    Status   CheckStatus
    Duration CheckDuration
    Time     time.Time
    Check    *Check
    Error    string
    Average  *LoadAverage
}

func NewErrorCheckResult(status CheckStatus, check *Check, diffTime CheckDuration, errorstr string) *CheckResult {
    return &CheckResult{
        Status:   status,
        Duration: diffTime,
        Time:     time.Now(),
        Check:    check,
        Error:    errorstr,
    }
}

func MakeErrorCheckResult(check *Check, diffTime CheckDuration, errorstr string) *CheckResult {
    return NewErrorCheckResult(CheckStatusError, check, diffTime, errorstr)
}

func MakeFailureCheckResult(check *Check, diffTime CheckDuration, errorstr string) *CheckResult {
    return NewErrorCheckResult(CheckStatusFailure, check, diffTime, errorstr)
}

func (c *CheckResult) String() string {
    // NOTE: we call duration.String() as Stringer is for pointers, we don't
    //       have a pointer here (Duration is an int64).

    if c.Error != "" {
        return fmt.Sprintf("%s %s %s %s %s \"%v\"", c.Time.Format(time.RFC3339), c.Status,
            c.Check.Name, c.Duration.String(), c.Average,
            strings.Replace(c.Error, "\"", "\\\"", -1))
    }

    return fmt.Sprintf("%s %s %s %s %s", c.Time.Format(time.RFC3339), c.Status,
        c.Check.Name, c.Duration.String(), c.Average)
}

func (result *CheckResult) AfterCheckFixes(check *Check) {
    switch result.Status {
    case CheckStatusSuccess:
        // If all went well, check if we exceeded the warning time.
        if check.Warning.Milliseconds() > 0 &&
            result.Duration.Milliseconds() >= check.Warning.Milliseconds() {
            result.Status = CheckStatusWarning
        }
    case CheckStatusUnknown:
        // At this point, if we don't have a status, we have a failure.
        result.Status = CheckStatusFailure
    }
}

