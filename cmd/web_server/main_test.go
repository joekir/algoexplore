package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestInit_withValidLengthField_ReturnsSerializedFHStruct(t *testing.T) {
	var jsonStr = []byte(`{"data_length": 15}`)
	req, err := http.NewRequest("POST", "/ctph/init", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/{algo}/init", Init)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusCreated)
	}

	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v\n", ctype, "application/json")
	}

	t.Logf("%#v\n", rr.Body.String())
}

func TestStepAlgo_noSession_Returns500(t *testing.T) {
	var jsonStr = []byte(`{"byte": 103}`)
	req, err := http.NewRequest("POST", "/ctph/step", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/{algo}/step", StepAlgo)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusPreconditionRequired {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusPreconditionRequired)
	}
}

func TestStepAlgo_withSession_StepsItByOne(t *testing.T) {
	var jsonStr = []byte(`{"data_length": 10}`)
	req, err := http.NewRequest("POST", "/ctph/init", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter() // need this to text mux Vars
	router.HandleFunc("/{algo}/init", Init)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusCreated)
	}

	cookies := rr.Result().Cookies()
	var ctphCookie *http.Cookie
	for i := range cookies {
		if cookies[i].Name == sessionCookieName {
			ctphCookie = cookies[i]
		}
	}

	if nil == ctphCookie {
		t.Error("First call did not return a cookie")
	}

	jsonStr = []byte(`{"byte": 103}`)
	req, err = http.NewRequest("POST", "/ctph/step", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	// These need to be re initialized to avoid
	// polution from first part of this test
	rr = httptest.NewRecorder()
	router = mux.NewRouter()
	router.HandleFunc("/{algo}/step", StepAlgo)
	req.AddCookie(ctphCookie)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v\n",
			status, http.StatusOK)
	}

	jsonStr, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Errorf("failed to read response body: %s", err.Error())
	}

	if !strings.Contains(string(jsonStr), `\"window\":[103`) {
		t.Fatalf("expected `window[103,...`, got %s", jsonStr)
	}
}
