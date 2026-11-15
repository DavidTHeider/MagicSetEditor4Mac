// **Ein Token bezieht kein Paket von innen** — SEC-P4
// (#2000),
// Befund P4 in `docs/PENTEST.md`.
//
// # Der Angriff
//
// Ein Automat mit einem Token `schreiben` ruft `POST /pakete` und nennt als
// Adresse einen Dienst, der **nur von diesem Rechner aus** zu erreichen ist —
// hier ein Lauscher auf der Schleife. Ginge der Bezug hinaus, hätte er über
// die Anwendung eine Anfrage an etwas gestellt, das er selbst nicht erreicht:
// die Gestalt eines SSRF, und an einem Metadatendienst oder einem Router im
// eigenen Netz wäre sie teuer.
//
// **Vier Schreibweisen derselben Schleife**, und jede ist eine eigene Lücke:
// `127.0.0.1`, `localhost`, `0.0.0.0` — auf macOS landet eine Verbindung
// dorthin bei diesem Rechner, gemessen am 2026-10-02 — und die
// IPv4-abgebildete Form `[::ffff:127.0.0.1]`.
//
// # Woran gemessen
//
// **Am Zähler der Gegenstelle**, nicht an der Antwort: eine Absage, die nach
// der Anfrage käme, wäre keine. Und die **Gegenprobe** fährt denselben Weg
// mit dem benannten Schalter für die Schleife — dort kommt die Anfrage an,
// sonst hieße die Null nur, dass der Weg gar nicht ging (§10).
package flaechen_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestEinTokenBeziehtKeinPaketVonInnen(t *testing.T) {
	var gefragt atomic.Int64

	innen := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt.Add(1)
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(innen.Close)

	port := innen.URL[strings.LastIndex(innen.URL, ":")+1:]

	// **Der Schalter ausdrücklich aus** — `eigenerBestand` setzt ihn für
	// jeden Lauf, und bei doppeltem Schlüssel gilt der letzte.
	umgebung := append(mitEigenerSaat(t), netz.SchleifeUmgebung+"=")

	gezeigt := profilkennungAus(t, mussKommandozeile(t, umgebung, "profil", "--json"))
	automat := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"schreiben"}`,
		"token-anlegen", gezeigt))

	alsAutomat := starteDienst(t, umgebung).mitNachweis(automat)

	for _, adresse := range []string{
		innen.URL,
		"http://localhost:" + port,
		"http://0.0.0.0:" + port,
		"http://[::ffff:127.0.0.1]:" + port,
	} {
		t.Run(adresse, func(t *testing.T) {
			vorher := gefragt.Load()

			status, antwort := alsAutomat.ruf(t, http.MethodPost, "/pakete",
				`{"adresse":"`+adresse+`","stand":"`+bezugsstand+
					`","unterverzeichnis":"pakete","name":"innen"}`)

			if nachher := gefragt.Load(); nachher != vorher {
				t.Errorf("der Bezug fragte den Dienst von innen %d-mal — Antwort %d: %s",
					nachher-vorher, status, antwort)
			}

			if status != http.StatusBadRequest {
				t.Errorf("Status %d statt 400: %s", status, antwort)
			}
		})
	}

	// **Die Gegenprobe**: mit dem Schalter erreicht derselbe Weg dieselbe
	// Gegenstelle — der Zähler misst also wirklich.
	t.Run("mit dem Schalter", func(t *testing.T) {
		geoeffnet := append(mitEigenerSaat(t), netz.SchleifeUmgebung+"=an")
		erlaubt := starteDienst(t, geoeffnet)
		vorher := gefragt.Load()

		erlaubt.ruf(t, http.MethodPost, "/pakete", `{"adresse":"`+innen.URL+
			`","stand":"`+bezugsstand+`","unterverzeichnis":"pakete","name":"innen"}`)

		if gefragt.Load() == vorher {
			t.Error("auch mit dem Schalter kam keine Anfrage an — der Zähler misst nichts")
		}
	})
}
