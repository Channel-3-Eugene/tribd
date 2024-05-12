package uriHandler

import (
	"log"
	"net/http"
)

type HTTPHandler struct {
	server   *http.Server
	dataChan chan []byte
	mode     Mode
	role     Role
}

func NewHTTPHandler(addr string, dataChan chan []byte) *HTTPHandler {
	handler := &HTTPHandler{
		dataChan: dataChan,
		server:   &http.Server{Addr: addr},
		mode:     Server,
		role:     Writer,
	}
	// Setup HTTP server with a handler method
	http.HandleFunc("/", handler.streamHandler)
	return handler
}

func (h *HTTPHandler) streamHandler(w http.ResponseWriter, r *http.Request) {
	// Set necessary headers for streaming MPEG-TS content
	w.Header().Set("Content-Type", "video/MP2T")

	// Ensure the writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Listen to the closing of the http connection via the CloseNotifier
	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Cannot stream", http.StatusInternalServerError)
		return
	}

	// This is where we stream the data
	for {
		select {
		case data, ok := <-h.dataChan:
			if !ok {
				return // if the channel is closed, stop the handler
			}
			_, err := w.Write(data)
			if err != nil {
				log.Printf("Error writing data to client: %v", err)
				return
			}
			flusher.Flush() // Trigger flush to send data immediately

		case <-closeNotifier.CloseNotify():
			log.Println("Client has disconnected")
			return
		}
	}
}

func (h *HTTPHandler) Open() error {
	log.Printf("Starting HTTP server at %s", h.server.Addr)
	if err := h.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (h *HTTPHandler) Close() error {
	log.Println("Shutting down HTTP server")
	return h.server.Close()
}
