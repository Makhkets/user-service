package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	fmt.Print(PasswordToHash("123321asSs", "ds@@akd12kd$xlk2@31kksnakn13)_-do10idj3-j8qeh813;mcxl'k=\\13i"))
}

func PasswordToHash(password, secretKey string) string {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	hashedPassword := hex.EncodeToString(h.Sum(nil))
	return hashedPassword
}
