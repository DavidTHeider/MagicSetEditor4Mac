package app

// **Ein Testlauf bekommt keinen schreibenden Ort im echten Zuhause** — A01
// (#2215).
//
// # Der Anlass
//
// Der Audit zu 0.8.0 sah während `make vorpruefung` zwei Drosseldateien und
// einen Scryfall-Eintrag im echten Anwendungsverzeichnis wachsen. Die
// Vorprüfung fuhr die Go-Läufe ohne eigenes Zuhause, und jedes Paket ohne
// den Boden aus `echteszuhause_test.go` schrieb dorthin, wo die Anwendung
// dieses Rechners ihre Daten hat.
//
// # Warum hier und nicht nur in den Läufen
//
// **Weil ein Lauf von Hand keinen Riegel fährt.** `scripts/ci.sh` und
// `scripts/vorpruefung.sh` setzen das Zuhause; `go test ./cmd/mse-rest/`
// tut es nicht. Die Klasse schließt erst, wo die Orte entstehen: an den
// fünf Stellen, die schreiben — Bestand, Scryfall, Teilungen, Bildvorrat,
// Downloads. **Sie scheitern dann laut**, und der Lauf sagt, welcher Test
// sein Zuhause vergessen hat (Zuordnung am Aufrufer statt an einer Datei).
//
// # Was er nicht sieht
//
// Orte, die nur gelesen oder angezeigt werden (`schwaerze`, der
// MSE2-Ort, die Betriebsauskunft) — sie bleiben, wie sie sind.

import (
	"errors"
	"os/user"
	"testing"
)

func TestImEchtenZuhauseGibtEsKeinenSchreibendenOrt(t *testing.T) {
	wer, fehler := user.Current()
	if fehler != nil || wer.HomeDir == "" {
		t.Skipf("ohne Benutzereintrag gibt es kein echtes Zuhause: %v", fehler)
	}

	t.Setenv("HOME", wer.HomeDir)

	orte := map[string]func() (string, error){
		"Datenbankpfad":       Datenbankpfad,
		"Scryfallverzeichnis": Scryfallverzeichnis,
		"Teilverzeichnis":     Teilverzeichnis,
		"zwischenspeicherort": zwischenspeicherort,
		"downloadortEcht":     downloadortEcht,
	}

	for name, ort := range orte {
		t.Run(name, func(t *testing.T) {
			pfad, fehler := ort()
			if !errors.Is(fehler, errEchtesZuhauseImTestlauf) {
				t.Errorf("im echten Zuhause gab %s %q, %v — erwartet war die Abweisung", name, pfad, fehler)
			}
		})
	}

	// **Und die Absprache der Drossel schaltet sich ab**, statt dort eine
	// Datei anzulegen — genau die zwei, die der Audit sah.
	if ort := ScryfallAbspracheort(); ort != "" {
		t.Errorf("die Drosselabsprache liegt im echten Zuhause: %s", ort)
	}
}

func TestEinEigenesZuhauseGibtSeineOrte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	for name, ort := range map[string]func() (string, error){
		"Datenbankpfad":       Datenbankpfad,
		"zwischenspeicherort": zwischenspeicherort,
		"downloadortEcht":     downloadortEcht,
	} {
		if pfad, fehler := ort(); fehler != nil || pfad == "" {
			t.Errorf("%s gab %q, %v", name, pfad, fehler)
		}
	}
}

// **Frei ist nur ein Zuhause unter der Zwischenablage** — dort legen
// `t.TempDir`, die Böden der TestMain und `mktemp` in `scripts/ci.sh` ihre
// an. Alles andere ist für einen Testlauf ein fremdes Zuhause, auch eines,
// das es nicht gibt: **fail closed**, und ohne `os/user` — den Systemnamen
// hält `TestKeinSystemnameFliesstStillEin` aus dem Produktivcode heraus.
func TestEinTestlaufBekommtNurUnterDerZwischenablageEinZuhause(t *testing.T) {
	faelle := []struct {
		name         string
		heim, ablage string
		frei         bool
	}{
		{"darunter", "/var/folders/x/T/mse-1", "/var/folders/x/T", true},
		{"darunter, Ablage mit Schrägstrich", "/var/folders/x/T/mse-1", "/var/folders/x/T/", true},
		{"die Ablage selbst", "/var/folders/x/T", "/var/folders/x/T", false},
		{"das echte", "/Users/jemand", "/var/folders/x/T", false},
		{"nur ähnlich benannt", "/var/folders/x/Tmp/mse", "/var/folders/x/T", false},
		// **Ohne Zuhause entscheidet das Betriebssystem** — es scheitert ohne
		// zu schreiben, und der Fehlerweg „ohne Benutzerverzeichnis" bleibt
		// prüfbar (Befund 8 des Zweitblicks über Phase A).
		{"leer", "", "/var/folders/x/T", true},
		// **Ohne Zwischenablage gibt es kein Darunter** — dann ist kein
		// Zuhause frei.
		{"ohne Ablage", "/var/folders/x/T/mse-1", "", false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if got := zuhauseFreiFuerTestlauf(fall.heim, fall.ablage); got != fall.frei {
				t.Errorf("zuhauseFreiFuerTestlauf(%q, %q) = %v", fall.heim, fall.ablage, got)
			}
		})
	}
}
