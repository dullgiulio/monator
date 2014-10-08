package monator

import (
    "time"
    "bytes"
    "encoding/json"
)

type CheckDuration time.Duration

func (d CheckDuration) Hours() float64 {
    return time.Duration(d).Hours()
}

func (d CheckDuration) Minutes() float64 {
    return time.Duration(d).Minutes()
}

func (d CheckDuration) Nanoseconds() int64 {
    return time.Duration(d).Nanoseconds()
}

func (d CheckDuration) Seconds() float64 {
    return time.Duration(d).Seconds()
}

func (d CheckDuration) Milliseconds() int64 {
    return time.Duration(d).Nanoseconds() / int64(time.Millisecond)
}

func (d CheckDuration) String() string {
    // Round up nanoseconds, they are stupid.
    rounded := ((time.Duration(d)).Nanoseconds() / int64(time.Millisecond)) * int64(time.Millisecond)
    return time.Duration(rounded).String()
}

// Unmarshaller for JSON decoding.
func (d *CheckDuration) UnmarshalJSON(data []byte) error {
    b := bytes.NewBuffer(data)
    dec := json.NewDecoder(b)

    var s string

    if err := dec.Decode(&s); err != nil {
        return err
    }
    if duration, err := time.ParseDuration(s); err != nil {
        return err
    } else {
        *d = CheckDuration(duration)
    }

    return nil
}

