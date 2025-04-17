# 'socket is not connected' errors

## Context
We maintain a proxy for local development of frontend services that handles TLS as well as some internal auth concerns.
Users frequently see errors like the following:

```
WARN	proxying request: write tcp [::1]:63845->[::1]:9091: write: socket is not connected
WARN	proxying request: read tcp [::1]:63844->[::1]:9091: read: socket is not connected
```

In particular, we are seeing this issue proxying a frontend that makes a LOT of
requests to load unbundled static assets in quick succession from the browser.

I spent some time debugging this and realized that switching to IPv4 fixed the issue.
My next step was to see if I could reproduce in an isolated environment outside
of our internal applications. I had Copilot generate me a test harness to
produce a lot of requests (`./frontend`) and wrote a basic proxy similar to
what we have internally. As a first pass, I didn't write HTTPS, though I
realized I actually couldn't reproduce the issue with plain HTTP.

I did find this open issue: https://github.com/golang/go/issues/68237

## Setup
In one window, start the frontend server:

```
go run ./frontend
```

In another, we'll start our proxy:

```
go run ./proxy -mode=MODE -usptream=UPSTREAM
```

Visit http://localhost:10000/ (or https://locahost:10000) to start the test.

## Results

System config:
- macOS 15.4 (24E248)
- WARP Version: 2025.1.861.0 (20250219.15)
- Chrome Version 135.0.7049.86 (Official Build) (arm64)

Request interval: 10ms
Parallel requests: 50

| Mode | Upstream | WARP | Result |
| ---- | -------- | ---- | ------ |
| `https` | localhost:9092 | On | `socket is not connected` errors |
| `https` | localhost:9092 | Off | No errors |
| `https` | 127.0.0.1:9092 | On | No errors |
| `http` | locahost:9092 | On | No errors |

I can also reproduce outside the browser using [vegeta](https://github.com/tsenart/vegeta):

```
$ echo "GET https://localhost:10000/api/data?id=2725" | vegeta attack -insecure -duration=10s -workers=50 -rate=500 | tee results.bin | vegeta report
Requests      [total, rate, throughput]         5000, 500.10, 403.76
Duration      [total, attack, wait]             9.999s, 9.998s, 524.417µs
Latencies     [min, mean, 50, 90, 95, 99, max]  202.375µs, 792.411µs, 465.142µs, 878.789µs, 1.233ms, 3.032ms, 1.003s
Bytes In      [total, mean]                     322960, 64.59
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           80.74%
Status Codes  [code:count]                      200:4037  502:963
Error Set:
502 Bad Gateway
```

The errors are consistently reproducible, although the results vary in magnitude between tests. You may have to run it a few times.

## Open questions / ideas
- Why isn't this reproducible with plaintext HTTP? Does it have something to do with HTTP/2 or TLS?
