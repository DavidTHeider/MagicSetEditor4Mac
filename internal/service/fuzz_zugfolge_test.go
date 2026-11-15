package service_test

// **Priorität und Stapel unter einer fremden Befehlsfolge** — QA-06b
// (#1967),
// Fläche 77 aus `docs/FUZZZIELE.md`.
//
// # Eine Folge mit Zustand
//
// **Ein einzelner Befehl ist hier keine Eingabe, sondern ein Schritt.** Die
// Fehler einer Regelmaschine liegen in Zuständen, die nur eine Folge
// erreicht: ein Zauber auf dem Stapel, während alle passen; ein Angriff im
// falschen Abschnitt; ein Zugende mit offener Priorität. Der Fuzzer schreibt
// die Folge — je drei Byte Befehl, Absender und Ziel — über einer Partie
// nach dem Beginn, wie [eineZugpartie] sie baut, in der Strenge, die er
// wählt.
//
// **Angewandt wird wie in der Partieleitung**: eine Partie, die nicht läuft,
// und ein Absender ohne Sitz kommen gar nicht erst zum Ausleger
// (`Partieleitung.pruefeAnnahme`).
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **nach jedem angenommenen Schritt** besteht die Partie
//     [domain.Partie.Pruefe] nicht
//   - **solange die Partie läuft**, liegt die Priorität bei jemandem, der
//     keinen Sitz hat oder schon ausgeschieden ist; der aktive Spieler hat
//     keinen Sitz
//   - „gepasst" nennt einen Sitz zweimal oder einen, den es nicht gibt
//     (CR 117.4, Fassung 2026-08-19: **wer** passt, nicht wie viele) — und
//     in einer laufenden Partie einen, der schon gegangen ist
//   - **zwei Objekte auf dem Stapel mit derselben Ordnung** — CR 405.2
//     legt jedes neue **über** alle vorhandenen; zwei Objekte mit derselben
//     Kennung irgendwo
//   - der Zug zählt rückwärts
//   - **die Eingabe des Auslegers ist danach eine andere** — abgelehnt oder
//     angenommen: die Leitung behält bei einer Ablehnung den Stand von
//     vorher, und der muss dann noch derselbe sein
//   - abgelehnt ohne Klasse — eine der drei aus §8: [domain.ErrValidation],
//     [domain.ErrNotFound], [domain.ErrConflict]; zweimal derselbe Schritt,
//     zwei Ergebnisse
//
// **[domain.ErrNotFound] fehlte hier vor dem ersten Lauf**, und der Lauf
// fand es: ein Grundland, das es nicht gibt, ist „nicht gefunden". Die
// Zusage war zu eng, nicht das Programm.
//
// **„Solange die Partie läuft" kam nach dem Saatlauf dazu**, und der Grund
// gehört hierher: wer bei zweien aufgibt, beendet die Partie (CR 104.2a),
// und eine beendete Partie nimmt keinen Befehl mehr an
// (`Partieleitung.pruefeAnnahme`). Die Zusage über eine laufende Partie war
// gemeint; die über eine beendete wäre eine über eine Anzeige.
//
// # Zwei und drei Sitze
//
// **Drei Sitze fährt das Ziel seit #1971.** Dort läuft die Partie nach einer
// Aufgabe weiter, und bis dahin behielt der Gehende die Priorität — die
// anderen konnten nicht passen. CR 800.4a und CR 800.4j (Fassung
// 2026-09-25) sagen, was stattdessen gilt; mit drei Sitzen sieht dieses Ziel
// es, und „gepasst" nennt seitdem nur, wer noch im Spiel ist (CR 117.4).
//
//	go test ./internal/service -run Fuzz -fuzz FuzzZugfolge -fuzztime 2m

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zugarten sind die Befehle, die Priorität und Stapel bewegen.
var zugarten = []string{
	service.BefehlPassen, service.BefehlAbkuerzen, service.BefehlZugende,
	service.BefehlWirken, service.BefehlLandspielen, service.BefehlLandmana,
	service.BefehlMana, service.BefehlAngreifer, service.BefehlBlocker,
	service.BefehlAbwerfen, service.BefehlNeutralisieren, service.BefehlAufgeben,
}

// **Nach jedem Schritt eine gültige Partie — und die Eingabe bleibt, wie
// sie war.**
func FuzzZugfolge(f *testing.F) {
	for _, saat := range [][]byte{
		{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0},
		{4, 0, 7, 6, 0, 0, 5, 0, 7, 3, 0, 7, 0, 0, 0, 0, 1, 0},
		{1, 0, 0, 1, 1, 0, 7, 0, 3, 8, 1, 4, 0, 0, 0, 0, 1, 0, 2, 0, 0},
		{3, 0, 9, 10, 1, 9, 0, 1, 0, 0, 0, 0, 11, 1, 0},
		{2, 2, 0, 9, 0, 1, 255, 0, 0},
	} {
		for wahl := range byte(6) {
			f.Add(wahl, saat)
		}
	}

	f.Fuzz(func(t *testing.T, wahl byte, folge []byte) {
		partie := zugpartieIn(t, wahl)

		for schritt := 0; schritt+2 < len(folge) && schritt < 3*48; schritt += 3 {
			if partie.Stand != domain.StandLaeuft {
				return
			}

			befehl := zugbefehl(partie, folge[schritt], folge[schritt+1], folge[schritt+2])
			if partie.SitzVon(befehl.Absender) == nil {
				continue
			}

			partie = zugschritt(t, partie, befehl)
		}
	})
}

// zugpartieIn baut die Partie nach dem Beginn — in einer der drei Strengen
// und mit zwei oder drei Sitzen.
func zugpartieIn(t *testing.T, wahl byte) domain.Partie {
	t.Helper()

	regelwerk := einRegelwerk()
	regelwerk.Strenge = []domain.Strenge{
		domain.StrengeSandkasten, domain.StrengeBegleitet, domain.StrengeStreng,
	}[int(wahl)%3]

	wuensche := zweiWuensche()
	if wahl/3%2 == 1 {
		wuensche = append(wuensche, service.Sitzwunsch{Kennung: "sitz-3", Name: "Carla", DeckID: 1})
	}

	partie, fehler := einSpielaufbau().Aufbauen(t.Context(), "p1", "magic",
		regelwerk, "sitz-1", wuensche)
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	return beginnAbschliessen(t, partie)
}

// zugbefehl macht aus drei Byte einen Befehl: Art, Absender, Ziel.
func zugbefehl(partie domain.Partie, art, wer, ziel byte) domain.Spielbefehl {
	absender := "fremd"
	if int(wer) < len(partie.Sitze) {
		absender = partie.Sitze[wer].Kennung
	}

	objekt, anderes := "", ""
	if len(partie.Objekte) > 0 && ziel != 255 {
		objekt = partie.Objekte[int(ziel)%len(partie.Objekte)].Kennung
		anderes = partie.Objekte[(int(ziel)/7)%len(partie.Objekte)].Kennung
	}

	gegner := partie.Sitze[(int(wer)+1)%len(partie.Sitze)].Kennung
	befehlsart := zugarten[int(art)%len(zugarten)]

	return domain.Spielbefehl{
		PartieID: partie.ID, Kennung: befehlsart, Absender: absender, Art: befehlsart,
		Nutzlast: map[string]any{
			"objekt": objekt, "art": string("WUBRGC"[int(ziel)%6]), "um": float64(int(ziel) % 4),
			"angriffe": []any{map[string]any{"kreatur": objekt, "ziel": gegner}},
			"bloecke":  []any{map[string]any{"kreatur": objekt, "angreifer": anderes}},
		},
	}
}

// zugschritt legt einen Befehl aus und hält die Zusagen über den Schritt.
func zugschritt(t *testing.T, partie domain.Partie, befehl domain.Spielbefehl) domain.Partie {
	t.Helper()

	vorher := alsJSON(t, partie)

	ereignisse, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(partie, befehl)
	nochmalEreignisse, nochmal, nochmalFehler := service.NeuerSandkasten(festeMischung{}).Auslegen(partie, befehl)

	if nachher := alsJSON(t, partie); nachher != vorher {
		t.Fatalf("%s von %s hat seine Eingabe verändert (%v)", befehl.Art, befehl.Absender, fehler)
	}

	if (fehler == nil) != (nochmalFehler == nil) || alsJSON(t, danach) != alsJSON(t, nochmal) ||
		alsJSON(t, ereignisse) != alsJSON(t, nochmalEreignisse) {
		t.Fatalf("%s zweimal ausgelegt, zwei Ergebnisse: %v / %v", befehl.Art, fehler, nochmalFehler)
	}

	if fehler != nil {
		if !errors.Is(fehler, domain.ErrValidation) && !errors.Is(fehler, domain.ErrNotFound) &&
			!errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("%s abgelehnt ohne Klasse: %v", befehl.Art, fehler)
		}

		return partie
	}

	zugstandPruefen(t, befehl, partie, danach)

	return danach
}

// zugstandPruefen hält die Partie nach einem angenommenen Schritt.
func zugstandPruefen(t *testing.T, befehl domain.Spielbefehl, vorher, danach domain.Partie) {
	t.Helper()

	if fehler := danach.Pruefe(); fehler != nil {
		t.Errorf("nach %s von %s ungültig: %v", befehl.Art, befehl.Absender, fehler)
	}

	draussen := map[string]bool{}
	for _, gegangen := range danach.Ausgang.Ausgeschieden {
		draussen[gegangen.Sitz] = true
	}

	zug := danach.Zug
	if danach.Stand == domain.StandLaeuft && zug.Prioritaet != "" &&
		(danach.SitzVon(zug.Prioritaet) == nil || draussen[zug.Prioritaet]) {
		t.Errorf("nach %s liegt die Priorität bei %q", befehl.Art, zug.Prioritaet)
	}

	if zug.Nummer > 0 && danach.SitzVon(zug.Aktiv) == nil {
		t.Errorf("nach %s ist %q aktiv", befehl.Art, zug.Aktiv)
	}

	if zug.Nummer < vorher.Zug.Nummer {
		t.Errorf("nach %s steht Zug %d, vorher %d", befehl.Art, zug.Nummer, vorher.Zug.Nummer)
	}

	gepasst := map[string]bool{}
	for _, wer := range zug.Gepasst {
		if gepasst[wer] || danach.SitzVon(wer) == nil || (danach.Stand == domain.StandLaeuft && draussen[wer]) {
			t.Errorf("nach %s nennt „gepasst“ %v", befehl.Art, zug.Gepasst)
		}

		gepasst[wer] = true
	}

	kennungen, ordnungen := map[string]bool{}, map[int]string{}
	for _, objekt := range danach.Objekte {
		if kennungen[objekt.Kennung] {
			t.Errorf("nach %s gibt es %q zweimal", befehl.Art, objekt.Kennung)
		}

		kennungen[objekt.Kennung] = true

		if objekt.Zone != domain.ZoneStapel {
			continue
		}

		if schon, da := ordnungen[objekt.Ordnung]; da {
			t.Errorf("nach %s liegen %q und %q auf dem Stapel an Stelle %d",
				befehl.Art, schon, objekt.Kennung, objekt.Ordnung)
		}

		ordnungen[objekt.Ordnung] = objekt.Kennung
	}
}

// alsJSON ist der Vergleich für einen Stand — die Partie trägt Karten und
// Scheiben, und ihre JSON-Form hält jedes Feld, das jemand liest.
func alsJSON(t *testing.T, wert any) string {
	t.Helper()

	roh, fehler := json.Marshal(wert)
	if fehler != nil {
		t.Fatalf("als JSON: %v", fehler)
	}

	return string(roh)
}
