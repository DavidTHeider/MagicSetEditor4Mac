// Was `ValidateBasics` **ohne** Spielpaket entscheiden kann.
//
// # Warum das mehr wurde
//
// #89 hielt
// fest: die Prüfung deckte zwei Klassen ab und ließ doppelte Setwerte und leere
// Feldnamen durch — und sie fragte nicht „mehr prüfen", sondern **wo**. Für
// doppelte Setwerte fiel das Urteil im Primärschlüssel von `set_values`; es
// stimmte, aber es entstand in `storage`, das die Regel nicht kennt.
//
// **Der Fall ist eingetreten:** ein `.mse-set` mit demselben Schlüssel zweimal
// auf Setebene ist eine gewöhnliche fremde Datei (§8). Bis hierher scheiterte
// ihr Import am Primärschlüssel — mit einem Konflikt statt einer Meldung, die
// sagt, **welches Feld** doppelt ist.
package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probeSet gibt ein Set, das die Prüfung besteht.
func probeSet() *domain.Set {
	return &domain.Set{
		Name: "Probeset", Game: "probe",
		Values: []domain.Value{{Field: "title", Text: "Probeset"}},
		Cards:  []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}},
	}
}

func TestValidateBasicsWeistDieSetwerteAb(t *testing.T) {
	faelle := []struct {
		name    string
		bauen   func(*domain.Set)
		meldung string
		art     error
	}{
		{
			// **Ein Konflikt, keine Prüfmeldung** — dieselbe Einordnung, die
			// ein doppeltes Kartenfeld seit Phase 2 bekommt. Sie hier
			// umzustellen hieße, an einem Werkstück über den **Ort** einer
			// Regel nebenbei ihre **Art** zu ändern; das gehört an einen Fall,
			// der es verlangt, und der ist nicht dieser (§10).
			name: "derselbe Setwert zweimal",
			bauen: func(s *domain.Set) {
				s.Values = append(s.Values, domain.Value{Field: "title", Text: "Noch eins"})
			},
			meldung: "title",
			art:     domain.ErrConflict,
		},
		{
			name: "ein Setwert ohne Feldnamen",
			bauen: func(s *domain.Set) {
				s.Values = append(s.Values, domain.Value{Field: "  ", Text: "x"})
			},
			meldung: "Namen",
			art:     domain.ErrValidation,
		},
		{
			name: "ein Kartenwert ohne Feldnamen",
			bauen: func(s *domain.Set) {
				s.Cards[0].Values = append(s.Cards[0].Values, domain.Value{Field: "", Text: "x"})
			},
			meldung: "Namen",
			art:     domain.ErrValidation,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			set := probeSet()
			fall.bauen(set)

			err := set.ValidateBasics()
			if err == nil {
				t.Fatal("kein Befund")
			}

			if !errors.Is(err, fall.art) {
				t.Errorf("%v ist nicht %v", err, fall.art)
			}

			if !strings.Contains(err.Error(), fall.meldung) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.meldung)
			}
		})
	}
}

// **Ein Set, das in Ordnung ist, bleibt in Ordnung.**
//
// Die Gegenprobe zu jeder neuen Prüfung: ein Wächter, der auch das Richtige
// abweist, ist keiner.
func TestValidateBasicsLaesstEinGutesSetDurch(t *testing.T) {
	if err := probeSet().ValidateBasics(); err != nil {
		t.Errorf("%v", err)
	}
}

// **Ein leerer Setname bleibt erlaubt** — und das ist eine Entscheidung.
//
// #89 führt ihn unter „durchgelassen". Er bleibt es: **keine Regel verbietet
// ihn.** `sets.name` ist `NOT NULL`, aber eine leere Zeichenkette ist nicht
// NULL, und kein gesichtetes Paket sagt, dass ein Set einen Titel tragen muss.
// Ihn jetzt abzuweisen hieße eine Regel zu erfinden — und die erste fremde
// Datei ohne Titel wäre unlesbar, ohne dass jemand sagen könnte, warum.
//
// Auch eine Ablehnung braucht Deckung (§10); hier ist sie.
func TestEinLeererSetnameBleibtErlaubt(t *testing.T) {
	set := probeSet()
	set.Name = ""

	if err := set.ValidateBasics(); err != nil {
		t.Errorf("%v — der leere Setname ist keine erfundene Regel wert", err)
	}
}

// **`PruefeFeldname` ist die Regel, wie ein Schreiber sie ruft.**
//
// Sie steht neben `feldnameBefund`, weil ein Aufrufer einen `error` umschließt
// und `doppelteWerte` einen `*FieldError` weiterreicht. Zwei Signaturen für
// eine Regel — und beide brauchen ihren Fall, sonst hält ein Wächter die halbe
// Zusage (§10).
func TestPruefeFeldnameIstDieRegelFuerSchreiber(t *testing.T) {
	for name, feld := range map[string]string{
		"leer":        "",
		"Leerzeichen": "   ",
		"Tabulator":   "\t",
	} {
		t.Run(name, func(t *testing.T) {
			err := domain.PruefeFeldname(feld)
			if err == nil {
				t.Fatal("kein Befund")
			}

			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("%v ist keine Prüfmeldung", err)
			}
		})
	}

	// Und die Gegenprobe: ein Name mit Leerraum außen ist ein Name.
	if err := domain.PruefeFeldname(" casting cost "); err != nil {
		t.Errorf("%v — ein Name mit Rand ist einer", err)
	}
}
