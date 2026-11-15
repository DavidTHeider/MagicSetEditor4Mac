package storage_test

// **Eine ausgestellte Einladung wird einmal eingelöst** — NET-01b
// (#1954).
//
// # Was hier gemessen wird
//
// **Dass eine fremde, verbrauchte oder abgelaufene Nonce nichts anlegt**
// (AK-02) — und dass eine Einlösung Kontakt und Vermerk zusammen schreibt
// oder keines von beiden.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

var einloesezeit = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

func einEinloesender(kennung string) domain.Peerkontakt {
	return domain.Peerkontakt{
		Kennung: kennung, Fingerabdruck: "abcd-" + kennung,
		SelbstgewaehlterName: "Bert", Adresse: "192.0.2.8:41234",
		Stand: domain.KontaktUnbestaetigt,
	}
}

func eineAusgestellte(t *testing.T) (ablage *storage.Kontaktablage, ausstellen func(nonce string)) {
	t.Helper()

	ablage = storage.NeueKontaktablage(frischeDatenbank(t))

	return ablage, func(nonce string) {
		t.Helper()

		if fehler := ablage.Ausgestellt(probekontext(t), nonce, domain.ZweckHandel,
			einloesezeit.Add(domain.KontakteinladungGueltig), einloesezeit); fehler != nil {
			t.Fatalf("Ausgestellt: %v", fehler)
		}
	}
}

func TestEineEinloesungLegtDenKontaktUnbestaetigtAn(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)
	ctx := probekontext(t)

	ausstellen("n1")

	if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), einloesezeit); fehler != nil {
		t.Fatalf("Einloesen: %v", fehler)
	}

	kontakt, fehler := ablage.Lesen(ctx, "bbert")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if kontakt.Stand != domain.KontaktUnbestaetigt || kontakt.SelbstgewaehlterName != "Bert" ||
		kontakt.Adresse != "192.0.2.8:41234" || len(kontakt.Rechte) != 0 ||
		!kontakt.Erstmals.Equal(einloesezeit) {
		t.Errorf("Kontakt: %+v", kontakt)
	}

	// **Ein zweites Mal ist verbraucht** — auch für jemand anderen.
	if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bcarla"), einloesezeit); !errors.Is(
		fehler, domain.ErrEinladungVerbraucht) {
		t.Errorf("zweimal: %v", fehler)
	}

	if _, fehler := ablage.Lesen(ctx, "bcarla"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der zweite steht da: %v", fehler)
	}
}

func TestFremdeUndAbgelaufeneNoncesLegenNichtsAn(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)
	ctx := probekontext(t)

	ausstellen("n1")

	if fehler := ablage.Einloesen(ctx, "fremd", einEinloesender("bbert"), einloesezeit); !errors.Is(
		fehler, domain.ErrEinladungFremd) {
		t.Errorf("fremd: %v", fehler)
	}

	spaeter := einloesezeit.Add(domain.KontakteinladungGueltig)
	if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), spaeter); !errors.Is(
		fehler, domain.ErrEinladungAbgelaufen) {
		t.Errorf("abgelaufen: %v", fehler)
	}

	if liste, fehler := ablage.Liste(ctx); fehler != nil || len(liste) != 0 {
		t.Errorf("angelegt: %+v, %v", liste, fehler)
	}
}

// **Eine Einladung gehört ihrem Profil** — ein anderes auf demselben
// Rechner kennt sie nicht.
func TestEineEinladungGehoertIhremProfil(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)

	ausstellen("n1")

	if fehler := ablage.Einloesen(storage.MitProfil(t.Context(), 2), "n1",
		einEinloesender("bbert"), einloesezeit); !errors.Is(fehler, domain.ErrEinladungFremd) {
		t.Errorf("ein anderes Profil: %v", fehler)
	}
}

// **Wer gesperrt ist, löst nichts ein** — und die Nonce bleibt, damit
// nach dem Entsperren nicht eine verbrauchte Einladung übrig ist.
func TestEinGesperrterLoestNichtsEin(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)
	ctx := probekontext(t)

	ausstellen("n1")

	if fehler := ablage.Begegnet(ctx, "bbert", "abcd", "Bert", einloesezeit); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Stand(ctx, "bbert", domain.KontaktGesperrt); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), einloesezeit); !errors.Is(
		fehler, domain.ErrEinloesungGesperrt) {
		t.Errorf("gesperrt: %v", fehler)
	}

	if fehler := ablage.Stand(ctx, "bbert", domain.KontaktUnbestaetigt); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), einloesezeit); fehler != nil {
		t.Errorf("nach dem Entsperren: %v", fehler)
	}
}

// **Ein bestätigter Kontakt bleibt bestätigt** — die Einlösung setzt keinen
// Vergleich zurück, und sie gibt kein Recht.
func TestEinBestaetigterBleibtBestaetigt(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)
	ctx := probekontext(t)

	ausstellen("n1")

	if fehler := ablage.Aufnehmen(ctx, domain.Kontaktaufnahme{
		Nonce: "angenommen", Kontakt: domain.Peerkontakt{
			Kennung: "bbert", Fingerabdruck: "abcd", SelbstgewaehlterName: "Bert",
			Stand: domain.KontaktBestaetigt, Rechte: []domain.Zweck{domain.ZweckPartie},
		},
	}, einloesezeit); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), einloesezeit); fehler != nil {
		t.Fatalf("Einloesen: %v", fehler)
	}

	kontakt, fehler := ablage.Lesen(ctx, "bbert")
	if fehler != nil || kontakt.Stand != domain.KontaktBestaetigt ||
		len(kontakt.Rechte) != 1 || kontakt.Adresse != "192.0.2.8:41234" {
		t.Errorf("Kontakt: %+v, %v", kontakt, fehler)
	}
}

// **Was lange abgelaufen ist, verschwindet** — beim nächsten Ausstellen,
// nach einer Woche Gnade, in der eine Wiederholung noch „verbraucht" hört.
func TestAlteAusgestellteVerschwinden(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)
	ctx := probekontext(t)

	ausstellen("alt")

	viel := einloesezeit.Add(domain.KontakteinladungGueltig + 8*24*time.Hour)
	if fehler := ablage.Ausgestellt(ctx, "neu", domain.ZweckHandel,
		viel.Add(domain.KontakteinladungGueltig), viel); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Einloesen(ctx, "alt", einEinloesender("bbert"), viel); !errors.Is(
		fehler, domain.ErrEinladungFremd) {
		t.Errorf("alt: %v", fehler)
	}
}

// **Eine Nonce wird einmal ausgestellt** — ein zweites Mal wäre ein
// Zufall, der keiner ist.
func TestEineNonceWirdEinmalAusgestellt(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)

	ausstellen("n1")

	if fehler := ablage.Ausgestellt(probekontext(t), "n1", domain.ZweckHandel,
		einloesezeit.Add(time.Hour), einloesezeit); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("zweimal: %v", fehler)
	}
}
