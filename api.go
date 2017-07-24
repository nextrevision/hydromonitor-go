package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type API struct {
	Router    *mux.Router
	datastore *Datastore
	state     *State
}

func NewAPI(datastore *Datastore, state *State) *API {
	a := &API{
		Router:    mux.NewRouter().StrictSlash(true),
		datastore: datastore,
		state:     state,
	}

	v1Router := mux.NewRouter()
	v1 := v1Router.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/devices", a.DevicesHandler).Methods("GET", "OPTIONS", "HEAD")
	v1.HandleFunc("/devices/{id}", a.DeviceHandler).Methods("GET", "OPTIONS")
	v1.HandleFunc("/devices/{id}", a.DeviceUpdateHandler).Methods("POST")
	v1.HandleFunc("/devices/{id}", a.DeviceDeleteHandler).Methods("DELETE")
	v1.HandleFunc("/devices/{id}/metrics", a.DeviceMetricsHandler).Methods("GET")
	v1.HandleFunc("/devices/{id}/latest", a.DeviceLatestMetricsHandler).Methods("GET")
	v1.HandleFunc("/devices/{id}/refresh", a.DeviceRefreshHandler).Methods("POST")

	a.Router.PathPrefix("/api/v1").Handler(v1Router)
	a.Router.PathPrefix("/").Handler(http.FileServer(http.Dir("./www")))
	return a
}

func (a *API) Start() {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	chain := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.LoggingHandler(os.Stderr, h)
		},
	).Then(a.Router)

	log.Info("[api] Listening on :8000")
	http.ListenAndServe(":8000", handlers.CORS(headersOk, originsOk, methodsOk)(chain))
}

func (a *API) DevicesHandler(w http.ResponseWriter, r *http.Request) {
	devices, err := a.datastore.GetDevicesWithMetrics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	respondJSON(w, devices)
}

func (a *API) DeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	device, err := a.datastore.GetDevice(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	device.LatestMetric, err = a.datastore.GetDeviceLatestMetrics(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	respondJSON(w, device)
}

func (a *API) DeviceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := a.datastore.DeleteDevice(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) DeviceUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	device, err := a.datastore.GetDevice(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Error(err)
		return
	}

	err = json.Unmarshal(body, &device)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, err.Error())
		return
	}

	if err := a.datastore.UpdateDevice(device); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) DeviceMetricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	limit := r.URL.Query().Get("limit")

	if limit == "" {
		limit = "24"
	}

	metrics, err := a.datastore.GetDeviceMetrics(id, limit)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	respondJSON(w, metrics)
}

func (a *API) DeviceLatestMetricsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	metrics, err := a.datastore.GetDeviceLatestMetrics(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	respondJSON(w, metrics)
}

func (a *API) DeviceRefreshHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := a.state.RefreshTilt(id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	metric, err := a.datastore.GetDeviceLatestMetrics(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	respondJSON(w, metric)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(payload)
}
