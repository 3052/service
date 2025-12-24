package nordVpn

import (
   "encoding/json"
   "net/http"
   "net/url"
   "strconv"
   "strings"
)

// limit <= -1 for default
// limit == 0 for all
func GetServers(limit int) ([]Server, error) {
   var req http.Request
   req.Header = http.Header{}
   req.URL = &url.URL{
      Scheme: "https",
      Host: "api.nordvpn.com",
      Path: "/v1/servers",
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

type ServerLoads []*ServerLoad

func (s ServerLoads) Marshal() ([]byte, error) {
   return json.Marshal(s)
}

func (s *ServerLoads) Unmarshal(data []byte) error {
   return json.Unmarshal(data, s)
}

type ServerLoad struct {
   Count    int
   Country  string
   City     string
   Hostname string
}

func GetServerLoads(servers []Server) ServerLoads {
   loads := make(ServerLoads, 0, len(servers))
   for _, server_var := range servers {
      if server_var.proxy_ssl() {
         var load ServerLoad
         load.Hostname = server_var.Hostname
         for _, location := range server_var.Locations {
            load.Country = location.Country.Code
            load.City = location.Country.City.DnsName
         }
         loads = append(loads, &load)
      }
   }
   return loads
}

func (s *Server) proxy_ssl() bool {
   for _, technology := range s.Technologies {
      if technology.Identifier == "proxy_ssl" {
         return true
      }
   }
   return false
}

func (s ServerLoads) Country(code string) (string, bool) {
   var load *ServerLoad
   for _, load1 := range s {
      if load1.Country == code {
         if load != nil {
            if load1.Count >= load.Count {
               continue
            }
         }
         load = load1
      }
   }
   if load != nil {
      load.Count++
      return load.Hostname, true
   }
   return "", false
}
