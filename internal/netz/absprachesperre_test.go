// **Die Sperre in der Absprache** — was ein 429 über Prozessgrenzen hinweg
// bewirkt (#1163,
// Befund Z7-03).
//
// Aus `prozessgrenze_test.go` herausgelöst, als es über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort steht, wie ein **Abstand** gehalten wird,
// hier, wie eine **Sperre** wirkt.
package netz_test

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Eine Sperre aus einem 429 gilt für alle Prozesse.**
//
// # Der Befund des Zweitblicks
//
// Sie blieb im empfangenden Prozess. Ein zweiter sendete während `Retry-After`
// weiter — und ignorierte das 429 damit **gemeinsam**, auch wenn der erste es
// einhielt. Scryfall schreibt: „It is not acceptable to ignore HTTP 429
// responses", und wiederholte 429 können zur dauerhaften Sperre führen.
func TestEineSperreGiltFuerAlleProzesse(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	// **Der eine trägt sie ein** …
	netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt }, nil).
		Sperren(jetzt.Add(time.Hour))

	// … **und der andere findet sie.** Zwei Grenzen auf derselben Datei
	// stehen hier für zwei Prozesse: was sie teilen, ist die Nachbardatei.
	var gewartet time.Duration

	zweite := netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) {
			gewartet = dauer
			jetzt = jetzt.Add(dauer)
		})

	if fehler := zweite.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet != time.Hour {
		t.Errorf("gewartet wurde %v — die Sperre des anderen Prozesses griff nicht", gewartet)
	}
}

// **Die längere Sperre gilt.**
//
// Zwei 429 kurz hintereinander sagen nicht, dass die zweite die erste ablöst —
// sie sagen, dass die Gegenstelle **mindestens** so lange nichts will.
func TestDieLaengereSperreGiltInDerAbsprache(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	grenze := netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt }, nil)

	grenze.Sperren(jetzt.Add(time.Hour))
	grenze.Sperren(jetzt.Add(time.Minute))

	var gewartet time.Duration

	zweite := netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) {
			gewartet = dauer
			jetzt = jetzt.Add(dauer)
		})

	if fehler := zweite.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet != time.Hour {
		t.Errorf("gewartet wurde %v — die kürzere Sperre hat die längere abgelöst", gewartet)
	}
}

// **Ohne Absprache geht auch das Sperren ins Leere — ohne Fehler.**
func TestSperrenOhneAbspracheTutNichts(t *testing.T) {
	t.Parallel()

	var keine *netz.Prozessgrenze

	keine.Sperren(time.Now())

	netz.NeueProzessgrenze("", time.Hour, time.Now, nil).Sperren(time.Now())

	// **Und eine Nachbardatei, die keine sein kann**, ebenso.
	ort := filepath.Join(t.TempDir(), "budget.drossel")
	if err := os.Mkdir(ort, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	netz.NeueProzessgrenze(ort, time.Hour, time.Now, nil).Sperren(time.Now())
}

// **Und ein Werk trägt die Sperre in jede Klasse ein.**
//
// Sie gilt der **Gegenstelle**, nicht einer Endpunktklasse — dieselbe
// Überlegung wie bei [netz.Drosselwerk.SperreAlle] im eigenen Prozess.
func TestEinWerkTraegtDieSperreInJedeKlasse(t *testing.T) {
	ort := t.TempDir()

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	werk := netz.NeuesDrosselwerk(func() time.Time { return jetzt }, func(time.Duration) {}).
		MitAbsprache(func() string { return ort })

	werk.Fuer("allgemein", time.Millisecond)
	werk.Fuer("suche", 500*time.Millisecond)

	werk.SperreAlle(jetzt.Add(time.Hour))

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 2 {
		t.Fatalf("es liegen %d Nachbardateien statt zwei", len(eintraege))
	}

	// **Beide tragen die Sperre.** Ein Werk, das nur die eine einträgt, ließe
	// die andere Klasse weiterlaufen — und ignorierte das 429 halb.
	for _, eintrag := range eintraege {
		var gewartet time.Duration

		if fehler := netz.NeueProzessgrenze(filepath.Join(ort, eintrag.Name()),
			time.Millisecond, func() time.Time { return jetzt },
			func(dauer time.Duration) { gewartet = dauer }).
			Warte(t.Context()); fehler != nil {
			t.Fatalf("warten: %v", fehler)
		}

		if gewartet != time.Hour {
			t.Errorf("%s wartete %v — die Sperre fehlt dort", eintrag.Name(), gewartet)
		}
	}
}

// **Ohne Ladeabsprache holt jeder selbst.**
//
// Sie ist eine Verbesserung und keine Bedingung: ein doppelter Ladevorgang ist
// besser als kein Ergebnis (§8).
func TestOhneLadeabspracheHoltJederSelbst(t *testing.T) {
	t.Parallel()

	// **Ein Verzeichnis am Ort der Nachbardatei**: `dateisperre.Oeffnen`
	// scheitert.
	ort := t.TempDir()
	if err := os.Mkdir(filepath.Join(ort, netz.Ladedatei), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt, _, fehler := netz.NeuerSpeicher(ort, time.Hour, time.Now).
		Hole(t.Context(), "karte/x", func() ([]byte, error) { return []byte("gut"), nil })
	if fehler != nil {
		t.Fatalf("holen ohne Ladeabsprache: %v", fehler)
	}

	if string(inhalt) != "gut" {
		t.Errorf("der Inhalt ist %q", inhalt)
	}
}

// **Und wenn ein anderer sie zu lange hält, ebenso.**
func TestNachDerLadefristHoltJederSelbst(t *testing.T) {
	ort := t.TempDir()

	griff, err := dateisperre.Oeffnen(filepath.Join(ort, netz.Ladedatei))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("allein nehmen: %v", fehler)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	alteFrist, alterTakt := netz.Ladefrist, netz.Ladetakt
	netz.Ladefrist, netz.Ladetakt = 40*time.Millisecond, 10*time.Millisecond

	t.Cleanup(func() { netz.Ladefrist, netz.Ladetakt = alteFrist, alterTakt })

	var geholt bool

	if _, _, fehler := netz.NeuerSpeicher(ort, time.Hour, time.Now).
		Hole(t.Context(), "karte/x", func() ([]byte, error) {
			geholt = true

			return []byte("gut"), nil
		}); fehler != nil {
		t.Fatalf("holen nach der Frist: %v", fehler)
	}

	if !geholt {
		t.Error("nach der Frist wurde nicht selbst geholt")
	}
}

// **Ein beendeter Kontext bricht auch das Warten auf eine Sperre ab.**
//
// Sie kommt aus einem 429 und kann eine Stunde dauern; wer die Antwort nicht
// mehr will, soll nicht darauf warten.
func TestEinBeendeterKontextBrichtDasWartenAufDieSperreAb(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "budget.drossel")
	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt }, nil).
		Sperren(jetzt.Add(time.Hour))

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	// **Ohne eingesetzten Schlaf**: geprüft wird die echte Uhr gegen den
	// Kontext.
	fehler := netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt }, nil).Warte(abgebrochen)
	if !errors.Is(fehler, context.Canceled) {
		t.Errorf("der Fehler ist %v — will context.Canceled", fehler)
	}
}

// **Ein Sendezeitpunkt aus der Zukunft hält niemanden auf.**
//
// Er entsteht bei einer verstellten Uhr oder einer veränderten Datei. Ohne
// eine Wache wartete der nächste Aufruf so lange, wie die Zukunft entfernt ist.
func TestEinSendezeitpunktAusDerZukunftHaeltNiemandenAuf(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "budget.drossel")
	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	// **Ein Sendezeitpunkt in einem Jahr**, ohne Sperre daneben.
	roh := make([]byte, 16)
	binary.BigEndian.PutUint64(roh,
		uint64(jetzt.Add(9000*time.Hour).UnixNano())) // #nosec G115 -- Zeit nach 1970

	if err := os.WriteFile(pfad, roh, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var gewartet time.Duration

	if fehler := netz.NeueProzessgrenze(pfad, 100*time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) { gewartet = dauer }).
		Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet > 100*time.Millisecond {
		t.Errorf("gewartet wurde %v — mehr als der Abstand", gewartet)
	}
}
