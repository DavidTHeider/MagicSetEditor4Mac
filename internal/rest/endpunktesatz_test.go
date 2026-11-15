// **Die Endpunkte am Set und an seinen Teilen** — aus
// `endpunktliste_test.go` herausgelöst am 2026-09-06, weil jene über die
// Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// **Dort stehen die Wege am Bestand als Ganzem** — Nutzer, Profile, Sicherung,
// Verschlüsselung, Aktualisierung; **hier die am Set und an dem, was in ihm
// liegt**: Stile, Karten, Pakete, Decks, Turniere, Sammlung.
package rest_test

import "net/http"

// endpunkteAmSet gibt die Wege, die ein Set oder seinen Inhalt betreffen.
func endpunkteAmSet() []endpunkt {
	return []endpunkt{
		// **Die Probleme eines ganzen Sets** (PERF02, #2352): lesen und prüfen.
		{name: "Probleme lesen", methode: http.MethodGet, weg: "/sets/1/probleme", erfolg: http.StatusOK},
		{name: "Probleme prüfen", methode: http.MethodPost, weg: "/sets/1/probleme", erfolg: http.StatusOK},

		// **Stile, Pakete, Kennzahlen, Verlauf und Einstellungen**
		// (#1107).
		{name: "Stile lesen", methode: http.MethodGet, weg: "/sets/1/stile", erfolg: http.StatusOK},
		{
			name: "Stilfolgen lesen", methode: http.MethodGet,
			weg: "/sets/1/stile/probe-schlicht/folgen", erfolg: http.StatusOK,
		},
		{
			name: "Stil wechseln", methode: http.MethodPut, weg: "/sets/1/stil",
			rumpf: `{"paket":"probe-schlicht"}`, erfolg: http.StatusOK,
		},
		{
			name: "Stiloptionen lesen", methode: http.MethodGet,
			weg: "/sets/1/stiloptionen", erfolg: http.StatusOK,
		},
		{
			name: "Stiloption setzen", methode: http.MethodPut,
			weg: "/sets/1/stiloptionen/rand", rumpf: `{"wert":"breit"}`, erfolg: http.StatusOK,
		},
		{
			name: "Stiloption zurücksetzen", methode: http.MethodDelete,
			weg: "/sets/1/stiloptionen/rand", erfolg: http.StatusOK,
		},
		// **Die Zusatzfelder einer Karte** (#2298).
		{
			name: "Zusatzfelder lesen", methode: http.MethodGet,
			weg: "/sets/1/karten/2/zusatzfelder", erfolg: http.StatusOK,
		},
		{
			name: "Zusatzwert setzen", methode: http.MethodPut,
			weg: "/sets/1/karten/2/zusatzfelder/vermerk", rumpf: `{"wert":"neu"}`, erfolg: http.StatusOK,
		},
		{name: "Pakete lesen", methode: http.MethodGet, weg: "/pakete", erfolg: http.StatusOK},
		{
			name: "Paket beziehen", methode: http.MethodPost, weg: "/pakete",
			rumpf:  `{"adresse":"https://example.invalid/x","stand":"abc1234","unterverzeichnis":"data","name":"probe"}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Statistik lesen", methode: http.MethodGet,
			weg: "/sets/1/statistik", erfolg: http.StatusOK,
		},
		{
			name: "Verteilung lesen", methode: http.MethodGet,
			weg: "/sets/1/statistik/name", erfolg: http.StatusOK,
		},
		{
			name: "Zurücknehmen", methode: http.MethodPost,
			weg: "/sets/1/zuruecknehmen", erfolg: http.StatusOK,
		},
		{
			name: "Wiederherstellen", methode: http.MethodPost,
			weg: "/sets/1/wiederherstellen", erfolg: http.StatusOK,
		},
		{
			name: "Einstellungen lesen", methode: http.MethodGet,
			weg: "/einstellungen", erfolg: http.StatusOK,
		},
		{
			name: "Einstellung setzen", methode: http.MethodPut,
			weg: "/einstellungen/sprache", rumpf: `{"wert":"en"}`, erfolg: http.StatusOK,
		},
		// **Der Einstellungsaustausch** — Werkstück O3 von Phase 5.2 (#1427):
		// Ausfuhr und Vorschau lesen, die Einfuhr schreibt gegen die Marke aus
		// der Vorschau.
		{
			name: "Einstellungen ausführen", methode: http.MethodGet,
			weg: "/einstellungen/ausfuhr", erfolg: http.StatusOK,
		},
		{
			name: "Einfuhr vorschauen", methode: http.MethodPost,
			weg: "/einstellungen/vorschau", rumpf: `{"inhalt":"{}"}`, erfolg: http.StatusOK,
		},
		{
			name: "Einstellungen einführen", methode: http.MethodPost,
			weg: "/einstellungen/einfuhr", erfolg: http.StatusOK,
			rumpf: `{"inhalt":"{}","revision":"abc123"}`,
		},
		{
			name: "Fassung lesen", methode: http.MethodGet,
			weg: "/aktualisierung", erfolg: http.StatusOK,
		},
		{
			name: "Fassung laden", methode: http.MethodPost,
			weg: "/aktualisierung/laden", erfolg: http.StatusCreated,
		},
	}
}
