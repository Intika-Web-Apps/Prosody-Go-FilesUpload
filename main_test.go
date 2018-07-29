package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func mockUpload() {
	os.MkdirAll(filepath.Dir(conf.Storedir+"thomas/abc/"), os.ModePerm)
	from, err := os.Open("./catmetal.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(conf.Storedir+"thomas/abc/catmetal.jpg", os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanup() {
	// Clean up
	if _, err := os.Stat(conf.Storedir); err == nil {
		// Delete existing catmetal picture
		err := os.RemoveAll(conf.Storedir)
		if err != nil {
			log.Println("Error while cleaning up:", err)
		}
	}
}

func TestReadConfig(t *testing.T) {
	// Set config
	err := readConfig("config.toml", &conf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUploadValid(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// Read catmetal file
	catmetalfile, err := ioutil.ReadFile("catmetal.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// Create request
	req, err := http.NewRequest("PUT", "/upload/thomas/abc/catmetal.jpg", bytes.NewBuffer(catmetalfile))
	q := req.URL.Query()
	q.Add("v", "e17531b1e88bc9a5cbf816eca8a82fc09969c9245250f3e1b2e473bb564e4be0")
	req.URL.RawQuery = q.Encode()

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusOK, rr.Body.String())
	}

	// clean up
	cleanup()
}

func TestUploadMissingMAC(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// Read catmetal file
	catmetalfile, err := ioutil.ReadFile("catmetal.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// Create request
	req, err := http.NewRequest("PUT", "/upload/thomas/abc/catmetal.jpg", bytes.NewBuffer(catmetalfile))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusConflict, rr.Body.String())
	}
}

func TestUploadInvalidMAC(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// Read catmetal file
	catmetalfile, err := ioutil.ReadFile("catmetal.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// Create request
	req, err := http.NewRequest("PUT", "/upload/thomas/abc/catmetal.jpg", bytes.NewBuffer(catmetalfile))
	q := req.URL.Query()
	q.Add("v", "abc")
	req.URL.RawQuery = q.Encode()

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusForbidden, rr.Body.String())
	}
}

func TestUploadInvalidMethod(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// Read catmetal file
	catmetalfile, err := ioutil.ReadFile("catmetal.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/upload/thomas/abc/catmetal.jpg", bytes.NewBuffer(catmetalfile))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusMethodNotAllowed, rr.Body.String())
	}
}

func TestDownloadHead(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// Mock upload
	mockUpload()

	// Create request
	req, err := http.NewRequest("HEAD", "/upload/thomas/abc/catmetal.jpg", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusOK, rr.Body.String())
	}

	// cleanup
	cleanup()
}

func TestDownloadGet(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// moch upload
	mockUpload()

	// Create request
	req, err := http.NewRequest("GET", "/upload/thomas/abc/catmetal.jpg", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusOK, rr.Body.String())
	}

	// cleanup
	cleanup()
}

func TestEmptyGet(t *testing.T) {
	// Set config
	readConfig("config.toml", &conf)

	// Create request
	req, err := http.NewRequest("GET", "", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRequest)

	// Send request and record response
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v. HTTP body: %s", status, http.StatusForbidden, rr.Body.String())
	}
}
