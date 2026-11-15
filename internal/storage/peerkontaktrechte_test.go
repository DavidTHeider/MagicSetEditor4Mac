package storage_test

// **Die Freigaben eines Kontakts** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass ein Recht nicht ohne Bestätigung entsteht und eine Sperre es
// mitnimmt.** Beide Sätze stehen in `domain`; hier wird geprüft, dass der Weg
// in den Bestand sie einhält — gegen den **gespeicherten** Stand und in
// derselben Transaktion, in der geschrieben wird.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein entzogenes Recht ist fort.**
func TestEinEntzogenesRechtIstFort(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	for _, zweck := range []domain.Zweck{
		domain.ZweckTeilen, domain.ZweckHandel,
	} {
		if fehler := ablage.Erlauben(
			ctx, "bgegenueber", zweck, derPeerlauf); fehler != nil {
			t.Fatalf("Erlauben: %v", fehler)
		}
	}

	if fehler := ablage.Entziehen(
		ctx, "bgegenueber", domain.ZweckHandel); fehler != nil {
		t.Fatalf("Entziehen: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if len(gelesen.Rechte) != 1 || gelesen.Rechte[0] != domain.ZweckTeilen {
		t.Fatalf("die Rechte sind %v", gelesen.Rechte)
	}
}

// **Eine Rotation nimmt jede Bestätigung und jedes Recht zurück.**
func TestEineRotationNimmtJedeBestaetigungZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if fehler := ablage.Erlauben(
		ctx, "bgegenueber", domain.ZweckTeilen, derPeerlauf); fehler != nil {
		t.Fatalf("Erlauben: %v", fehler)
	}

	// **Eine Sperre bleibt eine Sperre.** Wer sperrt, hat entschieden; eine
	// Rotation ist kein Grund, jemanden wieder hereinzulassen.
	begegne(t, ablage, ctx, "bgesperrter", "Gesperrter Rechner")

	if fehler := ablage.Stand(
		ctx, "bgesperrter", domain.KontaktGesperrt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if fehler := ablage.AllesZurueck(ctx); fehler != nil {
		t.Fatalf("AllesZurueck: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Stand != domain.KontaktUnbestaetigt {
		t.Errorf("der Stand ist %q", gelesen.Stand)
	}

	if len(gelesen.Rechte) != 0 {
		t.Errorf("die Rechte stehen noch: %v", gelesen.Rechte)
	}

	gesperrt, fehler := ablage.Lesen(ctx, "bgesperrter")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gesperrt.Stand != domain.KontaktGesperrt {
		t.Errorf("die Sperre ist fort: %q", gesperrt.Stand)
	}
}

// **Eine Begegnung macht niemanden bestätigt.**
//
// **Ein Recht an einem unbestätigten Gegenüber wird abgewiesen.**
//
// **Gegen den gespeicherten Stand, nicht gegen einen mitgegebenen.** Ein
// Aufrufer, der behauptet, das Gegenüber sei bestätigt, behauptet — die
// Tabelle weiß es.
func TestEinRechtAnEinemUnbestaetigtenWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Erlauben(
		ctx, "bgegenueber", domain.ZweckHandel, derPeerlauf); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("das Recht kam durch: %v", fehler)
	}

	// **Und ein fremder Zweck ebenso wenig.**
	if fehler := ablage.Erlauben(
		ctx, "bgegenueber", "alles", derPeerlauf); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Zweck kam durch: %v", fehler)
	}
}

// **Eine Sperre räumt die Rechte mit.**
//
// **Sonst stünden sie still da** — unsichtbar, weil `Darf` sie ohnehin nicht
// mehr gelten lässt, und wieder gültig, sobald jemand die Sperre aufhebt.
func TestEineSperreRaeumtDieRechteMit(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if fehler := ablage.Erlauben(
		ctx, "bgegenueber", domain.ZweckTeilen, derPeerlauf); fehler != nil {
		t.Fatalf("Erlauben: %v", fehler)
	}

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktGesperrt); fehler != nil {
		t.Fatalf("Sperren: %v", fehler)
	}

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Entsperren: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if len(gelesen.Rechte) != 0 {
		t.Fatalf("die Rechte kamen zurück: %v", gelesen.Rechte)
	}
}

// **Ein Stand, den es nicht gibt, wird abgewiesen.**
func TestEinStandDenEsNichtGibtWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", "vielleicht"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("er kam durch: %v", fehler)
	}
}

// **Ein zweimal gegebenes Recht ist eines.**
func TestEinZweimalGegebenesRechtIstEines(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	for range 2 {
		if fehler := ablage.Erlauben(
			ctx, "bgegenueber", domain.ZweckTeilen, derPeerlauf); fehler != nil {
			t.Fatalf("Erlauben: %v", fehler)
		}
	}

	// **Und ein zweimal entzogenes ist keines** — das Ziel ist erreicht.
	for range 2 {
		if fehler := ablage.Entziehen(
			ctx, "bgegenueber", domain.ZweckTeilen); fehler != nil {
			t.Fatalf("Entziehen: %v", fehler)
		}
	}

	gelesen, fehler := ablage.Lesen(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if len(gelesen.Rechte) != 0 {
		t.Fatalf("die Rechte sind %v", gelesen.Rechte)
	}
}
