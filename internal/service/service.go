//Package service represents price service logic
package service

import (
	"github.com/EgorBessonov/price-service/internal/model"
	priceService2 "github.com/EgorBessonov/price-service/protocol"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"sync"
)

//Service struct
type Service struct {
	ShareMap *model.ShareMap
	mutex    *sync.Mutex
	priceService2.UnimplementedPriceServer
}

//NewService returns new service instance
func NewService(shareMap *model.ShareMap, mutex *sync.Mutex) *Service {
	return &Service{
		ShareMap: shareMap,
		mutex:    mutex,
	}
}

//Get method get prices from redis stream and update price cache
func (s *Service) Get(request *priceService2.GetRequest, stream priceService2.Price_GetServer) error {
	newUser := uuid.New().String()
	shareChan := make(chan *model.Share)
	s.mutex.Lock()
	for _, shChan := range request.Name {
		s.ShareMap.Map[shChan][newUser] = &shareChan
	}
	s.mutex.Unlock()
	for {
		select {
		case <-stream.Context().Done():
			for _, shChan := range request.Name {
				delete(s.ShareMap.Map[shChan], newUser)
			}
			return nil
		default:
			updatedShare := <-shareChan
			err := stream.Send(&priceService2.GetResponse{Share: &priceService2.Share{
				Name: updatedShare.Name,
				Bid:  updatedShare.Bid,
				Ask:  updatedShare.Ask,
				Time: updatedShare.UpdatedAt,
			}})
			if err != nil {
				for _, shChan := range request.Name {
					delete(s.ShareMap.Map[shChan], newUser)
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("service: can't send message to gRPC stream")
			}
		}
	}
}
