package db

import (
	"time"
)

type User struct {
    ID        int
    Name      string
    Email     string
    Password  string 
    CreatedAt time.Time
}

type Keep struct {
    ID          int
    UserID      int
    Title       string
    Description string
    StoragePath string
    CreatedAt   time.Time
}