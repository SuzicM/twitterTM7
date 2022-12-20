package data

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	users = "users/%s"
	all   = "users"
)

func generateKey() (string, string) {
	id := uuid.New().String()
	return fmt.Sprintf(users, id), id
}

func constructKey(id string) string {
	return fmt.Sprintf(users, id)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func IsAlnumOrHyphen(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func isBlacklisted(password string) bool {
	file, err := os.Open("blacklist/blacklist-passwords.txt")

	if err != nil {
		log.Fatalf("failed to open")

	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if scanner.Text() == password {
			return true
		}
	}
	file.Close()
	return false
}

func ValidatePassword(s string) bool {
	if isBlacklisted(s) {
		return false
	}
	pass := 0
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			pass++
		case unicode.IsUpper(c):
			pass++
		case unicode.IsPunct(c) && c != ';' && c != '-' && c != '=':
			pass++
		case unicode.IsLower(c):
			pass++
		case unicode.IsLetter(c) || c == ' ':
			pass++
		default:
			return false
		}
	}
	return pass == len(s)
}

func ValidateName(user *User, logger *log.Logger) bool {
	reg, _ := regexp.Compile("^[a-zA-Z]+$")
	match := reg.MatchString(user.Name)

	if !match {
		logger.Println("Error: ValidateName")
	}

	return match
}

func ValidateLastName(user *User, logger *log.Logger) bool {
	reg, _ := regexp.Compile("^[a-zA-Z]+$")
	match := reg.MatchString(user.Surname)

	if !match {
		logger.Println("Error: ValidateLastName")
	}

	return match
}

func ValidateGender(user *User, logger *log.Logger) bool {
	reg, _ := regexp.Compile("^[a-zA-Z]+$")
	match := reg.MatchString(user.Gender)

	if !match {
		logger.Println("Error: ValidateGender")
	}

	return match
}

func ValidateResidance(user *User, logger *log.Logger) bool {
	reg, _ := regexp.Compile("^[a-z]+([a-zA-Z0-9]+)$")
	match := reg.MatchString(user.Gender)

	if !match {
		logger.Println("Error: ValidateResidance")
	}

	return match
}

func ValidateAge(user *User, logger *log.Logger) bool {
	reg, _ := regexp.Compile("^[0-9]+$")
	match := reg.MatchString(user.Age)

	if !match {
		logger.Println("Error: Age is invalid")
	}
	return match
}

func ValidateUsername(user *User, logger *log.Logger) bool {
	var valid = IsAlnumOrHyphen(user.Username)

	if !valid {
		logger.Println("Error: ValidateUsername")
	}

	return valid
}
