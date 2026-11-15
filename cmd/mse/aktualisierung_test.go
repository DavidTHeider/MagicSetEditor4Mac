// **`mse aktualisierung`** —
// #1100,
// Werkstück 7.
//
// # Warum dieser Befehl die Einstellung nicht liest
//
// **Weil sie eine andere Frage beantwortet.** Die Vorgabe „aus" verhindert,
// dass die Anwendung **von selbst** ins Netz greift (§16, Regel 6) — sie ist
// eine Aussage über ungefragtes Verhalten. Wer `mse aktualisierung` tippt,
// hat gefragt; ein Befehl, der daraufhin auf eine Einstellung verwiese, wäre
// eine Kommandozeile, die den Benutzer belehrt.
//
// **Das Fenster hält sich weiter daran**, und der Unterschied steht in
// `docs/FAEHIGKEITEN.md`.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// freigabeantwort schreibt die Freigabe als JSON.
//
// **Über [json.NewEncoder] und nicht als zusammengefügte Zeichenkette.** Der
// erste Anlauf setzte sie aus Stücken zusammen, und `anfrage.Host` fließt
// darin bis in die Antwort — gosec liest das als Einschleusung und hat recht:
// eine Gestalt, die man von Hand zusammenklebt, hält man von Hand richtig.
func freigabeantwort(
	t *testing.T, schreiben http.ResponseWriter, eigen, fassung string, groesse int,
) {
	t.Helper()

	type beigabe struct {
		Name    string `json:"name"`
		Groesse int    `json:"size"`
		Adresse string `json:"url"`
	}

	if fehler := json.NewEncoder(schreiben).Encode(struct {
		Marke    string    `json:"tag_name"`
		Seite    string    `json:"html_url"`
		Beigaben []beigabe `json:"assets"`
	}{
		Marke: "v" + fassung,
		Seite: "https://beispiel.test/v",
		Beigaben: []beigabe{
			{Name: "MagicSetEditor4Mac-" + fassung + ".dmg", Groesse: groesse, Adresse: eigen + "/abbild"},
			{Name: "pruefsummen.txt", Groesse: 8, Adresse: eigen + "/summen"},
			{Name: "pruefsummen.txt.asc", Groesse: 8, Adresse: eigen + "/signatur"},
		},
	}); fehler != nil {
		t.Errorf("schreiben: %v", fehler)
	}
}

// freigabestelle gibt eine Gegenstelle, die eine Freigabe ausliefert.
//
// **Mit beiden Beigaben**, denn ohne Prüfsummen wird gar nicht geladen — und
// eine Vorrichtung, die großzügiger ist als der Betrieb, prüft sich selbst
// (§3).
func freigabestelle(t *testing.T, fassung string, abbild []byte) {
	t.Helper()

	// **Mit gültiger Signatur**, denn ohne wird gar nicht geladen. Der
	// Schlüssel ist ein frisch erzeugter; die Prüfung selbst bleibt die echte
	// — getauscht ist nur, wessen Schlüssel gilt (§3).
	wer, schluesselfehler := openpgp.NewEntity("Probe", "Test", "p@beispiel.test", nil)
	if schluesselfehler != nil {
		t.Fatalf("Schlüssel erzeugen: %v", schluesselfehler)
	}

	vorherigerBund := app.MitFreigabebund(openpgp.EntityList{wer})
	t.Cleanup(func() { app.MitFreigabebund(vorherigerBund) })

	roh := sha256.Sum256(abbild)
	summen := hex.EncodeToString(roh[:]) + "  ./MagicSetEditor4Mac-" + fassung + ".dmg\n"

	var signatur strings.Builder
	if fehler := openpgp.ArmoredDetachSign(
		&signatur, wer, strings.NewReader(summen), nil,
	); fehler != nil {
		t.Fatalf("signieren: %v", fehler)
	}

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/latest"):
				// **Die Adresse kommt aus der Anfrage, nicht aus einer
				// geschlossenen Veränderlichen.** `httptest.NewServer` startet
				// seine Schleife **vor** der Rückgabe; wer danach `dienst.URL` in
				// eine Variable schreibt, die der Behandler liest, hat ein
				// Wettrennen — vom Race-Detektor am 2026-08-26 gefunden.
				eigen := "http://" + anfrage.Host

				freigabeantwort(t, schreiben, eigen, fassung, len(abbild))
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				if _, fehler := schreiben.Write([]byte(signatur.String())); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/summen"):
				if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			default:
				schreiben.Header().Set("Content-Length", strconv.Itoa(len(abbild)))
				if _, fehler := schreiben.Write(abbild); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			}
		}))

	t.Cleanup(dienst.Close)

	vorher := app.MitFreigabeadresse(dienst.URL)
	t.Cleanup(func() { app.MitFreigabeadresse(vorher) })
}

// **`--laden` holt das Abbild und nennt seinen Ort.**
//
// **Mit umgelenktem Zuhause** (§3): ohne das schriebe dieser Test in das
// echte Downloadverzeichnis dieses Rechners — grün, weil er wirklich
// funktioniert hat.
func TestAktualisierungLaedtUndNenntDenOrt(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	if err := os.Mkdir(filepath.Join(zuhause, "Downloads"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	freigabestelle(t, "99.0.0", []byte(strings.Repeat("abbild", 1000)))

	an, ausgabe := probeschreiber()

	if code := aktualisierungsbefehl(probekontext(t), an, []string{"--laden"}); code != CodeGut {
		t.Fatalf("der Code ist %d: %s", code, ausgabe())
	}

	pfad := filepath.Join(zuhause, "Downloads", "MagicSetEditor4Mac-99.0.0.dmg")
	if _, err := os.Stat(pfad); err != nil {
		t.Fatalf("die Datei liegt nicht da: %v", err)
	}

	if !strings.Contains(ausgabe(), pfad) {
		t.Errorf("die Ausgabe nennt den Ort nicht: %s", ausgabe())
	}

	// **Der Fortschritt ist zu sehen** — und er steht auf dem Fehlerstrom,
	// damit ein Skript den Pfad weiterreichen kann.
	if !strings.Contains(ausgabe(), "100 %") {
		t.Errorf("kein Fortschritt in der Ausgabe: %s", ausgabe())
	}
}

// **Eine Freigabe ohne Prüfsummen lädt nicht.**
func TestAktualisierungLaedtNichtOhnePruefsummen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, fehler := schreiben.Write([]byte(
				`{"tag_name":"v99.0.0","assets":[{"name":"x.dmg","url":"u"}]}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(dienst.Close)

	vorher := app.MitFreigabeadresse(dienst.URL)
	t.Cleanup(func() { app.MitFreigabeadresse(vorher) })

	an, ausgabe := probeschreiber()

	if code := aktualisierungsbefehl(probekontext(t), an, []string{"--laden"}); code != CodeFachlich {
		t.Errorf("der Code ist %d: %s", code, ausgabe())
	}
}

// probeschreiber gibt einen Schreiber und den Blick auf sein Ziel.
func probeschreiber() (an *schreiber, ausgabe func() string) {
	var raus, rfehler strings.Builder

	return &schreiber{ziel: &raus, fehler: &rfehler},
		func() string { return raus.String() + rfehler.String() }
}

// **Eine neuere Fassung wird genannt.**
func TestAktualisierungNenntEineNeuereFassung(t *testing.T) {
	freigabestelle(t, "99.0.0", nil)

	an, ausgabe := probeschreiber()

	if code := aktualisierungsbefehl(probekontext(t), an, nil); code != CodeGut {
		t.Fatalf("der Code ist %d: %s", code, ausgabe())
	}

	if !strings.Contains(ausgabe(), "99.0.0") {
		t.Errorf("die Ausgabe nennt die Fassung nicht: %s", ausgabe())
	}
}

// **Die laufende Fassung wird als solche gemeldet.**
func TestAktualisierungMeldetDieLaufendeFassung(t *testing.T) {
	freigabestelle(t, app.Fassung, nil)

	an, ausgabe := probeschreiber()

	if code := aktualisierungsbefehl(probekontext(t), an, nil); code != CodeGut {
		t.Fatalf("der Code ist %d: %s", code, ausgabe())
	}

	if strings.Contains(ausgabe(), "neuer") {
		t.Errorf("die laufende Fassung galt als neuer: %s", ausgabe())
	}
}

// **Eine unerreichbare Gegenstelle meldet sich fachlich.**
func TestAktualisierungMeldetEineUnerreichbareGegenstelle(t *testing.T) {
	vorher := app.MitFreigabeadresse("http://127.0.0.1:1")
	t.Cleanup(func() { app.MitFreigabeadresse(vorher) })

	an, ausgabe := probeschreiber()

	if code := aktualisierungsbefehl(probekontext(t), an, nil); code != CodeFachlich {
		t.Errorf("der Code ist %d: %s", code, ausgabe())
	}
}

// **Ein unbekanntes Argument wird abgewiesen.**
func TestAktualisierungWeistEinUnbekanntesArgumentAb(t *testing.T) {
	an, ausgabe := probeschreiber()

	if code := aktualisierungsbefehl(probekontext(t), an, []string{"--sofort"}); code != CodeBedienung {
		t.Errorf("der Code ist %d: %s", code, ausgabe())
	}
}
