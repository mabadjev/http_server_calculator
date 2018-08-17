package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

//this object holds our answer and is serialized for our response
type Answer struct {
	Action    string  `json:"action"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Answer    float64 `json:"answer"`
	Cached    bool    `json:"cached"`
	ErrorName error   `json:"error,omitempty"`
	timeout   *time.Timer
}

//Specify JSON serialization for our answer

func (u *Answer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		LastSeen int64  `json:"lastSeen"`
		Action    string  `json:"action"`
		X         float64 `json:"x"`
		Y         float64 `json:"y"`
		Answer    float64 `json:"answer"`
		Cached    bool    `json:"cached"`
		ErrorName error   `json:"error,omitempty"`
	}{
		ID:       u.ID,
		Name:     u.Name,
		LastSeen: u.LastSeen.Unix(),
	})
}

//Using synchronous map to avoid having to use ReadWriteMutex in conjunction with a map to avoid contention during high volume of requests
//According to documentation sync.Map is optimized for performance when many reads may be done on a map that is relatively static
//This is generally what occurs with the syncmap during a high volume of requests

//Map of URI: Answer or key: value  request: answer to return
//intended for [URI]Answer
var responseMap sync.Map

//Static string representing malformed request used
var CalcRequestError error = errors.New("Malformed calculator request")

//Static string representing a math error EG div by zero
var CalcMathError error = errors.New("Nonnumber math answer")

//Static integer representing duration before cache timeout in ms
const cacheDur = time.Minute

//
type MarshalledError error 

type DoMath func(float64, float64) (float64, error)

//Grabs the arguments from the query
//input: Values from request url
//output arguments or error if error is encountered
func extractArgs(v url.Values) (float64, float64, error) {

	rawx := v.Get("x")

	if len(rawx) == 0 {
		return -999, -999, CalcRequestError
	}

	rawy := v.Get("y")

	if len(rawy) == 0 {
		return -999, -999, CalcRequestError
	}

	x, err := strconv.ParseFloat(rawx, 64)
	if err != nil {
		return -999, -999, CalcRequestError
	}

	y, err := strconv.ParseFloat(rawy, 64)
	if err != nil {
		return -999, -999, CalcRequestError
	}

	return x, y, nil

}

//Returns json containing completed operation or error response if appropriate
//inputs: URL of request
//outputs: Answer struct with processed request
func assembleAnswer(u url.URL, math DoMath) Answer {

	v := u.Query()
	opp := u.Path
	op := opp[1:]

	//create new timer and start it off
	//it'll time an entry out automatically after the duration

	f := makeTimeout(u)
	tnew := time.AfterFunc(cacheDur, f)

	x, y, err := extractArgs(v)
	if err != nil {

		a := Answer{op, 0, 0, 0, false, err, tnew}
		return a

	}

	result, err := math(x, y)
	if err != nil {
		a := Answer{op, 0, 0, 0, false, err, tnew}
		return a
	}

	a := Answer{op, x, y, result, false, nil, tnew}
	return a

}

//builds a timeout to be used in the timer
func makeTimeout(u url.URL) func() {

	return func() {
		responseMap.Delete(u.RequestURI())
	}
}

//reset the timeout for a request
func resetTimeout(u url.URL) {

	//lookup the associated Answer
	aa, ok := responseMap.Load(u.RequestURI())

	//if we somehow miss, we raced and its already removed. In this case, let it go
	if ok == false {
		return
	}

	a := aa.(*Answer)

	t := a.timeout

	t.Stop()

	//create new timer and start it off
	//it'll time an entry out automatically after the duration

	f := makeTimeout(u)
	tnew := time.AfterFunc(cacheDur, f)

	//store the timer in the correct answer again
	a.timeout = tnew

}

//Generates handler for requests on the given port
//inputs: automatically gets responsewriter and request and math command
//outputs: handler function for handling requests
//Used as a closure to generate the appropriate

func handleCall(mathF DoMath) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		u := req.URL

		ii, ok := responseMap.Load(u.RequestURI())

		//if ok true, we had a hit in cache
		if ok == true {

			i := ii.(*Answer)

			j, err := json.Marshal(i)
			if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				http.Error(w, "Error building JSON", http.StatusInternalServerError)
				return
			}

			//return cached response
			w.Write(j)

			//pass request and answer to be recached and exit handler
			go resetTimeout(*u)

		} else {
			//Didn't hit in cache, so perform the operation and cache result
			a := assembleAnswer(*u, mathF)

			j, err := json.Marshal(a)
			if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				http.Error(w, "Error building JSON", http.StatusInternalServerError)
				return
			}

			//return response first, then cache afterwards for faster response
			w.Write(j)

			//set Cached flag true before caching
			a.Cached = true

			//Cache answer and end handler
			responseMap.Store(u.RequestURI(), &a)
		}
	}

}

func main() {

	//Accept input for port to monitor for calculator
	//Assume 8080 by default

	//Only take port for now
	if len(os.Args) > 2 {
		fmt.Printf("Error: only expected argument is port to monitor. Rerun with only one argument or no argument for default of 8080")
		os.Exit(1)
	}

	var portStr string

	//Default to 8080
	if len(os.Args) == 1 {
		portStr = ":8080"
	} else {
		//Don't check if port is valid here - will allow to error out below if invalid
		portStr = ":" + os.Args[1]
	}

	//handle add requests
	http.HandleFunc("/add", handleCall(AddFunc))

	//handle subtract requests
	http.HandleFunc("/subtract", handleCall(SubFunc))

	//handle multiply requests
	http.HandleFunc("/multiply", handleCall(MultFunc))

	//handle divide requests
	http.HandleFunc("/divide", handleCall(DiviFunc))

	log.Fatal(http.ListenAndServe(portStr, nil))
}

func AddFunc(x float64, y float64) (float64, error) { return x + y, nil }

func SubFunc(x float64, y float64) (float64, error) { return x - y, nil }

func MultFunc(x float64, y float64) (float64, error) { return x * y, nil }

func DiviFunc(x float64, y float64) (float64, error) {
	if y == 0 {
		return 0, CalcMathError
	}
	return x / y, nil
}