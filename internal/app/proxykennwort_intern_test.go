package app

// **Kein Proxykennwort in einer Antwort** — S13.NAT-02
// (#1706):
// „keine Geheimnisse in Standardanzeige oder Export".
//
// Die Ausfuhr ist es nicht: der Abschnitt `geraet` reist nie mit. Die
// Anzeige war es — der Netzstand nannte `HTTPS_PROXY` wörtlich, und die
// Einstellungsliste den gespeicherten Wert, beides samt Kennwort.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ohneGeheimnis prüft, dass eine Antwort das Kennwort nirgends nennt.
func ohneGeheimnis(t *testing.T, was string, antwort any) {
	t.Helper()

	roh, fehler := json.Marshal(antwort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if strings.Contains(string(roh), "geheim") {
		t.Errorf("%s nennt das Kennwort: %s", was, roh)
	}

	// **Die Gegenprobe, dass die Adresse überhaupt dasteht** (§3) — eine
	// leere Antwort nennt auch kein Kennwort.
	if !strings.Contains(string(roh), "anna:xxxxx@proxy.firma:3128") {
		t.Errorf("%s nennt die Adresse nicht: %s", was, roh)
	}
}

// **Ein Wert von vorher** — gesetzt, bevor die Prüfung ihn abwies.
func TestEinAlterProxywertZeigtSeinKennwortNicht(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(
		ctx, domain.NetzProxyHTTPS, "http://anna:geheim@proxy.firma:3128"); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung := NetzflaecheUeber(db)
	ohneGeheimnis(t, "der Netzstand", anwendung.Netzstand(ctx))
	ohneGeheimnis(t, "das Fenster", anwendung.Einstellungen(ctx))

	alle, fehler := EinstellungsnahtUeber(db).Alle(ctx)
	if fehler != nil {
		t.Fatal(fehler)
	}

	ohneGeheimnis(t, "die Flächen ohne Fenster", alle)
}

// **Ein Wert aus der Umgebung** — dort bleibt ein Kennwort erlaubt, denn
// Go schickt es als `Proxy-Authorization` mit; gezeigt wird es nicht.
func TestEinProxykennwortAusDerUmgebungSiehtNiemand(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://anna:geheim@proxy.firma:3128")

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	ohneGeheimnis(t, "der Netzstand", NetzflaecheUeber(db).Netzstand(ctx))
}

// **Und neu gesetzt wird keines** — der Fehler steht am Feld.
func TestEinProxykennwortWirdNichtGesetzt(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	antwort := NetzflaecheUeber(db).EinstellungSetzen(
		ctx, domain.NetzProxyHTTPS, "http://anna:geheim@proxy.firma:3128")
	if antwort.Fehler == nil || antwort.Fehler.Code != CodeValidierung {
		t.Fatalf("die Antwort ist %+v", antwort.Fehler)
	}

	// **Auch die Meldung nennt es nicht.**
	if strings.Contains(antwort.Fehler.Meldung, "geheim") {
		t.Errorf("die Meldung nennt das Kennwort: %q", antwort.Fehler.Meldung)
	}
}
