package main

import (
  "fmt"
  "strings"
  "os"
  "bytes"
  "encoding/base64"
  "encoding/json"
  "net/http"
  "io/ioutil"
)

type Listing struct {
  location  string
  price     int
  floor     string
  facing    string
  lift      string
  size      int
  status    string
  rooms     int
  baths     int
}

func ternary (b bool, opt1 string, opt2 string) string {
  if b {
    return opt1
  }
  return opt2
}

func defaultString (s interface{}, def string) string {
  if s == nil {
    return def
  }
  return s.(string)
}

func defaultBool (b interface{}, def bool) bool {
  if b == nil {
    return def
  }
  return b.(bool)
}

func getAccessToken () string {
  idealista_user := os.Getenv("IDEALISTA_GO_API_USER")
  idealista_pass := os.Getenv("IDEALISTA_GO_API_PASS")
  encoded_credentials :=
    base64.StdEncoding.EncodeToString([]byte(idealista_user + ":" + idealista_pass))
  idealista_auth_url := "https://api.idealista.com/oauth/token"

  req_body := []byte(`grant_type=client_credentials`)
  req, err := http.NewRequest("POST", idealista_auth_url, bytes.NewBuffer(req_body))
  req.Header.Add("Authorization", "Basic " + encoded_credentials)
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("Error on response:", err)
  }

  raw_token, _ := ioutil.ReadAll(resp.Body)
  var data map[string]interface{}
  json.Unmarshal(raw_token, &data)
  return data["access_token"].(string)
}

func getListings (token string, ids []string) map[string]Listing {
  idealista_search_url := "https://api.idealista.com/3.5/es/search?propertyType=homes" +
    "&locale=es&maxItems=50&numPage=1&operation=sale&order=publicationDate" +
    "&sort=desc&apikey=" + os.Getenv("IDEALISTA_API_USER") + "&language=es&" +
    "adIds=" + strings.Join(ids, "&adIds=") + "&center=37.383,-5.986&distance=50000"
  req, err := http.NewRequest("POST", idealista_search_url, nil)
  req.Header.Add("Authorization", "Bearer " + token)

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("Error on response:", err)
  }
  raw_listings, _ := ioutil.ReadAll(resp.Body)

  var data map[string]interface{}
  json.Unmarshal(raw_listings, &data)
  element_list := data["elementList"].([]interface{})

  results := make(map[string]Listing)
  for _, l := range element_list {
    l_map := l.(map[string]interface{})
    results[l_map["propertyCode"].(string)] = Listing{
      location: l_map["address"].(string),
      price: int(l_map["price"].(float64)),
      floor: defaultString(l_map["floor"], "Casa"),
      facing: ternary(l_map["exterior"].(bool), "Exterior", "Interior"),
      lift: ternary(defaultBool(l_map["hasLift"], true), "con", "sin"),
      size: int(l_map["size"].(float64)),
      status: defaultString(l_map["status"], ""),
      rooms: int(l_map["rooms"].(float64)),
      baths: int(l_map["bathrooms"].(float64)),
    }
  }
  return results
}

func getCSVRow (row string, l Listing) []string {
  return []string{
    strings.Split(row, ",")[0],
    strings.ReplaceAll(l.location, ",", "-"),
    fmt.Sprintf("%v", l.price),
    l.floor + "º " + l.facing + " " + l.lift + " ascensor",
    fmt.Sprintf("%v", l.size),
    "",
    l.status,
    fmt.Sprintf("%v", l.rooms),
    fmt.Sprintf("%v", l.baths),
    "",
    "",
    strings.Split(row, ",")[11],
    strings.Split(row, ",")[12],
    strings.Split(row, ",")[13],
    strings.Split(row, ",")[14],
    "",
  }
}

func main() {
  token := getAccessToken()
  csv_string := getPisosContent()

  var missing_listing_ids []string
  for _, row := range strings.Split(csv_string, "\n") {
    if strings.Contains(row, "idealista") {
      row_listing_id := strings.Split(row, "eble/")[1][:8]
      missing_listing_ids = append(missing_listing_ids, row_listing_id)
    }
  }
  listings := getListings(token, missing_listing_ids)

  new_csv := ""
  for _, row := range strings.Split(csv_string, "\n") {
    if strings.Contains(row, "idealista") && strings.Split(row, ",")[15] === "sold" {
      row_listing_id := strings.Split(row, "eble/")[1][:8]
      l, ok := listings[row_listing_id]
      if ok {
        new_csv = new_csv + strings.Join(getCSVRow(row, l), ",") + "\n"
      } else {
        new_csv = new_csv + row[:len(row)-1] + "sold\n"
      }
    } else {
      new_csv = new_csv + row[:len(row)-1] + "\n"
    }
  }
  ioutil.WriteFile("./pisos.csv", []byte(new_csv), 0644)
}
