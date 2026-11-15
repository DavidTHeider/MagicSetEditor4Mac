package domain_test

// **Wer wann handeln darf** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
//   - CR 117.3a — der aktive Spieler bekommt sie zu Beginn der meisten
//     Abschnitte, und im Enttappen bekommt sie **niemand**.
//   - CR 117.3d — wer passt, gibt sie an den Nächsten in Zugreihenfolge.
//   - CR 117.4 und CR 500.2 — passen **alle nacheinander**, endet der
//     Abschnitt; ein einzelnes Passen tut es nicht.
//   - CR 102.1 — der aktive Spieler ist der, dessen Zug es ist.
//
// **Und der Sonderfall mit einem Sitz** (#1460, Solo Goldfish): „alle
// nacheinander" ist dort **ein** Durchlauf. Ein Abschnitt, der auf einen
// zweiten Passenden wartete, den es nicht gibt, hinge.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Zug fängt an — und bleibt im Enttappen nicht stehen.**
//
// # Warum er dort nicht steht
//
// **Weil ihn nichts weiterbrächte.** Ein Abschnitt endet, wenn alle
// nacheinander passen (CR 500.2) — im Enttappen bekommt aber niemand
// Priorität (CR 502.4), also passt auch niemand. Er tut, was er zu tun hat
// (CR 502.3), und gibt weiter; der erste Halt ist die Versorgung
// (CR 503.1: der Versorgungsschritt hat keine zugbasierte Handlung, und
// der aktive Spieler bekommt sie, sobald er beginnt).
func TestDerZugFaengtAnUndHaeltInDerVersorgung(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	switch {
	case partie.Zug.Nummer != 1:
		t.Errorf("der Zug ist der %d.", partie.Zug.Nummer)
	case partie.Zug.Aktiv != "sitz-1":
		t.Errorf("aktiv ist %q", partie.Zug.Aktiv)
	case partie.Zug.Starter != "sitz-1":
		t.Errorf("angefangen hat %q", partie.Zug.Starter)
	case partie.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
		t.Errorf("es steht bei %+v", partie.Zug.Abschnitt)
	case partie.Zug.Prioritaet != "sitz-1":
		t.Errorf("Priorität hat %q", partie.Zug.Prioritaet)
	}
}

// **Im Versorgungsschritt hat der aktive Spieler sie** — CR 117.3a.
func TestDerAktiveSpielerBekommtSieZuerst(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-2")

	switch {
	case partie.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
		t.Fatalf("es steht bei %+v", partie.Zug.Abschnitt)
	case partie.Zug.Prioritaet != "sitz-2":
		t.Errorf("Priorität hat %q, erwartet sitz-2", partie.Zug.Prioritaet)
	}
}

// **Ein einzelnes Passen gibt sie weiter und beendet nichts** — CR 117.3d.
func TestEinPassenGibtSieWeiter(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	gepasst, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("Passen: %v", fehler)
	}

	switch {
	case gepasst.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
		t.Errorf("der Abschnitt endete schon: %+v", gepasst.Zug.Abschnitt)
	case gepasst.Zug.Prioritaet != "sitz-2":
		t.Errorf("Priorität hat %q, erwartet sitz-2", gepasst.Zug.Prioritaet)
	}
}

// **Passen alle nacheinander, endet der Abschnitt** — CR 117.4, CR 500.2.
func TestAlleNacheinanderBeendenDenAbschnitt(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	for _, wer := range []string{"sitz-1", "sitz-2"} {
		gepasst, fehler := partie.Passen(wer)
		if fehler != nil {
			t.Fatalf("Passen (%s): %v", wer, fehler)
		}

		partie = gepasst
	}

	switch {
	case partie.Zug.Abschnitt.Schritt != domain.SchrittZiehen:
		t.Errorf("es steht bei %+v, erwartet den Ziehschritt", partie.Zug.Abschnitt)
	case partie.Zug.Prioritaet != "sitz-1":
		t.Errorf("im neuen Abschnitt hat %q Priorität", partie.Zug.Prioritaet)
	case len(partie.Zug.Gepasst) != 0:
		t.Errorf("es gelten noch Passen von vorhin: %v", partie.Zug.Gepasst)
	}
}

// **Wer nicht an der Reihe ist, passt nicht.**
//
// **Und das ist kein Formfehler**: ein Passen des Falschen verschöbe die
// Reihenfolge, und danach wäre „alle nacheinander" eine andere Aussage.
func TestWerNichtDranIstPasstNicht(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	if _, fehler := partie.Passen("sitz-2"); fehler == nil {
		t.Error("ein Passen außer der Reihe kam durch")
	}

	// **Und wo niemand sie hat, passt auch niemand** — CR 502.4, CR 514.3.
	//
	// **Von Hand gestellt und nicht erspielt**: der Ablauf hält in keinem
	// dieser beiden Abschnitte an (siehe [domain.Partie.Weiter]), und genau
	// deshalb muss der Ausgang trotzdem einen haben — ein Zustand aus einem
	// Bestand kann dort stehen (§8).
	ohne := einePartie().ZugBeginnen("sitz-1")
	ohne.Zug.Abschnitt = domain.Abschnitt{
		Phase: domain.PhaseAnfang, Schritt: domain.SchrittEnttappen,
	}
	ohne.Zug.Prioritaet = ""

	if _, fehler := ohne.Passen("sitz-1"); fehler == nil {
		t.Error("im Enttappen wurde gepasst")
	}
}

// **Bei einem Sitz ist „alle nacheinander" ein Durchlauf.**
//
// **Solo Goldfish ist seit G3 erlaubt** (#1460, CR 100.1 dazu: die Regeln
// gelten für zwei oder mehr — ein Sitz ist ein Aufbau und kein Spiel). Ein
// Abschnitt, der auf einen zweiten Passenden wartete, hinge für immer.
func TestBeiEinemSitzReichtEinPassen(t *testing.T) {
	t.Parallel()

	allein := einePartie()
	allein.Sitze = allein.Sitze[:1]
	allein.Wirt = "sitz-1"

	partie := allein.ZugBeginnen("sitz-1")

	gepasst, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("Passen: %v", fehler)
	}

	if gepasst.Zug.Abschnitt.Schritt != domain.SchrittZiehen {
		t.Errorf("es steht bei %+v", gepasst.Zug.Abschnitt)
	}
}

// **Nach dem Aufräumen ist der Nächste am Zug** — CR 102.1.
func TestNachDemAufraeumenIstDerNaechsteDran(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	partie.Zug.Abschnitt = domain.Abschnitt{
		Phase: domain.PhaseEnde, Schritt: domain.SchrittAufraeumen,
	}

	weiter := partie.Weiter()

	// **Und der neue Zug hält wieder erst in der Versorgung** — das
	// Enttappen dazwischen ist gelaufen und hat nicht gewartet (CR 502.4).
	switch {
	case weiter.Zug.Nummer != 2:
		t.Errorf("es ist Zug %d", weiter.Zug.Nummer)
	case weiter.Zug.Aktiv != "sitz-2":
		t.Errorf("aktiv ist %q", weiter.Zug.Aktiv)
	case weiter.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
		t.Errorf("es steht bei %+v", weiter.Zug.Abschnitt)
	}
}

// **Ein Zustand, den jemand von Hand geändert hat, hält die Partie nicht an.**
//
// # Warum das zwei Läufe in einem sind
//
// **Weil beide Ausgänge derselbe Satz aus §8 sind**: fremde Daten sind der
// Normalfall, und ein Bestand, den jemand mit einem anderen Werkzeug
// geöffnet hat, darf eine Partie unbenutzbar machen — er tut es hier nicht.
//
//   - **Ohne Sitze** gibt es niemanden, der als Nächster drankäme.
//   - **Eine Kennung, die keinem Sitz gehört**, bekommt den ersten — sonst
//     stünde die Partie bei einem Namen, den es nicht gibt.
func TestEinKrummerZugstandHaeltDiePartieNichtAn(t *testing.T) {
	t.Parallel()

	ohneSitze := einePartie()
	ohneSitze.Sitze = nil
	ohneSitze.Zug = domain.Zugstand{
		Nummer: 1, Aktiv: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseEnde, Schritt: domain.SchrittAufraeumen,
		},
	}

	if aktiv := ohneSitze.Weiter().Zug.Aktiv; aktiv != "" {
		t.Errorf("ohne Sitze ist %q aktiv", aktiv)
	}

	fremd := einePartie()
	fremd.Zug = domain.Zugstand{
		Nummer: 1, Aktiv: "gibtsNicht",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseEnde, Schritt: domain.SchrittAufraeumen,
		},
	}

	if aktiv := fremd.Weiter().Zug.Aktiv; aktiv != "sitz-1" {
		t.Errorf("nach einer fremden Kennung ist %q aktiv", aktiv)
	}
}
