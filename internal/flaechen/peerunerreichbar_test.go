package flaechen_test

// **Ein Gegenüber, das nicht horcht, heißt so** — #2491.
//
// Gefunden mit der Hüllenprobe `p13abschluss09`: horchte Anna nicht, scheiterte
// „Jetzt übertragen" mit dem allgemeinen Satz zu `nichtGefunden` — „lade die
// Liste neu", obwohl nichts gelöscht war. **Hier über die Kommandozeile**: der
// Satz nennt, was ist, die zweite Zeile den Rat, und `--json` den Code und den
// Schlüssel, die das Fenster bekommt. **Der Code bleibt `nichtGefunden`**
// (§16: erweiternd).

import (
	"net"
	"strings"
	"testing"
)

func TestEinGegenueberDasNichtHorchtHeisstSo(t *testing.T) {
	anna, bert := zweiImLabor(t)

	// **Eine Adresse, an der gerade niemand annimmt.**
	lauscher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	adresse := lauscher.Addr().String()
	_ = lauscher.Close() //nolint:errcheck // er soll gerade nicht mehr da sein

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, adresse)
	einHandelImAusgang(t, anna, bert)

	aus, klage, code := kommandozeile(t, bert.umgebung, "handelszustellung", anna.kennung, "--json")
	if code == 0 {
		t.Fatalf("die Zustellung meldete Erfolg, obwohl niemand horchte: %q", aus)
	}

	for _, soll := range []string{"das Gegenüber ist nicht zu erreichen", "mse kontakt-adresse"} {
		if !strings.Contains(klage, soll) {
			t.Errorf("die Fehlerausgabe nennt %q nicht: %q", soll, klage)
		}
	}

	for _, soll := range []string{`"code":"nichtGefunden"`, `"schluessel":"peerNichtErreichbar"`} {
		if !strings.Contains(aus, soll) {
			t.Errorf("die Ausgabe mit --json trägt %s nicht: %q", soll, aus)
		}
	}

	// **Und die Nachricht bleibt, wo sie ist.**
	if stand := ausgangBei(t, bert, anna.kennung); stand.Offen != 1 {
		t.Errorf("im Ausgang steht %+v", stand)
	}
}

// **Und eines ohne Adresse** — gefunden mit der Hüllenprobe `p10handel09`:
// dort stand derselbe Satz, und der Rat ist ein Feld am Kontakt.
func TestEinGegenueberOhneAdresseHeisstSo(t *testing.T) {
	anna, bert := zweiImLabor(t)

	einHandelImAusgang(t, anna, bert)

	aus, klage, code := kommandozeile(t, bert.umgebung, "handelszustellung", anna.kennung, "--json")
	if code == 0 {
		t.Fatalf("die Zustellung meldete Erfolg ohne Adresse: %q", aus)
	}

	if !strings.Contains(klage, "keine Adresse eingetragen") || !strings.Contains(klage, "mse kontakt-adresse") {
		t.Errorf("die Fehlerausgabe: %q", klage)
	}

	if !strings.Contains(aus, `"schluessel":"peerOhneAdresse"`) {
		t.Errorf("die Ausgabe mit --json: %q", aus)
	}
}
