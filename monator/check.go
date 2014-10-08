package monator

import (
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "strings"
    "time"
)

const (
    CheckTypeHttp = "http"
)

type CheckResultChan chan *CheckResult
type DialTimeout func(network, addr string) (net.Conn, error)

type Check struct {
    Name       string        `json:"name"`
    Url        string        `json:"url"`
    Headers    []string      `json:"headers"`
    Method     string        `json:"method"`
    Contains   string        `json:"contains"`
    Timeout    CheckDuration `json:"timeout"`
    Warning    CheckDuration `json:"warning"`
    Frequency  CheckDuration `json:"frequency"`
    Retry      CheckDuration `json:"retry"`
    Type       string        `json:"type"`
    Average    *LoadAverage
    LastStatus CheckStatus
    client     *http.Client
    request    *http.Request
}

// TODO: Should be called: SetDefaults.
func (c *Check) Init() {
    if c.Average == nil {
        c.Average = new(LoadAverage)
    }

    if int64(c.Retry) == 0 {
        defaultValue, _ := time.ParseDuration("10s")
        c.Retry = CheckDuration(defaultValue)
    }

    if int64(c.Frequency) == 0 {
        defaultValue, _ := time.ParseDuration("1m")
        c.Frequency = CheckDuration(defaultValue)
    }

    // Timeout and Warning can be zero (they are not used.)
}

func (c *Check) checkContains(body string) bool {
    if c.Contains != "" {
        return strings.Contains(body, c.Contains)
    }

    return true
}

func (c *Check) HttpGetResult() *CheckResult {
    req, err := c.MakeHttpReqTimeout()
    if err != nil {
        return MakeErrorCheckResult(c, CheckDuration(0), err.Error())
    }

    beforeTime := time.Now()
    resp, err := c.GetClient().Do(req)
    if err != nil {
        return MakeErrorCheckResult(c, CheckDuration(0), err.Error())
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    diffTime := CheckDuration(time.Since(beforeTime))

    if err != nil {
        return MakeErrorCheckResult(c, diffTime, err.Error())
    }

    result := &CheckResult{
        Status:   CheckStatusUnknown,
        Duration: diffTime,
        Time:     beforeTime,
        Check:    c,
    }

    if resp.StatusCode == 200 {
        if c.checkContains(string(body)) {
            result.Status = CheckStatusSuccess
        } else {
            result.Status = CheckStatusError
            result.Error = "Contains check failed"
        }
    } else {
        result.Status = CheckStatusError
        result.Error = fmt.Sprintf("HTTP Status is '%s'", resp.Status)
    }

    return result
}

func (c *Check) GetClient() *http.Client {
    if c.client == nil {
        if c.Timeout.Milliseconds() > 0 {
            transport := &http.Transport{
                Dial: func(network, addr string) (net.Conn, error) {
                    return net.DialTimeout(network, addr, time.Duration(c.Timeout))
                },
            }

            c.client = &http.Client{
                Transport: transport,
            }
        } else {
            c.client = &http.Client{}
        }
    }

    return c.client
}

func (c *Check) MakeHttpReqTimeout() (*http.Request, error) {
    if c.request == nil {
        if c.Method == "" {
            c.Method = "GET"
        }

        req, err := http.NewRequest(c.Method, c.Url, nil)

        if len(c.Headers) > 0 {
            for i := range c.Headers {
                if strs := strings.SplitN(c.Headers[i], ": ", 2); err == nil {
                    // Special trick to set the Host header.
                    if strs[0] == "Host" {
                        req.Host = strs[1]
                    } else {
                        req.Header.Add(strs[0], strs[1])
                    }
                }
            }
        }

        if err == nil {
            c.request = req
        }

        return req, err
    }

    return c.request, nil
}

func (c *Check) SetHeaders(headers map[string]string) {
    for k, v := range headers {
        c.Headers = append(c.Headers, fmt.Sprintf("%s: %s", k, v))
    }
}

// TODO: Decouple this, make it an interface.
func (c *Check) getResult() *CheckResult {
    switch c.Type {
    case CheckTypeHttp:
        fallthrough
    default:
        return c.HttpGetResult()
    }

    // Not reached.
    panic("No type or incorrect type was specified")
    return nil
}

func (c *Check) PerformCheck(ch CheckResultChan) {
    for {
        result := c.getResult()

        ch <- result

        // If the check didn't succeed, try again immediately.
        // If it was a hard error don't retry immediately.
        if result.Status.In(CheckStatusSuccess, CheckStatusFailure) {
            time.Sleep(time.Duration(c.Frequency))
        } else {
            time.Sleep(time.Duration(c.Retry))
        }
    }
}

