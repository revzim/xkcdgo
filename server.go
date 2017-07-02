package main

import (
    "fmt"
    "net/http"
)

/*
handler method --
http.HandlerFunc
@receives: http.ResponseWriter 
        http.Request
http.ResponseWriter - assembles the HTTP server's response by writing to it
then we send data to HTTP client
http.Request - data struct represents the client HTTP request 
r.URL.Path - path component of request URL
[1:] - create subslice of Path from 1st char to end
*/
func handler(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Hey I love %s!", r.URL.Path[1:])
}

/*
main method --
call http.HandleFunc - http package to handle all requests to web root / with handler
- http.ListenAndServe - listen on port 8000 on any interface
*/
func main(){
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8000", nil)
}
