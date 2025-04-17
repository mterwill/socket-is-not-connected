package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Register handlers
	http.HandleFunc("/", serveHomePage)
	http.HandleFunc("/api/data", serveData)

	// Start the server
	port := 9092
	fmt.Printf("Server is running at http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// serveHomePage serves the static HTML page with JavaScript
func serveHomePage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Resource Loop Test</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        #log {
            border: 1px solid #ccc;
            padding: 10px;
            height: 300px;
            overflow-y: auto;
            margin-bottom: 20px;
            font-family: monospace;
        }
        .success { color: green; }
        .error { color: red; }
        .controls {
            display: grid;
            grid-template-columns: 150px 100px;
            grid-gap: 10px;
            margin-bottom: 20px;
        }
        .controls label {
            grid-column: 1;
        }
        .controls input {
            grid-column: 2;
        }
        .buttons {
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <h1>Resource Request Loop</h1>
    <div class="controls">
        <label>Request interval (ms):</label>
        <input type="number" id="interval" value="10" min="0">

        <label>Parallel requests:</label>
        <input type="number" id="parallel" value="50" min="1" max="100">
    </div>
    <div class="buttons">
        <button onclick="startRequests()">Start</button>
        <button onclick="stopRequests()">Stop</button>
        <button onclick="clearLog()">Clear Log</button>
    </div>
    <div>
        <h3>Request Log</h3>
        <div id="log"></div>
    </div>

    <script>
        let intervalId = null;
        let requestCounter = 0;
        let activeRequests = 0;
        const maxActiveRequests = 1000; // Safety limit

        function logMessage(message, isError = false) {
            const log = document.getElementById('log');
            const entry = document.createElement('div');
            entry.className = isError ? 'error' : 'success';
            entry.textContent = message;
            log.appendChild(entry);
            log.scrollTop = log.scrollHeight;
        }

        function clearLog() {
            document.getElementById('log').innerHTML = '';
        }

        function makeRequest() {
            const parallelCount = parseInt(document.getElementById('parallel').value);

            // Don't start too many parallel requests
            if (activeRequests >= maxActiveRequests) {
                logMessage('Warning: Too many active requests (' + activeRequests + '). Waiting for some to complete...', true);
                return;
            }

            // Start multiple parallel requests
            for (let i = 0; i < parallelCount; i++) {
                const requestId = ++requestCounter;
                const startTime = new Date();
                activeRequests++;

                logMessage('[' + startTime.toLocaleTimeString() + '] Request #' + requestId + ' started');

                fetch('/api/data?id=' + requestId)
                    .then(response => {
                        if (!response.ok) {
                            throw new Error('Server returned ' + response.status);
                        }
                        return response.json();
                    })
                    .then(data => {
                        const endTime = new Date();
                        const duration = endTime - startTime;
                        logMessage('[' + endTime.toLocaleTimeString() + '] Request #' + requestId + ' completed in ' + duration + 'ms: ' + JSON.stringify(data));
                    })
                    .catch(error => {
                        logMessage('Request #' + requestId + ' failed: ' + error, true);
                    })
                    .finally(() => {
                        activeRequests--;
                    });
            }
        }

        function startRequests() {
            if (intervalId) {
                stopRequests();
            }

            const interval = parseInt(document.getElementById('interval').value);
            const parallel = parseInt(document.getElementById('parallel').value);

            if (isNaN(interval) || interval < 0) {
                alert('Please enter a valid interval (minimum 0ms)');
                return;
            }

            if (isNaN(parallel) || parallel < 1) {
                alert('Please enter a valid number of parallel requests (minimum 1)');
                return;
            }

            logMessage('Starting ' + parallel + ' parallel requests every ' + interval + 'ms');
            makeRequest(); // Make first batch of requests immediately

            // Only set interval if > 0, otherwise it's a one-time batch
            if (interval > 0) {
                intervalId = setInterval(makeRequest, interval);
            }
        }

        function stopRequests() {
            if (intervalId) {
                clearInterval(intervalId);
                intervalId = null;
                logMessage('Request loop stopped. ' + activeRequests + ' requests still in progress.');
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// serveData simulates an API endpoint that returns JSON data
func serveData(w http.ResponseWriter, r *http.Request) {
	// Optional: Add artificial delay to simulate processing time
	// time.Sleep(200 * time.Millisecond)

	id := r.URL.Query().Get("id")

	// Create response data
	data := map[string]interface{}{
		"id":        id,
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
		"data":      fmt.Sprintf("Response data for request %s", id),
	}

	// Convert to JSON and send
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
