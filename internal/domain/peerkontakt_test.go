package domain_test

// **Was ein Kontakt darf** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Wogegen hier geprüft wird
//
// **Gegen zwei Sätze aus dem Issue, wörtlich genommen:**
//
//   - „Pairing allein gestattet keinen Import oder Handel."
//   - „Gegenüber per Fingerprint/Bestätigung prüfen; Anzeigenamen allein sind
//     kein Identitätsnachweis."
//
// Der erste ist eine Aussage über Rechte, der zweite eine über den Stand —
// und beide zusammen ergeben die Tabelle unten.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Tabelle, um die es geht.**
func TestWasEinKontaktDarf(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		kontakt domain.Peerkontakt
		zweck   domain.Zweck
		darf    bool
	}{
		{
			"ein ungeprüftes Gegenüber darf sich vorstellen",
			domain.Peerkontakt{Stand: domain.KontaktUnbestaetigt},
			domain.ZweckPaarung, true,
		},
		{
			"und sonst nichts",
			domain.Peerkontakt{Stand: domain.KontaktUnbestaetigt},
			domain.ZweckTeilen, false,
		},
		{
			"auch nicht handeln",
			domain.Peerkontakt{Stand: domain.KontaktUnbestaetigt},
			domain.ZweckHandel, false,
		},
		{
			"ein bestätigtes ohne Freigabe darf auch nichts",
			domain.Peerkontakt{Stand: domain.KontaktBestaetigt},
			domain.ZweckTeilen, false,
		},
		{
			"mit Freigabe darf es den freigegebenen Zweck",
			domain.Peerkontakt{
				Stand:  domain.KontaktBestaetigt,
				Rechte: []domain.Zweck{domain.ZweckTeilen},
			},
			domain.ZweckTeilen, true,
		},
		{
			"und nur den",
			domain.Peerkontakt{
				Stand:  domain.KontaktBestaetigt,
				Rechte: []domain.Zweck{domain.ZweckTeilen},
			},
			domain.ZweckHandel, false,
		},
		{
			"ein gesperrtes darf nichts",
			domain.Peerkontakt{
				Stand:  domain.KontaktGesperrt,
				Rechte: []domain.Zweck{domain.ZweckTeilen},
			},
			domain.ZweckTeilen, false,
		},
		{
			"auch sich nicht vorstellen",
			domain.Peerkontakt{Stand: domain.KontaktGesperrt},
			domain.ZweckPaarung, false,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if darf := fall.kontakt.Darf(fall.zweck); darf != fall.darf {
				t.Fatalf("darf ist %v, erwartet war %v", darf, fall.darf)
			}
		})
	}
}

// **Ein Kontakt, den es so nicht geben kann, wird abgewiesen.**
func TestEinUnmoeglicherKontaktWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Peerkontakt{
		"ohne Kennung": {Stand: domain.KontaktBestaetigt},
		"fremder Stand": {
			Kennung: "bwer", Stand: "vielleicht",
		},
		"fremder Zweck": {
			Kennung: "bwer",
			Stand:   domain.KontaktBestaetigt,
			Rechte:  []domain.Zweck{"alles"},
		},

		// **Der Satz, für den es diese Prüfung gibt.** Ein Recht an einem
		// ungeprüften Gegenüber hinge an einem Namen, den sich jeder geben
		// kann.
		"Recht ohne Bestätigung": {
			Kennung: "bwer",
			Stand:   domain.KontaktUnbestaetigt,
			Rechte:  []domain.Zweck{domain.ZweckTeilen},
		},
		"Recht trotz Sperre": {
			Kennung: "bwer",
			Stand:   domain.KontaktGesperrt,
			Rechte:  []domain.Zweck{domain.ZweckTeilen},
		},
	}

	for name, kontakt := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := kontakt.Validate(); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("er kam durch: %v", fehler)
			}
		})
	}
}

// **Und ein möglicher kommt durch.**
func TestEinMoeglicherKontaktKommtDurch(t *testing.T) {
	t.Parallel()

	kontakt := domain.Peerkontakt{
		Kennung:              "bwer",
		Fingerabdruck:        "abcd-ef01",
		SelbstgewaehlterName: "Fremder Rechner",
		EigenerName:          "Anna",
		Stand:                domain.KontaktBestaetigt,
		Rechte:               []domain.Zweck{domain.ZweckTeilen, domain.ZweckPartie},
	}

	if fehler := kontakt.Validate(); fehler != nil {
		t.Fatalf("er kam nicht durch: %v", fehler)
	}
}
