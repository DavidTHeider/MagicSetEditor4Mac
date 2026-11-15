// **Ein `leser` kommt hinein und schreibt nicht** — Werkstück 16 von Phase
// 4.12 (#1351),
// nach [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// # Die Abnahme dieses Werkstücks
//
// **Zwei Sätze, und der zweite ist neu**: `leser` darf in ein fremdes Profil,
// und wer als `leser` darin steht, schreibt nicht. Bis hierher galt der erste
// nicht — `freigabeReicht` verlangte `RolleMitarbeiter`.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitRolle baut zwei Nutzer und gibt dem Ersten die genannte Rolle auf dem
// Profil des Zweiten.
func mitRolle(t *testing.T, rolle domain.Rolle) (
	*service.Nutzerverwaltung, *nutzerdoppel,
) {
	t.Helper()

	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Erster"}, {ID: 2, Name: "Zweiter"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Meins"}},
			2: {{ID: 2, NutzerID: 2, Name: "Fremd"}},
		},
		zeiger: 1,
	}

	doppel := neuesFreigabendoppel()
	if rolle != "" {
		doppel.abgelegt = []domain.Profilfreigabe{
			{ProfilID: 2, NutzerID: 1, Rolle: rolle},
		}
	}

	// **Mit verschlossenem Faktor**, sonst käme jeder überall hinein und der
	// Lauf prüfte die Abwesenheit von Schutz (Werkstück 11).
	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFaktor(verschlossen()).
		MitFreigaben(service.NeueFreigabenverwaltung(doppel, ablage))

	return verwaltung, ablage
}

// **Ein `leser` kommt hinein.**
//
// **Bis Werkstück 16 kam er nicht** — `freigabeReicht` verlangte
// `RolleMitarbeiter`, und der Grund stand in
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md):
// ein Nur-Lese-Modus **innerhalb** eines Profils braucht einen lückenlosen
// Schreibschutz. Den gibt es jetzt, im Treiber.
func TestEinLeserKommtInEinFremdesProfil(t *testing.T) {
	verwaltung, ablage := mitRolle(t, domain.RolleLeser)

	if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler != nil {
		t.Fatalf("der Leser kam nicht hinein: %v", fehler)
	}

	// **Am Zeiger und nicht am Rückgabewert** (§10).
	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d, erwartet war 2", ablage.zeiger)
	}
}

// **Und ohne jede Rolle kommt weiterhin niemand hinein.**
//
// **Ohne diesen Lauf wäre der darüber auch dann grün, wenn die Prüfung ganz
// wegfiele** (§10: eine Gegenprobe, die nichts ändert, ist keine).
func TestOhneRolleKommtNiemandHinein(t *testing.T) {
	verwaltung, ablage := mitRolle(t, "")

	if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler == nil {
		t.Fatal("ohne Rolle ging das fremde Profil auf")
	}

	if ablage.zeiger != 1 {
		t.Errorf("der Zeiger steht auf %d — der Wechsel fand statt", ablage.zeiger)
	}
}

// **Und die Rolle im gezeigten Profil ist ablesbar** — daran hängt, ob der
// Bestand nur lesend geöffnet wird.
//
// **Erst seit Werkstück 17 ist sie das** (#1400): vorher galt, wer
// hineingewechselt war, als der Eigner — und ein Eigner ist `RolleVerwalter`.
func TestDieRolleImGezeigtenProfilIstAblesbar(t *testing.T) {
	for name, fall := range map[string]struct {
		rolle    domain.Rolle
		erwartet domain.Rolle
	}{
		"als Leser":       {domain.RolleLeser, domain.RolleLeser},
		"als Mitarbeiter": {domain.RolleMitarbeiter, domain.RolleMitarbeiter},
	} {
		t.Run(name, func(t *testing.T) {
			verwaltung, _ := mitRolle(t, fall.rolle)

			if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler != nil {
				t.Fatalf("hinein: %v", fehler)
			}

			rolle, fehler := verwaltung.RolleImGezeigten(t.Context())
			if fehler != nil {
				t.Fatalf("Rolle: %v", fehler)
			}

			if rolle != fall.erwartet {
				t.Errorf("Rolle %q, erwartet war %q", rolle, fall.erwartet)
			}
		})
	}

	// **Im eigenen Profil ist man Verwalter** — ohne Zeile im Bestand.
	t.Run("im eigenen Profil", func(t *testing.T) {
		verwaltung, _ := mitRolle(t, "")

		rolle, fehler := verwaltung.RolleImGezeigten(t.Context())
		if fehler != nil {
			t.Fatalf("Rolle: %v", fehler)
		}

		if rolle != domain.RolleVerwalter {
			t.Errorf("Rolle %q, erwartet war Verwalter", rolle)
		}
	})

	// **Und ein Lesefehler ist keine Rolle** (§8).
	t.Run("der Bestand klemmt", func(t *testing.T) {
		verwaltung, ablage := mitRolle(t, "")
		ablage.nutzerzeigerfehler = errors.New("die Platte klemmt")

		if _, fehler := verwaltung.RolleImGezeigten(t.Context()); fehler == nil {
			t.Error("aus einem Lesefehler wurde eine Rolle")
		}
	})
}
