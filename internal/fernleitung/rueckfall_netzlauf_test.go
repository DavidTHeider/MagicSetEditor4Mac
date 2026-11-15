//go:build netzlauf

package fernleitung

// **Der Wegfall der Pfadsuche während einer Verbindung, und was die Anzeige
// des Wegs daraus macht** — D03 (#2234).
//
// `fernweg_netzlauf_test.go` belegt die beiden Wege **je Prozessstart**:
// gewöhnlich direkt, mit `TS_DEBUG_ALWAYS_USE_DERP=1` über das Relais. Hier
// fällt der direkte Weg **während** der Verbindung weg und kommt wieder.
//
// # Wie nachgebildet wird, und was das heißt
//
// **Mit `TS_DEBUG_NEVER_DIRECT_UDP`**, zur Laufzeit über `envknob.Setenv`
// (`envknob/envknob.go`, Zeile 94). Nachgelesen am 2026-10-10 in
// tailscale.com v1.103.0-pre.0.20260916030321-a2263542f260: der Knopf sperrt
// **die Pfadsuche**, nicht die Daten — keine Disco-Pings an direkte Adressen
// mehr (`wgengine/magicsock/endpoint.go`, `startDiscoPingLocked` Zeile 1352,
// `sendDiscoPingsLocked` Zeile 1416), keine Antwort auf einen
// (`magicsock.go`, Zeile 2534). Ein direkter Weg gilt danach noch
// `trustUDPAddrDuration` = 6,5 s (`magicsock.go`, Zeile 4014); ist die Frist
// um, schickt magicsock **jedes Paket an beide**, den alten direkten Weg und
// DERP (`addrForSendLocked`, `endpoint.go` ab Zeile 632).
//
// # Was er belegt
//
// - **Die Anzeige folgt**: nach der Frist nennt der Stand das Relais und
//   keinen direkten Weg mehr — gemessen 6,58 s am 2026-10-10.
// - **Sie kehrt zurück**: ist die Pfadsuche wieder erlaubt, nennt der Stand
//   den direkten Weg.
// - **Kein Echo geht verloren**, über den ganzen Wechsel.
//
// # Was er **nicht** belegt
//
// - **Eine Lücke im Verkehr.** Der direkte Weg über die Schleife bleibt
//   offen, und die doppelt geschickten Pakete kommen über ihn zuerst an — die
//   Echos bleiben unter 10 ms. Die Lücke misst `udpsperre_netzlauf_test.go`,
//   das den UDP-Socket einer Seite wirklich sperrt: 6,46 s.
// - **Zwei NAT-Grenzen.** Beide Enden stehen auf demselben Mac.
//
//	go test -tags netzlauf ./internal/fernleitung/ -run AnzeigeFolgt -v -count=1

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
	"time"

	"tailscale.com/envknob"
)

const ohneDirekt = "TS_DEBUG_NEVER_DIRECT_UDP"

// echo ist ein Hin und Zurück über den Fernweg, mit seiner Zeit.
type echo struct {
	zeit   time.Time
	dauer  time.Duration
	direkt bool
	relais bool
}

func TestDieAnzeigeFolgtDemWegfallDerPfadsucheUndKehrtZurueck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() { envknob.Setenv(ohneDirekt, "") })

	dort := eineIdentitaet(t)
	hier := eineIdentitaet(t)

	horcher, fehler := Horche(dort)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	// **Die horchende Seite wirft jede Nachricht zurück**, bis der Test endet.
	go func() {
		offen, nahmfehler := horcher.Nimm(t.Context())
		if nahmfehler != nil {
			return
		}

		for {
			nachricht, lesefehler := offen.Empfange(t.Context())
			if lesefehler != nil {
				return
			}

			if offen.Sende(t.Context(), nachricht) != nil {
				return
			}
		}
	}()

	gewaehlt, fehler := Waehle(t.Context(), horcher.Einladung(), hier, dort.Kennung())
	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	t.Cleanup(func() { _ = gewaehlt.Schliesse() }) //nolint:errcheck // dito

	var gezaehlt int

	// hinUndZurueck schickt eine Nachricht und wartet auf ihr Echo. **Ein
	// verlorenes Echo ist ein Fehler**, kein Stocken: der Strom darf
	// langsamer werden, aber nichts verlieren.
	hinUndZurueck := func() echo {
		t.Helper()

		gezaehlt++
		inhalt := fmt.Appendf(nil, "%d", gezaehlt)
		anfang := time.Now()

		if sendefehler := gewaehlt.Sende(t.Context(), inhalt); sendefehler != nil {
			t.Fatalf("senden %d: %v", gezaehlt, sendefehler)
		}

		zurueck, lesefehler := gewaehlt.Empfange(t.Context())
		if lesefehler != nil {
			t.Fatalf("empfangen %d: %v", gezaehlt, lesefehler)
		}

		if !bytes.Equal(zurueck, inhalt) {
			t.Fatalf("Echo %d kam als %q", gezaehlt, zurueck)
		}

		wege := horcher.Wege()
		e := echo{zeit: anfang, dauer: time.Since(anfang)}

		if len(wege) == 1 {
			e.direkt = wege[0].Direkt != ""
			e.relais = wege[0].Relais != ""
		}

		return e
	}

	// abschnitt fährt Echos im Takt von 250 ms, bis `fertig` sagt, dass es
	// reicht, oder die Frist um ist.
	abschnitt := func(name string, frist time.Duration, fertig func(echo) bool) []echo {
		t.Helper()

		var reihe []echo

		for ende := time.Now().Add(frist); time.Now().Before(ende); time.Sleep(250 * time.Millisecond) {
			e := hinUndZurueck()
			reihe = append(reihe, e)

			if fertig(e) {
				break
			}
		}

		t.Logf("%s: %d Echos, längstes %v, Median %v, zuletzt direkt=%v relais=%v",
			name, len(reihe), laengstes(reihe), median(reihe), reihe[len(reihe)-1].direkt, reihe[len(reihe)-1].relais)

		return reihe
	}

	// **1 — der direkte Weg steht.** Er entsteht erst unter Verkehr.
	direkt := abschnitt("direkt", 30*time.Second, func(e echo) bool { return e.direkt })
	if !direkt[len(direkt)-1].direkt {
		t.Fatalf("in 30 s kein direkter Weg — ohne ihn gibt es nichts, wovon zurückzufallen wäre")
	}

	// **2 — die Pfadsuche fällt weg.** Gewartet wird, bis der Stand das
	// Relais ohne direkten Weg nennt: die Frist von 6,5 s und ein Herzschlag
	// von 3 s, dazu Spiel — **12 s**. Danach noch fünf Sekunden Verkehr.
	envknob.Setenv(ohneDirekt, "true")
	wegfall := time.Now()
	rueckfall := abschnitt("Rückfall", 30*time.Second, func(e echo) bool { return e.relais && !e.direkt })
	angezeigt := time.Since(wegfall)
	rueckfall = append(rueckfall, abschnitt("nach dem Wechsel der Anzeige", 5*time.Second, func(echo) bool { return false })...)

	if letzter := rueckfall[len(rueckfall)-1]; letzter.direkt || !letzter.relais {
		t.Errorf("30 s nach dem Wegfall nennt der Stand direkt=%v relais=%v — erwartet nur das Relais", letzter.direkt, letzter.relais)
	}

	if angezeigt > 12*time.Second {
		t.Errorf("der Stand nannte das Relais erst nach %v — erwartet binnen Frist und Herzschlag (12 s)", angezeigt)
	}

	t.Logf("der Stand nannte das Relais nach %v; längstes Echo dabei %v (über den offenen direkten Weg, siehe Kopf)", angezeigt.Round(time.Millisecond), laengstes(rueckfall))

	// **3 — die Pfadsuche geht wieder.** Gewartet wird auf den direkten Weg.
	envknob.Setenv(ohneDirekt, "")
	wiederkehr := time.Now()
	rueckkehr := abschnitt("Rückkehr", 30*time.Second, func(e echo) bool { return e.direkt })

	if !rueckkehr[len(rueckkehr)-1].direkt {
		t.Errorf("30 s nach der Wiederkehr kein direkter Weg — die Anzeige bleibt beim Relais")
	}

	t.Logf("der Stand nannte den direkten Weg nach %v wieder", time.Since(wiederkehr).Round(time.Millisecond))
}

func laengstes(reihe []echo) time.Duration {
	var laengste time.Duration

	for _, e := range reihe {
		laengste = max(laengste, e.dauer)
	}

	return laengste.Round(time.Millisecond)
}

func median(reihe []echo) time.Duration {
	dauern := make([]time.Duration, 0, len(reihe))
	for _, e := range reihe {
		dauern = append(dauern, e.dauer)
	}

	slices.Sort(dauern)

	return dauern[len(dauern)/2].Round(time.Millisecond)
}

// **Die Latenz über das Relais** — D03: mit `TS_DEBUG_ALWAYS_USE_DERP`, vor dem
// Aufbau gesetzt, bindet magicsock kein UDP für den direkten Weg
// (`magicsock.go`, Zeile 3706), und jedes Echo geht über ein DERP-Relais im
// Internet. Zwanzig Echos, Median und längstes.
func TestDasRelaisTraegtUndNenntSeineLatenz(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	envknob.Setenv("TS_DEBUG_ALWAYS_USE_DERP", "true")
	t.Cleanup(func() { envknob.Setenv("TS_DEBUG_ALWAYS_USE_DERP", "") })

	dort := eineIdentitaet(t)
	hier := eineIdentitaet(t)

	horcher, fehler := Horche(dort)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	go func() {
		offen, nahmfehler := horcher.Nimm(t.Context())
		if nahmfehler != nil {
			return
		}

		for {
			nachricht, lesefehler := offen.Empfange(t.Context())
			if lesefehler != nil || offen.Sende(t.Context(), nachricht) != nil {
				return
			}
		}
	}()

	gewaehlt, fehler := Waehle(t.Context(), horcher.Einladung(), hier, dort.Kennung())
	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	t.Cleanup(func() { _ = gewaehlt.Schliesse() }) //nolint:errcheck // dito

	reihe := make([]echo, 0, 20)

	for stelle := range 20 {
		inhalt := fmt.Appendf(nil, "%d", stelle)
		anfang := time.Now()

		if sendefehler := gewaehlt.Sende(t.Context(), inhalt); sendefehler != nil {
			t.Fatalf("senden %d: %v", stelle, sendefehler)
		}

		if _, lesefehler := gewaehlt.Empfange(t.Context()); lesefehler != nil {
			t.Fatalf("empfangen %d: %v", stelle, lesefehler)
		}

		reihe = append(reihe, echo{zeit: anfang, dauer: time.Since(anfang)})
	}

	wege := wartetAufDenWeg(t, horcher, true)
	if len(wege) != 1 || wege[0].Direkt != "" || wege[0].Relais == "" {
		t.Errorf("erwartet nur das Relais, der Stand nennt %+v", wege)
	}

	t.Logf("über das Relais %v: Median %v, längstes %v (20 Echos)", wege, median(reihe), laengstes(reihe))
}
