// **Wer die Sperre einträgt, und was ihn dabei aufhält** — die Wege von
// `Sperren` (#1163,
// Befund Z7-03).
//
// Aus `absprachesperre_test.go` herausgelöst, als es über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort steht, was eine eingetragene Sperre
// **bewirkt**, hier, wie sie **hineinkommt**.
package netz_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Eine belegte Absprache wird abgewartet, nicht aufgegeben.**
//
// # Der Befund des Nachlaufs
//
// `Sperren` versuchte die Nachbardatei **einmal** ohne Warten. War sie gerade
// belegt, wurde der 429-Zeitpunkt **gar nicht** eingetragen — und die anderen
// Prozesse sandten weiter. Hier stand ein Lauf, der genau das zusicherte:
// „nimmt die Sperre nicht an". Er prüfte das Gegenstück zur Zusage.
func TestEineBelegteAbspracheWirdAbgewartet(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")
	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("allein nehmen: %v", fehler)
	}

	// **Losgelassen wird, während `Sperren` wartet.** Ohne das prüfte der Lauf
	// die Frist und nicht das Warten.
	losgelassen := make(chan struct{})

	go func() {
		time.Sleep(50 * time.Millisecond)

		_ = griff.Schliessen() //nolint:errcheck // die Sperre fällt mit dem Griff

		close(losgelassen)
	}()

	netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt }, nil).
		Sperren(jetzt.Add(time.Hour))

	<-losgelassen

	// **Und die Sperre steht wirklich da.**
	var gewartet time.Duration

	if fehler := netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) {
			gewartet = dauer
			jetzt = jetzt.Add(dauer)
		}).Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet != time.Hour {
		t.Errorf("gewartet wurde %v — die Sperre wurde nicht eingetragen", gewartet)
	}
}

// **Und wenn sie auch nach der Frist belegt ist, bleibt die Sperre im eigenen
// Prozess.**
//
// Sie dort zu halten ist besser als sie zu verlieren; dafür sorgt
// `Drossel.Sperre`.
func TestEineDauerhaftBelegteAbspracheGibtAuf(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("allein nehmen: %v", fehler)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	alteFrist, alterTakt := netz.Prozessgrenzfrist, netz.Prozessgrenztakt
	netz.Prozessgrenzfrist, netz.Prozessgrenztakt = 40*time.Millisecond, 10*time.Millisecond

	t.Cleanup(func() { netz.Prozessgrenzfrist, netz.Prozessgrenztakt = alteFrist, alterTakt })

	beginn := time.Now()

	netz.NeueProzessgrenze(pfad, time.Millisecond, time.Now, nil).
		Sperren(time.Now().Add(time.Hour))

	if gedauert := time.Since(beginn); gedauert < 40*time.Millisecond {
		t.Errorf("es dauerte nur %v — es wurde gar nicht gewartet", gedauert)
	}
}

// **Eine lange Sperre hält die Absprache nicht besetzt.**
//
// # Der Befund des Nachlaufs
//
// `Warte` wartete die ganze Sperre **mit** dem Griff in der Hand. Dauert ein
// `Retry-After` länger als [netz.Prozessgrenzfrist], geben die anderen
// Prozesse auf und senden **ohne** Absprache — die gemeinsame Grenze fällt
// genau dann weg, wenn sie am meisten zählt.
//
// **Gemessen wird, ob ein zweiter Griff drankommt**, während der erste eine
// Stunde abwartet.
func TestEineLangeSperreHaeltDieAbspracheNichtBesetzt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")
	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt }, nil).
		Sperren(jetzt.Add(time.Hour))

	// **Mitten im Warten wird nachgesehen.** Der eingesetzte Schlaf ist die
	// Stelle, an der ein echter Prozess schliefe — dort muss die Nachbardatei
	// frei sein.
	var freiWaehrendDesWartens bool

	if fehler := netz.NeueProzessgrenze(pfad, time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) {
			if dauer >= time.Hour {
				freiWaehrendDesWartens = frei(t, pfad)
			}

			jetzt = jetzt.Add(dauer)
		}).Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if !freiWaehrendDesWartens {
		t.Error("die Absprache war während der Sperre besetzt — andere Prozesse " +
			"geben dann nach der Frist auf und senden ohne sie")
	}
}

// frei sagt, ob die Nachbardatei gerade zu haben ist.
func frei(t *testing.T, pfad string) bool {
	t.Helper()

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = griff.Schliessen() }() //nolint:errcheck // nur gefragt

	return griff.Allein() == nil
}
