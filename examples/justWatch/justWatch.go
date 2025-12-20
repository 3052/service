package main

import (
   "41.neocities.org/service/justWatch"
   "bytes"
   "errors"
   "flag"
   "fmt"
   "log"
   "net/http"
   "net/url"
   "os"
   "path"
   "strings"
   "time"
)

func (c *command) do_address() error {
   url_path, err := justWatch.GetPath(c.address)
   if err != nil {
      return err
   }
   var content justWatch.Content
   err = content.Fetch(url_path)
   if err != nil {
      return err
   }
   var allEnrichedOffers []justWatch.EnrichedOffer
   for _, tag := range content.HrefLangTags {
      locale, ok := justWatch.EnUs.Locale(&tag)
      if !ok {
         return errors.New("Locale")
      }
      log.Print(locale)
      offers, err := tag.Offers(locale)
      if err != nil {
         return err
      }
      for _, offer := range offers {
         allEnrichedOffers = append(allEnrichedOffers,
            justWatch.EnrichedOffer{Offer: offer, Locale: locale},
         )
      }
      time.Sleep(c.sleep)
   }
   enrichedOffers := justWatch.Deduplicate(allEnrichedOffers)
   enrichedOffers = justWatch.FilterOffers(
      enrichedOffers, strings.Split(c.filters, ",")...,
   )
   sortedUrls, groupedOffers := justWatch.GroupAndSortByURL(enrichedOffers)
   data := &bytes.Buffer{}
   for index, address := range sortedUrls {
      if index >= 1 {
         data.WriteByte('\n')
      }
      data.WriteString("##")
      fmt.Fprint(data, address)
      for _, enriched := range groupedOffers[address] {
         data.WriteByte('\n')
         data.WriteString("\ncountry = ")
         data.WriteString(enriched.Locale.Country)
         data.WriteString("\nname = ")
         data.WriteString(enriched.Locale.CountryName)
         data.WriteString("\nmonetization = ")
         data.WriteString(enriched.Offer.MonetizationType)
         if enriched.Offer.ElementCount >= 1 {
            data.WriteString("\ncount = ")
            fmt.Fprint(data, enriched.Offer.ElementCount)
         }
      }
   }
   name := path.Base(url_path) + ".md"
   log.Println("WriteFile", name)
   return os.WriteFile(name, data.Bytes(), os.ModePerm)
}

func main() {
   log.SetFlags(log.Ltime)
   http.DefaultTransport = &http.Transport{
      Protocols: &http.Protocols{}, // github.com/golang/go/issues/25793
      Proxy: func(req *http.Request) (*url.URL, error) {
         if req.URL.Path != "/graphql" {
            log.Println(req.Method, req.URL)
         }
         return http.ProxyFromEnvironment(req)
      },
   }
   err := new(command).run()
   if err != nil {
      log.Fatal(err)
   }
}

func (c *command) run() error {
   flag.StringVar(&c.address, "a", "", "address")
   flag.DurationVar(&c.sleep, "s", 99*time.Millisecond, "sleep")
   flag.StringVar(&c.filters, "f", "BUY,CINEMA,FAST,RENT", "filters")
   flag.Parse()

   if c.address != "" {
      return c.do_address()
   }
   flag.Usage()
   return nil
}

type command struct {
   address string
   filters string
   sleep   time.Duration
}
