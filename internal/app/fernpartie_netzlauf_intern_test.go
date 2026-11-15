//go:build netzlauf

package app

// **Eine Partie über fremde Netze, mit verdeckter Hand** — D04
// (#2235).
//
// Zwei Anwendungen mit eigenem Bestand und eigener Identität; der Wirt horcht
// über **tailcat** (`FernHorchStarten`), der Gast trägt die **Einladung** als
// Adresse ein und tritt bei. Darüber derselbe Weg wie im Betrieb: Handschlag,
// Partiesitzung, Tischfunk. Gemessen wird wie in D01 an der ganzen Sicht —
// das erste Tischbild beim Gast, das der Sicht gleicht, die der Wirt für
// Berts Sitz rechnet, binnen der Frist von 10 s, und kein geheimes Objekt
// verrät beim Gast Name oder Karte.
//
// **Zweimal**: so, wie tailcat wählt, und nur über das Relais
// (`TS_DEBUG_ALWAYS_USE_DERP`, vor dem Start gesetzt). Beide Male nennt der
// Fernstand des Wirts, welchen Weg die Verbindung nimmt.
//
// # Was er nicht belegt
//
// **Zwei NAT-Grenzen** — beide Anwendungen stehen auf diesem Mac, der direkte
// Weg ist die Schleife (`out-of-scope`, keine zwei Rechner).
//
//	go test -tags netzlauf ./internal/app/ -run FremdeNetze -v -count=1

import (
	"testing"
	"time"

	"tailscale.com/envknob"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einTischUeberFremdeNetze setzt den Gast über die Einladung an den Tisch —
// oder, mit `ueberTCP`, über eine Adresse im eigenen Netz: die Vergleichszeile.
func einTischUeberFremdeNetze(t *testing.T, ueberTCP bool) *tischUeberRelais {
	t.Helper()

	wirt, gast := zweiUeberTCP(t)
	sicht := &gastsicht{}
	gast.anwendung.melden = sicht.melden

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, bertsKarte := einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(wirt.ctx, partieID, "anna", []string{"bert"}); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	var adresse string

	if ueberTCP {
		gestartet := wirt.anwendung.HorchStarten(wirt.ctx, "127.0.0.1:0")
		if gestartet.Fehler != nil {
			t.Fatalf("HorchStarten: %+v", gestartet.Fehler)
		}

		t.Cleanup(wirt.anwendung.horchenBeenden)

		adresse = gestartet.Stand.Adresse
	} else {
		gestartet := wirt.anwendung.FernHorchStarten(wirt.ctx)
		if gestartet.Fehler != nil {
			t.Fatalf("FernHorchStarten: %+v", gestartet.Fehler)
		}

		t.Cleanup(wirt.anwendung.fernhorchenBeenden)

		adresse = gestartet.Stand.Einladung
	}

	if antwort := gast.anwendung.PeerkontaktAdresse(gast.ctx, string(wirt.kennung), adresse); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	tisch := &tischUeberRelais{
		wirt: wirt, gast: gast, sicht: sicht, partieID: partieID,
		stellen: ueberDenWirt(wirt), bertsKarte: bertsKarte,
	}

	beginn := time.Now()
	tisch.zurueck(t)
	tisch.konvergiert(t, "der Beitritt", beginn, func(PartieDTO) bool { return true })

	return tisch
}

func TestEinePartieGehtUeberFremdeNetze(t *testing.T) {
	for _, weg := range []struct {
		name      string
		ueberTCP  bool
		nurRelais bool
		erwartet  string
	}{
		{"über TCP im eigenen Netz", true, false, ""},
		{"wie tailcat wählt", false, false, "direkt"},
		{"nur über das Relais", false, true, "relais"},
	} {
		t.Run(weg.name, func(t *testing.T) {
			if !weg.ueberTCP {
				// **tailcat legt seinen Zustand unter das Zuhause.**
				t.Setenv("HOME", t.TempDir())
			}

			if weg.nurRelais {
				envknob.Setenv("TS_DEBUG_ALWAYS_USE_DERP", "true")
				t.Cleanup(func() { envknob.Setenv("TS_DEBUG_ALWAYS_USE_DERP", "") })
			}

			tisch := einTischUeberFremdeNetze(t, weg.ueberTCP)

			// **Anna zieht — verdeckt.** Die Karte geht aus ihrer Bibliothek
			// auf ihre Hand; beim Gast steht sie danach als geheimes Objekt
			// ohne Namen (geprüft in `konvergiert`).
			karte := inDerBibliothek(t, tisch.wirtssicht(t).Objekte, "anna")
			beginn := time.Now()

			befehl(t, tisch.stellen, tisch.partieID, "fern/ziehen", "objekt.schieben",
				map[string]any{"objekt": karte, "zone": "hand"})

			tisch.konvergiert(t, "ein verdeckter Zug über "+weg.name, beginn, func(sicht PartieDTO) bool {
				gezogen, da := objektIn(&sicht, karte)

				return da && gezogen.Zone == "hand" && gezogen.Geheim
			})

			// **Und Bert spielt zurück**, über den Gast und den Wirt.
			beginn = time.Now()

			if fehler := tisch.gast.anwendung.PartieWuenschen(tisch.gast.ctx, domain.Spielbefehl{
				Kennung: "fern/bert", Art: "objekt.schieben",
				Nutzlast: map[string]any{"objekt": tisch.bertsKarte, "zone": "hand"},
			}).Fehler; fehler != nil {
				t.Fatalf("PartieWuenschen: %+v", fehler)
			}

			tisch.konvergiert(t, "ein Zug des Gastes über "+weg.name, beginn, func(sicht PartieDTO) bool {
				gespielt, da := objektIn(&sicht, tisch.bertsKarte)

				return da && gespielt.Zone == "hand"
			})

			if weg.ueberTCP {
				return
			}

			// **Der Weg, den der Wirt nennt.** Direkt entsteht er erst unter
			// Verkehr — deshalb wird gewartet und nicht einmal gefragt.
			for frist := time.Now().Add(20 * time.Second); ; time.Sleep(250 * time.Millisecond) {
				wege := tisch.wirt.anwendung.FernHorchStand(tisch.wirt.ctx).Stand.Wege
				if len(wege) > 0 && wege[0].Art == weg.erwartet {
					t.Logf("der Wirt nennt %s: %+v", weg.erwartet, wege)

					break
				}

				if time.Now().After(frist) {
					t.Fatalf("nach 20 s nennt der Wirt %+v, erwartet zuerst %q", wege, weg.erwartet)
				}
			}
		})
	}
}
