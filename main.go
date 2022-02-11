package main

import (
	"context"
	"fmt"
	"github.com/EgorBessonov/price-service/internal/config"
	"github.com/EgorBessonov/price-service/internal/consumer"
	"github.com/EgorBessonov/price-service/internal/model"
	"github.com/EgorBessonov/price-service/internal/service"
	priceService "github.com/EgorBessonov/price-service/protocol"
	"github.com/caarlos0/env"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("price service: can't parse config")
	}
	redisClient, err := newRedisClient(&cfg)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("price service: can't create redis client")
	}
	shareMap := model.ShareMap{
		Map: map[int32]map[string]*chan *model.Share{
			1: {},
			2: {},
			3: {},
			4: {},
			5: {},
		},
	}
	mutex := sync.Mutex{}
	cons := consumer.NewConsumer(redisClient, cfg.RedisStreamName, &shareMap, &mutex)
	defer func() {
		if err := cons.RedisClient.Close(); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("price service: error while closing redis client")
		}
	}()
	s := service.NewService(&shareMap, &mutex)
	go newgRPCConnection(cfg.PriceServicePort, s)
	exitContext, cancelFunc := context.WithCancel(context.Background())
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go cons.GetPrices(exitContext)
	<-exitChan
	cancelFunc()
}

func newRedisClient(cfg *config.Config) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: "",
		DB:       0,
	})
	if _, err := redisClient.Ping().Result(); err != nil {
		return nil, fmt.Errorf("consumer: can't create new instance - %e", err)
	}
	return redisClient, nil
}

func newgRPCConnection(port string, s *service.Service) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("price service: can't create gRPC server")
	}
	gServer := grpc.NewServer()
	priceService.RegisterPriceServer(gServer, s)
	log.Printf("price service: listening at %s", lis.Addr())
	if err = gServer.Serve(lis); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("price service: gRPC server failed")
	}
}
