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

/*

GOAL: FILTER SPECIFIC CITIES BY COUNTRY IN CASE OF DUPLICATE CITIES

*/

// Global query value
var q string

// Global lang value
var ql string

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
				Uv           float64 `json:"uv"`
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

func main() {

	if validateArgs() {
		setQueryValue()
	} else {
		q = "Los_Angeles"
		ql = "en"
	}
	weatherReport()

}

// Get weather data
func weatherReport() {
	res, err := http.Get("https://api.weatherapi.com/v1/forecast.json?key=d48c3d2b3bad49b7af7180920252603&q=" + q + "&days=7&aqi=no&alerts=yes&lang=" + ql)

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
	f := fmt.Sprintf("\n%s\n", d)
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

			uvText := "UV-Index: N/A"
			if hour.Uv > 0 {
				uvText = uvIndex(hour.Uv)
			}

			output := fmt.Sprintf(
				"%s - %.0fF, %.0f%%, %s, %s",
				fDate,
				hour.TempF,
				hour.ChanceOfRain,
				hour.Condition.Text,
				uvText,
			)

			// Display weather report with sunrise time
			if parsedSunriseTime.Hour() == date.Hour() {

				// Output normal weather report
				chanceOfRain(output, hour.ChanceOfRain)

				color.RGB(255, 165, 0).Println("Sunrise: " + parsedSunriseTime.Format("15:04"))

				// Display weather report wtih sunset time
			} else if parsedSunsetTime.Hour() == date.Hour() {

				// Output normal weather report
				chanceOfRain(output, hour.ChanceOfRain)

				color.RGB(160, 32, 240).Println("Sunset: " + parsedSunsetTime.Format("15:04"))

			} else {
				// Output normal weather report
				chanceOfRain(output, hour.ChanceOfRain)
			}

		}
	}
}

// Check for valid number of command line arguments
func validateArgs() bool {
	fmt.Println(len(os.Args))
	return len(os.Args) >= 2
}

// if length of os.Args is 3 check if it is lang or city
// if length of os.Args is 5 do normal check

// Set the city to gather data from
func setQueryValue() {
	cityFlagCMD := flag.NewFlagSet("city", flag.ExitOnError)
	langFlagCMD := flag.NewFlagSet("lang", flag.ExitOnError)
	cityData := cityFlagCMD.String("data", "", "City Forecast")
	langData := langFlagCMD.String("data", "", "Lang Code")

	// Cecks number of CMD arguments
	switch len(os.Args) {

	// Cheks to query lang or city if args length is 3
	case 3:
		switch os.Args[1] {
		case "city":
			if err := cityFlagCMD.Parse(os.Args[2:3]); err != nil {
				log.Println("Cannot parse cmd flag", err)
			}

		case "lang":
			if err := langFlagCMD.Parse(os.Args[2:3]); err != nil {
				log.Println("Cannot parse cmd flag", err)
			}
		default:
			*cityData = "Los_Angeles"
			*langData = "en"
		}

		if cityFlagCMD.Parsed() {
			if *cityData == "" {
				os.Exit(1)
			}
			q = *cityData
			ql = "en"
		}
		if langFlagCMD.Parsed() {
			if *langData == "" {
				os.Exit(1)
			}
			ql = *langData
			q = "Los_Angeles"
		}

		// Queries lang and city if args length is 5
	case 5:
		switch os.Args[1] {
		case "city":
			if err := cityFlagCMD.Parse(os.Args[2:3]); err != nil {
				log.Println("Cannot parse cmd flag", err)
			}

		case "lang":
			if err := langFlagCMD.Parse(os.Args[2:3]); err != nil {
				log.Println("Cannot parse cmd flag", err)
			}

		}

		switch os.Args[3] {
		case "city":
			if err := cityFlagCMD.Parse(os.Args[4:]); err != nil {
				log.Println("Cannot parse cmd flag", err)
			}

		case "lang":
			if err := langFlagCMD.Parse(os.Args[4:]); err != nil {
				log.Println("Cannot parse cmd flag", err)
			}

		}
		if cityFlagCMD.Parsed() {
			if *cityData == "" {
				fmt.Println("city error")
				os.Exit(1)
			}
			q = *cityData
		}
		if langFlagCMD.Parsed() {
			if *langData == "" {
				fmt.Println("lang error")
				os.Exit(1)
			}
			ql = *langData
		}

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

// Specify color for UV-Index
func uvIndex(uv float64) string {
	var uvString string
	f := strconv.FormatFloat(uv, 'f', 1, 64)
	if uv <= 2 {
		uvString = color.GreenString("UV-Index: " + f)
	} else if uv >= 3 && uv <= 5 {
		uvString = color.HiWhiteString("UV-Index: " + f)
	} else if uv >= 6 && uv <= 7 {
		uvString = color.YellowString("UV-Index: " + f)
	} else if uv >= 8 && uv <= 10 {
		uvString = color.RedString("UV-Index: " + f)
	} else if uv >= 11 {
		uvString = color.BlueString("UV-Index: " + f)
	}

	return uvString
}

// Outputs line of certain color based on chance of rain
func chanceOfRain(output string, rain float64) {

	output = strings.TrimSpace(output)

	if rain >= 20 && rain < 70 {
		fmt.Println(color.YellowString(output))
	} else if rain > 70 {
		fmt.Println(color.RedString(output))
	} else {
		fmt.Println(color.GreenString(output))
	}
}
