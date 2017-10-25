package main


import (
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/mosesyu95/gomemcache/memcache"
	"log"
	"strings"
	"strconv"
	"os"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"net/http"
)


func Memcache()[]*model.MetricValue {
	log.Print("Memcache Monitor for falcon")
	ret:=[]*model.MetricValue{}
	for _,port := range strings.Split(cfg.Port,";"){
		_,err := strconv.Atoi(port)
		if err != nil{
			log.Fatalf("Port configure is illegal: %v", err)
			continue
		}
		FetchData(port,&ret)
	}
	return ret
}

func Hostname()string{
	hostname, _ := os.Hostname()
	return hostname
}

func FetchData(port string,data *([]*model.MetricValue))*([]*model.MetricValue){
	mc := memcache.New(Hostname() + ":" + port)
	rw_test:=RW_test(mc, port)
	*data = append(*data,rw_test)
	if rw_test.Value == 0 {
		goto A
	}
	data = statsinfo(mc,data, port)
A:
	return data
}

func RW_test(mc *memcache.Client,port string)*model.MetricValue{
	key:= "falcon_key"
	value:="falcon memcache testing"
	err := mc.Set(&memcache.Item{Key:key, Value:[]byte(value)})
	if err != nil {
		log.Print(err)
		return GaugeValue("memcache_read_write_test",0,"port="+port)
	}
	_,err = mc.Get(key)
	if err!= nil {
		log.Print(err)
		return GaugeValue("memcache_read_write_test",0,"port="+port)
	}
	return GaugeValue("memcache_read_write_test",1,"port="+port)
}

func statsinfo(mc *memcache.Client,data *([]*model.MetricValue),port string)*([]*model.MetricValue){
	m,err :=getinfo(mc)
	tag := "port="+port
	if err != nil {
		log.Print(err)
		return data
	}
	*data=append(*data,GaugeValue("memcache_curr_connections",m["curr_connections"],tag))
	*data=append(*data,GaugeValue("memcache_threads",m["threads"],tag))
	*data=append(*data,GaugeValue("memcache_curr_items",m["curr_items"],tag))
	*data=append(*data,GaugeValue("memcache_connections_free_percent",100*(m["connection_structures"]-m["curr_connections"])/m["connection_structures"],tag))
	*data=append(*data,GaugeValue("memcache_get_hits_percent",100*m["get_hits"]/(m["get_hits"]+m["get_misses"]),tag))
	*data=append(*data,GaugeValue("memcache_mem_free_percent",100*(m["limit_maxbytes"]-m["bytes"])/m["limit_maxbytes"],tag))
	*data=append(*data,CounterValue("memcache_cmd_get_speed",m["cmd_get"],tag))
	*data=append(*data,CounterValue("memcache_cmd_set_speed",m["cmd_set"],tag))
	*data=append(*data,CounterValue("memcache_cmd_flush_speed",m["cmd_flush"],tag))
	*data=append(*data,CounterValue("memcache_bytes_read_speed",m["bytes_read"],tag))
	*data=append(*data,CounterValue("memcache_bytes_written_speed",m["bytes_written"],tag))
	return data
}

func getinfo(mc *memcache.Client)(data map[string]int,err error){
	data = make(map[string]int)
	str,err :=mc.Get_stats()
	if err != nil{
		return data,err
	}
	STAT := strings.Split(str,"\n")
	for _,v := range STAT {
		stat := strings.Split(strings.TrimSpace(v)," ")
		if len(stat) !=3 {
			continue
		}
		a,_ :=strconv.Atoi(stat[2])
		data[stat[1]]=a
	}
	return data,nil

}
func sendData(data []*model.MetricValue) ([]byte, error) {

	js, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	log.Print("Send to %s, size: %d", cfg.FalconClient, len(data))
	for _, m := range data {
		log.Print("%s", m)
	}

	res, err := http.Post(cfg.FalconClient, "Content-Type: application/json", bytes.NewBuffer(js))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}


