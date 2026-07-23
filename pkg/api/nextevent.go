package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Ошибки бизнес-логики при обработке правил повторения задач. 
var (
	ErrEmptyRepeat   = errors.New("правило повторения не может быть пустой строкой")
	ErrUnknownRule   = errors.New("указано неизвестное правило повторения")
	ErrInvalidFormat = errors.New("нарушен формат параметров правила")
)

// NextDate вычисляет ближайшую дату выполнения задачи строго после даты now
// согласно стартовой даты dstart правилу повторения repeat.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// Очищаем входные параметры от пробелов по краям.
	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", ErrEmptyRepeat
	}
	
	startDate, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("неверный формат стартовой даты %q: %w", dstart, err)
	}

	// Сбрасыаем время у now для корректного сравнения календарных дан.
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	parts := strings.Fields(repeat)
	rule := parts[0]

	switch rule {
	case "y":		
		next := startDate
		for {
			next := next.AddDate(1, 0, 0)
			if next.After(now) {
				return next.Format(dateFormat), nil
			}
		}
	case "d":
		if len(parts) < 2 {
			return "", fmt.Errorf("%w: для правила 'd' не указано количество дней", ErrInvalidFormat)
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("%w: некорректный интервал дней (ожидается от 1 до 400)", ErrInvalidFormat)
		}

		// Выполняем математический прыжок по дням, если стартовая дата далеко в прошлом. 
		next := startDate
		if next.Before(now) {
			diffDays := int(now.Sub(next).Hours() / 24)
			if diffDays > days {				
				interfals := diffDays / days
				next = next.AddDate(0, 0, interfals*days)
			}
		}
		
		for !next.After(now) {
			next = next.AddDate(0, 0, days)
		}
		return next.Format(dateFormat), nil

	case "w":
		if len(parts) < 2 {
			return "", fmt.Errorf("%w: для правила 'w' не указаны дни недели", ErrInvalidFormat)
		}

		// Массив маски дней недели (индекс 0 - воскресенье, 1..6 - понедельник .. суббота).
		var activeDays [7]bool
		dayStrings := strings.Split(parts[1], ",")

		for _, dayStr := range dayStrings {
			dayNum, err := strconv.Atoi(dayStr)
			if err != nil || dayNum < 1 || dayNum > 7 {
				return "", fmt.Errorf("%w: не корректный день недели %q (ожидается от 1 до 7)", ErrInvalidFormat, dayStr)
			}
			// Приводим пользовательский день (7 = вс) к внутреннему формату time.Weekday (0 = вс).
			if dayNum == 7 {
				dayNum = 0
			}
			activeDays[dayNum] = true
		}
		
		next := startDate
		for {
			next = next.AddDate(0, 0, 1)
			if activeDays[next.Weekday()] && next.After(now) {
				return next.Format(dateFormat), nil
			}
		}

	case "m":
		if len(parts) < 2 {
			return "", fmt.Errorf("%w: для правила 'm' не указаны дни месяца", ErrInvalidFormat)
		}
		
		dayParts := strings.Split(parts[1], ",")
		allowedDays := make([]int, 0, len(dayParts))
		for _, dStr := range dayParts {
			d, err := strconv.Atoi(dStr)
			if err != nil || d == 0 || d < -2 || d > 31 {
				return "", fmt.Errorf("%w: некорректный день месяца %q", ErrInvalidFormat, dStr)
			}
			allowedDays = append(allowedDays, d)
		}
		
		allowedMonths := make([]bool, 13) // от 1 до 12 для прямой адресации.
		hasMonthFilter := false
		if len(parts) >= 3 {
			hasMonthFilter = true
			monthParts := strings.Split(parts[2], ",")
			for _, mStr := range monthParts {
				m, err := strconv.Atoi(mStr)
				if err != nil || m < 1 || m > 12 {
					return "", fmt.Errorf("%w: некорректный номер месяца %q", ErrInvalidFormat, mStr)
				}
				allowedMonths[m] = true
			}
		}

		next := startDate
		for {
			next = next.AddDate(0, 0, 1)

			// Если стоит фильтр по месяцам, а текущим месяц не подходит - пропускаем шаг.
			if hasMonthFilter && !allowedMonths[int(next.Month())] {
				continue
			}

			// Определяем последний день месяца путем получения нулевого дня следующего месяца.
			lastDayOfMonth := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, next.Location()).Day()
			
			dayMatch := false
			for _, d := range allowedDays {
				if d > 0 && next.Day() == d {
					dayMatch = true
					break
				}
				if d == -1 && next.Day() == lastDayOfMonth {
					dayMatch = true
					break
				}
				if d == -2 && next.Day() == lastDayOfMonth-1 {
					dayMatch = true
					break
				}
			}

			if dayMatch && next.After(now) {
				return next.Format(dateFormat), nil
			}
		}
	default:
		return "", fmt.Errorf("%w: %s", ErrUnknownRule, rule)
	}
}
