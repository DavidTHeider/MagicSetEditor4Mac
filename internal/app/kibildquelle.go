package app

// **Wer ein Bild erzeugen kann — und wer nicht** — Werkstück A12 von
// Phase 5.3 (#1451).
//
// # Warum eine zweite Schnittstelle und keine vierte Methode
//
// **Weil drei von vier Anbietern kein Bild erzeugen.** Eine vierte Methode an
// [KILaufzeitquelle] zwänge jeden von ihnen, sie zu haben — und drei davon
// müssten darin „kann ich nicht" sagen. Das ist eine Zusage, die nur aussieht
// wie eine: die Liste der Anbieter behauptete dann eine Fähigkeit, die es an
// drei Stellen nicht gibt.
//
// **So fragt die Naht stattdessen nach**, und die Antwort steht in
// [0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md):
//
//   - **OpenAI** kann es und gibt Bytes.
//   - **Ollama** kann es über HTTP **gar nicht** — der Blogbeitrag kündigt
//     einen Befehl an, die Dokumentation kennt weder Weg noch
//     Fähigkeitszeichenkette.
//   - **xAI** kann es und gibt eine **Adresse**, und die nimmt dieses
//     Programm nicht an.
//   - **Anthropic** liest Bilder und erzeugt keine.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIBildquelle ist, was eine Laufzeit können muss, um ein Bild zu erzeugen.
//
// **Eine Laufzeit erfüllt sie oder nicht** — geprüft wird mit einer
// Typzusicherung, nicht mit einer Liste daneben. Eine Liste liefe mit den
// Nähten auseinander.
type KIBildquelle interface {
	Bild(ctx context.Context, wunsch KIBildwunsch) (KIRohbild, error)
}

// KIBildwunsch ist, was die Naht an den Anbieter gibt.
type KIBildwunsch struct {
	Alias   string
	Auftrag string
	Groesse string
	Guete   string

	// Quelle gesetzt heißt: **bearbeiten** statt erzeugen.
	Quelle []byte
}

// KIRohbild ist, was zurückkam.
type KIRohbild struct {
	Inhalt    []byte
	Medientyp string
	Modell    string

	// Umschrift ist der Prompt, wie der Dienst ihn umgeschrieben hat.
	Umschrift string
}

// bildquelleAus sagt, ob dieser Anbieter Bilder erzeugt.
//
// **Die Meldung nennt den Anbieter**, damit niemand rät, welcher von vieren
// gemeint ist — und sie sagt, was stattdessen geht.
func bildquelleAus(quelle KILaufzeitquelle, anbieter string) (KIBildquelle, error) {
	bildquelle, kann := quelle.(KIBildquelle)
	if !kann {
		return nil, fmt.Errorf(
			"%w: %q erzeugt keine Bilder — dafür gibt es einen Anbieter, "+
				"dessen Modelle die Fähigkeit %q belegen",
			domain.ErrValidation, anbieter, domain.KIBildgenerierung)
	}

	return bildquelle, nil
}
