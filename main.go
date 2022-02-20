package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type requestPayloadStruct struct {
	ProxyCondition string `json:"proxy_condition"`
}

func getEnv(params ...string) string {
	if value, exists := os.LookupEnv(params[0]); exists {
		return value
	}
	return params[1]
}
func getListenAddress() string {
	port := getEnv("PORT", "8080")
	return ":" + port
}
func logSetup() {
	upstream_a := getEnv("UPSTREAM_A")
	upstream_b := getEnv("UPSTREAM_B")
	default_upstream := getEnv("DEFAULT_UPSTREAM")
	log.Println("Server will run on " + getListenAddress())
	log.Println("UPSTREAM_A: " + upstream_a)
	log.Println("UPSTREAM_B: " + upstream_b)
	log.Println("DEFAULT_UPSTREAM: " + default_upstream)

}
func requestBodyDecoder(request *http.Request) *json.Decoder {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
}

func parseRequestBody(request *http.Request) requestPayloadStruct {
	decoder := requestBodyDecoder(request)
	var requestPayload requestPayloadStruct
	err := decoder.Decode(&requestPayload)
	if err != nil {
		panic(err)
	}
	return requestPayload
}
func logRequestPayload(payload requestPayloadStruct, proxyUrl string) {
	log.Printf("proxy_condition: %s , proxy_url: %s", payload.ProxyCondition, proxyUrl)
}
func getProxyUrl(condition string) string {
	a_url := getEnv("UPSTREAM_A")
	b_url := getEnv("UPSTREAM_B")
	default_url := getEnv("DEFAULT_UPSTREAM")
	upperCond := strings.ToUpper(condition)
	switch upperCond {
	case "A":
		return a_url
	case "B":
		return b_url
	default:
		return default_url
	}
}
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host
	proxy.ServeHTTP(res, req)
}
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	requestPayload := parseRequestBody(req)
	url := getProxyUrl(requestPayload.ProxyCondition)
	logRequestPayload(requestPayload, url)
	serveReverseProxy(url, res, req)
}
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	logSetup()
	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}

}
