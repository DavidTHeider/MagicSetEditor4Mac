package service_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildquelle ist ein Doppel für [service.Fremdbilder].
type bildquelle struct {
	fund   scryfall.Bildfund
	fehler error
	rufe   int
	art    scryfall.Bildart
}

func (b *bildquelle) Bild(_ context.Context, _ *scryfall.Karte, art scryfall.Bildart) (
	scryfall.Bildfund, error,
) {
	b.rufe++
	b.art = art

	return b.fund, b.fehler
}

// bildkarten ist ein Doppel für [service.Fremdkarten].
type bildkarten struct {
	karte   *scryfall.Karte
	alter   time.Duration
	befunde []domain.Befund
	fehler  error
}

func (b *bildkarten) KarteNachName(_ context.Context, _ string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	return b.karte, b.alter, b.befunde, b.fehler
}

func bildwegFuer(t *testing.T, quelle *bildquelle, karten *bildkarten) (
	weg *service.Bildweg, ordner string,
) {
	t.Helper()

	ort := t.TempDir()

	return service.NeuerBildweg(karten, quelle, ort), ort
}

func TestBildwegHolen(t *testing.T) {
	t.Parallel()

	daten := []byte("\x89PNG\r\n\x1a\nBilddaten")

	gutefund := scryfall.Bildfund{
		Daten: daten, Art: scryfall.BildNormal, Kuenstler: "Christopher Rush",
		Nennung: "Illustration: Christopher Rush. " + scryfall.Copyright,
		Adresse: "https://scryfall.com/card/lea/161",
	}

	t.Run("legt das Bild als Datei ab und gibt den Weg", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: gutefund}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Lightning Bolt"}}
		weg, ordner := bildwegFuer(t, quelle, karten)

		holung, fehler := weg.Holen(t.Context(), "Lightning Bolt", scryfall.BildNormal)
		if fehler != nil {
			t.Fatalf("Holen: %v", fehler)
		}

		// **Die Datei ist der Zweck.** `BildAusDatei` und `BildEinsetzen`
		// nehmen einen Pfad — wer hier nur Bytes zurückgäbe, hätte einen Weg
		// gebaut, der nirgends ankommt.
		abgelegt, lesefehler := os.ReadFile(holung.Pfad) //nolint:gosec // Pfad aus dem Ordner dieses Tests
		if lesefehler != nil {
			t.Fatalf("die abgelegte Datei ist nicht lesbar: %v", lesefehler)
		}

		if !bytes.Equal(abgelegt, daten) {
			t.Errorf("die Datei enthält %q, erwartet die Bildbytes", abgelegt)
		}

		if filepath.Dir(holung.Pfad) != ordner {
			t.Errorf("die Datei liegt in %q, erwartet %q", filepath.Dir(holung.Pfad), ordner)
		}

		if holung.Kuenstler != "Christopher Rush" {
			t.Errorf("Kuenstler = %q", holung.Kuenstler)
		}

		if !strings.Contains(holung.Nennung, "Wizards of the Coast") {
			t.Errorf("Nennung = %q", holung.Nennung)
		}
	})

	t.Run("erkennt JPEG an seinen Bytes", func(t *testing.T) {
		t.Parallel()

		// **Die Endung kommt aus dem Inhalt.** `art_crop` liefert JPEG, die
		// `png`-Größe PNG — und eine Adresse kann eine Weiterleitung sein.
		jpeg := []byte("\xff\xd8\xff\xe0 und was folgt")
		quelle := &bildquelle{fund: scryfall.Bildfund{Daten: jpeg, Nennung: "n"}}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Foto"}}
		weg, _ := bildwegFuer(t, quelle, karten)

		holung, fehler := weg.Holen(t.Context(), "Foto", scryfall.BildNormal)
		if fehler != nil {
			t.Fatalf("Holen: %v", fehler)
		}

		if filepath.Ext(holung.Pfad) != ".jpg" {
			t.Errorf("Pfad = %q, erwartet eine .jpg-Endung", holung.Pfad)
		}
	})

	t.Run("gibt PNG die Endung .png", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: gutefund}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Bild"}}
		weg, _ := bildwegFuer(t, quelle, karten)

		holung, fehler := weg.Holen(t.Context(), "Bild", scryfall.BildNormal)
		if fehler != nil {
			t.Fatalf("Holen: %v", fehler)
		}

		if filepath.Ext(holung.Pfad) != ".png" {
			t.Errorf("Pfad = %q, erwartet eine .png-Endung", holung.Pfad)
		}
	})

	t.Run("reicht die gewünschte Art durch", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: gutefund}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Egal"}}
		weg, _ := bildwegFuer(t, quelle, karten)

		if _, fehler := weg.Holen(t.Context(), "Egal", scryfall.BildAusschnitt); fehler != nil {
			t.Fatalf("Holen: %v", fehler)
		}

		if quelle.art != scryfall.BildAusschnitt {
			t.Errorf("die Quelle bekam %q, erwartet den Ausschnitt", quelle.art)
		}
	})

	t.Run("weist ein Bild ohne Nennung ab", func(t *testing.T) {
		t.Parallel()

		// **Der Kern.** Ein Bild ohne den Satz, der danebenstehen muss, darf
		// die Anwendung nicht verlassen — auch nicht als Datei auf der Platte,
		// denn von dort holt es sich der Nächste ohne die Bedingung.
		quelle := &bildquelle{fund: scryfall.Bildfund{Daten: daten, Nennung: ""}}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Ohne"}}
		weg, ordner := bildwegFuer(t, quelle, karten)

		_, fehler := weg.Holen(t.Context(), "Ohne", scryfall.BildNormal)
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}

		eintraege, lesefehler := os.ReadDir(ordner)
		if lesefehler != nil {
			t.Fatalf("den Ordner lesen: %v", lesefehler)
		}

		if len(eintraege) != 0 {
			t.Errorf("es liegen %d Dateien im Ordner — nichts durfte geschrieben werden",
				len(eintraege))
		}
	})

	t.Run("weist ein leeres Bild ab", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: scryfall.Bildfund{Nennung: "irgendwas"}}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Leer"}}
		weg, _ := bildwegFuer(t, quelle, karten)

		if _, fehler := weg.Holen(t.Context(), "Leer", scryfall.BildNormal); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}
	})

	t.Run("reicht den Fehler der Kartensuche durch", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: gutefund}
		karten := &bildkarten{fehler: domain.ErrNotFound}
		weg, _ := bildwegFuer(t, quelle, karten)

		_, fehler := weg.Holen(t.Context(), "Gibtesnicht", scryfall.BildNormal)
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}

		if quelle.rufe != 0 {
			t.Errorf("das Bild wurde geholt, obwohl die Karte fehlt (%d Rufe)", quelle.rufe)
		}
	})

	t.Run("reicht den Fehler des Bildabrufs durch", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fehler: domain.ErrNotFound}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Ohne Bild"}}
		weg, _ := bildwegFuer(t, quelle, karten)

		if _, fehler := weg.Holen(t.Context(), "Ohne Bild", scryfall.BildKlein); !errors.Is(
			fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	t.Run("meldet das Alter des Standes", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: gutefund}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Alt"}, alter: 5 * time.Hour}
		weg, _ := bildwegFuer(t, quelle, karten)

		holung, fehler := weg.Holen(t.Context(), "Alt", scryfall.BildNormal)
		if fehler != nil {
			t.Fatalf("Holen: %v", fehler)
		}

		if holung.Alter != 5*time.Hour {
			t.Errorf("Alter = %v, erwartet 5h", holung.Alter)
		}
	})

	t.Run("meldet einen unbeschreibbaren Ordner", func(t *testing.T) {
		t.Parallel()

		quelle := &bildquelle{fund: gutefund}
		karten := &bildkarten{karte: &scryfall.Karte{Name: "Egal"}}
		weg := service.NeuerBildweg(karten, quelle,
			filepath.Join(t.TempDir(), "gibtesnicht", "auchnicht"))

		if _, fehler := weg.Holen(t.Context(), "Egal", scryfall.BildNormal); fehler == nil {
			t.Fatal("der unbeschreibbare Ordner blieb ohne Fehler")
		}
	})
}

// **Die toleranten Befunde der Kartenauskunft gehen mit** — Nachlauf über
// Phase 4.5, Befund 10.
//
// # Warum das kein Randfall ist
//
// §16 sagt: „Nichts wird still übernommen" und „ein unbekannter
// Aufzählungswert wird als **Befund** geführt und nicht verschluckt."
// `KarteNachName` liefert genau solche Befunde — ein Feld, das der Anbieter
// umbenannt hat, ein Aufzählungswert, den dieses Programm nicht kennt.
//
// **`Bildweg.Holen` warf sie mit `_` weg.** Der Feldabgleich reicht dieselben
// Befunde durch (`scryfallabgleich.go`); der Bildweg ist derselbe Ruf an
// dieselbe Quelle, und wer ihn nimmt, erfuhr nichts. Dieselbe Gestalt hat der
// Nachlauf in `zeileDeuten` gefunden und dort behoben — hier stand sie noch.
func TestBildwegReichtDieBefundeDurch(t *testing.T) {
	t.Parallel()

	befund := domain.Befund{
		Ort: "layout", Meldung: "unbekannter Wert »seltsam«", Art: domain.ArtUmgebung,
	}

	quelle := &bildquelle{fund: scryfall.Bildfund{
		Daten: []byte("\x89PNG\r\n\x1a\nBilddaten"), Art: scryfall.BildNormal,
		Kuenstler: "Christopher Rush",
		Nennung:   "Illustration: Christopher Rush. " + scryfall.Copyright,
		Adresse:   "https://scryfall.com/card/lea/161",
	}}
	karten := &bildkarten{
		karte:   &scryfall.Karte{Name: "Lightning Bolt"},
		befunde: []domain.Befund{befund},
	}

	weg, _ := bildwegFuer(t, quelle, karten)

	holung, fehler := weg.Holen(t.Context(), "Lightning Bolt", scryfall.BildNormal)
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if len(holung.Befunde) != 1 || holung.Befunde[0].Meldung != befund.Meldung {
		t.Errorf("die Befunde lauten %v", holung.Befunde)
	}
}

// **Und ohne Befunde ist die Liste leer, nicht nil.**
//
// Die Gegenprobe zur Zeile darüber: eine Zusicherung auf „mindestens einer"
// wäre auch dann grün, wenn immer derselbe Befund entstünde.
func TestBildwegOhneBefundeGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	quelle := &bildquelle{fund: scryfall.Bildfund{
		Daten: []byte("\x89PNG\r\n\x1a\nBilddaten"), Art: scryfall.BildNormal,
		Kuenstler: "Christopher Rush",
		Nennung:   "Illustration: Christopher Rush. " + scryfall.Copyright,
		Adresse:   "https://scryfall.com/card/lea/161",
	}}
	karten := &bildkarten{karte: &scryfall.Karte{Name: "Lightning Bolt"}}

	weg, _ := bildwegFuer(t, quelle, karten)

	holung, fehler := weg.Holen(t.Context(), "Lightning Bolt", scryfall.BildNormal)
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if len(holung.Befunde) != 0 {
		t.Errorf("die Befunde lauten %v", holung.Befunde)
	}
}
