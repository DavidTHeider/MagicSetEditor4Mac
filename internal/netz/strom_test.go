// **Ein Strom statt eines Rumpfes**
// (#1100).
//
// [Hole] liest die ganze Antwort in den Arbeitsspeicher und deckelt sie. Für
// ein Abbild von **188 MB** ist das der falsche Weg: der Deckel müsste größer
// sein als jede Datei, die je kommt, und der Speicher trüge sie ganz.
package netz_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Der Strom kommt stückweise und trägt seine Länge.**
func TestEinStromLiefertStueckweiseUndNenntSeineLaenge(t *testing.T) {
	t.Parallel()

	inhalt := make([]byte, 3<<20) // drei Mebibyte
	for i := range inhalt {
		inhalt[i] = byte(i % 251)
	}

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Content-Length", "3145728")
			if _, fehler := schreiben.Write(inhalt); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	defer gegenstelle.Close()

	antwort, err := netz.Strom(t.Context(), netz.NeuerKlient(netz.UeblicheFristen()), nil,
		netz.Kenner{Name: "Probe", Annahme: "*/*"}, gegenstelle.URL)
	if err != nil {
		t.Fatalf("Strom: %v", err)
	}

	defer func() {
		if fehler := antwort.Rumpf.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	if antwort.Status != netz.StatusGut {
		t.Fatalf("Status %d", antwort.Status)
	}

	// **Die Länge steht vor dem Lesen fest** — ohne sie gibt es keinen
	// Fortschritt, sondern eine wachsende Zahl ohne Ziel.
	if antwort.Laenge != int64(len(inhalt)) {
		t.Errorf("die Länge ist %d statt %d", antwort.Laenge, len(inhalt))
	}

	gelesen, lesefehler := io.ReadAll(antwort.Rumpf)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen) != len(inhalt) {
		t.Errorf("%d Bytes statt %d", len(gelesen), len(inhalt))
	}
}

// **Ein 429 sperrt die Drossel — auch hier.**
//
// **Und das ist der Grund, warum der Strom durch dieselbe Naht geht.** Ein
// zweiter Weg ins Netz, der die Drossel nicht kennt, wäre die Umgehung, die
// §16 Satz 4 verbietet.
func TestEinGedrosselterStromSperrtDieDrossel(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Retry-After", "30")
			schreiben.WriteHeader(http.StatusTooManyRequests)
		}))
	defer gegenstelle.Close()

	_, err := netz.Strom(t.Context(), netz.NeuerKlient(netz.UeblicheFristen()), nil,
		netz.Kenner{Name: "Probe", Annahme: "*/*"}, gegenstelle.URL)

	if !errors.Is(err, domain.ErrGedrosselt) {
		t.Errorf("der Fehler ist %v — will ErrGedrosselt", err)
	}
}

// **Ein Strom ohne Länge ist kein Fehler.**
//
// Eine Gegenstelle, die `Transfer-Encoding: chunked` schickt, nennt keine.
// Dann steht die Länge auf −1, und der Fortschritt zeigt, was gelesen wurde,
// statt einen Anteil zu erfinden.
func TestEinStromOhneLaengeIstKeinFehler(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			spueler, kann := schreiben.(http.Flusher)
			if !kann {
				t.Error("die Antwort lässt sich nicht durchdrücken")

				return
			}

			spueler.Flush()
			if _, fehler := schreiben.Write([]byte("etwas")); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	defer gegenstelle.Close()

	antwort, err := netz.Strom(t.Context(), netz.NeuerKlient(netz.UeblicheFristen()), nil,
		netz.Kenner{Name: "Probe", Annahme: "*/*"}, gegenstelle.URL)
	if err != nil {
		t.Fatalf("Strom: %v", err)
	}

	defer func() {
		if fehler := antwort.Rumpf.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	if antwort.Laenge != -1 {
		t.Errorf("die Länge ist %d — ohne Angabe wären −1 richtig", antwort.Laenge)
	}
}

// **Eine Gegenstelle, die es nicht gibt, meldet sich als Fehler.**
func TestEinStromInsLeereScheitert(t *testing.T) {
	t.Parallel()

	_, err := netz.Strom(t.Context(), netz.NeuerKlient(netz.UeblicheFristen()), nil,
		netz.Kenner{Name: "Probe", Annahme: "*/*"}, "http://127.0.0.1:1/nichts")

	if err == nil {
		t.Error("der Strom ins Leere kam durch")
	}
}

// **Eine Adresse, die keine ist, wird gemeldet, bevor etwas hinausgeht.**
//
// Dieselbe Zusage wie bei [netz.Hole] — und sie steht hier eigens, weil ein
// Strom seinen eigenen Weg durch [http.NewRequestWithContext] nimmt.
func TestEinStromMitUnbrauchbarerAdresseWirdGemeldet(t *testing.T) {
	t.Parallel()

	if _, fehler := netz.Strom(t.Context(), netz.KlientUeber(http.DefaultClient), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"},
		"http://[::1]:namedport/"); fehler == nil {
		t.Error("eine unbrauchbare Adresse kam durch")
	}
}

// **Ein Abbruch trifft schon die Drossel.**
//
// Wer abbricht, während gewartet wird, soll nicht erst noch 188 MB anfangen.
func TestEinStromGibtBeimAbbruchInDerDrosselAuf(t *testing.T) {
	t.Parallel()

	drossel := netz.NeueDrossel(0, time.Now, nil)
	drossel.Sperre(time.Now().Add(time.Hour))

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := netz.Strom(ctx, netz.KlientUeber(http.DefaultClient), drossel,
		netz.Kenner{Name: "test", Annahme: "*/*"}, "http://127.0.0.1:1/"); fehler == nil {
		t.Error("der Abbruch in der Drossel kam durch")
	}
}
