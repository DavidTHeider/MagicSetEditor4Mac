package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Composition Root wird gefahren, nicht behauptet**
// (#997).
//
// # Warum das ein eigener Test ist
//
// `internal/rest` prüft den Weg **bis** zum Anwendungsfall über Doubles. Was
// hier geprüft wird, ist die Verdrahtung dahinter: dass der Bestand aufgeht,
// migriert wird, die Nähte zusammenkommen und wirklich etwas herauskommt.
// Genau dort entstand in Phase 5 der Fehler, den kein Test sah (#130).
func TestDerDienstBedientEineEchteAnfrage(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	ctx, beenden := context.WithCancel(probekontext(t))

	angesagt := &sofort{ziel: &bytes.Buffer{}}

	fertig := make(chan error, 1)
	go func() { fertig <- lauf(ctx, []string{"-port", "0"}, angesagt) }()

	adresse, marke := warteAufAnsage(t, angesagt)

	// **Eine echte Anfrage über eine echte Verbindung.**
	anfrage, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://"+adresse+"/sets", http.NoBody)
	if err != nil {
		t.Fatalf("Anfrage bauen: %v", err)
	}

	anfrage.Header.Set("Authorization", "Bearer "+marke)

	antwort, fehler := http.DefaultClient.Do(anfrage)
	if fehler != nil {
		t.Fatalf("fragen: %v", fehler)
	}

	defer schliesse(t, antwort.Body)

	if antwort.StatusCode != http.StatusOK {
		t.Errorf("Status %d", antwort.StatusCode)
	}

	var rumpf struct {
		Sets []map[string]string `json:"sets"`
	}

	if err := json.NewDecoder(antwort.Body).Decode(&rumpf); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v", err)
	}

	// **Ein frischer Bestand hat kein Set — und `[]`, nicht `null`.**
	if rumpf.Sets == nil {
		t.Error("die leere Liste kam als null an")
	}

	beenden()

	if fehler := <-fertig; fehler != nil {
		t.Errorf("der Dienst endete mit einem Fehler: %v", fehler)
	}
}

// **Ohne Marke kommt auch über eine echte Verbindung niemand durch.**
//
// Die Gegenprobe zum Fall darüber: ohne sie bliebe offen, ob die Wachen im
// gebauten Programm überhaupt hängen.
func TestUeberDieEchteVerbindungGiltDieMarkeAuch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	ctx, beenden := context.WithCancel(probekontext(t))

	angesagt := &sofort{ziel: &bytes.Buffer{}}

	fertig := make(chan error, 1)
	go func() { fertig <- lauf(ctx, nil, angesagt) }()

	adresse, _ := warteAufAnsage(t, angesagt)

	anfrage, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://"+adresse+"/sets", http.NoBody)
	if err != nil {
		t.Fatalf("Anfrage bauen: %v", err)
	}

	antwort, fehler := http.DefaultClient.Do(anfrage)
	if fehler != nil {
		t.Fatalf("fragen: %v", fehler)
	}

	defer schliesse(t, antwort.Body)

	if antwort.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status %d, erwartet war %d", antwort.StatusCode, http.StatusUnauthorized)
	}

	beenden()

	if fehler := <-fertig; fehler != nil {
		t.Errorf("der Dienst endete mit %v", fehler)
	}
}

// **Ein unbrauchbarer Port wird abgewiesen, bevor irgendetwas aufgeht.**
func TestEinUnbrauchbarerPortWirdAbgewiesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	for _, port := range []string{"-1", "70000"} {
		t.Run(port, func(t *testing.T) {
			fehler := lauf(probekontext(t), []string{"-port", port}, &bytes.Buffer{})
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v ist keine Validierung", fehler)
			}
		})
	}
}

// **Und ein Argument, das es nicht gibt, ebenso.**
func TestEinUnbekanntesArgumentWirdAbgewiesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if fehler := lauf(probekontext(t), []string{"-gibtsnicht"}, &bytes.Buffer{}); fehler == nil {
		t.Error("das unbekannte Argument kam durch")
	}
}

// **Ohne Zuhause gibt es keinen Bestandsort.**
func TestOhneZuhauseKeinDienst(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if fehler := lauf(probekontext(t), nil, &bytes.Buffer{}); fehler == nil {
		t.Error("ohne Zuhause lief der Dienst an")
	}
}

// **Und wenn die Ansage nicht ankommt, läuft der Dienst nicht an.**
//
// Wer die Marke nicht bekommt, kann den Dienst nicht benutzen; ein Server, der
// dann auf eine Marke wartet, die niemand kennt, wäre schlimmer als ein
// Fehlschlag.
func TestOhneAnsageKeinDienst(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	fehler := lauf(probekontext(t), nil, stockend{})
	if fehler == nil {
		t.Fatal("ohne Ansage lief der Dienst an")
	}

	if !strings.Contains(fehler.Error(), "Marke ansagen") {
		t.Errorf("der Fehler nennt den Grund nicht: %v", fehler)
	}
}

// schliesse macht einen Rumpf zu und meldet, wenn es nicht ging.
func schliesse(t *testing.T, was io.Closer) {
	t.Helper()

	if err := was.Close(); err != nil {
		t.Errorf("schließen: %v", err)
	}
}

// warteAufAnsage liest Adresse und Marke aus der Ansage.
func warteAufAnsage(t *testing.T, angesagt *sofort) (adresse, marke string) {
	t.Helper()

	// **Sechzig Sekunden, nicht zehn** (2026-09-09): unter `-race` und neben
	// zwei rechnenden Blender-Prozessen brauchte der Dienst länger als zehn bis
	// zur Ansage, und zwei Pipelineläufe endeten hier — bei einem Dienst, der
	// allein in einer Sekunde ansagt. Die Frist ist ein Netz gegen ein Hängen,
	// keine Messung.
	frist := time.After(60 * time.Second)

	for {
		select {
		case <-frist:
			t.Fatalf("der Dienst hat nichts angesagt: %q", angesagt.gelesen())
		default:
		}

		// **Je Vorsatz und nicht je Zeilennummer**: seit N6 (#1572) steht
		// jenseits der Schleife eine dritte Zeile dazwischen. Hier bindet
		// zwar jeder Lauf auf der Schleife — aber ein Leser, der bei der
		// ersten TLS-Bindung still danebenläge, ist einer zu viel.
		adresse, marke = "", ""

		for _, zeile := range strings.Split(
			strings.TrimSpace(angesagt.gelesen()), "\n") {
			if wert, ja := strings.CutPrefix(
				zeile, "Dienst auf http://"); ja {
				adresse = wert
			}

			if wert, ja := strings.CutPrefix(zeile, "Marke "); ja {
				marke = wert
			}
		}

		if adresse != "" && marke != "" {
			return adresse, marke
		}
	}
}

// sofort schreibt durch und ist dabei gegen gleichzeitiges Lesen gesichert.
//
// **Die Sperre ist nötig**, weil der Dienst in einer Goroutine schreibt und der
// Test daneben liest — ohne sie meldet `-race` zu Recht einen Wettlauf.
type sofort struct {
	ziel     *bytes.Buffer
	schranke sync.Mutex
}

// gelesen gibt, was bisher angesagt wurde.
func (s *sofort) gelesen() string {
	s.schranke.Lock()
	defer s.schranke.Unlock()

	return s.ziel.String()
}

func (s *sofort) Write(roh []byte) (int, error) {
	s.schranke.Lock()
	defer s.schranke.Unlock()

	return s.ziel.Write(roh) //nolint:wrapcheck // ein Puffer scheitert nicht
}

// stockend nimmt nichts an — wie eine geschlossene Röhre.
type stockend struct{}

func (stockend) Write([]byte) (int, error) {
	return 0, os.ErrClosed
}
