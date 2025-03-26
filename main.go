package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
)

type Weather struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		TzID    string `json:"tz_id"`
	} `json:"location"`
	Current struct {
		TempF     float64 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	}
	Forecast struct {
		Forecastday []struct {
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
	res, err := http.Get("http://api.weatherapi.com/v1/forecast.json?key=d48c3d2b3bad49b7af7180920252603&q=" + q + "&days=1&aqi=no&alerts=no")

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

	location, current, hours := w.Location, w.Current, w.Forecast.Forecastday[0].Hour
	locTZ, err := time.LoadLocation(w.Location.TzID)
	if err != nil {
		panic(err)
	}
	now := time.Now().In(locTZ)

	fmt.Printf(
		"%s, %s - %.0f, %s\n",
		location.Name,
		location.Country,
		current.TempF,
		current.Condition.Text,
	)

	for _, hour := range hours {

		date := time.Unix(hour.TimeEpoch, 0).In(locTZ)

		if date.Before(now) {
			continue
		}

		output := fmt.Sprintf(
			"%s - %.0fF, %.0f%%, %s\n",
			date.Format("15:05"),
			hour.TempF,
			hour.ChanceOfRain,
			hour.Condition.Text,
		)

		if hour.ChanceOfRain >= 20 && hour.ChanceOfRain <= 50 {
			color.Yellow(output)
		} else if hour.ChanceOfRain > 50 {
			color.Red(output)
		} else {
			color.Green(output)
		}

	}

}
