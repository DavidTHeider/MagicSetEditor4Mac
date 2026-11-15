package rest_test

// **Die Wege der Marken an jedem Fachobjekt** — UX-12
// (#1684).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge.

import "net/http"

// endpunkteMarken sind die neun Wege der Marken an jedem Fachobjekt.
func endpunkteMarken() []endpunkt {
	return []endpunkt{
		{name: "den Markenkatalog lesen", methode: http.MethodGet, weg: "/objektmarken", erfolg: http.StatusOK},
		{name: "die Marken je Objekt lesen", methode: http.MethodGet, weg: "/objektmarken/deck", erfolg: http.StatusOK},
		{
			name: "nach Marken filtern", methode: http.MethodGet,
			weg: "/objektmarken/deck/filter?marke=rot&marke=blau&verknuepfung=oder", erfolg: http.StatusOK,
		},
		{name: "die Marken eines Objekts lesen", methode: http.MethodGet, weg: "/objektmarken/deck/11", erfolg: http.StatusOK},

		// **Ein Paket heißt `art/name`** — mit `%2F` im Weg.
		{name: "eine Marke an ein Paket hängen", methode: http.MethodPut, weg: "/objektmarken/paket/game%2Fmagic/Rot", erfolg: http.StatusOK},
		{name: "eine Marke von einem Objekt nehmen", methode: http.MethodDelete, weg: "/objektmarken/deck/11/rot", erfolg: http.StatusOK},
		{
			name: "eine Marke vergeben", methode: http.MethodPost, weg: "/objektmarken/deck/vergabe",
			rumpf: `{"kennungen":["11","12"],"marke":"rot"}`, erfolg: http.StatusOK,
		},
		{
			name: "eine Marke umbenennen", methode: http.MethodPost, weg: "/marken/umbenennung",
			rumpf: `{"alt":"rot","neu":"blau","zusammenfuehren":true}`, erfolg: http.StatusOK,
		},
		{name: "eine Marke überall entfernen", methode: http.MethodDelete, weg: "/marken/rot", erfolg: http.StatusOK},
	}
}
