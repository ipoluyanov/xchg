package http_server

import (
	"github.com/ipoluianov/xchg/internal/logger"
	"time"
)

func (c *HttpServer) purgeRoutine() {
	ticker := time.NewTicker(c.config.PurgeInterval)
	for {
		select {
		case <-c.stopPurgeRoutineCh:
			return
		case <-ticker.C:
		}

		c.mtx.Lock()
		logger.Println("purging begin")
		for id, l := range c.listeners {
			if time.Now().Sub(l.LastGetDT()) > 10*time.Second {
				logger.Println("purging", id)
				delete(c.listeners, id)
			}
		}
		logger.Println("purging end")
		c.mtx.Unlock()
	}
}
