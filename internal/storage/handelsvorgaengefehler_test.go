package storage_test

// **Was die Handelsablage abweist** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Warum die Fehlerwege eine eigene Datei haben
//
// **Weil sie eine andere Frage beantworten** (§2): nebenan steht, **was** die
// Ablage hält; hier steht, was geschieht, wenn der Kontext kein Profil trägt,
// die Datenbank zu ist oder ein von Hand geänderter Bestand etwas enthält,
// das kein Angebot ist.
//
// **Der letzte Fall ist der, für den es §8 gibt**: fremde Daten sind der
// Normalfall, und ein Befund ist kein Absturz.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Profil im Kontext geschieht nichts.**
func TestDieHandelsablageBrauchtEinProfil(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueHandelsablage(frischeDatenbank(t))
	ohne := t.Context()

	if fehler := ablage.Schreiben(ohne, einVorgangsangebot()); fehler == nil {
		t.Error("geschrieben wurde ohne Profil")
	}

	if _, fehler := ablage.Vorgang(ohne, "h1"); fehler == nil {
		t.Error("gelesen wurde ohne Profil")
	}

	if _, fehler := ablage.Revisionen(ohne, "h1"); fehler == nil {
		t.Error("die Kette kam ohne Profil")
	}

	if _, fehler := ablage.Letzte(ohne); fehler == nil {
		t.Error("die Liste kam ohne Profil")
	}
}

// **Eine geschlossene Datenbank meldet sich.**
func TestDieHandelsablageMeldetEineGeschlosseneDatenbank(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage, ctx := storage.NeueHandelsablage(db), probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := ablage.Schreiben(ctx, einVorgangsangebot()); fehler == nil {
		t.Error("geschrieben wurde in eine geschlossene Datenbank")
	}

	if _, fehler := ablage.Vorgang(ctx, "h1"); fehler == nil {
		t.Error("gelesen wurde aus einer geschlossenen Datenbank")
	}

	if _, fehler := ablage.Revisionen(ctx, "h1"); fehler == nil {
		t.Error("die Kette kam aus einer geschlossenen Datenbank")
	}

	if _, fehler := ablage.Letzte(ctx); fehler == nil {
		t.Error("die Liste kam aus einer geschlossenen Datenbank")
	}
}

// **Was kein Angebot ist, wird als Befund gemeldet und nicht geraten.**
//
// **Von Hand geänderte Bestände sind der Normalfall** (§8) — ein `sqlite3`
// auf der Kommandozeile genügt. Die Anwendung darf daran nicht umfallen, und
// sie darf sich auch nichts zusammenreimen.
func TestEineUnlesbareRevisionIstEinBefund(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage, ctx := storage.NeueHandelsablage(db), probekontext(t)

	if fehler := ablage.Schreiben(ctx, einVorgangsangebot()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if _, fehler := db.ExecContext(ctx,
		`UPDATE handelsrevisionen SET angebot = '{kein json'`,
	); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if _, fehler := ablage.Vorgang(ctx, "h1"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("beim Lesen kam %v", fehler)
	}

	if _, fehler := ablage.Revisionen(ctx, "h1"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("bei der Kette kam %v", fehler)
	}

	if _, fehler := ablage.Letzte(ctx); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("bei der Liste kam %v", fehler)
	}
}
