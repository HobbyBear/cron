package cron

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"
)

var client = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0, // use default DB
})

type memoryStorage struct {
	s map[string]time.Time
	r map[string]struct{}
}

func (m *memoryStorage) DelRetryEntry(taskId string) {
	client.HDel("retry", taskId)
}

func (m *memoryStorage) GetRetryEntryList() []string {
	list := make([]string, 0)
	res, _ := client.HGetAll("retry").Result()
	for taskId, _ := range res {
		list = append(list, taskId)
	}
	return list
}

func (m *memoryStorage) PutRetryEntry(taskId string) {
	client.HSet("retry", taskId, "")
}

func (m *memoryStorage) GetEntry(taskId string) *Entry {
	nextTime, err := client.HGet("mis", taskId).Result()
	if err != nil {
		return nil
	}
	t, _ := strconv.Atoi(nextTime)
	return &Entry{TaskId: taskId, Next: time.Unix(int64(t), 0)}
}

func (m *memoryStorage) PutEntry(taskId string, nextTime time.Time) {
	client.HSet("mis", taskId, nextTime.Unix())
}

func TestMisfire(t *testing.T) {
	c := newWithSeconds()
	const (
		taskId = "taskId"
	)
	storage := &memoryStorage{s: map[string]time.Time{}, r: map[string]struct{}{}}

	fmt.Println(os.Getpid())

	i := 0

	c.storage = storage

	var process = func(ext ExtContext) {
		if ext.IsInterrupt {
			panic("hahahah")
		}
		log.Println("start===============", i, ext)
		time.Sleep(5 * time.Second)
		log.Println("end==============", i, ext)
		i++
	}

	process2 := NewChain(Recover(PrintfLogger(log.New(os.Stdout, "", 0)))).
		Then(JobFunc(process))

	_, err := c.AddTask("0 0/1 * * * ?", taskId, process2, 30)
	if err != nil {
		t.Fatal(err)
	}
	c.Start()
	// 模拟定时任务执行
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	<-sigterm
	c.Stop()

	//// 模拟重启服务
	//c2 := newWithSeconds()
	//c2.storage = storage
	//_, err = c2.AddTask("0 0/1 * * * ?", taskId, process, 30)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//c2.run()
}
