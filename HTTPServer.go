package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

//HTTPServer for handling web request
type HTTPServer struct {
	httpPort        int
	blockChain      *Blockchain
	isNewBlockAdded *bool
}

//DataMessage constains data for new block
type DataMessage struct {
	Message string
}

func (s *HTTPServer) startServer() error {
	mux := s.makeMuxRouter()

	log.Println("\nHTTP listening on ", s.httpPort)
	ser := &http.Server{
		Addr:           ":" + strconv.Itoa(s.httpPort),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := ser.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *HTTPServer) makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/bc", s.handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/mine", s.handleWriteBlock).Methods("POST")
	return muxRouter
}

func (s *HTTPServer) handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(s.blockChain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

func (s *HTTPServer) handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var msg DataMessage

	//decode and save data as DataMessage object
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&msg); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, *s.blockChain)
		return
	}
	defer r.Body.Close()

	//add data to the block
	s.blockChain.addBlock([]byte(msg.Message))
	*(s.isNewBlockAdded) = true
	respondWithJSON(w, r, http.StatusCreated, *s.blockChain)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload Blockchain) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
