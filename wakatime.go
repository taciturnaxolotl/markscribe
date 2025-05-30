package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func wakatimeData() (WakatimeUserStats, error) {
	// Use hackatime/v1 path for Hackatime compatibility, falls back to standard WakaTime path
	endpoint := wakatimeClient.baseurl + "/users/current/stats/last_7_days"
	if wakatimeClient.baseurl == "https://hackatime.hackclub.com/api/v1/" {
		endpoint = "https://hackatime.hackclub.com/api/hackatime/v1/users/current/stats/last_7_days"
	}
	
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return WakatimeUserStats{}, err // Return empty struct and error
	}

	// Encode API key to base64
	encodedKey := base64.StdEncoding.EncodeToString([]byte(wakatimeClient.apikey))
	
	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encodedKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return WakatimeUserStats{}, err // Return empty struct and error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WakatimeUserStats{}, fmt.Errorf("wakatime API returned status code %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse the JSON response and populate WakatimeUserStats struct
	decoder := json.NewDecoder(resp.Body)
	var stats WakatimeDataRes
	err = decoder.Decode(&stats)
	if err != nil {
		return WakatimeUserStats{}, fmt.Errorf("error decoding response: %w", err)
	}

	return stats.Data, nil
}

func formatTime(hours int, minutes int, seconds int) string {
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func bar(percentage float64, barWidth int) string {
	bar := ""
	for i := 0; i < barWidth; i++ {
		if float64(i) < percentage/(100/float64(barWidth)) {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return fmt.Sprintf("%s  %.2f%%", bar, percentage)
}

func wakatimeCategoryBar(count int, category any) string {
	typedCategory := wakatimeTransformType(category)

	// sort languages by percentage
	for i := range typedCategory {
		if i >= count {
			typedCategory = typedCategory[:count]
			break
		}
		for j := i + 1; j < len(typedCategory); j++ {
			if typedCategory[i].Percent < typedCategory[j].Percent {
				typedCategory[i], typedCategory[j] = typedCategory[j], typedCategory[i]
			}
		}
	}

	// get the longest name and time
	longestName, longestTime := wakatimeOffsets(count, typedCategory)

	for i, c := range typedCategory {
		if i >= count {
			break
		}
		typedCategory[i].Name = fmt.Sprintf("%-*s", longestName+2, c.Name)
		typedCategory[i].Digital = fmt.Sprintf("%-*s", longestTime+2, formatTime(c.Hours, c.Minutes, c.Seconds))
	}

	// generate the lines in the format: name bar percent%
	var lines []string
	for _, c := range typedCategory {
		lines = append(lines, fmt.Sprintf("%s %s %s", c.Name, c.Digital, bar(c.Percent, 25)))
	}

	return strings.Join(lines, "\n")
}

func wakatimeDoubleCategoryBar(title string, category any, title2 string, category2 any, count int) string {
	typedCategory := wakatimeTransformType(category)
	typedCategory2 := wakatimeTransformType(category2)

	// sort languages by percentage
	for i := range typedCategory {
		if i >= count {
			typedCategory = typedCategory[:count]
			break
		}
		for j := i + 1; j < len(typedCategory); j++ {
			if typedCategory[i].Percent < typedCategory[j].Percent {
				typedCategory[i], typedCategory[j] = typedCategory[j], typedCategory[i]
			}
		}
	}

	for i := range typedCategory2 {
		if i >= count {
			typedCategory2 = typedCategory2[:count]
			break
		}
		for j := i + 1; j < len(typedCategory2); j++ {
			if typedCategory2[i].Percent < typedCategory2[j].Percent {
				typedCategory2[i], typedCategory2[j] = typedCategory2[j], typedCategory2[i]
			}
		}
	}

	// get the longest name and time from both categories and pick larger values
	longestName1, longestTime1 := wakatimeOffsets(count, typedCategory)
	longestName2, longestTime2 := wakatimeOffsets(count, typedCategory2)
	longestName := max(longestName2, longestName1)
	longestTime := max(longestTime2, longestTime1)

	for i, c := range typedCategory {
		if i >= count {
			break
		}
		typedCategory[i].Name = fmt.Sprintf("%-*s", longestName+2, c.Name)
		typedCategory[i].Digital = fmt.Sprintf("%-*s", longestTime+2, formatTime(c.Hours, c.Minutes, c.Seconds))
	}

	for i, c := range typedCategory2 {
		if i >= count {
			break
		}
		typedCategory2[i].Name = fmt.Sprintf("%-*s", longestName+2, c.Name)
		typedCategory2[i].Digital = fmt.Sprintf("%-*s", longestTime+2, formatTime(c.Hours, c.Minutes, c.Seconds))
	}

	// generate the lines in the format: name bar percent%
	var lines []string

	lines = append(lines, title)
	for _, c := range typedCategory {
		lines = append(lines, fmt.Sprintf("%s %s %s", c.Name, c.Digital, bar(c.Percent, 25)))
	}

	lines = append(lines, "")

	lines = append(lines, title2)
	for _, c := range typedCategory2 {
		lines = append(lines, fmt.Sprintf("%s %s %s", c.Name, c.Digital, bar(c.Percent, 25)))
	}

	return strings.Join(lines, "\n")
}

func wakatimeTransformType(category any) []WakatimeCategoryType {
	var typedCategory []WakatimeCategoryType

	switch v := category.(type) {
	case []WakatimeCategoryType:
		typedCategory = v
	case []WakatimeMachines:
		// Convert WakatimeMachines to WakatimeCategoryType
		typedCategory = make([]WakatimeCategoryType, len(v))
		for i, machine := range v {
			typedCategory[i] = WakatimeCategoryType{
				Name:         machine.Name,
				TotalSeconds: machine.TotalSeconds,
				Percent:      machine.Percent,
				Digital:      machine.Digital,
				Text:         machine.Text,
				Hours:        machine.Hours,
				Minutes:      machine.Minutes,
				Seconds:      machine.Seconds,
			}
		}
	default:
		panic("unknown category type")
	}

	return typedCategory
}

func wakatimeOffsets(count int, typedCategory []WakatimeCategoryType) (int, int) {
	// pad the name of the language so that they are all equal in lengh to the longest name plus 2 spaces
	longestName := 0
	longestTime := 0
	for i, c := range typedCategory {
		if i >= count {
			break
		}
		if len(c.Name) > longestName {
			longestName = len(c.Name)
		}
		time := len(formatTime(c.Hours, c.Minutes, c.Seconds))
		if time > longestTime {
			longestTime = time
		}
	}

	return longestName, longestTime
}
