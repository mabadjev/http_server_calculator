package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMissing(t *testing.T) {

	req, err := http.NewRequest("GET", "/add?x=&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(AddFunc))

	handler.ServeHTTP(rr, req)

	expected := `{"action":"add","x":0,"y":0,"answer":0,"cached":false,"error":"Malformed calculator request"}`

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestBadInput(t *testing.T) {

	req, err := http.NewRequest("GET", "/add?x=a&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(AddFunc))

	handler.ServeHTTP(rr, req)

	expected := `{"action":"add","x":0,"y":0,"answer":0,"cached":false,"error":"Malformed calculator request"}`

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestCalculatorAdd(t *testing.T) {

	req, err := http.NewRequest("GET", "/add?x=4&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(AddFunc))

	handler.ServeHTTP(rr, req)

	expected := `{"action":"add","x":4,"y":2,"answer":6,"cached":false}`

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}
