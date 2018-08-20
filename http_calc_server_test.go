package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMissing(t *testing.T) {

	req, err := http.NewRequest("GET", "/add?x=&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(AddFunc))

	handler.ServeHTTP(rr, req)

	expected := `Malformed calculator request` + "\n"

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

	expected := `Malformed calculator request` + "\n"

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

func TestCalculatorSubtract(t *testing.T) {

	req, err := http.NewRequest("GET", "/subtract?x=4.2&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(SubFunc))

	handler.ServeHTTP(rr, req)

	expected := `{"action":"subtract","x":4.2,"y":2,"answer":2.2,"cached":false}`

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestCalculatorMultiply(t *testing.T) {

	req, err := http.NewRequest("GET", "/multiply?x=3&y=5", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(MultFunc))

	handler.ServeHTTP(rr, req)

	expected := `{"action":"multiply","x":3,"y":5,"answer":15,"cached":false}`

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestCalculatorBadDivide(t *testing.T) {

	req, err := http.NewRequest("GET", "/divide?x=6&y=0", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(DiviFunc))

	handler.ServeHTTP(rr, req)

	expected := `Nonnumber math answer` + "\n"

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestCalculatorDivide(t *testing.T) {

	req, err := http.NewRequest("GET", "/divide?x=6&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(DiviFunc))

	handler.ServeHTTP(rr, req)

	expected := `{"action":"divide","x":6,"y":2,"answer":3,"cached":false}`

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestCacheSimple(t *testing.T) {

	req, err := http.NewRequest("GET", "/divide?x=8&y=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(DiviFunc))

	handler.ServeHTTP(rr, req)

	expected1 := `{"action":"divide","x":8,"y":2,"answer":4,"cached":false}`

	handler.ServeHTTP(rr, req)

	expected2 := `{"action":"divide","x":8,"y":2,"answer":4,"cached":true}`

	expected := expected1 + expected2

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}

func TestCacheExpire(t *testing.T) {

	req, err := http.NewRequest("GET", "/divide?x=9&y=3", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCall(DiviFunc))

	handler.ServeHTTP(rr, req)

	expected1 := `{"action":"divide","x":9,"y":3,"answer":3,"cached":false}`

	time.Sleep(time.Minute + time.Second*5)
	handler.ServeHTTP(rr, req)

	expected2 := `{"action":"divide","x":9,"y":3,"answer":3,"cached":false}`

	time.Sleep(time.Second * 5)
	handler.ServeHTTP(rr, req)

	expected3 := `{"action":"divide","x":9,"y":3,"answer":3,"cached":true}`

	expected := expected1 + expected2 + expected3

	if rr.Body.String() != expected {

		t.Errorf("Did not receive expected result: got %s expected %s", rr.Body.String(), expected)

	}

}
