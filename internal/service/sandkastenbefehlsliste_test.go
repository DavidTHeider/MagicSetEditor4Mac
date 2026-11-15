package service_test

// **Die Liste der Sandkastenbefehle** — der Wächter über ihre Zahl.
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was Blick und Mischung tun**; hier steht, **wie viele
// Befehle es überhaupt gibt und dass sie alle durch einen Weg gehen**. Das
// ist eine Frage an die Liste und nicht an eine Zone — und
// `sandkastenzonen_test.go` stand mit ihr bei 403 Zeilen (§2).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Und beide stehen in der Liste der erlaubten Befehle.**
//
// **Fünfzehn seit M17** (#1592)
// — passen, abwerfen, Leben, Vergiftungsmarken und markierter Schaden kamen
// dazu. Sie gehen durch denselben Weg und stehen damit an allen vier Flächen
// zugleich da (§16).
//
// **Siebzehn seit M1** (#1603):
// die **Abkürzung** (CR 732) kam dazu — sie spart zwanzig Passen je Zug — und
// das **vorzeitige Zugende** (CR 724).
//
// **Neunzehn seit M3** (#1605):
// Mana **nachtragen** (CR 106.4) und ein **Grundland tappen** (CR 305.6).
// Die zweite ist die einzige Manafähigkeit, die im Regelwerk steht statt auf
// der Karte — ohne sie gäbe es in einer Partie kein Mana.
//
// **Einundzwanzig seit M4** (#1606):
// einen Zauber **wirken** (CR 601.2) und ein **Land spielen** (CR 116.2a) —
// die einzige der zwölf Sonderhandlungen, die ohne Kartentext geht.
//
// **Dreiundzwanzig seit M6** (#1608):
// **Angreifer** und **Blocker** erklären (CR 508.1, CR 509.1). Der
// **Kampfschaden** steht ausdrücklich **nicht** dabei: er ist eine
// zugbasierte Handlung (CR 703.4m) und geschieht von selbst — ein Befehl
// dafür wäre eine Wahl, wo die Regeln keine lassen.
//
// **Einunddreißig seit M7** (#1609):
// **acht Handlungen mit Schlüsselwort** (CR 701) — zerstören, opfern,
// exilieren, neutralisieren, kämpfen, mahlen, hellsehen, überwachen. Es sind
// die acht von neunundsechzig, die das **Regelwerk allein** beschreibt; die
// übrigen einundsechzig sind Bausteine eines Kartentextes und brauchen `M8`.
//
// **Dreiunddreißig seit M15** (#1617):
// **Münzwurf** (CR 705) und **Würfel** (CR 706). Beide ändern nichts am
// Tisch — sie hinterlassen ein Ergebnis im Strom, und der hängt an der
// Hashkette.
//
// **Siebenunddreißig seit M12** (#1614):
// ein **Spielstein** (CR 701.7a), ein **Emblem** (CR 114.1), ein
// **Kontrollwechsel** (CR 110.2) und das **Umwandeln** einer zweiseitigen
// Karte (CR 701.27a). Was die vier eint, ist nicht eine Regelnummer, sondern
// die Frage: welche Gestalt hat das, was auf dem Tisch liegt?
//
// **Und der achtunddreißigste ist die Legendenwahl** (CR 704.5j) — der
// einzige Befehl, der eine **zustandsbasierte Aktion abschließt**. Die
// übrigen vierundzwanzig aus CR 704.5 geschehen von selbst; diese eine
// verlangt eine Wahl, und M5 führte sie deshalb als Grenze.
//
// **Vierundvierzig seit M9** (#1611):
// die **Reihung** (CR 603.3b). Sie ist die zweite ihrer Art — eine Wahl, die
// das Regelwerk verlangt, und keine Handlung: beherrscht ein Sitz mehrere
// Fähigkeiten, die zugleich ausgelöst haben, nennt er die Reihenfolge, in
// der sie auf den Stapel kommen.
//
// **Fünfzig seit** #1839:
// der **Beginn** (CR 103.5) — Starthand, Mulligan, Unterlegen und Behalten.
// Bis dahin lag jede Bibliothek nach dem Anfangen in der Reihenfolge der
// Deckliste, und niemand hatte eine Hand.
//
// **Einundfünfzig seit P3** (#2371):
// die **Geste** (`objekt.ziehen`) — Objekt, Zone und Ziel-Sitz, und der Kern
// deutet, was sie meint.
func TestDerSandkastenNenntEinundfuenfzigBefehle(t *testing.T) {
	t.Parallel()

	befehle := service.Sandkastenbefehle()

	if len(befehle) != 51 {
		t.Fatalf("%d Befehle, erwartet einundfünfzig: %v",
			len(befehle), befehle)
	}

	gesucht := map[string]bool{
		service.BefehlBlick: false, service.BefehlMischen: false,

		// **Häufen und Enthäufen** — CR 400.5, eine Anordnung und keine
		// Regelbewegung
		// (#1756).
		service.BefehlHaeufen: false, service.BefehlEnthaeufen: false,

		// **Der Beginn** — CR 103.5 (#1839).
		service.BefehlStarthand: false, service.BefehlMulligan: false,
		service.BefehlUnterlegen: false, service.BefehlBehalten: false,
	}

	for _, befehl := range befehle {
		if _, dabei := gesucht[befehl]; dabei {
			gesucht[befehl] = true
		}
	}

	for befehl, dabei := range gesucht {
		if !dabei {
			t.Errorf("%q steht nicht in Sandkastenbefehle()", befehl)
		}
	}
}
