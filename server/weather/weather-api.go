package main

import "fmt"

/**
https://openweathermap.org/current#one
https://api.openweathermap.org/data/2.5/weather?lat={lat}&lon={lon}&appid={API key}

https://openweathermap.org/api/one-call-3
https://api.openweathermap.org/data/3.0/onecall?lat={lat}&lon={lon}&exclude={part}&appid={API key}
*/

type WeatherAPI interface {
	GetUrl(lat, long string) string
}

type Api25 struct {
	apiKey string
}

func (api *Api25) GetUrl(lat, long string) string {
	return fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat={%s}&lon={%s}&appid={%s}", lat, long, api.apiKey)
}

func testFunc(weatherApi WeatherAPI, lat, long string) {
	weatherApi.GetUrl(lat, long)
}

func main() {

	x := &Api25{}

	testFunc(x, "", "")

}
