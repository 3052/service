package main

import (
   "bytes"
   "encoding/json"
   "fmt"
   "io"
   "log"
   "net/http"
   "net/url"
)

func main() {
   http.DefaultTransport = &http.Transport{
      Proxy: func(req *http.Request) (*url.URL, error) {
         log.Println(req.Method, req.URL)
         return nil, nil
      },
   }
   // Make an HTTP GET request to the URL
   res, err := http.Get("https://www.justwatch.com/us")
   if err != nil {
      log.Fatalf("failed to get URL: %v", err)
   }
   defer res.Body.Close()
   if res.StatusCode != 200 {
      log.Fatalf("request failed with status code: %d %s", res.StatusCode, res.Status)
   }
   // Read the entire response body into a byte slice.
   bodyBytes, err := io.ReadAll(res.Body)
   if err != nil {
      log.Fatalf("failed to read response body: %v", err)
   }
   // --- Step 1: Extract the JSON using bytes.Cut ---
   // Find the start of the data.
   _, after, found := bytes.Cut(bodyBytes, []byte("window.__DATA__="))
   if !found {
      log.Fatal("could not find 'window.__DATA__=' in the response body")
   }
   // Find the end of the data.
   jsonData, _, found := bytes.Cut(after, []byte("</script>"))
   if !found {
      log.Fatal("could not find closing '</script>' tag after the data")
   }
   // --- Step 2: Unmarshal the JSON ---
   // No trimming is performed.
   // Define the struct to match the nested JSON structure.
   var result struct {
      State struct {
         Constant struct {
            Providers []struct {
               HasTitles bool
               Slug      string 
            }
         }
      }
   }
   // Unmarshal the raw JSON byte slice into the struct.
   if err := json.Unmarshal(jsonData, &result); err != nil {
      // This is where the program will fail.
      log.Fatalf("failed to unmarshal JSON: %v", err)
   }
   // --- Step 3: Use the structured data (this part will not be reached) ---
   var i int
   for _, provider := range result.State.Constant.Providers {
      if provider.HasTitles {
         i++
         fmt.Println(i, provider.Slug)
      }
   }
}
