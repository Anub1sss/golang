package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Request представляет тело запроса
type Request struct {
	Expression string `json:"expression"`
}

// Response представляет тело ответа
type Response struct {
	Result *float64 `json:"result,omitempty"`
	Error  *string  `json:"error,omitempty"`
}

// Calculate вычисляет значение арифметического выражения
func Calculate(expression string) (float64, error) {
	// Убираем пробелы
	expression = strings.ReplaceAll(expression, " ", "")

	// Проверяем на недопустимые символы
	if !isValidExpression(expression) {
		return 0, errors.New("Expression is not valid")
	}

	// Вычисляем выражение
	result, err := evaluate(expression)
	if err != nil {
		return 0, errors.New("Expression is not valid")
	}

	return result, nil
}

// Проверяет выражение на допустимость символов
func isValidExpression(expression string) bool {
	for _, char := range expression {
		if !(char >= '0' && char <= '9' || char == '+' || char == '-' || char == '*' || char == '/' || char == '.') {
			return false
		}
	}
	return true
}

// Выполняет простое вычисление выражения (поддерживает +, -, *, /)
func evaluate(expression string) (float64, error) {
	// Разбиваем на числа и операции
	var numbers []float64
	var operators []rune
	var currentNum strings.Builder

	for _, char := range expression {
		if char >= '0' && char <= '9' || char == '.' {
			currentNum.WriteRune(char)
		} else if char == '+' || char == '-' || char == '*' || char == '/' {
			// Завершаем текущее число
			num, err := strconv.ParseFloat(currentNum.String(), 64)
			if err != nil {
				return 0, err
			}
			numbers = append(numbers, num)
			currentNum.Reset()

			// Добавляем оператор
			operators = append(operators, char)
		} else {
			return 0, errors.New("Invalid character in expression")
		}
	}

	// Добавляем последнее число
	if currentNum.Len() > 0 {
		num, err := strconv.ParseFloat(currentNum.String(), 64)
		if err != nil {
			return 0, err
		}
		numbers = append(numbers, num)
	}

	// Выполняем операции (* и / сначала)
	for i := 0; i < len(operators); {
		if operators[i] == '*' || operators[i] == '/' {
			// Выполняем операцию
			var result float64
			if operators[i] == '*' {
				result = numbers[i] * numbers[i+1]
			} else {
				if numbers[i+1] == 0 {
					return 0, errors.New("Division by zero")
				}
				result = numbers[i] / numbers[i+1]
			}

			// Обновляем массивы
			numbers[i] = result
			numbers = append(numbers[:i+1], numbers[i+2:]...)
			operators = append(operators[:i], operators[i+1:]...)
		} else {
			i++
		}
	}

	// Выполняем операции (+ и -)
	for len(operators) > 0 {
		var result float64
		if operators[0] == '+' {
			result = numbers[0] + numbers[1]
		} else {
			result = numbers[0] - numbers[1]
		}

		// Обновляем массивы
		numbers[0] = result
		numbers = append(numbers[:1], numbers[2:]...)
		operators = operators[1:]
	}

	if len(numbers) == 1 {
		return numbers[0], nil
	}
	return 0, errors.New("Evaluation error")
}

// CalculateHandler обрабатывает HTTP-запросы
func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request format"}`, http.StatusBadRequest)
		return
	}

	result, err := Calculate(req.Expression)
	if err != nil {
		errorMessage := err.Error()
		http.Error(w, `{"error":"`+errorMessage+`"}`, http.StatusUnprocessableEntity)
		return
	}

	resp := Response{Result: &result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/api/v1/calculate", CalculateHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
