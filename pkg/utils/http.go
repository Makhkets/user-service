package utils

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func GetFingerprint(headers http.Header) string {
	var headerList []string
	for key, values := range headers {
		if shouldIncludeHeader(key) { // Проверяем, нужен ли нам этот заголовок
			headerList = append(headerList, key+": "+strings.Join(values, ","))
		}
	}
	sort.Strings(headerList)
	sortedHeaders := strings.Join(headerList, ",")
	hash := sha256.Sum256([]byte(sortedHeaders))
	return fmt.Sprintf("%x", hash)
}

func shouldIncludeHeader(key string) bool {
	switch key {
	case "User-Agent", "Accept-Encoding", "Accept-Language":
		return true
	default:
		return false
	}
}

//func CheckRequestJson(c *gin.Context) error {
//	var data map[string]interface{}
//	err := c.BindJSON(&data)
//	if err != nil {
//		return err
//	}
//
//	// Проверяем на пустоту
//	if len(data) == 0 {
//		return fmt.Errorf("JSON is empty")
//	}
//
//	return nil
//}
