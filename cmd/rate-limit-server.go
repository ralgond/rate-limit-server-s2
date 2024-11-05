package main

import (
	"context"
	"fmt"
	"github.com/ralgond/rate-limit-server-s2/internal/config"
	"github.com/ralgond/rate-limit-server-s2/internal/ratelimit"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"net/url"
	"os"
	"time"
)

// 创建一个 Transport，并设置连接池参数

var (
	rlRdbS2 *redis.Client
	usRdbS2 *redis.Client
	rlS2    *ratelimit.RateLimit
	ctxS2   context.Context
)

func handle(ctx *fasthttp.RequestCtx) {
	xRealIP := string(ctx.Request.Header.Peek("X-Real-IP"))
	if xRealIP == "" {
		xRealIP = ctx.RemoteAddr().String()
		ip, _, err := net.SplitHostPort(xRealIP)
		if err == nil {
			xRealIP = ip
		}
	}

	xRealMethod := string(ctx.Request.Header.Peek("X-Real-Method"))

	log.Printf("body_size:%d\n", len(ctx.Request.Body()))

	URL, err := url.Parse(string(ctx.Request.URI().FullURI()))
	if err != nil {
		ctx.Error("StatusInternalServerError", fasthttp.StatusInternalServerError)
		return
	}

	shouldBeLimited := false
	sessionId := string(ctx.Request.Header.Cookie("sessionId"))
	if sessionId == "" {
		log.Printf("====>ShouldBlockWithIP: addr=%v, method=%s, URL=%s",
			ctx.RemoteAddr().String(), xRealMethod, URL.String())
		_shouldBeLimited, _matched, err1 := rlS2.ShouldBlockWithIP(xRealMethod, URL, ctx, xRealIP)
		if err1 != nil {
			ctx.Error("StatusInternalServerError", fasthttp.StatusInternalServerError)
			return
		}
		if !_matched {
			log.Println("here 2.")
			ctx.Error("StatusForbidden", fasthttp.StatusForbidden)
			return
		}
		shouldBeLimited = _shouldBeLimited
	} else {
		value, err := usRdbS2.Get(ctx, sessionId).Result()

		// fmt.Printf("err: ====>%v", err)
		if err == nil && value == "" {
			ctx.Error("StatusForbidden", fasthttp.StatusForbidden)
			return
		}

		if err != nil {
			ctx.Error("StatusInternalServerError", fasthttp.StatusInternalServerError)
			return
		}

		_shouldBeLimited, _matched, err1 := rlS2.ShouldBlockWithSession(string(ctx.Request.Header.Method()), URL, ctx, sessionId)
		if err1 != nil {
			ctx.Error("StatusInternalServerError", fasthttp.StatusInternalServerError)
			return
		}
		if !_matched {
			ctx.Error("StatusForbidden", fasthttp.StatusForbidden)
			return
		}
		shouldBeLimited = _shouldBeLimited
	}

	if shouldBeLimited {
		ctx.Error("Requests are being made too frequently.", fasthttp.StatusTooManyRequests)
	} else {
		ctx.Error("OK", fasthttp.StatusOK)
	}
}

func runCoreS2() {
	conf, err := config.LoadConfig("./configs/rate-limit-server.xml")
	if err != nil {
		fmt.Println("load configuration failed, err: ", err)
		os.Exit(-1)
	}

	logFile, err := os.OpenFile(conf.Log.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	os.Stdin.Close()
	os.Stdout.Close()
	os.Stderr.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)
	log.Printf("To start http\n")

	m := func(ctx *fasthttp.RequestCtx) {
		handle(ctx)
	}

	serverS2 := &fasthttp.Server{
		Handler:      m,
		ReadTimeout:  time.Duration(conf.Timeout.FrontendReadTimeoutMS.Value) * time.Millisecond,
		WriteTimeout: time.Duration(conf.Timeout.FrontendWriteTimeoutMS.Value) * time.Millisecond,
		Concurrency:  conf.Frontend.MaxConcurrency,
	}

	rlRdbS2 = redis.NewClient(&redis.Options{
		Addr:     conf.RateLimitRedis.Address,
		Password: "",
		DB:       0,
		PoolSize: conf.RateLimitRedis.PoolSize,
	})

	usRdbS2 = redis.NewClient(&redis.Options{
		Addr:     conf.UserSessionRedis.Address,
		Password: "",
		DB:       0,
		PoolSize: conf.UserSessionRedis.PoolSize,
	})

	rlS2 = ratelimit.NewRateLimit(rlRdbS2, usRdbS2, conf)

	err = serverS2.ListenAndServe(conf.Frontend.Address)
	if err != nil {
		log.Fatalf("ListenAndServe failed, err=%v", err)
	}
}

func main() {
	runCoreS2()
}
