package storage

// **Die Fehlerwege der Turnierfassung** — Befund 7 des Nachlaufs über Phase
// 4.7.
//
// **Eine eigene Datei**, damit `turnierfehler_intern_test.go` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie
// der **Turnierweg** seine Fehler durchreicht, hier, wie seine **Buchführung
// über die Fassung** es tut.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg der Buchführung reicht seinen Fehler durch.**
func TestDieFassungsbuchfuehrungReichtIhreFehlerDurch(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
		ruf     func(*Turnierablage) error
	}{
		{
			"das Fortschreiben beim Anmelden scheitert", "stolper-fassung-schreiben",
			func(a *Turnierablage) error {
				_, err := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

				return err
			},
		},
		{
			"das Fortschreiben beim Legen scheitert", "stolper-fassung-schreiben",
			func(a *Turnierablage) error {
				_, err := a.RundeLegen(probekontext(t), 1, 1,
					[]domain.Paarung{{A: 1, B: 2}})

				return err
			},
		},
		{
			"das Abmelden beginnt nicht", "stolper-abmelden-beginnen",
			func(a *Turnierablage) error {
				return a.Aussteigen(probekontext(t), 1, true)
			},
		},
		{
			"das Turnier des Teilnehmers ist unlesbar", "stolper-teilnehmer-turnier",
			func(a *Turnierablage) error {
				return a.Aussteigen(probekontext(t), 1, true)
			},
		},
		{
			"das Fortschreiben beim Abmelden scheitert", "stolper-fassung-schreiben",
			func(a *Turnierablage) error {
				return a.Aussteigen(probekontext(t), 1, true)
			},
		},
		{
			"das Eintragen beginnt nicht", "stolper-eintragen-beginnen",
			func(a *Turnierablage) error {
				return a.ErgebnisSetzen(probekontext(t), 1, 1,
					&domain.Ergebnis{SpieleA: 2})
			},
		},
		{
			"das Turnier der Paarung ist unlesbar", "stolper-paarungs-turnier",
			func(a *Turnierablage) error {
				return a.ErgebnisSetzen(probekontext(t), 1, 1,
					&domain.Ergebnis{SpieleA: 2})
			},
		},
		{
			"das Fortschreiben gegen die Fassung scheitert",
			"stolper-fassung-schreiben",
			func(a *Turnierablage) error {
				return a.ErgebnisSetzen(probekontext(t), 1, 1,
					&domain.Ergebnis{SpieleA: 2})
			},
		},
		{
			"das Zählen der Fassung scheitert", "stolper-fassung-zaehlen",
			func(a *Turnierablage) error {
				return a.ErgebnisSetzen(probekontext(t), 1, 1,
					&domain.Ergebnis{SpieleA: 2})
			},
		},
		{
			"das Nachschlagen der Fassung scheitert",
			"stolper-fassung-nachschlagen",
			func(a *Turnierablage) error {
				return a.ErgebnisSetzen(probekontext(t), 1, 1,
					&domain.Ergebnis{SpieleA: 2})
			},
		},
		{
			"das Schreiben des Ergebnisses scheitert", "stolper-schreiben",
			func(a *Turnierablage) error {
				return a.ErgebnisSetzen(probekontext(t), 1, 1,
					&domain.Ergebnis{SpieleA: 2})
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.ruf(NeueTurnierablage(db)))
		})
	}
}

// **Ein Turnier, das es nicht gibt, ist etwas anderes als eines, das
// überholt ist.**
//
// Zwei Antworten auf denselben Fehlschlag, und sie haben zwei Abhilfen:
// „gibt es nicht" heißt aufräumen, „jemand anderes war schneller" heißt
// nachladen (§8). Wer sie zusammenlegte, schickte jemanden auf die falsche
// Suche.
func TestEineFehlendeFassungIstNichtGefunden(t *testing.T) {
	db, err := sql.Open("stolper-fassung-fehlt", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	fehler := NeueTurnierablage(db).ErgebnisSetzen(probekontext(t), 1, 1,
		&domain.Ergebnis{SpieleA: 2})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", fehler)
	}
}
