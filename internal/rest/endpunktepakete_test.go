package rest_test

// **Plan, Ersetzen und Zurück** — DOM-01.AK-02 und AK-03
// (#1696).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge.

import "net/http"

// endpunktePakete sind die Wege rund um den Plan einer Paketänderung und die
// Bezugsstände.
func endpunktePakete() []endpunkt {
	return []endpunkt{
		{
			// **Ein `GET`**: der Plan ändert nichts.
			name: "Paketplan lesen", methode: http.MethodGet,
			weg: "/pakete/plan?vorgang=entfernen&art=mse-style&name=probe", erfolg: http.StatusOK,
		},
		{
			name: "Paket ersetzen", methode: http.MethodPut, weg: "/pakete/ersetzt",
			rumpf: `{"quelle":"/tmp/probe.mse-style","bestaetigt":true}`, erfolg: http.StatusOK,
		},
		{
			name: "Paket zurück", methode: http.MethodPost,
			weg: "/pakete/mse-style/probe/zurueck", erfolg: http.StatusOK,
		},
		{
			// **Am Set**, denn der Bedarf ist eine Frage an ein Set (AK-04).
			name: "Paketbedarf eines Sets", methode: http.MethodGet,
			weg: "/sets/1/paketbedarf", erfolg: http.StatusOK,
		},
		// **Die Stände der bezogenen Bäume** — G04 (#2251).
		{
			name: "Bezugsstände lesen", methode: http.MethodGet,
			weg: "/paketbezuege?fragen=1", erfolg: http.StatusOK,
		},
		{
			name: "Bezugsvorschau lesen", methode: http.MethodGet,
			weg: "/paketbezuege/M15-Magic-Pack/vorschau?stand=" + bezugsziel, erfolg: http.StatusOK,
		},
		{
			name: "Bezug aktualisieren", methode: http.MethodPost,
			weg:   "/paketbezuege/M15-Magic-Pack/aktualisierung",
			rumpf: `{"stand":"` + bezugsziel + `","bestaetigt":true}`, erfolg: http.StatusOK,
		},
		{
			name: "Bezug zurück", methode: http.MethodPost,
			weg: "/paketbezuege/M15-Magic-Pack/zurueck", erfolg: http.StatusOK,
		},
	}
}

// bezugsziel ist ein Commit, wie eine Aktualisierung ihn verlangt.
const bezugsziel = "c8e38c9b138c5e9c66199fd099a353bce6ff5ec7"
