package justWatch

import (
   "cmp"
   "slices"
)

type Offer struct {
   ElementCount     int
   MonetizationType string
   StandardWebUrl   string
}

type Locale struct {
   FullLocale  string
   Country     string
   CountryName string
}

type EnrichedOffer struct {
   Locale *Locale
   Offer  *Offer
}

// Deduplicate removes true duplicates where both the Offer and Locale are identical.
func Deduplicate(offers []EnrichedOffer) []EnrichedOffer {
   // 1. Sort the slice. This brings identical EnrichedOffers next to each other.
   // This part is correct as it compares the underlying values.
   slices.SortFunc(offers, func(a, b EnrichedOffer) int {
      if n := cmp.Compare(a.Offer.StandardWebUrl, b.Offer.StandardWebUrl); n != 0 {
         return n
      }
      if n := cmp.Compare(a.Offer.MonetizationType, b.Offer.MonetizationType); n != 0 {
         return n
      }
      if n := cmp.Compare(a.Offer.ElementCount, b.Offer.ElementCount); n != 0 {
         return n
      }
      return cmp.Compare(a.Locale.FullLocale, b.Locale.FullLocale)
   })
   // 2. Compact the sorted slice, removing consecutive duplicates.
   // CORRECTED: The comparison function now compares the actual data fields,
   // not the memory addresses of the pointers.
   return slices.CompactFunc(offers, func(a, b EnrichedOffer) bool {
      return a.Offer.StandardWebUrl == b.Offer.StandardWebUrl &&
         a.Offer.MonetizationType == b.Offer.MonetizationType &&
         a.Offer.ElementCount == b.Offer.ElementCount &&
         a.Locale.FullLocale == b.Locale.FullLocale
   })
}
