package service

// **Die setweiten Stiloptionen gehen in den Block `styling:`** — F05
// (#2247).
//
// # Wie das gefunden wurde
//
// Im Setdatei-Rundlauf gegen MSE2. Zwei Fehler an derselben Naht:
//
//  1. **Der eingelesene Block landete unter `set info:`** — als Zeile
//     `styling:` mit flachen `paket/option`-Werten. MSE2 meldet beim Öffnen
//     „Unexpected key: 'styling'"; jedes Set, das diese Anwendung schrieb,
//     öffnete in MSE2 mit einer Warnung.
//  2. **Eine Stiloption, die hier gesetzt wurde, kam gar nicht in die Datei.**
//     `mse stiloption <set> setzen "chop top" 7`, dann `mse sichern`: kein
//     `chop` im geschriebenen Set — weder MSE2 noch diese Anwendung sah den
//     Wert nach dem Wiederöffnen.
//
// MSE2 legt die setweiten Optionen je Stilpaket ab (`Set::stylingDataFor`,
// `src/data/set.cpp`, Zeilen 99–101, Fassung f9d9356 vom 2021-07-06), und so
// steht es in jedem kanonischen Set: `styling:`, darunter der Name des
// Stilpakets, darunter die Optionen.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

const stilblockQuelle = "game: magic\n" +
	"stylesheet: m15-altered\n" +
	"set_info:\n" +
	"\ttitle: Probe\n" +
	"styling:\n" +
	"\tmagic-m15-altered:\n" +
	"\t\ttext_box_mana_symbols: magic-mana-small.mse-symbol-font\n" +
	"\tmagic-m15-saga:\n" +
	"\t\ttext_box_mana_symbols: magic-mana-small.mse-symbol-font\n" +
	"card:\n" +
	"\tname: Eine\n"

func stilblockGeschrieben(t *testing.T, quelle string, werte, stilwerte []domain.Value, kopf domain.SetHead) string {
	t.Helper()

	baum, fehler := msepaket.Zerteile("set", []byte(quelle))
	if fehler != nil {
		t.Fatalf("zerteilen: %v", fehler)
	}

	setwerteAngleichen(baum, werte)
	stilblockAngleichen(baum, kopf, stilwerte, "")

	return string(msepaket.Schreibe(baum))
}

func TestSetweiteStiloptionenStehenImBlockStyling(t *testing.T) {
	t.Parallel()

	kopf := domain.SetHead{Game: "magic", Style: "m15-altered"}
	eingelesen := []domain.Value{
		{Field: "title", Text: "Probe"},
		{Field: Setstilfeld, Text: "magic-m15-altered/text_box_mana_symbols: magic-mana-small.mse-symbol-font\n"},
	}

	t.Run("der eingelesene Block steht nicht unter set info", func(t *testing.T) {
		t.Parallel()

		if aus := stilblockGeschrieben(t, stilblockQuelle, eingelesen, nil, kopf); aus != stilblockQuelle {
			t.Errorf("ohne Änderung änderte sich die Datei:\n%s", aus)
		}
	})

	t.Run("eine gesetzte Option steht unter dem Stilpaket", func(t *testing.T) {
		t.Parallel()

		aus := stilblockGeschrieben(t, stilblockQuelle, eingelesen,
			[]domain.Value{{Field: "chop top", Text: "7"}}, kopf)

		soll := "\tmagic-m15-altered:\n" +
			"\t\ttext_box_mana_symbols: magic-mana-small.mse-symbol-font\n" +
			"\t\tchop_top: 7\n" +
			"\tmagic-m15-saga:\n"
		if !strings.Contains(aus, soll) {
			t.Errorf("die Option steht nicht im Block ihres Stils:\n%s", aus)
		}
	})

	t.Run("eine geänderte Option ersetzt ihre Zeile", func(t *testing.T) {
		t.Parallel()

		aus := stilblockGeschrieben(t, stilblockQuelle, eingelesen,
			[]domain.Value{{Field: "text box mana symbols", Text: "magic-mana-large.mse-symbol-font"}}, kopf)

		if strings.Count(aus, "text_box_mana_symbols") != 2 ||
			!strings.Contains(aus, "\tmagic-m15-altered:\n\t\ttext_box_mana_symbols: magic-mana-large.mse-symbol-font\n") {
			t.Errorf("die Zeile wurde nicht ersetzt:\n%s", aus)
		}
	})

	t.Run("eine Option als eingerückter Block wird zu einer Zeile", func(t *testing.T) {
		t.Parallel()

		block := strings.Replace(stilblockQuelle, "\t\ttext_box_mana_symbols: magic-mana-small.mse-symbol-font\n\tmagic-m15-saga:",
			"\t\tother options:\n\t\t\terste\n\t\t\tzweite\n\tmagic-m15-saga:", 1)

		aus := stilblockGeschrieben(t, block, eingelesen, []domain.Value{{Field: "other options", Text: "dritte"}}, kopf)
		// **Der Schlüssel bleibt, wie die Datei ihn schrieb** — nur der Wert
		// ist neu.
		if !strings.Contains(aus, "\t\tother options: dritte\n\tmagic-m15-saga:") || strings.Contains(aus, "zweite") {
			t.Errorf("der alte Block blieb:\n%s", aus)
		}
	})

	t.Run("ohne Block und ohne Stilpaket darin wird beides angelegt", func(t *testing.T) {
		t.Parallel()

		ohne := "game: magic\nstylesheet: m15\nset_info:\n\ttitle: Probe\ncard:\n\tname: Eine"

		aus := stilblockGeschrieben(t, ohne, []domain.Value{{Field: "title", Text: "Probe"}},
			[]domain.Value{{Field: "chop top", Text: "7"}}, domain.SetHead{Game: "magic", Style: "m15"})

		if !strings.HasSuffix(aus, "\tname: Eine\nstyling:\n\tmagic-m15:\n\t\tchop_top: 7\n") {
			t.Errorf("der Block wurde nicht angelegt:\n%q", aus)
		}
	})

	t.Run("das Stilpaket unter seinem rohen Namen wird gefunden", func(t *testing.T) {
		t.Parallel()

		roh := strings.ReplaceAll(stilblockQuelle, "\tmagic-m15-altered:", "\tm15-altered:")

		aus := stilblockGeschrieben(t, roh, nil, []domain.Value{{Field: "chop top", Text: "7"}}, kopf)
		if strings.Contains(aus, "magic-m15-altered") || !strings.Contains(aus, "\t\tchop_top: 7\n\tmagic-m15-saga:") {
			t.Errorf("ein zweiter Block für denselben Stil:\n%s", aus)
		}
	})

	t.Run("die Zeile einer früheren Abschrift unter set info verschwindet", func(t *testing.T) {
		t.Parallel()

		verirrt := strings.Replace(stilblockQuelle, "\ttitle: Probe\n",
			"\ttitle: Probe\n\tstyling:\n\t\tmagic-m15-altered/text_box_mana_symbols: magic-mana-small.mse-symbol-font\n", 1)

		if aus := stilblockGeschrieben(t, verirrt, eingelesen, nil, kopf); aus != stilblockQuelle {
			t.Errorf("die verirrte Zeile blieb stehen:\n%s", aus)
		}
	})

	t.Run("ein Setfeld styling in anderer Gestalt bleibt", func(t *testing.T) {
		t.Parallel()

		eigenes := strings.Replace(stilblockQuelle, "\ttitle: Probe\n", "\ttitle: Probe\n\tstyling: rot\n", 1)

		if aus := stilblockGeschrieben(t, eigenes, eingelesen, nil, kopf); aus != eigenes {
			t.Errorf("ein Setfeld wurde entfernt:\n%s", aus)
		}
	})
}

func TestEineVerirrteStilzeileWirdNichtEingelesen(t *testing.T) {
	t.Parallel()

	baum, fehler := msepaket.Zerteile("set", []byte(strings.Replace(stilblockQuelle, "\ttitle: Probe\n",
		"\ttitle: Probe\n\tstyling:\n\t\tmagic-alt/option: falsch\n", 1)))
	if fehler != nil {
		t.Fatalf("zerteilen: %v", fehler)
	}

	werte := werteAus(baum.Ersten("set info"))
	for _, wert := range werte {
		if wert.Field == Setstilfeld {
			t.Errorf("die verirrte Zeile wurde ein Setwert: %q", wert.Text)
		}
	}

	if len(werte) != 1 || werte[0].Field != "title" {
		t.Errorf("Setwerte %+v", werte)
	}
}
