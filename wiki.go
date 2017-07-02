package main

import (
    "errors"
    "regexp"
    "io/ioutil"
    "net/http"
    "html/template"
)
/*
template.Must - convenience wrapper that panics when passed a non-nil error val
otherwise returns the *template unaltered
A panic is appropriate here, if the templates can't be loaded, only thing to do is exit pgrm
ParseFiles func takes any number of string args that identify our template files and parses those files into templates taht are named after base file name
** if adding more files to program, add to ParseFiles args
*/
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
/*
validPath - MustCompile will parse and compile the regexp
and return a regexp.
Regexep.MustCompile is distinct form Compile in that it will panic if the exp compiliation fails
*/
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")


/*
Page struct -
Title - title of page
Body - body of page (byte slice for io lib)
*/
type Page struct {
    Title string
    Body []byte
}

/*
save method --
@receiver: p - pointer to Page
@params: N/A
@return: error
= save Page's Body to a text file Title as file name
returns error val b/c return type of WriteFile is error
0600 - indicates file should be created with read-write permissions for curr user only
*/
func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

/*
loadPage method --
@parameter: title (string)
@returns: pointer to a Page literal constructed with proper title and body
= constructs the file name from title param, reads file's contents into new body
*** io.ReadFile returns []byte and error, in loadPage, error isn't being handled yet, "blank identifier" represented by _ is used to throw away the error return val
if second param nil, page successfully loaded, if not, error handled
*/
func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

/*
getTitle method --
if title is valid, it will be returned along with nil error value
if title is invalid, func will write a 404 not found error at HTTP conn
and return an error to the handler
*/
func getTitle(w http.ResponseWriter, r *http.Request) (string, error){
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil{
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil
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
viewHandler method --
gets page title from path sliced from /view/
loads page data, form, http.ResponseWriter
if /view/page doesn't exist, redirect client to edit Page so content may be created
http.Redirect - adds HTTP status code of http.SattusFound (302) and Location header to response
*/
func viewHandler(w http.ResponseWriter, r *http.Request, title string){
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

/*
editHandler method --
loads page or if doesn't exist, creates empty Page struct
displays HTML form
template.ParseFiles - reads contents of filename and returns a *template.Template
t.Execute - exdecutes template, writing generated HTML to http.ResponseWWriter
.Title and .Body indentifiers refer to p.Title and p.Body
== Template directives are enclosed in double {{}}
printf "%s" .Body - function call outputs .Body as string instead of stream of bytes
same as fmt.Printf
*/
func editHandler(w http.ResponseWriter, r *http.Request, title string){
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}
/*
saveHandler method --
handle submission of forms located on edit pages
page title and form's only field, Body, are stored in new Page
save() method called to write data to file and client redirected to /view/ page
FormValue - returns type string, must convert to []byte slice before it will fit into Page struct
accepts a title (string)
*/
func saveHandler(w http.ResponseWriter, r *http.Request, title string){
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

/*
makeHandler method --
wrapper function takes handler functions and returns a function of ttype http.HandlerFunc
fn is enclosed by closure, fn will be one of our save, edit, or view handlers
closure returned by makeHandler is a function that takes http.ResponseWriter and http.Request
then extracts title from request path, validates with TitleValidator regexp.
If title is invalid, error will be wwritten, ResponseWriter, using http.NotFound
If title is valid, enclosed handdler function fn will be called with the ResponseWriter, Request and title as args
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
        fn(w, r, m[2])
    }
}

/*
/view/ - view the page created from title
/edit/ - edit the page created from title
/save/ - save the page created from title
*/
func main(){
    /*p1 := &Page{Title: "TestPage", Body: []byte("This is a test Page.")}
    p1.save()
    p2, _ := loadPage("TestPage")
    fmt.Println(string(p2.Body))
    */
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.ListenAndServe(":8000", nil)
}
