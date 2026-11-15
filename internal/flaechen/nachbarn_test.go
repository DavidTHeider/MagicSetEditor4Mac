package flaechen_test

// **Wer in der Nähe ist — über jede Fläche** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass dieselbe Frage überall dieselbe Antwort bekommt** (§16). Eine
// Nachbarschaft, die an der Kommandozeile jemanden zeigte und im Dienst nicht,
// wäre zwei verschiedene Auskünfte über dasselbe Netz.
//
// # Warum er über die Rückschleife geht
//
// **Weil ein Lauf nicht ins Netz ruft** (§3). mDNS ist Multicast; über `en0`
// fragte dieser Lauf jeden Rechner im WLAN und fände, was zufällig dort steht.
// `MSE_ENTDECKUNGSSCHNITTSTELLE=lo0` hält ihn auf diesem Rechner.
//
// **Was er damit nicht zeigt**: ein zweiter Rechner, ein WLAN mit
// Client-Isolation und die lokale Netzwerkberechtigung von macOS. Das steht in
// `docs/HAERTUNG.md`.

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// ohneNetzerlaubnisSatz ist der Satz, mit dem der Kern eine verweigerte
// Multicastsendung meldet.
//
// **Ein Stück des Satzes und nicht der ganze**: davor steht „Validierung
// fehlgeschlagen", dahinter der Vorgang — beides wechselt.
const ohneNetzerlaubnisSatz = "ins lokale Netz ging nichts hinaus"

// ohneNetzrouteSatz ist die Deutung von `ENETUNREACH` (NET-02.NAT-03,
// #1706) — auch sie heißt: dieser Rechner lässt die Suche nicht hinaus.
const ohneNetzrouteSatz = "für das lokale Netz gibt es keine Route"

func TestDieNachbarschaftUeberJedeFlaeche(t *testing.T) {
	if _, fehler := net.InterfaceByName("lo0"); fehler != nil {
		t.Skipf("dieser Rechner hat kein lo0: %v", fehler)
	}

	// **Und ein eigener Schlüsselbund** — dasselbe wie bei der Peeridentität
	// daneben, und aus demselben Grund: `mse nachbarn` braucht den Ausweis
	// dieses Profils, und ohne `MSE_SCHLUESSELBUND` holte er ihn aus dem
	// **Anmeldebund dieses Rechners**.
	//
	// **Hier stand es zuerst nicht**, und der Preis war sichtbar: macOS
	// fragte am 2026-09-14 mitten im Lauf nach der Erlaubnis für
	// „peeridentitaet". Ein Lauf, der einen Menschen um eine Entscheidung
	// bittet, hat die Grenze aus §3 schon überschritten — gleich wie er
	// ausgeht.
	umgebung := append(
		mitEigenerSaat(t),
		app.Entdeckungsschnittstelle+"=lo0",
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	// **Beide Flächen antworten, und beide antworten in derselben Gestalt.**
	//
	// **Einzeln und nicht über den Dreiervergleich**: die Kommandozeile
	// schreibt bei einer leeren Liste **nichts**, und ein gemeinsamer Kern
	// wäre die leere Zeichenkette.
	//
	// # Warum hier nicht „niemand ist da" steht
	//
	// **Weil `go test` die Pakete nebeneinander fährt und über `lo0` jeder
	// jeden hört.** `internal/app` und `internal/entdeckung` kündigen in
	// ihren eigenen Läufen auf derselben Schnittstelle an. Trifft eine ihrer
	// Ankündigungen in das Zeitfenster dieses Laufs, steht sie hier als
	// **Fund** — mit gültiger Kennung, denn `hashicorp/mdns` sendet nur
	// vollständige Einträge.
	//
	// **Hier stand zuerst „die Kommandozeile fand jemanden" als Fehler**,
	// und darunter die Erklärung, eine Antwort könne **vor ihrem TXT-Satz**
	// eintreffen. **Die Erklärung war falsch**: `ServiceEntry.complete()`
	// verlangt `hasTXT`, nachgelesen in `client.go` (v1.0.7, Z. 38). Das
	// gezählte `verworfen 1` vom 2026-09-14 kam aus einem **Rennen** —
	// `query` schrieb `InfoFields` neu, während `deute` es las. Das ist
	// behoben; die Abhängigkeit vom Netz dieses Rechners bleibt.
	//
	// **Also prüft dieser Lauf, was er wirklich prüfen kann** (§10): dass
	// die Fähigkeit an beiden Flächen verdrahtet ist, antwortet und eine
	// **Liste** gibt — nicht, was in dieser Liste steht. Was dort steht,
	// prüft `internal/app` über die echten Nähte, und dort ist der eigene
	// Widerhall die Zusage.
	// **Und `lo0` allein genügt nicht** — Tor vom 2026-09-20.
	//
	// Die Schnittstelle trägt Multicast; ob der **Prozess** darüber senden
	// darf, ist eine zweite Frage. Unter macOS ab 15 braucht er dafür die
	// Erlaubnis für das lokale Netzwerk, und ohne sie weist der Kernel jede
	// Sendung ab — für IPv4 wie IPv6 und auf jeder Schnittstelle,
	// nachgemessen an diesem Tag. Dieser Lauf war daran rot, ohne dass etwas
	// am Code falsch gewesen wäre.
	//
	// **Gefragt wird die Antwort des Produkts und keine eigene Netzprobe**:
	// `entdeckung.deuteSuchfehler` deutet die drei Ursachen bereits als
	// fehlende Erlaubnis, und eine zweite Antwort daneben liefe mit der
	// ersten auseinander. Hier kommt sie als **Text** zurück, weil eine
	// Kommandozeile Text gibt — den Schlüssel trägt sie nicht.
	//
	// **Laut übersprungen**, und was das kostet, steht in `docs/HAERTUNG.md`.
	aus, fehlerstrom, code := kommandozeile(t, umgebung, "nachbarn")
	if code != 0 && (strings.Contains(fehlerstrom+aus, ohneNetzerlaubnisSatz) ||
		strings.Contains(fehlerstrom+aus, ohneNetzrouteSatz)) {
		t.Skipf("dieser Rechner ließ die Suche nicht hinaus — unter macOS ab "+
			"15 braucht das laufende Programm die Erlaubnis für das lokale "+
			"Netzwerk (Systemeinstellungen → Datenschutz & Sicherheit → "+
			"Lokales Netzwerk): %s%s", aus, fehlerstrom)
	}

	if code != 0 {
		t.Fatalf("mse nachbarn: %d %s%s", code, aus, fehlerstrom)
	}

	for _, zeile := range strings.Split(strings.TrimSpace(aus), "\n") {
		if zeile == "" || strings.HasPrefix(zeile, "verworfen\t") {
			continue
		}

		// Ein Fund trägt Kennung, Name und Adresse, getrennt durch Tabulator.
		if teile := strings.Split(zeile, "\t"); len(teile) < 2 {
			t.Errorf("die Kommandozeile gab %q — das ist kein Fund", zeile)
		}
	}

	// **Und der Dienst gibt eine Liste und kein `null`** — sonst verschwände
	// sie aus der Antwort, und wer sie liest, hielte „niemand da" für „keine
	// Angabe".
	aus = mussREST(t, an, http.MethodGet, "/peer/nachbarn", "")

	var antwort struct {
		Nachbarn []map[string]any `json:"nachbarn"`
	}

	if fehler := json.Unmarshal([]byte(aus), &antwort); fehler != nil {
		t.Fatalf("der Dienst gab kein JSON: %v — %s", fehler, aus)
	}

	if antwort.Nachbarn == nil {
		t.Errorf("der Dienst gab `null` statt einer Liste: %s", aus)
	}

	// **Dass es kein MCP-Werkzeug gibt, steht hier nicht** — und das ist
	// Absicht: `internal/werkzeuge/faehigkeiten_test.go` hält beide
	// Richtungen zwischen `docs/FAEHIGKEITEN.md` und dem Quelltext, und eine
	// zweite Prüfung daneben liefe beim nächsten Werkzeug auseinander (§10).
	// Der Grund für das Nein steht in der Fähigkeitentabelle.
}
