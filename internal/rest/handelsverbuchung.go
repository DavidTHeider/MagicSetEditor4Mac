package rest

// **Die Verbuchung im Dienst** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum ein Schritt eine Ressource ist und kein Verb im Pfad
//
// **Weil der Handel die Ressource ist und der Schritt sein Zustand.** `POST`
// auf `/handel/{vorgang}/schritt` trägt, **welcher** Schritt gemeint ist, im
// Rumpf; ein Pfad je Schritt machte aus einer Zustandsmaschine drei
// Endpunkte, die nichts voneinander wissen.
//
// **Und das Journal ist ein `GET`** — es entsteht nichts, es wird gelesen.

import (
	"context"
	"net/http"
)

// Verbuchungsnaht ist, was der Dienst für die Verbuchung braucht.
type Verbuchungsnaht interface {
	Handelsschritt(ctx context.Context, vorgang string, wunsch Schrittwunsch) (
		Handelsstand, error)
	Handelsjournal(ctx context.Context, vorgang string) ([]Journalzeile, error)

	// Handelsbeleg gibt Stand und Beleg, **ohne** einen Schritt zu gehen
	// (Hüllenprobe `nat04handel`, #1672).
	Handelsbeleg(ctx context.Context, vorgang string) (Handelsstand, error)
}

// Handelsstand ist der Stand eines Handels samt seinem Beleg.
type Handelsstand struct {
	Vorgang string `json:"vorgang"`
	Zustand string `json:"zustand"`

	// Abdruck, Unterschrift und Gebucht stehen nur mit einem Beleg da.
	Abdruck      string `json:"abdruck,omitempty"`
	Unterschrift string `json:"unterschrift,omitempty"`
	Beidseitig   bool   `json:"beidseitig"`
	Gebucht      string `json:"gebucht,omitempty"`

	// Gegenzeichner und Gegengezeichnet sagen, **wessen** geprüfter Beleg
	// hier liegt und seit wann (#1650) — leer, solange keiner da ist.
	Gegenzeichner   string `json:"gegenzeichner,omitempty"`
	Gegengezeichnet string `json:"gegengezeichnet,omitempty"`

	// Befund ist der jüngste abgewiesene Gegenbeleg (#2500) — fehlt, solange
	// keiner abgewiesen wurde.
	Befund *Gegenbelegbefund `json:"befund,omitempty"`
}

// Gegenbelegbefund sagt, dass ein Gegenbeleg nicht passte — `mangel` ist
// eines aus `unterschrift`, `vorgang`, `revision`, `parteien`, `zeilen`.
type Gegenbelegbefund struct {
	Von       string `json:"von"`
	Abdruck   string `json:"abdruck"`
	Mangel    string `json:"mangel"`
	Grund     string `json:"grund"`
	Empfangen string `json:"empfangen"`
}

// Journalzeile ist eine Bewegung im eigenen Bestand.
type Journalzeile struct {
	ID       int64  `json:"id"`
	Art      string `json:"art"`
	Anlass   string `json:"anlass"`
	Vorgang  string `json:"vorgang,omitempty"`
	Bezug    int64  `json:"bezug,omitempty"`
	Grund    string `json:"grund,omitempty"`
	Printing string `json:"printing"`

	// Bezeichnung ist, wie die Karte heißt — beim Lesen nachgeschlagen und
	// **nicht** im Journal gespeichert.
	Bezeichnung string `json:"bezeichnung"`
	Menge       int    `json:"menge"`
	Zustand     string `json:"zustand"`
	Sprache     string `json:"sprache"`
	Finish      string `json:"finish"`

	// ExemplarID, Lagerort und Notiz sind **lokal** und stehen nicht im
	// Beleg, der hinausgeht — hier stehen sie, weil dieser Dienst der eigene
	// ist und nicht der des Gegenübers.
	ExemplarID int64  `json:"exemplarId,omitempty"`
	Lagerort   string `json:"lagerort,omitempty"`
	Notiz      string `json:"notiz,omitempty"`

	Gebucht string `json:"gebucht"`
}

// Schrittwunsch ist, was im Rumpf eines Schritts steht.
type Schrittwunsch struct {
	// Schritt ist `verbuchen`, `abgleichen` oder `bestaetigen`.
	Schritt string `json:"schritt"`

	// Wer ist die Partei, deren Stimme ankommt — nur beim Bestätigen.
	Wer string `json:"wer"`

	// Freigabe ist der Nachweis, den ein Mensch am Fenster erteilt hat (T10,
	// #1475).
	//
	// **Ausgestellt wird sie hier nicht.** `Nachweise.Ausstellen` nimmt nur
	// den Modus `fenster`; wer über einen Port anfragt, sitzt nicht vor dem
	// Bildschirm, an dem jemand zustimmt. Was dieser Weg kann, ist eine
	// **vorher** erteilte, eng begrenzte Freigabe einlösen — für diesen
	// Vorgang, in diesem Stand, einmal.
	//
	// **Abgleichen braucht keine**: es bindet nichts.
	Freigabe string `json:"freigabe"`

	// Gesehen ist der Abdruck der Revision, die der Schritt meint (#2496):
	// ist der Vorgang inzwischen eine andere, ist das ein Konflikt. **Nur vor
	// dem Abschluss** — danach ändert sich die Revision nicht mehr, und ein
	// Abdruck dort wird abgewiesen statt still übergangen. Leer heißt: nicht
	// gefragt (§16: erweiternd).
	Gesehen string `json:"gesehen,omitempty"`
}

// handelsschrittGehen fährt einen Schritt eines Handels.
func (d *Dienst) handelsschrittGehen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Schrittwunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	stand, fehler := d.naehte.Verbuchung.Handelsschritt(anfrage.Context(),
		anfrage.PathValue("vorgang"), wunsch)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// handelsbelegZeigen gibt Stand und Beleg eines Handels.
func (d *Dienst) handelsbelegZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	stand, fehler := d.naehte.Verbuchung.Handelsbeleg(anfrage.Context(),
		anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// handelsjournalZeigen gibt die Bewegungen zu einem Vorgang.
func (d *Dienst) handelsjournalZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	zeilen, fehler := d.naehte.Verbuchung.Handelsjournal(anfrage.Context(),
		anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, zeilen)
}
