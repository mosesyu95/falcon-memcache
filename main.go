package main

import (
	log "github.com/Sirupsen/logrus"
	goconf "github.com/akrennmair/goconf"
	"github.com/open-falcon/falcon-plus/common/model"
	"flag"
	"os"
	"time"
)

type Cfg struct {
	LogFile      string
	LogLevel     int
	FalconClient string
	Endpoint     string
	Port            string
}

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
	for {
		time.Sleep(60 * time.Second)
		data := Memcache()
		mvs := []*model.MetricValue{}
		for _, mv := range data {
			mvs = append(mvs, mv)
		}

		now := time.Now().Unix()
		for j := 0; j < len(mvs); j++ {
			mvs[j].Step = 60
			mvs[j].Endpoint = Hostname()
			mvs[j].Timestamp = now
		}
		sendData(mvs)
		msg, err := sendData(data)
		if err != nil {

		}
		log.Infof("Send response %s", string(msg))
	}
}
