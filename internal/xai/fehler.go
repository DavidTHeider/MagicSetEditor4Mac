package xai

// **Was ein Fehler von xAI heißt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Was hier **nicht** steht, und warum
//
// **Die Gestalt des Fehlerrumpfs.** Nachgelesen am 2026-09-12 unter
// <https://docs.x.ai/developers/debugging>: die Dokumentation nennt die
// Statuscodes und ihre Bedeutung — und **keine** Felder. Weder `error` noch
// `code` noch `message` stehen dort.
//
// **Also wird tolerant gelesen und streng gedeutet**: was an bekannten
// Feldern dasteht, reist als Grund mit; was fehlt, fehlt. Die Fehlerart kommt
// aus dem **Status**, denn der ist dokumentiert. Wer eine Feldgestalt
// annähme, hätte sie erfunden (§16, Regel 1).
//
// # Zwei Statuscodes, die die anderen nicht haben
//
// **415 und 422.** Der eine heißt „falscher Inhaltstyp", der andere „ein Feld
// im Rumpf hat ein ungültiges Format" — beide sind Eingabefehler und keine
// Erlaubnisfragen.

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// fehlerhuelle ist, was ein Fehlerrumpf **vielleicht** trägt.
//
// **Jedes Feld wahlfrei.** Die Dokumentation sagt nichts über ihre Gestalt;
// was ankommt, wird gelesen, und was fehlt, wird nicht erfunden.
type fehlerhuelle struct {
	// Meldung ist `error` als Zeichenkette — die Gestalt, die xAI-Antworten
	// in der Praxis tragen.
	Meldung string `json:"error"`

	// Code ist eine Kennzeichnung, wo eine dasteht.
	Code string `json:"code"`
}

// fehlerAus macht aus einer abgelehnten Antwort einen Fehler dieses Programms.
func fehlerAus(antwort netz.Antwort, weg string) error {
	var huelle fehlerhuelle

	// **Ein unlesbarer Rumpf ist kein Grund, den Status zu verschweigen** —
	// und hier ist er sogar der Normalfall: die Gestalt steht nirgends.
	_ = json.Unmarshal(antwort.Rumpf, &huelle) //nolint:errcheck // siehe oben

	grund := fmt.Errorf("%s: %s", weg, meldungAus(antwort, huelle))

	return domain.MitGrund(artFuer(antwort.Status), grund,
		"xAI hat die Anfrage abgelehnt (%d)", antwort.Status)
}

// meldungAus nimmt, was dasteht — und sonst den rohen Rumpf.
//
// **Der rohe Rumpf ist hier die ehrlichere Auskunft**: eine leere Meldung
// verschwiege, dass der Dienst etwas gesagt hat, das dieses Programm nicht
// einordnen kann.
func meldungAus(antwort netz.Antwort, huelle fehlerhuelle) string {
	if huelle.Meldung != "" {
		if huelle.Code != "" {
			return huelle.Meldung + " (" + huelle.Code + ")"
		}

		return huelle.Meldung
	}

	if len(antwort.Rumpf) == 0 {
		return "ohne Rumpf"
	}

	return string(antwort.Rumpf)
}

// artFuer ordnet einen Status der Fehlerart dieses Programms zu.
//
// **Aus dem Status und nicht aus dem Rumpf** — er ist das Einzige, was
// dokumentiert ist.
func artFuer(status int) error {
	switch status {
	case netz.StatusNichtGefunden:
		return domain.ErrNotFound

	case netz.StatusOhneZugang, netz.StatusVerboten:
		// **Erlaubnis und nicht Validierung**: 401 heißt „kein oder ein
		// ungültiger Zugang", 403 „dein Team lässt dich nicht".
		return domain.ErrErlaubnis

	case netz.StatusGedrosselt:
		// **Ein Konflikt und keine Erlaubnisfrage.** Anders als bei
		// Anthropic und OpenAI kennt xAI hier **kein** zweites Ding: die
		// Dokumentation nennt für 429 genau einen Grund — zu schnell
		// gefragt —, und der geht vorüber.
		return domain.ErrConflict

	case netz.StatusWiderspruch:
		return domain.ErrConflict
	}

	return domain.ErrValidation
}
