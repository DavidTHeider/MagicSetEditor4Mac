package main

// **Der Dienst auf mehreren Adressen** — G08
// (#2272).
//
// `--adresse` lässt sich wiederholen, und ohne sie gilt die Einstellung
// `netz.lauschen` — so startet auch der LaunchAgent, der keine Adresse
// mitgibt, dort, wo gewählt wurde. Jede angesagte Adresse bedient eine
// echte Anfrage mit der Marke.

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// alleAngesagten wartet auf die Marke und gibt jede angesagte Adresse.
func alleAngesagten(t *testing.T, angesagt *sofort) (adressen []string, marke string) {
	t.Helper()

	frist := time.After(60 * time.Second)

	for {
		select {
		case <-frist:
			t.Fatalf("der Dienst hat nichts angesagt: %q", angesagt.gelesen())
		case <-time.After(20 * time.Millisecond):
		}

		adressen, marke = nil, ""

		for _, zeile := range strings.Split(strings.TrimSpace(angesagt.gelesen()), "\n") {
			if wert, ja := strings.CutPrefix(zeile, "Dienst auf http://"); ja {
				adressen = append(adressen, wert)
			}

			if wert, ja := strings.CutPrefix(zeile, "Marke "); ja {
				marke = wert
			}
		}

		if marke != "" {
			return adressen, marke
		}
	}
}

func jedeBedient(t *testing.T, ctx context.Context, adressen []string, marke string) {
	t.Helper()

	for _, adresse := range adressen {
		anfrage, fehler := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+adresse+"/sets", http.NoBody)
		if fehler != nil {
			t.Fatalf("Anfrage bauen: %v", fehler)
		}

		anfrage.Header.Set("Authorization", "Bearer "+marke)

		antwort, fehler := http.DefaultClient.Do(anfrage)
		if fehler != nil {
			t.Fatalf("über %s: %v", adresse, fehler)
		}

		_ = antwort.Body.Close() //nolint:errcheck // gelesen wird nur der Status

		if antwort.StatusCode != http.StatusOK {
			t.Errorf("über %s: %d", adresse, antwort.StatusCode)
		}
	}
}

func lauscheMit(t *testing.T, argumente []string, lauschwahl string) []string {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	if lauschwahl != "" {
		db, fehler := storage.Oeffnen(probekontext(t), filepath.Join(zuhause, "bestand.db"))
		if fehler != nil {
			t.Fatal(fehler)
		}

		if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
			t.Fatal(fehler)
		}

		if antwort := app.DienstflaecheUeber(db).EinstellungSetzen(probekontext(t), domain.NetzLauschen, lauschwahl); antwort.Fehler != nil {
			t.Fatalf("EinstellungSetzen: %+v", antwort.Fehler)
		}

		schliesse(t, db)
	}

	ctx, beenden := context.WithCancel(probekontext(t))
	angesagt := &sofort{ziel: &bytes.Buffer{}}

	fertig := make(chan error, 1)
	go func() { fertig <- lauf(ctx, argumente, angesagt) }()

	adressen, marke := alleAngesagten(t, angesagt)
	jedeBedient(t, ctx, adressen, marke)

	beenden()

	if fehler := <-fertig; fehler != nil {
		t.Errorf("lauf: %v", fehler)
	}

	return adressen
}

func TestDerDienstLauschtAufJederGenanntenAdresse(t *testing.T) {
	adressen := lauscheMit(t, []string{"-port", "0", "-adresse", "127.0.0.1", "-adresse", "::1"}, "")

	if len(adressen) != 2 || !strings.HasPrefix(adressen[0], "127.0.0.1:") || !strings.HasPrefix(adressen[1], "[::1]:") {
		t.Errorf("angesagt sind %v", adressen)
	}
}

func TestOhneAdresseGiltDieEinstellung(t *testing.T) {
	adressen := lauscheMit(t, []string{"-port", "0"}, "::1,127.0.0.1")

	if len(adressen) != 2 || !strings.HasPrefix(adressen[0], "[::1]:") || !strings.HasPrefix(adressen[1], "127.0.0.1:") {
		t.Errorf("angesagt sind %v", adressen)
	}
}

func TestOhneAdresseUndOhneEinstellungGiltDieSchleife(t *testing.T) {
	adressen := lauscheMit(t, []string{"-port", "0"}, "")

	if len(adressen) != 1 || !strings.HasPrefix(adressen[0], "127.0.0.1:") {
		t.Errorf("angesagt sind %v", adressen)
	}
}
