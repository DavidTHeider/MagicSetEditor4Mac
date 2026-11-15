// **Was beim Fortschreiben der Listenwahl geprüft wird** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Eine zweite Datei**, weil `kartenliste_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt einer Frage: dort steht, **was die Liste
// zeigt**, hier, **was beim Ändern der Wahl gilt**.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Wahl wird fortgeschrieben — Spalten und Ordnung.**
func TestDieWahlWirdFortgeschrieben(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	bau := kartenliste(doppel)

	wahl := []domain.Spaltenwahl{{Feld: "typ", Sichtbar: true, Stelle: 10}}
	if err := bau.SpaltenSetzen(t.Context(), 1, wahl); err != nil {
		t.Fatalf("Spalten: %v", err)
	}

	if len(doppel.geschrieben) != 1 || doppel.geschrieben[0].Feld != "typ" {
		t.Errorf("geschrieben wurde %+v", doppel.geschrieben)
	}

	ordnung := domain.Kartenordnung{Feld: "name", Aufsteigend: false}
	if err := bau.SortierungSetzen(t.Context(), 1, ordnung); err != nil {
		t.Fatalf("Sortierung: %v", err)
	}

	if doppel.fortgeschrieben == nil || doppel.fortgeschrieben.Feld != "name" {
		t.Errorf("fortgeschrieben wurde %+v", doppel.fortgeschrieben)
	}
}

// **Eine Wahl über ein Feld, das es nicht gibt, wird abgewiesen.**
//
// §8 und §16, Regel 6: nichts wird still übernommen. Ein Tippfehler in einem
// Feldnamen legte sonst eine Spalte an, die niemals erscheint — und niemand
// sähe, warum.
func TestEineWahlUeberEinUnbekanntesFeldWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()

	fehler := kartenliste(doppel).SpaltenSetzen(t.Context(), 1,
		[]domain.Spaltenwahl{{Feld: "gibtsnicht", Sichtbar: true}})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler: %v", fehler)
	}

	if doppel.geschrieben != nil {
		t.Errorf("geschrieben wurde trotzdem: %+v", doppel.geschrieben)
	}

	// **Ebenso eine Sortierung nach einem unbekannten Feld.**
	if err := kartenliste(doppel).SortierungSetzen(t.Context(), 1,
		domain.Kartenordnung{Feld: "gibtsnicht"}); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Sortierung: %v", err)
	}

	// **Aber „nach nichts sortieren" ist erlaubt.**
	if err := kartenliste(doppel).SortierungSetzen(t.Context(), 1,
		domain.Kartenordnung{Feld: "", Aufsteigend: true}); err != nil {
		t.Errorf("die leere Sortierung: %v", err)
	}
}

// **Jeder Lesefehler kommt durch.**
func TestJederLesefehlerDerKartenlisteKommtDurch(t *testing.T) {
	t.Parallel()

	for name, verletze := range map[string]func(*listendoppel){
		"Set unlesbar":  func(d *listendoppel) { d.lesefehler = domain.ErrBelegt },
		"Wahl unlesbar": func(d *listendoppel) { d.wahlfehler = domain.ErrBelegt },

		// **Und die Sortierung getrennt von der Spaltenwahl** — sie sind zwei
		// Abfragen, und die zweite käme sonst nie an die Reihe.
		"Sortierung unlesbar": func(d *listendoppel) { d.ordnungfehler = domain.ErrBelegt },

		// **Und die Karten selbst** — sie kommen zuletzt, und deshalb
		// braucht dieser Fall ein eigenes Feld: ein Lesefehler am Setkopf
		// hielte den Lauf schon davor an.
		"Karten unlesbar": func(d *listendoppel) { d.kartenfehler = domain.ErrBelegt },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := dreiFelder()
			verletze(doppel)

			if _, fehler := kartenliste(doppel).Aufbau(
				t.Context(), 1); !errors.Is(fehler, domain.ErrBelegt) {
				t.Errorf("Aufbau: %v", fehler)
			}
		})
	}
}

// **Und auch die beiden Schreibwege lesen zuerst.**
//
// Sie fragen nach dem **Spiel**, bevor sie schreiben — die Wahl gehört ihm.
// Ein unlesbarer Setkopf hält sie an; unlesbare **Felder** nicht, denn ohne
// Vorlage wird nicht geprüft (der Lauf darunter).
func TestDieSchreibwegeDerKartenlisteLesenZuerst(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.lesefehler = domain.ErrBelegt

	if fehler := kartenliste(doppel).SpaltenSetzen(
		t.Context(), 1, nil); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("Spalten: %v", fehler)
	}

	if fehler := kartenliste(doppel).SortierungSetzen(
		t.Context(), 1, domain.Kartenordnung{}); !errors.Is(
		fehler, domain.ErrBelegt) {
		t.Errorf("Sortierung: %v", fehler)
	}
}

// **Ein Feld ohne eigene Stelle steht hinten, nicht vorn.**
//
// Die Null in `card list column:` heißt „ungesagt", nicht „ganz vorne" — wer
// sie als Stelle läse, schöbe jedes namenlose Feld an den Anfang der Liste.
func TestEinFeldOhneStelleStehtHinten(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.wahl = []domain.Spaltenwahl{{Feld: "typ", Sichtbar: true, Stelle: 10}}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	letzte := liste.Spalten[len(liste.Spalten)-1]
	if letzte.Feld != "ohne stelle" {
		t.Errorf("hinten steht %q", letzte.Feld)
	}
}

// **Zwei Felder auf derselben Stelle ordnen nach ihrem Namen.**
//
// Sonst hinge die Reihenfolge an der Laune der Sortierung — und wechselte
// zwischen zwei Läufen.
func TestZweiFelderAufDerselbenStelleOrdnenNachNamen(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.felder = []domain.Field{
		listenfeld("zebra", "", 100, false, true, domain.ListeLinks),
		listenfeld("adler", "", 100, false, true, domain.ListeLinks),
	}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.Spalten[0].Feld != "adler" || liste.Spalten[1].Feld != "zebra" {
		t.Errorf("%+v", liste.Spalten)
	}
}

// **Und jeder Schreibfehler ebenso.**
func TestJederSchreibfehlerDerKartenlisteKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.schreibfehler = domain.ErrBelegt

	if err := kartenliste(doppel).SpaltenSetzen(t.Context(), 1, nil); !errors.Is(
		err, domain.ErrBelegt) {
		t.Errorf("Spalten: %v", err)
	}

	if err := kartenliste(doppel).SortierungSetzen(t.Context(), 1,
		domain.Kartenordnung{}); !errors.Is(err, domain.ErrBelegt) {
		t.Errorf("Sortierung: %v", err)
	}
}

// **Ein Set, das es nicht gibt, wird gemeldet.**
func TestEinUnbekanntesSetHatKeineKartenliste(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.spiel = ""

	if _, fehler := kartenliste(doppel).Aufbau(t.Context(), 1); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Aufbau: %v", fehler)
	}
}

// **Ohne lesbare Vorlage kommen die Spalten aus den Karten.**
//
// Dieselbe Duldung wie beim Ersetzen und beim Feldschreiben: ein Set aus MTG
// JSON bliebe sonst ohne Kartenliste, solange sein Spielpaket fehlt — und die
// Liste ist der Weg zu jeder Karte.
//
// **Was das Paket sagen würde, weiß dann niemand**: keine Aufschrift, keine
// Reihenfolge, kein Verbot. Was bleibt, sind die Felder, die die Karten
// führen — in der Reihenfolge, in der sie stehen.
func TestOhneVorlageKommenDieSpaltenAusDenKarten(t *testing.T) {
	t.Parallel()

	for name, verletze := range map[string]func(*listendoppel){
		"keine Felder":    func(d *listendoppel) { d.felder = nil },
		"Felder unlesbar": func(d *listendoppel) { d.feldfehler = domain.ErrNotFound },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := dreiFelder()
			verletze(doppel)

			liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
			if fehler != nil {
				t.Fatalf("Aufbau: %v", fehler)
			}

			// `name` und `typ` — die beiden Felder, die die Karten führen.
			if len(liste.Spalten) != 2 ||
				liste.Spalten[0].Feld != "name" || liste.Spalten[1].Feld != "typ" {
				t.Fatalf("%+v", liste.Spalten)
			}

			// **Die erste ist sichtbar, die übrigen nicht.** Ohne Vorschlag
			// des Pakets ist das die einzige Wahl, die eine Liste brauchbar
			// lässt — der Name steht vorn.
			if !liste.Spalten[0].Sichtbar || liste.Spalten[1].Sichtbar {
				t.Errorf("Sichtbarkeit: %+v", liste.Spalten)
			}

			// **Und die Antwort sagt, dass geraten wurde.**
			if !liste.OhneVorlage {
				t.Error("die Antwort verschweigt die fehlende Vorlage")
			}
		})
	}
}

// **Mit Vorlage steht der Hinweis nicht da.**
func TestMitVorlageStehtBeiDerListeKeinHinweis(t *testing.T) {
	t.Parallel()

	liste, fehler := kartenliste(dreiFelder()).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.OhneVorlage {
		t.Error("der Hinweis steht auch mit Vorlage da")
	}
}

// **Und eine Wahl ohne Vorlage wird nicht gegen sie geprüft.**
//
// Sonst ließe sich in einem Set ohne Spielpaket keine Spalte wählen — und die
// Duldung darüber wäre nur die halbe.
func TestEineWahlOhneVorlageWirdNichtGeprueft(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.feldfehler = domain.ErrNotFound

	if err := kartenliste(doppel).SpaltenSetzen(t.Context(), 1,
		[]domain.Spaltenwahl{{Feld: "name", Sichtbar: true}}); err != nil {
		t.Errorf("Spalten: %v", err)
	}

	if err := kartenliste(doppel).SortierungSetzen(t.Context(), 1,
		domain.Kartenordnung{Feld: "name"}); err != nil {
		t.Errorf("Sortierung: %v", err)
	}
}
