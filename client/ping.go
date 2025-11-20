package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	lib "github.com/vszholobov/tetrisLib"
)

type PingMeasurer struct {
	pingMeasures *ttlcache.Cache[uuid.UUID, time.Time]
	actualPing   lib.Ping
}

func MakePingMeasurer() *PingMeasurer {
	cache := ttlcache.New[uuid.UUID, time.Time](
		ttlcache.WithTTL[uuid.UUID, time.Time](time.Minute),
	)
	return &PingMeasurer{
		pingMeasures: cache,
		actualPing:   0,
	}
}

func (pingMeasurer *PingMeasurer) addMeasure() uuid.UUID {
	pingUuid := uuid.New()
	pingMeasurer.pingMeasures.Set(pingUuid, time.Now(), ttlcache.DefaultTTL)
	return pingUuid
}

func (pingMeasurer *PingMeasurer) getMeasure(uuid uuid.UUID) (time.Time, bool) {
	startTime := pingMeasurer.pingMeasures.Get(uuid)
	if startTime != nil {
		return startTime.Value(), true
	} else {
		return time.Time{}, false
	}
}

func (pingMeasurer *PingMeasurer) pongHandler() func(appData string) error {
	return func(appData string) error {
		pingUuid, _ := uuid.FromBytes([]byte(appData))
		if startTime, exists := pingMeasurer.getMeasure(pingUuid); exists {
			pingMeasurer.actualPing = lib.Ping(time.Since(startTime).Milliseconds())
		}
		return nil
	}
}
