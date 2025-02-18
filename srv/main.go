package main

import (
	"net/http"
)

const WEBSITE_CONTENT = `
<html>
    <head>
    </head>
    <body>
        <a href="https://www.youtube.com/watch?v=dQw4w9WgXcQ">ur gay</a>
    </body>
</html>
`

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(WEBSITE_CONTENT))
	})

  http.ListenAndServe("10.69.42.16:8000", mux);
}
