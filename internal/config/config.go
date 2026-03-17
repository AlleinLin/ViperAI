package config

import (
	"sync"

	"github.com/BurntSushi/toml"
)

type AppConfig struct {
	Name string `toml:"name"`
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	Charset  string `toml:"charset"`
}

type CacheConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type QueueConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	VHost    string `toml:"vhost"`
}

type AuthConfig struct {
	Secret   string `toml:"secret"`
	Issuer   string `toml:"issuer"`
	Subject  string `toml:"subject"`
	Duration int    `toml:"duration"`
}

type MailConfig struct {
	Address    string `toml:"address"`
	AuthCode   string `toml:"auth_code"`
	SMTPServer string `toml:"smtp_server"`
	SMTPPort   int    `toml:"smtp_port"`
}

type AIModelConfig struct {
	Provider       string `toml:"provider"`
	EmbeddingModel string `toml:"embedding_model"`
	ChatModel      string `toml:"chat_model"`
	BaseURL        string `toml:"base_url"`
	Dimension      int    `toml:"dimension"`
	DocDirectory   string `toml:"doc_directory"`
}

type VoiceConfig struct {
	APIKey    string `toml:"api_key"`
	SecretKey string `toml:"secret_key"`
}

type Configuration struct {
	App      AppConfig      `toml:"app"`
	Database DatabaseConfig `toml:"database"`
	Cache    CacheConfig    `toml:"cache"`
	Queue    QueueConfig    `toml:"queue"`
	Auth     AuthConfig     `toml:"auth"`
	Mail     MailConfig     `toml:"mail"`
	AIModel  AIModelConfig  `toml:"ai_model"`
	Voice    VoiceConfig    `toml:"voice"`
}

type CacheKeyConfig struct {
	CaptchaPrefix   string
	IndexName       string
	IndexNamePrefix string
}

var DefaultCacheKeyConfig = CacheKeyConfig{
	CaptchaPrefix:   "captcha:%s",
	IndexName:       "knowledge:%s:idx",
	IndexNamePrefix: "knowledge:%s:",
}

var (
	instance *Configuration
	once     sync.Once
)

func Load() *Configuration {
	once.Do(func() {
		instance = &Configuration{}
		if _, err := toml.DecodeFile("config/settings.toml", instance); err != nil {
			instance = getDefaultConfig()
		}
	})
	return instance
}

func getDefaultConfig() *Configuration {
	return &Configuration{
		App: AppConfig{
			Name: "ViperAI",
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "",
			Database: "viperai",
			Charset:  "utf8mb4",
		},
		Cache: CacheConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Queue: QueueConfig{
			Host:     "localhost",
			Port:     5672,
			User:     "guest",
			Password: "guest",
			VHost:    "/",
		},
		Auth: AuthConfig{
			Secret:   "viperai-default-secret-key",
			Issuer:   "viperai",
			Subject:  "auth",
			Duration: 24,
		},
		Mail: MailConfig{
			Address:    "",
			AuthCode:   "",
			SMTPServer: "smtp.example.com",
			SMTPPort:   587,
		},
		AIModel: AIModelConfig{
			Provider:       "openai",
			EmbeddingModel: "text-embedding-ada-002",
			ChatModel:      "gpt-3.5-turbo",
			BaseURL:        "https://api.openai.com/v1",
			Dimension:      1536,
			DocDirectory:   "uploads",
		},
		Voice: VoiceConfig{
			APIKey:    "",
			SecretKey: "",
		},
	}
}

func Get() *Configuration {
	if instance == nil {
		return Load()
	}
	return instance
}
