// Package dateisperre ist die beratende Sperre auf einer Datei.
//
// # Warum es dieses Paket gibt
//
// Weil es sonst zwei Kopien gäbe — dieselbe Überlegung, aus der
// `internal/netz` entstanden ist. Zwei Wege in diesem Programm brauchen eine
// Absprache **über Prozessgrenzen hinweg**
// (#1163):
//
//   - `internal/storage` hält den Bestand über einen Tausch (Befund Z6-02),
//   - `internal/netz` teilt sich das Hostbudget einer Gegenstelle (Z7-03).
//
// Beide brauchen dasselbe: eine Nachbardatei, ein `flock`, und einen Weg, der
// auf Windows nichts tut. Zwei Kopien davon liefen an dem Tag auseinander, an
// dem eine von ihnen etwas dazulernte.
//
// # Ein Blatt (§2)
//
// Es kennt **nichts** aus diesem Projekt — auch nicht `domain`. Was es tut,
// ist Betriebssystem und nicht Fachlichkeit.
//
// # Was sie nicht kann
//
//   - **Sie ist beratend.** Wer die Datei ohne diese Nachbardatei anfasst,
//     wird nicht aufgehalten; kein Unix-Dateisystem hält jemanden auf.
//   - **Sie überlebt keinen Absturz**, und das ist richtig: das
//     Betriebssystem gibt sie mit dem Prozess frei. Eine Sperre, die einen
//     Absturz überlebte, machte ihren Gegenstand für immer unbenutzbar.
//   - **Auf Windows gibt es sie nicht.** `flock` fehlt dort, und der Weg über
//     `LockFileEx` wäre eine Abhängigkeit für eine Plattform, die dieses
//     Projekt beschreibt und nicht baut (§12).
package dateisperre

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Griff ist eine offene Sperrdatei.
//
// **`flock` gilt für die offene Datei und nicht für den Prozess.** Zwei Griffe
// auf dieselbe Datei stehen einander im Weg; wer eine Sperre je Prozess will,
// hält **einen** Griff und reicht ihn herum.
type Griff struct {
	datei *os.File
}

// Oeffnen legt die Sperrdatei an und gibt einen Griff darauf.
//
// **Sie bleibt leer.** Was zählt, ist die Sperre des Betriebssystems auf ihr;
// ein Inhalt wäre eine zweite Auskunft über dieselbe Sache.
func Oeffnen(pfad string) (*Griff, error) {
	// #nosec G304 -- die Sperrdatei, die der Aufrufer benennt
	datei, err := os.OpenFile(pfad, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("die Sperrdatei %s anlegen: %w", pfad, err)
	}

	return &Griff{datei: datei}, nil
}

// Teilen nimmt die geteilte Sperre, ohne zu warten.
func (g *Griff) Teilen() error { return teilen(g.datei) }

// Allein nimmt die ausschließliche Sperre, ohne zu warten.
func (g *Griff) Allein() error { return allein(g.datei) }

// Schliessen gibt den Griff und damit die Sperre zurück.
func (g *Griff) Schliessen() error { return g.datei.Close() }

// TeilenBis versucht die geteilte Sperre, bis die Frist um ist.
//
// **Gewartet und nicht abgewiesen**: wer während eines fremden Alleinanspruchs
// dazukommt, soll den **fertigen** Stand bekommen. Und nicht unbegrenzt: ein
// Anspruch, der hängenbleibt, hielte sonst jeden Aufrufer für immer auf.
func (g *Griff) TeilenBis(ctx context.Context, frist, takt time.Duration) error {
	return warteAuf(ctx, g.Teilen, frist, takt)
}

// AlleinBis versucht die ausschließliche Sperre, bis die Frist um ist.
func (g *Griff) AlleinBis(ctx context.Context, frist, takt time.Duration) error {
	return warteAuf(ctx, g.Allein, frist, takt)
}

// Erbfrist ist, wie lange eine Sperre belegt bleiben muss, damit sie als
// gehalten zählt — #2401
// (#2401).
//
// **Ein `flock` hängt an der offenen Dateibeschreibung**, und ein Kind, das
// irgendein Teil des Prozesses gerade startet, hält sie zwischen `fork` und
// `exec` mit. Gemessen am 2026-10-14: 672 von 20.000 Versuchen fanden die
// Sperre so belegt, während daneben `/usr/bin/true` lief. Ein Halter hält,
// solange er etwas tut; ein Erbe geht nach Millisekunden.
const Erbfrist = 100 * time.Millisecond

// erbtakt ist, wie oft innerhalb der [Erbfrist] gefragt wird.
const erbtakt = 5 * time.Millisecond

// AlleinOhneErbe nimmt die alleinige Sperre und wartet dafür einen Erben ab —
// höchstens die [Erbfrist]. **Für jede Probe, die sonst einmal fragte**: ein
// Kindprozess im selben Prozess ist kein zweites Programm.
//
// **Nicht `fcntl` statt `flock`**: jene Sperren werden nicht vererbt, gelten
// aber je Prozess — zwei Griffe im selben Prozess sperrten einander nicht
// mehr, und das Schließen irgendeines Griffs gäbe alle frei.
func (g *Griff) AlleinOhneErbe(ctx context.Context) error {
	return g.AlleinBis(ctx, Erbfrist, erbtakt)
}

// warteAuf versucht es, bis es gelingt, die Frist um ist oder der Kontext endet.
//
// # Warum der Kontext hierher gehört
//
// Ohne ihn wartete ein längst abgebrochener Aufruf die volle Frist ab — und
// nähme danach womöglich noch eine Sperre, die niemand mehr braucht.
// Angestrichen vom Zweitblick über Phase 3.2
// (#1163).
func warteAuf(ctx context.Context, versuchen func() error, frist, takt time.Duration) error {
	bis := time.Now().Add(frist)

	uhr := time.NewTicker(takt)
	defer uhr.Stop()

	for {
		fehler := versuchen()
		if fehler == nil || time.Now().After(bis) {
			return fehler
		}

		select {
		case <-uhr.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// LiesVonVorn liest vom Anfang der Sperrdatei.
//
// **Vom Anfang und nicht von der Schreibmarke**: derselbe Griff wird gelesen
// und geschrieben, und ohne das Zurücksetzen läse der zweite Durchgang hinter
// dem Ende.
func (g *Griff) LiesVonVorn(nach []byte) (int, error) {
	if _, err := g.datei.Seek(0, 0); err != nil {
		return 0, fmt.Errorf("die Sperrdatei zurücksetzen: %w", err)
	}

	return g.datei.Read(nach)
}

// SchreibVonVorn schreibt an den Anfang der Sperrdatei.
func (g *Griff) SchreibVonVorn(roh []byte) error {
	if _, err := g.datei.Seek(0, 0); err != nil {
		return fmt.Errorf("die Sperrdatei zurücksetzen: %w", err)
	}

	if _, err := g.datei.Write(roh); err != nil {
		return fmt.Errorf("die Sperrdatei schreiben: %w", err)
	}

	return nil
}
