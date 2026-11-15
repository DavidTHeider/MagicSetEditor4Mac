package app

// **Die Marktlinks an der Naht** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
// **Dass der Öffner das Argument bekommt, das er bekommen soll** — und dass
// eine Adresse von außen ihn nie erreicht. Das Issue verlangt es wörtlich:
// „Test prüft Argument des Öffners."
//
// **Und dass eine Liste leer und nicht `null` ankommt.**

import (
	"context"
	"errors"
	"testing"
)

func TestDieMarktlinkeKommenDurchDieNaht(t *testing.T) {
	anwendung := mitSaat(t)
	ctx := t.Context()

	antwort := anwendung.Marktlinks(ctx, "", "Lightning Bolt")
	if antwort.Fehler != nil {
		t.Fatalf("Marktlinks: %+v", antwort.Fehler)
	}

	if len(antwort.Links) != 6 {
		t.Fatalf("es kamen %d Links", len(antwort.Links))
	}

	for _, link := range antwort.Links {
		if link.Anbieter == "" || link.Art == "" || link.Adresse == "" {
			t.Errorf("der Link ist %+v", link)
		}
	}
}

// **Ein klagender Bestand klagt, und die Liste ist leer statt `null`.**
//
// **Mit einem geschlossenen Bestand**, weil ein Druck nachgeschlagen wird:
// ohne ihn gäbe es keinen Lesefehler, und der Zweig bliebe unbetreten (§10).
func TestEinKlagenderBestandGibtKeineMarktlinke(t *testing.T) {
	anwendung := mitSaat(t)

	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	antwort := anwendung.Marktlinks(t.Context(), "bolt-m10", "")
	if antwort.Fehler == nil {
		t.Error("die Links kamen aus einem geschlossenen Bestand durch")
	}

	if antwort.Links == nil {
		t.Error("die Liste ist nil statt leer")
	}
}

// **Der Öffner bekommt genau die geprüfte Adresse — und sonst keine.**
func TestDerOeffnerAnDerNahtBekommtNurGeprueftes(t *testing.T) {
	anwendung := mitSaat(t)
	ctx := t.Context()

	gesehen := ""

	anwendung.MitOeffner(func(_ context.Context, adresse string) error {
		gesehen = adresse

		return nil
	})

	const gut = "https://www.tcgplayer.com/product/33632"

	if antwort := anwendung.MarktlinkOeffnen(
		ctx, "tcgplayer", gut); antwort.Fehler != nil {
		t.Fatalf("MarktlinkOeffnen: %+v", antwort.Fehler)
	}

	if gesehen != gut {
		t.Errorf("der Öffner bekam %q", gesehen)
	}

	// **Und was von außen kommt, erreicht ihn nicht.**
	gesehen = ""

	if antwort := anwendung.MarktlinkOeffnen(
		ctx, "tcgplayer", "https://boese.example/"); antwort.Fehler == nil {
		t.Error("der fremde Wirt kam durch")
	}

	if gesehen != "" {
		t.Errorf("der Öffner bekam %q", gesehen)
	}
}

// **Ein klagender Öffner kommt als Klage an.**
func TestEinKlagenderOeffnerAnDerNahtKommtAlsKlageAn(t *testing.T) {
	anwendung := mitSaat(t)

	anwendung.MitOeffner(func(context.Context, string) error {
		return errors.New("kein Browser")
	})

	if antwort := anwendung.MarktlinkOeffnen(t.Context(), "gatherer",
		"https://gatherer.wizards.com/"); antwort.Fehler == nil {
		t.Error("die Klage ging verloren")
	}
}

// **Ohne gesetzten Öffner sagt es die Fläche.**
//
// **Ein Absturz wäre schlimmer als ein Nein**: eine Fläche, die keinen
// Browser hat, soll es sagen — und nicht das Fenster für immer warten lassen.
func TestOhneGesetztenOeffnerSagtEsDieFlaeche(t *testing.T) {
	anwendung := mitSaat(t)
	anwendung.MitOeffner(nil)

	if antwort := anwendung.MarktlinkOeffnen(t.Context(), "gatherer",
		"https://gatherer.wizards.com/"); antwort.Fehler == nil {
		t.Error("ohne Öffner wurde geöffnet")
	}
}
