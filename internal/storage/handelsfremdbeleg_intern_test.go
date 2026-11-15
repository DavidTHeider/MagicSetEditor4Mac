package storage

// **Ein Fremdbeleg, der sich nicht nachschlagen lässt** — TRADE-03
// (#1717).
//
// **Innen und mit geschlossener Datenbank**: der Vergleich läuft nur, wenn
// das Einfügen eben einen vorhandenen fand — ein Fehler dort und nicht beim
// Einfügen lässt sich an einer echten Datei nicht herstellen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinUnlesbarerFremdbelegIstKeinKonflikt(t *testing.T) {
	fehler := derselbeFremdbeleg(t.Context(), geschlosseneDatenbank(t), 1,
		domain.Fremdbeleg{Vorgang: "v"})
	if fehler == nil || errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("derselbeFremdbeleg: %v", fehler)
	}
}
