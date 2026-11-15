//go:build fremdpakete

package service_test

// **`average opening hand` — die Abnahme, die möglich ist** —
// #1355.
//
// # Warum das Orakel hier nicht antwortet
//
// Das Issue verlangt „denselben Wert wie MSE2, gefragt über
// `magicseteditor --cli`". **Das geht nicht**, und zwar nicht an dieser
// Umsetzung:
//
//	$ echo 'average_opening_hand_statistic(card: set.cards.0)' \
//	    | magicseteditor --cli Colorless.mse-set -q --raw
//	ERROR: Variable not set: global_value
//	         in function average_opening_hand_statistic
//
// `global_script:` ist ein Schlüssel, den **MSE 2.1.2 nicht kennt**
// (`IMPLEMENT_REFLECTION_NO_GET_MEMBER(StatsDimension)`, Fassung `f9d9356`
// vom 2021-07-06, führt elf Angaben ohne ihn). `magic.mse-game` schreibt
// `mse version: 2.5.8`; unser Orakel ist 2.1.2, und ein neuerer Quelltext
// liegt hier nicht vor — gesucht in `/Users/david/mse2-quelle`, in
// `/Users/david/magic-set-generator` und im ganzen Pack.
//
// # Was stattdessen geprüft wird
//
// **Eine Eigenschaft, die aus der Formel des Pakets folgt.**
// `average_opening_hand_global_statistic` rechnet je Kartenart
//
//	7.0 * anzahl / gesamtzahl
//
// über **dieselbe** Grundgesamtheit. Die neun Durchschnitte summieren sich
// deshalb auf **sieben** — eine Eröffnungshand hat sieben Karten. Eine Karte
// kann mehrere Arten tragen (`Artifact Creature`), also ist sieben die
// **Untergrenze**; nach oben begrenzt die Zahl der Arten je Karte.
//
// **Gesammelt wird über alle Karten und nicht aus einer Zeile.** Das
// Kartenskript zeigt nur die Arten, die **diese** Karte trägt:
//
//	result := for x from 0 to 8 do (if types[x] then global_value[x] else "")
//
// Die erste Karte von `Colorless.mse-set` ist ein Land und gibt deshalb
// `Land — 0.83` und sonst nichts. Der erste Anlauf hier erwartete die Summe
// aus **einer** Zeile und war rot — die Erwartung war falsch, nicht die
// Rechnung. Was gilt, ist die Summe über die **verschiedenen** Gruppen.
//
// **Das ist weniger als ein Orakelvergleich, und der Satz steht hier, damit
// niemand mehr hineinliest** (§10).

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// TestDieEroeffnungshandRechnetSiebenKarten ist die Abnahme.
func TestDieEroeffnungshandRechnetSiebenKarten(t *testing.T) {
	setpfad := kanonischesSet(t, "Colorless.mse-set")

	paket, fehler := msepaket.Oeffne(filepath.Join(packort, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser(packort)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	var dimension service.Auszaehlangabe

	for _, eine := range service.Auszaehlangaben(paket.Auszaehldimensionen()) {
		if eine.Name == "average opening hand" {
			dimension = eine
		}
	}

	if dimension.Globalskript == "" {
		t.Fatal("die Dimension trägt kein globales Skript")
	}

	karten, setangaben := kartenUndAngabenAusSet(t, setpfad)

	speicher := service.NeuerStilspeicher(&setkopfprobe{
		kopf:  domain.SetHead{ID: 1, Game: "magic", Style: "m15-altered"},
		werte: setangaben,
	}, packort)

	vorlage := magicvorlage(spiel)
	vorlage.Setangaben = setangaben

	// **Ohne `split list:`** — verglichen wird der rohe Wert des Skripts.
	dimension.Zerlegen = false
	dimension.LeereZeigen = true

	werte, befunde := service.Auszaehlwerte(
		t.Context(), karten, vorlage,
		func(karte *domain.CardRow) service.Vorlage {
			eigene, _, da := speicher.StilFuerKarte(t.Context(), karte)
			if !da {
				return vorlage
			}

			eigene.Spielskript = vorlage.Spielskript
			eigene.Setfelder = vorlage.Setfelder
			eigene.Kartenfelder = vorlage.Kartenfelder
			eigene.Setangaben = setangaben

			return eigene
		}, dimension)

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) == 0 {
		t.Fatal("keine Werte")
	}

	t.Logf("die erste Karte gibt %q", werte[0])

	// **Es ist eine Rechnung und keine Meldung.**
	// **Die Nullbreiten-Leerzeichen sind Absicht des Pakets** — es setzt sie
	// vor jede Gruppe, um die Reihenfolge zu erzwingen (`statistics_script`,
	// Zeile 645: „There are zero width spaces after the initial \",\" to
	// impose ordering"). Als Escape geschrieben, weil ein unsichtbares
	// Zeichen im Quelltext niemandem hilft.
	if strings.HasPrefix(strings.TrimLeft(werte[0], ",\u200b"), "!!") {
		t.Fatalf("die Dimension gibt eine Meldung statt einer Rechnung: %q", werte[0])
	}

	// **Jede Gruppe einmal**, gesammelt über alle Karten. Dieselbe Gruppe
	// trägt bei jeder Karte denselben Wert — sie kommt ja aus dem globalen
	// Skript.
	gruppen := map[string]float64{}

	for _, wert := range werte {
		for _, stueck := range strings.Split(wert, ",") {
			teile := strings.Split(stueck, "—")
			if len(teile) != 2 {
				continue
			}

			name := strings.TrimSpace(strings.Trim(teile[0], "\u200b"))

			zahl, err := strconv.ParseFloat(strings.TrimSpace(teile[1]), 64)
			if err != nil {
				t.Errorf("%q ist keine Zahl: %v", teile[1], err)

				continue
			}

			if vorher, da := gruppen[name]; da && vorher != zahl {
				t.Errorf("die Gruppe %q trägt zwei Werte: %.2f und %.2f — "+
					"das globale Skript hat mehr als einmal gerechnet",
					name, vorher, zahl)
			}

			gruppen[name] = zahl
		}
	}

	summe := 0.0
	for _, zahl := range gruppen {
		summe += zahl
	}

	t.Logf("%d Gruppen, Summe %.2f", len(gruppen), summe)

	// **Sieben, mit Luft nach oben für Karten mit zwei Arten.**
	//
	// Eine `Artifact Creature` zählt in zwei Gruppen; nach unten kann die
	// Summe aber nicht: jede gezählte Karte trägt mindestens eine Art.
	if summe < 6.9 {
		t.Errorf("die neun Durchschnitte summieren sich auf %.2f — erwartet mindestens 7", summe)
	}

	if summe > 14.1 {
		t.Errorf("die Summe ist %.2f — mehr als zwei Arten je Karte im Mittel", summe)
	}
}
