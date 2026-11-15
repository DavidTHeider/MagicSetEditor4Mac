// **Die Wege der Symbolbibliothek** — DOM-02
// (#1697).

package rest_test

import "net/http"

// endpunkteSymbolbibliothek sind die zwölf Wege.
func endpunkteSymbolbibliothek() []endpunkt {
	return []endpunkt{
		{name: "Symbolbibliothek", methode: http.MethodGet, weg: "/symbole", erfolg: http.StatusOK},
		{
			name: "Symbol aufnehmen", methode: http.MethodPost, weg: "/symbole",
			rumpf: `{"setID":1,"name":"Krone"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Symbol einlesen", methode: http.MethodPost, weg: "/symbole/einlesen",
			rumpf: `{"inhalt":"mse_version: 0.3.5","name":"Reif"}`, erfolg: http.StatusCreated,
		},
		{name: "Symbol als Text", methode: http.MethodGet, weg: "/symbole/1", erfolg: http.StatusOK},
		{
			name: "Symbol duplizieren", methode: http.MethodPost, weg: "/symbole/1/kopie",
			rumpf: `{"name":"Zacke"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Symbol umbenennen", methode: http.MethodPut, weg: "/symbole/1/name",
			rumpf: `{"name":"Reif"}`, erfolg: http.StatusOK,
		},
		{
			name: "Symbol löschen", methode: http.MethodDelete, weg: "/symbole/1?bestaetigt=true",
			erfolg: http.StatusOK,
		},
		{
			name: "Symbol einsetzen", methode: http.MethodPut, weg: "/sets/1/bibliothekssymbol",
			rumpf: `{"id":1}`, erfolg: http.StatusOK,
		},

		// **Die Bearbeitung** — DOM-02.AK-02.
		{name: "Symbolteile", methode: http.MethodGet, weg: "/symbole/1/teile", erfolg: http.StatusOK},
		{
			name: "Symbol sichern", methode: http.MethodPut, weg: "/symbole/1/teile",
			rumpf: `{"teile":[]}`, erfolg: http.StatusOK,
		},
		{name: "Symbol zurücknehmen", methode: http.MethodPost, weg: "/symbole/1/zuruecknahme", erfolg: http.StatusOK},
		{name: "Symbol nachziehen", methode: http.MethodPost, weg: "/symbole/1/nachzug", erfolg: http.StatusOK},
	}
}
