package service

// **Was für einen Handel festgehalten wird** —
// #1643.
//
// # Warum ein Angebot das nicht sagen kann
//
// **Weil es einen Druck nennt und keine Zeile des eigenen Bestands.**
// [domain.Handelsposition] trägt `Printing`, `DruckID`, `Menge`, `Guete`,
// `Sprache` und `Finish`; wer drei `NM`-Exemplare desselben Drucks in drei
// Lagerorten hat, muss sagen, **welches** weggeht. Das rät kein Weg, und
// einer, der es täte, legte die falsche Karte auf den Tisch.
//
// # Und warum es das überhaupt gibt
//
// **Weil `Verbuchungsstelle.abgaengeFuer` die Abgangszeilen eines Belegs aus
// genau diesen Zusagen zieht.** Ohne sie konnte nur die **empfangende** Seite
// verbuchen: wer Karten weggab, bekam „es fehlen N Stück an bestätigter
// Übergabe" — richtig gemeldet und ohne einen Weg, es in Ordnung zu bringen.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zusagenablage ist, was das Werk vom Bestand braucht.
//
// **Die Schnittstelle steht hier und nicht in `storage`** (§2): sie
// beschreibt, was dieser Anwendungsfall verlangt.
type Zusagenablage interface {
	// Festhalten trägt eine Zusage ein — oder sagt, warum nicht.
	//
	// **`vertrag` sind die Zeilen, die dieser Handel von hier nimmt.** Er
	// gilt in **derselben** Transaktion wie das Eintragen: eine Prüfung
	// davor ist eine Vorschau und keine Grenze (Befund 1 des Nachlaufs zum
	// Zweitblick über Phase 5.6). **Die Zeilen und nicht ihre Summe**
	// (#2492): erst mit dem Exemplar zusammen sagen sie, ob es passt und wie
	// viel von seiner Variante gebunden werden darf.
	Festhalten(ctx context.Context, zusage domain.Reservierung,
		vertrag domain.Zusagevertrag) error

	// FuerVorgang gibt die Zusagen eines Handels über alle Besitzzeilen.
	FuerVorgang(ctx context.Context, vorgang string) ([]domain.Reservierung, error)

	// Uebergeben vermerkt, dass sie über den Tisch gegangen sind.
	Uebergeben(ctx context.Context, vorgang string) error

	// Freigeben nimmt sie weg — was übergeben ist, bleibt.
	Freigeben(ctx context.Context, vorgang string) error
}

// Zusagenwunsch ist, was ein Mensch benennt.
//
// **Ein Verbund und keine sechs Argumente**: fünf Werte in einer Reihe, von
// denen zwei Zahlen sind, sind eine Einladung, zwei davon zu vertauschen.
type Zusagenwunsch struct {
	Vorgang string

	// Wer ist die Partei, die weggibt — dieser Rechner.
	//
	// **Sie steht dabei, weil der Vertrag sagt, wie viel sie weggibt**
	// (Befund 1 des Zweitblicks über Phase 5.6): ohne sie prüfte nur der
	// **Bestand**, ob eine Zusage hineinpasst, und nicht der **Vertrag**.
	// Wer bei einem Handel über ein Stück zweimal festhielt, gab beim
	// Verbuchen zwei weg.
	Wer string

	ExemplarID int64
	Menge      int

	// Fassung ist die Fassung der Besitzzeile, die der Mensch **gesehen**
	// hat (§7, Regel 2). Hat inzwischen jemand geschrieben, ist die Zusage
	// ein Konflikt und kein stiller Erfolg.
	Fassung int64

	// Frist ist, bis wann die Zusage gilt — der Nullwert heißt: ohne.
	//
	// # Warum der Nullwert der Normalfall ist
	//
	// **Weil der Handel selbst die Frist ist.** Ein zweiter Wecker liefe
	// neben dem Vorgang her und ließe Exemplare frei werden, während
	// verhandelt wird; was einen Handel beendet, gibt sie ohnehin frei
	// ([Handelsleitung.Schreiten]).
	Frist time.Time
}

// Zusagenwerk hält Exemplare für einen Handel fest.
type Zusagenwerk struct {
	ablage Zusagenablage
	stand  Handelsstand

	// schranke ist die Handelsrichtlinie (#2435) — samt der Grenze, wie
	// lange eine Zusage festhalten darf.
	schranke Handelsschranke
	uhr      func() time.Time
}

// NeuesZusagenwerk baut es über der Ablage und dem Stand der Vorgänge.
func NeuesZusagenwerk(
	ablage Zusagenablage, stand Handelsstand,
) *Zusagenwerk {
	return &Zusagenwerk{ablage: ablage, stand: stand, schranke: ohneEinstellung{}, uhr: time.Now}
}

// MitHandelsschranke setzt die Handelsrichtlinie, die gilt (#2435).
func (w *Zusagenwerk) MitHandelsschranke(schranke Handelsschranke) *Zusagenwerk {
	w.schranke = schranke

	return w
}

// MitUhr setzt die Uhr, an der eine Frist gemessen wird — für den Test.
func (w *Zusagenwerk) MitUhr(uhr func() time.Time) *Zusagenwerk {
	w.uhr = uhr

	return w
}

// Festhalten hält eine benannte Besitzzeile für einen Vorgang fest.
//
// **Die Revision kommt aus dem Stand und nicht aus dem Ruf.** Eine Zusage
// gilt einer Angebotsrevision (`T2`); wer sie mitschickte, könnte eine
// nennen, die es nicht mehr gibt — und der Vorgang muss ohnehin gelesen
// werden, um zu wissen, dass es ihn gibt.
func (w *Zusagenwerk) Festhalten(
	ctx context.Context, wunsch Zusagenwunsch,
) (domain.Reservierung, error) {
	heute, fehler := w.stand.Stand(ctx, wunsch.Vorgang)
	if fehler != nil {
		return domain.Reservierung{}, fehler
	}

	// **Und die Schranke, samt ihrer Frist** (#2435): ohne Wunsch hält eine
	// Zusage so lange, wie die Grenze sagt.
	frist, fehler := zusagefrist(ctx, w.schranke, heute, w.uhr(), wunsch.Frist)
	if fehler != nil {
		return domain.Reservierung{}, fehler
	}

	zusage := domain.Reservierung{
		Vorgang: heute.Vorgang, Revision: heute.Revision,
		ExemplarID: wunsch.ExemplarID, Menge: wunsch.Menge,
		Fassung: wunsch.Fassung, Frist: frist,
	}

	// **Ob sie hineinpasst, entscheidet der Bestand** — in **einer**
	// Transaktion mit dem Eintragen ([domain.ReservierungPasst]). Eine
	// Prüfung hier davor wäre eine zweite Wahrheit und ein Zeitfenster.
	if fehler := w.ablage.Festhalten(
		ctx, zusage, vonHier(heute, wunsch.Wer)); fehler != nil {
		return domain.Reservierung{}, fehler
	}

	return zusage, nil
}

// Zusagen gibt, was für einen Vorgang festgehalten ist.
func (w *Zusagenwerk) Zusagen(
	ctx context.Context, vorgang string,
) ([]domain.Reservierung, error) {
	return w.ablage.FuerVorgang(ctx, vorgang)
}

// vonHier gibt die Zeilen, die dieser Vorgang von dieser Partei nimmt.
//
// **Das ist der Vertrag** (Befund 1 des Zweitblicks über Phase 5.6):
// [domain.ReservierungPasst] fragt, ob die **Besitzzeile** die Zusage
// hergibt; ob der **Vertrag** sie hergibt, fragte niemand. Wer bei einem
// Handel über ein Stück zweimal festhielt, gab beim Verbuchen zwei weg.
//
// **Gerechnet und geprüft wird im Bestand** — in derselben Transaktion wie
// das Eintragen (Befund 1 des Nachlaufs), und seit #2492 je Variante.
func vonHier(angebot domain.Handelsangebot, wer string) domain.Zusagevertrag {
	vertrag := domain.Zusagevertrag{}

	for _, stelle := range angebot.Positionen {
		if stelle.Von == wer {
			vertrag = append(vertrag, stelle)
		}
	}

	return vertrag
}
