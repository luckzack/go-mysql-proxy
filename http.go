package main

import (
	"fmt"
	"github.com/gogoods/mysql-proxy/conf"
	"log"
	"net/http"

	"encoding/json"
	"github.com/gogoods/mysql-proxy/chat"
	"github.com/gorilla/websocket"
	"github.com/olekukonko/tablewriter"
)

const (
	websocketRoute = "/ws"
	webRoute       = "/"
)

func runHttpServerOld(hub *chat.Hub) {
	// Websockets endpoint
	http.HandleFunc(websocketRoute, func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// Proper handling 'close' message from the peer
		// See https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages for details
		go func() {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
			}
		}()

		client := chat.NewClient(conn, hub)

		hub.RegisterClient(client)

		go client.Process()
	})

	// Query execution endpoint
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		type Data struct {
			Database   string
			Query      string
			Parameters []string
		}

		var parsedData Data
		data := r.PostFormValue("data")
		json.Unmarshal([]byte(data), &parsedData)

		columns, rows, err := getQueryResults(parsedData.Database, parsedData.Query, parsedData.Parameters, *mysqlDsn)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		if len(columns) > 0 {
			table := tablewriter.NewWriter(w)
			table.SetAutoFormatHeaders(false)
			table.SetColWidth(1000)
			table.SetHeader(columns)
			table.AppendBulk(rows)
			table.Render()
		} else {
			fmt.Fprint(w, "Empty response")
		}
	})

	http.Handle(webRoute, http.FileServer(FS(*useLocalUI)))

	log.Fatal(http.ListenAndServe(*guiAddr, nil))

}

func addProxyRoute(hub *chat.Hub, proxyAlias string) {
	// Websockets endpoint
	http.HandleFunc(websocketRoute+"/"+proxyAlias, func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// Proper handling 'close' message from the peer
		// See https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages for details
		go func() {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
			}
		}()

		client := chat.NewClient(conn, hub)

		hub.RegisterClient(client)

		go client.Process()
	})

	// Query execution endpoint
	http.HandleFunc("/execute"+"/"+proxyAlias, func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		type Data struct {
			Database   string
			Query      string
			Parameters []string
		}

		var parsedData Data
		data := r.PostFormValue("data")
		json.Unmarshal([]byte(data), &parsedData)

		columns, rows, err := getQueryResults(parsedData.Database, parsedData.Query, parsedData.Parameters, *mysqlDsn)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		if len(columns) > 0 {
			table := tablewriter.NewWriter(w)
			table.SetAutoFormatHeaders(false)
			table.SetColWidth(1000)
			table.SetHeader(columns)
			table.AppendBulk(rows)
			table.Render()
		} else {
			fmt.Fprint(w, "Empty response")
		}
	})

	//sync.Once(func() {
	//	http.Handle(webRoute, http.FileServer(FS(_useLocalUI)))
	//
	//	fmt.Printf("Web gui available at `http://%s/` \n", addr)
	//	log.Fatal(http.ListenAndServe(addr, nil))
	//
	//})
}

func runServer(proxies []*conf.ProxyConfig, addr string, _useLocalUI bool) {
	http.HandleFunc("/apply", func(w http.ResponseWriter, r *http.Request) {
		bs, _ := json.Marshal(proxies)
		w.Write(bs)
	})

	http.Handle(webRoute, http.FileServer(FS(_useLocalUI)))

	fmt.Printf("Web gui available at `http://%s/` \n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
