package main
import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

type templateHandler struct {
	once sync.Once
	filename string
	templ *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(
			template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "App address")
	flag.Parse()
	r := newRoom()
	http.HandleFunc("/", (&templateHandler{filename: "chat.html"}).ServeHTTP)
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting server on ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Listen and Serve", err)
	}
   
}
