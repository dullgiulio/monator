package monator

const (
    CheckStatusUnknown = iota
    CheckStatusSuccess
    CheckStatusWarning
    CheckStatusError
    CheckStatusFailure
    CheckStatusAverage
)

type CheckStatus int64

func (s CheckStatus) String() string {
    switch s {
    case CheckStatusUnknown:
        return "UNK"
    case CheckStatusSuccess:
        return "OK"
    case CheckStatusWarning:
        return "WARN"
    case CheckStatusError:
        return "ERR"
    case CheckStatusFailure:
        return "FAIL"
    case CheckStatusAverage:
        return "AVG"
    default:
        return "UNK"
    }
}

func (s CheckStatus) In(stata ...CheckStatus) bool {
    for _, status := range stata {
        if s == status {
            return true
        }
    }

    return false
}

