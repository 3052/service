package main

import (
   "41.neocities.org/service/justWatch"
   "errors"
   "flag"
   "fmt"
   "log"
   "net/http"
   "os"
   "path"
   "strings"
   "time"
)

func main() {
   log.SetFlags(log.Ltime)
   http.DefaultClient.Transport = &justWatch.Transport
   var set flag_set
   flag.StringVar(&set.address, "a", "", "address")
   flag.DurationVar(&set.sleep, "s", 99*time.Millisecond, "sleep")
   flag.StringVar(&set.filters, "f", "BUY,CINEMA,FAST,RENT", "filters")
   flag.Parse()
   if set.address != "" {
      err := set.do_address()
      if err != nil {
         log.Fatal(err)
      }
   } else {
      flag.Usage()
   }
}

func (f *flag_set) do_address() error {
   url_path, err := justWatch.GetPath(f.address)
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
      time.Sleep(f.sleep)
   }
   enrichedOffers := justWatch.Deduplicate(allEnrichedOffers)
   enrichedOffers = justWatch.FilterOffers(
      enrichedOffers, strings.Split(f.filters, ",")...,
   )
   sortedUrls, groupedOffers := justWatch.GroupAndSortByURL(enrichedOffers)
   var data []byte
   for index, url := range sortedUrls {
      if index >= 1 {
         data = append(data, '\n')
      }
      data = fmt.Appendln(data, url)
      for offerIndex, enrichedOffer := range groupedOffers[url] {
         if offerIndex >= 1 {
            data = append(data, '\n')
         }
         data = fmt.Appendln(data, "country =", enrichedOffer.Locale.Country)
         data = fmt.Appendln(data, "name =", enrichedOffer.Locale.CountryName)
         data = fmt.Appendln(data, "monetization =", enrichedOffer.Offer.MonetizationType)
         if enrichedOffer.Offer.ElementCount >= 1 {
            data = fmt.Appendln(data, "count =", enrichedOffer.Offer.ElementCount)
         }
      }
   }
   return write_file(path.Base(url_path), data)
}

func write_file(name string, data []byte) error {
   log.Println("WriteFile", name)
   return os.WriteFile(name, data, os.ModePerm)
}

type flag_set struct {
   address string
   sleep   time.Duration
   filters string
}
