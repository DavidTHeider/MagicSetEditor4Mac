package app

// **Der Beginn, bevor eine Vorrichtung spielt** —
// #1851.
//
// # Warum es diese Hilfe gibt
//
// **Weil der Beginn vom Aufbau an läuft** — CR 103.8 (Fassung 2026-09-25):
// bis jeder Sitz seine Starthand behalten hat, nimmt der Sandkasten nur ihn
// an. Ein Lauf, der an der Naht einen Befehl im Spiel prüft, bringt ihn
// deshalb zuerst zu Ende — **über dieselbe Fläche, über die er spielt**, und
// nicht, indem er den Stand im Bestand umschreibt (§3).
//
// # Und warum die Objekte danach gesucht werden
//
// **Weil die Starthand die Bibliothek mischt** (CR 103.3): welche Karte auf
// der Hand liegt und welche in der Bibliothek, ist danach Zufall. Ein Lauf,
// der „das erste Objekt" nähme, prüfte einmal eine Handkarte und einmal eine
// aus der Bibliothek — gesucht wird deshalb im Stand ([inDerBibliothek]).

import (
	"context"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// befehlsweg stellt einen Spielbefehl über eine Fläche.
type befehlsweg func(befehl domain.Spielbefehl) BefehlAntwort

// ueberDieBindungen stellt ihn über das Fenster.
func ueberDieBindungen(ctx context.Context, gebunden *Bindungen) befehlsweg {
	return func(befehl domain.Spielbefehl) BefehlAntwort {
		return gebunden.PartieBefehl(ctx, befehl.PartieID, befehl.Kennung,
			befehl.Absender, befehl.Art, befehl.Nutzlast, "")
	}
}

// ueberDieFlaeche stellt ihn über die kopflose Fläche.
func ueberDieFlaeche(ctx context.Context, flaeche *Partieflaeche) befehlsweg {
	return func(befehl domain.Spielbefehl) BefehlAntwort {
		return flaeche.PartieBefehl(ctx, befehl)
	}
}

// beginnAbschliessen bringt den Beginn zu Ende — für jeden Sitz die
// Starthand, dann behalten (CR 103.5) — und gibt die Fassung danach.
//
// **Jeder Sitz, den die Partie nennt**, und nicht zwei feste: der Beginn ist
// erst vorbei, wenn der letzte behalten hat.
func beginnAbschliessen(
	t *testing.T, stellen befehlsweg, partie PartieDTO,
) string {
	t.Helper()

	fassung := partie.Fassung

	for _, sitz := range partie.Sitze {
		for _, art := range []string{"spieler.starthand", "spieler.behalten"} {
			antwort := stellen(domain.Spielbefehl{
				PartieID: partie.ID, Kennung: art + "/" + sitz.Kennung,
				Absender: sitz.Kennung, Art: art,
			})
			if antwort.Fehler != nil {
				t.Fatalf("%s für %s: %+v", art, sitz.Kennung, antwort.Fehler)
			}

			fassung = antwort.Fassung
		}
	}

	return fassung
}

// inDerBibliothek gibt die Kennung eines Objekts, das in der Bibliothek von
// `wer` liegt — gesucht im Stand, nicht an einer festen Stelle.
//
// **Die Bibliothek und nicht die Hand**: bis #1851 lag dort jedes Objekt, und
// die Läufe, die eines schieben oder tappen, taten das mit einer Karte aus
// ihr. Sie tun es weiter.
func inDerBibliothek(t *testing.T, objekte []ObjektDTO, wer string) string {
	t.Helper()

	for _, objekt := range objekte {
		if objekt.Beherrscher == wer && objekt.Zone == "bibliothek" {
			return objekt.Kennung
		}
	}

	t.Fatalf("in der Bibliothek von %s liegt nichts: %+v", wer, objekte)

	return ""
}

// naechsteFassung ist die Fassung, die ein Befehl nach dieser hinterlässt.
func naechsteFassung(t *testing.T, fassung string) string {
	t.Helper()

	zahl, fehler := strconv.ParseInt(fassung, 10, 64)
	if fehler != nil {
		t.Fatalf("die Fassung %q ist keine Zahl", fassung)
	}

	return strconv.FormatInt(zahl+1, 10)
}
