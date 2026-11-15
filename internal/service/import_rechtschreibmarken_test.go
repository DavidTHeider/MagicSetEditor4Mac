package service_test

// **Rechtschreibmarken sind kein Wert** — F05
// (#2247).
//
// # Wie das gefunden wurde
//
// Im Setdatei-Rundlauf gegen MSE2 (`internal/flaechen/setrundlauf_fremd_test.go`):
// MSE2 schrieb die eigene Saat mit `write_set_file` zurück, und in drei Werten
// standen danach `<error-spelling:en_US>Ein</error-spelling:en_US>` und
// `<error-spelling:en_US:/magic.mse-game/dictionary/magic-words>` — die
// Unterringelung seines Editors, weil das englische Wörterbuch kein Deutsch
// kennt. Diese Anwendung las sie als Teil des Werts; der Volltext fand damit
// „spelling", „dictionary" und „magic" in jeder markierten Karte.
//
// # Was MSE2 mit ihnen tut
//
// `check_spelling` nimmt die alten Marken heraus, bevor es neu prüft
// (`src/script/functions/spelling.cpp`, Zeilen 57–71, Fassung f9d9356 vom
// 2021-07-06), und die Vorlagen nehmen sie vor dem Zeichnen heraus
// (`magic.mse-game/script`, `remove_tag@(tag: "<error-spelling")`). Sie sind
// eine Auskunft an den Editor und gehören nicht zum Text.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestRechtschreibmarkenAusMSE2WerdenNichtGelesen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "markiert.mse-set")
	if fehler := os.MkdirAll(ort, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	set := "mse version: 2.0.2\n" +
		"game: probe\n" +
		"set info:\n" +
		"\ttitle: <error-spelling:en_US>Grüße</error-spelling:en_US> aus Köln\n" +
		"card:\n" +
		"\tname: Feuerball\n" +
		"\trule_text:\n" +
		"\t\t<error-spelling:en_US:/magic.mse-game/dictionary/magic-words>Erste</error-spelling:en_US:/magic.mse-game/dictionary/magic-words> Zeile\n" +
		"\t\tmit <b>Fett</b> und \\<kein Tag>\n"

	if fehler := os.WriteFile(filepath.Join(ort, "set"), []byte(set), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	speicher := &speicherDouble{}
	if _, fehler := service.NeuerImporteur(speicher).Set(t.Context(), ort); fehler != nil {
		t.Fatalf("Import: %v", fehler)
	}

	regel, _ := speicher.set.Cards[0].Value("rule text")
	if soll := "Erste Zeile\nmit <b>Fett</b> und \\<kein Tag>"; regel != soll {
		t.Errorf("rule text = %q, erwartet %q — die Marke ist Teil des Werts geblieben", regel, soll)
	}

	for _, wert := range speicher.set.Values {
		if wert.Field == "title" && wert.Text != "Grüße aus Köln" {
			t.Errorf("title = %q — die Marke ist Teil des Setwerts geblieben", wert.Text)
		}
	}
}
