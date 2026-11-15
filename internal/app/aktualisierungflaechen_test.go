// **Dieselbe Fähigkeit ohne Fenster** —
// #1100,
// Werkstück 7.
//
// Kommandozeile und Dienst gehen über [app.NeuesteFassung] und
// [app.FassungLaden]. Beide bauen ihre Naht **selbst** — sie haben kein
// gestartetes `App` —, und deshalb steht ihr Weg hier und nicht bei den
// Bindungen.
package app_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
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

// freigabegegenstelle gibt eine Gegenstelle mit Freigabe, Summen und Abbild.
func freigabegegenstelle(t *testing.T, fassung string, abbild []byte) {
	t.Helper()

	// **Mit gültiger Signatur**, denn ohne wird gar nicht geladen. Der
	// Schlüssel ist ein frisch erzeugter; die Prüfung selbst bleibt die
	// echte — getauscht ist nur, wessen Schlüssel gilt (§3).
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

// **Die Auskunft nennt beide Fassungen.**
func TestDieFassungsauskunftNenntBeideFassungen(t *testing.T) {
	eigenesZuhause(t)

	freigabegegenstelle(t, "99.0.0", nil)

	auskunft, err := app.NeuesteFassung(probekontext(t))
	if err != nil {
		t.Fatalf("NeuesteFassung: %v", err)
	}

	if auskunft.Laufend != app.Fassung || auskunft.Freigegeben != "99.0.0" {
		t.Errorf("die Auskunft ist %+v", auskunft)
	}

	if !auskunft.Neuer {
		t.Error("99.0.0 gilt nicht als neuer")
	}
}

// **Eine unlesbare Fassung wird gemeldet.**
func TestEineUnlesbareFreigabeMeldetSich(t *testing.T) {
	eigenesZuhause(t)

	freigabegegenstelle(t, "keine Zahl", nil)

	if _, err := app.NeuesteFassung(probekontext(t)); err == nil {
		t.Error("eine unlesbare Fassung kam durch")
	}
}

// **Eine unerreichbare Gegenstelle wird gemeldet.**
func TestEineUnerreichbareGegenstelleMeldetSich(t *testing.T) {
	eigenesZuhause(t)

	vorher := app.MitFreigabeadresse("http://127.0.0.1:1")
	t.Cleanup(func() { app.MitFreigabeadresse(vorher) })

	if _, err := app.NeuesteFassung(probekontext(t)); err == nil {
		t.Error("eine unerreichbare Gegenstelle kam durch")
	}

	if _, err := app.FassungLaden(probekontext(t), nil); err == nil {
		t.Error("das Laden ins Leere kam durch")
	}
}

// **Geladen wird ins Downloadverzeichnis, und der Ort kommt zurück.**
//
// **Mit umgelenktem Zuhause** (§3): ohne das schriebe dieser Test in das echte
// Downloadverzeichnis dieses Rechners — grün, weil er wirklich funktioniert
// hat.
func TestFassungLadenLegtDasAbbildAb(t *testing.T) {
	eigenesZuhause(t)

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	if err := os.Mkdir(filepath.Join(zuhause, "Downloads"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	freigabegegenstelle(t, "99.0.0", []byte(strings.Repeat("dmg", 500)))

	var zuletzt int64

	pfad, err := app.FassungLaden(probekontext(t), func(gelesen, _ int64) {
		zuletzt = gelesen
	})
	if err != nil {
		t.Fatalf("FassungLaden: %v", err)
	}

	if pfad != filepath.Join(zuhause, "Downloads", "MagicSetEditor4Mac-99.0.0.dmg") {
		t.Errorf("die Datei liegt unter %q", pfad)
	}

	if zuletzt != 1500 {
		t.Errorf("der Fortschritt endete bei %d", zuletzt)
	}
}

// **Ohne Heimatverzeichnis wird das gesagt, nicht geraten.**
func TestFassungLadenOhneZuhauseMeldetSich(t *testing.T) {
	t.Setenv("HOME", "")

	freigabegegenstelle(t, "99.0.0", []byte("x"))

	if _, err := app.FassungLaden(probekontext(t), nil); err == nil {
		t.Error("ohne Heimatverzeichnis kam das Laden durch")
	}
}

// **Die laufende Fassung wird über keine Fläche geladen.**
//
// # Der Befund
//
// `FassungLaden` holte die jüngste Freigabe und lud sie — **ohne** die
// Fassungen zu vergleichen. Das Fenster prüfte selbst, die Kommandozeile
// prüfte selbst, und der Dienst rief unmittelbar hierher: ein `POST
// /aktualisierung/laden` lud damit auch dieselbe oder eine ältere Fassung.
//
// Genau die Sorte Abweichung, die §16 verbietet — und der Flächenlauf sah sie
// nicht, weil sein einziger Erfolgsfall immer eine künstlich neuere Fassung
// lieferte. Angestrichen vom Gesamtreview vor 0.2.0.
//
// **Die Prüfung steht jetzt hier**, an der gemeinsamen Stelle. Wer sie an
// jeder Fläche einzeln schriebe, hätte sie an der nächsten wieder vergessen.
func TestDieLaufendeFassungWirdUeberKeineFlaecheGeladen(t *testing.T) {
	eigenesZuhause(t)

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	if err := os.Mkdir(filepath.Join(zuhause, "Downloads"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	freigabegegenstelle(t, app.Fassung, []byte("egal"))

	if _, err := app.FassungLaden(probekontext(t), nil); err == nil {
		t.Error("die laufende Fassung wurde geladen")
	}

	eintraege, err := os.ReadDir(filepath.Join(zuhause, "Downloads"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 0 {
		t.Errorf("es blieb etwas liegen: %+v", eintraege)
	}
}

// **Zwei Ladevorgänge zugleich gibt es nicht.**
//
// §7, Regel 4. Das Fenster hält seinen eigenen Wächter im Haken; der Dienst
// hatte keinen, und jeder `POST /aktualisierung/laden` fing einen weiteren
// Ladevorgang an. Angestrichen vom Gesamtreview vor 0.2.0.
func TestZweiLadevorgaengeZugleichGibtEsNicht(t *testing.T) {
	eigenesZuhause(t)

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	if err := os.Mkdir(filepath.Join(zuhause, "Downloads"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	freigabegegenstelle(t, "99.0.0", []byte(strings.Repeat("dmg", 500)))

	// **Der zweite geht los, während der erste im Fortschritt hängt.** Ohne
	// diese Verzahnung liefe der zweite nach dem ersten, und der Riegel bliebe
	// ungeprüft.
	angefangen := make(chan struct{})
	weiter := make(chan struct{})

	var zweiter error

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		<-angefangen

		_, zweiter = app.FassungLaden(probekontext(t), nil)

		close(weiter)
	}()

	einmal := sync.OnceFunc(func() { close(angefangen); <-weiter })

	if _, err := app.FassungLaden(probekontext(t), func(int64, int64) {
		einmal()
	}); err != nil {
		t.Fatalf("der erste Ladevorgang: %v", err)
	}

	<-fertig

	if !errors.Is(zweiter, domain.ErrConflict) {
		t.Errorf("der zweite Ladevorgang gab %v — will ErrConflict", zweiter)
	}
}

// **Eine unlesbare Fassung führt zu keinem Ladevorgang.**
func TestEineUnlesbareFreigabeFuehrtZuKeinemLaden(t *testing.T) {
	eigenesZuhause(t)

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	freigabegegenstelle(t, "keine Zahl", []byte("egal"))

	if _, err := app.FassungLaden(probekontext(t), nil); err == nil {
		t.Error("eine unlesbare Fassung kam durch")
	}
}

// eigenesZuhause gibt dem Lauf ein Benutzerverzeichnis für sich.
//
// # Warum das hier steht und nicht in `freigabegegenstelle`
//
// **Weil ein Lauf ausdrücklich ohne eines auskommen muss**
// (`TestFassungLadenOhneZuhauseMeldetSich`). Eine Vorrichtung, die es allen
// gäbe, nähme ihm seinen Gegenstand.
//
// # Und warum es ihn braucht
//
// Die Drossel legt seit Phase 3.2 eine Nachbardatei im Anwendungsverzeichnis
// an (#1163).
// Ohne ein eigenes Zuhause schrieben die Läufe in das **echte** dieses
// Rechners — und blieben dabei grün. §3 verlangt beides: ein Zuhause **und**
// eine umgelenkte Adresse; die Adresse war seit je umgelenkt.
func eigenesZuhause(t *testing.T) {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	// **Mit einem Downloadordner darin.** Ein echtes Zuhause hat einen; ohne
	// ihn scheiterte das Laden an einem fehlenden Verzeichnis — und der Lauf
	// prüfte, dass es fehlt, statt was er prüfen soll (§10).
	if fehler := os.MkdirAll(filepath.Join(zuhause, "Downloads"), 0o700); fehler != nil {
		t.Fatalf("den Downloadordner anlegen: %v", fehler)
	}
}
