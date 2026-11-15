package storage

// **Was der Abschluss meldet, wenn etwas stolpert** — Werkstück `T6` von
// Phase 5.6 (#1471).
//
// **Eine zweite Datei neben `stolperabschluss_test.go`**, weil beide zusammen
// über die Zeilengrenze gingen (§2). Der Schnitt folgt der Frage: dort, **wo**
// gestolpert wird und was der Treiber antwortet — hier, **was der Aufrufer
// davon merkt**.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineAbgangszeile ist die Zeile, mit der die Helfer gefahren werden.
func eineAbgangszeile(menge int) domain.Journalzeile {
	return domain.Journalzeile{
		Art: domain.JournalAbgang, Anlass: domain.AnlassTausch,
		Vorgang: "v-1", Printing: "bolt-m10", Menge: menge,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil, ExemplarID: 1,
	}
}

// **Jeder Weg des Abgangs reicht seinen Fehler durch.**
func TestJederWegDesAbgangsReichtSeinenFehlerDurch(t *testing.T) {
	for name, fall := range map[string]struct {
		treiber string
		menge   int
	}{
		"die Besitzzeile lässt sich nicht lesen": {"stolper-besitz-lesen", 2},
		"eine Zusage lässt sich nicht lesen":     {"stolper-deckung-scan", 2},
		"die Zusagen brechen ab":                 {"stolper-deckung-ende", 2},
		"das Mindern scheitert":                  {"stolper-mindern", 2},
		"das Nachzählen beim Mindern klagt":      {"stolper-mindern-zaehlen", 2},
		"das Löschen scheitert":                  {"stolper-bestand-loesch", 4},
		"das Nachzählen beim Löschen klagt":      {"stolper-loesch-zaehlen", 4},
		"es liegt weniger da, als zugesagt ist":  {"stolper-zuwenig", 2},
		"das Mindern trifft keine Zeile":         {"stolper-stumm", 2},
		"die Zusagen lassen sich nicht abfragen": {"stolper-deckung-fragen", 2},
	} {
		t.Run(name, func(t *testing.T) {
			db := stolperbestand(t, fall.treiber)

			if fehler := bestandAbgang(t.Context(), db, 1, "v-1",
				eineAbgangszeile(fall.menge)); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Und jeder Weg des Zugangs ebenso.**
func TestJederWegDesZugangsReichtSeinenFehlerDurch(t *testing.T) {
	zeile := eineAbgangszeile(1)
	zeile.Art, zeile.ExemplarID = domain.JournalZugang, 0

	for name, treiber := range map[string]string{
		"der Druck lässt sich nicht anlegen": "stolper-zugang-druck",
		"das Eintragen scheitert":            "stolper-zugang",
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := bestandZugang(
				t.Context(), stolperbestand(t, treiber), 1, zeile,
			); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Und jeder Weg des Journals und des Belegs.**
func TestJederWegDesJournalsReichtSeinenFehlerDurch(t *testing.T) {
	zeile := eineAbgangszeile(1)

	for name, fall := range map[string]struct {
		treiber string
		fahre   func(*sql.DB) error
	}{
		"das Schreiben scheitert": {"stolper-journal-schreib", func(db *sql.DB) error {
			return journalSchreiben(t.Context(), db, 1,
				[]domain.Journalzeile{zeile})
		}},
		"eine Zeile lässt sich nicht lesen": {"stolper-journal-scan", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Journal(probekontext(t), "v-1")

			return fehler
		}},
		"die Zeilen brechen ab": {"stolper-journal-ende", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Journal(probekontext(t), "v-1")

			return fehler
		}},
		"eine Uhr ist keine": {"stolper-journal-uhr", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Journal(probekontext(t), "v-1")

			return fehler
		}},
		"die Abfrage geht gar nicht": {"stolper-fragen", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Journal(probekontext(t), "v-1")

			return fehler
		}},
		"der Beleg trägt keine Uhr": {"stolper-beleg-uhr", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Beleg(probekontext(t), "v-1")

			return fehler
		}},
		"der Beleg lässt sich nicht lesen": {"stolper-lesen", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Beleg(probekontext(t), "v-1")

			return fehler
		}},
		"die Uhr des Fremdbelegs ist keine": {"stolper-fremdbeleg-uhr", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Beleg(probekontext(t), "v-1")

			return fehler
		}},
		"die Uhr des Befunds ist keine": {"stolper-befund-uhr", func(db *sql.DB) error {
			_, fehler := NeueAbschlussablage(db).Beleg(probekontext(t), "v-1")

			return fehler
		}},
		"der Fremdbeleg lässt sich nicht ablegen": {"stolper-fremdbeleg", func(db *sql.DB) error {
			return NeueAbschlussablage(db).Gegenzeichnen(probekontext(t),
				domain.Fremdbeleg{Vorgang: "v-1", Von: "bert", Abdruck: "a1"})
		}},
		"die Gegenzeichnung beginnt nicht": {"stolper-abschlussbeginn", func(db *sql.DB) error {
			return NeueAbschlussablage(db).Gegenzeichnen(probekontext(t),
				domain.Fremdbeleg{Vorgang: "v-1", Von: "bert", Abdruck: "a1"})
		}},
		"das Gegenzeichnen scheitert": {"stolper-gegenzeichnen", func(db *sql.DB) error {
			return NeueAbschlussablage(db).Gegenzeichnen(probekontext(t),
				domain.Fremdbeleg{Vorgang: "v-1", Von: "bert", Abdruck: "a1"})
		}},
		"das Nachzählen beim Gegenzeichnen klagt": {"stolper-gegen-zaehlen", func(db *sql.DB) error {
			return NeueAbschlussablage(db).Gegenzeichnen(probekontext(t),
				domain.Fremdbeleg{Vorgang: "v-1", Von: "bert", Abdruck: "a1"})
		}},
		"das Einlösen der Zusagen scheitert": {"stolper-zusagen-loesen", func(db *sql.DB) error {
			return zusagenEinloesen(t.Context(), db, 1, "v-1")
		}},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := fall.fahre(stolperbestand(t, fall.treiber)); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Und jeder Weg des Nachschreibens.**
func TestJederWegDesNachschreibensReichtSeinenFehlerDurch(t *testing.T) {
	bauen := func() domain.Handelsabschluss {
		return domain.Handelsabschluss{
			Beleg:        domain.Handelsbeleg{Vorgang: "v-1", Abdruck: "a"},
			Unterschrift: "u",
			Vermerk: &domain.Inboxeintrag{
				Schluessel: "s", Nutzlastabdruck: "n", Ergebnis: "{}",
			},
			Ausgang: &domain.Belegnachricht{
				NachrichtID: "n-1", Empfaenger: "b", Nachricht: "{}",
			},
		}
	}

	for name, treiber := range map[string]string{
		"der Beleg lässt sich nicht ablegen":      "stolper-beleg-schreib",
		"der Vermerk lässt sich nicht schreiben":  "stolper-vermerk",
		"die Belegnachricht lässt sich nicht ein": "stolper-belegausgang",
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := nachschreiben(
				t.Context(), stolperbestand(t, treiber), 1, bauen()); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Und jeder Weg im Abschluss selbst** — von `BEGIN` bis `COMMIT`.
//
// **Hier wird der ganze Weg gefahren und nicht ein Helfer**, weil genau das
// die Frage ist: bleibt an einer Stelle mittendrin etwas stehen? Die Helfer
// oben sagen, dass jeder einzeln redet; dieser Lauf sagt, dass der Weg
// darüber es auch tut.
func TestJederWegImAbschlussReichtSeinenFehlerDurch(t *testing.T) {
	bauen := func() domain.Handelsabschluss {
		angebot := domain.Handelsangebot{
			Vertragsfassung: 1, Vorgang: "v-1", Revision: 1,
			Art: domain.HandelTausch, Initiator: "a",
			Parteien: []string{"a", "b"}, Zustand: domain.HandelLokalVerbucht,
		}

		return domain.Handelsabschluss{
			Angebot: angebot,
			Beleg: domain.Handelsbeleg{
				Vorgang: "v-1", Abdruck: angebot.Abdruck(),
				Zeilen: []domain.Journalzeile{eineAbgangszeile(2)},
			},
			Unterschrift: "u",
		}
	}

	for name, treiber := range map[string]string{
		"der Zug beginnt nicht":        "stolper-abschlussbeginn",
		"die Belegsuche stolpert":      "stolper-belegsuche",
		"der Bestand stolpert":         "stolper-mindern",
		"das Journal stolpert":         "stolper-journal-schreib",
		"das Einlösen stolpert":        "stolper-zusagen-loesen",
		"die Revision stolpert":        "stolper-revision",
		"der Beleg stolpert":           "stolper-beleg-schreib",
		"das Abschließen selbst klagt": "stolper-abschlusscommit",
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := NeueAbschlussablage(stolperbestand(t, treiber)).
				Abschliessen(probekontext(t), bauen()); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Und die Zusagen eines Vorgangs melden ihre Lesefehler.**
//
// **Eine eigene Stelle**, weil die Abfrage selbst scheitern muss: die Zeilen
// dahinter prüft `stolperreservierungen_test.go` schon über den anderen Weg —
// beide gehen durch denselben Ableser.
func TestDieZusagenEinesVorgangsMeldenIhreLesefehler(t *testing.T) {
	if _, fehler := NeueReservierungsablage(
		stolperbestand(t, "stolper-vorgangszusagen")).
		FuerVorgang(probekontext(t), "v-1"); fehler == nil {
		t.Error("der Lesefehler ging verloren")
	}
}
