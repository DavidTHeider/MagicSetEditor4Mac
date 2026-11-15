// **Stufe 2: holen**
// (#1100).
//
// > Sie lädt sie ins **Standard-Downloadverzeichnis** und bittet um die
// > Installation.
//
// **Die Prüfsumme wird geprüft, bevor irgendetwas gemeldet wird.**
// `pruefsummen.txt` liegt jeder Freigabe bei, und seit dem 2026-08-25 auch
// eine Signatur darüber (#892).
package aktualisierung_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abbild ist ein Inhalt, der groß genug für mehrere Blöcke ist.
func abbild() []byte {
	inhalt := make([]byte, 1<<20)
	for i := range inhalt {
		inhalt[i] = byte(i % 251)
	}

	return inhalt
}

// summe gibt die SHA-256 eines Inhalts, wie `shasum -a 256` sie schreibt.
func summe(inhalt []byte) string {
	roh := sha256.Sum256(inhalt)

	return hex.EncodeToString(roh[:])
}

// gegenstelleMit gibt einen Dienst, der Abbild, Prüfsummen und Signatur gibt.
//
// **Mit Signatur**, denn ohne wird gar nicht geladen — und eine Vorrichtung,
// die großzügiger ist als der Betrieb, prüft den falschen Weg (§3). Der
// Schlüssel dazu ist ein eigener; warum, steht in `probeschluessel_test.go`.
func gegenstelleMit(t *testing.T, inhalt []byte, pruefsummen string) *httptest.Server {
	t.Helper()

	signiere := probeschluessel(t)
	signatur := signiere(pruefsummen)

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if strings.HasSuffix(anfrage.URL.Path, "/signatur") {
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			if strings.HasSuffix(anfrage.URL.Path, "/pruefsummen") {
				if _, fehler := schreiben.Write([]byte(pruefsummen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			// **Die Länge steht dran.** GitHub liefert seine Beigaben mit
			// `Content-Length`; ohne sie würde dieser Dienst ab 2 KiB
			// stückeln, und der Test prüfte einen Fortschritt, den es so
			// nicht gibt.
			schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))
			// **Ein Abbruch ist hier kein Fehlschlag.** Wer das Ziel nicht
			// anlegen kann, schließt den Rumpf ungelesen — und dann
			// scheitert dieses Schreiben mit „connection reset by peer".
			// Das ist die geprüfte Wirkung und nicht ihr Gegenteil.
			if _, fehler := schreiben.Write(inhalt); fehler != nil {
				t.Logf("der Klient ging weg: %v", fehler)
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst
}

// **Das Abbild wird geladen, geprüft und liegt danach da.**
func TestEineFreigabeWirdGeladenUndGeprueft(t *testing.T) {
	inhalt := abbild()
	gegenstelle := gegenstelleMit(t, inhalt,
		summe(inhalt)+"  ./MagicSetEditor4Mac-0.2.0.dmg\n")

	ziel := t.TempDir()

	var zuletzt, gesamt int64

	pfad, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, ziel, func(gelesen, ganz int64) {
			zuletzt, gesamt = gelesen, ganz
		})
	if err != nil {
		t.Fatalf("Laden: %v", err)
	}

	if filepath.Base(pfad) != "MagicSetEditor4Mac-0.2.0.dmg" {
		t.Errorf("die Datei heißt %q", filepath.Base(pfad))
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen) != len(inhalt) {
		t.Errorf("%d Bytes statt %d", len(gelesen), len(inhalt))
	}

	// **Der Fortschritt ist zu sehen** — 188 MB sind kein Augenblick.
	if zuletzt != int64(len(inhalt)) || gesamt != int64(len(inhalt)) {
		t.Errorf("der Fortschritt endete bei %d von %d", zuletzt, gesamt)
	}
}

// **Eine falsche Prüfsumme lässt keine Datei zurück.**
//
// **Das ist der Kern von Stufe 2.** Eine geladene Datei, die nicht zu ihrer
// Prüfsumme passt, ist entweder halb angekommen oder ausgetauscht — und in
// beiden Fällen darf sie nicht dastehen, wo jemand sie anklickt.
func TestEineFalschePruefsummeLaesstNichtsZurueck(t *testing.T) {
	inhalt := abbild()
	gegenstelle := gegenstelleMit(t, inhalt,
		strings.Repeat("0", 64)+"  ./MagicSetEditor4Mac-0.2.0.dmg\n")

	ziel := t.TempDir()

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, ziel, nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", err)
	}

	eintraege, lesefehler := os.ReadDir(ziel)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(eintraege) != 0 {
		t.Errorf("es blieb etwas liegen: %+v", eintraege)
	}
}

// **Eine Freigabe ohne Prüfsummen wird abgewiesen — vor dem Laden.**
//
// 188 MB zu holen, um danach zu sagen „ich kann es nicht prüfen", wäre die
// falsche Reihenfolge.
func TestEineFreigabeOhnePruefsummenWirdAbgewiesen(t *testing.T) {
	gegenstelle := gegenstelleMit(t, abbild(), "")

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung: "0.2.0",
			Abbild:  beigabe("x.dmg", gegenstelle.URL+"/abbild"),
		}, t.TempDir(), nil)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Und eine ohne Abbild ebenso.**
func TestEineFreigabeOhneAbbildWirdAbgewiesen(t *testing.T) {
	gegenstelle := gegenstelleMit(t, abbild(), "x  ./y\n")

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, t.TempDir(), nil)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Eine Prüfsummenliste ohne die eigene Zeile wird abgewiesen.**
//
// Sie kann fremde Zeilen tragen — das `.dmg`, zwei Stücklisten —, aber wenn
// die des Abbilds fehlt, gibt es nichts zu vergleichen.
func TestEinePruefsummenlisteOhneDieEigeneZeileWirdAbgewiesen(t *testing.T) {
	inhalt := abbild()
	gegenstelle := gegenstelleMit(t, inhalt,
		summe(inhalt)+"  ./stueckliste-go.json\n")

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, t.TempDir(), nil)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Eine Gegenstelle, die es nicht gibt, meldet sich als Fehler.**
func TestEinLadenInsLeereScheitert(t *testing.T) {
	gegenstelle := gegenstelleMit(t, abbild(), "x  ./MagicSetEditor4Mac-0.2.0.dmg\n")

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", "http://127.0.0.1:1/nichts"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, t.TempDir(), nil)

	if err == nil {
		t.Error("das Laden ins Leere kam durch")
	}

	// **Am Netz gescheitert, nicht an der Prüfung** (SEC03, #2356): mit einem
	// Abbild ohne Fassung im Namen wäre dieser Lauf an der Namensprüfung rot
	// geworden und hätte das Laden ins Leere gar nicht erreicht — so stand er
	// nach dem Einbau der Prüfung da und war grün.
	if errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist ein Konflikt (%v) — das Laden kam nicht bis ans Netz", err)
	}
}

// beigabe baut eine Beigabe mit Namen und Adresse.
func beigabe(name, adresse string) aktualisierung.Beigabe {
	return aktualisierung.Beigabe{Name: name, Adresse: adresse}
}
