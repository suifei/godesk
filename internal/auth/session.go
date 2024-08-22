package auth

import (
    "time"
)

type Session struct {
    Token string
    UserID string
    ExpiresAt time.Time
}

func NewSession(userID string, duration time.Duration) *Session {
    return &Session{
        Token: generateToken(),
        UserID: userID,
        ExpiresAt: time.Now().Add(duration),
    }
}

func (s *Session) IsValid() bool {
    return time.Now().Before(s.ExpiresAt)
}

func generateToken() string {
    // 实现一个安全的令牌生成方法
    // 这里仅作为示例
    return "example-token"
}