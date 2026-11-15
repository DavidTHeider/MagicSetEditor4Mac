package domain_test

// **Was über ein erzeugtes Bild feststeht** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass ein Bildauftrag nur nennt, was auch angeboten wird.** Größe und Güte
// stehen als Aufzählung da, nicht als freier Text: beide gehen in den Preis
// ein, und ein Wort, das der Dienst nicht kennt, ist ein 400, für den jemand
// gewartet hat.
//
// **Und dass eine fehlende Angabe als fehlend dasteht.** Die Ergänzung des
// Issues sagt es scharf: „Fehlende Angaben als unbekannt führen; Seed allein
// garantiert keine pixelgleiche Reproduktion." Eine Herkunft, die einen Seed
// erfindet, ist schlimmer als eine, die keinen nennt.

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einBildauftrag ist ein gültiger.
func einBildauftrag() domain.KIBildauftrag {
	return domain.KIBildauftrag{
		Aufgabe: "ein Blitz über einer Ebene", Feld: "image",
		Groesse: domain.KIBildgroesseHoch, Guete: domain.KIBildgueteMittel,
	}
}

// **Ein vollständiger Auftrag geht durch.**
func TestEinVollstaendigerBildauftragGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := einBildauftrag().Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}
}

// **Und jeder unvollständige nicht.**
func TestEinUnvollstaendigerBildauftragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.KIBildauftrag){
		"ohne Aufgabe":     func(a *domain.KIBildauftrag) { a.Aufgabe = "" },
		"nur Leerraum":     func(a *domain.KIBildauftrag) { a.Aufgabe = "  \n " },
		"ohne Feld":        func(a *domain.KIBildauftrag) { a.Feld = "" },
		"erfundene Größe":  func(a *domain.KIBildauftrag) { a.Groesse = "4000x4000" },
		"erfundene Güte":   func(a *domain.KIBildauftrag) { a.Guete = "beste" },
		"die Güte »auto«":  func(a *domain.KIBildauftrag) { a.Guete = "auto" },
		"die Güte »xhigh«": func(a *domain.KIBildauftrag) { a.Guete = "xhigh" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			auftrag := einBildauftrag()
			kaputt(&auftrag)

			if fehler := auftrag.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", name, fehler)
			}
		})
	}
}

// **Eine zu lange Aufgabe auch nicht.**
//
// **Weil der Prompt bezahlt wird.** Ein Feld ohne Grenze ist ein Feld, in das
// jemand ein halbes Buch einfügt — versehentlich, beim Kopieren.
func TestEineZuLangeAufgabeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	auftrag := einBildauftrag()
	auftrag.Aufgabe = strings.Repeat("x", domain.KIBildaufgabeHoechstens+1)

	if fehler := auftrag.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die lange Aufgabe kam durch: %v", fehler)
	}
}

// **Leere Größe und leere Güte gehen durch — sie heißen „der Dienst
// entscheidet".**
func TestOhneGroesseUndGueteEntscheidetDerDienst(t *testing.T) {
	t.Parallel()

	auftrag := einBildauftrag()
	auftrag.Groesse = ""
	auftrag.Guete = ""

	if fehler := auftrag.Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}
}

// **Die drei Größen und die drei Güten stehen als Aufzählung da.**
func TestDieAufzaehlungenSindDrei(t *testing.T) {
	t.Parallel()

	if groessen := domain.KIBildgroessen(); len(groessen) != 3 {
		t.Errorf("die Größen sind %v", groessen)
	}

	if gueten := domain.KIBildgueten(); len(gueten) != 3 {
		t.Errorf("die Güten sind %v", gueten)
	}

	// **Und eine Kopie, keine Sicht auf das Original.**
	domain.KIBildgroessen()[0] = "verbogen"

	if domain.KIBildgroessen()[0] == "verbogen" {
		t.Error("die Aufzählung ließ sich von außen ändern")
	}
}

// **Eine Herkunft ohne Angaben sagt »unbekannt« — und erfindet nichts.**
func TestEineHerkunftOhneAngabenSagtUnbekannt(t *testing.T) {
	t.Parallel()

	herkunft := domain.KIBildherkunft{
		Anbieter: "openai", Modell: "gpt-image-2.5-flare",
		Aufgabe: "ein Blitz", Auftrag: "ki-4-1",
		Zeitpunkt: time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC),
	}

	if !herkunft.Kette.Unbekannt() {
		t.Error("die Artefaktkette gilt als bekannt")
	}

	if len(herkunft.Parameter) != 0 {
		t.Errorf("es stehen Parameter da: %+v", herkunft.Parameter)
	}

	// **Der Seed steht nicht dabei, weil ihn niemand genannt hat.**
	if wert, da := herkunft.Parameterwert("seed"); da {
		t.Errorf("ein Seed wurde erfunden: %q", wert)
	}
}

// **Ein genannter Parameter kommt heraus, wie er hineinging.**
func TestEinGenannterParameterKommtHeraus(t *testing.T) {
	t.Parallel()

	herkunft := domain.KIBildherkunft{Parameter: []domain.KIBildparameter{
		{Name: "groesse", Wert: "1024x1536"},
		{Name: "guete", Wert: "medium"},
	}}

	if wert, da := herkunft.Parameterwert("groesse"); !da || wert != "1024x1536" {
		t.Fatalf("die Größe ist %q (%v)", wert, da)
	}
}

// **Die Herkunft wird zur Feldherkunft aus A11 — ohne die Bildparameter.**
//
// # Warum eine Umrechnung und keine zweite Tafel
//
// **Weil ein Bildfeld ein Kartenfeld ist.** Die Tafel aus A11 hält je Feld,
// woher sein Wert stammt; ein Bild ist keine Ausnahme davon. Was sie **nicht**
// hält, sind Größe, Güte und Umschrift — die stehen daneben, weil sie nur für
// Bilder gelten.
func TestDieBildherkunftWirdZurFeldherkunft(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)

	herkunft := domain.KIBildherkunft{
		Anbieter: "openai", Modell: "gpt-image-2.5-flare",
		Aufgabe: "ein Blitz", Umschrift: "ein Blitz, Ölgemälde",
		Auftrag: "ki-4-1", Zeitpunkt: jetzt,
		Parameter: []domain.KIBildparameter{{Name: "guete", Wert: "medium"}},
	}

	feld := herkunft.AlsFeldherkunft("image", "sha256:abc")

	if feld.Feld != "image" || feld.Bezug != "sha256:abc" {
		t.Fatalf("die Feldherkunft ist %+v", feld)
	}

	if feld.Anbieter != "openai" || feld.Modell != "gpt-image-2.5-flare" ||
		feld.Aufgabe != "ein Blitz" || feld.Auftrag != "ki-4-1" ||
		!feld.Zeitpunkt.Equal(jetzt) {
		t.Fatalf("die Feldherkunft ist %+v", feld)
	}
}
