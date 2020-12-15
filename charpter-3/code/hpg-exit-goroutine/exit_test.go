package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func do(taskCh chan int) {
	for {
		select {
		case t := <-taskCh:
			time.Sleep(time.Millisecond)
			fmt.Printf("task %d is done\n", t)
		}
	}
}

func sendTasks() {
	taskCh := make(chan int, 10)
	go do(taskCh)
	for i := 0; i < 1000; i++ {
		taskCh <- i
	}
}

func TestDo(t *testing.T) {
	t.Log(runtime.NumGoroutine())
	sendTasks()
	time.Sleep(time.Second)
	t.Log(runtime.NumGoroutine())
}

func doCheckClose(taskCh chan int) {
	for {
		select {
		case t, beforeClosed := <-taskCh:
			if !beforeClosed {
				fmt.Println("taskCh has been closed")
				return
			}
			time.Sleep(time.Millisecond)
			fmt.Printf("task %d is done\n", t)
		}
	}
}

func sendTasksCheckClose() {
	taskCh := make(chan int, 10)
	go doCheckClose(taskCh)
	for i := 0; i < 1000; i++ {
		taskCh <- i
	}
	close(taskCh)
}

func TestDoCheckClose(t *testing.T) {
	t.Log(runtime.NumGoroutine())
	sendTasksCheckClose()
	time.Sleep(time.Second)
	runtime.GC()
	t.Log(runtime.NumGoroutine())
}
