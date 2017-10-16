package main

import (
	log "github.com/Sirupsen/logrus"
	goconf "github.com/akrennmair/goconf"
	"github.com/bradfitz/mosesyu95/memcache"
	"flag"
	"os"
	"strings"
	"strconv"
	"time"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
)

type Cfg struct {
	LogFile      string
	LogLevel     int
	FalconClient string
	Endpoint     string
	Port            string
}

var old_m map[string]int
var cfg Cfg


func init() {
	var cfgFile string
	flag.StringVar(&cfgFile, "c", "mc.cfg", "myMon configure file")
	flag.Parse()

	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			log.WithField("cfg", cfgFile).Fatalf("myMon config file does not exists: %v", err)
		}
	}

	if err := cfg.readConf(cfgFile); err != nil {
		log.Fatalf("Read configure file failed: %v", err)
	}

	// Init log file
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.Level(cfg.LogLevel))

	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			log.SetOutput(f)
			return
		}
	}
	log.SetOutput(os.Stderr)
}


func (conf *Cfg) readConf(file string) error {
	c, err := goconf.ReadConfigFile(file)
	if err != nil {
		return err
	}

	conf.LogFile, err = c.GetString("default", "log_file")
	if err != nil {
		return err
	}

	conf.LogLevel, err = c.GetInt("default", "log_level")
	if err != nil {
		return err
	}

	conf.FalconClient, err = c.GetString("default", "falcon_client")
	if err != nil {
		return err
	}

	conf.Endpoint, err = c.GetString("default", "endpoint")
	if err != nil {
		return err
	}
	conf.Port, err = c.GetString("memcache", "port")
	return err
}

func main() {
	log.Info("Memcache Monitor for falcon")
	//go timeout()
	for _,port := range strings.Split(cfg.Port,";"){
		_,err := strconv.Atoi(port)
		if err != nil{
			log.Fatalf("Port configure is illegal: %v", err)
			continue
		}
		go FetchData(port)
	}
	select {}

}


func FetchData(port string){
	mc := memcache.New("127.0.0.1:" + port)
	for {
		data := make([]*MetaData, 0)
		data = append(data, RW_test(mc, data, port))
		data = statsinfo(mc, data, port)

		for _, v := range data {
			log.Info(v)
		}


		msg, err := sendData(data)
		if err != nil {

		}
		log.Infof("Send response %s", string(msg))
		time.Sleep(60*time.Second)
	}
}

func sendData(data []*MetaData) ([]byte, error) {

	js, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	log.Debugf("Send to %s, size: %d", cfg.FalconClient, len(data))
	for _, m := range data {
		log.Debugf("%s", m)
	}

	res, err := http.Post(cfg.FalconClient, "Content-Type: application/json", bytes.NewBuffer(js))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}


func statsinfo(mc *memcache.Client,data []*MetaData,port string)([]*MetaData){
	m,err :=getinfo(mc)
	if err != nil {
		log.Error(err)
		return data
	}
	if len(old_m)==0 {
		log.Error(len(old_m))
		old_m = m
		log.Error(len(old_m))
		return data

	}
	data=append(data,NewMetric("curr_connections",port,m["curr_connections"]))
	data=append(data,NewMetric("threads",port,m["threads"]))
	data=append(data,NewMetric("curr_items",port,m["curr_items"]))
	data=append(data,NewMetric("connections_free_percent",port,100*(m["connection_structures"]-m["curr_connections"])/m["connection_structures"]))
	data=append(data,NewMetric("get_hits_percent",port,100*m["get_hits"]/(m["get_hits"]+m["get_misses"])))
	data=append(data,NewMetric("mem_free_percent",port,100*m["bytes"]/m["limit_maxbytes"]))

	cmd_get_per_minute:=m["cmd_get"]-old_m["cmd_get"]
	if cmd_get_per_minute >= 0 {
		data=append(data,NewMetric("cmd_get_per_minute",port,cmd_get_per_minute))
	}
	cmd_set_per_minute:=m["cmd_set"]-old_m["cmd_set"]
	if cmd_set_per_minute >=0 {
		data=append(data,NewMetric("cmd_set_per_minute",port,cmd_set_per_minute))
	}
	cmd_flush_per_minute:=m["cmd_flush"]-old_m["cmd_flush"]
	if cmd_flush_per_minute >= 0 {
		data = append(data, NewMetric("cmd_flush_per_minute", port, cmd_flush_per_minute))
	}
	bytes_read_per_minute:=m["bytes_read"]-old_m["bytes_read"]
	if bytes_read_per_minute >=0 {
		data = append(data, NewMetric("bytes_read_per_minute", port, bytes_read_per_minute))
	}
	bytes_written_per_minute:=m["bytes_written"]-old_m["bytes_written"]
	if bytes_written_per_minute >=0 {
		data = append(data, NewMetric("bytes_written_per_minute", port, bytes_written_per_minute))
	}
	old_m = m
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

func RW_test(mc *memcache.Client,data []*MetaData,port string)*MetaData{
	key:= "falcon_key"
	value:="falcon memcache testing"
	err := mc.Set(&memcache.Item{Key:key, Value:[]byte(value)})
	if err != nil {
		log.Info(err)
		return NewMetric("write_read_test",port,0)
	}
	_,err = mc.Get(key)
	if err!= nil {
		log.Error(err)
		return NewMetric("write_read_test",port,0)
	}
	return NewMetric("write_read_test",port,1)
}