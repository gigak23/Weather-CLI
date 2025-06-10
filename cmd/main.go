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

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: Please set the WEATHER_API_KEY environment variable.")
		os.Exit(1)
	}

	for _, arg := range os.Args {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			printHelp()
			return
		}
	}

	if validateArgs() {
		setQueryValue()
	} else {
		q = "Los_Angeles"
		ql = "en"
	}

	weatherReport(apiKey)

}

// Command-line help function
func printHelp() {
	fmt.Print(`
Usage:
  go run main.go [flag] [value] [optional second flag] [value]

Flags:
  city -data=<city_name>     Specify the city to get weather for (e.g., city -data "Los_Angeles")
  lang -data=<lang_code>     Specify the language code (e.g., lang -data "es")
  -help                      Show this help message

Examples:
  go run main.go
      -> Defaults to Los Angeles and English

  go run main.go city -data="Tokyo"
      -> Weather for Tokyo in English

  go run main.go lang -data="fr"
      -> Weather for Los Angeles in French

  go run main.go city -data="Tokyo" lang -data="ja"
      -> Weather for Tokyo in Japanese
`)
}

// Get weather data
func weatherReport(apiKey string) {

	res, err := http.Get("https://api.weatherapi.com/v1/forecast.json?key=" + apiKey + "&q=" + q + "&days=7&aqi=no&alerts=yes&lang=" + ql)

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

	d := "3-day Forecast"
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

			sunsetLang, ok := translationsSunset[ql]
			if !ok {
				sunsetLang = "Sunset"
			}
			sunriseLang, ok := translationsSunrise[ql]
			if !ok {
				sunriseLang = "Sunrise"
			}

			// Display weather report with sunrise time
			if parsedSunriseTime.Hour() == date.Hour() {

				// Output normal weather report
				chanceOfRain(output, hour.ChanceOfRain)

				color.RGB(255, 165, 0).Println(sunriseLang + " " + parsedSunriseTime.Format("15:04"))

				// Display weather report wtih sunset time
			} else if parsedSunsetTime.Hour() == date.Hour() {

				// Output normal weather report
				chanceOfRain(output, hour.ChanceOfRain)

				color.RGB(160, 32, 240).Println(sunsetLang + " " + parsedSunsetTime.Format("15:04"))

			} else {
				// Output normal weather report
				chanceOfRain(output, hour.ChanceOfRain)
			}

		}
	}
}

// Check for valid number of command line arguments
func validateArgs() bool {
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
	switch {
	case rain < 20:
		color.Green(output)
	case rain < 50:
		color.HiYellow(output)
	case rain < 80:
		color.Yellow(output)
	default:
		color.Red(output)
	}
}

var translationsSunset = map[string]string{
	"ar":     "غروب",
	"bn":     "সূর্যাস্ত",
	"bg":     "Залез",
	"zh":     "日落",
	"zh_tw":  "日落",
	"cs":     "Západ slunce",
	"da":     "Solnedgang",
	"nl":     "Zonsondergang",
	"fi":     "Auringonlasku",
	"fr":     "Coucher de soleil",
	"de":     "Sonnenuntergang",
	"el":     "Ηλιοβασίλεμα",
	"hi":     "सूर्यास्त",
	"hu":     "Napnyugta",
	"it":     "Tramonto",
	"ja":     "日没",
	"jv":     "Sunset",
	"ko":     "일몰",
	"zh_cmn": "日落",
	"mr":     "सूर्यास्त",
	"pl":     "Zachód słońca",
	"pt":     "Por do sol",
	"pa":     "ਸੂਰਜ ਡੁੱਬਣ",
	"ro":     "Apus de soare",
	"ru":     "Закат",
	"sr":     "Залазак сунца",
	"si":     "හිරු බැස යෑමයි",
	"sk":     "Západ slnka",
	"es":     "Atardecer",
	"sv":     "Solnedgång",
	"ta":     "சூரிய அஸ்தமனம்",
	"te":     "సూర్యాస్తమయం",
	"tr":     "Gün batımı",
	"uk":     "Захід сонця",
	"ur":     "غروب آفتاب",
	"vi":     "Hoàng hôn",
	"zh_wuu": "晚霞",
	"zh_hsn": "夕阳",
	"zh_yue": "日落",
	"zu":     "Ukushona kwelanga",
}

var translationsSunrise = map[string]string{
	"ar":     "شروق الشمس",
	"bn":     "সূর্যোদয়",
	"bg":     "Изгрев",
	"zh":     "日出",
	"zh_tw":  "日出",
	"cs":     "Východ slunce",
	"da":     "Solopgang",
	"nl":     "Zonsopgang",
	"fi":     "Auringonnousu",
	"fr":     "Lever du soleil",
	"de":     "Sonnenaufgang",
	"el":     "Ανατολή ηλίου",
	"hi":     "सूर्योदय",
	"hu":     "Napkelte",
	"it":     "Alba",
	"ja":     "日の出",
	"jv":     "Srengenge munggah",
	"ko":     "일출",
	"zh_cmn": "日出",
	"mr":     "सूर्योदय",
	"pl":     "Wschód słońca",
	"pt":     "Nascer do sol",
	"pa":     "ਸੂਰਜ ਚੜ੍ਹਨਾ",
	"ro":     "Răsărit",
	"ru":     "Восход",
	"sr":     "Излазак сунца",
	"si":     "සඳළුව",
	"sk":     "Východ slnka",
	"es":     "Amanecer",
	"sv":     "Soluppgång",
	"ta":     "சூரிய உதயம்",
	"te":     "సూర్యోదయం",
	"tr":     "Gün doğumu",
	"uk":     "Схід сонця",
	"ur":     "طلوع آفتاب",
	"vi":     "Bình minh",
	"zh_wuu": "日出",
	"zh_hsn": "日出",
	"zh_yue": "日出",
	"zu":     "Ukuphuma kwelanga",
}
