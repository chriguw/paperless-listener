package naming

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	documentIDRegex = regexp.MustCompile(`/documents/(\d+)/?`)
	yearRegex       = regexp.MustCompile(`\b\d{2}\.\d{2}\.(\d{2}|\d{4})\b`)
	amountRegex     = regexp.MustCompile(`CHF\s+([\d']+(?:\.\d{1,2})?)`)
	dateRegex       = regexp.MustCompile(`\d{1,2}\.\s*(Januar|Februar|März|April|Mai|Juni|Juli|August|September|Oktober|November|Dezember)\s+(\d{4})`)
)

var monthMap = map[string]string{
	"Januar":    "01",
	"Februar":   "02",
	"März":      "03",
	"April":     "04",
	"Mai":       "05",
	"Juni":      "06",
	"Juli":      "07",
	"August":    "08",
	"September": "09",
	"Oktober":   "10",
	"November":  "11",
	"Dezember":  "12",
}

type Result struct {
	Title1     string
	Title2     string
	Title3     string
	Title4     string
	Year       string
	DateString string
	Amount     float64
}

func CheckTitles(content string, title1, title2, title3, title4 map[string]string, yearKeywords, amountKeywords []string) Result {
	res := Result{}

	for _, line := range strings.Split(content, "\n") {
		if res.DateString == "" {
			res.DateString = findDate(strings.TrimSpace(line))
		}

		res.Title1 = firstMatch(res.Title1, line, title1)
		res.Title2 = firstMatch(res.Title2, line, title2)
		res.Title3 = firstMatch(res.Title3, line, title3)
		res.Title4 = firstMatch(res.Title4, line, title4)

		if res.Year == "" {
			for _, kw := range yearKeywords {
				if strings.Contains(line, kw) {
					year, err := extractYear(line)
					if err == nil {
						res.Year = year
					}
					break
				}
			}
		}

		if res.Amount == 0 {
			for _, kw := range amountKeywords {
				if strings.Contains(line, kw) {
					amount, err := extractCHFAmount(line)
					if err == nil {
						res.Amount = amount
					}
					break
				}
			}
		}
	}

	return res
}

func BuildFinalTitle(title1, title2, title3, title4, dateString, year string) string {
	titles := make([]string, 0, 6)
	if title1 != "" {
		titles = append(titles, title1)
	}
	if title2 != "" {
		titles = append(titles, title2)
	}
	if title3 != "" && !strings.Contains(strings.ToLower(title2), strings.ToLower(title3)) {
		titles = append(titles, title3)
	}
	if title4 != "" {
		titles = append(titles, title4)
	}
	if dateString != "" {
		titles = append(titles, dateString)
	}
	if year != "" && dateString == "" {
		titles = append(titles, year)
	}
	return strings.Join(titles, "_")
}

func ExtractDocumentID(input string) (string, error) {
	var data struct {
		Pfad string `json:"pfad"`
	}

	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	match := documentIDRegex.FindStringSubmatch(data.Pfad)
	if len(match) < 2 {
		return "", fmt.Errorf("document ID not found in URL")
	}

	return match[1], nil
}

func firstMatch(current, line string, lookup map[string]string) string {
	if current != "" {
		return current
	}
	for key, value := range lookup {
		if strings.Contains(strings.ToLower(line), strings.ToLower(key)) {
			return value
		}
	}
	return ""
}

func extractYear(input string) (string, error) {
	matches := yearRegex.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", fmt.Errorf("no date found in input")
	}

	year := matches[1]
	if len(year) == 2 {
		y, _ := strconv.Atoi(year)
		year = fmt.Sprintf("20%02d", y)
	}
	return year, nil
}

func extractCHFAmount(text string) (float64, error) {
	match := amountRegex.FindStringSubmatch(text)
	if match == nil {
		return 0, fmt.Errorf("no CHF amount found in text")
	}

	cleaned := strings.ReplaceAll(match[1], "'", "")
	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount: %w", err)
	}
	return amount, nil
}

func findDate(text string) string {
	m := dateRegex.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return m[2] + "_" + monthMap[m[1]]
}

