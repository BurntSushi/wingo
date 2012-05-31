package config

import (
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wini"
)

func SetString(k wini.Key, place *string) {
	if v, ok := GetLastString(k); ok {
		*place = v
	}
}

func GetLastString(k wini.Key) (string, bool) {
	vals := k.Strings()
	if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return "", false
	}

	return vals[len(vals)-1], true
}

func SetBool(k wini.Key, place *bool) {
	if v, ok := GetLastBool(k); ok {
		*place = v
	}
}

func GetLastBool(k wini.Key) (bool, bool) {
	vals, err := k.Bools()
	if err != nil {
		logger.Warning.Println(err)
		return false, false
	} else if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return false, false
	}

	return vals[len(vals)-1], true
}

func SetInt(k wini.Key, place *int) {
	if v, ok := GetLastInt(k); ok {
		*place = int(v)
	}
}

func GetLastInt(k wini.Key) (int, bool) {
	vals, err := k.Ints()
	if err != nil {
		logger.Warning.Println(err)
		return 0, false
	} else if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return 0, false
	}

	return vals[len(vals)-1], true
}

func SetFloat(k wini.Key, place *float64) {
	if v, ok := GetLastFloat(k); ok {
		*place = float64(v)
	}
}

func GetLastFloat(k wini.Key) (float64, bool) {
	vals, err := k.Floats()
	if err != nil {
		logger.Warning.Println(err)
		return 0.0, false
	} else if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return 0.0, false
	}

	return vals[len(vals)-1], true
}
