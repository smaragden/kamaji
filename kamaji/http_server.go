package kamaji

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

type ClientList struct {
	sync.RWMutex
	Clients []string
}

var ClientListInst ClientList

func updateClients(cm *ClientManager) {
	for {
		var clients []string
		index := 0
		for p, v := range cm.itemToPool {
			client := p.(*Client)
			clientStr := fmt.Sprintf("<li>Client[%d]: %s, %s</li>", index, client.ID.String(), v.Name)
			clients = append(clients, clientStr)
			index++
		}
		ClientListInst.Lock()
		ClientListInst.Clients = clients
		ClientListInst.Unlock()
		time.Sleep(time.Second)
	}

}

func HttpServe(cm *ClientManager) {
	go updateClients(cm)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	var resp string
	ClientListInst.Lock()
	fmt.Println(ClientListInst.Clients)
	for _, c := range ClientListInst.Clients {
		resp = fmt.Sprintf("%s\n%s", resp, c)
	}
	ClientListInst.Unlock()
	resp = fmt.Sprintf("<html><body><ul>%s</ul></body></html>", resp)
	fmt.Println(resp)
	fmt.Fprintf(w, resp)
}
