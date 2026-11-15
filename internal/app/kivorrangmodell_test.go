package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Welches Ollama-Modell der Vorrang fragt** — #2054, aus AI-02 (#1743).
// Ollama nennt das zuletzt gezogene zuerst; ohne Wahl nimmt der Vorrang es,
// mit Wahl das gewählte, und ein gewähltes, das fehlt, wird gemeldet.
func TestDerVorrangNimmtDasGewaehlteModell(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, gewaehlt, modell, meldung string
	}{
		{name: "ohne Wahl das erste, das Ollama nennt", modell: "qwen3:0.6b"},
		{name: "mit Wahl das gewählte", gewaehlt: "gemma3:1b", modell: "gemma3:1b"},
		{name: "ein gewähltes, das fehlt, wird gemeldet", gewaehlt: "mistral:7b", meldung: "mistral:7b"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			welt := einekiwelt(t)
			welt.laufzeit.modelle = []domain.KIModell{einKoenner("qwen3:0.6b"), einKoenner("gemma3:1b")}

			if fall.gewaehlt != "" {
				if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
					domain.KIVorrangmodell, fall.gewaehlt); fehler != nil {
					t.Fatalf("%v", fehler)
				}
			}

			antwort := vorrang(t, welt, "name")

			if fall.meldung != "" {
				if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, fall.meldung) {
					t.Fatalf("erwartet eine Meldung über %q: %+v", fall.meldung, antwort)
				}

				if len(welt.laufzeit.auftraege) != 0 {
					t.Errorf("statt des gewählten wurde ein anderes gefragt: %v", welt.laufzeit.auftraege)
				}

				return
			}

			// **`adressen` hält den gefragten Alias** — `antwort.Modell` ist,
			// was das Doppel als Antwortenden nennt, und das steht fest.
			if antwort.Fehler != nil || antwort.Anbieter != "ollama" ||
				len(welt.laufzeit.adressen) != 1 || welt.laufzeit.adressen[0] != fall.modell {
				t.Fatalf("gefragt %v: %+v", welt.laufzeit.adressen, antwort)
			}
		})
	}
}

// **Der Katalog sagt, wen der Vorrang fragt** — das zweite Kriterium von
// #2054: ohne Wahl steht es am Formular, welches Modell gefragt wird.
func TestDerKatalogNenntDieReiheDesVorrangs(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.modelle = []domain.KIModell{einKoenner("qwen3:0.6b"), einKoenner("gemma3:1b")}

	katalog := welt.anwendung.KIModelle(probekontext(t))
	if katalog.Fehler != nil || strings.Join(katalog.Vorrang, ",") != "ollama/qwen3:0.6b" {
		t.Fatalf("Reihe %v, Fehler %+v", katalog.Vorrang, katalog.Fehler)
	}

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		domain.KIVorrangmodell, "mistral:7b"); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	katalog = welt.anwendung.KIModelle(probekontext(t))
	if len(katalog.Vorrang) != 0 || !strings.Contains(strings.Join(katalog.Befunde, " "), "mistral:7b") {
		t.Fatalf("Reihe %v, Befunde %v", katalog.Vorrang, katalog.Befunde)
	}

	// **Und eigens für das Feld** — Z4 aus dem Zweitblick über Phase 6.11:
	// unter allen Befunden fände das Formular den seinen nicht.
	if len(katalog.Vorrangbefunde) != 1 || !strings.Contains(katalog.Vorrangbefunde[0], "mistral:7b") {
		t.Errorf("Vorrangbefunde %v", katalog.Vorrangbefunde)
	}
}
