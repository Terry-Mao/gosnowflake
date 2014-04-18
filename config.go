package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"github.com/golang/glog"
	"runtime"
)

var (
	// global config object
	goConf   = goconf.New()
	MyConf   *Config
	confPath string
)

type Config struct {
	User    string `goconf:"base:user"`
	PidFile string `goconf:"base:pid_file"`
	Dir     string `goconf:"base:dir"`
	MaxProc int    `goconf:"base:max_proc"`
}

func init() {
	flag.StringVar(&confPath, "conf", "./user_account.conf", " set user_account config file path")
}

// Init init the configuration file.
func Init() error {
	MyConf = &Config{
		User:    "nobody",
		PidFile: "/tmp/user_account.pid",
		Dir:     "/dev/null",
		MaxProc: runtime.NumCPU(),
	}
	if err := goConf.Parse(confPath); err != nil {
		glog.Errorf("goconf.Parse(\"%s\") error(%v)", confPath, err)
		return err
	}
	if err := goConf.Unmarshal(MyConf); err != nil {
		glog.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}
	return nil
}

// Reload reload the configuration file.
func Reload() error {
	glog.Infof("config file: \"%s\" reload", confPath)
	t, err := goConf.Reload()
	if err != nil {
		glog.Errorf("goconf.Reload() error(%v)", err)
		glog.Warningf("confi file: \"%s\" reload failed, use original one", confPath)
		return err
	}
	goConf = t
	// new a Config for Unmarshal
	myConf := &Config{}
	if err := goConf.Unmarshal(myConf); err != nil {
		glog.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}
	// atomic replace MyConf
	MyConf = myConf
	return nil
}
