// **Die Absprache für sich** — was sie in **einem** Prozess tut
// (#1163,
// Befund Z7-03).
//
// Der Mehrprozesslauf steht in `prozessdrossel_test.go`; er belegt die Zusage.
// Hier stehen die Wege daneben: die eingesetzte Uhr, der Kontext, die Fälle,
// in denen es **ohne** Absprache weitergeht.
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

// **Der erste Aufruf wartet nicht, der zweite schon.**
func TestDieAbspracheHaeltDenAbstand(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	var gewartet []time.Duration

	grenze := netz.NeueProzessgrenze(pfad, 100*time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) {
			gewartet = append(gewartet, dauer)
			jetzt = jetzt.Add(dauer)
		})

	for range 3 {
		if fehler := grenze.Warte(t.Context()); fehler != nil {
			t.Fatalf("warten: %v", fehler)
		}
	}

	// **Zweimal gewartet und nicht dreimal**: der erste Aufruf trifft auf eine
	// leere Datei, und die ist „lange her".
	if len(gewartet) != 2 {
		t.Fatalf("gewartet wurde %v", gewartet)
	}

	for _, dauer := range gewartet {
		if dauer != 100*time.Millisecond {
			t.Errorf("gewartet wurde %v statt des Abstands", dauer)
		}
	}
}

// **Ohne Uhr gilt die echte.**
func TestOhneUhrGiltDieEchte(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	grenze := netz.NeueProzessgrenze(pfad, time.Nanosecond, nil, func(time.Duration) {})

	if fehler := grenze.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}
}

// **Ein beendeter Kontext bricht das Warten ab.**
//
// **Ohne eingesetzten Schlaf**, denn genau der Weg wird geprüft: die echte Uhr
// gegen den Kontext.
func TestEinBeendeterKontextBrichtDasWartenAb(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	grenze := netz.NeueProzessgrenze(pfad, time.Hour, time.Now, nil)

	// **Der erste Aufruf setzt den Zeitpunkt**, der zweite muss warten.
	if fehler := grenze.Warte(t.Context()); fehler != nil {
		t.Fatalf("der erste Aufruf: %v", fehler)
	}

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if fehler := grenze.Warte(abgebrochen); !errors.Is(fehler, context.Canceled) {
		t.Errorf("der Fehler ist %v — will context.Canceled", fehler)
	}
}

// **Ohne Absprache geht es weiter.**
//
// Sie ist eine Verbesserung des Schutzes und keine Bedingung: der Aufrufer hat
// seine eigene Drossel. Eine Anwendung, die wegen einer Nachbardatei keine
// Karte mehr sucht, wäre schlechter als eine, die je Prozess drosselt.
func TestOhneAbspracheGehtEsWeiter(t *testing.T) {
	t.Run("ohne Pfad", func(t *testing.T) {
		if fehler := netz.NeueProzessgrenze("", time.Hour, time.Now, nil).
			Warte(t.Context()); fehler != nil {
			t.Errorf("ohne Pfad: %v", fehler)
		}
	})

	t.Run("ohne Nachbardatei", func(t *testing.T) {
		// **Ein Verzeichnis am Ort der Datei**: `os.OpenFile` scheitert.
		ort := filepath.Join(t.TempDir(), "budget.drossel")
		if err := os.Mkdir(ort, 0o700); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if fehler := netz.NeueProzessgrenze(ort, time.Hour, time.Now, nil).
			Warte(t.Context()); fehler != nil {
			t.Errorf("ohne Nachbardatei: %v", fehler)
		}
	})

	t.Run("ohne Grenze", func(t *testing.T) {
		var keine *netz.Prozessgrenze

		if fehler := keine.Warte(t.Context()); fehler != nil {
			t.Errorf("ohne Grenze: %v", fehler)
		}
	})
}

// **Eine unlesbare Absprache gilt als „lange her".**
//
// Sie entsteht beim ersten Mal leer; ein Fehler darauf hielte den ersten
// Aufruf jedes frischen Rechners auf.
func TestEineUnlesbareAbspracheGiltAlsLangeHer(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	if err := os.WriteFile(pfad, []byte("zu kurz"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var gewartet bool

	grenze := netz.NeueProzessgrenze(pfad, time.Hour, time.Now,
		func(time.Duration) { gewartet = true })

	if fehler := grenze.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet {
		t.Error("es wurde gewartet — eine unlesbare Angabe wurde für einen Zeitpunkt gehalten")
	}
}

// **Ein Werk mit Absprache gibt Drosseln, die sie benutzen.**
func TestEinWerkMitAbspracheReichtSieDurch(t *testing.T) {
	ort := t.TempDir()

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	werk := netz.NeuesDrosselwerk(
		func() time.Time { return jetzt },
		func(dauer time.Duration) { jetzt = jetzt.Add(dauer) }).
		MitAbsprache(func() string { return ort })

	if gesetzt := werk.Abspracheort(); gesetzt != ort {
		t.Fatalf("der Abspracheort ist %q statt %q", gesetzt, ort)
	}

	drossel := werk.Fuer("allgemein", 100*time.Millisecond)

	for range 2 {
		if fehler := drossel.Warte(t.Context()); fehler != nil {
			t.Fatalf("warten: %v", fehler)
		}
	}

	// **Die Nachbardatei liegt da** — und ihr Name ist gehasht, nicht der
	// Schlüssel: ein `/` darin schriebe sie woanders hin.
	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 1 {
		t.Fatalf("im Verzeichnis liegen %d Einträge: %v", len(eintraege), eintraege)
	}

	if name := eintraege[0].Name(); name == "allgemein.drossel" {
		t.Errorf("die Nachbardatei heißt wie der Schlüssel: %q", name)
	}
}

// **Ein Zeitpunkt aus der Zukunft hält niemanden auf.**
//
// Er entsteht bei einer verstellten Uhr oder einer veränderten Datei. Ohne
// eine Wache wartete der nächste Aufruf so lange, wie die Zukunft entfernt ist
// — bei einem verstellten Jahr sind das Monate.
func TestEinZeitpunktAusDerZukunftHaeltNiemandenAuf(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	kuenftig := make([]byte, 8)
	binary.BigEndian.PutUint64(kuenftig,
		uint64(time.Now().Add(9000*time.Hour).UnixNano())) // #nosec G115 -- Zeit nach 1970

	if err := os.WriteFile(pfad, kuenftig, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var gewartet time.Duration

	grenze := netz.NeueProzessgrenze(pfad, 100*time.Millisecond, time.Now,
		func(dauer time.Duration) { gewartet = dauer })

	if fehler := grenze.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet > 100*time.Millisecond {
		t.Errorf("gewartet wurde %v — mehr als der Abstand", gewartet)
	}
}

// **Eine Absprache mit einer Null darin gilt als „lange her".**
//
// Sie entsteht, wenn die Nachbardatei angelegt, aber noch nicht beschrieben
// wurde — der Augenblick zwischen `O_CREATE` und dem ersten Vermerk.
func TestEineAbspracheMitNullGiltAlsLangeHer(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	if err := os.WriteFile(pfad, make([]byte, 8), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var gewartet bool

	grenze := netz.NeueProzessgrenze(pfad, time.Hour, time.Now,
		func(time.Duration) { gewartet = true })

	if fehler := grenze.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gewartet {
		t.Error("es wurde gewartet — eine Null wurde für einen Zeitpunkt gehalten")
	}
}

// **Hält ein anderer Prozess die Absprache zu lange, geht es ohne sie
// weiter.**
//
// Ein Anspruch, der hängenbleibt, hielte sonst jede Anfrage für immer auf.
// Ohne Absprache drosselt der Aufrufer je Prozess — schlechter, aber besser
// als eine Anwendung, die stehen bleibt.
func TestNachDerFristGehtEsOhneAbspracheWeiter(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	// **Ein Griff, der sie hält.** Er wird nicht geschlossen, solange dieser
	// Lauf dauert; `flock` gilt für die offene Datei, also steht er auch dem
	// eigenen Prozess im Weg.
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

	var gewartet bool

	grenze := netz.NeueProzessgrenze(pfad, time.Hour, time.Now,
		func(time.Duration) { gewartet = true })

	beginn := time.Now()

	if fehler := grenze.Warte(t.Context()); fehler != nil {
		t.Fatalf("warten: %v", fehler)
	}

	if gedauert := time.Since(beginn); gedauert < 40*time.Millisecond {
		t.Errorf("es dauerte nur %v — es wurde gar nicht gewartet", gedauert)
	}

	if gewartet {
		t.Error("es wurde ein Abstand gehalten, obwohl die Absprache nicht zu haben war")
	}
}

// **Ein beendeter Kontext bricht auch das Warten auf die Absprache ab.**
//
// # Der Befund des Zweitblicks
//
// Der Kontext wurde erst **nach** der Sperrnahme ausgewertet. Ein längst
// abgebrochener MCP- oder REST-Aufruf wartete deshalb bis zu fünf Sekunden auf
// eine Sperre, die niemand mehr brauchte — und verbrauchte danach womöglich
// noch einen Zeitstempel.
func TestEinBeendeterKontextBrichtDasWartenAufDieAbspracheAb(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	// **Ein Griff, der sie hält** — der eigene Prozess reicht: `flock` gilt
	// für die offene Datei.
	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("allein nehmen: %v", fehler)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	beginn := time.Now()

	fehler := netz.NeueProzessgrenze(pfad, time.Hour, time.Now, nil).Warte(abgebrochen)
	if !errors.Is(fehler, context.Canceled) {
		t.Errorf("der Fehler ist %v — will context.Canceled", fehler)
	}

	// **Und zwar sofort.** Ohne die Auswertung dauerte es die volle Frist.
	if gedauert := time.Since(beginn); gedauert > netz.Prozessgrenzfrist/2 {
		t.Errorf("es dauerte %v — der Kontext wurde erst danach gelesen", gedauert)
	}
}

// **Ein Werk ohne Absprache gibt keinen Ort und keine Grenze.**
//
// Der Normalfall in jedem Lauf, der das Drosselwerk einzeln baut: ohne
// `MitAbsprache` bleibt alles beim Stand vor Phase 3.2.
func TestEinWerkOhneAbspracheGibtKeinenOrt(t *testing.T) {
	t.Parallel()

	werk := netz.NeuesDrosselwerk(time.Now, func(time.Duration) {})

	if ort := werk.Abspracheort(); ort != "" {
		t.Errorf("ohne Absprache entstand der Ort %q", ort)
	}

	// **Und die Drossel läuft trotzdem.**
	if fehler := werk.Fuer("allgemein", time.Millisecond).Warte(t.Context()); fehler != nil {
		t.Errorf("warten ohne Absprache: %v", fehler)
	}
}

// **Und ein Werk, dessen Naht einen leeren Ort gibt, ebenso.**
//
// Der Fall entsteht ohne Benutzerverzeichnis: `app.ScryfallAbspracheort` gibt
// dann eine leere Zeichenkette, und die Drossel muss trotzdem laufen.
func TestEinWerkMitLeeremOrtGibtKeineGrenze(t *testing.T) {
	t.Parallel()

	werk := netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}).
		MitAbsprache(func() string { return "" })

	if fehler := werk.Fuer("allgemein", time.Millisecond).Warte(t.Context()); fehler != nil {
		t.Errorf("warten mit leerem Ort: %v", fehler)
	}
}
