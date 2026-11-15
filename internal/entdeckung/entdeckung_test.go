package entdeckung_test

// **Wer in der Nähe ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum dieser Lauf über die Rückschleife geht
//
// **Weil ein Test nicht ins Netz ruft** (§3). mDNS ist Multicast, und ein Lauf
// über `en0` fragte jeden Rechner im WLAN — und fände, was zufällig dort
// steht. `lo0` trägt Multicast (nachgemessen am 2026-09-14:
// `flags=up|loopback|multicast|running`), und was darüber geht, verlässt
// diesen Rechner nicht.
//
// **Was er damit nicht zeigt**: ein zweiter Rechner, ein WLAN mit
// Client-Isolation und die lokale Netzwerkberechtigung von macOS. Das steht in
// `docs/HAERTUNG.md`.

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// rueckschleife gibt die Schnittstelle, über die dieser Lauf spricht.
func rueckschleife(t *testing.T) *net.Interface {
	t.Helper()

	schnitt, fehler := net.InterfaceByName("lo0")
	if fehler != nil {
		t.Skipf("dieser Rechner hat kein lo0: %v", fehler)
	}

	return schnitt
}

// ohneNetzerlaubnisUeberspringen bricht ab, wo der Rechner die Sendung verbot.
//
// # Warum die Antwort des Produkts das Tor ist
//
// **Weil sie schon da war.** `deuteSuchfehler` bildet `EHOSTUNREACH`,
// `ENETUNREACH` und `EPERM` auf [domain.ErrOhneNetzerlaubnis] ab — genau das,
// was macOS ab 15 meldet, wenn ein Programm keine Erlaubnis für das lokale
// Netzwerk hat. Eine eigene Netzprobe daneben wäre eine **zweite** Antwort auf
// dieselbe Frage; läuft sie mit der ersten auseinander, überspringt der Lauf
// aus einem Grund, den es nicht gibt (oder läuft in einen, den es gibt).
//
// # Warum es das Tor braucht
//
// **Weil vier Läufe in drei Paketen daran rot waren** — gemessen am
// 2026-09-20: `sendto: no route to host`, und zwar für IPv4 wie IPv6 und auf
// jeder Schnittstelle. Ein Lauf, der eine Fähigkeit dieses Rechners braucht,
// muss sie erfragen; für `lo0` tut das [rueckschleife], für die Erlaubnis tat
// es niemand.
//
// # Und warum laut
//
// **Weil ein stilles SKIP in diesem Repositorium schon einmal sechs Tage
// gekostet hat.** Die Meldung nennt Ursache und Weg dorthin. Was dieses Tor
// kostet, steht in `docs/HAERTUNG.md`: an einem Tag ohne die Erlaubnis ist die
// Entdeckung im lokalen Netz **nicht** geprüft.
func ohneNetzerlaubnisUeberspringen(t *testing.T, fehler error) {
	t.Helper()

	if errors.Is(fehler, domain.ErrOhneNetzerlaubnis) {
		t.Skipf("dieser Rechner ließ die Suche nicht hinaus (%v) — unter "+
			"macOS ab 15 braucht das laufende Programm die Erlaubnis für das "+
			"lokale Netzwerk (Systemeinstellungen → Datenschutz & Sicherheit "+
			"→ Lokales Netzwerk). Was dieser Lauf sagen soll, ist damit heute "+
			"nicht geprüft", fehler)
	}
}

// eineKennung gibt eine echte Peerkennung.
func eineKennung(t *testing.T) peer.Kennung {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich.Kennung()
}

// **Wer sich meldet, wird gefunden.**
func TestWerSichMeldetWirdGefunden(t *testing.T) {
	kennung := eineKennung(t)
	schnitt := rueckschleife(t)

	anmeldung, fehler := entdeckung.Melde(entdeckung.Angebot{
		Kennung:       kennung,
		Name:          "Davids Rechner",
		Port:          41234,
		Schnittstelle: schnitt,
	})
	if fehler != nil {
		t.Fatalf("Melde: %v", fehler)
	}

	t.Cleanup(func() {
		_ = anmeldung.Beende() //nolint:errcheck // nur ein Lauf
	})

	ergebnis, fehler := entdeckung.Suche(t.Context(), entdeckung.Frage{
		Frist:         2 * time.Second,
		Schnittstelle: schnitt,
	})

	ohneNetzerlaubnisUeberspringen(t, fehler)

	if fehler != nil {
		t.Fatalf("Suche: %v", fehler)
	}

	var gefunden *entdeckung.Fund

	for i := range ergebnis.Funde {
		if ergebnis.Funde[i].Kennung == kennung {
			gefunden = &ergebnis.Funde[i]
		}
	}

	if gefunden == nil {
		t.Fatalf("nicht gefunden — es kamen %+v", ergebnis)
	}

	if gefunden.Name != "Davids Rechner" {
		t.Errorf("der Name ist %q", gefunden.Name)
	}

	if gefunden.Port != 41234 {
		t.Errorf("der Port ist %d", gefunden.Port)
	}
}

// **Wer nicht mehr meldet, wird nicht mehr gefunden.**
//
// **Das ist die Zusage „Peer verschwunden"** aus dem Issue: eine Entdeckung,
// die einen Rechner weiterführte, der längst weg ist, schickte jemanden an
// eine Adresse, an der niemand sitzt.
func TestWerNichtMehrMeldetWirdNichtMehrGefunden(t *testing.T) {
	kennung := eineKennung(t)
	schnitt := rueckschleife(t)

	anmeldung, fehler := entdeckung.Melde(entdeckung.Angebot{
		Kennung: kennung, Name: "Kurz da", Port: 41235,
		Schnittstelle: schnitt,
	})
	if fehler != nil {
		t.Fatalf("Melde: %v", fehler)
	}

	if befund := anmeldung.Beende(); befund != nil {
		t.Fatalf("Beende: %v", befund)
	}

	ergebnis, fehler := entdeckung.Suche(t.Context(), entdeckung.Frage{
		Frist: time.Second, Schnittstelle: schnitt,
	})

	ohneNetzerlaubnisUeberspringen(t, fehler)

	if fehler != nil {
		t.Fatalf("Suche: %v", fehler)
	}

	for _, fund := range ergebnis.Funde {
		if fund.Kennung == kennung {
			t.Fatalf("der Beendete steht noch da: %+v", fund)
		}
	}
}

// **Eine Anmeldung ohne Kennung gibt es nicht.**
func TestEineAnmeldungOhneKennungGibtEsNicht(t *testing.T) {
	t.Parallel()

	if _, fehler := entdeckung.Melde(entdeckung.Angebot{
		Port: 41234,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und eine ohne Port ebenso.**
//
// **Anders als beim Horchen**: dort heißt `0` „such mir einen aus", hier wäre
// es eine Ankündigung, die niemanden irgendwohin schickt.
func TestEineAnmeldungOhnePortGibtEsNicht(t *testing.T) {
	t.Parallel()

	for name, port := range map[string]int{
		"null":    0,
		"negativ": -1,
		"zu groß": 65536,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := entdeckung.Melde(entdeckung.Angebot{
				Kennung: eineKennung(t), Port: port,
			}); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Eine Suche ohne Frist sucht trotzdem nicht ewig.**
//
// **„Discovery bleibt kein unbegrenzter Spinner"** steht in der Abnahme des
// Issues, und eine Vorgabe ist die Stelle, an der das eingelöst wird.
func TestEineSucheOhneFristSuchtNichtEwig(t *testing.T) {
	t.Parallel()

	angefangen := time.Now()

	_, fehler := entdeckung.Suche(t.Context(), entdeckung.Frage{
		Schnittstelle: rueckschleife(t),
	})

	// **Auch hier, und das ist kein Formalismus**: ohne Erlaubnis gibt die
	// Suche **sofort** zurück — der Lauf wäre grün und hätte über die Frist
	// nichts gesagt. Ein Tor, das hier fehlte, machte aus einem roten Lauf
	// einen grünen ohne Aussage.
	ohneNetzerlaubnisUeberspringen(t, fehler)

	if fehler != nil {
		t.Fatalf("Suche: %v", fehler)
	}

	// **Frist plus Nachlauf**, und die zweite Zahl ist der Grund, warum es
	// diesen Lauf gibt: `QueryParam.Timeout` begrenzt nur das **Zuhören**.
	// Der erste Anlauf verlangte `< Hoechstfrist` und maß 6,0 Sekunden bei
	// fünf Sekunden Frist — ein Kontext bindet seither die Wanduhr.
	grenze := entdeckung.Hoechstfrist + entdeckung.Nachlauf
	if gedauert := time.Since(angefangen); gedauert > grenze {
		t.Fatalf("die Suche dauerte %s, erlaubt sind %s", gedauert, grenze)
	}
}
