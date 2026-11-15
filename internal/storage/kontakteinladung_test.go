package storage_test

// **Eine Einladung annehmen** — NET-01
// (#1705).
//
// # Was hier gemessen wird
//
// **Dass eine Annahme ganz oder gar nicht geschieht** (AK-04): Kontakt,
// Freigabe und verbrauchte Nonce stehen zusammen da oder keines davon.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func eineAufnahme(nonce string, stand domain.Kontaktstand,
	rechte ...domain.Zweck,
) domain.Kontaktaufnahme {
	return domain.Kontaktaufnahme{
		Nonce: nonce,
		Kontakt: domain.Peerkontakt{
			Kennung: "banna", Fingerabdruck: "abcd-ef01",
			SelbstgewaehlterName: "Anna", EigenerName: "Anna vom Stammtisch",
			Adresse: "192.168.1.42:41234", Stand: stand, Rechte: rechte,
			Vorgaenger: "baltanna",
		},
	}
}

func TestAufnahmeLegtDenKontaktAn(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	if fehler := ablage.Aufnehmen(ctx, eineAufnahme(
		"n1", domain.KontaktBestaetigt, domain.ZweckHandel), derPeerlauf); fehler != nil {
		t.Fatalf("Aufnehmen: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "banna")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Stand != domain.KontaktBestaetigt ||
		gelesen.EigenerName != "Anna vom Stammtisch" ||
		gelesen.Adresse != "192.168.1.42:41234" ||
		gelesen.Vorgaenger != "baltanna" ||
		!slices.Equal(gelesen.Rechte, []domain.Zweck{domain.ZweckHandel}) ||
		!gelesen.Erstmals.Equal(derPeerlauf) {
		t.Fatalf("gelesen: %+v", gelesen)
	}

	angenommen, fehler := ablage.Angenommen(ctx, "n1")
	if fehler != nil || !angenommen {
		t.Fatalf("Angenommen = %v, %v", angenommen, fehler)
	}
}

func TestAufnahmeEinZweitesMalScheitertGanz(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	if fehler := ablage.Aufnehmen(ctx,
		eineAufnahme("n1", domain.KontaktUnbestaetigt), derPeerlauf); fehler != nil {
		t.Fatalf("Aufnehmen: %v", fehler)
	}

	// **Dieselbe Nonce, und diesmal mit Vergleich und Freigabe**: nichts
	// davon darf ankommen, sonst wäre die Wiederholung eine Hintertür.
	fehler := ablage.Aufnehmen(ctx, eineAufnahme(
		"n1", domain.KontaktBestaetigt, domain.ZweckHandel), derPeerlauf)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("zweite Annahme: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "banna")
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen.Stand != domain.KontaktUnbestaetigt || len(gelesen.Rechte) != 0 {
		t.Fatalf("die Wiederholung hat gewirkt: %+v", gelesen)
	}
}

func TestAufnahmeAnEinemBekannten(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "banna", "Anna alt")

	if fehler := ablage.Stand(ctx, "banna", domain.KontaktBestaetigt); fehler != nil {
		t.Fatal(fehler)
	}

	// **Ein unbestätigtes Annehmen nimmt keine Bestätigung zurück** — und
	// ein leerer eigener Name keinen vorhandenen.
	aufnahme := eineAufnahme("n2", domain.KontaktUnbestaetigt)
	aufnahme.Kontakt.EigenerName = ""
	aufnahme.Kontakt.Adresse = ""

	if fehler := ablage.Aufnehmen(ctx, aufnahme, derPeerlauf); fehler != nil {
		t.Fatalf("Aufnehmen: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "banna")
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen.Stand != domain.KontaktBestaetigt || gelesen.SelbstgewaehlterName != "Anna" ||
		gelesen.Adresse != "" || gelesen.Vorgaenger != "baltanna" {
		t.Fatalf("gelesen: %+v", gelesen)
	}
}

func TestAufnahmeAnEinemGesperrten(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "banna", "Anna")

	if fehler := ablage.Stand(ctx, "banna", domain.KontaktGesperrt); fehler != nil {
		t.Fatal(fehler)
	}

	fehler := ablage.Aufnehmen(ctx,
		eineAufnahme("n3", domain.KontaktBestaetigt, domain.ZweckHandel), derPeerlauf)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Aufnahme eines Gesperrten: %v", fehler)
	}

	gelesen, fehler := ablage.Lesen(ctx, "banna")
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen.Stand != domain.KontaktGesperrt || len(gelesen.Rechte) != 0 {
		t.Fatalf("die Sperre ist gefallen: %+v", gelesen)
	}

	// **Und die Nonce ist nicht verbraucht** — gescheitert heißt: nichts.
	if angenommen, befund := ablage.Angenommen(ctx, "n3"); befund != nil || angenommen {
		t.Fatal("die Nonce ist verbraucht, ohne dass etwas geschah")
	}
}

func TestAufnahmeMitRechtOhneBestaetigung(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	fehler := ablage.Aufnehmen(ctx,
		eineAufnahme("n4", domain.KontaktUnbestaetigt, domain.ZweckHandel), derPeerlauf)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Recht ohne Vergleich: %v", fehler)
	}

	if _, fehler := ablage.Lesen(ctx, "banna"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("ein halber Kontakt steht da: %v", fehler)
	}
}

func TestAufnahmeGehoertDemProfil(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)

	if fehler := ablage.Aufnehmen(probekontext(t),
		eineAufnahme("n5", domain.KontaktUnbestaetigt), derPeerlauf); fehler != nil {
		t.Fatal(fehler)
	}

	anderes := zweitesProfil(t, storage.NeueNutzerablage(db), "Zweites")
	if angenommen, befund := ablage.Angenommen(imProfil(t, anderes), "n5"); befund != nil || angenommen {
		t.Fatal("die Nonce eines Profils gilt im anderen")
	}
}
