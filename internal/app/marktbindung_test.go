package app_test

// **Die Naht zum Marktplatz** — Werkstück A4 von Phase 5.3
// (#1443).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Marktplatz, den es nicht gibt, wird abgewiesen — und der Name steht
// in der Meldung.**
func TestEinUnbekannterMarktplatzWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		Marktangebote(probekontext(t), "cardmarket", "abc", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("cardmarket kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "cardmarket") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Eine Ausführung, die es nicht gibt, ebenso.**
//
// **Und sie wird vor dem Schlüsselbund geprüft**: sonst hinge die Meldung
// davon ab, ob jemand einen Zugang eingetragen hat.
func TestEineUnbekannteAusfuehrungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		Marktangebote(probekontext(t), "cardtrader", "abc", "", "glaenzend", nil)

	if antwort.Fehler == nil {
		t.Fatal("eine erfundene Ausführung kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "glaenzend") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Ohne Bestand gibt es keine Auskunft.**
func TestOhneBestandGibtEsKeineMarktauskunft(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").
		Marktangebote(probekontext(t), "cardtrader", "abc", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam eine Auskunft")
	}
}

// **Ohne Zugangsmarke wird nicht gefragt.**
//
// Der Schlüsselbund ist eingesetzt und leer — ohne ihn führe der Lauf in den
// echten Bund dieses Rechners (§3).
func TestOhneMarkeGibtEsKeineMarktauskunft(t *testing.T) {
	t.Parallel()

	antwort := app.MarktAnwendungUeber(frischerBestand(t)).
		MitMarktbund(marktbund{}).
		Marktangebote(probekontext(t), "cardtrader", "abc", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ohne Marke kam eine Auskunft")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "marke") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}
