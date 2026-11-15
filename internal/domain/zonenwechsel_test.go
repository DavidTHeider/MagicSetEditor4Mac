package domain_test

// **Ein Zonenwechsel macht ein neues Objekt** — Werkstück M18 der
// Regelmaschine
// (#1619).
//
// # Die Regel, und warum sie hier ganz gilt
//
// **CR 400.7** (Fassung 2026-08-19): was die Zone wechselt, ist ein neues
// Objekt ohne Erinnerung an sein voriges Dasein. Die Regel nennt **zwölf**
// Ausnahmen (400.7a–m) — und **jede einzelne** spricht von Effekten,
// Fähigkeiten oder Auslösern. Keines davon gibt es hier; deshalb greift die
// Regel ohne Abstrich, und die Ausnahmen kommen mit `M8` und `M9`.
//
// # Was vorher geschah
//
// **Zähler und Anhängsel überlebten die Reise.** Eine Kreatur mit drei
// Marken kam mit drei Marken aus dem Friedhof zurück, und eine Aura hing
// weiter an einer Karte, die es auf dem Tisch nicht mehr gab. G4 führte das
// als **Grenze in der Regelmatrix** — sie fällt jetzt weg.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einZonenwechseltisch gibt eine Partie mit einer belegten Kreatur.
func einZonenwechseltisch(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineGestaltenpartie(t)
	partie.Zug.Nummer = 4

	partie.Objekte[1].Zaehler = map[string]int{domain.MarkePlusEins: 3}
	partie.Objekte[1].Schaden = 2
	partie.Objekte[1].Verdeckt = true
	partie.Objekte[1].Seite = 0
	partie.Objekte[1].Seit = 1

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/aura/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "baer",
		HaengtAn: "sitz-1/baer/0",
	})

	return partie
}

// **Marken bleiben nicht** — CR 400.7 mit CR 122.2.
func TestMarkenUeberlebenDenZonenwechselNicht(t *testing.T) {
	t.Parallel()

	danach, fehler := einZonenwechseltisch(t).Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	baer := objektMit(t, danach, "sitz-1/baer/0")

	if anzahl := baer.Zaehler[domain.MarkePlusEins]; anzahl != 0 {
		t.Errorf("auf dem Bären liegen noch %d Marken", anzahl)
	}
}

// **Und der Zustand ebenso wenig** — getappt und verdeckt sind Zustände
// eines **Permanents** (CR 110.5), und im Friedhof liegt keines.
func TestZustandUeberlebtDenZonenwechselNicht(t *testing.T) {
	t.Parallel()

	danach, fehler := einZonenwechseltisch(t).Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	baer := objektMit(t, danach, "sitz-1/baer/0")

	if baer.Getappt {
		t.Error("der Bär liegt getappt im Friedhof")
	}

	if baer.Verdeckt {
		t.Error("der Bär liegt verdeckt im Friedhof")
	}

	if baer.Schaden != 0 {
		t.Errorf("der Bär trägt %d markierten Schaden", baer.Schaden)
	}
}

// **Ein Anhängsel hängt danach an nichts** — CR 400.7 aus der anderen
// Richtung: die Aura wird kein neues Objekt, aber das, woran sie hing,
// schon.
func TestEinAnhaengselHaengtNachDemZonenwechselAnNichts(t *testing.T) {
	t.Parallel()

	danach, fehler := einZonenwechseltisch(t).Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	// **Die Aura folgt von selbst** — CR 704.5m: eine Aura, die an nichts
	// hängt, geht in den Friedhof. Das ist `M5`, und hier steht, dass sie
	// **nicht mehr hängt**.
	if aura := objektMit(t, danach,
		"sitz-1/aura/0"); aura.HaengtAn != "" {
		t.Errorf("die Aura hängt weiter an %q", aura.HaengtAn)
	}
}

// **Und was selbst die Zone wechselt, hängt danach an nichts** — dieselbe
// Regel, dieselbe Richtung wie die Marken.
func TestEinAnhaengselLoestSichBeimEigenenZonenwechsel(t *testing.T) {
	t.Parallel()

	danach, fehler := einZonenwechseltisch(t).Exilieren("sitz-1/aura/0")
	if fehler != nil {
		t.Fatalf("exilieren: %v", fehler)
	}

	if aura := objektMit(t, danach,
		"sitz-1/aura/0"); aura.HaengtAn != "" {
		t.Errorf("die Aura hängt weiter an %q", aura.HaengtAn)
	}
}

// **Wer auf den Tisch kommt, kommt mit dem laufenden Zug** — CR 302.6.
func TestWerAufDenTischKommtIstNeuDabei(t *testing.T) {
	t.Parallel()

	danach, fehler := einZonenwechseltisch(t).Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	// **Auch im Friedhof** — die Zahl sagt, seit wann dieses Objekt so
	// dasteht, und das ist seit diesem Zug. Sie zählt erst auf dem
	// Schlachtfeld (CR 302.6); dass sie **nicht stehen bleibt**, ist der
	// Punkt: ein Objekt, das mit `Seit: 1` zurückkäme, dürfte in Zug 4
	// angreifen, obwohl es gerade erst gekommen ist.
	if baer := objektMit(t, danach, "sitz-1/baer/0"); baer.Seit != 4 {
		t.Errorf("der Bär ist seit Zug %d im Friedhof", baer.Seit)
	}
}

// **Und die Seite fällt auf die Vorderseite zurück** — CR 712.8a: außerhalb
// von Schlachtfeld und Stapel gelten deren Merkmale.
func TestDieSeiteFaelltBeimZonenwechselZurueck(t *testing.T) {
	t.Parallel()

	partie := einZonenwechseltisch(t)

	gedreht, fehler := partie.Umwandeln("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("umwandeln: %v", fehler)
	}

	danach, fehler := gedreht.Zerstoeren("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if seite := objektMit(t, danach,
		"sitz-1/wandler/0").Seite; seite != 0 {
		t.Errorf("im Friedhof liegt Seite %d", seite)
	}
}

// **Was in den Friedhof kommt, kommt obenauf** — CR 404.1.
//
// **Die andere Hälfte von Befund 2** aus `docs/REGELABDECKUNG.md`: der
// Friedhof gilt seit G9 als geordnete Zone, und bis zum 2026-09-17 bekam
// jede Karte darin die Null — alle lagen an derselben Stelle, und CR 404.3
// hatte nichts, worauf es zeigen konnte.
func TestWasInDenFriedhofKommtKommtObenauf(t *testing.T) {
	t.Parallel()

	erst, fehler := einZonenwechseltisch(t).Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("erstes zerstören: %v", fehler)
	}

	dann, fehler := erst.Zerstoeren("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("zweites zerstören: %v", fehler)
	}

	// **Null liegt oben** — der Grund steht bei `Spielobjekt.Ordnung`.
	if oben := objektMit(t, dann,
		"sitz-1/wandler/0").Ordnung; oben != 0 {
		t.Errorf("der zuletzt gelegte liegt an Stelle %d", oben)
	}

	if unten := objektMit(t, dann,
		"sitz-1/baer/0").Ordnung; unten != 1 {
		t.Errorf("der zuerst gelegte liegt an Stelle %d", unten)
	}
}

// **Und ein fremder Friedhof zählt nicht mit** — CR 404.1: er gehört einem
// Spieler.
func TestEinFremderFriedhofZaehltNichtMit(t *testing.T) {
	t.Parallel()

	partie := einZonenwechseltisch(t)
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/karte/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneFriedhof, Karte: "baer",
	})

	danach, fehler := partie.Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if fremd := objektMit(t, danach,
		"sitz-2/karte/0").Ordnung; fremd != 0 {
		t.Errorf("die fremde Karte ist auf Stelle %d gerückt", fremd)
	}
}
