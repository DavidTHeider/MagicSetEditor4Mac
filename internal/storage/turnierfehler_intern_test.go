package storage

// **Die Fehlerwege des Turnierspeichers** — Werkstück 3 von Phase 4.7
// (#1294).
//
// **Ein interner Lauf**, weil er die Stolpertreiber braucht. Gegen eine echte
// SQLite-Datei sind diese Zweige nicht zu betreten: `RowsAffected` scheitert
// dort nie, und ein Verbund, der mitten in der Schleife abbricht, ebenso
// wenig.
//
// **Warum sie trotzdem geprüft werden.** Im Ernstfall entscheiden sie, ob die
// Anwendung „die Paarung stand nicht da" von „das Zählen ging schief"
// unterscheidet — und ob ein Lesefehler als Fehler ankommt oder als leeres
// Turnier.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg der Ablage reicht seinen Fehler durch.**
func TestDieTurnierablageReichtIhreFehlerDurch(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
		ruf     func(*Turnierablage) error
	}{
		{"das Auflisten scheitert", "stolper-fragen", func(a *Turnierablage) error {
			_, err := a.Turniere(probekontext(t), "")

			return err
		}},
		{
			"ein Turnier der Liste ist unlesbar", "stolper-turnierliste-scan",
			func(a *Turnierablage) error {
				_, err := a.Turniere(probekontext(t), "")

				return err
			},
		},
		{
			"die Liste bricht nach der Schleife ab", "stolper-turnierliste-ende",
			func(a *Turnierablage) error {
				_, err := a.Turniere(probekontext(t), "")

				return err
			},
		},
		{"das Abmelden scheitert", "stolper-schreiben", func(a *Turnierablage) error {
			return a.Aussteigen(probekontext(t), 1, true)
		}},
		{"das Löschen scheitert", "stolper-schreiben", func(a *Turnierablage) error {
			return a.Loeschen(probekontext(t), 1)
		}},
		{
			"das Zählen beim Löschen scheitert", "stolper-zaehlen",
			func(a *Turnierablage) error {
				return a.Loeschen(probekontext(t), 1)
			},
		},
		{"das Anlegen scheitert", "stolper-schreiben", func(a *Turnierablage) error {
			_, err := a.Anlegen(probekontext(t), domain.Turnier{Name: "T", Spiel: "magic"})

			return err
		}},
		{"das Lesen des Turniers scheitert", "stolper-fragen", func(a *Turnierablage) error {
			_, err := a.Turnier(probekontext(t), 1)

			return err
		}},
		{"die nächste Stelle ist unlesbar", "stolper-fragen", func(a *Turnierablage) error {
			_, err := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

			return err
		}},
		{"das Anmelden scheitert", "stolper-schreiben", func(a *Turnierablage) error {
			_, err := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

			return err
		}},
		{
			"der Zug der Anmeldung beginnt nicht", "stolper-runden-beginn",
			func(a *Turnierablage) error {
				_, err := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

				return err
			},
		},
		{
			"die Anmeldung schreibt sich nicht fest", "stolper-runden-commit",
			func(a *Turnierablage) error {
				_, err := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

				return err
			},
		},
		{
			"die Kennung des Teilnehmers ist unlesbar", "stolper-teilnehmer-kennung",
			func(a *Turnierablage) error {
				_, err := a.Anmelden(probekontext(t), 1, domain.Teilnehmer{Name: "Ada"})

				return err
			},
		},
		{"die Rundennummer ist unlesbar", "stolper-fragen", func(a *Turnierablage) error {
			_, err := a.RundeLegen(probekontext(t), 1, 1, []domain.Paarung{{A: 1, B: 2}})

			return err
		}},
		{"das Anlegen der Runde scheitert", "stolper-schreiben", func(a *Turnierablage) error {
			_, err := a.RundeLegen(probekontext(t), 1, 1, []domain.Paarung{{A: 1, B: 2}})

			return err
		}},
		{"das Ergebnis lässt sich nicht setzen", "stolper-schreiben", func(a *Turnierablage) error {
			return a.ErgebnisSetzen(probekontext(t), 1, 1, &domain.Ergebnis{SpieleA: 2})
		}},
		{
			"die Teilnehmer sind nicht abfragbar", "stolper-teilnehmer-abfrage",
			func(a *Turnierablage) error {
				_, err := a.Turnier(probekontext(t), 1)

				return err
			},
		},
		{
			"ein Teilnehmer ist unlesbar", "stolper-teilnehmer-scan",
			func(a *Turnierablage) error {
				_, err := a.Turnier(probekontext(t), 1)

				return err
			},
		},
		{
			"das Lesen bricht nach der Schleife ab", "stolper-teilnehmer-ende",
			func(a *Turnierablage) error {
				_, err := a.Turnier(probekontext(t), 1)

				return err
			},
		},
		{
			"der Verlauf ist nicht abfragbar", "stolper-verlauf-abfrage",
			func(a *Turnierablage) error {
				_, err := a.Turnier(probekontext(t), 1)

				return err
			},
		},
		{
			"eine Verlaufszeile ist unlesbar", "stolper-verlauf-scan",
			func(a *Turnierablage) error {
				_, err := a.Turnier(probekontext(t), 1)

				return err
			},
		},
		{
			"der Zug beginnt nicht", "stolper-runden-beginn",
			func(a *Turnierablage) error {
				_, err := a.RundeLegen(probekontext(t), 1, 1, []domain.Paarung{{A: 1, B: 2}})

				return err
			},
		},
		{
			"das Festschreiben scheitert", "stolper-runden-commit",
			func(a *Turnierablage) error {
				_, err := a.RundeLegen(probekontext(t), 1, 1, []domain.Paarung{{A: 1, B: 2}})

				return err
			},
		},
		{
			"die Kennung der Runde ist unlesbar", "stolper-runden-kennung",
			func(a *Turnierablage) error {
				_, err := a.RundeLegen(probekontext(t), 1, 1, []domain.Paarung{{A: 1, B: 2}})

				return err
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

			zweiterSchreibzaehler = 0

			pruefeStolperfehler(t, fall.ruf(NeueTurnierablage(db)))
		})
	}
}

// **Eine Runde ohne Paarung ist keine.**
func TestEineRundeOhnePaarungWirdAbgewiesen(t *testing.T) {
	if _, err := new(Turnierablage).RundeLegen(probekontext(t), 1, 1, nil); !errors.Is(
		err, domain.ErrValidation,
	) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Der Kartenname reicht seinen Lesefehler durch.**
//
// **Am [SetLeser] und nicht an der Turnierablage** — die Frage gehört
// dorthin, wo Karten liegen. Ohne diesen Lauf bliebe der Zweig unbetreten,
// und im Ernstfall käme statt eines Fehlers ein leerer Name in einen Abzug,
// der ein eingereichtes Dokument sein soll.
func TestDerKartennameReichtSeinenLesefehlerDurch(t *testing.T) {
	db, err := sql.Open("stolper-fragen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	_, fehler := NeuerSetLeser(db).NameZurKarte(probekontext(t), 1)
	pruefeStolperfehler(t, fehler)
}

// **Und der Feldwert reicht seinen ebenso durch** (#1376).
//
// Dieselbe Zusage wie darüber, und sie steht daneben statt darin: die beiden
// Nähte liegen in verschiedenen Dateien, und eine gemeinsame Prüfung ließe
// nicht sehen, welche von beiden ihren Fehler verschluckt.
func TestDerFeldwertReichtSeinenLesefehlerDurch(t *testing.T) {
	db, err := sql.Open("stolper-fragen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	_, fehler := NeuerSetLeser(db).FeldZurKarte(
		probekontext(t), 1, "casting cost")
	pruefeStolperfehler(t, fehler)
}
