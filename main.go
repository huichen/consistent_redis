package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/huichen/consistent_service"
	"log"
	"math/rand"
	"strings"
	"time"
)

var (
	endPoints   = flag.String("endpoints", "", "Comma-separated endpoints of your etcd cluster, each starting with http://.")
	serviceName = flag.String("service_name", "", "Name of your service in etcd.")
)

type ConsistentRedisClient struct {
	service   consistent_service.ConsistentService
	conns     map[string]redis.Conn
	duplicate int
}

func (client *ConsistentRedisClient) Init(endpoints []string, servicename string) error {
	client.duplicate = 2
	client.conns = make(map[string]redis.Conn)
	return client.service.Connect(servicename, endpoints)
}

func (client *ConsistentRedisClient) Set(key string, value string) error {
	nodes, _ := client.service.GetNodes(key, client.duplicate)
	if nodes != nil {
		log.Printf("assigned to node: %v", nodes)
	} else {
		log.Printf("no assignment")
	}

	if len(nodes) == 0 {
		return errors.New("Set error: can't get access to any redis node")
	}

	for _, node := range nodes {
		if _, ok := client.conns[node]; !ok {
			conn, err := redis.Dial("tcp", node)
			if err != nil {
				return err
			}
			client.conns[node] = conn
		}
		err := client.conns[node].Send("SET", key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (client *ConsistentRedisClient) Get(key string) (value string, err error) {
	nodes, _ := client.service.GetNodes(key, client.duplicate)
	if nodes != nil {
		log.Printf("assigned to node: %v", nodes)
	} else {
		log.Printf("no assignment")
	}

	if len(nodes) == 0 {
		err = errors.New("Get error: can't get access to any redis node")
		return
	}

	for _, node := range nodes {
		if _, ok := client.conns[node]; !ok {
			client.conns[node], err = redis.Dial("tcp", node)
			if err != nil {
				return
			}
		}
		client.conns[node].Send("GET", key)
		value, err = redis.String(client.conns[node].Receive())
		if err == nil {
			return
		}
	}
	return
}

func (client *ConsistentRedisClient) Close() {
	for _, v := range client.conns {
		v.Close()
	}
}

func main() {
	// Parsing flags
	flag.Parse()
	ep := strings.Split(*endPoints, ",")
	if len(ep) == 0 {
		log.Fatal("Can't parse --endpoints")
	}
	if *serviceName == "" {
		log.Fatal("--service_name can't be empty")
	}

	var client ConsistentRedisClient
	err := client.Init(ep, *serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	//rand.Seed(time.Now().UTC().UnixNano())

	t1 := time.Now()
	for i := 0; i < 10000; i++ {
		n := rand.Intn(10000000)
		key := fmt.Sprintf("%d", n)
		value := fmt.Sprintf("%d", n)

		err := client.Set(key, value)
		if err != nil {
			log.Fatal(err)
		}

		/*
			val, err := client.Get(key)
			if err != nil {
				log.Fatal(err)
			}
		*/
		log.Printf("%d key(%s) = %s", i, key, value)
	}

	t2 := time.Now()
	t := t2.Sub(t1).Seconds()
	log.Printf("%f", float64(t))
}
