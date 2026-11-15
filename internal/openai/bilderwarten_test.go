package openai_test

// **Ein Bildauftrag, der wirklich dauert** — Werkstück A17 von Phase 5.3
// (#1532).
//
// # Was hier gemessen wird
//
// **Der Weg durch die Naht, während die Gegenstelle schweigt.** Jedes Doppel
// in A12 antwortet **sofort**; das ist die Vorrichtung, die großzügiger ist
// als der Betrieb (§3) — und damit war ungeprüft, was in genau der Minute
// geschieht, in der das Programm nichts tun kann.
//
// Drei Gestalten, und sie sind nicht dasselbe:
//
//  1. **Langsam, aber innerhalb der Kopffrist** — das Bild kommt an.
//  2. **Jenseits der Kopffrist** — der Auftrag scheitert, und zwar **bevor**
//     der Kopf da ist.
//  3. **Abgebrochen** — die Gegenstelle **sieht** den Abbruch. Gezählt wird
//     an ihr und nicht am Aufrufer: eine Zusicherung über den eigenen
//     Rückgabewert sagt nichts darüber, ob draußen etwas endet.
//
// # Und warum kein echtes Bildmodell
//
// **Entscheidung des Auftraggebers vom 2026-09-12**: die KI-Modelle werden am
// Ende von 0.5.0 gelöscht. Was dieser Lauf **nicht** belegt, ist deshalb, ob
// ein echtes Modellbild in Format, Größe und Farbraum so ankommt, wie dieses
// Programm es erwartet — das sagt nur ein Lauf gegen den echten Dienst, und
// der kostet Geld (0007, dieselbe Logik).

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// langsamerBilddienst antwortet nach `dauer` — und hört auf den Abbruch.
//
// **Beides gehört zusammen.** Eine Gegenstelle, die schläft und den Abbruch
// **nicht** hört, hielte `server.Close` und damit die ganze Teststufe fest —
// der Fehler ist in diesem Projekt schon einmal aufgetreten.
type langsamerBilddienst struct {
	server *httptest.Server

	// begonnen und abgebrochen zählen, was die Gegenstelle wirklich gesehen
	// hat. **An ihr und nicht am Aufrufer**: dass ein Ruf zurückkommt, sagt
	// nichts darüber, ob draußen etwas endet.
	begonnen    atomic.Int32
	abgebrochen atomic.Int32
}

// einLangsamerBilddienst stellt ihn hin.
func einLangsamerBilddienst(t *testing.T, dauer time.Duration) *langsamerBilddienst {
	t.Helper()

	dienst := &langsamerBilddienst{}
	bild := base64.StdEncoding.EncodeToString(einPNG(t))

	dienst.server = httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			// **Erst den Rumpf lesen, dann warten.** Ohne das merkt die
			// Gegenstelle den Abbruch **nicht**: `net/http` startet den
			// Hintergrundleser, der einen Verbindungsabbruch bemerkt, erst
			// wenn der Anfragerumpf verbraucht ist — und ein Bildauftrag ist
			// ein POST mit Rumpf.
			//
			// **Gemessen am 2026-09-12**: ohne diese zwei Zeilen schlief der
			// Handler seine vollen 30 Sekunden weiter, während der Aufrufer
			// längst zurück war. Der erste Anlauf dieses Laufs hat genau das
			// als „die Gegenstelle sah keinen Abbruch" gemeldet — und es war
			// eine Eigenschaft der **Vorrichtung**, nicht der Naht.
			_, _ = io.Copy(io.Discard, gestellt.Body) //nolint:errcheck // die Gegenstelle

			dienst.begonnen.Add(1)

			// **Kein Kopf, kein Rumpf, solange gerechnet wird** — so
			// antwortet ein Bildmodell: erst wenn es fertig ist.
			select {
			case <-time.After(dauer):
			case <-gestellt.Context().Done():
				dienst.abgebrochen.Add(1)

				return
			}

			_ = json.NewEncoder(schreiben).Encode(map[string]any{ //nolint:errcheck // die Gegenstelle
				"created": 1,
				"data":    []map[string]any{{"b64_json": bild}},
			})
		}))

	t.Cleanup(dienst.server.Close)

	return dienst
}

// klientMit baut den Klienten mit dieser Kopffrist.
//
// **Die Stillstandsfrist steht weit darüber**, und das ist keine Beiläufigkeit
// (§10): beide enden in einer Meldung mit dem Wort „timeout", und wer sie
// gleich groß setzt, weiß nicht, welche zugeschlagen hat. Der erste Anlauf
// gab ihr fünf Sekunden — die Gegenprobe (Kopffrist auf 60 s) blieb grün,
// weil dann **die andere** Frist traf. Gemessen am 2026-09-12: 5,28 s statt
// der erwarteten 0,3 s, und der Lauf sagte nichts über die Kopffrist.
func (d *langsamerBilddienst) klientMit(kopffrist time.Duration) *openai.Klient {
	return openai.NeuerKlient(d.server.URL, netz.NeuerKlient(netz.Fristen{
		Verbindung: 5 * time.Second,
		Kopf:       kopffrist,
		Stillstand: 30 * time.Second,
	}), openai.Zugang{Schluessel: "geheim"})
}

// einAuftrag ist der kleinste, der durchgeht.
func einAuftrag() openai.Bildauftrag {
	return openai.Bildauftrag{
		Alias: "gpt-image-langsam", Auftrag: "ein Blitz",
		Groesse: "1024x1024", Guete: "medium",
	}
}

// **Ein langsamer Auftrag innerhalb der Kopffrist kommt durch.**
//
// **Die Hälfte, ohne die die andere nichts sagte** (§10): eine Frist, die
// jede Antwort abschnitte, ließe keinen Bildauftrag je zurückkommen — und der
// Lauf darunter wäre grün, ohne etwas über Fristen zu wissen.
func TestEinLangsamerBildauftragInnerhalbDerFristKommtDurch(t *testing.T) {
	t.Parallel()

	dienst := einLangsamerBilddienst(t, 400*time.Millisecond)

	begonnen := time.Now()

	bild, fehler := dienst.klientMit(5*time.Second).Bild(
		context.Background(), einAuftrag())
	if fehler != nil {
		t.Fatalf("der langsame Auftrag: %v", fehler)
	}

	if len(bild.Inhalt) == 0 {
		t.Fatal("das Bild kam leer zurück")
	}

	// **Und er hat wirklich gewartet.** Ein Lauf, der sofort zurückkäme,
	// hätte die Gestalt nicht geprüft, um die es hier geht.
	if gedauert := time.Since(begonnen); gedauert < 300*time.Millisecond {
		t.Errorf("der Auftrag war nach %s zurück — dann war er nicht langsam",
			gedauert)
	}
}

// **Ein Auftrag jenseits der Kopffrist scheitert — und die Meldung sagt es.**
//
// **Nicht „unbekannter Fehler"** (§8): wer eine Minute gewartet hat, soll
// lesen, dass eine Frist zugeschlagen hat, und nicht raten.
func TestEinBildauftragJenseitsDerKopffristScheitert(t *testing.T) {
	t.Parallel()

	dienst := einLangsamerBilddienst(t, 30*time.Second)

	_, fehler := dienst.klientMit(300*time.Millisecond).Bild(
		context.Background(), einAuftrag())
	if fehler == nil {
		t.Fatal("der schweigende Dienst kam durch — dann gibt es keine Frist")
	}

	// **Und der Fehler nennt **diese** Frist und keine andere.**
	//
	// Bis D02 (#2233) stand hier der Wortlaut von `net/http`, „timeout
	// awaiting response headers". Seitdem trägt der Fehler Klasse und Frist:
	// ein Test auf bloß „zu spät" wäre grün, wenn stattdessen die
	// Stillstandsfrist zuschlüge — und das ist eine andere Aussage über eine
	// andere Naht.
	var gegenstelle *domain.Gegenstellenfehler
	if !errors.Is(fehler, domain.ErrKeineAntwort) || !errors.As(fehler, &gegenstelle) ||
		gegenstelle.Frist != 300*time.Millisecond {
		t.Errorf("der Fehler ist %v — er nennt nicht die Kopffrist", fehler)
	}

	// **Und die Gegenstelle hat den Abbruch gesehen.** Ohne das liefe sie
	// weiter und rechnete an einem Auftrag, den niemand mehr will.
	warteAufAbbruch(t, dienst)
}

// **Ein abgebrochener Auftrag endet auch draußen.**
//
// **Gezählt an der Gegenstelle** (§10, die Reichweite einer Messung): dass
// `Bild` zurückkommt, sagt nichts darüber, ob die Verbindung fällt.
func TestEinAbgebrochenerBildauftragEndetAuchDraussen(t *testing.T) {
	t.Parallel()

	dienst := einLangsamerBilddienst(t, 30*time.Second)

	ctx, abbrechen := context.WithCancel(context.Background())

	fertig := make(chan error, 1)

	go func() {
		_, fehler := dienst.klientMit(30*time.Second).Bild(ctx, einAuftrag())
		fertig <- fehler
	}()

	// **Erst der Beleg, dass die Gegenstelle den Auftrag hat** — ein Abbruch
	// vor dem Absenden prüfte nichts.
	warteAufBeginn(t, dienst)

	abbrechen()

	select {
	case fehler := <-fertig:
		if !errors.Is(fehler, context.Canceled) {
			t.Errorf("der abgebrochene Auftrag endete mit %v", fehler)
		}

	case <-time.After(10 * time.Second):
		t.Fatal("der abgebrochene Auftrag kam nicht zurück")
	}

	warteAufAbbruch(t, dienst)
}

// warteAufBeginn wartet, bis die Gegenstelle einen Auftrag hat.
func warteAufBeginn(t *testing.T, dienst *langsamerBilddienst) {
	t.Helper()

	warteAuf(t, func() bool { return dienst.begonnen.Load() > 0 },
		"die Gegenstelle hat den Auftrag nie gesehen")
}

// warteAufAbbruch wartet, bis sie ihn als abgebrochen gezählt hat.
func warteAufAbbruch(t *testing.T, dienst *langsamerBilddienst) {
	t.Helper()

	warteAuf(t, func() bool { return dienst.abgebrochen.Load() > 0 },
		"die Gegenstelle sah keinen Abbruch — dann rechnet sie weiter")
}

// warteAuf pollt eine Bedingung, bis sie hält oder die Frist abläuft.
func warteAuf(t *testing.T, haelt func() bool, klage string) {
	t.Helper()

	frist := time.After(10 * time.Second)

	for {
		if haelt() {
			return
		}

		select {
		case <-frist:
			t.Fatal(klage)
		case <-time.After(10 * time.Millisecond):
		}
	}
}
