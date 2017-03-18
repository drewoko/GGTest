package main

import (
	"github.com/drewoko/comfyconf"
	"log"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"regexp"
	"sync"
	"time"
)

var CHAN string

func main() {

	conf := comfyconf.New(comfyconf.NewFlags())

	url := conf.String("url", "url", "", "")
	cons := conf.Int("cons", "cons", 100, "")
	timeout := conf.Int("timeout", "timeout", 0, "")

	conf.Parse()

	fmt.Println(*url)
	fmt.Println(*cons)
	fmt.Println(*timeout)

	if *url == "" {
		log.Fatalln("channel url is empty")
	}

	_, body, _ := gorequest.New().Get(*url).End()

	CHAN = regexp.MustCompile(`var src = "(.+)"`).FindAllStringSubmatch(body, -1)[0][1]

	fmt.Println(CHAN)

	var wg sync.WaitGroup

	for i := 0; i< *cons; i++ {
		go initGoodGame(&wg)
		time.Sleep(time.Duration(*timeout))
	}

	wg.Wait()
}
