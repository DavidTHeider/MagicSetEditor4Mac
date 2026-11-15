package app

// **Hintergrundaufgaben aus einer Naht** — Werkstück U11 von Phase 5.1
// (#1422).
//
// Drei Wege laufen länger als ein Klick: Preise für die Sammlung
// (`MassePreise`), die Sammeldaten (`ScryfallSammeldatenHolen`) und eine
// neue Fassung (`NeueFassungLaden`). Jeder trägt sich beim Beginn ins
// [service.Auftragsbuch] ein und beim Ende wieder aus — und `Auftraege`
// liest das Buch für die Übersicht. **Nur am Fenster**: ein Auftrag lebt im
// Prozess, der ihn angestoßen hat; eine Kommandozeile ist ihr eigener.

import (
	"context"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// AuftragDTO ist ein Auftrag, wie das Fenster ihn sieht.
type AuftragDTO struct {
	ID          string `json:"id"`
	Art         string `json:"art"`
	Was         string `json:"was"`
	Fortschritt string `json:"fortschritt"`
	Stand       string `json:"stand"`
	Grund       string `json:"grund"`
	Begonnen    string `json:"begonnen"`
	// Sekunden ist die Dauer bis jetzt oder bis zum Ende.
	Sekunden int `json:"sekunden"`

	// Erledigt, Gesamt und Einheit sind der gezählte Stand eines Datei-Laufs
	// (FMT-03) — das Fenster setzt ihn in seiner Sprache; `gesamt` 0 heißt
	// „nicht bekannt".
	Erledigt int    `json:"erledigt"`
	Gesamt   int    `json:"gesamt"`
	Einheit  string `json:"einheit"`

	// Abbrechbar sagt, ob der laufende Auftrag sich anhalten lässt.
	Abbrechbar bool `json:"abbrechbar"`
}

// AuftraegeAntwort ist das Buch samt Umschlag.
type AuftraegeAntwort struct {
	Auftraege []AuftragDTO `json:"auftraege"`
	Laufend   int          `json:"laufend"`
	Fehler    *Fehler      `json:"fehler"`
}

// Auftraege gibt die laufenden und die zuletzt beendeten Aufträge.
func (a *App) Auftraege(ctx context.Context) AuftraegeAntwort {
	ctx = a.imProfil(ctx)
	_ = ctx

	antwort := AuftraegeAntwort{Auftraege: []AuftragDTO{}}
	jetzt := a.uhr()

	for _, auftrag := range a.auftragsbuch().Auftraege() {
		if auftrag.Stand == domain.AuftragLaeuft {
			antwort.Laufend++
		}

		antwort.Auftraege = append(antwort.Auftraege, auftragAlsDTO(auftrag, jetzt))
	}

	return antwort
}

// auftragAlsDTO übersetzt einen Auftrag für das Fenster.
func auftragAlsDTO(auftrag domain.Auftrag, jetzt time.Time) AuftragDTO {
	return AuftragDTO{
		ID: strconv.FormatInt(auftrag.ID, 10), Art: auftrag.Art, Was: auftrag.Was,
		Fortschritt: auftrag.Fortschritt, Stand: string(auftrag.Stand), Grund: auftrag.Grund,
		Begonnen: auftrag.Begonnen.UTC().Format(time.RFC3339),
		Sekunden: int(auftrag.Dauer(jetzt).Seconds()),
		Erledigt: auftrag.Gezaehlt.Erledigt, Gesamt: auftrag.Gezaehlt.Gesamt, Einheit: auftrag.Gezaehlt.Einheit,
		Abbrechbar: auftrag.Abbrechbar,
	}
}

// auftragsbuch gibt das Buch — und legt es beim ersten Ruf an.
func (a *App) auftragsbuch() *service.Auftragsbuch {
	a.auftragsriegel.Lock()
	defer a.auftragsriegel.Unlock()

	if a.auftraege == nil {
		a.auftraege = service.NeuesAuftragsbuch(a.uhr)
	}

	return a.auftraege
}

// auftragBeginnen trägt einen Auftrag ein und gibt zurück, was ihn beendet.
func (a *App) auftragBeginnen(art, was string) func(fehler error) {
	buch := a.auftragsbuch()
	id := buch.Beginnen(art, was)

	return func(fehler error) { buch.Beenden(id, fehler) }
}

// Auftraege — die Fensterbindung.
func (b *Bindungen) Auftraege(ctx context.Context) AuftraegeAntwort {
	return b.app(ctx).Auftraege(b.ruf(ctx))
}
