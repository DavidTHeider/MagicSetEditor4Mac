package storage

// **Die Fehlerwege des Nutzerspeichers** — Werkstück 1 von Phase 4.12
// (#1351).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt kein echter Treiber her: eine Abfrage, die scheitert; eine Zeile, die
// sich nicht lesen lässt; ein Abbruch **nach** der Schleife.

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg der Nutzerablage reicht seinen Fehler durch.**
func TestJederNutzerwegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*Nutzerablage) error
	}{
		"die Nutzerabfrage scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.Nutzer(probekontext(t))

			return err
		}},
		"eine Nutzerzeile lässt sich nicht lesen": {"stolper-nutzer-scan", func(a *Nutzerablage) error {
			_, err := a.Nutzer(probekontext(t))

			return err
		}},

		// **Und dieselbe krumme Zeile bei der Suche nach einer Kennung** —
		// Werkstück 6 (#1351). Sie liest **eine** Zeile und unterscheidet
		// „gibt es nicht" von „lässt sich nicht lesen"; ohne diesen Fall
		// bliebe die zweite Hälfte der Unterscheidung unbetreten und ein
		// Lesefehler käme als [domain.ErrNotFound] an der Oberfläche an.
		"die Kennungssuche liest eine krumme Zeile": {
			"stolper-nutzer-scan", func(a *Nutzerablage) error {
				_, err := a.NutzerZuKennung(probekontext(t), "wer@example.org")

				return err
			},
		},
		"die Nutzerzeilen brechen ab": {"stolper-nutzer-ende", func(a *Nutzerablage) error {
			_, err := a.Nutzer(probekontext(t))

			return err
		}},
		"die Profilabfrage scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.Profile(probekontext(t), 1)

			return err
		}},
		"eine Profilzeile lässt sich nicht lesen": {"stolper-profil-scan", func(a *Nutzerablage) error {
			_, err := a.Profile(probekontext(t), 1)

			return err
		}},
		"die Profilzeilen brechen ab": {"stolper-profil-ende", func(a *Nutzerablage) error {
			_, err := a.Profile(probekontext(t), 1)

			return err
		}},
		// **Und alle Profile für den Neuanfang** — Befund 1 des Zweitblicks
		// über Phase G (#2282).
		"die Abfrage aller Profile scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := AlleProfile(probekontext(t), a.db)

			return err
		}},
		"eine Zeile aller Profile lässt sich nicht lesen": {"stolper-profil-scan", func(a *Nutzerablage) error {
			_, err := AlleProfile(probekontext(t), a.db)

			return err
		}},
		"die Zeilen aller Profile brechen ab": {"stolper-profil-ende", func(a *Nutzerablage) error {
			_, err := AlleProfile(probekontext(t), a.db)

			return err
		}},
		"das Anlegen scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			_, err := a.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})

			return err
		}},
		"die Nutzersuche scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.ProfilAnlegen(probekontext(t), domain.Profil{NutzerID: 1, Name: "X"})

			return err
		}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.fahre(NeueNutzerablage(db)))
		})
	}
}

// **Und die Wege, die kein echter Treiber hergibt.**
//
// `LastInsertId` scheitert dort nie, `Begin` und `Commit` ebenso wenig — nur
// ein erfundener Treiber zeigt, dass sie behandelt sind.
func TestDieUnsichtbarenNutzerwegeSindBehandelt(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*Nutzerablage) error
	}{
		"die Nutzerkennung": {"stolper-nutzer-kennung", func(a *Nutzerablage) error {
			_, err := a.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})

			return err
		}},
		"die Profilkennung": {"stolper-profil-kennung", func(a *Nutzerablage) error {
			_, err := a.ProfilAnlegen(probekontext(t), domain.Profil{NutzerID: 1, Name: "X"})

			return err
		}},
		"der Beginn": {"stolper-eintragen-beginnen", func(a *Nutzerablage) error {
			_, err := a.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})

			return err
		}},
		"das Festschreiben": {"stolper-commit", func(a *Nutzerablage) error {
			_, err := a.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})

			return err
		}},
		"das zweite Einfügen": {"stolper-zweites-schreiben", func(a *Nutzerablage) error {
			_, err := a.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})

			return err
		}},
		"das Einfügen des Profils": {"stolper-schreiben", func(a *Nutzerablage) error {
			_, err := a.ProfilAnlegen(probekontext(t), domain.Profil{NutzerID: 1, Name: "X"})

			return err
		}},
		// **Die Suche glückt, das Schreiben scheitert.** Über einen
		// geschlossenen Bestand wäre schon die Suche rot, und der zweite Weg
		// bliebe unbetreten (§3).
		"das Setzen des Zeigers": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.ProfilzeigerSetzen(probekontext(t), 1)
		}},
		"das Lesen des Zeigers": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.Profilzeiger(probekontext(t))

			return err
		}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			// **Der Zähler je Fall** — ohne die Zeile traf „das zweite
			// Einfügen" seine Anweisung nur, wenn vorher niemand gezählt
			// hatte; seit `RETURNING` zählt schon die erste mit.
			zweiterSchreibzaehler = 0

			pruefeStolperfehler(t, fall.fahre(NeueNutzerablage(db)))
		})
	}
}

// **Jeder Weg der Anbieterkonten reicht seinen Fehler durch.**
//
// **Über Stolpertreiber und nicht über einen geschlossenen Bestand**: dort
// scheiterte schon die erste Abfrage, und die Wege dahinter blieben unbetreten
// (§3).
func TestJederKontoWegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*Nutzerablage) error
	}{
		"die Nutzersuche scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.KontoVerbinden(probekontext(t), domain.Anbieterkonto{
				NutzerID: 1, Anbieter: "github", Subjekt: "1",
			})

			return err
		}},
		"das Einfügen scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			_, err := a.KontoVerbinden(probekontext(t), domain.Anbieterkonto{
				NutzerID: 1, Anbieter: "github", Subjekt: "1",
			})

			return err
		}},
		"die Kennung lässt sich nicht lesen": {"stolper-nutzer-kennung", func(a *Nutzerablage) error {
			_, err := a.KontoVerbinden(probekontext(t), domain.Anbieterkonto{
				NutzerID: 1, Anbieter: "github", Subjekt: "1",
			})

			return err
		}},
		"die Suche nach dem Nutzer scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.NutzerZuKonto(probekontext(t), "github", "1")

			return err
		}},
		"die Kontenabfrage scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.Konten(probekontext(t), 1)

			return err
		}},
		"eine Kontozeile lässt sich nicht lesen": {"stolper-konto-scan", func(a *Nutzerablage) error {
			_, err := a.Konten(probekontext(t), 1)

			return err
		}},
		"die Kontozeilen brechen ab": {"stolper-konto-ende", func(a *Nutzerablage) error {
			_, err := a.Konten(probekontext(t), 1)

			return err
		}},
		"das Lösen scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.KontoLoesen(probekontext(t), 1, "github")
		}},
		"das Zählen beim Lösen scheitert": {"stolper-zaehlen", func(a *Nutzerablage) error {
			return a.KontoLoesen(probekontext(t), 1, "github")
		}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.fahre(NeueNutzerablage(db)))
		})
	}
}

// **Jeder Weg der Nutzerpflege reicht seinen Fehler durch.**
func TestJederPflegewegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*Nutzerablage) error
	}{
		"das Ändern des Nutzers scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			_, err := a.NutzerAendern(probekontext(t), domain.Nutzer{ID: 1, Name: "X"})

			return err
		}},
		"das Zählen beim Ändern des Nutzers scheitert": {"stolper-zaehlen", func(a *Nutzerablage) error {
			_, err := a.NutzerAendern(probekontext(t), domain.Nutzer{ID: 1, Name: "X"})

			return err
		}},
		"das Ändern des Profils scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			_, err := a.ProfilAendern(probekontext(t), domain.Profil{ID: 1, NutzerID: 1, Name: "X"})

			return err
		}},
		"das Zählen beim Ändern des Profils scheitert": {"stolper-zaehlen", func(a *Nutzerablage) error {
			_, err := a.ProfilAendern(probekontext(t), domain.Profil{ID: 1, NutzerID: 1, Name: "X"})

			return err
		}},
		"das Beginnen scheitert": {"stolper-nutzer-beginn", func(a *Nutzerablage) error {
			return a.NutzerLoeschen(probekontext(t), 1, "Ich")
		}},
		"das Nachschlagen scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			return a.NutzerLoeschen(probekontext(t), 1, "Ich")
		}},
		"das Zählen der Nutzer scheitert": {"stolper-nutzer-zaehlen", func(a *Nutzerablage) error {
			return a.NutzerLoeschen(probekontext(t), 1, "Ich")
		}},
		"das Zählen der Profile scheitert": {"stolper-profil-zaehlen", func(a *Nutzerablage) error {
			return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 0)
		}},
		"das Löschen des Nutzers scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.NutzerLoeschen(probekontext(t), 1, "Ich")
		}},
		"das Löschen des Profils scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 0)
		}},
		"das Beginnen beim Profil scheitert": {"stolper-nutzer-beginn", func(a *Nutzerablage) error {
			return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 0)
		}},

		// **Und das Setzen des Ausweichzeigers** — Befund 1 des Zweitblicks
		// über die Phasen 4.0 bis 4.13 (2026-09-07). Es steht seither **in**
		// der Löschtransaktion und hat damit einen eigenen Fehlerweg; jeder
		// Test davor gab `ausweich: 0` und ging daran vorbei.
		"das Setzen des Ausweichzeigers scheitert": {
			"stolper-schreiben", func(a *Nutzerablage) error {
				return a.ProfilLoeschen(probekontext(t), 1, "Mein Profil", 2)
			},
		},

		// **Die vier Wege von [Nutzerablage.ZeigerSetzen]**, aus demselben
		// Befund. Sie stellt beide Zeiger in einer Transaktion; jeder Schritt
		// darin kann scheitern, und keiner darf einen halben Stand lassen.
		"das Beginnen der Zeigerumstellung scheitert": {
			"stolper-nutzer-beginn", func(a *Nutzerablage) error {
				return a.ZeigerSetzen(probekontext(t), 1, 1)
			},
		},
		"die Suche nach dem Nutzer scheitert": {
			"stolper-fragen", func(a *Nutzerablage) error {
				return a.ZeigerSetzen(probekontext(t), 1, 1)
			},
		},
		"das Schreiben der Zeiger scheitert": {
			"stolper-schreiben", func(a *Nutzerablage) error {
				return a.ZeigerSetzen(probekontext(t), 1, 1)
			},
		},
		"das Festschreiben der Zeiger scheitert": {
			"stolper-nutzer-commit", func(a *Nutzerablage) error {
				return a.ZeigerSetzen(probekontext(t), 1, 1)
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.fahre(NeueNutzerablage(db)))
		})
	}
}

// pruefeStolperfehler verlangt einen Fehler — **und zwar den der Stelle**:
// einer an der Spaltenzahl heißt, dass die Vorrichtung nicht mehr zur
// Anweisung passt, und dann prüft der Lauf etwas anderes als seinen Fall.
// So standen die Profilzeilen bei acht Spalten, als elf gelesen wurden, und
// das Einfügen bei zwei, als drei zurückkamen — grün, an der falschen
// Anweisung (Zweitblick über Phase B, Befund 3).
func pruefeStolperfehler(t *testing.T, fehler error) {
	t.Helper()

	switch {
	case fehler == nil:
		t.Error("der Fehler ging verloren")
	case strings.Contains(fehler.Error(), "destination arguments"):
		t.Errorf("gescheitert an der Spaltenzahl der Vorrichtung, nicht an der Stelle: %v", fehler)
	}
}
