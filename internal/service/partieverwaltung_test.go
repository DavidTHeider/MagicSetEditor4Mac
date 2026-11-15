package service_test

// **Löschen und Kopieren einer Partie** — PLAY-03.AK-05
// (#1711):
// „Löschen betrifft eigene Entwürfe nach definierter Regel, Kopien werden
// neue Übungszustände ohne alte Peer-Zustimmungen."

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineAbgelegtePartie stellt eine Partie in den Speicher und gibt beide.
func eineAbgelegtePartie(
	t *testing.T, stand domain.Partiestand, freigaben int,
) (*service.Partieleitung, *partieprobe) {
	t.Helper()

	leitung, speicher := eineLeitung(t)

	partie := einePartie()
	partie.Stand = stand
	speicher.partien[partie.ID] = partie

	speicher.freigaben = map[string][]domain.Offenlegung{}

	for nummer := range freigaben {
		speicher.freigaben[partie.ID] = append(
			speicher.freigaben[partie.ID],
			domain.Offenlegung{Von: "sitz-1", An: []string{"wer"}},
		)

		_ = nummer
	}

	return leitung, speicher
}

// **Die Löschregel, alle drei Fälle** — und der Satz nennt jedes Mal, woran
// es liegt (§6): „darf nicht" allein ließe jemanden raten, ob der Stand oder
// eine Einwilligung im Weg steht, und beides führt zu einem anderen nächsten
// Schritt.
func TestPartieLoeschenFolgtDerRegel(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		stand     domain.Partiestand
		freigaben int
		teil      string
	}{
		{name: "ein laufender Entwurf geht", stand: domain.StandLaeuft},
		{
			name:  "eine beendete bleibt nachvollziehbar",
			stand: domain.StandBeendet, teil: "nachvollziehbar",
		},
		{
			name:  "eine mit Einwilligung bleibt",
			stand: domain.StandLaeuft, freigaben: 1, teil: "Einwilligung",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			leitung, speicher := eineAbgelegtePartie(
				t, fall.stand, fall.freigaben)

			fehler := leitung.Loeschen(context.Background(), "p1")

			if fall.teil == "" {
				if fehler != nil {
					t.Fatalf("Löschen: %v", fehler)
				}

				if len(speicher.geloescht) != 1 {
					t.Fatalf("gelöscht wurden %v", speicher.geloescht)
				}

				return
			}

			if !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.teil) {
				t.Errorf("die Meldung %q nennt %q nicht", fehler, fall.teil)
			}

			if len(speicher.geloescht) != 0 {
				t.Errorf("gelöscht wurde trotzdem: %v", speicher.geloescht)
			}
		})
	}
}

// **Eine Partie, die es nicht gibt, ist ein Befund** — und der Weg dorthin
// führt über den Lesefehler, nicht über die Regel.
func TestPartieLoeschenMeldetDieUnbekannte(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)

	fehler := leitung.Loeschen(context.Background(), "gibtsnicht")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Der zweite Lesefehler ist ein eigener Weg** — wer nur den ersten prüft,
// hat den halben Erfolgsweg gemessen (§10).
func TestPartieLoeschenMeldetDenZaehlfehler(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineAbgelegtePartie(t, domain.StandLaeuft, 0)
	speicher.stolpert = map[string]error{
		"offenlegungszahl": errors.New("die Tafel schweigt"),
	}

	fehler := leitung.Loeschen(context.Background(), "p1")
	if fehler == nil || !strings.Contains(fehler.Error(), "schweigt") {
		t.Fatalf("erwartet den Zählfehler, bekam %v", fehler)
	}

	if len(speicher.geloescht) != 0 {
		t.Errorf("gelöscht wurde trotzdem: %v", speicher.geloescht)
	}
}

// **Eine beendete Partie lässt sich kopieren** — das ist der Sinn eines
// Übungszustands, und die Löschregel gilt dafür ausdrücklich nicht.
func TestPartieKopierenGehtAuchBeiEinerBeendeten(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineAbgelegtePartie(t, domain.StandBeendet, 1)

	kopie, fehler := leitung.Kopieren(context.Background(), "p1", "p2")
	if fehler != nil {
		t.Fatalf("Kopieren: %v", fehler)
	}

	if kopie.ID != "p2" {
		t.Fatalf("die Kopie heißt %q", kopie.ID)
	}

	if kopie.Stand != domain.StandLaeuft {
		t.Errorf("die Kopie steht auf %q", kopie.Stand)
	}

	// **Die Einwilligung der Vorlage geht nicht mit** — wörtlich „ohne alte
	// Peer-Zustimmungen". Am Doppel heißt das: die Kopie hat keine eigene
	// Zeile in der Tafel.
	zahl, fehler := speicher.Offenlegungszahl(context.Background(), "p2")
	if fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	if zahl != 0 {
		t.Errorf("die Kopie trägt %d Einwilligungen", zahl)
	}
}
