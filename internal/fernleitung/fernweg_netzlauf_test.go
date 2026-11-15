//go:build netzlauf

package fernleitung

// **Ein offener Fernstrom über echtes DERP, und was der Stand darüber sagt** —
// SEC-21 (#2162).
//
// # Was hier wirklich geschieht
//
// Ein Horcher und ein Wählender **in diesem Prozess**, verbunden über
// tailcat — also über ein DERP-Relais von Tailscale im Internet, ohne Konto.
// Über den Strom geht eine Nachricht hin und eine zurück: das ist die
// Positivkontrolle, dass wirklich eine Verbindung steht. Danach wird gefragt,
// was der Horcher über sie sagt.
//
// **Zwei Fälle**: der gewöhnliche Weg, auf dem magicsock einen direkten Pfad
// versuchen darf, und einer mit `TS_DEBUG_ALWAYS_USE_DERP=1`, der UDP für den
// direkten Weg abschaltet (`wgengine/magicsock/debugknobs.go`, Zeile 37,
// tailscale.com v1.101.0-pre.0.20260720143344-246c82a658b3) — dann geht alles
// über das Relais.
//
// # Was er **nicht** belegt
//
// **Zwei Rechner.** Beide Enden stehen auf demselben Mac; ein direkter Weg
// zwischen ihnen ist ein Weg über die Schleife, kein Weg durch ein NAT.
// Zwei physische Macs bleiben außerhalb des Auftrags.
//
// # Der Weg, seit tailcat v0.7.0
//
// **Bis v0.6.0 nannte tailcat ihn nicht**: `Server.Status()` baute den Stand
// ohne `WantPeers`, und die Liste blieb bei offener Verbindung leer (SEC-21).
// v0.7.0 setzt es (`tailcat.go`, `locoBackend.Status`, nachgelesen am
// 2026-10-07; upstream #116). **Dieser Lauf verlangt den Weg jetzt** — K01
// (#2187):
// gewöhnlich einen direkten, mit erzwungenem DERP das Relais und keinen
// direkten. Der direkte Weg entsteht erst, wenn Verkehr durch den Tunnel geht;
// deshalb wird bis zu einer Frist nachgefragt.
//
//	go test -tags netzlauf ./internal/fernleitung/ -run Fernweg -v -count=1
//	TS_DEBUG_ALWAYS_USE_DERP=1 go test -tags netzlauf ./internal/fernleitung/ -run Fernweg -v -count=1

import (
	"os"
	"testing"
	"time"
)

func TestEinOffenerFernwegNenntSeinenWeg(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dort := eineIdentitaet(t)
	hier := eineIdentitaet(t)

	horcher, fehler := Horche(dort)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	t.Logf("TS_DEBUG_ALWAYS_USE_DERP=%q", os.Getenv("TS_DEBUG_ALWAYS_USE_DERP"))

	genommen := make(chan error, 1)

	go func() {
		offen, nahmfehler := horcher.Nimm(t.Context())
		if nahmfehler != nil {
			genommen <- nahmfehler

			return
		}

		nachricht, lesefehler := offen.Empfange(t.Context())
		if lesefehler != nil {
			genommen <- lesefehler

			return
		}

		genommen <- offen.Sende(t.Context(), append([]byte("zurück: "), nachricht...))
	}()

	anfang := time.Now()

	gewaehlt, fehler := Waehle(t.Context(), horcher.Einladung(), hier, dort.Kennung())
	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	t.Cleanup(func() { _ = gewaehlt.Schliesse() }) //nolint:errcheck // dito

	if sendefehler := gewaehlt.Sende(t.Context(), []byte("Mausbraten")); sendefehler != nil {
		t.Fatalf("senden: %v", sendefehler)
	}

	antwort, fehler := gewaehlt.Empfange(t.Context())
	if fehler != nil {
		t.Fatalf("empfangen: %v", fehler)
	}

	if horchfehler := <-genommen; horchfehler != nil {
		t.Fatalf("die horchende Seite: %v", horchfehler)
	}

	t.Logf("Echo nach %v: %q (%d Bytes)", time.Since(anfang).Round(time.Millisecond), antwort, len(antwort))

	if string(antwort) != "zurück: Mausbraten" {
		t.Fatalf("das Echo ist %q", antwort)
	}

	nurRelais := os.Getenv("TS_DEBUG_ALWAYS_USE_DERP") != ""
	wege := wartetAufDenWeg(t, horcher, nurRelais)
	t.Logf("tailcat nennt %d Wege: %+v", len(wege), wege)
	t.Logf("der Horcher zählt %d angebunden", horcher.Angebunden())

	if len(wege) != 1 {
		t.Fatalf("tailcat nennt %d Wege, erwartet einen: %+v — ohne ihn zeigt die Anzeige »unbestimmt« (SEC-21)", len(wege), wege)
	}

	switch weg := wege[0]; {
	case nurRelais && (weg.Relais == "" || weg.Direkt != ""):
		t.Errorf("mit erzwungenem DERP erwartet: Relais und kein direkter Weg — %+v", weg)
	case !nurRelais && weg.Direkt == "":
		t.Errorf("gewöhnlich erwartet: ein direkter Weg — %+v", weg)
	}

	if horcher.Angebunden() != 1 {
		t.Errorf("%d angebunden, erwartet 1", horcher.Angebunden())
	}
}
