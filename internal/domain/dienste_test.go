package domain_test

// **Die vier Kartendienste als Einstellung** — Werkstück 2 von Phase 4.8
// (#1321),
// nach [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Dienst hat einen Schalter, und drei von vieren stehen auf aus.**
//
// Vermerk 0063, Verworfen: „Ein Dienst, dessen Bedingungen jemand nicht
// gelesen hat, soll nicht in seinem Namen fragen." Scryfall bleibt an, weil
// es heute schon an ist und die Anwendung ohne es weniger könnte als gestern.
func TestJederDienstHatEinenSchalter(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		dienst  domain.Dienst
		vorgabe string
	}{
		{domain.DienstScryfall, domain.DienstAn},
		{domain.DienstMTGAPI, domain.DienstAus},
		{domain.DienstMTGJSON, domain.DienstAus},
		{domain.DienstScrydex, domain.DienstAus},
		{domain.DienstCardTrader, domain.DienstAus},
	} {
		t.Run(string(fall.dienst), func(t *testing.T) {
			t.Parallel()

			schluessel := domain.DienstEinstellung(fall.dienst)

			if !slices.Contains(domain.Einstellungsreihe(), schluessel) {
				t.Fatalf("%q steht nicht in der Einstellungsreihe", schluessel)
			}

			if vorgabe := domain.EinstellungsVorgabe(schluessel); vorgabe != fall.vorgabe {
				t.Errorf("Vorgabe ist %q, erwartet %q", vorgabe, fall.vorgabe)
			}

			if werte := domain.ErlaubteEinstellungswerte(schluessel); !slices.Equal(
				werte, []string{domain.DienstAus, domain.DienstAn}) {
				t.Errorf("erlaubt sind %v, erwartet aus und an", werte)
			}
		})
	}
}

// **Die Aufzählung der Dienste ist eine und nicht zwei.**
//
// Wer einen achten Dienst aufnimmt, trägt ihn in [domain.Dienste] ein — und
// bekommt Schalter, Vorgabe und Prüfung daraus. Eine zweite Liste daneben
// wäre die Falle, die bei den Einstellungen schon zweimal zugeschnappt ist
// (siehe den Kopf von `einstellung.go`).
func TestJederDienstDerAufzaehlungHatSeineZeile(t *testing.T) {
	t.Parallel()

	if len(domain.Dienste()) != 8 {
		t.Fatalf("es sind %d Dienste, erwartet 8", len(domain.Dienste()))
	}

	for _, dienst := range domain.Dienste() {
		schluessel := domain.DienstEinstellung(dienst)

		if !slices.Contains(domain.Einstellungsreihe(), schluessel) {
			t.Errorf("%q hat keine Zeile in der Einstellungsreihe", schluessel)
		}
	}
}

// **Ein unsinniger Wert wird abgewiesen.**
func TestEinUnsinnigerDienstwertWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	err := domain.PruefeEinstellung(
		domain.DienstEinstellung(domain.DienstScrydex), "vielleicht")

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Und die beiden, die einen Schlüssel brauchen, sagen es.**
//
// Vermerk 0063, Entscheidung 3: „an, aber Schlüssel fehlt" ist ein eigener
// Zustand. Wer nicht weiß, welcher Dienst überhaupt einen braucht, kann ihn
// nicht unterscheiden.
func TestDreiDiensteBrauchenEinenSchluessel(t *testing.T) {
	t.Parallel()

	for dienst, erwartet := range map[domain.Dienst]bool{
		domain.DienstScryfall:   false,
		domain.DienstMTGAPI:     false,
		domain.DienstMTGJSON:    true,
		domain.DienstScrydex:    true,
		domain.DienstCardTrader: true,
	} {
		if dienst.BrauchtSchluessel() != erwartet {
			t.Errorf("%s.BrauchtSchluessel() ist %v, erwartet %v",
				dienst, dienst.BrauchtSchluessel(), erwartet)
		}
	}
}

// **Scrydex braucht zwei Angaben, nicht eine.**
//
// `X-Api-Key` **und** `X-Team-ID` (0056, nachgelesen am 2026-08-30). Ein
// Dienst, der nur nach einem Schlüssel fragte, wäre eingerichtet und
// antwortete trotzdem nicht.
func TestScrydexBrauchtZweiAngaben(t *testing.T) {
	t.Parallel()

	for dienst, erwartet := range map[domain.Dienst][]string{
		domain.DienstScryfall:   nil,
		domain.DienstMTGAPI:     nil,
		domain.DienstMTGJSON:    {domain.ZugangMarke},
		domain.DienstScrydex:    {domain.ZugangSchluessel, domain.ZugangTeam},
		domain.DienstCardTrader: {domain.ZugangMarke},
	} {
		if angaben := dienst.Zugangsangaben(); !slices.Equal(angaben, erwartet) {
			t.Errorf("%s verlangt %v, erwartet %v", dienst, angaben, erwartet)
		}
	}
}

// **Ein Dienst, den es nicht gibt, verlangt keine Zugangsangaben.**
//
// Der Fall entsteht, wenn jemand eine Kennung aus einer Einstellung liest,
// die aus einer älteren Fassung stammt. Eine leere Liste ist die richtige
// Antwort: „ich weiß von keinem Zugang" — und nicht ein Absturz.
func TestEinUnbekannterDienstVerlangtNichts(t *testing.T) {
	t.Parallel()

	fremd := domain.Dienst("gibtsnicht")

	if angaben := fremd.Zugangsangaben(); len(angaben) != 0 {
		t.Errorf("ein unbekannter Dienst verlangt %v", angaben)
	}

	if fremd.BrauchtSchluessel() {
		t.Error("ein unbekannter Dienst behauptet, er brauche einen Schlüssel")
	}
}

// **Das Konto im Schlüsselbund trägt Dienst und Angabe.**
//
// **Ohne Leerzeichen und Sonderzeichen**, weil es auf der Befehlszeile von
// `security` mitreist — `internal/schluesselbund` weist alles andere ab.
func TestDasZugangskontoNenntBeides(t *testing.T) {
	t.Parallel()

	konto := domain.ZugangKonto(domain.DienstScrydex, domain.ZugangTeam)

	if konto != "scrydex.team" {
		t.Errorf("das Konto heißt %q, erwartet »scrydex.team«", konto)
	}
}

// **»Der Dienst ist aus« ist ein eigener Ausgang.**
//
// Die Abhilfe ist eine andere als bei einer Validierung: nicht „berichtige
// die Eingabe", sondern „schalte ihn ein, wenn du ihn willst" (§8).
func TestEinAbgeschalteterDienstIstEinEigenerAusgang(t *testing.T) {
	t.Parallel()

	err := domain.DienstIstAus(domain.DienstScrydex)

	if !domain.IstDienstAus(err) {
		t.Errorf("%v gilt nicht als »Dienst aus«", err)
	}

	if !strings.Contains(err.Error(), "scrydex") {
		t.Errorf("die Meldung %q nennt den Dienst nicht", err)
	}

	// **Und ein anderer Fehler gilt nicht als solcher** — sonst sagte die
	// Prüfung bei jedem Fehler „schalte ihn ein".
	if domain.IstDienstAus(domain.ErrNotFound) {
		t.Error("ErrNotFound gilt als »Dienst aus«")
	}
}

// **Ein fehlender Zugang nennt Dienst und Angabe** — Zusatzbefund 1 des
// Nachlaufs über Phase 4.8.
//
// Vorher verschwand ein Dienst ohne Zugang schlicht aus der Quellenkarte, und
// die Meldung riet danach zwischen zwei Möglichkeiten. Wer sie liest, soll
// wissen, **was** er eintragen muss und **wo** (§8).
func TestEinFehlenderZugangNenntDienstUndAngabe(t *testing.T) {
	t.Parallel()

	err := domain.ZugangFehlt(domain.DienstScrydex, domain.ZugangTeam)

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v, erwartet ErrValidation", err)
	}

	for _, teil := range []string{"scrydex", domain.ZugangTeam, "Einstellungen"} {
		if !strings.Contains(err.Error(), teil) {
			t.Errorf("die Meldung %q nennt %q nicht", err, teil)
		}
	}
}

// **Ein abgeschalteter Dienst trägt seinen Namen** — I05
// (#2263):
// die Oberfläche nennt ihn im Rat, und dafür muss sie ihn lesen können.
func TestEinAbgeschalteterDienstTraegtSeinenNamen(t *testing.T) {
	t.Parallel()

	fehler := domain.DienstIstAus(domain.DienstScrydex)

	var aus *domain.AbgeschalteterDienst
	if !errors.As(fehler, &aus) || aus.Dienst != domain.DienstScrydex {
		t.Fatalf("der Fehler trägt keinen Dienst: %#v", fehler)
	}

	if !errors.Is(fehler, domain.ErrDienstAus) {
		t.Error("der Fehler ist kein ErrDienstAus")
	}

	if fehler.Error() != "der Dienst ist abgeschaltet: scrydex darf nicht gefragt werden — "+
		"unter Einstellungen lässt er sich einschalten" {
		t.Errorf("die Meldung ist %q", fehler.Error())
	}
}
