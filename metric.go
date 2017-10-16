package main

import (
	"fmt"
	"os"
	"time"
)


type MetaData struct {
	Metric      string      `json:"metric"`      //key
	Endpoint    string      `json:"endpoint"`    //hostname
	Value       interface{} `json:"value"`       // number or string
	CounterType string      `json:"counterType"` // GAUGE  原值   COUNTER 差值(ps)
	Tags        string      `json:"tags"`        // port=3306,k=v
	Timestamp   int64       `json:"timestamp"`
	Step        int64       `json:"step"`
}

func (m *MetaData) String() string {
	s := fmt.Sprintf("MetaData Metric:%s Endpoint:%s Value:%v CounterType:%s Tags:%s Timestamp:%d Step:%d",
		m.Metric, m.Endpoint, m.Value, m.CounterType, m.Tags, m.Timestamp, m.Step)
	return s
}

func NewMetric(name string,port string,value int) *MetaData {
	return &MetaData{
		Metric:      name,
		Endpoint:    hostname(),
		Value:		 value,
		CounterType: "GAUGE",
		Tags:        fmt.Sprintf("port=%s", port),
		Timestamp:   time.Now().Unix(),
		Step:        60,
	}
}
func hostname() string {
	host, err := os.Hostname()
	if err != nil {

	}
	return host
}


//time	1320756409	服务器当前Unix时间戳
//curr_connections	41	当前连接数量
//connection_structures	46	Memcached分配的连接结构数量
//cmd_get	164377	get命令请求次数
//cmd_set	58617	set命令请求次数
//cmd_flush	0	flush命令请求次数
//get_hits	105598	get命令命中次数
//get_misses	58779	get命令未命中次数
//bytes_read	262113283	读取总字节数
//bytes_written	460023263	发送总字节数
//limit_maxbytes	536870912	分配的内存总大小（字节）
//threads	4	当前线程数
//bytes	1941693	当前存储占用的字节数
//curr_items	476	当前存储的数据总数
