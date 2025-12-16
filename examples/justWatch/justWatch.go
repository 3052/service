package main

import (
   "41.neocities.org/service/justWatch"
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
   var data []byte
   for index, address := range sortedUrls {
      if index >= 1 {
         data = append(data, '\n')
      }
      data = fmt.Appendln(data, "##", address)
      for _, enrichedOffer := range groupedOffers[address] {
         data = fmt.Appendln(data, "\ncountry =", enrichedOffer.Locale.Country)
         data = fmt.Appendln(data, "name =", enrichedOffer.Locale.CountryName)
         data = fmt.Appendln(data, "monetization =", enrichedOffer.Offer.MonetizationType)
         if enrichedOffer.Offer.ElementCount >= 1 {
            data = fmt.Appendln(data, "count =", enrichedOffer.Offer.ElementCount)
         }
      }
   }
   name := path.Base(url_path) + ".md"
   log.Println("WriteFile", name)
   return os.WriteFile(name, data, os.ModePerm)
}
