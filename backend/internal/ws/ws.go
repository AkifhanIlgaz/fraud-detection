// Package ws, WebSocket bağlantılarını yöneten Hub'ı sağlar.
// Fiber'ın fasthttp pipeline'ıyla çakışmaması için standart net/http
// üzerinde ayrı bir sunucu olarak çalışır.
package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Geliştirme ortamında tüm origin'lere izin verilir.
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Hub, bağlı WebSocket istemcilerini yönetir.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	log.Printf("[ws] istemci bağlandı — toplam: %d", len(h.clients))
}

func (h *Hub) unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	log.Printf("[ws] istemci ayrıldı — toplam: %d", len(h.clients))
}

// Broadcast, data'yı bağlı tüm istemcilere gönderir.
// Yazma hatası olan bağlantıyı kapatır ve hub'dan çıkarır.
func (h *Hub) Broadcast(data []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[ws] broadcast yazma hatası: %v", err)
			h.unregister(c)
			c.Close()
		}
	}
}

// ListenAndServe, addr üzerinde bağımsız bir net/http WebSocket sunucusu başlatır.
// /alerts endpoint'ini açar ve bağlanan her istemciyi hub'a kaydeder.
func (h *Hub) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[ws] upgrade hatası: %v", err)
			return
		}
		h.register(conn)
		defer func() {
			h.unregister(conn)
			conn.Close()
		}()
		// Read pump: istemci ayrılınca ReadMessage hata döner.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})

	log.Printf("[ws] sunucu başladı — ws://%s/transactions", addr)
	return http.ListenAndServe(addr, mux)
}
