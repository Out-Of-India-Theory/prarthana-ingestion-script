package util

import (
	"log"
	"net/http"
	"regexp"
	"strings"
)

func ExtractLanguageKeys(fieldMap map[string]int, prefix string) map[string]int {
	keys := make(map[string]int)
	for key, value := range fieldMap {
		if strings.HasPrefix(key, prefix) {
			lang := strings.TrimPrefix(key, prefix)
			keys[lang] = value
		}
	}
	return keys
}

func GetSplittedString(str string) []string {
	strs := strings.Split(str, ",")
	var rStrs []string
	for _, s := range strs {
		if s != "" {
			rStrs = append(rStrs, strings.TrimSpace(s))
		}
	}
	return rStrs
}

func GetDaysFromTitle(title string) []int {
	daysMap := map[string]int{
		"monday prayer":    1,
		"tuesday prayer":   2,
		"wednesday prayer": 3,
		"thursday prayer":  4,
		"friday prayer":    5,
		"saturday prayer":  6,
		"sunday prayer":    0,
	}

	titleLower := strings.ToLower(title)
	if day, exists := daysMap[titleLower]; exists {
		return []int{day}
	}
	return nil
}

func UrlExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		log.Printf("Error checking URL: %s, %v", url, err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func SanitizeString(input string) string {
	// Remove square brackets and their contents
	reSquareBrackets := regexp.MustCompile(`\[.*?\]`)
	cleaned := reSquareBrackets.ReplaceAllString(input, "")

	// Remove parentheses
	reParentheses := regexp.MustCompile(`\(|\)`)
	cleaned = reParentheses.ReplaceAllString(cleaned, "")

	// Replace hyphens and spaces with underscores
	reSpaceHyphen := regexp.MustCompile(`[\s\-]+`)
	cleaned = reSpaceHyphen.ReplaceAllString(cleaned, "_")

	// Trim leading and trailing underscores, if any
	cleaned = strings.Trim(cleaned, "_")

	return cleaned
}
