package main

import (
   "bytes"
   "encoding/json"
   "flag"
   "fmt"
   "io"
   "log"
   "net/http"
   "net/url"
   "os"
   "sort" // Imported the sort package
   "strings"
)

func main() {
   log.SetFlags(log.Ltime)
   http.DefaultTransport = &http.Transport{
      Protocols: &http.Protocols{}, // github.com/golang/go/issues/25793
      Proxy: func(req *http.Request) (*url.URL, error) {
         log.Println(req.Method, req.URL)
         return nil, nil
      },
   }
   // Define flags
   countryCode := flag.String("a", "", "Country code to process (e.g., 'us')")
   jsonFile := flag.String("b", "", "JSON file with a list of provider URLs")

   flag.Parse()

   // If no flags are provided, print help
   if *countryCode == "" && *jsonFile == "" {
      flag.Usage()
      return
   }

   if *countryCode != "" {
      processCountry(*countryCode, nil)
   }

   if *jsonFile != "" {
      file, err := os.ReadFile(*jsonFile)
      if err != nil {
         log.Fatalf("failed to read json file: %v", err)
      }

      var providerURLs []string
      if err := json.Unmarshal(file, &providerURLs); err != nil {
         log.Fatalf("failed to unmarshal json file: %v", err)
      }

      countriesToProviders := make(map[string]map[string]bool)

      for _, providerURL := range providerURLs {
         parsedURL, err := url.Parse(providerURL)
         if err != nil {
            log.Printf("failed to parse URL %s: %v", providerURL, err)
            continue
         }

         pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
         if len(pathParts) != 3 {
            log.Printf("invalid provider URL format: %s. Expected 3 path components like /<country>/<...>/<slug>", providerURL)
            continue
         }

         country := pathParts[0]
         providerSlug := pathParts[2]

         if _, ok := countriesToProviders[country]; !ok {
            countriesToProviders[country] = make(map[string]bool)
         }
         countriesToProviders[country][providerSlug] = true
      }

      // --- Start of Ordering Logic ---
      // Get the keys (country codes) from the map
      var countries []string
      for country := range countriesToProviders {
         countries = append(countries, country)
      }

      // Sort the country codes alphabetically
      sort.Strings(countries)
      // --- End of Ordering Logic ---

      // Iterate over the sorted slice of countries
      for _, country := range countries {
         providerFilter := countriesToProviders[country]
         processCountry(country, providerFilter)
      }
   }
}

func processCountry(countryCode string, providerFilter map[string]bool) {
   res, err := http.Get(fmt.Sprintf("https://www.justwatch.com/%s", countryCode))
   if err != nil {
      log.Printf("failed to get URL for country %s: %v", countryCode, err)
      return
   }
   defer res.Body.Close()

   if res.StatusCode != 200 {
      log.Printf("request failed for country %s with status code: %d %s", countryCode, res.StatusCode, res.Status)
      return
   }

   bodyBytes, err := io.ReadAll(res.Body)
   if err != nil {
      log.Printf("failed to read response body for country %s: %v", countryCode, err)
      return
   }

   _, after, found := bytes.Cut(bodyBytes, []byte("window.__DATA__="))
   if !found {
      log.Printf("could not find 'window.__DATA__=' in the response body for country %s", countryCode)
      return
   }

   jsonData, _, found := bytes.Cut(after, []byte("</script>"))
   if !found {
      log.Printf("could not find closing '</script>' tag after the data for country %s", countryCode)
      return
   }

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
   if err := json.Unmarshal(jsonData, &result); err != nil {
      log.Printf("failed to unmarshal JSON for country %s: %v", countryCode, err)
      return
   }
   var i int
   for _, provider := range result.State.Constant.Providers {
      if provider.HasTitles {
         if providerFilter == nil || providerFilter[provider.Slug] {
            i++
            fmt.Println(i, provider.Slug)
         }
      }
   }
}
