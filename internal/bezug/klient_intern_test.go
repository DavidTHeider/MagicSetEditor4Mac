// Der Klient — Frist, Selbstnennung und was beides nicht abdeckt.
//
// # Warum ein Test **im** Paket
//
// Wie bei `grenze_intern_test.go`: die Fristen sind für den Betrieb gesetzt und
// für einen Test zu lang. Von innen lassen sie sich für die Dauer eines Falls
// herunterdrehen, und dann fährt der Test **denselben** Weg.
//
// # Der Anlass
//
// `internal/bezug` benutzte `http.DefaultClient`
// (#899).
// Dessen `Timeout` ist null, und der `ctx` kommt aus Wails und trägt keine
// Frist: ein hängender Server hielt den Paketbezug beliebig lange fest, und
// abbrechen konnte nur, wer die Anwendung beendete.
//
// Und er nannte sich nicht. Für den Paketbezug ist das unhöflich; für den
// zweiten Netzverbraucher, den Phase V bringt, ist es ein
// Ausschlusskriterium — Scryfall verlangt `User-Agent` und `Accept`
// ausdrücklich (§16).
package bezug

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mitKopffrist dreht die Wartezeit auf den Antwortkopf herunter.
func mitKopffrist(t *testing.T, neu time.Duration) {
	t.Helper()

	alt := kopffrist
	kopffrist = neu

	t.Cleanup(func() { kopffrist = alt })
}

// **Ein Server, der nicht antwortet, hält den Bezug nicht fest.**
func TestEinHaengenderServerLaeuftInDieFrist(t *testing.T) {
	mitKopffrist(t, 50*time.Millisecond)

	losgelassen := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-losgelassen
	}))

	// **Erst den Server schließen, wenn der Bearbeiter zurück ist.** Ein
	// `Close` auf einen laufenden Bearbeiter wartet auf ihn — und der wartet
	// auf den Kanal.
	t.Cleanup(func() {
		close(losgelassen)
		server.Close()
	})

	beginn := time.Now()

	_, err := laden(context.Background(), server.URL, "MagicSetEditor4Mac/1.2.3")
	if err == nil {
		t.Fatal("der hängende Server kam durch")
	}

	if gedauert := time.Since(beginn); gedauert > time.Second {
		t.Errorf("es dauerte %v — die Frist steht bei %v", gedauert, kopffrist)
	}
}

// **Und der Klient nennt sich, mit beiden Kopfzeilen.**
//
// `User-Agent` **und** `Accept`: Scryfall verlangt beide, und die Regel steht
// in §16 für jeden fremden Dienst, nicht nur für den einen.
func TestDerKlientNenntSichMitBeidenKopfzeilen(t *testing.T) {
	var gesehen http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gesehen = r.Header.Clone()

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if _, err := laden(context.Background(), server.URL, "MagicSetEditor4Mac/1.2.3"); err != nil {
		t.Fatalf("laden: %v", err)
	}

	kenner := gesehen.Get("User-Agent")
	if !strings.Contains(kenner, "MagicSetEditor4Mac") {
		t.Errorf("der User-Agent nennt die Anwendung nicht: %q", kenner)
	}

	// **Mit Fassung**, damit eine Gegenstelle einen Fehlerbericht zuordnen
	// kann. Eine Selbstnennung ohne Fassung sagt nur die Hälfte — welche
	// Fassung es ist, entscheidet der Composition Root, nicht dieses Paket.
	if !strings.Contains(kenner, "/") {
		t.Errorf("der User-Agent nennt keine Fassung: %q", kenner)
	}

	if gesehen.Get("Accept") == "" {
		t.Error("es fehlt die Kopfzeile Accept")
	}
}

// **Auch die Standprüfung nennt sich.**
//
// Sie ist der **zweite** Aufruf ins Netz, und ein Klient, der nur an einer
// Stelle richtig gebaut ist, ist die halbe Zusage aus §10.
func TestAuchDieStandpruefungNenntSich(t *testing.T) {
	var gesehen http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gesehen = r.Header.Clone()

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	quelle := Repositorium{Adresse: server.URL, Stand: "v1.0.0", Marken: server.URL + "/"}
	if err := standPruefen(context.Background(), quelle, "MagicSetEditor4Mac/1.2.3"); err != nil {
		t.Fatalf("standPruefen: %v", err)
	}

	if !strings.Contains(gesehen.Get("User-Agent"), "MagicSetEditor4Mac") {
		t.Errorf("der User-Agent nennt die Anwendung nicht: %q", gesehen.Get("User-Agent"))
	}
}

// **Ein 429 wird als das benannt, was es ist.**
//
// # Warum keine Wiederholung
//
// Ein Bezug beginnt mit einer ausdrücklichen Zustimmung (#56) und lädt ein
// Archiv von bis zu zwei Gigabyte. Eine selbsttätige Wiederholung machte aus
// einer Handlung, die jemand angestoßen hat, eine, die im Hintergrund
// weiterläuft — und die Gegenstelle sagt mit 429 gerade, dass sie das nicht
// will.
//
// **Die Drossel samt Rückzug gehört zu Phase V**
// (#888),
// wo ein Dienst viele kleine Anfragen stellt und ein `Retry-After` etwas
// bedeutet. Hier bleibt es bei der Auskunft — und die sagt jetzt, was zu tun
// ist, statt eine Statuszeile weiterzureichen.
func TestEinZuVieleAnfragenWirdBenannt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	_, err := laden(context.Background(), server.URL, "MagicSetEditor4Mac/1.2.3")
	if err == nil {
		t.Fatal("der 429 kam durch")
	}

	for _, muss := range []string{"zu viele", "später"} {
		if !strings.Contains(err.Error(), muss) {
			t.Errorf("die Meldung sagt %q nicht: %v", muss, err)
		}
	}
}

// **Ein Rumpf, der nach dem Kopf stehenbleibt, hält den Bezug nicht fest** —
// #916.
//
// # Warum die Kopffrist dafür nicht reicht
//
// `ResponseHeaderTimeout` misst bis zur Statuszeile. Wer sie schickt und dann
// schweigt, ist danach unbeobachtet — und genau das war der Befund E-37 des
// Audits: „ein hängender Server hält den Paketbezug beliebig lange fest".
// #899 hat die halbe Klasse geschlossen und es im Kommentar gesagt; der
// Zweitblick über Phase R hat die andere Hälfte eingefordert.
func TestEinStehenderRumpfLaeuftInDieFrist(t *testing.T) {
	mitLeerlauffrist(t, 50*time.Millisecond)

	losgelassen := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1024")
		w.WriteHeader(http.StatusOK)

		if spuelbar, kann := w.(http.Flusher); kann {
			spuelbar.Flush()
		}

		<-losgelassen
	}))

	t.Cleanup(func() {
		close(losgelassen)
		server.Close()
	})

	beginn := time.Now()

	_, err := laden(context.Background(), server.URL, "MagicSetEditor4Mac/1.2.3")
	if err == nil {
		t.Fatal("der stehende Rumpf kam durch")
	}

	if gedauert := time.Since(beginn); gedauert > 2*time.Second {
		t.Errorf("es dauerte %v — die Leerlauffrist steht bei %v", gedauert, leerlauffrist)
	}
}

// **Ein langsamer, aber fortschreitender Bezug wird nicht abgebrochen.**
//
// Der Gegenbeleg, ohne den die Abhilfe schlimmer wäre als der Fehler: ein
// Paket darf zwei Gigabyte wiegen, und über eine langsame Leitung dauert das.
// Abgebrochen wird **Stillstand**, nicht Langsamkeit.
func TestEinLangsamerAberFortschreitenderBezugLaeuftDurch(t *testing.T) {
	mitLeerlauffrist(t, 200*time.Millisecond)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		spuelbar, kann := w.(http.Flusher)

		for i := 0; i < 6; i++ {
			if _, err := w.Write([]byte("x")); err != nil {
				return
			}

			if kann {
				spuelbar.Flush()
			}

			time.Sleep(50 * time.Millisecond)
		}
	}))
	t.Cleanup(server.Close)

	gelesen, err := laden(context.Background(), server.URL, "MagicSetEditor4Mac/1.2.3")
	if err != nil {
		t.Fatalf("ein fortschreitender Bezug wurde abgebrochen: %v", err)
	}

	if len(gelesen) != 6 {
		t.Errorf("%d Bytes, erwartet 6", len(gelesen))
	}
}

// **Auch die Standprüfung meldet ein 429 als das, was es ist** —
// #917.
//
// Sie läuft **vor** dem Archivbezug. Drosselt die Gegenstelle schon dort, sagte
// die Meldung bis zum 2026-08-23 „%q ist keine Freigabe dieses Repositoriums"
// — und die Person suchte den Fehler an ihrer Eingabe.
func TestAuchDieStandpruefungMeldetEin429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	quelle := Repositorium{Adresse: server.URL, Stand: "v1.0.0", Marken: server.URL + "/"}

	err := standPruefen(context.Background(), quelle, "MagicSetEditor4Mac/1.2.3")
	if err == nil {
		t.Fatal("der 429 kam durch")
	}

	for _, muss := range []string{"zu viele", "später"} {
		if !strings.Contains(err.Error(), muss) {
			t.Errorf("die Meldung sagt %q nicht: %v", muss, err)
		}
	}
}

// mitLeerlauffrist dreht die Frist für den Lesefortschritt herunter.
func mitLeerlauffrist(t *testing.T, neu time.Duration) {
	t.Helper()

	alt := leerlauffrist
	leerlauffrist = neu

	t.Cleanup(func() { leerlauffrist = alt })
}

// Hier stand `TestEineGeschlosseneVerbindungMeldetSich`.
//
// **Er ist mit `fristverbindung` nach `internal/netz` gewandert**
// (#1002) und
// heißt dort `TestEineGeschlosseneVerbindungLiestNicht`. Der Transport steht
// jetzt an einer Stelle für alle Dienste; ihn hier weiter zu prüfen hieße, eine
// Kopie zu prüfen, die es nicht mehr gibt.
