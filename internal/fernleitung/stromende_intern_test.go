package fernleitung

// **Jeder Weg, auf dem ein Strom endet, zählt ab** — angestrichen vom
// Zweitblick über Phase H (#2086): abgezählt wurde nur beim `Close` der
// Hülle. Schließt tailcat den Strom darunter — beim Niederlegen seines
// Netzstapels —, sah der Zähler es nie; und ein Strom, der nach dem
// Niederlegen noch einlief, landete in der Schlange.

import (
	"net"
	"testing"
	"time"
)

// angebundenNach baut einen Horcher über einem Probeserver.
func angebundenNach(t *testing.T) (*Horcher, *probeserver) {
	t.Helper()

	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	return horcher, gebaut
}

func TestEinStromEndetAuchUnterDerHuelle(t *testing.T) {
	for _, fall := range []struct {
		name  string
		enden func(hiesig, dortig net.Conn)
		lesen bool
	}{
		{"das Gegenüber legt auf, gelesen", func(hiesig, _ net.Conn) { _ = hiesig.Close() }, true},           //nolint:errcheck // eine Pipe schließt ohne Fehler
		{"der Strom darunter schließt, gelesen", func(_, dortig net.Conn) { _ = dortig.Close() }, true},      //nolint:errcheck // dito
		{"der Strom darunter schließt, geschrieben", func(_, dortig net.Conn) { _ = dortig.Close() }, false}, //nolint:errcheck // dito
	} {
		t.Run(fall.name, func(t *testing.T) {
			horcher, gebaut := angebundenNach(t)

			hiesig, dortig := net.Pipe()
			gebaut.annehmen(dortig)
			strom := horcher.anbinden(<-horcher.stroeme)

			fall.enden(hiesig, dortig)

			if fall.lesen {
				_, _ = strom.Read(make([]byte, 1)) //nolint:errcheck // das Ende ist der Fall
			} else {
				_, _ = strom.Write([]byte{1}) //nolint:errcheck // dito
			}

			if zahl := horcher.Angebunden(); zahl != 0 {
				t.Errorf("nach dem Ende: %d angebunden, erwartet 0", zahl)
			}

			_ = strom.Close() //nolint:errcheck // ein zweites Ende zählt nicht noch einmal

			if zahl := horcher.Angebunden(); zahl != 0 {
				t.Errorf("nach dem Schließen: %d angebunden, erwartet 0", zahl)
			}
		})
	}
}

// **Eine abgelaufene Frist ist kein Ende** — der Strom lebt, und wer danach
// liest, bekommt Daten.
func TestEineFristIstKeinStromende(t *testing.T) {
	horcher, gebaut := angebundenNach(t)

	hiesig, dortig := net.Pipe()
	t.Cleanup(func() { _ = hiesig.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten
	gebaut.annehmen(dortig)
	strom := horcher.anbinden(<-horcher.stroeme)

	_ = strom.SetReadDeadline(time.Now().Add(-time.Second)) //nolint:errcheck // eine Pipe kennt Fristen
	_, _ = strom.Read(make([]byte, 1))                      //nolint:errcheck // die Frist ist der Fall

	if zahl := horcher.Angebunden(); zahl != 1 {
		t.Errorf("nach der Frist: %d angebunden, erwartet 1", zahl)
	}
}

// **Nach dem Niederlegen wartet kein Strom** — weder einer, der vorher in
// der Schlange stand, noch einer, der danach einläuft.
func TestNachDemSchliessenWartetKeinStrom(t *testing.T) {
	horcher, gebaut := angebundenNach(t)

	for range Schlangenlaenge {
		_, dortig := net.Pipe()
		gebaut.annehmen(dortig)
	}

	if fehler := horcher.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	for range Schlangenlaenge {
		_, dortig := net.Pipe()
		gebaut.annehmen(dortig)
	}

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("nach dem Niederlegen: %d angebunden, erwartet 0", zahl)
	}

	if wartend := len(horcher.stroeme); wartend != 0 {
		t.Errorf("nach dem Niederlegen warten %d Ströme", wartend)
	}
}

// **Wer noch wartet, ist nicht angebunden** — Nachlauf über Phase H: ein
// Strom, der in der Schlange stirbt, ohne dass jemand liest, blieb gezählt,
// bis ihn jemand nahm. Gezählt wird deshalb erst, was [Horcher.Nimm]
// übernimmt.
func TestEinWartenderStromZaehltNicht(t *testing.T) {
	horcher, gebaut := angebundenNach(t)

	hiesig, dortig := net.Pipe()
	gebaut.annehmen(dortig)
	_ = dortig.Close() //nolint:errcheck // der Strom stirbt in der Schlange
	_ = hiesig.Close() //nolint:errcheck // dito

	if zahl := horcher.Angebunden(); zahl != 0 {
		t.Errorf("ein wartender, toter Strom: %d angebunden, erwartet 0", zahl)
	}
}
