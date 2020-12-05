package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joekir/algoexplore"
	_ "github.com/joekir/algoexplore/internal/algos/ctph"
)

const (
	algoStateGob               = "ALGO_GOB"
	sessionCookieName          = "SESSION"
	portEnvVarName             = "PORT"
	cookieSessionKeyEnvVarName = "COOKIE_SESSION_KEY"
)

var (
	listeningPort  = os.Getenv(portEnvVarName)
	availableAlgos []string
	cookieStore    *sessions.CookieStore
)

func main() {
	if len(listeningPort) < 1 {
		listeningPort = "8080"
	}

	router := mux.NewRouter()
	router.HandleFunc("/{algo}/init", Init).Methods("POST")
	router.HandleFunc("/{algo}/step", StepAlgo).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../../web/")))
	cookieStore = sessions.NewCookieStore([]byte(os.Getenv(cookieSessionKeyEnvVarName)))

	log.Printf("Listening on %s\n", listeningPort)
	log.Fatal(http.ListenAndServe(":"+listeningPort, router))
}

type hashReq struct {
	DataLength int `json:"data_length"`
}

func validateAlgo(vars map[string]string) *algoexplore.AlgoInfo {
	algoName := vars["algo"]

	algoInfo, err := algoexplore.GetAlgo(algoName)
	log.Printf("algoInfo: %#v\n", algoInfo)
	if err != nil {
		log.Fatal("valid algorithm not found")
		return nil
	}

	return &algoInfo
}

func Init(w http.ResponseWriter, r *http.Request) {
	algoInfo := validateAlgo(mux.Vars(r))

	var h hashReq
	if err := algoexplore.StrictUnmarshalJSON(&(r.Body), &h); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.DataLength <= 0 {
		http.Error(w, "Invalid 'data_length'", http.StatusUnprocessableEntity)
		return
	}

	// A session is always returned
	session, _ := cookieStore.Get(r, sessionCookieName)

	if !session.IsNew {
		log.Println("Deleting old cookie")
		session.Options.MaxAge = -1
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var err error
		session, err = cookieStore.New(r, sessionCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	algo := algoInfo.New()
	algo.Init(algo, h.DataLength)
	state := algo.SerializeState()
	log.Printf("state: %#v\n", state)

	session.Values[algoStateGob] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(state)
}

type stepReq struct {
	Data byte `json:"byte"`
}

func StepAlgo(w http.ResponseWriter, r *http.Request) {
	algoInfo := validateAlgo(mux.Vars(r))
	session, err := cookieStore.Get(r, sessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew {
		http.Error(w, "no session detected", http.StatusPreconditionRequired)
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var s stepReq
	if err := decoder.Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.Data == 0x0 {
		// You could argue that 0x0 is a legitimate state, however in ascii it is NUL
		// Hence it's unlikely to be a legit input, however this is a default input if the
		// Client doesn't have a valid one, so we should return
		log.Printf("no data provided")
		http.Error(w, "No data provided, no state to update", http.StatusNoContent)
		return
	}

	algo := algoInfo.New()
	sessionCookie := session.Values[algoStateGob]
	algo.Step(sessionCookie.(string), s.Data)

	session.Values[algoStateGob] = sessionCookie
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessionCookie)
}
