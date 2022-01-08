package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	// Postgres
	DBHost, DBPort, DBUser, DBName string
	DBPassword                     string `json:"-"`

	// Redis
	RedisHost, RedisPort string
	RedisPassword        string `json:"-"`
	RedisDB              int
}

func Init() *Config {
	cfg := &Config{
		DBHost:        GetEnv("POSTGRES_DB_HOST", "localhost"),
		DBPort:        GetEnv("POSTGRES_DB_PORT", "5432"),
		DBUser:        GetEnv("POSTGRES_DB_USER", "postgres"),
		DBPassword:    GetEnv("POSTGRES_DB_PASSWORD", "postgres"),
		DBName:        GetEnv("POSTGRES_DB_NAME", "postgres"),
		RedisHost:     GetEnv("REDIS_HOST", "localhost"),
		RedisPort:     GetEnv("REDIS_PORT", "6379"),
		RedisDB:       mustParseInt(GetEnv("REDIS_DB", "0")),
		RedisPassword: GetEnv("REDIS_PASSWORD", ""),
	}
	return cfg
}

func (c Config) Dump(w io.Writer) error {
	fmt.Fprint(w, "config: ")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(c)
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func mustParseInt(x string) (y int) {
	y, err := strconv.Atoi(x)
	if err != nil {
		panic("cant parse interger")
	}
	return
}

func (c Config) DBConnURI() string {
	cs := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timezone=UTC",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword), c.DBHost, url.QueryEscape(c.DBPort), c.DBName)
	return cs
}
