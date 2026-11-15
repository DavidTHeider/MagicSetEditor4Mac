//go:build fremdpakete

// **Die Verweisskripte an einem echten Spielpaket** — Werkstück D von Phase
// 4.11 (#1342).
//
// # Was dieser Lauf hält, das kein Doppel halten kann
//
// **Die Zahlen, mit denen die Entscheidungen begründet sind.** Dass eine
// Parameterart ohne Verweisskripte die schlichte Form bekommt, ist nur so
// lange eine sinnvolle Regel, wie es solche Arten wirklich gibt.
//
//	MSE_FREMDPAKETE=~/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/msepaket/ -run Verweise -v
package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDieVerweiseVonMagic(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	arten := paket.Schluesselwortparameterarten()

	mitVerweisen, skripte := 0, 0

	var mana msepaket.Schluesselwortparameterart

	for _, art := range arten {
		if len(art.Verweisskripte) > 0 {
			mitVerweisen++
			skripte += len(art.Verweisskripte)
		}

		if art.Name == "mana" {
			mana = art
		}
	}

	t.Logf("%d Parameterarten, %d mit Verweisen, %d Skripte",
		len(arten), mitVerweisen, skripte)

	// **Die Art `mana` bringt drei mit** — gezählt am 2026-08-31 in
	// `keywords_en`. Als Gleichheit und nicht als Untergrenze: eine
	// Untergrenze bliebe grün, wenn zwei wegfielen.
	if len(mana.Verweisskripte) != 3 {
		t.Fatalf("die Art mana hat %d Verweisskripte", len(mana.Verweisskripte))
	}

	if mana.Verweisskripte[0].Name != "normal" ||
		mana.Verweisskripte[0].Skript != `\{{input}\}` {
		t.Errorf("das erste lautet %+v", mana.Verweisskripte[0])
	}

	// **Und es gibt Arten ohne** — sonst wäre die schlichte Form `{paramN}`
	// eine Regel ohne Fall.
	if mitVerweisen == len(arten) {
		t.Error("jede Parameterart trägt Verweisskripte — dann hat {paramN} keinen Fall")
	}
}
