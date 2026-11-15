package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)


func main() {
	originServers:=make([]*url.URL,5)
	for i:=0;i<5;i++{
		parsedURL,err:=url.Parse(fmt.Sprintf("http://127.0.0.1:800%d",i))
		if err!=nil{
			log.Fatalf("Couldn't parse the given address: %v",err)
		}
		originServers[i]=parsedURL
	}
}