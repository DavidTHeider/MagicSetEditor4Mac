package storage_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineKarte legt ein Set mit einer Karte an und gibt deren Kennung.
func eineKarte(t *testing.T, db *sql.DB, werte ...domain.Value) int64 {
	t.Helper()

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{{Values: werte}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return karten[0].ID
}

// **Schreiben gibt den vorigen Wert — in einem Zug.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Zweitblick über Phase 9. Der MCP-Server las den Vorwert und schrieb
// **danach**, in einem getrennten Aufruf. Im ausdrücklich normalen
// Zwei-Schreiber-Betrieb ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md))
// kann das Fenster dazwischen denselben Wert ändern — und die zugesagte
// Rücknahme führte dann auf einen Wert, den es nie gab.
//
// Ein Zug heißt: eine Transaktion.
func TestSchreibenGibtDenVorigenWert(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	vorher, hatte, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, "name", "Eisball")
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if !hatte || vorher != "Flammenball" {
		t.Errorf("vorher: %q (da: %v)", vorher, hatte)
	}

	// Und ein Feld, das es nicht gab, sagt das — statt einen leeren Wert zu
	// behaupten, den es nie gab.
	vorher, hatte, err = schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, "rule text", "Wachsamkeit")
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if hatte || vorher != "" {
		t.Errorf("vorher: %q (da: %v) — das Feld gab es nicht", vorher, hatte)
	}
}

// **Ein Feld, das es vorher nicht gab, lässt sich wieder entfernen.**
//
// Ohne das ist die Rücknahme eine halbe Zusage: `vorherDa: false` sagt, dass
// das Feld fehlte, und `FeldSetzen` konnte es nur **leeren**. Eine leere Zeile
// ist aber nicht dasselbe wie keine Zeile — die Domäne unterscheidet beides
// ausdrücklich (`Card.Value` gibt zwei Werte zurück).
func TestEinFeldLaesstSichEntfernen(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, "rule text", "Wachsamkeit"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	vorher, hatte, _, err := schreiber.FeldEntfernen(probekontext(t), karteID, "rule text")
	if err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if !hatte || vorher != "Wachsamkeit" {
		t.Errorf("vorher: %q (da: %v)", vorher, hatte)
	}

	karte, err := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if _, da := karte.Value("rule text"); da {
		t.Error("das Feld ist noch da — geleert ist nicht entfernt")
	}
}

// Eine Karte, die es nicht gibt, ist auch hier ein ErrNotFound.
func TestEineFehlendeKarteWirdBeimAendernBenannt(t *testing.T) {
	db := frischeDatenbank(t)
	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), 999, "name", "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler: %v", err)
	}

	if _, _, _, err := schreiber.FeldEntfernen(probekontext(t), 999, "name"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler: %v", err)
	}
}

// **Zwei Schreiber auf demselben Feld geben zwei verschiedene Vorwerte.**
//
// Das ist die Zusage, um die es geht: wer schreibt, bekommt den Wert, den *er*
// überschrieben hat — nicht den, den jemand anders vorfand. Ohne die
// Transaktion wären beide Vorwerte derselbe, und eine der beiden Rücknahmen
// führte auf einen Stand, den es nie gab.
func TestZweiSchreiberBekommenVerschiedeneVorwerte(t *testing.T) {
	// **Eine Datei und zwei Verbindungen, nicht der Arbeitsspeicher.**
	//
	// `frischeDatenbank` gibt eine Datenbank im Arbeitsspeicher mit
	// `SetMaxOpenConns(1)` — zwei Aufrufer liefen darauf reihum, und der Test
	// prüfte die Zusage nicht, die er behauptet. Der Nachlauf zum Zweitblick
	// über Phase 9 hat es angestrichen.
	ort := filepath.Join(t.TempDir(), "bestand.db")

	erste, zweite := verbindung(t, ort), verbindung(t, ort)

	if err := storage.Migriere(probekontext(t), erste, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	karteID := eineKarte(t, erste, domain.Value{Field: "name", Text: "Anfang"})

	schreiber := []*storage.Kartenschreiber{
		storage.NeuerKartenschreiber(erste),
		storage.NeuerKartenschreiber(zweite),
	}

	var (
		warten  sync.WaitGroup
		schloss sync.Mutex
		gesehen []string
	)

	for lauf := range 2 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			vorher, _, err := schreiber[lauf].FeldSetzenMitVorwert(
				probekontext(t), karteID, "name", "Lauf "+string(rune('A'+lauf)))
			if err != nil {
				t.Errorf("schreiben: %v", err)

				return
			}

			schloss.Lock()
			gesehen = append(gesehen, vorher)
			schloss.Unlock()
		}()
	}

	warten.Wait()

	if len(gesehen) != 2 || gesehen[0] == gesehen[1] {
		t.Errorf("Vorwerte: %v — zwei Schreiber sahen denselben Stand", gesehen)
	}
}
