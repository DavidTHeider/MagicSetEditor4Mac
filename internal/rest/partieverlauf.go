package rest

// **Der Verlauf einer Partie über den Dienst** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum er gefiltert ankommt
//
// **Weil `service` ihn filtert** (`domain.RueckschauFuer`), bevor er hier
// vorbeikommt. Diese Fläche sieht den ungefilterten Strom nie — und das ist
// die Zusage aus dem Issue: „kein Replay exportiert unbemerkt fremde private
// Informationen."
//
// **Die Sicht steht in der Abfrage und nicht im Zugang.** Wer sie aus dem
// Token nähme, ließe beim Self-Play einen Sitz den Verlauf des anderen sehen —
// dort ist dasselbe Profil der Normalfall (G8,
// #1465).

import (
	"context"
	"net/http"
)

// Verlaufswerkstatt ist der Weg zum Verlauf.
type Verlaufswerkstatt interface {
	PartieVerlauf(ctx context.Context, kennung, fuer string) (
		[]Partieereignis, error)
}

// partieVerlauf gibt den Ereignisstrom, wie ein Sitz ihn sehen darf.
func (d *Dienst) partieVerlauf(schreiben http.ResponseWriter, anfrage *http.Request) {
	fuer, gut := pflichttext(
		schreiben, "fuer", anfrage.URL.Query().Get("fuer"))
	if !gut {
		return
	}

	verlauf, fehler := d.naehte.Partienverlauf.PartieVerlauf(
		anfrage.Context(), anfrage.PathValue("partieID"), fuer)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste und kein `null`** — ein Verbraucher, der darüber
	// liefe, bekäme einen Fehler statt einer leeren Anzeige.
	if verlauf == nil {
		verlauf = []Partieereignis{}
	}

	ja(schreiben, http.StatusOK, verlauf)
}
