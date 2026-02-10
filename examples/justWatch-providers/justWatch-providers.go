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
   "sort"
   "strings"
)

// ProviderInfo holds the data for a single provider.
type ProviderInfo struct {
   Slug string
   Rank int
}

// processCountry fetches and parses provider data for a country.
// It returns a slice of ProviderInfo structs, ordered by their rank on JustWatch.
func processCountry(countryCode string, providerFilter map[string]bool) []ProviderInfo {
   res, err := http.Get(fmt.Sprintf("https://www.justwatch.com/%s", countryCode))
   if err != nil {
      log.Printf("failed to get URL for country %s: %v", countryCode, err)
      return nil
   }
   defer res.Body.Close()

   if res.StatusCode != 200 {
      log.Printf("request failed for country %s with status code: %d %s", countryCode, res.StatusCode, res.Status)
      return nil
   }

   bodyBytes, err := io.ReadAll(res.Body)
   if err != nil {
      log.Printf("failed to read response body for country %s: %v", countryCode, err)
      return nil
   }

   _, after, found := bytes.Cut(bodyBytes, []byte("window.__DATA__="))
   if !found {
      log.Printf("could not find 'window.__DATA__=' in the response body for country %s", countryCode)
      return nil
   }

   jsonData, _, found := bytes.Cut(after, []byte("</script>"))
   if !found {
      log.Printf("could not find closing '</script>' tag after the data for country %s", countryCode)
      return nil
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
      return nil
   }

   var foundProviders []ProviderInfo
   rankCounter := 0
   for _, provider := range result.State.Constant.Providers {
      if provider.HasTitles {
         rankCounter++
         if providerFilter == nil || providerFilter[provider.Slug] {
            foundProviders = append(foundProviders, ProviderInfo{Slug: provider.Slug, Rank: rankCounter})
         }
      }
   }
   return foundProviders
}

func main() {
   log.SetFlags(log.Ltime)
   http.DefaultTransport = &http.Transport{
      Protocols: &http.Protocols{}, // github.com/golang/go/issues/25793
      Proxy: func(req *http.Request) (*url.URL, error) {
         log.Println(req.Method, req.URL)
         return nil, nil
      },
   }

   countryCode := flag.String("a", "", "Country code to process (e.g., 'us')")
   jsonFile := flag.String("b", "", "JSON file with a list of provider URLs")
   flag.Parse()

   if *countryCode == "" && *jsonFile == "" {
      flag.Usage()
      return
   }

   // Handle -a flag
   if *countryCode != "" {
      providers := processCountry(*countryCode, nil)
      for _, provider := range providers {
         fmt.Printf("%d. (%s) %s\n", provider.Rank, *countryCode, provider.Slug)
      }
   }

   // Handle -b flag
   if *jsonFile != "" {
      file, err := os.ReadFile(*jsonFile)
      if err != nil {
         log.Fatalf("failed to read json file: %v", err)
      }

      var providerURLs []string
      if err := json.Unmarshal(file, &providerURLs); err != nil {
         log.Fatalf("failed to unmarshal json file: %v", err)
      }

      countriesToProvidersList := make(map[string][]string)
      for _, providerURL := range providerURLs {
         parsedURL, err := url.Parse(providerURL)
         if err != nil {
            log.Printf("failed to parse URL %s: %v", providerURL, err)
            continue
         }
         pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
         if len(pathParts) != 3 {
            log.Printf("invalid provider URL format: %s", providerURL)
            continue
         }
         country, providerSlug := pathParts[0], pathParts[2]
         countriesToProvidersList[country] = append(countriesToProvidersList[country], providerSlug)
      }

      type FinalResult struct {
         Country string
         Slug    string
         Rank    int
      }
      var finalResults []FinalResult

      for country, providers := range countriesToProvidersList {
         providerFilter := make(map[string]bool)
         for _, slug := range providers {
            providerFilter[slug] = true
         }

         foundProviders := processCountry(country, providerFilter)
         for _, pInfo := range foundProviders {
            finalResults = append(finalResults, FinalResult{Country: country, Slug: pInfo.Slug, Rank: pInfo.Rank})
         }
      }

      // Sort the final combined list by the rank from JustWatch
      sort.Slice(finalResults, func(i, j int) bool {
         return finalResults[i].Rank < finalResults[j].Rank
      })

      // Print the final, sorted, and consolidated list
      for i, result := range finalResults {
         fmt.Printf("%d. (%s) %s\n", i+1, result.Country, result.Slug)
      }
   }
}
