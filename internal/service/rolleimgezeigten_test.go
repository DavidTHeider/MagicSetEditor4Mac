// **Welche Rolle der arbeitende Mensch im gezeigten Profil hat** — Werkstück
// 16 von Phase 4.12
// (#1351),
// nach [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// **Eigene Datei**, weil `wechselerlaubnis_test.go` mit diesen Läufen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **wer
// hinein darf**, hier, **was er drinnen darf**.

package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Freigabennaht ist jeder Verwalter.**
//
// # Warum das die richtige Antwort ist und nicht die bequeme
//
// **Weil ohne Freigaben niemand in ein fremdes Profil kommt.** `freigabeReicht`
// weist dann jeden Wechsel ab — wer also überhaupt irgendwo steht, steht im
// eigenen, und dort ist er Verwalter.
//
// **Was das kostet, gehört gesagt**: die Antwort ist erschlossen und nicht
// gelesen. Sie stimmt, solange die Zeile in `freigabeReicht` steht, die ohne
// Naht abweist — und die steht direkt daneben.
func TestOhneFreigabennahtIstJederVerwalter(t *testing.T) {
	t.Parallel()

	ablage := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{1: {{ID: 1, NutzerID: 1, Name: "Meins"}}},
		zeiger:  1,
	}

	rolle, fehler := service.NeueNutzerverwaltung(ablage).RolleImGezeigten(t.Context())
	if fehler != nil {
		t.Fatalf("Rolle im gezeigten: %v", fehler)
	}

	if rolle != domain.RolleVerwalter {
		t.Errorf("Rolle = %q, erwartet war %q", rolle, domain.RolleVerwalter)
	}
}

// **Und ein stolpernder Zeiger reißt die Frage ab.**
//
// **Der Nutzerzeiger steht, der Profilzeiger klemmt** — sonst scheiterte schon
// `WerArbeitet`, und der zweite Zweig bliebe unbetreten (§10: ein Lauf, der
// scheitert, bevor er die Stelle erreicht, die er meint).
func TestEinStolpernderProfilzeigerReisstDieRollenfrageAb(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("der Zeiger klemmt")

	ablage := &nutzerdoppel{
		nutzer:       []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile:      map[int64][]domain.Profil{1: {{ID: 1, NutzerID: 1, Name: "Meins"}}},
		zeiger:       1,
		nutzerzeiger: 1,
		zeigerfehler: klemmt,
	}

	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFreigaben(service.NeueFreigabenverwaltung(neuesFreigabendoppel(), ablage))

	if _, fehler := verwaltung.RolleImGezeigten(t.Context()); !errors.Is(fehler, klemmt) {
		t.Errorf("Fehler = %v, erwartet war %v", fehler, klemmt)
	}
}
