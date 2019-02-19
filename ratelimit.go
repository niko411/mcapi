package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/OneOfOne/cmap/stringcmap"
)

const (
	rateLimitThreshold = 3
	banThreshold       = 200
)

var rateLimit *stringcmap.CMap

func init() {
	rateLimit = stringcmap.New()

	go processRateLimit()
}

type CloudFlareConfiguration struct {
	Target string `json:"target"`
	Value  string `json:"value"`
}

type CloudFlareData struct {
	Mode          string                  `json:"mode"`
	Configuration CloudFlareConfiguration `json:"configuration"`
	Notes         string                  `json:"notes"`
}

var blocked []string
var blockSync sync.RWMutex

func blockFromCloudFlare(ip string) {
	blockSync.RLock()
	for _, against := range blocked {
		if ip == against {
			blockSync.RUnlock()
			return
		}
	}
	blockSync.RUnlock()

	log.Printf("Blocking IP %s at CloudFlare level ...\n", ip)

	blockSync.Lock()
	blocked = append(blocked, ip)
	blockSync.Unlock()

	j, err := json.Marshal(CloudFlareData{
		Mode: "block",
		Configuration: CloudFlareConfiguration{
			Target: "ip",
			Value:  ip,
		},
		Notes: "blocked by count threshold",
	})

	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/user/firewall/access_rules/rules", bytes.NewReader(j))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", os.Getenv("CLOUDFLARE_EMAIL"))
	req.Header.Set("X-Auth-Key", os.Getenv("CLOUDFLARE_AUTH"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	resp.Body.Close()

	log.Printf("Blocked IP %s at CloudFlare level\n", ip)
}

func processRateLimit() {
	for range time.Tick(time.Second) {
		rateLimit.ForEach(func(ip string, val interface{}) bool {
			i, ok := val.(int)

			if !ok {
				return true
			}

			i -= rateLimitThreshold

			if i <= 0 {
				rateLimit.Delete(ip)
			} else {
				rateLimit.Set(ip, i)
			}

			return true
		})
	}
}

func shouldRateLimit(ip string) (bool, int) {
	item := rateLimit.Get(ip)

	if item == nil {
		return false, -1
	}

	if i, ok := item.(int); ok {
		if i > rateLimitThreshold {
			incrRateLimit(ip)
			if i > 200 {
				blockFromCloudFlare(ip)
			}
			return true, i
		}
	}

	return false, -1
}

func incrRateLimit(ip string) {
	item := rateLimit.Get(ip)

	if item == nil {
		rateLimit.Set(ip, 1)
	} else if i, ok := item.(int); ok {
		rateLimit.Set(ip, i+1)
	}
}
