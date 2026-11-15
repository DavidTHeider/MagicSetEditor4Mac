// **Die Vorschauen der Einfuhr** — DOM-03.AK-01
// (#1698).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge.
package rest_test

import "net/http"

// endpunkteEinfuhrvorschau sind die zwei Wege, die sagen, was eine Einfuhr täte.
func endpunkteEinfuhrvorschau() []endpunkt {
	return []endpunkt{
		{
			name: "Liste vorschauen", methode: http.MethodPost,
			weg:    "/sammlung/liste/vorschau",
			rumpf:  `{"zeilen":[{"printing":"probe-druck","menge":1}]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Sammlung vorschauen", methode: http.MethodPost,
			weg:    "/sammlung/vorschau",
			rumpf:  `{"text":"printing,menge\nprobe-druck,1\n"}`,
			erfolg: http.StatusOK,
		},
	}
}
