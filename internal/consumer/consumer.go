//Package consumer applies for communication with redis stream
package consumer

import (
	"context"
	"github.com/EgorBessonov/price-service/internal/model"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"sync"
)

//Consumer struct
type Consumer struct {
	RedisClient *redis.Client
	StreamName  string
	ShareMap    *model.ShareMap
	mutex       *sync.Mutex
}

//NewConsumer returns new consumer instance
func NewConsumer(redisClient *redis.Client, streamName string, shareMap *model.ShareMap, mutex *sync.Mutex) *Consumer {
	return &Consumer{
		RedisClient: redisClient,
		StreamName:  streamName,
		ShareMap:    shareMap,
		mutex:       mutex,
	}
}

//GetPrices method read share prices from redis stream
func (consumer *Consumer) GetPrices(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, err := consumer.RedisClient.XRead(&redis.XReadArgs{
				Streams: []string{consumer.StreamName, "$"},
				Count:   1,
				Block:   0,
			}).Result()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("consumer: can't read message")
			}
			bytes := result[0].Messages[0]
			share := model.Share{}
			for _, sh := range bytes.Values {
				if err = share.UnmarshalBinary([]byte(sh.(string))); err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("consumer: can't parse message")
				}
				consumer.mutex.Lock()
				for _, ch := range consumer.ShareMap.Map[share.Name] {
					*ch <- &share
				}
				consumer.mutex.Unlock()
			}
		}
	}
}
