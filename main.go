package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Weather struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		TzID    string `json:"tz_id"`
		Region  string `json:"region"`
	} `json:"location"`
	Current struct {
		TempF     float64 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	}
	Forecast struct {
		Forecastday []struct {
			Astro struct {
				Sunrise string `json:"sunrise"`
				Sunset  string `json:"sunset"`
			}
			Hour []struct {
				TimeEpoch    int64   `json:"time_epoch"`
				TempF        float64 `json:"temp_f"`
				ChanceOfRain float64 `json:"chance_of_rain"`
				Condition    struct {
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

var q string

func main() {

	if len(os.Args) >= 2 {
		cityFlagCMD := flag.NewFlagSet("city", flag.ExitOnError)
		cityData := cityFlagCMD.String("data", "Los_Angeles", "City Forecast")
		switch os.Args[1] {
		case "city":
			err := cityFlagCMD.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("Cannot parse argument")
				panic(err)
			}
		}
		if cityFlagCMD.Parsed() {
			if *cityData == "" {
				os.Exit(1)
			}
			q = *cityData
		}
	} else {
		q = "Los_Angeles"
	}
	res, err := http.Get("http://api.weatherapi.com/v1/forecast.json?key=d48c3d2b3bad49b7af7180920252603&q=" + q + "&days=7&aqi=no&alerts=no")

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Errorf("Wheather API not available")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("weather.json", body, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	var w Weather

	err = json.Unmarshal(body, &w)
	if err != nil {
		panic(err)
	}

	location, current := w.Location, w.Current

	locTZ, err := time.LoadLocation(w.Location.TzID)
	if err != nil {
		panic(err)
	}
	now := time.Now().In(locTZ)

	d := "7-day Forecast"
	f := fmt.Sprintf("%s\n\n", d)
	color.Cyan(f)

	fmt.Printf(
		"%s\n%s, %s - %.0f, %s\n\n",
		location.Country,
		location.Name,
		location.Region,
		current.TempF,
		current.Condition.Text,
	)

	for day := range w.Forecast.Forecastday {

		hours, sun := w.Forecast.Forecastday[day].Hour, w.Forecast.Forecastday[day].Astro

		fmt.Println()

		fmt.Println("Day:", day)
		fmt.Println()
		s := fmt.Sprintf("\nSunrise:%s\nSunset:%s\n\n", sun.Sunrise, sun.Sunset)
		color.Magenta(s)

		for _, hour := range hours {

			date := time.Unix(hour.TimeEpoch, 0).In(locTZ)

			if date.Before(now) {
				continue
			}

			fDate := date.Format("15:05")
			srise := strings.Split(sun.Sunrise, " ")
			sset := strings.Split(sun.Sunset, " ")

			srise[0] += ":00"
			sset[0] += ":00"

			var sriseRes string
			var ssetRes string

			for _, s := range srise {
				sriseRes += s
			}

			for _, s := range sset {
				ssetRes += s
			}

			inputFormat := "03:04:05PM"

			parsedSunriseTime, err := time.Parse(inputFormat, sriseRes)
			if err != nil {
				log.Fatal("Time cannot be parsed")
			}

			parsedSunsetTime, err := time.Parse(inputFormat, ssetRes)
			if err != nil {
				log.Fatal("Time cannot be parsed")
			}

			var output string
			if parsedSunriseTime.Hour() == date.Hour() {
				output = fmt.Sprintf(
					"%s - %.0fF, %.0f%%, %s\nSunrise: %s",
					fDate,
					hour.TempF,
					hour.ChanceOfRain,
					hour.Condition.Text,
					parsedSunriseTime.Format("15:04"),
				)
			} else if parsedSunsetTime.Hour() == date.Hour() {
				output = fmt.Sprintf(
					"%s - %.0fF, %.0f%%, %s\nSunset: %s",
					fDate,
					hour.TempF,
					hour.ChanceOfRain,
					hour.Condition.Text,
					parsedSunsetTime.Format("15:04"),
				)
			} else {
				output = fmt.Sprintf(
					"%s - %.0fF, %.0f%%, %s\n",
					fDate,
					hour.TempF,
					hour.ChanceOfRain,
					hour.Condition.Text,
				)
			}

			if hour.ChanceOfRain >= 20 && hour.ChanceOfRain < 70 {
				color.Yellow(output)
			} else if hour.ChanceOfRain > 70 {
				color.Red(output)
			} else {
				color.Green(output)
			}
		}
	}

}
