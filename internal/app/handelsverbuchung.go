package app

// **Die Verbuchung eines Handels am Fenster** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum drei Rufe und nicht einer
//
// **Weil drei Dinge geschehen, und zwei davon warten auf einen Menschen.**
// Verbuchen bewegt den Bestand, Abgleichen meldet der Gegenseite, dass hier
// fertig ist, Bestätigen nimmt ihre Stimme entgegen. Ein Ruf, der alles täte,
// müsste die zweite Unterschrift erfinden — und genau das verbietet dieses
// Werkstück.
//
// # Und warum der Beleg mit zurückkommt
//
// **Weil der Zustand allein nicht sagt, was unterschrieben wurde.**
// „lokalVerbucht" ist eine Marke; der Beleg ist das Papier dazu.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// HandelsbelegDTO ist der abgelegte Beleg eines Abschlusses.
type HandelsbelegDTO struct {
	Vorgang string `json:"vorgang"`

	// Abdruck ist der Stand, der abgeschlossen wurde.
	Abdruck string `json:"abdruck"`

	Unterschrift string `json:"unterschrift"`

	// Beidseitig sagt, ob auch die Gegenseite unterschrieben hat.
	Beidseitig bool `json:"beidseitig"`

	// Gegenzeichner ist die Peerkennung dessen, dessen geprüfter Beleg hier
	// liegt — **leer**, solange keiner da ist.
	Gegenzeichner string `json:"gegenzeichner"`

	// Gegengezeichnet ist der Augenblick, in dem er ankam — **leer**,
	// solange keiner da ist.
	Gegengezeichnet string `json:"gegengezeichnet"`

	Gebucht string `json:"gebucht"`

	// Befund ist der jüngste abgewiesene Gegenbeleg — `null`, solange
	// keiner abgewiesen wurde (#2500). **Erweiternd** an MCP (0033, Regel 6).
	Befund *GegenbelegbefundDTO `json:"befund"`
}

// GegenbelegbefundDTO sagt, dass ein Gegenbeleg nicht passte.
type GegenbelegbefundDTO struct {
	// Von ist die Peerkennung des Absenders.
	Von string `json:"von"`

	// Abdruck ist der, den der abgewiesene Beleg nannte.
	Abdruck string `json:"abdruck"`

	// Mangel ist eines aus `domain.DieGegenbelegmaengel` — das Fenster
	// übersetzt es.
	Mangel string `json:"mangel"`

	// Grund ist der Satz der Prüfung, deutsch — für Protokoll und Skript.
	Grund string `json:"grund"`

	Empfangen string `json:"empfangen"`
}

// uhrzeitOderLeer gibt die Uhr als Text — **leer**, wenn es keine gibt.
//
// **Und nicht `0001-01-01T00:00:00Z`**: eine Nullzeit sieht aus wie eine
// Angabe, und eine Fläche, die sie anzeigte, behauptete einen Augenblick im
// Jahr eins.
func uhrzeitOderLeer(wann time.Time) string {
	if wann.IsZero() {
		return ""
	}

	return wann.UTC().Format(time.RFC3339)
}

// JournalzeileDTO ist eine Bewegung im eigenen Bestand.
//
// **Mit Lagerort und Besitzzeile** — das Journal ist das eigene Buch und
// nicht der Beleg, der hinausgeht.
type JournalzeileDTO struct {
	ID int64 `json:"id"`

	// Art ist `zugang` oder `abgang`.
	Art string `json:"art"`

	// Anlass ist, warum gebucht wurde.
	Anlass string `json:"anlass"`

	Vorgang string `json:"vorgang"`

	// Bezug und Grund stehen an einer Korrektur und sonst nirgends.
	Bezug int64  `json:"bezug"`
	Grund string `json:"grund"`

	Printing string `json:"printing"`

	// Bezeichnung ist, wie die Karte heißt — beim Lesen nachgeschlagen.
	//
	// **Nicht im Journal gespeichert**: dort steht der öffentliche
	// Ausschnitt, und ein eingefrorener Name wäre falsch, sobald jemand den
	// Druck berichtigt.
	Bezeichnung string `json:"bezeichnung"`

	Menge   int    `json:"menge"`
	Zustand string `json:"zustand"`
	Sprache string `json:"sprache"`
	Finish  string `json:"finish"`

	ExemplarID int64  `json:"exemplarId"`
	Lagerort   string `json:"lagerort"`
	Notiz      string `json:"notiz"`

	Gebucht string `json:"gebucht"`
}

// VerbuchungAntwort ist der Stand eines Handels samt seinem Beleg.
type VerbuchungAntwort struct {
	Vorgang string `json:"vorgang"`

	// Zustand ist der Zustand des Handels nach diesem Ruf.
	Zustand string `json:"zustand"`

	// Beleg gilt nur mit HatBeleg.
	Beleg    HandelsbelegDTO `json:"beleg"`
	HatBeleg bool            `json:"hatBeleg"`

	Fehler *Fehler `json:"fehler"`
}

// JournalAntwort sind die Bewegungen zu einem Vorgang.
type JournalAntwort struct {
	Vorgang string            `json:"vorgang"`
	Zeilen  []JournalzeileDTO `json:"zeilen"`
	Fehler  *Fehler           `json:"fehler"`
}

// Handelsverbuchung bucht einen übergebenen Handel im eigenen Bestand.
//
// **Ein zweiter Ruf bewegt nichts** — der Beleg wird vorher nachgeschlagen.
// **`nachweis` ist die Freigabe eines Menschen** — sie kommt aus
// [App.HandelFreigabeErteilen] und gilt für **diesen** Vorgang in **diesem**
// Stand (T10, #1475).
func (a *App) Handelsverbuchung(
	ctx context.Context, vorgang, nachweis string,
) VerbuchungAntwort {
	ctx = a.imProfil(ctx)

	angebot, fehler := a.verbuchungsstelle().Verbuchen(ctx, vorgang, nachweis)

	return a.nachVerbuchung(ctx, vorgang, angebot, fehler)
}

// Handelsabgleich meldet, dass hier fertig ist und die Gegenseite fehlt.
func (a *App) Handelsabgleich(
	ctx context.Context, vorgang string,
) VerbuchungAntwort {
	ctx = a.imProfil(ctx)

	angebot, fehler := a.verbuchungsstelle().Abgleichen(ctx, vorgang)

	return a.nachVerbuchung(ctx, vorgang, angebot, fehler)
}

// Handelsbestaetigung nimmt eine Stimme für den beidseitigen Abschluss.
//
// **`wer` ist die Partei, deren Stimme ankommt.** Erst wenn alle gestimmt
// haben, wird der Beleg beidseitig.
// **Und auch sie braucht die Freigabe eines Menschen** (T10, #1475): `wer`
// ist frei gewählt, und ohne Riegel ließe sich damit die Stimme des
// Gegenübers einsetzen.
func (a *App) Handelsbestaetigung(
	ctx context.Context, vorgang, wer, nachweis string,
) VerbuchungAntwort {
	ctx = a.imProfil(ctx)

	angebot, fehler := a.verbuchungsstelle().Bestaetigen(
		ctx, vorgang, wer, nachweis)

	return a.nachVerbuchung(ctx, vorgang, angebot, fehler)
}

// Handelsbeleg gibt Stand und Beleg eines Handels — **rein lesend**.
//
// **Gefunden mit zwei Fenstern** (Hüllenprobe `nat04handel`, #1672): wer
// unterschrieben hat, sagte jede Fläche nur als Antwort auf den eigenen
// Schritt. Ein Gegenbeleg, der danach über die Leitung kam, blieb
// unsichtbar, und das Fenster sagte weiter „Nur du hast unterschrieben".
func (a *App) Handelsbeleg(ctx context.Context, vorgang string) VerbuchungAntwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.handelsleitung().Stand(ctx, vorgang)

	return a.nachVerbuchung(ctx, vorgang, stand, fehler)
}

// Handelsjournal gibt die Bewegungen zu einem Vorgang.
//
// **Ein leerer Vorgang gibt, was ohne Handel gebucht wurde** — Kauf, Verkauf,
// Geschenk, Einfuhr und Korrektur.
func (a *App) Handelsjournal(
	ctx context.Context, vorgang string,
) JournalAntwort {
	ctx = a.imProfil(ctx)

	eintraege, fehler := a.verbuchungsstelle().Journal(ctx, vorgang)
	if fehler != nil {
		return JournalAntwort{
			Vorgang: vorgang, Zeilen: []JournalzeileDTO{},
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	return JournalAntwort{Vorgang: vorgang, Zeilen: alsJournalzeilen(eintraege)}
}

// nachVerbuchung baut die Antwort — samt dem Beleg, der dazu dasteht.
func (a *App) nachVerbuchung(
	ctx context.Context, vorgang string,
	angebot domain.Handelsangebot, fehler error,
) VerbuchungAntwort {
	if fehler != nil {
		return VerbuchungAntwort{
			Vorgang: vorgang, Fehler: Uebersetze(ctx, fehler),
		}
	}

	antwort := VerbuchungAntwort{
		Vorgang: vorgang, Zustand: string(angebot.Zustand),
	}

	// **Der Beleg wird gelesen und nicht gefolgert.** Aus „lokalVerbucht"
	// folgt, dass einer dastehen **sollte** — was dasteht, sagt der Bestand.
	beleg, lesefehler := a.abschlussablage().Beleg(ctx, vorgang)
	if lesefehler != nil {
		return VerbuchungAntwort{
			Vorgang: vorgang, Zustand: antwort.Zustand,
			Fehler: Uebersetze(ctx, lesefehler),
		}
	}

	if beleg != nil {
		antwort.Beleg, antwort.HatBeleg = alsHandelsbelegDTO(*beleg), true
	}

	return antwort
}

// alsHandelsbelegDTO macht aus dem Beleg seine Gestalt für das Fenster.
func alsHandelsbelegDTO(beleg domain.Belegstand) HandelsbelegDTO {
	return HandelsbelegDTO{
		Vorgang: beleg.Vorgang, Abdruck: beleg.Abdruck,
		Unterschrift: beleg.Unterschrift, Beidseitig: beleg.Beidseitig,
		Gegenzeichner:   beleg.Gegenzeichner,
		Gegengezeichnet: uhrzeitOderLeer(beleg.Gegengezeichnet),
		Gebucht:         beleg.Gebucht.UTC().Format(time.RFC3339),
		Befund:          alsBefundDTO(beleg.Befund),
	}
}

// alsBefundDTO gibt den Befund für das Fenster — `nil` ohne einen.
func alsBefundDTO(befund *domain.Gegenbelegbefund) *GegenbelegbefundDTO {
	if befund == nil {
		return nil
	}

	return &GegenbelegbefundDTO{
		Von: befund.Von, Abdruck: befund.Abdruck, Mangel: string(befund.Mangel),
		Grund: befund.Grund, Empfangen: uhrzeitOderLeer(befund.Empfangen),
	}
}

// alsJournalzeilen macht aus den Bewegungen ihre Gestalt für das Fenster.
func alsJournalzeilen(eintraege []service.Journaleintrag) []JournalzeileDTO {
	dtos := make([]JournalzeileDTO, 0, len(eintraege))

	for _, eintrag := range eintraege {
		zeile := eintrag.Zeile

		dtos = append(dtos, JournalzeileDTO{
			Bezeichnung: eintrag.Bezeichnung,
			ID:          zeile.ID, Art: string(zeile.Art),
			Anlass: string(zeile.Anlass), Vorgang: zeile.Vorgang,
			Bezug: zeile.Bezug, Grund: zeile.Grund,
			Printing: zeile.Printing, Menge: zeile.Menge,
			Zustand: string(zeile.Zustand), Sprache: zeile.Sprache,
			Finish: string(zeile.Finish), ExemplarID: zeile.ExemplarID,
			Lagerort: zeile.Lagerort, Notiz: zeile.Notiz,
			Gebucht: zeile.Gebucht.UTC().Format(time.RFC3339),
		})
	}

	return dtos
}
