//go:build netzlauf

package fernleitung

// **UDP fällt wirklich weg — und wie lange der Verkehr dann steht** — D03
// (#2234).
//
// `rueckfall_netzlauf_test.go` sperrt mit `TS_DEBUG_NEVER_DIRECT_UDP` nur die
// Pfadsuche; die Daten liefen weiter direkt, und eine Lücke war so nicht zu
// messen. Hier wird **der UDP-Socket der horchenden Seite selbst** gesperrt,
// ohne Systemeinstellung:
//
// # Wie gesperrt wird
//
// magicsock bindet seine Sockets neu (`Conn.Rebind`), und ist dabei
// `TS_DEBUG_ALWAYS_USE_DERP` gesetzt, nimmt es statt eines Sockets seinen
// **Verwerfer**: Schreiben geht still verloren, Lesen bringt nie etwas
// (`newBlockForeverConn`, `wgengine/magicsock/magicsock.go` Zeile 3706;
// `blockforever_conn.go` Zeile 33) — das Verhalten einer Firewall, die UDP
// verwirft. DERP spricht TCP und bleibt. Geöffnet wird ebenso: Schalter
// zurück, wieder `Rebind`, ein neuer Socket. Nachgelesen am 2026-10-10 in
// tailscale.com v1.103.0-pre.0.20260916030321-a2263542f260.
//
// **An `Rebind` kommt nur, wer den `magicsock.Conn` hat**, und tailcat reicht
// ihn nicht heraus. [magicsockVon] liest ihn aus den unexportierten Feldern
// des gepinnten tailcat v0.7.0 (`Server.lb.sys.MagicSock`) — und scheitert
// laut, wenn sich der Weg ändert.
//
// # Wer es weiß, und wer nicht
//
// `Rebind` setzt die Wege **der gesperrten Seite** zurück
// (`resetEndpointStates`, Zeile 3829): sie schickt sofort über DERP. **Die
// wählende Seite weiß nichts** — sie schickt weiter an den toten Port, bis
// ihr direkter Weg nach 6,5 s seine Geltung verliert. Gemessen wird an ihr:
// das ist die Lage eines Menschen, dessen Gegenüber hinter eine Firewall
// gerät. **Was das nicht ist**: eine Firewall auf der eigenen Seite — dort
// wüsste auch die gesperrte Seite nichts und schickte bis zum Ende ihrer
// eigenen Frist an den toten Weg.
//
//	go test -tags netzlauf ./internal/fernleitung/ -run GesperrtesUDP -v -count=1

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"tailscale.com/envknob"
	"tailscale.com/tsd"
	"tailscale.com/wgengine/magicsock"
)

// magicsockVon gibt den `magicsock.Conn` hinter einem Horcher — über die
// unexportierten Felder von tailcat v0.7.0. **Jeder Schritt prüft seinen
// Typ**: ein anderer Aufbau nach einem Heben soll den Lauf anhalten, nicht ins
// Leere greifen.
func magicsockVon(t *testing.T, horcher *Horcher) *magicsock.Conn {
	t.Helper()

	naht, istTailcat := horcher.server.(*tailcatserver)
	if !istTailcat {
		t.Fatalf("der Horcher hat keinen tailcat-Server, sondern %T", horcher.server)
	}

	lb := reflect.ValueOf(naht.unten).Elem().FieldByName("lb")
	if !lb.IsValid() || lb.Kind() != reflect.Pointer || lb.IsNil() {
		t.Fatalf("tailcat.Server.lb fehlt oder ist leer — tailcat hat seinen Aufbau geändert")
	}

	//nolint:gosec // ein Messlauf liest ein unexportiertes Feld einer gepinnten Abhängigkeit
	backend := reflect.NewAt(lb.Type(), unsafe.Pointer(lb.UnsafeAddr())).Elem().Elem()

	sys := backend.FieldByName("sys")
	if !sys.IsValid() || sys.Type() != reflect.TypeFor[tsd.System]() {
		t.Fatalf("locoBackend.sys ist kein tsd.System — tailcat hat seinen Aufbau geändert")
	}

	//nolint:gosec // dito
	system := (*tsd.System)(unsafe.Pointer(sys.UnsafeAddr()))

	conn, gesetzt := system.MagicSock.GetOK()
	if !gesetzt || conn == nil {
		t.Fatal("tsd.System.MagicSock ist nicht gesetzt")
	}

	return conn
}

func TestEinGesperrtesUDPFaelltAufsRelaisZurueckUndKehrtZurueck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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

	conn := magicsockVon(t, horcher)

	var gezaehlt int

	// hinUndZurueck schickt eine Nachricht und wartet auf ihr Echo — **ein
	// verlorenes Echo ist ein Fehler**: der Strom darf stocken, nicht
	// verlieren.
	hinUndZurueck := func() echo {
		t.Helper()

		gezaehlt++
		inhalt := fmt.Appendf(nil, "%d", gezaehlt)
		anfang := time.Now()

		if sendefehler := gewaehlt.Sende(t.Context(), inhalt); sendefehler != nil {
			t.Fatalf("senden %d: %v", gezaehlt, sendefehler)
		}

		if _, lesefehler := gewaehlt.Empfange(t.Context()); lesefehler != nil {
			t.Fatalf("empfangen %d: %v", gezaehlt, lesefehler)
		}

		e := echo{zeit: anfang, dauer: time.Since(anfang)}
		if wege := horcher.Wege(); len(wege) == 1 {
			e.direkt, e.relais = wege[0].Direkt != "", wege[0].Relais != ""
		}

		return e
	}

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

		t.Logf("%s: %d Echos, längstes %v, Median %v", name, len(reihe), laengstes(reihe), median(reihe))

		return reihe
	}

	// **1 — direkt**, und ein paar Echos, damit der Median etwas sagt.
	direkt := abschnitt("direkt", 30*time.Second, func(e echo) bool { return e.direkt })
	if !direkt[len(direkt)-1].direkt {
		t.Fatalf("in 30 s kein direkter Weg")
	}

	direkt = append(direkt, abschnitt("direkt, eingeschwungen", 3*time.Second, func(echo) bool { return false })...)

	// **2 — UDP der horchenden Seite gesperrt.** Die Lücke ist das längste
	// Echo danach: die wählende Seite schickt an den toten Port, bis ihr
	// direkter Weg seine Geltung verliert.
	envknob.Setenv("TS_DEBUG_ALWAYS_USE_DERP", "true")
	conn.Rebind()
	conn.ReSTUN("D03: UDP gesperrt")

	gesperrt := abschnitt("UDP gesperrt", 30*time.Second, func(echo) bool { return false })
	luecke := laengstes(gesperrt)
	ueberRelais := median(gesperrt[len(gesperrt)/2:])

	// **Und wirklich über das Relais**: ein Echo über DERP im Internet dauert
	// länger als eines über die Schleife. Hier lag der Irrtum des ersten
	// Entwurfs — 1 ms „über das Relais" war der direkte Weg.
	if ueberRelais < 5*time.Millisecond || ueberRelais < 3*median(direkt) {
		t.Errorf("gesperrt dauert ein Echo im Median %v, direkt %v — das ist kein Weg über ein Relais", ueberRelais, median(direkt))
	}

	if luecke > 30*time.Second {
		t.Errorf("die Lücke war %v — länger als die Frist des Abschnitts", luecke)
	}

	t.Logf("Lücke beim Sperren: %v; danach über das Relais im Median %v (direkt %v)", luecke, ueberRelais, median(direkt))

	// **3 — UDP wieder offen.** Gewartet wird, bis der Stand den direkten Weg
	// nennt **und** ein Echo wieder so schnell ist wie direkt.
	envknob.Setenv("TS_DEBUG_ALWAYS_USE_DERP", "")
	conn.Rebind()
	conn.ReSTUN("D03: UDP offen")

	offen := time.Now()
	schnell := 3 * max(median(direkt), time.Millisecond)
	rueckkehr := abschnitt("UDP offen", 90*time.Second, func(e echo) bool { return e.direkt && e.dauer < schnell })

	if letzter := rueckkehr[len(rueckkehr)-1]; !letzter.direkt || letzter.dauer >= schnell {
		t.Errorf("90 s nach dem Öffnen: direkt=%v, Echo %v — der Fernweg bleibt beim Relais", letzter.direkt, letzter.dauer)
	}

	t.Logf("zurück auf dem direkten Weg nach %v", time.Since(offen).Round(time.Millisecond))
}
