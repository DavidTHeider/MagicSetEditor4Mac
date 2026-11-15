// **Die Wege für Turniere** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, weil `endpunktedecks_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht die
// **Sammlung**, hier der **Verlauf**.

package rest_test

import "net/http"

// endpunkteTurniere sind die fünfzehn Wege für Turniere.
func endpunkteTurniere() []endpunkt {
	return []endpunkt{
		{
			name: "Turniere", methode: http.MethodGet, weg: "/turniere",
			erfolg: http.StatusOK,
		},
		{
			name: "Turnier anlegen", methode: http.MethodPost, weg: "/turniere",
			rumpf:  `{"name":"Küchentisch","spiel":"magic","notiz":""}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Turnier", methode: http.MethodGet, weg: "/turniere/1",
			erfolg: http.StatusOK,
		},
		{
			name: "Turnier löschen", methode: http.MethodDelete, weg: "/turniere/1",
			erfolg: http.StatusNoContent,
		},
		{
			name: "Teilnehmer anmelden", methode: http.MethodPost,
			weg:    "/turniere/1/teilnehmer",
			rumpf:  `{"name":"Ada"}`,
			erfolg: http.StatusCreated,
		},
		{
			// **`PUT` und nicht `DELETE`**: er steigt aus, er verschwindet
			// nicht. Seine Quote geht weiter in die seiner Gegner ein.
			name: "Teilnehmer aussteigen", methode: http.MethodPut,
			weg:    "/turniere/1/teilnehmer/2",
			rumpf:  `{"ausgestiegen":true}`,
			erfolg: http.StatusNoContent,
		},
		{
			name: "Rangliste", methode: http.MethodGet, weg: "/turniere/1/rangliste",
			erfolg: http.StatusOK,
		},
		{
			// **`GET`, weil nichts entsteht** — das ist der ganze Unterschied
			// zu `POST /runden`.
			name: "Paarungsvorschlag", methode: http.MethodGet,
			weg: "/turniere/1/vorschlag", erfolg: http.StatusOK,
		},
		{
			name: "Runde legen", methode: http.MethodPost, weg: "/turniere/1/runden",
			erfolg: http.StatusCreated,
		},
		{
			// **Eine Vorlage** — DOM-04 (#1699): ein neues Turnier entsteht.
			name: "Turnier kopieren", methode: http.MethodPost,
			weg:    "/turniere/1/kopie",
			rumpf:  `{"name":"Küchentisch II","mitTeilnehmenden":true}`,
			erfolg: http.StatusCreated,
		},
		{
			// **Die Folgen einer Berichtigung** — DOM-04.AK-02 (#1699): ein
			// `POST`, weil der Stand in den Rumpf gehört, und ein `200`, weil
			// nichts entsteht.
			name: "Folgen einer Berichtigung", methode: http.MethodPost,
			weg:    "/turniere/1/paarungen/3/folgen",
			rumpf:  `{"spieleB":2}`,
			erfolg: http.StatusOK,
		},
		{
			// **Abschluss und Protokoll** — DOM-04.AK-04 (#1699). Der
			// Abschluss ist eine Ressource des Turniers: `POST` setzt ihn,
			// `DELETE` nimmt ihn zurück.
			name: "Turnier abschließen", methode: http.MethodPost,
			weg: "/turniere/1/abschluss", erfolg: http.StatusNoContent,
		},
		{
			name: "Turnier wiedereröffnen", methode: http.MethodDelete,
			weg: "/turniere/1/abschluss", erfolg: http.StatusNoContent,
		},
		{
			name: "Protokoll", methode: http.MethodGet,
			weg: "/turniere/1/protokoll", erfolg: http.StatusOK,
		},
		{
			// **Datei und Bericht** — DOM-04.AK-03 (#1699).
			name: "Turnier ausgeben", methode: http.MethodGet,
			weg: "/turniere/1/ausgabe", erfolg: http.StatusOK,
		},
		{
			name: "Turnierbericht", methode: http.MethodGet,
			weg: "/turniere/1/bericht", erfolg: http.StatusOK,
		},
		{
			name: "Turnier einlesen", methode: http.MethodPost,
			weg: "/turniere/einlesen", rumpf: `{"inhalt":"{}"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Ergebnis setzen", methode: http.MethodPut,
			weg:    "/turniere/1/paarungen/3",
			rumpf:  `{"spieleA":2,"spieleB":0}`,
			erfolg: http.StatusNoContent,
		},
	}
}
