package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

func FormatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func PasswordToHash(password, secretKey string) string {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	hashedPassword := hex.EncodeToString(h.Sum(nil))
	return hashedPassword
}

func HasNil(slice ...interface{}) bool {
	for _, v := range slice {
		for _, j := range v.([]interface{}) {
			if j == nil {
				return true
			}
		}
	}
	return false
}

func CheckEmptyFields(s interface{}) []string {
	detectedFields := []string{}
	v := reflect.ValueOf(s)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		if field.Interface() != reflect.Zero(fieldType.Type).Interface() {
			detectedFields = append(detectedFields, fieldType.Name)
		}
	}
	return detectedFields
}

func ContainsStringInArray(substr string, arr []string) bool {
	for _, field := range arr {
		if strings.ToLower(substr) == strings.ToLower(field) {
			return true
		}
	}
	return false
}

func GetIdField(id any) string {
	return fmt.Sprintf("user%v", id)
}
