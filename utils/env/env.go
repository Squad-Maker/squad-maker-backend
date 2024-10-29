package env

import (
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"
)

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func GetStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}

func GetSecretKey(key string) ([]byte, error) {
	s, err := GetStr(key)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(s, "base64:") {
		s = s[7:]
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return []byte(s), nil
}

func GetInt(key string) (int, error) {
	s, err := GetStr(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func GetInt32(key string) (int32, error) {
	i, err := GetInt(key)
	return int32(i), err
}

func GetUInt8(key string) (uint8, error) {
	i, err := GetInt(key)
	return uint8(i), err
}

func GetUInt32(key string) (uint32, error) {
	i, err := GetInt(key)
	return uint32(i), err
}

func GetBool(key string) (bool, error) {
	s, err := GetStr(key)
	if err != nil {
		return false, err
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return v, nil
}
