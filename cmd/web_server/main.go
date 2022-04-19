package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joekir/algoexplore"
	"github.com/joekir/algoexplore/internal/algos/ctph"
)

const (
	algoState                  = "ALGO_STATE"
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
	cookieStore = sessions.NewCookieStore([]byte(os.Getenv(cookieSessionKeyEnvVarName)))
}

func main() {
	if len(listeningPort) < 1 {
		listeningPort = "8080"
	}

	router := mux.NewRouter()
	router.HandleFunc("/{algo}/init", Init).Methods("POST")
	router.HandleFunc("/{algo}/step", StepAlgo).Methods("POST")
	workingDir, err := os.Getwd()

	flag.Parse()
	if err != nil {
		glog.Fatal(err)
	}
	staticDir := path.Join(workingDir, "/static/")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	glog.Infof("Listening on %s\n", listeningPort)
	glog.Fatal(http.ListenAndServe(":"+listeningPort, router))
}

type hashReq struct {
	DataLength int `json:"data_length"`
}

func validateAlgo(vars map[string]string) *algoexplore.AlgoPlugin {
	algoName := vars["algo"]

	algo, err := algoexplore.GetAlgo(algoName)
	if err != nil {
		glog.Fatal("valid algorithm not found")
		return nil
	}

	return &algo
}

func Init(w http.ResponseWriter, r *http.Request) {
	algo := validateAlgo(mux.Vars(r))

	var h hashReq
	var body io.Reader
	body = r.Body
	if err := algoexplore.StrictUnmarshalJSON(&body, &h); err != nil {
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
		glog.Infoln("Deleting old cookie")
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

	// https://golang.org/doc/effective_go.html#type_switch
	var state string
	switch algo := (*algo).(type) {
	case *ctph.Ctph:
		algo.Init(h.DataLength)
		state = algo.SerializeState()
		glog.Infoln("registering ctph algorithm")
	default:
		glog.Fatal("Unable to concretize algo type")
	}
	glog.Infof("state: %#v\n", state)

	session.Values[algoState] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		glog.Errorln("no data provided")
		http.Error(w, "No data provided, no state to update", http.StatusNoContent)
		return
	}

	state := session.Values[algoState].(string)

	// https://golang.org/doc/effective_go.html#type_switch
	switch algo := (*algoInfo).(type) {
	case *ctph.Ctph:
		if err := algo.DeserializeState(state); err != nil {
			glog.Fatalf("Failed to deserialize state: %s", err.Error())
		}
		algo.Step(s.Data)
		state = algo.SerializeState()
	default:
		glog.Fatal("Unable to concretize algo type")
	}
	glog.Infof("state: %#v\n", state)

	session.Values[algoState] = state
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
