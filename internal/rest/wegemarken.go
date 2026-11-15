package rest

// **Die Wege der Marken an jedem Fachobjekt** — UX-12
// (#1684).
//
// **Eine eigene Datei**, weil `wegetafel.go` an der Zeilengrenze aus §2
// steht; `vollstaendig_test.go` liest sie mit.

import "net/http"

// objektmarkenwege meldet die Wege an.
func (d *Dienst) objektmarkenwege(weg *http.ServeMux) {
	weg.HandleFunc("GET /objektmarken", d.markenkatalogZeigen)
	weg.HandleFunc("GET /objektmarken/{art}", d.markenJeObjekt)
	weg.HandleFunc("GET /objektmarken/{art}/filter", d.markenFiltern)
	weg.HandleFunc("POST /objektmarken/{art}/vergabe", d.markeVergeben)
	weg.HandleFunc("GET /objektmarken/{art}/{kennung}", d.objektmarkenZeigen)
	weg.HandleFunc("PUT /objektmarken/{art}/{kennung}/{marke}", d.objektmarkeSetzen)
	weg.HandleFunc("DELETE /objektmarken/{art}/{kennung}/{marke}", d.objektmarkeEntfernen)
	weg.HandleFunc("POST /marken/umbenennung", d.markeUmbenennen)
	weg.HandleFunc("DELETE /marken/{marke}", d.markeUeberallEntfernen)
}
