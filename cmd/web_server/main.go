package main

import (
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	ctph "github.com/joekir/algoexplore/internal/algos/ctph"
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

func init() {
	// Server side storage
	cookieStore = sessions.NewCookieStore([]byte(os.Getenv(cookieSessionKeyEnvVarName)))
	gob.RegisterName(algoStateGob, &ctph.FuzzyHash{})
	gob.Register(&ctph.RollingHash{})
	if len(listeningPort) < 1 {
		listeningPort = "8080"
	}

	algos, err := ioutil.ReadDir("../../internal/algos/")
	if err != nil {
		log.Fatal(err)
	}
	for _, algo := range algos {
		availableAlgos = append(availableAlgos, algo.Name())
	}
	// log.Printf("availableAlgos: %#v\n", availableAlgos)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{algo}/init", Init).Methods("POST")
	router.HandleFunc("/{algo}/step", StepAlgo).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../../web/")))

	log.Printf("Listening on %s\n", listeningPort)
	log.Fatal(http.ListenAndServe(":"+listeningPort, router))
}

type hashReq struct {
	DataLength int `json:"data_length"`
}

func validateAlgo(vars map[string]string) string {
	algo := vars["algo"]
	for _, a := range availableAlgos {
		if a == algo {
			return a
		}
	}

	log.Fatal("valid algorithm not found")
	return ""
}

func Init(w http.ResponseWriter, r *http.Request) {
	algo := validateAlgo(mux.Vars(r))
	decoder := json.NewDecoder(r.Body)

	var h hashReq
	err := decoder.Decode(&h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Init (algo=%s) request received: %#v\n", algo, h)

	if h.DataLength <= 0 {
		http.Error(w, "Invalid 'data_length'", http.StatusUnprocessableEntity)
		return
	}

	// A session is always returned
	session, _ := cookieStore.Get(r, sessionCookieName)

	if !session.IsNew {
		log.Println("Deleting old cookie")
		session.Options.MaxAge = -1
		session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session, err = cookieStore.New(r, sessionCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fh := ctph.NewFuzzyHash(h.DataLength)
	session.Values[algoStateGob] = fh
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fh)
}

type stepReq struct {
	Data byte `json:"byte"`
}

func StepAlgo(w http.ResponseWriter, r *http.Request) {
	algo := validateAlgo(mux.Vars(r))
	session, err := cookieStore.Get(r, sessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// This should be acting on a server-side struct persisted with gob
	// Hence a session is needed from /Init call
	if session.IsNew {
		http.Error(w, "no session detected", http.StatusPreconditionRequired)
		return
	}

	fhCookie := session.Values[algoStateGob]
	fh, ok := fhCookie.(*ctph.FuzzyHash)
	if !ok {
		http.Error(w, "Unable to retrieve FuzzyHash from session obj",
			http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var s stepReq
	err = decoder.Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.Data == 0x0 {
		// You could argue that 0x0 is a legitimate state, however in ascii it is NUL
		// Hence it's unlikely to be a legit input, however this is a default input if the
		// Client doesn't have a valid one, so we should return
		http.Error(w, "No data provided, no state to update", http.StatusNoContent)
		return
	}

	log.Printf("StepAlgo (algo=%s) request received: %#v\n", algo, s)

	fh.Step(s.Data)
	session.Values[algoStateGob] = fh
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fh)
}
