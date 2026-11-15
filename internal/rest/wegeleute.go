// **Die Wege zu Nutzern und Profilen** — abgetrennt aus `wegetafel.go` am
// 2026-09-08, als `GET /kontext` (Werkstück O1 von Phase 5.2, #1425) die
// Tafel über 400 Zeilen brachte (§2).
//
// **Dieselbe Tafel, zwei Dateien.** Was hier steht, gehört zur Wegetafel und
// wird von denselben Wächtern gelesen — `vollstaendig_test.go` hier und
// `restspalte_test.go` in `internal/werkzeuge` lesen beide Dateien.

package rest

import "net/http"

// leutewegeEintragen trägt die Wege ein, die an Nutzern und Profilen hängen.
func (d *Dienst) leutewegeEintragen(weg *http.ServeMux) {
	// **Nutzer und Profile** (#1351). Das Profil hängt am Nutzer, und der Pfad
	// sagt es.
	weg.HandleFunc("GET /nutzer", d.nutzer)
	weg.HandleFunc("POST /nutzer", d.nutzerAnlegen)
	weg.HandleFunc("POST /nutzer/{nutzerID}/profile", d.profilAnlegen)

	// **`GET /profil` und `PUT /profil`** — Werkstück 3 (#1351).
	//
	// **Einzahl und ohne Kennung im Weg**: es gibt genau eines, in dem
	// gearbeitet wird. `PUT /profil/{id}` läse sich wie „ändere dieses
	// Profil"; gemeint ist „arbeite ab jetzt in diesem".
	weg.HandleFunc("GET /profil", d.gezeigtesProfil)
	weg.HandleFunc("PUT /profil", d.profilWechseln)

	// **`GET /werarbeitet` und `PUT /nutzer`** — Werkstück 17 (#1400).
	//
	// **Getrennt von `/profil`**, weil es zwei Fragen sind: *wer* arbeitet und
	// *worin*. Wer in ein fremdes Profil gewechselt ist, bleibt, wer er ist.
	weg.HandleFunc("GET /werarbeitet", d.werArbeitet)

	// **`GET /kontext`** — Werkstück O1 von Phase 5.2 (#1425).
	weg.HandleFunc("GET /kontext", d.arbeitskontext)
	weg.HandleFunc("PUT /nutzer", d.nutzerWechseln)

	// **Die Anbieterkonten hängen am Nutzer**, und der Pfad sagt es —
	// Werkstück 4 (#1351).
	weg.HandleFunc("GET /nutzer/{nutzerID}/konten", d.anbieterkonten)
	weg.HandleFunc("DELETE /nutzer/{nutzerID}/konten/{anbieter}", d.kontoLoesen)

	// **Ändern und Löschen** — Werkstück 5 (#1351). **Der Name bestätigt**,
	// und er steht in der Abfrage: ein `DELETE` hat keinen Rumpf (#1218).
	weg.HandleFunc("PUT /nutzer/{nutzerID}", d.nutzerAendern)
	weg.HandleFunc("DELETE /nutzer/{nutzerID}", d.nutzerLoeschen)
	weg.HandleFunc("PUT /nutzer/{nutzerID}/bild", d.nutzerbildSetzen)
	weg.HandleFunc("DELETE /nutzer/{nutzerID}/bild", d.nutzerbildEntfernen)
	weg.HandleFunc("GET /nutzer/{nutzerID}/einstellungen", d.nutzereinstellungen)
	weg.HandleFunc("PUT /nutzer/{nutzerID}/einstellungen/{schluessel}",
		d.nutzereinstellungSetzen)
	weg.HandleFunc("GET /nutzer/{nutzerID}/personendaten", d.personendaten)
	weg.HandleFunc("PUT /profile/{profilID}/badge", d.profilbadgeSetzen)
	weg.HandleFunc("DELETE /profile/{profilID}/badge", d.profilbadgeEntfernen)
	// **Die Zugriffstoken eines Profils** — Werkstück 10 (#1351), Punkt 13.
	// **Unter dem Profil und nicht unter dem Nutzer**: ein Token gehört einer
	// Arbeitsumgebung, nicht einem Menschen.
	weg.HandleFunc("GET /profile/{profilID}/token", d.token)
	weg.HandleFunc("POST /profile/{profilID}/token", d.tokenAnlegen)
	weg.HandleFunc("PUT /token/{tokenID}", d.tokenAendern)
	weg.HandleFunc("DELETE /token/{tokenID}", d.tokenLoeschen)

	// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von
	// Phase 5.2 (#1426). **`/token/ablaufwahlen` vor `/token/{tokenID}`**
	// ist keine Frage der Reihenfolge: ein fester Abschnitt schlägt im Muster
	// den Platzhalter, und die Methoden sind ohnehin verschieden.
	weg.HandleFunc("POST /token/{tokenID}/rotation", d.tokenRotieren)
	weg.HandleFunc("DELETE /profile/{profilID}/token", d.tokenAlleWiderrufen)
	weg.HandleFunc("GET /token/ablaufwahlen", d.tokenAblaufwahlen)

	// **Die Vorgaben eines Profils** — Werkstück 15 (#1351), Punkte 4 bis 7.
	// **`PUT` auf die Art**: sie ist der Ort, und derselbe Ruf zweimal führt
	// zum selben Zustand.
	weg.HandleFunc("GET /profile/{profilID}/vorgaben", d.vorgaben)
	weg.HandleFunc("PUT /profile/{profilID}/vorgaben/{art}", d.vorgabeSetzen)

	// **Die Freigaben eines Profils** — Werkstück 11 (#1351). **`PUT` auf den
	// Nutzer**: er ist der Ort, und derselbe Ruf zweimal führt zum selben
	// Zustand.
}
