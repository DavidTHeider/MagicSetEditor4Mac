package service_test

// **Die Sicherungsprobe: sichern und zurücklesen** — Werkstück O6 von Phase
// 5.2 (#1430).
//
// **Hier und nicht in `internal/app`**: dort liegt nur die Zeile, die die zwei
// Schichten verbindet — was eine Probe ist und wann sie gilt, entscheidet
// dieses Paket.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Eine Sicherung, die sich nicht zurücklesen lässt, ist keine** — und
// genau das sagt die Probe. Ohne diesen Lauf wäre der Zweig eine Behauptung
// über den Ernstfall (§3).
func TestEineUnlesbareSicherungFaelltBeimZurueclesenAuf(t *testing.T) {
	t.Parallel()

	probe := service.SicherungsprobeUeber(func(_ context.Context, ziel string) error {
		// **Etwas schreiben, das kein Archiv ist**: so verhält sich eine
		// Sicherung, die unterwegs abgebrochen wurde.
		return os.WriteFile(ziel, []byte("kein Archiv"), 0o600)
	})

	fehler := probe.SichereUndPruefe(t.Context())
	if fehler == nil {
		t.Fatal("die unlesbare Sicherung kam durch")
	}

	if !strings.Contains(fehler.Error(), "zurücklesen") {
		t.Errorf("der Satz sagt nicht, an welchem Schritt es lag: %v", fehler)
	}
}

// **Und ohne Platz für die Probe gibt es keine.** Der Zweig ist über die
// Umgebung erreichbar, und ein Zweig, den kein Lauf betritt, wäre eine Lücke
// mit einer Ausrede davor.
func TestOhneOrtFuerDieProbeGibtEsKeine(t *testing.T) {
	t.Setenv("TMPDIR", "/gibtesnicht/und/wird/es/nie")

	probe := service.SicherungsprobeUeber(func(context.Context, string) error { return nil })

	if fehler := probe.SichereUndPruefe(t.Context()); fehler == nil {
		t.Error("ohne Ort entstand eine Probe")
	}
}

// **Und wenn schon das Schreiben scheitert, ist die Probe dort zu Ende** —
// der Satz nennt den Schritt, denn „Platte voll" und „Archiv kaputt" führen
// zu verschiedenen Handgriffen.
func TestEineProbeDieSichNichtSchreibenLaesstNenntDenSchritt(t *testing.T) {
	t.Parallel()

	probe := service.SicherungsprobeUeber(func(context.Context, string) error {
		return errors.New("die Platte ist voll")
	})

	fehler := probe.SichereUndPruefe(t.Context())
	if fehler == nil {
		t.Fatal("die gescheiterte Sicherung kam durch")
	}

	if !strings.Contains(fehler.Error(), "schreiben") {
		t.Errorf("der Satz sagt nicht, an welchem Schritt es lag: %v", fehler)
	}
}

// **Eine echte Sicherung geht durch** — und ohne diesen Lauf wäre jeder Satz
// oben eine Aussage darüber, dass die Probe **immer** scheitert (§10).
//
// **Mit einem wirklichen Archiv und nicht mit einer Attrappe**: die Aussage
// dieser Naht ist das Zurücklesen, und ein Doppel, das `EntpackenNur`
// umginge, prüfte den Weg nicht, den es hier gibt.
func TestEineEchteSicherungBestehtDieProbe(t *testing.T) {
	t.Parallel()

	quelle := filepath.Join(t.TempDir(), "bestand.db")
	if fehler := os.WriteFile(quelle, []byte("kein echter Bestand, aber eine Datei"), 0o600); fehler != nil {
		t.Fatalf("die Quelle schreiben: %v", fehler)
	}

	probe := service.SicherungsprobeUeber(func(_ context.Context, ziel string) error {
		return sicherung.Packen(ziel,
			[]sicherung.Teil{{Name: "bestand.db", Quelle: quelle}},
			sicherung.Angaben{Fassung: "0.5.0", Erstellt: time.Now()})
	})

	if fehler := probe.SichereUndPruefe(t.Context()); fehler != nil {
		t.Errorf("die echte Sicherung fiel durch: %v", fehler)
	}
}
