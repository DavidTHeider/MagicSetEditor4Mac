package service_test

// **Die übrigen Wege der Turnierleitung** — Werkstück 4 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turnierlauf_test.go` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort steht der **ganze
// Ablauf**, hier stehen die Wege daneben — auflisten, abmelden, löschen, und
// was passiert, wenn der Speicher schweigt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Turniere lassen sich auflisten — je Spiel und alle.**
func TestTurniereLassenSichAuflisten(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	werk.turnierAnlegen(t)

	if _, err := werk.leitung.Anlegen(t.Context(),
		domain.Turnier{Name: "Anderswo", Spiel: "vs"}); err != nil {
		t.Fatalf("das zweite: %v", err)
	}

	alle, err := werk.leitung.Turniere(t.Context(), "")
	if err != nil {
		t.Fatalf("auflisten: %v", err)
	}

	if len(alle) != 2 {
		t.Errorf("%d Turniere ohne Spiel, erwartet 2", len(alle))
	}

	magic, err := werk.leitung.Turniere(t.Context(), "magic")
	if err != nil {
		t.Fatalf("auflisten: %v", err)
	}

	if len(magic) != 1 || magic[0].Spiel != "magic" {
		t.Errorf("%+v", magic)
	}
}

// **Ein Teilnehmer steigt aus und bleibt in der Rangliste.**
//
// Appendix C: seine Matchquote geht weiter in die Gegnerquote jedes Gegners
// ein, den er hatte. Wer ihn entfernte, veränderte die Rangliste aller
// anderen.
func TestEinAusgestiegenerBleibtGezaehlt(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	wer := werk.anmelden(t, turnier.ID, "Ada", "Bert")

	if err := werk.leitung.Aussteigen(t.Context(), turnier.ID, wer[1], true); err != nil {
		t.Fatalf("aussteigen: %v", err)
	}

	rangliste, err := werk.leitung.Rangliste(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("Rangliste: %v", err)
	}

	if len(rangliste) != 2 {
		t.Fatalf("%d Ränge, erwartet 2", len(rangliste))
	}

	var gefunden bool

	for _, rang := range rangliste {
		if rang.TeilnehmerID == wer[1] {
			gefunden = rang.Ausgestiegen
		}
	}

	if !gefunden {
		t.Error("der Ausgestiegene ist in der Rangliste nicht als solcher zu sehen")
	}
}

// **Ein Turnier lässt sich löschen.**
func TestEinTurnierLaesstSichLoeschen(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	if err := werk.leitung.Loeschen(t.Context(), turnier.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if _, err := werk.leitung.Turnier(t.Context(), turnier.ID); !errors.Is(
		err, domain.ErrNotFound,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Jeder Weg reicht einen Speicherfehler durch.**
//
// **Zum Ergebnis gehört auch das Nein** (§10): ein Weg, von dem nur der
// Erfolgsfall belegt ist, hat die Hälfte ausgelassen.
func TestJederWegReichtEinenSpeicherfehlerDurch(t *testing.T) {
	t.Parallel()

	schweigen := errors.New("der Bestand schweigt")

	faelle := map[string]func(*turnierwerk) error{
		"anlegen": func(w *turnierwerk) error {
			_, err := w.leitung.Anlegen(t.Context(),
				domain.Turnier{Name: "T", Spiel: "magic"})

			return err
		},
		"lesen": func(w *turnierwerk) error {
			_, err := w.leitung.Turnier(t.Context(), 1)

			return err
		},
		"auflisten": func(w *turnierwerk) error {
			_, err := w.leitung.Turniere(t.Context(), "")

			return err
		},
		"anmelden": func(w *turnierwerk) error {
			_, err := w.leitung.Anmelden(t.Context(), 1,
				domain.Teilnehmer{Name: "Ada"})

			return err
		},
		"aussteigen": func(w *turnierwerk) error {
			return w.leitung.Aussteigen(t.Context(), 1, 1, true)
		},
		"aussteigen, Lesen": func(w *turnierwerk) error {
			// **Ein eigener Fall**: der Weg liest **zuerst** das Turnier, und
			// ein Fehler dort kommt aus einem anderen Zweig als einer beim
			// Schreiben. Die Vorrichtung lässt beides zugleich scheitern —
			// dieser Fall belegt, dass auch der zweite Zweig meldet.
			return w.leitung.Aussteigen(t.Context(), 1, 1, false)
		},
		"Rangliste": func(w *turnierwerk) error {
			_, err := w.leitung.Rangliste(t.Context(), 1)

			return err
		},
		"vorschlagen": func(w *turnierwerk) error {
			_, err := w.leitung.Paarungsvorschlag(t.Context(), 1)

			return err
		},
		"Runde legen": func(w *turnierwerk) error {
			_, err := w.leitung.RundeLegen(t.Context(), 1)

			return err
		},
		"Ergebnis setzen": func(w *turnierwerk) error {
			return w.leitung.ErgebnisSetzen(t.Context(), 1, 1, 1,
				&domain.Ergebnis{SpieleA: 2})
		},
		"löschen": func(w *turnierwerk) error {
			return w.leitung.Loeschen(t.Context(), 1)
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werk := neuesTurnierwerk(t)
			werk.turniere.fehler = schweigen

			if err := ruf(werk); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Ein Kartenname, der sich nicht lesen lässt, verhindert die Anmeldung.**
//
// **Und keine halbe Liste**: ein Abzug, dem eine Zeile fehlt, wäre ein
// eingereichtes Dokument, das nie so eingereicht wurde.
func TestEinUnlesbarerKartennameVerhindertDieAnmeldung(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.decks.deck = domain.Deck{
		ID: 7, Spiel: "magic", Plaetze: []domain.Deckplatz{
			{KarteID: 3, Menge: 1, Bereich: domain.BereichHauptdeck},
		},
	}
	werk.karten.fehler = errors.New("die Karte schweigt")

	if _, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{Name: "Ada", DeckID: 7}); err == nil {
		t.Error("angemeldet, erwartet abgewiesen")
	}

	if werk.turniere.angemeldet != 0 {
		t.Error("angemeldet wurde trotzdem")
	}
}

// **Eine offene Paarding hält das Turnier an — auch eine ältere.**
//
// # Der Befund
//
// Der erste Anlauf sah nur die **letzte** Runde an und begründete das damit,
// eine ältere offene Runde sei „ein Befund für den Turnierleiter, aber kein
// Grund anzuhalten". Die Begründung trägt nicht: eine Paarung ohne Ergebnis
// zählt in der Rangliste **gar nicht**, und die nächste Runde paart dann nach
// einem unvollständigen Stand — genau das, wogegen die Prüfung steht.
//
// **Angestrichen von Befund 7 des Zweitblicks über Phase 4.7.** Der alte Test
// schrieb das Verhalten sogar fest; er belegte, dass es gewollt war, nicht,
// dass es richtig ist.
//
// **Was es kostet, gehört gesagt**: wer ein Ergebnis nachtragen will und
// solange weiterspielen, kann das jetzt nicht mehr. Der Preis ist eine
// Rangliste, die stimmt.
func TestEineAeltereOffeneRundeHaeltAn(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.anmelden(t, turnier.ID, "Ada", "Bert")

	erste, err := werk.leitung.RundeLegen(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("die erste Runde: %v", err)
	}

	if setzen := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
		werk.fassung(t, turnier.ID), erste.Paarungen[0].ID, &domain.Ergebnis{SpieleA: 2}); setzen != nil {
		t.Fatalf("Ergebnis: %v", setzen)
	}

	zweite, err := werk.leitung.RundeLegen(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("die zweite Runde: %v", err)
	}

	// **Die erste wieder öffnen, die zweite abschließen.**
	if ruecknahme := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
		werk.fassung(t, turnier.ID), erste.Paarungen[0].ID, nil); ruecknahme != nil {
		t.Fatalf("zurücknehmen: %v", ruecknahme)
	}

	if zweites := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
		werk.fassung(t, turnier.ID), zweite.Paarungen[0].ID, &domain.Ergebnis{SpieleA: 2}); zweites != nil {
		t.Fatalf("Ergebnis: %v", zweites)
	}

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); !errors.Is(
		err, domain.ErrValidation,
	) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung — die erste Runde "+
			"hat ein offenes Ergebnis", err)
	}
}

// **Und ein Teilnehmer dieses Turniers steigt aus.**
//
// **Der Erfolgsweg neben der Abweisung**: ohne ihn bliebe die Zeile, die den
// Speicher wirklich ruft, unbetreten — und die Grenze aus Befund 6 sähe aus
// wie eine, die alles abweist.
func TestEinEigenerTeilnehmerSteigtAus(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	wer := werk.anmelden(t, turnier.ID, "Ada")[0]

	if err := werk.leitung.Aussteigen(
		t.Context(), turnier.ID, wer, true); err != nil {
		t.Fatalf("aussteigen: %v", err)
	}

	gelesen, lesefehler := werk.leitung.Turnier(t.Context(), turnier.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !gelesen.Teilnehmer[0].Ausgestiegen {
		t.Error("er ist nicht ausgestiegen")
	}
}

// **Ein Teilnehmer eines anderen Turniers lässt sich nicht abmelden.**
//
// # Der Befund
//
// Die REST-Route lautet `/turniere/{turnierID}/teilnehmer/{teilnehmerID}`,
// und die Kennung des Turniers wurde nicht ausgewertet: der Dienst änderte
// den Teilnehmer allein über seine globale Kennung. Eine Anfrage unter
// Turnier A konnte jemanden aus Turnier B abmelden.
//
// **Beim Ergebnis war die Grenze da** — dort prüft der Dienst, ob die Paarung
// zu diesem Turnier gehört. Zwei benachbarte Wege, eine Zusage. Befund 6 des
// Zweitblicks über Phase 4.7.
func TestEinFremderTeilnehmerLaesstSichNichtAbmelden(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)

	erstes := werk.turnierAnlegen(t)

	zweites, err := werk.leitung.Anlegen(t.Context(),
		domain.Turnier{Name: "Anderswo", Spiel: "magic"})
	if err != nil {
		t.Fatalf("das zweite Turnier: %v", err)
	}

	fremder := werk.anmelden(t, zweites.ID, "Bert")[0]

	if err := werk.leitung.Aussteigen(
		t.Context(), erstes.ID, fremder, true); !errors.Is(
		err, domain.ErrNotFound,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	// **Und er ist nicht abgemeldet.**
	gelesen, lesefehler := werk.leitung.Turnier(t.Context(), zweites.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen.Teilnehmer[0].Ausgestiegen {
		t.Error("er wurde trotzdem abgemeldet")
	}
}
