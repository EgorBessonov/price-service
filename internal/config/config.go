//Package config represents config info of price service
package config

// Config type store all env info of price service
type Config struct {
	RedisURL         string `env:"REDISURL" envDefault:"127.0.0.1:49153"`
	RedisStreamName  string `env:"RSTREAMNAME" envDefault:"PriceGenerator"`
	PriceServicePort string `env:"PRICESERVICEPORT" envDefault:"localhost:8083"`
}
