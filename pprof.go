package main

import (
	"github.com/golang/glog"
	"net/http"
	"net/http/pprof"
)

// InitPprof start http pprof.
func InitPprof() {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	for _, addr := range MyConf.PprofBind {
		go func() {
			if err := http.ListenAndServe(addr, pprofServeMux); err != nil {
				glog.Errorf("http.ListenAndServe(\"%s\", pproServeMux) error(%v)", addr)
				panic(err)
			}
		}()
	}
}
