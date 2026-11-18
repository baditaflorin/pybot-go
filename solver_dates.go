package pybot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	datesBetweenRE  = regexp.MustCompile(`(?i)\bhow many days between\s+(\d{4}-\d{2}-\d{2})\s+and\s+(\d{4}-\d{2}-\d{2})\b`)
	relativeDateRE  = regexp.MustCompile(`(?i)\b(\d{1,5})\s+(days?|weeks?|months?|years?)\s+(after|before|from)\s+(\d{4}-\d{2}-\d{2})\b`)
	relativeTodayRE = regexp.MustCompile(`(?i)\b(\d{1,5})\s+(days?|weeks?|months?|years?)\s+(after|before|from)\s+today\b`)
	untilDateRE     = regexp.MustCompile(`(?i)\bhow many days (?:until|from today to)\s+(\d{4}-\d{2}-\d{2})\b`)
	dateIntentRE    = regexp.MustCompile(`(?i)\b(date|days between|days until|weeks from today|months from today|years from today|what day is it|what day will it be)\b`)
	singleDateRE    = regexp.MustCompile(`(?i)\b(?:what day|what date|which weekday|what weekday)\b.*?(\d{4}-\d{2}-\d{2})\b`)
	todayDateRE     = regexp.MustCompile(`(?i)\b(?:what(?:'s| is) today's date|what day is it today|what is the date today)\b`)
	tomorrowRE      = regexp.MustCompile(`(?i)\b(?:what(?:'s| is) the date|what day is it) (tomorrow|yesterday)\b`)
)

func solveDateQuestion(text string, now time.Time) (Response, bool) {
	text = replaceSpokenNumbers(text)
	if match := datesBetweenRE.FindStringSubmatch(text); len(match) == 3 {
		first, errFirst := parseISODate(match[1])
		second, errSecond := parseISODate(match[2])
		if errFirst != nil || errSecond != nil {
			return badDateResponse(), true
		}
		days := int(second.Sub(first).Hours() / 24)
		if days < 0 {
			days = -days
		}
		return Response{Text: fmt.Sprintf("There are %d days between those dates.", days), Module: "reasoning/date"}, true
	}
	if todayDateRE.MatchString(text) {
		date := dateOnly(now)
		return Response{Text: fmt.Sprintf("Today is %s (%s).", date.Format("2006-01-02"), date.Weekday()), Module: "reasoning/date"}, true
	}
	if match := tomorrowRE.FindStringSubmatch(text); len(match) == 2 {
		date := dateOnly(now)
		if strings.EqualFold(match[1], "tomorrow") {
			date = date.AddDate(0, 0, 1)
		} else {
			date = date.AddDate(0, 0, -1)
		}
		return Response{Text: fmt.Sprintf("%s is %s.", date.Format("2006-01-02"), date.Weekday()), Module: "reasoning/date"}, true
	}
	if match := untilDateRE.FindStringSubmatch(text); len(match) == 2 {
		target, err := parseISODate(match[1])
		if err != nil {
			return badDateResponse(), true
		}
		start := dateOnly(now)
		days := int(target.Sub(start).Hours() / 24)
		return Response{Text: fmt.Sprintf("That date is %d days from today (%s).", days, start.Format("2006-01-02")), Module: "reasoning/date"}, true
	}
	if match := relativeDateRE.FindStringSubmatch(text); len(match) == 5 {
		date, err := parseISODate(match[4])
		if err != nil {
			return badDateResponse(), true
		}
		amount, _ := strconv.Atoi(match[1])
		result, err := shiftDate(date, amount, match[2], match[3])
		if err != nil {
			return badDateResponse(), true
		}
		return Response{Text: fmt.Sprintf("%s is %s.", result.Format("2006-01-02"), result.Weekday()), Module: "reasoning/date"}, true
	}
	if match := relativeTodayRE.FindStringSubmatch(text); len(match) == 4 {
		amount, _ := strconv.Atoi(match[1])
		result, err := shiftDate(dateOnly(now), amount, match[2], match[3])
		if err != nil {
			return badDateResponse(), true
		}
		return Response{Text: fmt.Sprintf("%s is %s.", result.Format("2006-01-02"), result.Weekday()), Module: "reasoning/date"}, true
	}
	if match := singleDateRE.FindStringSubmatch(text); len(match) == 2 {
		date, err := parseISODate(match[1])
		if err != nil {
			return badDateResponse(), true
		}
		return Response{Text: fmt.Sprintf("%s is %s.", date.Format("2006-01-02"), date.Weekday()), Module: "reasoning/date"}, true
	}
	if dateIntentRE.MatchString(text) {
		return Response{Text: "I can calculate dates using ISO dates such as 2026-11-18. For example: \"30 days after 2026-11-18\".", Module: "reasoning/date", NeedsClarification: true}, true
	}
	return Response{}, false
}

func parseISODate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, time.UTC)
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func shiftDate(date time.Time, amount int, unit, direction string) (time.Time, error) {
	if amount > 10000 {
		return time.Time{}, fmt.Errorf("date offset too large")
	}
	if strings.EqualFold(direction, "before") {
		amount = -amount
	}
	switch strings.TrimSuffix(strings.ToLower(unit), "s") {
	case "day":
		return date.AddDate(0, 0, amount), nil
	case "week":
		return date.AddDate(0, 0, amount*7), nil
	case "month":
		return addMonthsClamped(date, amount), nil
	case "year":
		return addMonthsClamped(date, amount*12), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported date unit")
	}
}

func addMonthsClamped(date time.Time, months int) time.Time {
	first := time.Date(date.Year(), date.Month()+time.Month(months), 1, 0, 0, 0, 0, time.UTC)
	lastDay := first.AddDate(0, 1, -1).Day()
	day := date.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, time.UTC)
}

func badDateResponse() Response {
	return Response{Text: "I couldn't read that date. Use YYYY-MM-DD, such as 2026-11-18.", Module: "reasoning/date", NeedsClarification: true}
}
