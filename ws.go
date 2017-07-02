package main

import (
	
	"github.com/nishanths/go-xkcd"
	"net/http"
	"log"
	"regexp"
	/*
	"errors"
	"io/ioutil"
	"fmt"
	*/
	"html/template"
	"time"
	"os"
	
)

//template for page
var templates = template.Must(template.ParseFiles("xkcdindex.html"))

/*
validPath - MustCompile will parse and compile the regexp
and return a regexp.
Regexep.MustCompile is distinct form Compile in that it will panic if the exp compiliation fails
*/
var validPath = regexp.MustCompile("^/(xkcd)") ///([a-zA-Z0-9]+)$

//struct for page
//xkcd
type Page struct {
	Title string
	Comic
}
type Comic struct {
	Alt string
	ImgUrl string
	Number int
	cTitle string
	Transcript string
}

/*
save method --
saves previous comics
*/
func (p *Page) save() int {
	filename := "prevcomics.txt"
	f, err := os.OpenFile (filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	dlog := "[" + time.Now().String() + "]\n" + p.Comic.cTitle + "+" + string(p.Comic.Number) + ":" + p.Comic.ImgUrl + "\n"
	n, err := f.WriteString(dlog)
	f.Close()
	return n
}

/*
loadPage method --
loads the comic page
*/
func loadPage() (*Page, error){
	client := xkcd.NewClient()
	comic, err := client.Random()
	if err != nil { log.Fatal(err) }
	return &Page{Title: comic.SafeTitle,
		Comic: Comic{ Alt: comic.Alt,
				ImgUrl: comic.ImageURL,
				Number: comic.Number,
				cTitle: comic.Title,
				Transcript: comic.Transcript} }, nil
}

/*
general renderTemplate func
http.Eerror sends specified internalservice error response code and err msg
*/
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

/*
comicHandler method --

*/
func comicHandler(w http.ResponseWriter, r *http.Request, title string){
	p, err := loadPage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "xkcdindex", p)
}


/*
makeHandler method --
wrapper function takes handler functions and returns a function of type http.HandlerFunc
fn is enclosed by closure, fn will be one of the pages available
closure returned by makeHandler is a function that takes http.ResponseWriter and http.Request
then extracts title from request path, validates with TitleValidator regexp.
If title is invalid, error will be written, ResponseWriter, using http.NotFound
If title is valid, enclosed handler function fn will be called with the ResponseWriter, Request and title as args
*/
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        //extract page title from Request
        //call provided handler 'fn'
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil{
            http.NotFound(w, r)
            return
        }
        fn(w, r, "")
    }
}

func main(){
	http.HandleFunc("/xkcd/", makeHandler(comicHandler))
	http.ListenAndServe(":8000", nil)
}