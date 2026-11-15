package netz_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Server, der schweigt, hält uns nicht fest**
// (#1002).
//
// # Warum das die wichtigste der drei Fristen ist
//
// `ResponseHeaderTimeout` misst bis zur Statuszeile. Wer sie schickt und dann
// **schweigt**, war danach unbeobachtet — das war Audit-Befund E-37, und
// #916 hat
// ihn geschlossen. Hier wird er gehalten, jetzt an einer Stelle für alle
// Dienste.
//
// # Wie er hergestellt wird
//
// Ein Server, der den Kopf schickt und danach nichts mehr — die Verbindung
// bleibt offen, es kommt nur nichts. Ohne die Frist bliebe der Test bis zur
// Zeitgrenze stehen.
func TestEinSchweigenderServerHaeltUnsNichtFest(t *testing.T) {
	t.Parallel()

	losgelassen := make(chan struct{})
	t.Cleanup(func() { close(losgelassen) })

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("Content-Length", "1000")
			schreiben.WriteHeader(http.StatusOK)

			if flusher, geht := schreiben.(http.Flusher); geht {
				flusher.Flush()
			}

			// Und dann schweigt er, bis der Test vorbei ist.
			select {
			case <-losgelassen:
			case <-anfrage.Context().Done():
			}
		}))
	t.Cleanup(gegenstelle.Close)

	klient := netz.NeuerKlient(netz.Fristen{
		Verbindung: time.Second,
		Kopf:       time.Second,
		Stillstand: 200 * time.Millisecond,
	})

	los := time.Now()

	_, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20)
	if fehler == nil {
		t.Fatal("der schweigende Server kam durch")
	}

	// **Und zwar schnell.** Ohne die Stillstandsfrist liefe der Test bis zur
	// Zeitgrenze des Pakets.
	if gedauert := time.Since(los); gedauert > 5*time.Second {
		t.Errorf("es dauerte %v — die Stillstandsfrist greift nicht", gedauert)
	}
}

// **Und ein Server, der langsam aber stetig schickt, läuft durch.**
//
// # Warum dieser Fall danebensteht
//
// Ohne ihn wäre auch eine Frist grün, die **die Dauer** begrenzt statt des
// Stillstands — und die bräche jeden großen Bezug über eine langsame Leitung
// ab. Eine Bulk-Datei von Scryfall wiegt 74 MB.
func TestEinLangsamerServerLaeuftDurch(t *testing.T) {
	t.Parallel()

	const stuecke = 5

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			flusher, geht := schreiben.(http.Flusher)

			for range stuecke {
				schreibe(t, schreiben, []byte("xxxxx"))

				if geht {
					flusher.Flush()
				}

				time.Sleep(60 * time.Millisecond)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	// **Die Frist ist kürzer als die Gesamtdauer und länger als eine Pause.**
	// Genau daran zeigt sich, dass sie den Stillstand misst.
	klient := netz.NeuerKlient(netz.Fristen{
		Verbindung: time.Second,
		Kopf:       time.Second,
		Stillstand: 200 * time.Millisecond,
	})

	antwort, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20)
	if fehler != nil {
		t.Fatalf("der langsame Server wurde abgebrochen: %v", fehler)
	}

	if len(antwort.Rumpf) != stuecke*5 {
		t.Errorf("es kamen %d Bytes an, erwartet waren %d", len(antwort.Rumpf), stuecke*5)
	}
}

// **Die üblichen Fristen sind gesetzt und in der richtigen Ordnung.**
//
// # Was dieser Test hält, und was nicht
//
// Er hält, dass die drei Werte **da** sind und zueinander passen: eine
// Stillstandsfrist, die kürzer wäre als der Verbindungsaufbau, ergäbe keinen
// Sinn. Er hält **nicht**, dass die Zahlen richtig gewählt sind — das ist eine
// Frage an die Leitung und an die Gegenstelle, und sie steht am Typ.
func TestDieUeblichenFristenSindGesetzt(t *testing.T) {
	t.Parallel()

	fristen := netz.UeblicheFristen()

	if fristen.Verbindung <= 0 || fristen.Kopf <= 0 || fristen.Stillstand <= 0 {
		t.Fatalf("eine Frist ist nicht gesetzt: %+v", fristen)
	}

	if fristen.Verbindung > fristen.Kopf {
		t.Errorf("der Verbindungsaufbau (%v) darf nicht länger dauern als der Kopf (%v)",
			fristen.Verbindung, fristen.Kopf)
	}
}

// **Ein Wähler, der nicht durchkommt, meldet es.**
//
// Der Fehlerweg des Wählers — er wird sonst von keinem Test betreten.
func TestEinWaehlerDerNichtDurchkommtMeldetEs(t *testing.T) {
	t.Parallel()

	// **Eine Adresse, an der niemand lauscht.** Port 0 auf der Rückschleife
	// wird vom Betriebssystem sofort abgewiesen.
	lauscher, fehler := net.Listen("tcp", "127.0.0.1:0") //nolint:noctx // ein Lauscher, der sofort wieder zugeht
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	adresse := lauscher.Addr().String()

	if err := lauscher.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	klient := netz.NeuerKlient(netz.UeblicheFristen())

	if _, err := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "test", Annahme: "*/*"},
		"http://"+adresse+"/", 1<<20); err == nil {
		t.Error("eine Verbindung ins Leere kam durch")
	}
}

// **Die langen Fristen sind länger als die üblichen — und der Grund steht am
// Typ.**
//
// **Der Wächter hält die Ordnung und nicht die Zahl**: dass 180 Sekunden die
// richtige Lesefrist für den CardTrader-Export sind, steht beim Anbieter und
// in [0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md);
// hier steht, dass sie **länger** ist als die übliche — sonst wäre die
// eigene Frist keine (Befund 6 des Zweitblicks über Phase 5.6).
func TestDieLangenFristenSindLaenger(t *testing.T) {
	t.Parallel()

	lang, ueblich := netz.LangeFristen(), netz.UeblicheFristen()

	switch {
	case lang.Kopf <= ueblich.Kopf:
		t.Errorf("der Kopf ist %v statt länger als %v", lang.Kopf, ueblich.Kopf)
	case lang.Stillstand <= ueblich.Stillstand:
		t.Errorf("der Stillstand ist %v statt länger als %v",
			lang.Stillstand, ueblich.Stillstand)
	case lang.Verbindung != ueblich.Verbindung:
		t.Errorf("der Verbindungsaufbau ist %v — dass ein Rechner nicht "+
			"antwortet, merkt man so schnell wie sonst", lang.Verbindung)
	}
}
