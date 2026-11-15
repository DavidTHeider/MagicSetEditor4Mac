package app

// **Wie der Tisch technisch steht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wofür es das gibt
//
// **Weil „es ruckelt" keine Auskunft ist.** Das Issue verlangt Diagnostik für
// „Route/RTT/Queue/Desync/Snapshot/Framezeit **ohne Kartengeheimnisse**" — und
// der Nachsatz ist der Punkt: was hier steht, dürfte jeder sehen, auch das
// Gegenüber.
//
// # Was ausdrücklich nicht darinsteht
//
// **Kein Abdruck.** Er ist ein Hash über **Karten**; wer ihn weitergäbe, gäbe
// eine Prüfsumme weiter, gegen die sich raten lässt — das Issue nennt es beim
// Namen („Geteilter Eventhash ermöglicht keine simple Wörterbuchsuche auf
// einzelne geheime Karten"). Die **Folge** steht da: sie zählt den Strom und
// sagt nichts über seinen Inhalt. Ein Lauf hält das gegen die kodierte
// Antwort.
//
// **Keine Bildzeit.** Sie entsteht im Fenster und wird dort gemessen; eine
// Zahl, die diese Naht nennte, hätte sie von niemandem (§10). Das Fenster
// zeigt sie neben diesen Werten und sagt, woher sie kommt.
//
// # Und warum der RTT die Rückkehr ist
//
// **Weil es die eine Runde gibt, die dieses Protokoll ohnehin dreht.** Ein
// eigener Takt wäre eine zweite Wahrheit über dieselbe Leitung und kostete
// bei jedem Schlag eine Nachricht. Gemessen wird zwischen dem
// Wiederaufnahmewunsch und dem ersten, was darauf zurückkam —
// [App.Merke] schließt die Messung, weil **jeder** Rückweg dort vorbeikommt.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die beiden Rollen, die ein Programm an einem Tisch haben kann.
const (
	rolleWirt = "wirt"
	rolleGast = "gast"
)

// LeitungszeileDTO ist eine Leitung dieses Tisches.
type LeitungszeileDTO struct {
	// Sitz ist, wer an ihrem anderen Ende sitzt — beim Gast leer, denn dort
	// ist es der Wirt und kein Sitz.
	Sitz string `json:"sitz"`

	// Weg ist der Kanal: die **Route** aus dem Issue.
	Weg string `json:"weg"`

	// Gegenueber ist der Fingerabdruck — die Gestalt, die ein Mensch
	// vergleichen kann.
	Gegenueber string `json:"gegenueber"`

	// SeitSekunden ist, wie lange diese Leitung steht.
	SeitSekunden int64 `json:"seitSekunden"`

	// Wartend und Verworfen sind die **Warteschlange**.
	Wartend   int `json:"wartend"`
	Verworfen int `json:"verworfen"`
}

// PartiediagnoseAntwort sagt, wie es um die Leitungen dieses Tisches steht.
type PartiediagnoseAntwort struct {
	// Laeuft sagt, ob überhaupt ein Tisch offensteht.
	Laeuft bool `json:"laeuft"`

	// Rolle ist `wirt` oder `gast` — leer, wenn nichts läuft.
	Rolle string `json:"rolle"`

	Partie string `json:"partie"`

	// Pausiert heißt: die Leitung riss, ohne dass jemand aufgestanden ist.
	Pausiert bool `json:"pausiert"`

	// Stand ist die Folge, bei der dieses Programm im Strom steht.
	Stand int64 `json:"stand"`

	// StandBekannt trennt „bei null" von „weiß ich nicht".
	//
	// **Zwei Felder und kein `-1`.** Eine Zahl, die einen zweiten Sinn
	// trägt, wird irgendwo addiert (§8).
	StandBekannt bool `json:"standBekannt"`

	// RueckkehrMillis ist die gemessene Runde — siehe Kopf.
	RueckkehrMillis int64 `json:"rueckkehrMillis"`

	// RueckkehrBekannt sagt, ob überhaupt gemessen wurde. „Fehlende
	// Messgrößen erscheinen unbekannt" steht so in der Abnahme des Issues.
	RueckkehrBekannt bool `json:"rueckkehrBekannt"`

	Leitungen []LeitungszeileDTO `json:"leitungen"`

	Fehler *Fehler `json:"fehler"`
}

// Partiediagnose sagt, wie es um die Leitungen steht.
func (a *App) Partiediagnose(ctx context.Context) PartiediagnoseAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return PartiediagnoseAntwort{
			Leitungen: []LeitungszeileDTO{},
			Fehler:    Uebersetze(ctx, fehler),
		}
	}

	// **Der eigene Tisch zuerst.** Wer freigegeben hat, ist Wirt — auch wenn
	// er daneben an einem fremden Tisch säße; die Frage „wie steht meine
	// Leitung" hat dann zwei Antworten, und die des eigenen Tisches ist die,
	// an der andere hängen.
	if antwort, wirt := a.wirtsdiagnose(); wirt {
		return antwort
	}

	return a.gastdiagnose()
}

// wirtsdiagnose gibt die Sicht des Wirts — und ob er einer ist.
func (a *App) wirtsdiagnose() (PartiediagnoseAntwort, bool) {
	offen, _ := a.offeneFreigabe()
	if offen == nil {
		return PartiediagnoseAntwort{}, false
	}

	return PartiediagnoseAntwort{
		Laeuft: true, Rolle: rolleWirt, Partie: offen.Partie(),
		Leitungen: leitungszeilen(a.partie.runde.Leitungen()),
	}, true
}

// gastdiagnose gibt die Sicht des Gastes — oder eine leere Antwort.
func (a *App) gastdiagnose() PartiediagnoseAntwort {
	griff, partie, _ := a.beitritt()

	antwort := PartiediagnoseAntwort{Leitungen: []LeitungszeileDTO{}}

	if partie == "" {
		return antwort
	}

	antwort.Laeuft, antwort.Rolle, antwort.Partie = true, rolleGast, partie
	antwort.Pausiert = a.istPausiert()

	folge, _ := a.Stand()
	antwort.Stand, antwort.StandBekannt = folge, true

	if dauer, gemessen := a.letzteRueckkehr(); gemessen {
		antwort.RueckkehrMillis = dauer.Milliseconds()
		antwort.RueckkehrBekannt = true
	}

	// **Ein pausierter Gast hat keine Leitung mehr**, und `griff` ist dann
	// null. Sein Platz bleibt trotzdem — genau das trennt „pausiert" von
	// „aufgestanden".
	if griff != nil {
		antwort.Leitungen = leitungszeilen(
			[]service.Sitzungsstand{griff.Stand()})
	}

	return antwort
}

// leitungszeilen bildet die Stände ab.
func leitungszeilen(staende []service.Sitzungsstand) []LeitungszeileDTO {
	zeilen := make([]LeitungszeileDTO, 0, len(staende))

	for _, stand := range staende {
		seit := int64(0)
		if !stand.Seit.IsZero() {
			seit = int64(time.Since(stand.Seit).Seconds())
		}

		zeilen = append(zeilen, LeitungszeileDTO{
			Sitz:         stand.Kennung,
			Weg:          stand.Kanal,
			Gegenueber:   stand.Fingerabdruck,
			SeitSekunden: seit,
			Wartend:      stand.Wartend,
			Verworfen:    stand.Verworfen,
		})
	}

	return zeilen
}
