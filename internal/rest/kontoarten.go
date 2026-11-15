// **Die vier Kontoarten über REST** — Werkstück 6 von Phase 4.12
// (#1351).
//
// # Warum hier ein Kennwort im Rumpf stehen darf und über MCP nicht
//
// **Weil niemand mitliest.** Der Dienst lauscht auf `127.0.0.1`, der Rumpf
// geht von einem Prozess zum anderen und landet in keinem Verlauf. Über MCP
// stünde derselbe Wert im Mitschnitt eines Sprachmodells — dort fehlen
// `konto_neu` und `konto_anmelden` deshalb, und zwar begründet.
//
// **Im Pfad steht er trotzdem nie**: eine Anfragezeile landet in jedem
// Zugriffsprotokoll, ein Rumpf nicht. Deshalb `POST` mit JSON und kein `GET`
// mit Parametern.
//
// # Was das Anmelden hier ist und was nicht
//
// **Es ist eine Auskunft, kein Schlüssel.** Der Dienst kennt keine Sitzungen:
// wer ihn erreicht, erreicht auch alles andere an ihm — das ändert Werkstück
// 10 mit den technischen Profilen. Bis dahin sagt `POST /konten/anmeldung`,
// **ob** ein Kennwort stimmt, und öffnet nichts, was vorher zu war.
//
// **Was das trotzdem ist: ein Orakel.** Wer den Port erreicht, kann Kennwörter
// durchprobieren. Der Riegel dagegen ist Argon2id selbst — 19 MiB und zwei
// Durchgänge je Versuch (`internal/kennwort`).
//
// **Gemessen am 2026-09-05**: 21 ms je Ruf von `kennwort.Huete`
// (`go test -bench`, 20 Läufe, dieser Rechner). **So weit reicht die Zahl und
// nicht weiter** (§10): sie gilt für **einen** Versuch nacheinander, also rund
// 48 in der Sekunde. `p=1` heißt, dass ein Ruf **einen** Kern benutzt — wer
// zehn Anfragen nebeneinander schickt, bekommt das Zehnfache, und 19 MiB je
// Ruf sind dabei die Grenze, nicht die Rechenzeit.
//
// **Das ist kein Ersatz für eine Drossel.** Es ist der Grund, warum bis
// Werkstück 10 (technische Profile und Zugriffstoken) keine gebaut wird: der
// Dienst hat noch keinen Begriff davon, wer da fragt.

package rest

import "net/http"

// kontoEintragen legt ein Konto unter einer Adresse oder einer Rufnummer an.
func (d *Dienst) kontoEintragen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Kennung  string `json:"kennung"`
		Kennwort string `json:"kennwort"`
		Name     string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	eingetragen, fehler := d.naehte.Nutzer.KontoEintragen(
		anfrage.Context(), rumpf.Kennung, rumpf.Kennwort, rumpf.Name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"nutzer": eingetragen})
}

// kontoAnmelden sagt, wer unter dieser Kennung mit diesem Kennwort arbeitet.
//
// **`200` und nicht `201`**: hier entsteht nichts.
func (d *Dienst) kontoAnmelden(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Kennung  string `json:"kennung"`
		Kennwort string `json:"kennwort"`

		// **Der zweite Faktor** — Werkstück 7 (#1351). Leer heißt „über den
		// Sensor"; wer keinen Faktor hat, lässt das Feld weg.
		Nachweis string `json:"nachweis"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	herein, fehler := d.naehte.Nutzer.KontoAnmelden(
		anfrage.Context(), rumpf.Kennung, rumpf.Kennwort, rumpf.Nachweis)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"nutzer": herein})
}

// kontoAnonym legt ein Konto ohne jede Angabe an.
//
// **Ohne Rumpf, weil es nichts anzugeben gibt.** Ein leerer `POST` ist genau
// die Gestalt der Sache.
func (d *Dienst) kontoAnonym(schreiben http.ResponseWriter, anfrage *http.Request) {
	angelegt, fehler := d.naehte.Nutzer.KontoAnonym(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"nutzer": angelegt})
}
