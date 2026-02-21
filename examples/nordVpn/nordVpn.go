package main

import (
   "41.neocities.org/service/nordVpn"
   "errors"
   "flag"
   "fmt"
   "io"
   "log"
   "net/http"
   "net/url"
   "os"
   "os/exec"
   "path/filepath"
   "time"
)

func do_country(name, code string) error {
   data, err := read_file(name)
   if err != nil {
      return err
   }
   var loads nordVpn.ServerLoads
   err = loads.Unmarshal(data)
   if err != nil {
      return err
   }
   country, ok := loads.Country(code)
   if !ok {
      return errors.New(".Country")
   }
   data, err = loads.Marshal()
   if err != nil {
      return err
   }
   err = write_file(name, data)
   if err != nil {
      return err
   }
   user, err := output("credential", "-h=api.nordvpn.com", "-k=user")
   if err != nil {
      return err
   }
   password, err := output("credential", "-h=api.nordvpn.com")
   if err != nil {
      return err
   }
   fmt.Println(nordVpn.FormatProxy(string(user), string(password), country))
   return nil
}

func do_write(name string) error {
   servers, err := nordVpn.GetServers(0)
   if err != nil {
      return err
   }
   data, err := nordVpn.GetServerLoads(servers).Marshal()
   if err != nil {
      return err
   }
   return write_file(name, data)
}

func write_file(name string, data []byte) error {
   log.Println("WriteFile", name)
   return os.WriteFile(name, data, os.ModePerm)
}

func read_file(name string) ([]byte, error) {
   file, err := os.Open(name)
   if err != nil {
      return nil, err
   }
   defer file.Close()
   info, err := file.Stat()
   if err != nil {
      return nil, err
   }
   const month = 30 * 24 * time.Hour
   if time.Since(info.ModTime()) >= month {
      return nil, errors.New("ModTime")
   }
   return io.ReadAll(file)
}
func main() {
   log.SetFlags(log.Ltime)
   http.DefaultTransport = &http.Transport{
      Proxy: func(req *http.Request) (*url.URL, error) {
         log.Println(req.Method, req.URL)
         return http.ProxyFromEnvironment(req)
      },
   }
   write := flag.Bool("w", false, "write")
   country_code := flag.String("c", "", "country code")
   flag.Parse()
   cache, err := os.UserCacheDir()
   if err != nil {
      log.Fatal(err)
   }
   cache = filepath.ToSlash(cache) + "/nordVpn/nordVpn.json"
   switch {
   case *country_code != "":
      err = do_country(cache, *country_code)
   case *write:
      err = do_write(cache)
   default:
      flag.Usage()
   }
   if err != nil {
      log.Fatal(err)
   }
}

func output(name string, arg ...string) ([]byte, error) {
   command := exec.Command(name, arg...)
   log.Println("Output", command.Args)
   return command.Output()
}
