package domain_test

// **Scrydex vergibt eigene Kartenkennungen** — Werkstück A5 von Phase 5.3
// (#1444).
//
// # Warum das ein eigenes System ist
//
// **Weil Scrydex nicht Scryfalls Kennungen führt.** Sein Preisverlauf hängt
// an `/magicthegathering/v1/cards/<id>/price_history`, und `<id>` ist **seine**
// Kennung. Wer dort eine Scryfall-UUID einsetzte, fragte nach einer Karte, die
// es dort nicht gibt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Scrydex-Kennung ist gültig und nennt genau einen Druck.**
func TestEineScrydexkennungIstGueltig(t *testing.T) {
	t.Parallel()

	kennung := domain.Druckkennung{
		System: domain.SystemScrydex, Semantik: domain.SemantikDruck,
		Wert: "mtg-m11-149",
	}

	if befunde := kennung.Validate(); len(befunde) != 0 {
		t.Errorf("die Kennung wurde abgewiesen: %v", befunde)
	}

	if !kennung.Eindeutig() {
		t.Error("die Kennung gilt nicht als eindeutig")
	}
}

// **Sie ist keine UUID, und das ist Absicht.**
//
// **Die Gegenprobe zur Regel in `Validate`**: Scryfall und MTGJSON vergeben
// UUIDs, und dort wird die Gestalt geprüft. Scrydex vergibt, was es will —
// gemessen am 2026-09-02 sah eine Kennung wie `mtg-m11-149` aus. Wer hier
// eine UUID verlangte, wiese jede echte Kennung ab.
func TestEineScrydexkennungBrauchtKeineUUID(t *testing.T) {
	t.Parallel()

	fuerScryfall := domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
		Wert: "mtg-m11-149",
	}

	if len(fuerScryfall.Validate()) == 0 {
		t.Error("Scryfall nahm eine Kennung ohne UUID-Gestalt an")
	}
}

// **Die Aufzählung der Systeme ist eine Kopie.**
//
// # Warum das geprüft wird
//
// **Weil ein zweiter Leser sie sonst ändern könnte.** Sie steht ein zweites
// Mal in der Wanderung (`CHECK (system IN …)`), und ein Wächter in
// `internal/storage` hält beide zusammen — wer hier hineinschriebe, verschöbe
// die eine Hälfte still.
func TestDieSystemaufzaehlungIstEineKopie(t *testing.T) {
	t.Parallel()

	erste := domain.Kennungssysteme()

	if len(erste) == 0 {
		t.Fatal("es gibt keine Kennungssysteme")
	}

	erste[0] = "erfunden"

	if domain.Kennungssysteme()[0] == "erfunden" {
		t.Error("die Aufzählung ließ sich von außen ändern")
	}
}
