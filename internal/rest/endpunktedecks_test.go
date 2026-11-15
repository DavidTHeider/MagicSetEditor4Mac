// **Die Wege für Decks und Formate** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort stehen die
// Wege **daneben** — Sammlung, Preise, Pakete —, hier die **Decks**.

package rest_test

import "net/http"

// endpunkteDecks sind die Wege für Decks und Formate — **gezählt wird in
// `FAEHIGKEITEN.md`** (#1287), eine Zahl hier liefe davon.
func endpunkteDecks() []endpunkt {
	return []endpunkt{
		{
			name: "Decks", methode: http.MethodGet, weg: "/decks",
			erfolg: http.StatusOK,
		},
		{
			name: "Deck anlegen", methode: http.MethodPost, weg: "/decks",
			rumpf:  `{"name":"Aggro","spiel":"magic","notiz":""}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Deck", methode: http.MethodGet, weg: "/decks/1",
			erfolg: http.StatusOK,
		},
		{
			name: "Deck umbenennen", methode: http.MethodPut, weg: "/decks/1",
			rumpf:  `{"fassung":1,"name":"Midrange","notiz":""}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Deck löschen", methode: http.MethodDelete, weg: "/decks/1",
			erfolg: http.StatusOK,
		},
		{
			// **`POST`, weil ein neues Deck entsteht** — DOM-03.AK-02 (#1698).
			name: "Deck kopieren", methode: http.MethodPost, weg: "/decks/1/kopie",
			rumpf:  `{"name":"Aggro II"}`,
			erfolg: http.StatusCreated,
		},
		{
			// **`POST`, weil ein neues Deck entsteht** — FMT-04 (#1693).
			name: "Deck aus einem Set", methode: http.MethodPost, weg: "/decks/aus-set",
			rumpf:  `{"setID":4,"name":"","menge":2}`,
			erfolg: http.StatusCreated,
		},
		{
			// **`POST` ohne Rumpf** — ein zweiter Ruf legt nichts doppelt an,
			// die Antwort nennt dann, was schon gewünscht ist.
			name: "Fehlkarten wünschen", methode: http.MethodPost,
			weg: "/decks/1/fehlkarten/wunschliste", erfolg: http.StatusOK,
		},
		{
			// **`PUT` auf die Sammlung der Plätze**: derselbe Bezug im selben
			// Bereich ist eine höhere Menge und kein zweiter Platz.
			name: "Deckplatz legen", methode: http.MethodPut,
			weg:    "/decks/1/plaetze",
			rumpf:  `{"printing":"a1b2c3d4-0000-0000-0000-00000000000a","menge":4,"bereich":"hauptdeck"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Deckplatz nehmen", methode: http.MethodDelete,
			weg: "/decks/1/plaetze/2", erfolg: http.StatusOK,
		},
		{
			// **Der Bereichswechsel ist ein Weg** — Befund 7 des Zweitblicks
			// über Phase 5.1.
			name: "Deckplatz verschieben", methode: http.MethodPut,
			weg:    "/decks/1/plaetze/2/bereich",
			rumpf:  `{"bereich":"sideboard"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Deckformat wählen", methode: http.MethodPut,
			weg: "/decks/1/format", rumpf: `{"formatID":5}`, erfolg: http.StatusOK,
		},
		{
			name: "Deck prüfen", methode: http.MethodGet, weg: "/decks/1/pruefung",
			erfolg: http.StatusOK,
		},
		{
			name: "Deckbestand", methode: http.MethodGet, weg: "/decks/1/bestand",
			erfolg: http.StatusOK,
		},
		// **Die Decklisteneinfuhr** — Werkstück T8 von Phase 5.6 (#1473).
		//
		// **`POST` für beides**: die Liste steht im Rumpf, und das
		// Übernehmen ist an den Stand gebunden — ein zweiter Ruf mit
		// demselben Stand wird abgewiesen und ist damit nicht idempotent.
		{
			name: "Deckliste deuten", methode: http.MethodPost,
			weg:    "/decks/1/liste/vorschau",
			rumpf:  `{"text":"4 Lightning Bolt (M11) 149"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Deckliste übernehmen", methode: http.MethodPost,
			weg: "/decks/1/liste",
			rumpf: `{"stand":"abc","zeilen":[{"printing":` +
				`"a1b2c3d4-0000-0000-0000-00000000000a","menge":4,` +
				`"bereich":"hauptdeck"}]}`,
			erfolg: http.StatusOK,
		},
		// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
		{
			name: "Deckzuordnungen", methode: http.MethodGet, weg: "/decks/1/zuordnungen",
			erfolg: http.StatusOK,
		},
		{
			name: "Zuordnen", methode: http.MethodPut, weg: "/decks/1/zuordnungen",
			rumpf:  `{"platzID":2,"exemplarID":3,"menge":1}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Zuordnung lösen", methode: http.MethodDelete,
			weg: "/decks/1/zuordnungen/4", erfolg: http.StatusOK,
		},
		{
			name: "Fehlkarten", methode: http.MethodGet, weg: "/decks/1/fehlkarten",
			erfolg: http.StatusOK,
		},
		{
			name: "Deckliste", methode: http.MethodGet, weg: "/decks/1/liste",
			erfolg: http.StatusOK,
		},
		{
			name: "Decks eines Exemplars", methode: http.MethodGet,
			weg: "/sammlung/exemplare/3/decks", erfolg: http.StatusOK,
		},
		{
			name: "Formate", methode: http.MethodGet, weg: "/formate",
			erfolg: http.StatusOK,
		},
		{
			name: "Format anlegen", methode: http.MethodPost, weg: "/formate",
			rumpf:  `{"name":"Standard","spiel":"magic","mindestgroesse":60}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Formatbauarten", methode: http.MethodGet, weg: "/formatbauarten",
			erfolg: http.StatusOK,
		},
		{
			name: "Format setzen", methode: http.MethodPut, weg: "/formate/5",
			rumpf:  `{"fassung":1,"name":"Standard","spiel":"magic"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Format löschen", methode: http.MethodDelete, weg: "/formate/5",
			erfolg: http.StatusOK,
		},
	}
}
