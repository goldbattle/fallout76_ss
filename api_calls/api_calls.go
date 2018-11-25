package api_calls

import (
	"encoding/json"
	"fallout76_ss/api_db"
	"fmt"
	"github.com/tcnksm/go-httpstat"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Our api ticker
var ticker time.Ticker

// Frequency we poll at
var freq time.Duration

// Sets the polling frequency/seconds
// So if this is 1 => poll every second
func SetPollingFrequency(secs int) {
	freq = time.Duration(secs)
}

// This function will startup a seperate thread for our API calls
// This should poll the api every second and get the server status!
// We also record the timing information of the api so we can compute the statistics (evenutally lol)
func StartAPITicker() {

	// Make the ticker
	ticker := time.NewTicker(time.Second * freq).C
	go func() {
		for {
			select {
			case <-ticker:

				// Open the database
				db := api_db.OpenDatabase()
				defer db.Close()

				// Create a new HTTP request
				req, err := http.NewRequest("GET", "https://api.bethesda.net/status/ext-server-status?product_id=8", nil)
				if err != nil {
					fmt.Errorf("%v", err)
					continue
				}

				// Create a httpstat powered context
				var result httpstat.Result
				ctx := httpstat.WithHTTPStat(req.Context(), &result)
				req = req.WithContext(ctx)

				// Send request with our http client
				// We would normally have keep alive, but this would allow golang to reuse the connection
				// We don't want to reuse the connection, as we want to track how long the API calls take
				var client = &http.Client{
					Timeout: time.Second * freq,
					Transport: &http.Transport{
						DisableKeepAlives: true,
					},
				}
				//client := http.DefaultClient
				res, err := client.Do(req)
				if err != nil {
					fmt.Errorf("%v", err)
					continue
				}

				// Save the response!
				bytes, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Errorf("%v", err)
					continue
				}
				res.Body.Close()

				// Get the timing results
				ms_dns := float64(result.DNSLookup) / float64(time.Millisecond)
				ms_tcp := float64(result.TCPConnection) / float64(time.Millisecond)
				ms_tls := float64(result.TLSHandshake) / float64(time.Millisecond)
				ms_server := float64(result.ServerProcessing) / float64(time.Millisecond)
				ms_content := float64(result.ContentTransfer(time.Now())) / float64(time.Millisecond)
				ms_total := float64(result.Total(time.Now())) / float64(time.Millisecond)

				// Decode the message!
				var dat ExternalServerStatus
				if err := json.Unmarshal(bytes, &dat); err != nil {
					log.Fatal(err)
					continue
				}
				//fmt.Println(dat.Platform.Code)
				//fmt.Println(dat.Platform.Message)
				for k := range dat.Platform.Response {
					fmt.Printf("[api-ticker}: api response %s = %s\n", k, dat.Platform.Response[k])
				}

				// status 0 = unknown, status 1 = up, status 2 = down
				status := 0

				// If we have the fallout76 field, then lets parse it to see if it is up or down...
				if value, ok := dat.Platform.Response["fallout76"]; ok {
					if value == "UP" || value == "up" {
						status = 1
					} else if value == "DOWN" || value == "down" {
						status = 2
					} else {
						status = 0
					}
				}

				// Append to our timing database
				api_db.InsertTiming(db, "fallout67", status, string(bytes), ms_dns, ms_tcp, ms_tls, ms_server, ms_content, ms_total)

			}
		}

	}()

}

// This will stop the ticker from polling the API
// This would happen if we are shutting down the server...
func StopAPITicker() {
	log.Printf("Stopping API ticker from polling....")
	ticker.Stop()
}
