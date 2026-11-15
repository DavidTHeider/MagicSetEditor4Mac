// **Der Dateibaum eines Pakets** — FMT-02.AK-01
// (#1691).
//
// Einordnung nach §16: **erweiternd** — drei Wege kommen hinzu, und
// `GET /pakete/{paket}/dateien` nennt zusätzlich das Beiwerk.
//
//   - `POST …/dateien` legt eine Vorlage an — `name` und `inhalt` im Rumpf,
//     `201` mit dem Stand. **Nicht unter `…/dateien/{datei...}`**: der Weg
//     überschnitte sich mit `POST /pakete/{art}/{name}/zurueck`, und die
//     Wegetafel weist das beim Start ab. **`name` und nicht `datei`**: ein
//     Feld `datei` ist im Dienst ein Pfad auf der Platte
//     (`domain.IstPfadfeld`) und müsste mit `/` beginnen,
//   - `PATCH …/dateien/{datei...}` benennt um — `neu` und `stand` im Rumpf,
//   - `DELETE …/dateien/{datei...}?stand=…` entfernt in die Sicherung, wie
//     `DELETE /pakete/{art}/{name}?pfad=…` seine Bestätigung in der Abfrage
//     trägt.

package rest

import (
	"net/http"
)

func (d *Dienst) paketdateiAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Name   string `json:"name"`
		Inhalt string `json:"inhalt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	stand, fehler := d.naehte.Werkstatt.DateiAnlegen(anfrage.Context(),
		anfrage.PathValue("paket"), rumpf.Name, rumpf.Inhalt)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"stand": stand})
}

func (d *Dienst) paketdateiUmbenennen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Neu   string `json:"neu"`
		Stand string `json:"stand"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if fehler := d.naehte.Werkstatt.DateiUmbenennen(anfrage.Context(),
		anfrage.PathValue("paket"), anfrage.PathValue("datei"), rumpf.Neu, rumpf.Stand); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"datei": rumpf.Neu})
}

func (d *Dienst) paketdateiEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	datei := anfrage.PathValue("datei")

	if fehler := d.naehte.Werkstatt.DateiEntfernen(anfrage.Context(),
		anfrage.PathValue("paket"), datei, anfrage.URL.Query().Get("stand")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"entfernt": datei})
}
