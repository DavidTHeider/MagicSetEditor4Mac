package app_test

// **Die Anmeldebindungen** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Anbieterliste kommt ohne Kontakt nach draußen.**
//
// **Und das ist die Zusage, die dahintersteht** (§16, Regel 5): wer die Seite
// öffnet, hat nichts angestoßen. Erst ein Klick auf einen Anbieter macht einen
// Browser auf.
func TestDieAnbieterlisteKommtOhneKontaktNachDraussen(t *testing.T) {
	antwort := angelaufeneAnwendung(t).Anbieter(t.Context())

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Anbieter) != 4 {
		t.Errorf("%d Anbieter: %v", len(antwort.Anbieter), antwort.Anbieter)
	}
}

// **Eine Anmeldung ohne eingetragene Kennung ist ein fachliches Nein.**
//
// # Der Normalfall, nicht der Sonderfall
//
// **Dieses Projekt liefert keine Kennungen mit** (Vermerk 0066). Wer sich
// anmelden will, registriert seine eigene Anwendung beim Anbieter — und bis
// dahin sagt die Anwendung das, statt einen Browser auf eine Adresse zu
// schicken, die der Anbieter abweist.
func TestEineAnmeldungOhneKennungIstEinFachlichesNein(t *testing.T) {
	antwort := angelaufeneAnwendung(t).Anmelden(t.Context(), "github")

	if antwort.Fehler == nil {
		t.Fatal("die Anmeldung ohne Kennung kam durch")
	}

	if antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("Code %q", antwort.Fehler.Code)
	}
}

// **Die Konten eines Nutzers lassen sich lesen und lösen.**
func TestDieKontenLassenSichLesenUndLoesen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	konten := &kontenprobe{konten: []domain.Anbieterkonto{
		{Anbieter: "github", Anzeigename: "Wer"},
	}}

	anwendung.MitAnmeldung(nil, konten)

	antwort := anwendung.Anbieterkonten(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Konten) != 1 || antwort.Konten[0].Anbieter != "github" {
		t.Fatalf("Konten %+v", antwort.Konten)
	}

	if geloest := anwendung.KontoLoesen(t.Context(), "1", "github"); geloest.Fehler != nil {
		t.Fatalf("%+v", geloest.Fehler)
	}

	if konten.geloest != "github" {
		t.Errorf("gelöst wurde %q", konten.geloest)
	}
}

// **Eine unbrauchbare Nutzerkennung wird an beiden Wegen abgewiesen.**
func TestEineUnbrauchbareNutzerkennungWirdAbgewiesen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitAnmeldung(nil, &kontenprobe{})

	if antwort := anwendung.Anbieterkonten(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("das Lesen kam durch")
	}

	if antwort := anwendung.KontoLoesen(t.Context(), "keine Zahl", "github"); antwort.Fehler == nil {
		t.Error("das Lösen kam durch")
	}
}

// **Und jeder Fehler der Ablage reicht durch.**
func TestJederKontenfehlerReichtDurch(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitAnmeldung(nil, &kontenprobe{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.Anbieterkonten(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("das Lesen kam durch")
	}

	if antwort := anwendung.KontoLoesen(t.Context(), "1", "github"); antwort.Fehler == nil {
		t.Error("das Lösen kam durch")
	}
}

// kontenprobe ist ein Kontenführer im Arbeitsspeicher.
type kontenprobe struct {
	konten  []domain.Anbieterkonto
	geloest string
	fehler  error
}

func (p *kontenprobe) Konten(context.Context, int64) ([]domain.Anbieterkonto, error) {
	return p.konten, p.fehler
}

func (p *kontenprobe) KontoLoesen(_ context.Context, _ int64, anbieter string) error {
	if p.fehler != nil {
		return p.fehler
	}

	p.geloest = anbieter

	return nil
}

// **Nach der Anmeldung steht der Nutzer samt seinen Profilen da.**
//
// # Warum das eigens geprüft wird
//
// **Weil die Seite nach der Anmeldung dasselbe zeigen soll wie vorher.** Ein
// Nutzer ohne seine Profile sähe aus wie einer, der keine hat — und der Weg
// dorthin geht über die Übersicht, nicht über eine zweite Abfrage (§7,
// Regel 2).
func TestNachDerAnmeldungStehtDerNutzerMitProfilen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	anwendung.MitAnmeldung(
		nahtGegen(t, &anmeldedienstprobe{nutzerID: 1}, nil), &kontenprobe{})

	antwort := anwendung.Anmelden(t.Context(), "github")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Nutzer == nil || len(antwort.Nutzer.Profile) != 1 {
		t.Fatalf("Nutzer %+v", antwort.Nutzer)
	}
}

// **Ein angemeldeter Nutzer, den die Übersicht nicht führt, ist ein Fehler.**
//
// Der Fall kann nicht vorkommen — die Anmeldung hat ihn gerade geliefert. Wenn
// er doch vorkommt, sind zwei Stellen auseinander, und eine leere Antwort sähe
// aus wie „niemand angemeldet" (§8).
func TestEinNutzerDenDieUebersichtNichtFuehrtIstEinFehler(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	anwendung.MitAnmeldung(
		nahtGegen(t, &anmeldedienstprobe{nutzerID: 4711}, nil), &kontenprobe{})

	antwort := anwendung.Anmelden(t.Context(), "github")
	if antwort.Fehler == nil {
		t.Fatal("der unbekannte Nutzer kam durch")
	}

	if antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q", antwort.Fehler.Code)
	}
}

// **Ein Bestand, der die Übersicht nicht hergibt, meldet sich.**
func TestEineUnlesbareUebersichtNachDerAnmeldungMeldetSich(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	// **Die Übersicht klagt** — und die Anmeldung selbst glückt. Ohne diese
	// Trennung sähe der Lauf so aus, als wäre die Anmeldung gescheitert.
	anwendung.MitNutzern(&nutzerfuehrer{fehler: errors.New("die Platte schweigt")})
	anwendung.MitAnmeldung(
		nahtGegen(t, &anmeldedienstprobe{nutzerID: 1}, nil), &kontenprobe{})

	if antwort := anwendung.Anmelden(t.Context(), "github"); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}
