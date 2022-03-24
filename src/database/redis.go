package database

import (
	"github.com/go-redis/redis/v8"
	"context"
	"time"
	"fmt"
)

var CacheChannel chan string // we are specifing it is a channel of string type
var Cache *redis.Client

func SetupRedis() {
	Cache = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB: 0,
	})
}

// CHANNELS
// un chanel funciona del siguiente modo:
// hay dos tipos que se envían una clave, que en este caso es un string:
// - Sender: recibe la key, en este caso es ClearCache. Para recibir la key utilizamos el operador "<-"

func SetupCacheChannel() {
	CacheChannel = make(chan string)

	// receiving channel THIS IS ALWAYS WAITING
	go func(ch chan string) {
		// with this for we keep the function running all the time (bucle infinito)
		for {
			time.Sleep(5 * time.Second)
			key := <-ch
			Cache.Del(context.Background(), key)
			fmt.Println("Cache cleared", key ) // we are printing the value of the channel
		}
	}(CacheChannel)
}

// THIS UPDATES THE CHANNEL (event handler)
func ClearCache(keys ...string) { // we are saying could be multiple string with the three dots (...)
	for _, key := range keys {
		CacheChannel <- key
	}
}