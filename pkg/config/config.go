package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
)

type Config struct {
	DBHost, DBPort, DBUser, DBName string
	DBPassword                     string `json:"-"`
}

func Init() (*Config, error) {
	cfg := &Config{
		DBHost:     GetEnv("POSTGRES_DB_HOST", "localhost"),
		DBPort:     GetEnv("POSTGRES_DB_PORT", "5432"),
		DBUser:     GetEnv("POSTGRES_DB_USER", "postgres"),
		DBPassword: GetEnv("POSTGRES_DB_PASSWORD", "postgres"),
		DBName:     GetEnv("POSTGRES_DB_Name", "postgres"),
	}
	return cfg, nil
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

func (c Config) DBConnURI() string {
	cs := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timezone=UTC",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword), c.DBHost, url.QueryEscape(c.DBPort), c.DBName)
	return cs
}
