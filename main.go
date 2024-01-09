package main

import (
	"flag"
	"fmt"
	"github.com/gogoods/mysql-proxy/conf"
	"time"

	"github.com/gogoods/mysql-proxy/chat"
)

var (
	proxyAddr  = flag.String("proxy", "127.0.0.1:4041", "Proxy <host>:<port>")
	mysqlAddr  = flag.String("mysql", "127.0.0.1:3306", "MySQL <host>:<port>")
	guiAddr    = flag.String("gui", "127.0.0.1:9999", "Web UI <host>:<port>")
	useLocalUI = flag.Bool("use-local", false, "Use local UI instead of embed")
	mysqlDsn   = flag.String("mysql-dsn", "", "MySQL DSN for query execution capabilities")

	config = flag.String("config-file", "./conf/config.yaml", "Specify a config file")
)

func appReadyInfo(appReadyChan chan bool) {
	<-appReadyChan
	time.Sleep(1 * time.Second)
	fmt.Printf("Forwarding queries from `%s` to `%s` \n", *proxyAddr, *mysqlAddr)
	fmt.Printf("Web gui available at `http://%s` \n", *guiAddr)
}

func main() {
	runNew()
}

func runOld() {

	flag.Parse()

	cmdChan := make(chan chat.Cmd)
	cmdResultChan := make(chan chat.CmdResult)
	connStateChan := make(chan chat.ConnState)
	appReadyChan := make(chan bool)

	hub := chat.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	go runHttpServerOld(hub)
	go appReadyInfo(appReadyChan)

	p := MySQLProxyServer{cmdChan, cmdResultChan, connStateChan, appReadyChan, *mysqlAddr, *proxyAddr}
	p.run()
}

func runNew() {
	flag.Parse()

	conf.ParseConfig(*config)

	for _, proxy := range conf.Config().Proxies {
		if !proxy.Enabled {
			continue
		}
		go runProxy(proxy)
	}
	runServer(conf.Config().Proxies, conf.Config().GUI, !conf.Config().UseEmbedUI)
	select {}
}

func runProxy(conf *conf.ProxyConfig) {
	cmdChan := make(chan chat.Cmd)
	cmdResultChan := make(chan chat.CmdResult)
	connStateChan := make(chan chat.ConnState)
	appReadyChan := make(chan bool)

	hub := chat.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	go addProxyRoute(hub, conf.Alias)

	go func() {
		<-appReadyChan
		time.Sleep(1 * time.Second)
		fmt.Printf("[%s]Forwarding queries from `%s` to `%s` \n", conf.Alias, conf.Mysql, conf.Listen)

	}()

	p := MySQLProxyServer{cmdChan, cmdResultChan, connStateChan, appReadyChan, conf.Mysql, conf.Listen}
	p.run()
}
