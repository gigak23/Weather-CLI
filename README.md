# 🌦️ Weather CLI Tool

A simple and customizable command-line interface (CLI) tool built in Go that fetches real-time weather data and a 3-day forecast using the [WeatherAPI](https://www.weatherapi.com/). It supports querying by city and changing the output language.

---

## 🚀 Features

- Fetches current weather conditions and 3-day forecasts.
- Supports multiple languages for condition descriptions and labels (e.g., sunrise/sunset).
- Automatically adjusts for the local timezone of the queried city.
- Writes the full API response to a `weather.json` file for reference or further processing.

---

## 🛠️ Requirements

- Go 1.18 or later
- Set your API key via environment variable:

bash
```bash
export WEATHER_API_KEY="your_api_key_here"
```
powershell
```powershell
$env:WEATHER_API_KEY="your_api_key_here"
```
cmd
```cmd
set WEATHER_API_KEY=your_api_key_here
```

---

## 📦 Installation

### Option 1: Download Binaries (Recommended)

> Binaries for Linux, macOS, and Windows will be available in the [Releases](https://github.com/gigak23/Weather-CLI/releases/tag/v1.0.0) section of the GitHub repo.

1. Download the binary for your operating system.
2. Rename the file to `weather` (or `weather.exe` for Windows).
3. Move it to a directory in your `PATH` (e.g., `/usr/local/bin` or `C:\Windows\System32`).

### Option 2: Clone and Build from Source

```bash
git clone https://github.com/gigak23/Weather-CLI.git
cd Weather-CLI
go build -o weather main.go
```

Then run with:

```bash
./weather
```

---

## 🌍 Usage

### Basic Command

```bash
./weather
```

Displays the weather forecast for **Los Angeles** in **English**.

---

### Command Flags

| Flag      | Description                           | Example                                  |
|-----------|---------------------------------------|------------------------------------------|
| `city`    | Specify a city name                   | `city -data="Tokyo"`                     |
| `lang`    | Set language code for translations    | `lang -data="fr"`                        |
| `-help`   | Show usage instructions               | `./weather -help`                        |

---

### Example Commands

```bash
./weather city -data="New_York"
./weather lang -data="es"
./weather city -data="Paris" lang -data="fr"
./weather -help
```

---

## 🧾 Notes

- The city input should use underscores (`_`) instead of spaces (e.g., `New_York`, `San_Francisco`).
- The response from the WeatherAPI is saved to a local file named `weather.json`.
- All times are displayed in the queried city’s local timezone.

---

## 📄 License

MIT License
