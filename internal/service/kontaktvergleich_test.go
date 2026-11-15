package service_test

// **Bestätigen mit dem Vergleichscode** — NET-01
// (#1705).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestBestaetigenMitDemVergleichscode(t *testing.T) {
	t.Parallel()

	ich, erzeugt := peer.Erzeugen(nil)
	if erzeugt != nil {
		t.Fatal(erzeugt)
	}

	werk := service.NeuesKontaktwerk(neuesKontaktdoppel(), einePeeruhr).
		MitAusweis(festerAusweis{ich: ich})
	kennung, _ := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatal(fehler)
	}

	// **Der Code eines anderen Paars ist ein Irrtum** — derselbe Ausgang
	// wie beim falschen Abdruck.
	anderer, _ := eineFremdeKennung(t)
	falsch := domain.Vergleichscode(string(ich.Kennung()), string(anderer))

	if fehler := werk.Bestaetigen(t.Context(), kennung, falsch); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der falsche Code kam durch: %v", fehler)
	}

	richtig := domain.Vergleichscode(string(kennung), string(ich.Kennung()))
	if fehler := werk.Bestaetigen(t.Context(), kennung, richtig); fehler != nil {
		t.Fatalf("Bestaetigen: %v", fehler)
	}

	kontakt, fehler := werk.Lesen(t.Context(), kennung)
	if fehler != nil || kontakt.Stand != domain.KontaktBestaetigt {
		t.Fatalf("Stand %q", kontakt.Stand)
	}
}

func TestBestaetigenOhneAusweisKenntKeinenCode(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, _ := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatal(fehler)
	}

	// **Ohne Ausweis gibt es kein Paar**, also nur den Abdruck.
	if fehler := werk.Bestaetigen(t.Context(), kennung,
		domain.Vergleichscode(string(kennung), "x")); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Bestaetigen: %v", fehler)
	}
}

func TestBestaetigenMitKaputtemAusweis(t *testing.T) {
	t.Parallel()

	werk := service.NeuesKontaktwerk(neuesKontaktdoppel(), einePeeruhr).
		MitAusweis(kaputterAusweis{})
	kennung, abdruck := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatal(fehler)
	}

	// **Der Abdruck braucht keinen Ausweis** und geht weiter; ein Code
	// braucht ihn, und sein Fehler geht durch.
	if fehler := werk.Bestaetigen(t.Context(), kennung, "11111 22222"); fehler == nil ||
		errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Bestaetigen mit Code: %v", fehler)
	}

	if fehler := werk.Bestaetigen(t.Context(), kennung, abdruck); fehler != nil {
		t.Fatalf("Bestaetigen mit Abdruck: %v", fehler)
	}
}
