package app

// **Ein Tausch des Bestands beendet die Sitzung nicht** — Befund 1 des
// Zweitblicks über Phase 1 von 0.9.0 (F04, #2343). Zurückspielen,
// Verschlüsseln, Neuanfang und ein abgebrochener Start räumen den Bestand
// über denselben Weg ab wie der Schluss; die Marke gehört aber dem Prozess
// und nicht dem Bestand.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// marken zählt die Marken im Ordner.
func marken(t *testing.T, ordner string) int {
	t.Helper()

	gefunden, fehler := filepath.Glob(filepath.Join(ordner, "fenster-*.json"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	return len(gefunden)
}

func TestEinTauschLaesstDieMarkeDerSitzungStehen(t *testing.T) {
	for name, tausch := range map[string]func(*App) error{
		"ein Tausch": func(a *App) error {
			return a.imTausch(func() error {
				if fehler := a.zumTauschSchliessen(); fehler != nil {
					return fehler
				}

				return a.wiederOeffnen(context.Background())
			})
		},
		"ein abgebrochener Start": func(a *App) error {
			fehler := a.abbrechen(errors.New("die Wanderung scheiterte"))
			if a.lauf.tor.Load() != torOffen {
				t.Fatalf("das Tor steht danach auf %d", a.lauf.tor.Load())
			}

			return errors.Join(fehler, a.Starten(context.Background()))
		},
	} {
		t.Run(name, func(t *testing.T) {
			ordner := t.TempDir()
			anwendung := gestartet(t)

			if err := anwendung.StartmarkeSetzen(ordner); err != nil {
				t.Fatal(err)
			}

			_ = tausch(anwendung) //nolint:errcheck // geprüft wird die Marke

			if zahl := marken(t, ordner); zahl != 1 {
				t.Fatalf("nach dem Tausch liegen %d Marken, erwartet die eigene — ein "+
					"späterer Absturz bliebe ohne Hinweis", zahl)
			}

			frist, aufhoeren := context.WithTimeout(context.Background(), 2*time.Second)
			defer aufhoeren()

			_ = anwendung.Schliessen(frist, nil) //nolint:errcheck // geprüft wird die Marke

			if zahl := marken(t, ordner); zahl != 0 {
				t.Errorf("nach dem Schluss liegen noch %d Marken", zahl)
			}
		})
	}
}

// **Ein Ende ohne Geduld nimmt die Marke mit, sobald der Tausch fertig ist**
// — Befund 1 des Nachlaufs über denselben Zweitblick. Das Wiederöffnen
// schloss danach den Bestand selbst, die Marke blieb liegen, und der nächste
// Start meldete einen Abbruch, den es nicht gab.
func TestEinEndeOhneGeduldNimmtDieMarkeNachDemTauschMit(t *testing.T) {
	a, _ := tauschbereit(t)
	ordner := t.TempDir()

	if err := a.StartmarkeSetzen(ordner); err != nil {
		t.Fatal(err)
	}

	angefangen, weiter := haltendesOeffnen(a)

	getauscht := make(chan error, 1)
	go func() { getauscht <- a.imTausch(schliessenUndOeffnen(a)) }()

	<-angefangen

	frist, aufhoeren := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer aufhoeren()

	if fehler := a.Schliessen(frist, nil); !errors.Is(fehler, ErrSchlussNachFrist) {
		t.Fatalf("das Ende gab %v, erwartet war der Schluss nach der Frist", fehler)
	}

	// **Solange der Tausch läuft, bleibt sie**: er kann noch abstürzen.
	if zahl := marken(t, ordner); zahl != 1 {
		t.Fatalf("während der Tausch läuft, liegen %d Marken", zahl)
	}

	close(weiter)

	// **Ohne Geduld heißt abgebrochen** — der Tausch endet mit dem Ende der
	// Anwendung, und die Marke geht trotzdem erst danach.
	if fehler := <-getauscht; !errors.Is(fehler, context.Canceled) {
		t.Fatalf("der Tausch endete mit %v, erwartet war der Abbruch", fehler)
	}

	if zahl := marken(t, ordner); zahl != 0 {
		t.Errorf("nach dem Tausch liegen noch %d Marken — der nächste Start meldete einen Abbruch", zahl)
	}
}
