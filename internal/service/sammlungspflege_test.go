// **Die Sammlung pflegen** — Werkstück 2 von Phase 3.4
// (#1165).
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sammlungsdoppel ist eine Ablage im Arbeitsspeicher.
type sammlungsdoppel struct {
	zeilen   []*domain.Exemplar
	naechste int64

	// fehler ist, was jeder Weg zurückgibt — für die Fehlerläufe.
	fehler error
}

// Anlegen legt eine Zeile an — **mit Fassung 1**, wie der Bestand.
//
// Die Migration sagt `fassung INTEGER NOT NULL DEFAULT 1`. Ein Doppel, das
// eine Null hinterließe, ließe jede Zusage über Fassungen ins Leere laufen
// (§3, „eine Vorrichtung, die großzügiger ist als der Betrieb").
func (s *sammlungsdoppel) Anlegen(_ context.Context, exemplar *domain.Exemplar) (int64, error) {
	if s.fehler != nil {
		return 0, s.fehler
	}

	s.naechste++
	kopie := *exemplar
	kopie.ID = s.naechste
	kopie.Fassung = 1
	s.zeilen = append(s.zeilen, &kopie)

	return kopie.ID, nil
}

// Aendern schreibt eine Zeile zurück — **gegen ihre Fassung**, und erhöht sie.
//
// # Warum das Doppel es tut
//
// Weil der Bestand es tut: `UPDATE … WHERE id = ? AND fassung = ?` samt
// `fassung = fassung + 1`. Ein Doppel ohne diese Prüfung ist großzügiger als
// der Betrieb — und damit hätte §7 Regel 2 hier keinen Lauf, der sie prüft.
//
// Angestrichen am 2026-08-30 bei den Massenvorgängen
// (#1292):
// ein Lauf erwartete eine übergangene Zeile mit veralteter Fassung und bekam
// drei geänderte.
func (s *sammlungsdoppel) Aendern(_ context.Context, exemplar *domain.Exemplar) error {
	if s.fehler != nil {
		return s.fehler
	}

	for i, zeile := range s.zeilen {
		if zeile.ID != exemplar.ID {
			continue
		}

		if zeile.Fassung != exemplar.Fassung {
			return domain.ErrConflict
		}

		kopie := *exemplar
		kopie.Fassung = zeile.Fassung + 1
		s.zeilen[i] = &kopie

		return nil
	}

	return domain.ErrNotFound
}

// Loeschen nimmt eine Zeile heraus — **gegen ihre Fassung**.
//
// # Warum das Doppel sie prüft
//
// Weil der Bestand es tut: `DELETE FROM sammlung WHERE id = ? AND fassung = ?`.
// Ein Doppel, das die Fassung übergeht, ist **großzügiger als der Betrieb**
// (§3) — und damit hätte die Zusage aus §7 Regel 2 hier keinen Lauf, der sie
// prüft. Angestrichen am 2026-08-30 beim Löschen in Masse
// (#1292):
// der Test erwartete eine übergangene Zeile und bekam drei gelöschte.
func (s *sammlungsdoppel) Loeschen(_ context.Context, id, fassung int64) error {
	if s.fehler != nil {
		return s.fehler
	}

	for i, zeile := range s.zeilen {
		if zeile.ID != id {
			continue
		}

		if zeile.Fassung != fassung {
			return domain.ErrConflict
		}

		s.zeilen = append(s.zeilen[:i], s.zeilen[i+1:]...)

		return nil
	}

	return domain.ErrNotFound
}

func (s *sammlungsdoppel) Liste(context.Context) ([]*domain.Exemplar, error) {
	return s.zeilen, s.fehler
}

func (s *sammlungsdoppel) Fuer(_ context.Context, printing string) (
	[]*domain.Exemplar, error,
) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	gefunden := []*domain.Exemplar{}

	for _, zeile := range s.zeilen {
		if zeile.Printing == printing {
			gefunden = append(gefunden, zeile)
		}
	}

	return gefunden, nil
}

// NachDruck gibt die Zeilen einer Druckzeile — wie der Bestand (A2, #1441).
func (s *sammlungsdoppel) NachDruck(_ context.Context, druckID int64) (
	[]*domain.Exemplar, error,
) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	gefunden := []*domain.Exemplar{}

	for _, zeile := range s.zeilen {
		if zeile.DruckID == druckID {
			gefunden = append(gefunden, zeile)
		}
	}

	return gefunden, nil
}

// **Was durchkommt, ist geglättet und vollständig.**
//
// # Warum das hier steht und nicht im Speicher
//
// Der Speicher schreibt, was er bekommt. Ob ein `zustand: "prima"` in den
// Bestand darf, ist eine **fachliche** Frage — und die gehört in die Schicht,
// die sie einmal beantwortet, statt in jede der vier Flächen.
func TestEinAngelegtesExemplarIstGeglaettet(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsdoppel{}
	pflege := service.NeueSammlungspflege(doppel)

	if _, err := pflege.Anlegen(t.Context(), domain.Exemplar{
		Printing: "  abc\n",
		Menge:    1,
	}); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	abgelegt := doppel.zeilen[0]

	if abgelegt.Printing != "abc" {
		t.Errorf("der Druck ist %q", abgelegt.Printing)
	}

	// **Die Vorgaben sind gesetzt, bevor geschrieben wird.** Sonst stünde im
	// Bestand ein leerer Zustand, und jede Auswertung müsste ihn deuten.
	if abgelegt.Zustand != domain.ZustandNearMint || abgelegt.Sprache != "en" {
		t.Errorf("die Vorgaben fehlen: %+v", *abgelegt)
	}
}

// **Was nicht stimmt, kommt gar nicht erst in den Bestand.**
func TestEinUngueltigesExemplarErreichtDenBestandNicht(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsdoppel{}

	_, err := service.NeueSammlungspflege(doppel).
		Anlegen(t.Context(), domain.Exemplar{Printing: "", Menge: 0})

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v — will ErrValidation", err)
	}

	if len(doppel.zeilen) != 0 {
		t.Errorf("es wurde etwas abgelegt: %+v", doppel.zeilen)
	}
}

// **Der Befund hängt an einem Feld — und es ist das erste falsche.**
//
// # Was hier ausdrücklich **nicht** zugesagt wird
//
// Dass alle drei falschen Felder auf einmal ankommen. `app.Fehler` trägt ein
// Feld und einen Satz; wer drei falsch ausgefüllt hat, korrigiert sie
// nacheinander. Der Grund und sein Preis stehen bei `vorbereiten` — hier steht
// die Zusage, die wirklich gilt: **es kommt ein Feldbefund an, und er nennt
// sein Feld.**
func TestEinUngueltigesExemplarNenntSeinFeld(t *testing.T) {
	t.Parallel()

	_, err := service.NeueSammlungspflege(&sammlungsdoppel{}).
		Anlegen(t.Context(), domain.Exemplar{Menge: 0, Zustand: "prima"})

	var befund *domain.FieldError
	if !errors.As(err, &befund) {
		t.Fatalf("der Fehler ist kein Feldbefund: %v", err)
	}

	if befund.Field != "printing" {
		t.Errorf("der Befund hängt an %q", befund.Field)
	}
}

// **Eine Änderung ohne Kennung ist keine.**
//
// Ohne sie schriebe der Speicher gegen `id = 0` — und träfe nichts, was aber
// nach „gibt es nicht" aussähe statt nach „du hast nicht gesagt, was".
func TestEineAenderungOhneKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	err := service.NeueSammlungspflege(&sammlungsdoppel{}).
		Aendern(t.Context(), domain.Exemplar{Printing: "abc", Menge: 1})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **Eine Änderung wird ebenso geglättet und geprüft.**
func TestEineAenderungWirdGeprueft(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsdoppel{}
	pflege := service.NeueSammlungspflege(doppel)

	id, err := pflege.Anlegen(t.Context(), domain.Exemplar{Printing: "abc", Menge: 1})
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	fehler := pflege.Aendern(t.Context(), domain.Exemplar{
		ID: id, Printing: "abc", Menge: 0,
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", fehler)
	}
}

// **Löschen, Lesen und Nachschlagen reichen durch.**
func TestDieSammlungspflegeReichtDurch(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsdoppel{}
	pflege := service.NeueSammlungspflege(doppel)

	id, err := pflege.Anlegen(t.Context(), domain.Exemplar{Printing: "abc", Menge: 2})
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	fuer, fuerfehler := pflege.Fuer(t.Context(), "abc")
	if fuerfehler != nil || len(fuer) != 1 {
		t.Fatalf("Fuer: %v, %d Zeilen", fuerfehler, len(fuer))
	}

	if fehler := pflege.Loeschen(t.Context(), id, 1); fehler != nil {
		t.Fatalf("Loeschen: %v", fehler)
	}

	liste, listenfehler := pflege.Liste(t.Context())
	if listenfehler != nil || len(liste) != 0 {
		t.Errorf("Liste: %v, %d Zeilen", listenfehler, len(liste))
	}
}

// **Ein Fehler der Ablage kommt durch — auf jedem Weg.**
func TestEinFehlerDerAblageKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Ablage stolpert")
	pflege := service.NeueSammlungspflege(&sammlungsdoppel{fehler: gestolpert})

	for name, versuch := range map[string]func() error{
		"Anlegen": func() error {
			_, err := pflege.Anlegen(t.Context(),
				domain.Exemplar{Printing: "abc", Menge: 1})

			return err
		},
		"Aendern": func() error {
			return pflege.Aendern(t.Context(),
				domain.Exemplar{ID: 1, Printing: "abc", Menge: 1})
		},
		"Loeschen": func() error { return pflege.Loeschen(t.Context(), 1, 1) },
		"Liste": func() error {
			_, err := pflege.Liste(t.Context())

			return err
		},
		"Fuer": func() error {
			_, err := pflege.Fuer(t.Context(), "abc")

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !errors.Is(versuch(), gestolpert) {
				t.Error("der Fehler der Ablage kam nicht durch")
			}
		})
	}
}
