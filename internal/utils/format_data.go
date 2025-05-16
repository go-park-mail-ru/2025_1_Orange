package utils

import (
	"fmt"
	"strings"
	"time"
)

func FormatDataToString(date time.Time) string {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	return fmt.Sprintf("%d %s %d г.", date.Day(), months[date.Month()-1], date.Year())
}

func FormatWorkExperienceDate(startDateStr string, endDateStr string, untilNow bool) (string, error) {
	months := []string{
		"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return "", fmt.Errorf("неправильный формат даты начала работы: %v", err)
	}

	var endDate time.Time
	if untilNow {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return "", fmt.Errorf("неправильный формат даты окончания работы: %v", err)
		}
	}

	return fmt.Sprintf("%s %d — %s %d", months[startDate.Month()-1], startDate.Year(), months[endDate.Month()-1], endDate.Year()), nil
}

func ExtractYearFromDate(dateStr string) (string, error) {
	parts := strings.Split(dateStr, "-")
	if len(parts) < 1 {
		return "", fmt.Errorf("неверный формат даты")
	}
	return parts[0], nil
}
