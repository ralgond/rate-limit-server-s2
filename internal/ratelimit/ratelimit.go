package ratelimit

import (
	"context"
	"github.com/ralgond/rate-limit-server-s2/internal/config"
	"github.com/redis/go-redis/v9"
	"net/url"
)

type RateLimit struct {
	redisCellClient    *redis.Client
	redisSessionClient *redis.Client
	conf               *config.Config
}

func NewRateLimit(redisCellClient *redis.Client,
	redisSessionClient *redis.Client,
	conf *config.Config) *RateLimit {
	return &RateLimit{
		redisCellClient:    redisCellClient,
		redisSessionClient: redisSessionClient,
		conf:               conf,
	}
}

func (rl *RateLimit) ShouldBlockWithIP(Method string, URL *url.URL, ctx context.Context, xRealIp string) (bool, bool, error) {
	ret := true
	var err error = nil
	matched := false
	for _, url := range rl.conf.RateLimitWithIP.RateLimitURL {
		if url.Method != Method {
			continue
		}
		// fmt.Printf("r=%p\n", url.Pattern)
		_, ok := url.Pattern.Match(URL)
		if ok {
			matched = true
			key := "ip_" + xRealIp
			value1, err1 := rl.redisCellClient.Do(ctx, "CL.THROTTLE", key,
				url.LimitParam[0],
				url.LimitParam[1],
				url.LimitParam[2]).Int64Slice()
			if err1 == nil {
				// fmt.Printf("ShouldBlockWithIP: %v\n", value1)
				if value1[0] == 1 {
					ret = true
				} else {
					ret = false
				}
			} else {
				err = err1
			}

			break
		}
	}

	return ret, matched, err
}

func (rl *RateLimit) ShouldBlockWithSession(Method string, URL *url.URL, ctx context.Context, sessionId string) (bool, bool, error) {
	ret := true
	var err error = nil
	matched := false
	for _, url := range rl.conf.RateLimitWithToken.RateLimitURL {
		if url.Method != Method {
			continue
		}
		_, ok := url.Pattern.Match(URL)
		if ok {
			matched = true
			key := "si_" + sessionId
			value1, err1 := rl.redisCellClient.Do(ctx, "CL.THROTTLE", key,
				url.LimitParam[0],
				url.LimitParam[1],
				url.LimitParam[2]).Int64Slice()
			if err1 == nil {
				// fmt.Printf("ShouldBlockWithSession: %v\n", value1)
				if value1[0] == 1 {
					ret = true
				} else {
					ret = false
				}
			} else {
				err = err1
			}

			break
		}
	}

	return ret, matched, err
}
