// **Keine geklemmte Abfrage arbeitet ohne Profil** — Werkstück 2 von Phase
// 4.12 (#1351).
package storage_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// TestKeineGeklemmteAbfrageArbeitetOhneProfil geht jede Naht mit einem
// **nackten** Kontext an.
//
// # Warum eine Tabelle und nicht fünfzig Tests
//
// **Weil es eine Klasse ist und kein Fall** (§10). Jede dieser Abfragen
// klemmt auf `profil_id`, und jede holt das Profil vorher aus dem Kontext.
// Fünfzig Tests, die je einen davon prüfen, sagen fünfzigmal dasselbe — und
// der einundfünfzigste, den jemand vergisst, fällt niemandem auf.
//
// # Was er hält
//
// Der Aufruf **scheitert**, und die Meldung nennt das Profil. Ein Lesevorgang,
// der ohne Profil eine leere Liste gäbe, wäre der schlimmere Fehler: die
// Anwendung sähe aus, als wären alle Daten weg.
//
// # Was er **nicht** hält
//
// **Er sieht nicht, ob die Abfrage danach richtig klemmt.** Dass sie es tut,
// hält `internal/werkzeuge/profilklausel_test.go` am Quelltext; dass die
// Trennung wirkt, halten die Läufe in `trennung_test.go`.
func TestKeineGeklemmteAbfrageArbeitetOhneProfil(t *testing.T) {
	db := frischeDatenbank(t)

	leser := storage.NeuerSetLeser(db)
	speicher := storage.NeuerSetSpeicher(db)
	decks := storage.NeueDeckablage(db)
	formate := storage.NeueFormatablage(db)
	turniere := storage.NeueTurnierablage(db)
	sammlung := storage.NeuerSammlungsspeicher(db)
	listen := storage.NeueListenablage(db)
	ersetzungen := storage.NeueAutoersetzungsablage(db)

	// **Vier Nähte bekommen eine Welt.** `Anmelden`, `Aussteigen`,
	// `RundeLegen` und `ErgebnisSetzen` fragten ihr Profil bis #1815 erst in
	// `fortschreiben`; seitdem holen sie es zuerst. Die Welt bleibt, damit ihr
	// Scheitern hier nur am fehlenden Profil liegen kann.
	turnier, teilnehmer, runden, paarung := turnierMitRunde(t, turniere)

	faelle := map[string]func(context.Context) error{
		"Setkoepfe":      func(c context.Context) error { _, e := leser.Setkoepfe(c); return e },
		"Setkopf":        func(c context.Context) error { _, e := leser.Setkopf(c, 1); return e },
		"Quelle":         func(c context.Context) error { _, e := leser.Quelle(c, 1); return e },
		"HerkunftSetzen": func(c context.Context) error { return leser.HerkunftSetzen(c, 1, "x") },
		"SpielZurKarte":  func(c context.Context) error { _, e := leser.SpielZurKarte(c, 1); return e },
		// **Diese drei klemmen seit #1815 in ihrer eigenen Abfrage.**
		"NameZurKarte": func(c context.Context) error { _, e := leser.NameZurKarte(c, 1); return e },
		"FeldZurKarte": func(c context.Context) error { _, e := leser.FeldZurKarte(c, 1, "name"); return e },
		"KarteLesen":   func(c context.Context) error { _, e := leser.KarteLesen(c, 1); return e },
		// **Und die Vorprüfung jeder Massenänderung** fragt über das Set.
		"FelderSetzen": func(c context.Context) error {
			_, e := storage.NeuerKartenschreiber(db).FelderSetzen(c, []int64{1}, "name", "x")

			return e
		},
		"SetAnlegen": func(c context.Context) error {
			_, e := speicher.SetAnlegen(c, &domain.Set{Name: "x", Game: "y"}, nil)

			return e
		},
		"SetBenutzt":  func(c context.Context) error { return speicher.SetBenutzt(c, 1) },
		"SetLoeschen": func(c context.Context) error { return speicher.SetLoeschen(c, 1, "x") },
		"StilSetzen": func(c context.Context) error {
			return storage.NeueStilablage(db).StilSetzen(c, 1, "s", "t", nil)
		},
		"SetwertSetzen": func(c context.Context) error {
			_, _, e := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(c, 1, "f", "w")

			return e
		},

		"Decks": func(c context.Context) error { _, e := decks.Decks(c, "magic"); return e },
		"Deck":  func(c context.Context) error { _, e := decks.Deck(c, 1); return e },
		"DeckAnlegen": func(c context.Context) error {
			_, e := decks.Anlegen(c, domain.Deck{Name: "d", Spiel: "magic"})
			return e
		},
		"DeckAendern": func(c context.Context) error {
			_, e := decks.Aendern(c, domain.Deck{ID: 1, Name: "d", Spiel: "magic"})
			return e
		},
		"DeckLoeschen":  func(c context.Context) error { return decks.Loeschen(c, 1) },
		"DecksMitKarte": func(c context.Context) error { _, e := decks.DecksMitKarte(c, 1); return e },

		"Formate": func(c context.Context) error { _, e := formate.Formate(c, "magic"); return e },
		"Format":  func(c context.Context) error { _, e := formate.Format(c, 1); return e },
		"FormatAnlegen": func(c context.Context) error {
			_, e := formate.Anlegen(c, domain.Format{Name: "f", Spiel: "magic"})
			return e
		},
		"FormatAendern": func(c context.Context) error {
			_, e := formate.Aendern(c, domain.Format{ID: 1, Name: "f", Spiel: "magic"})
			return e
		},
		"FormatLoeschen": func(c context.Context) error { return formate.Loeschen(c, 1) },

		"Turniere": func(c context.Context) error { _, e := turniere.Turniere(c, "magic"); return e },
		"Turnier":  func(c context.Context) error { _, e := turniere.Turnier(c, 1); return e },
		"TurnierAnlegen": func(c context.Context) error {
			_, e := turniere.Anlegen(c, domain.Turnier{Name: "t", Spiel: "magic"})
			return e
		},
		"TurnierLoeschen": func(c context.Context) error { return turniere.Loeschen(c, 1) },
		// **Diese vier holen das Profil seit #1815 selbst** und reichen es an
		// `fortschreiben` weiter — vorher kam es erst dort, und ohne sie bliebe
		// ihr eigener Zweig unbetreten.
		"Anmelden": func(c context.Context) error {
			_, e := turniere.Anmelden(c, turnier, domain.Teilnehmer{Name: "wer"})

			return e
		},
		"Aussteigen": func(c context.Context) error { return turniere.Aussteigen(c, teilnehmer, true) },
		"RundeLegen": func(c context.Context) error {
			_, e := turniere.RundeLegen(c, turnier, runden+1, []domain.Paarung{{A: teilnehmer}})

			return e
		},
		"ErgebnisSetzen": func(c context.Context) error {
			return turniere.ErgebnisSetzen(c, 1, paarung, &domain.Ergebnis{SpieleA: 2})
		},

		"Sammlungsliste": func(c context.Context) error { _, e := sammlung.Liste(c); return e },
		"SammlungFuer":   func(c context.Context) error { _, e := sammlung.Fuer(c, "p"); return e },
		"ExemplarAnlegen": func(c context.Context) error {
			_, e := sammlung.Anlegen(c, &domain.Exemplar{Printing: "p", Menge: 1})
			return e
		},
		"ExemplarAendern": func(c context.Context) error {
			return sammlung.Aendern(c, &domain.Exemplar{ID: 1, Printing: "p", Menge: 1})
		},
		"ExemplarLoeschen": func(c context.Context) error { return sammlung.Loeschen(c, 1, 1) },

		"Spalten": func(c context.Context) error { _, e := listen.Spalten(c, "magic"); return e },
		"SpaltenSetzen": func(c context.Context) error {
			return listen.SpaltenSetzen(c, "magic", []domain.Spaltenwahl{{Feld: "name"}})
		},
		"Sortierung": func(c context.Context) error { _, e := listen.Sortierung(c, "magic"); return e },
		"SortierungSetzen": func(c context.Context) error {
			return listen.SortierungSetzen(c, "magic", domain.Kartenordnung{Feld: "name"})
		},

		"Einstellungen":      func(c context.Context) error { _, e := storage.NeueEinstellungen(db).Alle(c); return e },
		"EinstellungSetzen":  func(c context.Context) error { return storage.NeueEinstellungen(db).Setzen(c, "s", "w") },
		"Exportoptionen":     func(c context.Context) error { _, e := storage.NeueExportoptionen(db).Fuer(c, "magic"); return e },
		"ExportoptionSetzen": func(c context.Context) error { return storage.NeueExportoptionen(db).Setzen(c, "magic", "s", "w") },

		"Ersetzungsregeln": func(c context.Context) error { _, e := ersetzungen.Regeln(c, "magic"); return e },
		"ErsetzungSetzen": func(c context.Context) error {
			return ersetzungen.Setzen(c, "magic", domain.Autoersetzungsregel{Muster: "a", Ersatz: "b"})
		},
		"ErsetzungLoeschen":        func(c context.Context) error { return ersetzungen.Loeschen(c, "magic", "a") },
		"ErsetzungenZuruecksetzen": func(c context.Context) error { return ersetzungen.Zuruecksetzen(c, "magic") },
		"Ersetzungsschalter":       func(c context.Context) error { _, e := ersetzungen.Schalter(c, "magic"); return e },
		"ErsetzungsschalterSetzen": func(c context.Context) error { return ersetzungen.SchalterSetzen(c, "magic", true) },
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			// **Nackt und nicht abgebrochen**: ein abgebrochener Kontext
			// scheiterte auch ohne die Klemme, und der Test sagte nichts.
			fehler := ruf(t.Context())
			if fehler == nil {
				t.Fatal("der Aufruf glückte ohne Profil")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("kein Validierungsfehler: %v", fehler)
			}

			if !strings.Contains(fehler.Error(), "Profil") {
				t.Errorf("die Meldung nennt das Profil nicht: %v", fehler)
			}
		})
	}
}

// turnierMitRunde legt ein Turnier mit Teilnehmer und Runde an — **mit**
// Profil, denn hier wird die Welt gebaut und nicht geprüft.
//
// Sie gibt zusätzlich den Rundenstand, gegen den `RundeLegen` rechnet, und
// eine Paarungskennung für `ErgebnisSetzen`.
func turnierMitRunde(t *testing.T, ablage *storage.Turnierablage) (
	turnierID, teilnehmerID int64, runden int, paarungID int64,
) {
	t.Helper()

	turnier, fehler := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Probe", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("das Turnier anlegen: %v", fehler)
	}

	teilnehmer, anmeldefehler := ablage.Anmelden(probekontext(t), turnier.ID,
		domain.Teilnehmer{Name: "wer"})
	if anmeldefehler != nil {
		t.Fatalf("anmelden: %v", anmeldefehler)
	}

	runde, rundenfehler := ablage.RundeLegen(probekontext(t), turnier.ID, 1,
		[]domain.Paarung{{A: teilnehmer.ID}})
	if rundenfehler != nil {
		t.Fatalf("die Runde legen: %v", rundenfehler)
	}

	return turnier.ID, teilnehmer.ID, runde.Nummer, runde.Paarungen[0].ID
}
