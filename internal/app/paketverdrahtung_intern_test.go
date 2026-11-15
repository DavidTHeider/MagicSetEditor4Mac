// Der Wächter über die Naht des Paketkatalogs (#54).
//
// **Dieselbe Frage wie bei #58 und #206**: nicht, ob der Katalog gebaut ist,
// sondern ob die laufende Anwendung ihn führt — und zwar über **dieselben**
// Wurzeln, aus denen Stil, Bild und Schrift kommen. Zwei Listen wären zwei
// Wahrheiten darüber, was installiert ist.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

// **Der Start führt den Katalog, und er kennt die Saat.**
//
// # Die Gegenprobe, die dahintersteht
//
// Wer in `app.go` die Wurzeln des Verwalters durch etwas anderes ersetzt —
// etwa nur das Bezugsverzeichnis —, macht diesen Lauf rot: die Saat stünde
// dann in keinem Katalog, und die Paketseite zeigte eine leere Liste, obwohl
// fünf Pakete installiert sind.
func TestDerStartFuehrtDenPaketkatalog(t *testing.T) {
	anwendung := mitSaat(t)

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Pakete: %+v", antwort.Fehler)
	}

	if len(antwort.Befunde) > 0 {
		t.Fatalf("Befunde: %v", antwort.Befunde)
	}

	arten := map[string]int{}
	for _, paket := range antwort.Pakete {
		arten[paket.Art]++
	}

	if arten["mse-game"] < 1 || arten["mse-style"] < 1 {
		t.Fatalf("der Katalog kennt die Saat nicht: %+v", antwort.Pakete)
	}

	for _, paket := range antwort.Pakete {
		if paket.Name == "probe" && !strings.HasSuffix(paket.Pfad, "probe.mse-game") {
			t.Errorf("das Spielpaket liegt an %q", paket.Pfad)
		}
	}
}

// **Ohne Saat und ohne bezogenes Paket bleibt der Katalog leer — und heil.**
//
// Der frische Rechner: keine Liste, kein Fehler, keine Befunde. Ein Katalog,
// der auf einem leeren Verzeichnis meldete, machte aus dem Normalfall einen
// Zwischenfall.
func TestOhneVorlagenBleibtDerKatalogLeer(t *testing.T) {
	anwendung := ohneSaat(t)

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Pakete: %+v", antwort.Fehler)
	}

	if len(antwort.Pakete) != 0 {
		t.Errorf("Pakete: %+v", antwort.Pakete)
	}

	if len(antwort.Befunde) != 0 {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}
}

// **Ein frisch installiertes Paket ist sofort da — ohne Neustart.**
//
// # Woher der Fall kommt
//
// Aus dem Zweitblick über Phase 7. `Paketwurzeln` meldet, **wo Pakete liegen**;
// ein leeres Paketverzeichnis trägt keines und fiel heraus. Wer danach ein
// Paket installierte, legte es an einen Ort, den weder Bildquelle noch
// Stilspeicher kannten.
//
// **Die Gegenprobe, die dahintersteht:** wer in `app.go` den verwalteten Ort
// wieder aus den Wurzeln nimmt, macht diesen Lauf rot — der Stilspeicher findet
// den eben installierten Stil dann nicht.
func TestEinInstalliertesPaketIstSofortDa(t *testing.T) {
	verwaltet := t.TempDir()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.paketort = func() (string, error) { return verwaltet, nil }
	anwendung.saatort = func() string { return "" }

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	// **Der Bestand ist leer, das Paketverzeichnis auch** — der frische Rechner.
	wurzel, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	for _, name := range []string{"probe.mse-game", "probe-schlicht.mse-style"} {
		quelle := filepath.Join(wurzel, "saat", name)

		if antwort := anwendung.PaketInstallieren(t.Context(), quelle); antwort.Fehler != nil {
			t.Fatalf("%s installieren: %+v", name, antwort.Fehler)
		}
	}

	// **Und jetzt, ohne Neustart:** der Stilspeicher findet ihn.
	antwort := anwendung.SetOeffnen(t.Context(), saatset(t))
	if antwort.Fehler != nil {
		t.Fatalf("öffnen: %+v", antwort.Fehler)
	}

	karten := anwendung.Karten(t.Context(), antwort.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	if nennt(vorschau.Vorschau.Befunde, "Paketverzeichnis") ||
		nennt(vorschau.Vorschau.Befunde, "in keinem Paketverzeichnis") {
		t.Errorf("das installierte Paket wurde nicht gefunden: %v", vorschau.Vorschau.Befunde)
	}
}
