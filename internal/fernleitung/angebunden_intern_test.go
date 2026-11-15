package fernleitung

// **Der Horcher zählt, wer angebunden ist** — SEC-21
// (#2162).
//
// tailcat v0.6.0 nannte in `Server.Status()` **keine** Gegenstelle: sein
// `locoBackend.Status` baute den `ipnstate.StatusBuilder` ohne `WantPeers`,
// die Liste der Wege blieb auch bei offener Verbindung leer, und die
// Oberfläche sagte „noch hat niemand diese Einladung benutzt". Seit v0.7.0
// nennt es Gegenstellen (K01, #2187) — aber Gegenstellen sind keine Ströme:
// ob gerade jemand angebunden ist, weiß weiter nur, wer die Ströme selbst
// zählt.

import (
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// **Ein Strom zählt, solange er offen ist** — von der Übernahme durch `Nimm`
// bis zum Schließen, auf welcher Seite auch immer es geschieht; wer noch in der
// Schlange wartet, zählt nicht (Nachlauf über Phase H).
func TestDerHorcherZaehltOffeneStroeme(t *testing.T) {
	dort := eineIdentitaet(t)
	hier := eineIdentitaet(t)

	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(dort)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Fatalf("vor dem ersten Strom: %d angebunden", zahl)
	}

	hiesig, dortig := net.Pipe()
	gebaut.annehmen(dortig)

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("in der Schlange: %d angebunden, erwartet 0", zahl)
	}

	genommen := make(chan *leitung.Leitung, 1)

	go func() {
		offen, _ := horcher.Nimm(t.Context()) //nolint:errcheck // geprüft wird unten über das Ergebnis
		genommen <- offen
	}()

	if _, fehler := leitung.UeberStrom(t.Context(), hiesig, hier, dort.Kennung()); fehler != nil {
		t.Fatalf("UeberStrom: %v", fehler)
	}

	offen := <-genommen
	if offen == nil {
		t.Fatal("der Horcher nahm die Leitung nicht")
	}

	if zahl := horcher.Angebunden(); zahl != 1 {
		t.Errorf("nach dem Handschlag: %d angebunden, erwartet 1", zahl)
	}

	// **Das Gegenüber legt auf, dann schließt diese Seite** — über eine Pipe
	// wartete ein `close_notify` sonst auf einen Leser, den es nicht gibt.
	_ = hiesig.Close()    //nolint:errcheck // eine Pipe schließt ohne Fehler
	_ = offen.Schliesse() //nolint:errcheck // das Gegenüber ist fort; gezählt wird trotzdem ab

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("nach dem Schließen: %d angebunden, erwartet 0", zahl)
	}
}

// **Ein Handschlag, der scheitert, bindet niemanden an** — der Strom wird
// geschlossen, und die Zahl fällt zurück.
func TestEinGescheiterterHandschlagZaehltNicht(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	hiesig, dortig := net.Pipe()
	gebaut.annehmen(dortig)

	// **Das Gegenüber legt auf, statt TLS zu sprechen.**
	_ = hiesig.Close() //nolint:errcheck // eine Pipe schließt ohne Fehler

	if _, fehler := horcher.Nimm(t.Context()); fehler == nil {
		t.Fatal("ein Handschlag ohne Gegenüber ging durch")
	}

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("nach dem gescheiterten Handschlag: %d angebunden", zahl)
	}
}

// **Ein überzähliger Strom wird nie angebunden** — die volle Schlange
// schließt ihn sofort; wer in ihr wartet, zählt erst, wenn `Nimm` ihn
// übernimmt.
func TestEinUeberzaehligerStromZaehltNicht(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	var letzter net.Conn
	for range Schlangenlaenge + 1 {
		hiesig, dortig := net.Pipe()
		gebaut.annehmen(dortig)
		letzter = hiesig
	}

	if _, lesefehler := letzter.Read(make([]byte, 1)); lesefehler == nil {
		t.Error("der überzählige Strom blieb offen")
	}

	if wartend := len(horcher.stroeme); wartend != Schlangenlaenge {
		t.Errorf("%d warten, erwartet %d", wartend, Schlangenlaenge)
	}

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("%d angebunden, erwartet 0", zahl)
	}
}

// **Ein zweimal geschlossener Strom zählt einmal ab** — eine Leitung schließt
// ihren Strom, und ein Aufräumen danach darf die Zahl nicht unter null ziehen.
func TestEinZweimalGeschlossenerStromZaehltEinmalAb(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	for range 2 {
		_, dortig := net.Pipe()
		gebaut.annehmen(dortig)
	}

	strom := horcher.anbinden(<-horcher.stroeme)
	_ = strom.Close() //nolint:errcheck // eine Pipe schließt ohne Fehler
	_ = strom.Close() //nolint:errcheck // dito

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("%d angebunden, erwartet 0", zahl)
	}
}
