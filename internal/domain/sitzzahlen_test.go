package domain_test

// **Was am Sitz steht, steht auch in der Sicht** — PLAY-03
// (#1711).
//
// # Der Anlass
//
// Bis zum 2026-09-21 kopierte [domain.Partie.Sicht] **acht** der zwölf
// Felder eines Sitzes. Strahlung, Commander, Commandersteuer,
// Commanderschaden und Leergezogen blieben liegen — gerechnet, gespeichert,
// geprüft und nie gezeigt.
//
// **Keines davon ist geheim.** CR 728.1 setzt Strahlungsmarken an den
// **Spieler**, CR 400.2 macht die Kommandozone öffentlich, CR 903.8 zählt
// Würfe, die alle gesehen haben, CR 903.10a nennt Kampfschaden, und der
// Griff in eine leere Bibliothek (CR 704.5b) geschieht am Tisch.
//
// # Was hier geprüft wird und was daneben
//
// **Hier: dass die Zahlen ankommen.** Dass sie über die Naht kommen, hält
// `internal/werkzeuge/sitzfelder_test.go`; dass eine Fläche sie zeigt, die
// Läufe über `Spielerleiste` und `sitzkopf`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einSitzMitZahlen gibt eine Partie, deren erster Sitz alle fünf trägt.
func einSitzMitZahlen() domain.Partie {
	partie := einePartie()

	partie.Sitze[0].Strahlung = 3
	partie.Sitze[0].Commander = "sitz-1/held/0"
	partie.Sitze[0].Commandersteuer = 2
	partie.Sitze[0].Commanderschaden = map[string]int{
		"sitz-2/drache/0": 7,
		"sitz-2/baer/0":   1,
	}
	partie.Sitze[0].Leergezogen = true

	// **Und der Beginn** (#1839): wie oft neu gezogen wurde und was noch
	// unter die Bibliothek muss, wird am Tisch erklärt (CR 103.5).
	partie.Sitze[0].Eroeffnung = domain.EroeffnungOffen
	partie.Sitze[0].Mulligans = 2
	partie.Sitze[0].Unterlegen = 1

	return partie
}

func TestSichtTraegtDieZahlenDesSitzes(t *testing.T) {
	sicht := einSitzMitZahlen().Sicht("sitz-2")

	if len(sicht.Sitze) == 0 {
		t.Fatal("die Sicht hat keinen Sitz")
	}

	erster := sicht.Sitze[0]

	if erster.Strahlung != 3 {
		t.Errorf("die Strahlung steht bei %d, erwartet 3", erster.Strahlung)
	}

	if erster.Commander != "sitz-1/held/0" {
		t.Errorf("der Commander ist %q", erster.Commander)
	}

	if erster.Commandersteuer != 2 {
		t.Errorf("die Steuer steht bei %d, erwartet 2", erster.Commandersteuer)
	}

	if !erster.Leergezogen {
		t.Error("der Vermerk über die leere Bibliothek fehlt")
	}

	if erster.Eroeffnung != domain.EroeffnungOffen || erster.Mulligans != 2 ||
		erster.Unterlegen != 1 {
		t.Errorf("der Beginn fehlt: %q, %d, %d", erster.Eroeffnung,
			erster.Mulligans, erster.Unterlegen)
	}

	if schaden := erster.Commanderschaden["sitz-2/drache/0"]; schaden != 7 {
		t.Errorf("der Commanderschaden steht bei %d, erwartet 7", schaden)
	}
}

// **Ein Gegenüber sieht dieselben Zahlen.** Sie sind öffentlich; eine Sicht,
// die sie nur dem eigenen Sitz gäbe, machte aus einer Buchführung ein
// Geheimnis — derselbe Satz wie beim Lebensstand.
func TestDieZahlenGeltenFuerJedenAmTisch(t *testing.T) {
	partie := einSitzMitZahlen()

	fuerSichSelbst := partie.Sicht("sitz-1").Sitze[0]
	fuerDenAnderen := partie.Sicht("sitz-2").Sitze[0]

	if fuerSichSelbst.Commandersteuer != fuerDenAnderen.Commandersteuer ||
		fuerSichSelbst.Strahlung != fuerDenAnderen.Strahlung {
		t.Errorf("zwei Sichten, zwei Zahlen: %+v gegen %+v",
			fuerSichSelbst, fuerDenAnderen)
	}
}

// **Die Tabelle der Sicht gehört ihr allein** — dieselbe Regel wie beim
// Manavorrat daneben: eine Sicht ist eine Projektion und kein Griff in den
// Zustand. Wer sie durchreichte, ließe einen Empfänger den Spielstand
// ändern.
func TestDerCommanderschadenDerSichtIstEineKopie(t *testing.T) {
	partie := einSitzMitZahlen()

	sicht := partie.Sicht("sitz-1")
	sicht.Sitze[0].Commanderschaden["sitz-2/drache/0"] = 21

	if partie.Sitze[0].Commanderschaden["sitz-2/drache/0"] != 7 {
		t.Errorf("die Sicht hat in den Zustand geschrieben: %d",
			partie.Sitze[0].Commanderschaden["sitz-2/drache/0"])
	}
}

// **Ohne Schaden steht eine leere Tabelle da und kein `nil`.** An der Naht
// ist das ein Unterschied: `nil` wird zu `null`, und das Fenster liest
// etwas anderes als eine leere Tabelle (§3, „zwischen den Prüfebenen").
func TestOhneCommanderschadenStehtEineLeereTabelle(t *testing.T) {
	sicht := einePartie().Sicht("sitz-1")

	if sicht.Sitze[0].Commanderschaden == nil {
		t.Error("die Tabelle ist nil statt leer")
	}

	if len(sicht.Sitze[0].Commanderschaden) != 0 {
		t.Errorf("die Tabelle ist nicht leer: %v",
			sicht.Sitze[0].Commanderschaden)
	}
}
