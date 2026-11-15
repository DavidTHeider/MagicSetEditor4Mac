// **Die Wege für Turniere** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, weil `wegetafel.go` mit DOM-04
// (#1699)
// über die Zeilengrenze aus §2 ginge — dieselbe Form wie `peerwege`.

package rest

import "net/http"

// turnierwege trägt die Turnierwege ein.
//
// **Der Vorschlag ist ein `GET`**, weil er nichts anlegt — das ist der ganze
// Unterschied zu `POST /runden`.
func (d *Dienst) turnierwege(weg *http.ServeMux) {
	weg.HandleFunc("GET /turniere", d.turniere)
	weg.HandleFunc("POST /turniere", d.turnierAnlegen)
	weg.HandleFunc("GET /turniere/{turnierID}", d.turnier)
	weg.HandleFunc("DELETE /turniere/{turnierID}", d.turnierLoeschen)
	weg.HandleFunc("POST /turniere/{turnierID}/teilnehmer", d.turnierAnmelden)
	weg.HandleFunc("PUT /turniere/{turnierID}/teilnehmer/{teilnehmerID}",
		d.turnierAussteigen)
	weg.HandleFunc("GET /turniere/{turnierID}/rangliste", d.turnierRangliste)
	weg.HandleFunc("GET /turniere/{turnierID}/vorschlag", d.turnierVorschlag)
	weg.HandleFunc("POST /turniere/{turnierID}/runden", d.turnierRunde)
	weg.HandleFunc("PUT /turniere/{turnierID}/paarungen/{paarungID}",
		d.turnierErgebnis)

	// **Eine Vorlage** — DOM-04.AK-01 und AK-03 (#1699): ein `POST`, weil
	// ein neues Turnier entsteht.
	weg.HandleFunc("POST /turniere/{turnierID}/kopie", d.turnierKopieren)

	// **Abschluss und Protokoll** — DOM-04.AK-04 (#1699). Der Abschluss ist
	// eine Ressource des Turniers: `POST` setzt ihn, `DELETE` nimmt ihn
	// zurück.
	weg.HandleFunc("POST /turniere/{turnierID}/abschluss", d.turnierAbschliessen)
	weg.HandleFunc("DELETE /turniere/{turnierID}/abschluss", d.turnierWiedereroeffnen)
	weg.HandleFunc("GET /turniere/{turnierID}/protokoll", d.turnierProtokoll)

	// **Die Folgen einer Berichtigung** — DOM-04.AK-02 (#1699): ein `POST`,
	// weil der Stand in den Rumpf gehört; es entsteht nichts.
	weg.HandleFunc("POST /turniere/{turnierID}/paarungen/{paarungID}/folgen",
		d.turnierFolgen)

	// **Datei und Bericht** — DOM-04.AK-03 (#1699). Das Einlesen ist ein
	// `POST` auf die Sammlung, weil ein neues Turnier entsteht.
	weg.HandleFunc("GET /turniere/{turnierID}/ausgabe", d.turnierAusgabe)
	weg.HandleFunc("GET /turniere/{turnierID}/bericht", d.turnierBericht)
	weg.HandleFunc("POST /turniere/einlesen", d.turnierEinlesen)
}
