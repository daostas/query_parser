package qp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type (
	QueryParam  struct{ Key, Type, Sign, Value string }
	QueryParams []QueryParam
)

func NewQueryParam(key, _type, sign, value string) QueryParam {
	return QueryParam{
		Key:   key,
		Type:  _type,
		Sign:  sign,
		Value: value,
	}
}

// String Создает строку для sql запроса из структуры
func (qp *QueryParam) String() string {
	return fmt.Sprintf("%s%s %s %s", qp.Key, qp.Type, qp.Sign, qp.Value)
}

func (qp *QueryParam) Marshal() (data []byte, err error) {
	return json.Marshal(qp)
}

func (qp *QueryParam) Unmarshal(data []byte) (err error) {
	if data != nil {
		if len(data) != 0 {
			return json.Unmarshal(data, &qp)
		}
	}
	return
}

func (qp *QueryParams) Marshal() (data []byte, err error) {
	return json.Marshal(qp)
}

func (qp *QueryParams) Unmarshal(data []byte) (err error) {
	if data != nil {
		if len(data) != 0 {
			return json.Unmarshal(data, &qp)
		}
	}
	return
}

// QueryFormat - Функция, которая формирует структуру по key и value приходящих в query params в запросе в api
func QueryFormat(key, value string) (qp QueryParam, err error) {
	key = strings.Trim(key, " ")
	// По умолчанию ставим знак равно
	sign := "="
	// Доступные на данный момент функции
	r := regexp.MustCompile(`\[(>|<|>-|<-|!|<>|~\*|~|!~\*|!~|\+|!\+|!%|%|!:|:|like|not_like|between|not_between|similar_to|not_similar_to)]$`)
	if r.MatchString(key) {
		matches := r.FindStringSubmatch(key)
		// Достаем знак, который пришел в query параметрах
		sign = matches[1]
		// Убираем найденный знак из ключа
		key = r.ReplaceAllString(key, "")
		// Обработка знака на валидные для sql запроса в базу PostgreSql
		switch sign {
		case ">-":
			sign = ">="
		case "<-":
			sign = "<="
		case "<>", "!":
			sign = "!="
		case "%", "like":
			sign = "like"
		case "!%", "not_like":
			sign = "not like"
		case "+", "similar_to":
			sign = "similar to"
		case "!+", "not_similar_to":
			sign = "not similar to"
		case ":", "between":
			sign = "between"
		case "!:", "not_between":
			sign = "not between"
		case ">", "<", "~", "~*", "!~", "!~*":
			// В этих случаях не меняется
		}
	}

	// Обработка всех значений и запись в QueryParam
	err = qp.FillQueryParam(key, sign, value)
	return
}

// ReplaceForSql - Функция для экранирование ковычек в value и добавления их по обе стороны от value для sql запроса.
// Если value будет равен "null", ковычки добавлены не будут
func ReplaceForSql(value string) string {
	if value != "null" {
		value = fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
	}
	return value
}

// isNumber - Функция, которая определяет, пришло ли число в value.
// Если пришло не число, то вызывает функцию ReplaceForSql для экранирования строки и добавления ковычек
func isNumber(value string) string {
	r := regexp.MustCompile(`^\d*$`)
	if !r.MatchString(value) {
		value = ReplaceForSql(value)
	}
	return value
}

// isArrayOperation - Функция, которая проверяет, пытаются ли вызвать операцию ANY(ARRAY[]) или ALL(ARRAY[]).
// Если да, то возращает true, тип операции в переменной _type и необработнное значение массива для ARRAY[] в переменной _value.
func isArrayOperation(value string) (_type, _value string, ok bool) {
	r := regexp.MustCompile(`^(.+)\(\[(.+)]\)$`)
	if r.MatchString(value) {
		matches := r.FindStringSubmatch(value)
		_type = matches[1]
		if _type == "any" || _type == "all" {
			_value = fmt.Sprintf("[%s]", matches[2])
			ok = true
		}
	}
	return
}

// parseArrayValue - Функция для парсинга и экранирование value для ANY(ARRAY[]) или ALL(ARRAY[]).
// Возращает false, если пришло пустое значение. Если typification = true, сохраняет тип исходного value и экранирует,
// если пришла строка. Если typification = false, все значения переводит в строковый тип и экранирует
func parseArrayValue(value string, typification bool, isLike bool) (string, bool, error) {
	var res []string
	var arr []any

	err := json.Unmarshal([]byte(value), &arr)
	if err != nil {
		return value, false, err
	}

	for _, v := range arr {
		str := fmt.Sprint(v)

		if isLike {
			str = strings.ReplaceAll(str, "%", `\%`)
			str = strings.ReplaceAll(str, "_", `\_`)
			str = "%" + str + "%"
		}

		if typification {
			str = isNumber(str)
		} else {
			str = ReplaceForSql(str)
		}
		res = append(res, str)
	}

	if len(res) != 0 {
		value = "(array["
		for i, v := range res {
			value += v
			if i < len(res)-1 {
				value += ", "
			} else {
				value += "])"
			}
		}
	} else {
		return "", false, nil
	}
	return value, true, nil
}

// FillQueryParam Обрабатывает пришедшие значения и заполняет структура QueryParam
func (qp *QueryParam) FillQueryParam(key, sign, value string) (err error) {
	ok := false
	qp.Sign = sign
	qp.Key = key

	// Проверка на ANY(ARRAY[]) или ALL(ARRAY[])
	_type, _value, isArray := isArrayOperation(value)
	if isArray {
		qp.Sign += " " + _type
	}

	switch sign {
	// Cлучай со знаками, которые поддерживают работу с LIKE
	case "like", "not like", "=", "!=", "~", "~*", "!~", "!~*", "similar to", "not similar to":
		qp.Type = "::text"
		if isArray {
			// Если ANY(ARRAY[]) или ALL(ARRAY[])
			if sign == "similar to" || sign == "not similar to" {
				err = fmt.Errorf(`wrong value format for "%s" operation`, sign)
			}
			qp.Value, ok, err = parseArrayValue(_value, false, sign == "like" || sign == "not like")
			if !ok && err == nil {
				err = fmt.Errorf(`wrong value format for "%s" operation`, _type+"(array[])")
			}
		} else {
			// Смена знаков если пришел null value
			if value == "null" {
				switch sign {
				case "=":
					qp.Sign = "is"
				case "!=":
					qp.Sign = "is not"
				default:
					err = fmt.Errorf(`wrong value "null" for "%s" operation`, sign)
				}
			}
			// Экранизация символов для корректной работы like и not like
			if sign == "like" || sign == "not like" {
				value = strings.ReplaceAll(value, "%", `\%`)
				value = strings.ReplaceAll(value, "_", `\_`)
				value = "%" + value + "%"
			}
			qp.Value = ReplaceForSql(value)
		}
	// Случай between и not between
	case "between", "not between":
		if isArray {
			// Возврат ошибки, потому что between не поддерживает ANY(ARRAY[]) или ALL(ARRAY[])
			err = fmt.Errorf(`wrong value format for "%s" operation`, sign)
		} else {
			r := regexp.MustCompile(`^\[(.+)\{;}(.+)]$`)
			if r.MatchString(value) {
				matches := r.FindStringSubmatch(value)
				if len(matches) == 3 {
					if matches[1] == "null" || matches[2] == "null" {
						err = fmt.Errorf(`wrong value "null" for "%s" operation`, sign)
					}
					qp.Value = fmt.Sprintf("%s and %s", isNumber(matches[1]), isNumber(matches[2]))
				}
			} else {
				err = fmt.Errorf(`wrong value format for "%s" operation`, sign)
			}
		}
	// Все остальные случаи
	default:
		if isArray {
			qp.Value, isArray, err = parseArrayValue(_value, true, false)
			if !isArray {
				err = fmt.Errorf(`wrong value format for "%s" operation`, _type+"(array[])")
			}
		} else {
			qp.Value = isNumber(value)
		}
	}
	return
}
