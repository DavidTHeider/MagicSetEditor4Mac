package domain_test

// **Ein markierter Block in einer fremden Datei** — OPS-09.AK-01 und AK-02
// (#1907):
// einsetzen und widerrufen sind umkehrbar und idempotent, und widerrufen
// hinterlässt die Datei **byte-gleich** zum Stand davor.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const probeblock = "export PATH=\"$HOME/.local/bin:$PATH\"\n"

func TestEinBlockGehtHineinUndSpurlosWiederHinaus(t *testing.T) {
	t.Parallel()

	for name, vorher := range map[string]string{
		"leer":                  "",
		"mit Zeilenende":        "alias ll='ls -l'\n",
		"ohne Zeilenende":       "alias ll='ls -l'",
		"mit Leerzeile am Ende": "eins\n\n",
		"mit fremdem Kommentar": "# >>> conda initialize >>>\nconda\n# <<< conda initialize <<<\n",
	} {
		eingesetzt := domain.BlockEinsetzen([]byte(vorher), probeblock)

		if domain.BlockStand(eingesetzt, probeblock) != domain.BlockEingetragen {
			t.Errorf("%s: nach dem Einsetzen %q", name, domain.BlockStand(eingesetzt, probeblock))
		}

		// **Zweimal einsetzen ändert nichts mehr.**
		if zweimal := domain.BlockEinsetzen(eingesetzt, probeblock); !bytes.Equal(zweimal, eingesetzt) {
			t.Errorf("%s: zweimal eingesetzt\n%q\n%q", name, zweimal, eingesetzt)
		}

		if zurueck := domain.BlockEntfernen(eingesetzt); string(zurueck) != vorher {
			t.Errorf("%s: zurück %q, vorher %q", name, zurueck, vorher)
		}
	}
}

// **Ein Block mit anderem Inhalt ist veraltet** — etwa nachdem die Anwendung
// verschoben wurde; einsetzen ersetzt ihn an seiner Stelle.
func TestEinVeralteterBlockWirdAnSeinerStelleErsetzt(t *testing.T) {
	t.Parallel()

	alt := domain.BlockEinsetzen([]byte("vorn\n"), "command = \"/alt/mse-mcp\"\n")
	mitHinten := append(append([]byte{}, alt...), []byte("hinten\n")...)

	if stand := domain.BlockStand(mitHinten, "command = \"/neu/mse-mcp\"\n"); stand != domain.BlockVeraltet {
		t.Errorf("Stand %q", stand)
	}

	neu := domain.BlockEinsetzen(mitHinten, "command = \"/neu/mse-mcp\"\n")
	if !bytes.HasPrefix(neu, []byte("vorn\n")) || !bytes.HasSuffix(neu, []byte("hinten\n")) ||
		bytes.Contains(neu, []byte("/alt/")) || !bytes.Contains(neu, []byte("/neu/")) {
		t.Errorf("ersetzt: %q", neu)
	}

	if zurueck := domain.BlockEntfernen(neu); string(zurueck) != "vorn\nhinten\n" {
		t.Errorf("zurück: %q", zurueck)
	}
}

// **Ohne Block ist nichts eingetragen** — und entfernen ändert dann nichts.
func TestOhneBlockIstNichtsEingetragen(t *testing.T) {
	t.Parallel()

	vorher := []byte("export PATH=\"$HOME/.local/bin:$PATH\"\n")

	if stand := domain.BlockStand(vorher, probeblock); stand != domain.BlockFehlt {
		t.Errorf("Stand %q — dieselbe Zeile ohne Marke ist nicht unsere", stand)
	}

	if !bytes.Equal(domain.BlockEntfernen(vorher), vorher) {
		t.Error("ohne Block wurde etwas entfernt")
	}
}

// **Ein halber Block** — Anfang ohne Ende — wird nicht angefasst: er ist
// nicht mehr der, den dieses Programm geschrieben hat.
func TestEinHalberBlockIstFremd(t *testing.T) {
	t.Parallel()

	halb := []byte(domain.BlockAnfang + "\nexport X=1\n")

	if stand := domain.BlockStand(halb, probeblock); stand != domain.BlockFremd {
		t.Errorf("Stand %q", stand)
	}

	if !bytes.Equal(domain.BlockEntfernen(halb), halb) {
		t.Error("ein halber Block wurde angefasst")
	}

	if !bytes.Equal(domain.BlockEinsetzen(halb, probeblock), halb) {
		t.Error("in einen halben Block wurde eingesetzt")
	}
}

// **Eine Anfangsmarke ohne Zeilenende** ist kein Block, den dieses Programm
// geschrieben hat — sie ist fremd.
func TestEinAnfangOhneZeilenendeIstFremd(t *testing.T) {
	t.Parallel()

	if stand := domain.BlockStand([]byte("export X=1\n"+domain.BlockAnfang), "x\n"); stand != domain.BlockFremd {
		t.Errorf("Stand %q", stand)
	}
}

// **Ein Block ohne letztes Zeilenende bekommt eines** — sonst klebte die
// Endmarke an seiner letzten Zeile.
func TestEinBlockOhneZeilenendeBekommtEines(t *testing.T) {
	t.Parallel()

	eingesetzt := domain.BlockEinsetzen(nil, "x")
	if !strings.Contains(string(eingesetzt), "x\n"+domain.BlockEnde) ||
		domain.BlockStand(eingesetzt, "x") != domain.BlockEingetragen {
		t.Errorf("eingesetzt: %q", eingesetzt)
	}

	if uebrig := domain.BlockEntfernen(eingesetzt); len(uebrig) != 0 {
		t.Errorf("übrig: %q", uebrig)
	}
}
