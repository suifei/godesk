package auth

import (
    "crypto/sha256"
    "encoding/hex"
)

type User struct {
    Username string
    PasswordHash string
}

func NewUser(username, password string) *User {
    return &User{
        Username: username,
        PasswordHash: hashPassword(password),
    }
}

func (u *User) Authenticate(password string) bool {
    return u.PasswordHash == hashPassword(password)
}

func hashPassword(password string) string {
    hash := sha256.Sum256([]byte(password))
    return hex.EncodeToString(hash[:])
}