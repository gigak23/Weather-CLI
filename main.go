package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Weather data
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
			Date  string `json:"date"`
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

// Global query value
var q string

func main() {

	if validateArgs() {
		setQueryValue()
	} else {
		q = "Los_Angeles"
	}
	weatherReport()

}

// Get weather data
func weatherReport() {
	res, err := http.Get("https://api.weatherapi.com/v1/forecast.json?key=d48c3d2b3bad49b7af7180920252603&q=" + q + "&days=7&aqi=no&alerts=no")

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatal("Weather API not available")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("weather.json", body, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	outputData(body)

}

// Output weather data
func outputData(body []byte) {

	var w Weather

	err := json.Unmarshal(body, &w)
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

		d, hours, sun := w.Forecast.Forecastday[day].Date, w.Forecast.Forecastday[day].Hour, w.Forecast.Forecastday[day].Astro
		weekDay := dayOfTheWeek(d)
		fmt.Println()
		fmt.Println(weekDay)
		fmt.Println()
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

// Check for valid number of command line arguments
func validateArgs() bool {
	return len(os.Args) >= 2
}

// Set the city to gather data from
func setQueryValue() {
	cityFlagCMD := flag.NewFlagSet("city", flag.ExitOnError)
	cityData := cityFlagCMD.String("data", "", "City Forecast")
	switch os.Args[1] {
	case "city":
		err := cityFlagCMD.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Cannot parse argument")
			panic(err)
		}
	default:
		*cityData = "Los_Angeles"
	}
	if cityFlagCMD.Parsed() {
		if *cityData == "" {
			os.Exit(1)
		}
		q = *cityData
	}
}

// Find day of the week
func dayOfTheWeek(date string) string {

	t := []int{0, 3, 2, 5, 0, 3, 5, 1, 4, 6, 2, 4}

	splitDate := strings.Split(date, "-")

	year, err := strconv.Atoi(splitDate[0])
	if err != nil {
		log.Fatal("Cannot parse date")
	}
	month, err := strconv.Atoi(splitDate[1])
	if err != nil {
		log.Fatal("Cannot parse date")
	}
	day, err := strconv.Atoi(splitDate[2])
	if err != nil {
		log.Fatal("Cannot parse date")
	}

	if month < 3 {
		year -= 1
	}

	dayValue := (year + year/4 - year/100 + year/400 + t[month-1] + day) % 7
	var weekDay string

	switch dayValue {
	case 0:
		weekDay = "Sunday"
	case 1:
		weekDay = "Monday"
	case 2:
		weekDay = "Tuesday"
	case 3:
		weekDay = "Wednesday"
	case 4:
		weekDay = "Thursday"
	case 5:
		weekDay = "Friday"
	case 6:
		weekDay = "Saturday"
	}

	return weekDay

}
