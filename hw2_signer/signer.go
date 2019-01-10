package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)
	wg := &sync.WaitGroup{}

	for _, job := range jobs {
		wg.Add(1)
		go startWorker(job, in, out, wg)
		in = out
		out = make(chan interface{}, 100)
	}
	wg.Wait()
}

func startWorker(job job, in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	job(in, out)
	close(out)
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	quota := make(chan int, 1)
	for i := range in {
		wg.Add(1)
		go calcSingleHash(strconv.Itoa(i.(int)), out, wg, quota)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go calcMultiHash(i.(string), out, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var result []string
	for i := range in {
		result = append(result, i.(string))
	}
	sort.Strings(result)
	out <- strings.Join(result, "_")
}

func calcSingleHash(data string, out chan interface{}, wg *sync.WaitGroup, quota chan int) {
	defer wg.Done()
	crc32Data := make(chan string)
	crc32Md5 := make(chan string)
	md5 := make(chan string)

	go calcCrc32(data, crc32Data)
	go calcMd5(data, md5, quota)
	go calcCrc32(<-md5, crc32Md5)

	out <- (<-crc32Data + "~" + <-crc32Md5)
}

func calcCrc32(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func calcMd5(data string, out chan string, quota chan int) {
	quota <- 0
	out <- DataSignerMd5(data)
	<-quota
}

func calcMultiHash(data string, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var chans = []chan string{
		make(chan string, 1),
		make(chan string, 1),
		make(chan string, 1),
		make(chan string, 1),
		make(chan string, 1),
		make(chan string, 1),
	}

	for i := 0; i < 6; i++ {
		go calcCrc32(strconv.Itoa(i)+data, chans[i])
	}
	out <- (<-chans[0] + <-chans[1] + <-chans[2] + <-chans[3] + <-chans[4] + <-chans[5])
}
