package main

import (
	"time"
)

import (
	"log"
	"runtime"
)

type GoroutinePool struct {
	Queue  chan func() error
	Number int
	Total  int

	result         chan error
	finishCallback func()
}

func (self *GoroutinePool) Init(number int, total int) {
	self.Queue = make(chan func() error, total)
	self.Number = number
	self.Total = total
	self.result = make(chan error, total)
}

func (self *GoroutinePool) routine(i int) {
	defer func() {
		log.Printf("goroutine %d done...", i)
		if err := recover(); err != nil {
			stack := make([]byte, 1024)
			stack = stack[:runtime.Stack(stack, false)]

			f := "PANIC: %s\n%s"
			log.Printf(f, err, stack)
		}
	}()

	for {
		task, ok := <-self.Queue
		if !ok {
			break
		}

		retry := 3

		var err error
		for retry > 0 {
			err = task()
			if err == nil {
				break
			}
			retry--
			log.Printf("goroutine %d retrying...%d chance left, err:%s", i, retry, err)
			time.Sleep(2 * time.Second)
		}

		self.result <- err
	}
}
func (self *GoroutinePool) Start() {
	for i := 0; i < self.Number; i++ {
		var i = i
		go self.routine(i)
	}

	for j := 0; j < self.Total; j++ {
		res, ok := <-self.result
		if !ok {
			break
		}

		if res != nil {
			log.Println(res)
		}
	}

	if self.finishCallback != nil {
		self.finishCallback()
	}
}

func (self *GoroutinePool) Stop() {
	close(self.Queue)
	close(self.result)
}

func (self *GoroutinePool) AddTask(task func() error) {
	self.Queue <- task
}

func (self *GoroutinePool) SetFinishCallback(callback func()) {
	self.finishCallback = callback
}
