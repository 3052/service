package nordVpn

import (
   "encoding/json"
   "net/http"
   "net/url"
   "strconv"
   "strings"
)

func (s *Server) ProxySsl() bool {
   for _, technology := range s.Technologies {
      if technology.Identifier == "proxy_ssl" {
         return true
      }
   }
   return false
}

func (s *Server) Country(code string) bool {
   for _, location := range s.Locations {
      if location.Country.Code == code {
         return true
      }
   }
   return false
}

type Server struct {
   Hostname     string
   Status       string
   Technologies []struct {
      Identifier string
   }
   Locations []struct {
      Country struct {
         City struct {
            DnsName string `json:"dns_name"`
         }
         Code string
      }
   }
}

// limit <= -1 for default
// limit == 0 for all
func GetServers(limit int) ([]Server, error) {
   var req http.Request
   req.Header = http.Header{}
   req.URL = &url.URL{
      Scheme: "https",
      Host:   "api.nordvpn.com",
      Path:   "/v1/servers",
   }
   if limit >= 0 {
      req.URL.RawQuery = "limit=" + strconv.Itoa(limit)
   }
   resp, err := http.DefaultClient.Do(&req)
   if err != nil {
      return nil, err
   }
   defer resp.Body.Close()
   var result []Server
   err = json.NewDecoder(resp.Body).Decode(&result)
   if err != nil {
      return nil, err
   }
   return result, nil
}

func FormatProxy(username, password, hostname string) string {
   var data strings.Builder
   data.WriteString("https://")
   data.WriteString(username)
   data.WriteByte(':')
   data.WriteString(password)
   data.WriteByte('@')
   data.WriteString(hostname)
   data.WriteString(":89")
   return data.String()
}
