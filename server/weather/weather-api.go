package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ykm7/mini-farm-tracker/core"
)

type WeatherResponse struct {
	Coord      *Coordinates `json:"coord,omitempty"`
	Weather    []Weather    `json:"weather,omitempty"`
	Base       string       `json:"base,omitempty"`
	Main       *MainWeather `json:"main,omitempty"`
	Visibility int          `json:"visibility,omitempty"`
	Wind       *Wind        `json:"wind,omitempty"`
	Rain       *Rain        `json:"rain,omitempty"`
	Clouds     *Clouds      `json:"clouds,omitempty"`
	Dt         int64        `json:"dt,omitempty"`
	Sys        *Sys         `json:"sys,omitempty"`
	Timezone   int          `json:"timezone,omitempty"`
	ID         int          `json:"id,omitempty"`
	Name       string       `json:"name,omitempty"`
	Cod        int          `json:"cod,omitempty"`
}

type Coordinates struct {
	Lon float64 `json:"lon,omitempty"`
	Lat float64 `json:"lat,omitempty"`
}

type Weather struct {
	ID          int    `json:"id,omitempty"`
	Main        string `json:"main,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type MainWeather struct {
	Temp      float64 `json:"temp,omitempty"`
	FeelsLike float64 `json:"feels_like,omitempty"`
	TempMin   float64 `json:"temp_min,omitempty"`
	TempMax   float64 `json:"temp_max,omitempty"`
	Pressure  int     `json:"pressure,omitempty"`
	Humidity  int     `json:"humidity,omitempty"`
	SeaLevel  int     `json:"sea_level,omitempty"`
	GrndLevel int     `json:"grnd_level,omitempty"`
}

type Wind struct {
	Speed float64 `json:"speed,omitempty"`
	Deg   int     `json:"deg,omitempty"`
	Gust  float64 `json:"gust,omitempty"`
}

type Rain struct {
	OneHour float64 `json:"1h,omitempty"`
}

type Clouds struct {
	All int `json:"all,omitempty"`
}

type Sys struct {
	Type    int    `json:"type,omitempty"`
	ID      int    `json:"id,omitempty"`
	Country string `json:"country,omitempty"`
	Sunrise int64  `json:"sunrise,omitempty"`
	Sunset  int64  `json:"sunset,omitempty"`
}

/**
https://openweathermap.org/current#one
https://api.openweathermap.org/data/2.5/weather?lat={lat}&lon={lon}&appid={API key}

https://openweathermap.org/api/one-call-3
https://api.openweathermap.org/data/3.0/onecall?lat={lat}&lon={lon}&exclude={part}&appid={API key}
*/

type WeatherAPI interface {
	GetUrl(lat, long string) string
}

type Api25WeatherAPIImpl struct {
	apiKey string
}

func (api *Api25WeatherAPIImpl) GetUrl(lat, long string) string {

	units := "metric"

	return fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&appid=%s&units=%s", lat, long, api.apiKey, units)
}

func testFunc(weatherApi WeatherAPI, lat, long string) string {
	return weatherApi.GetUrl(lat, long)
}

func main() {
	envs := core.ReadEnvs()

	x := &Api25WeatherAPIImpl{apiKey: envs.Open_weather_api}

	// Home - -33.6668553,115.0810599
	url := testFunc(x, "-33.6668553", "115.0810599")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var weather WeatherResponse
	err = json.Unmarshal(body, &weather)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println(string(body))

	log.Println("%+v", weather)
}
