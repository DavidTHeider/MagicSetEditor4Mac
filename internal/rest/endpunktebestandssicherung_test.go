package rest_test

// **Die Wege der Bestandssicherung** — aus `endpunktliste_test.go`
// herausgezogen, als OPS01 (#2347) den Sicherungsstand dazubrachte und die
// Datei über die Zeilengrenze ging (§2).

import "net/http"

// endpunkteBestandssicherung sind die Wege an `/bestand`.
func endpunkteBestandssicherung() []endpunkt {
	return []endpunkt{
		// **Die Sicherung des ganzen Bestands**
		// (#1169).
		// Sie hängt an keinem Set — deshalb `/bestand` und nicht `/sets/{id}/…`.
		//
		// **Zurückgespielt wird hier nicht**, und der Grund steht in
		// `bestandssicherung.go`: der Dienst hält eine Verbindung über seine
		// Laufzeit, und alle seine Nähte haben sie sich abgeholt.
		{
			name: "Bestand sichern", methode: http.MethodPost, weg: "/bestand/sicherung",
			rumpf: `{"ziel":"/tmp/x.msesicherung"}`, erfolg: http.StatusCreated,
		},
		{
			name: "in eine Sicherung sehen", methode: http.MethodGet,
			weg: "/bestand/sicherung?pfad=/tmp/x.msesicherung", erfolg: http.StatusOK,
		},
		// **Nur lesend** (#1351, Werkstück 14): das Umschalten tauscht die
		// Bestandsdatei, und der Grund gegen einen Schaltweg hier ist wörtlich
		// der von oben — der Dienst hält eine Verbindung über seine Laufzeit.
		{
			name: "die Verschlüsselung ansehen", methode: http.MethodGet,
			weg: "/bestand/verschluesselung", erfolg: http.StatusOK,
		},
		// **Wann zuletzt gesichert wurde** (OPS01, #2347).
		{
			name: "das Alter der Sicherung ansehen", methode: http.MethodGet,
			weg: "/bestand/sicherungsstand", erfolg: http.StatusOK,
		},
	}
}
