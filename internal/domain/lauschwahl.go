package domain

// **Wo die Dienste lauschen — eine oder mehrere Adressen** — G08
// (#2272).
//
// Bis dahin nahm jeder Dienst **eine** Adresse ([Bindung]). Wer ihn im
// eigenen Netz und zugleich auf der Schleife erreichbar haben wollte, hatte
// die Wahl zwischen `0.0.0.0` — jede Schnittstelle, auch die, die er nicht
// meinte — und einer. Die Lauschwahl nennt **jede gemeinte Adresse**, und
// jede bekommt ihren Lauscher.
//
// # Was sie nicht tut
//
// **Sie löst keinen Namen auf** und kennt keine Schnittstelle beim Namen
// (`en0`): eine Schnittstelle trägt mehrere Adressen, und welche gemeint ist,
// sagt die Wahl — die Liste dieses Rechners steht in `internal/entdeckung`.
// **Und „alle" steht allein**: neben `0.0.0.0` eine einzelne Adresse zu
// nennen hieße, zwei Lauscher auf denselben Port zu legen, und der zweite
// scheiterte.

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"
)

// NetzLauschen ist die Einstellung der Lauschwahl — **leer heißt die
// Schleife**, sonst Adressen mit Komma getrennt.
//
// **Unter `netz.` und nicht unter `dienst.`**: jeder Schlüssel der Gestalt
// `dienst.<name>` ist ein Dienstschalter, und „nur lokal" legte ihn auf
// `aus` (`mitNurLokal` in `geltung.go`).
const NetzLauschen = "netz.lauschen"

// Lauschwahl ist, wo die Dienste lauschen — in der Reihenfolge der Wahl.
type Lauschwahl struct{ adressen []string }

// LauschwahlAus liest eine Lauschwahl und prüft jede Adresse.
func LauschwahlAus(text string) (Lauschwahl, error) {
	if strings.TrimSpace(text) == "" {
		return Lauschwahl{adressen: []string{Schleifenadresse}}, nil
	}

	var adressen []string

	for _, teil := range strings.Split(text, ",") {
		adresse := strings.TrimSpace(teil)
		if adresse == "" {
			return Lauschwahl{}, fmt.Errorf("%w: %q enthält einen leeren Eintrag", ErrValidation, text)
		}

		if _, fehler := netip.ParseAddr(adresse); fehler != nil {
			return Lauschwahl{}, fmt.Errorf(
				"%w: %q ist keine Adresse — gemeint ist die Adresse einer Schnittstelle, "+
					"ohne Port und ohne Namen", ErrValidation, adresse)
		}

		if slices.Contains(adressen, adresse) {
			return Lauschwahl{}, fmt.Errorf("%w: %s steht zweimal da", ErrValidation, adresse)
		}

		adressen = append(adressen, adresse)
	}

	if len(adressen) > 1 && (slices.Contains(adressen, JedeSchnittstelle) || slices.Contains(adressen, "::")) {
		return Lauschwahl{}, fmt.Errorf(
			"%w: „alle\" schließt jede andere Adresse ein — wähle alle oder einzelne", ErrValidation)
	}

	return Lauschwahl{adressen: adressen}, nil
}

// PruefeLauschwahl ist die Prüfung der Einstellung.
func PruefeLauschwahl(text string) error {
	_, fehler := LauschwahlAus(text)

	return fehler
}

// Adressen gibt die Adressen der Wahl als Kopie.
func (w Lauschwahl) Adressen() []string { return slices.Clone(w.adressen) }

// Text gibt die Wahl in der Gestalt der Einstellung.
func (w Lauschwahl) Text() string { return strings.Join(w.adressen, ",") }

// VerlaesstDenRechner sagt, ob **eine** der Adressen von außen erreichbar ist
// — dann gelten TLS und Token für den ganzen Dienst (0115).
func (w Lauschwahl) VerlaesstDenRechner() bool {
	for _, bindung := range w.Bindungen(0) {
		if bindung.VerlaesstDenRechner() {
			return true
		}
	}

	return false
}

// Bindungen gibt je Adresse eine [Bindung] auf demselben Port.
func (w Lauschwahl) Bindungen(port int) []Bindung {
	bindungen := make([]Bindung, 0, len(w.adressen))
	for _, adresse := range w.adressen {
		bindungen = append(bindungen, Bindung{Adresse: adresse, Port: port})
	}

	return bindungen
}

// Lauschbefund nennt eine gewählte Adresse, auf der kein Lauscher entstand,
// und den Grund — für die Ansage und das Fenster.
type Lauschbefund struct {
	Adresse string
	Grund   string
}

// Gelungene gibt die Bindungen, auf denen ein Lauscher entstand — in der
// Reihenfolge der Wahl und damit in der der angesagten Adressen.
//
// **Befund 3 des Zweitblicks über Phase G** (#2272): die Aufrufer paarten die
// angesagten Adressen über die Stelle in der **gewählten** Liste. Fiel eine
// Bindung aus, rückte jede danach auf und bekam die Adresse einer anderen —
// mit falschen Dienstnamen und einer Erreichbarkeit, die an der falschen
// Stelle fragt.
//
// **Ein Befund nimmt genau eine Bindung mit** — zwei mit derselben Adresse,
// von denen eine scheitert, lassen die andere stehen.
func Gelungene(bindungen []Bindung, fehlende []Lauschbefund) []Bindung {
	offen := map[string]int{}
	for _, befund := range fehlende {
		offen[befund.Adresse]++
	}

	gelungen := make([]Bindung, 0, len(bindungen))

	for _, bindung := range bindungen {
		if offen[bindung.Wirkadresse()] > 0 {
			offen[bindung.Wirkadresse()]--

			continue
		}

		gelungen = append(gelungen, bindung)
	}

	return gelungen
}
