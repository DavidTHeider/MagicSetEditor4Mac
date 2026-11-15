package peer

// **Was ein Handschlag prüft, bevor er weitergeht** — Werkstück N1 von Phase
// 5.4 (#1454).

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// lesAngebot packt eine Nachricht aus und liest ein Angebot daraus.
func lesAngebot(roh []byte) (Huelle, Angebot, error) {
	huelle, fehler := Auspacke(roh)
	if fehler != nil {
		return Huelle{}, Angebot{}, fehler
	}

	if huelle.Art != ArtAngebot {
		return Huelle{}, Angebot{}, fmt.Errorf(
			"%w: hier steht ein Angebot und kein %q", domain.ErrValidation,
			huelle.Art)
	}

	var angebot Angebot

	if fehler := json.Unmarshal(huelle.Rumpf, &angebot); fehler != nil {
		return Huelle{}, Angebot{}, domain.MitGrund(
			domain.ErrValidation, fehler, "der Rumpf ist kein Angebot")
	}

	return huelle, angebot, nil
}

// pruefeGegenueber hält das andere Ende gegen das, was erwartet wird.
//
// **Mit sich selbst gibt es keinen Handschlag.** Er belegte nichts — und er
// fiele durch jede Prüfung, weil beide Seiten dieselbe Zufallszahl vermerken
// wollten.
func pruefeGegenueber(ich, gegenueber, erwartet Kennung) error {
	if gegenueber == ich {
		return fmt.Errorf(
			"%w: mit sich selbst gibt es keinen Handschlag",
			domain.ErrValidation)
	}

	if erwartet != "" && erwartet != gegenueber {
		return fmt.Errorf(
			"%w: erwartet war %s, gemeldet hat sich %s",
			domain.ErrValidation, Fingerabdruck(erwartet),
			Fingerabdruck(gegenueber))
	}

	return nil
}

// pruefeFassung ist die festgelegte Ablehnung aus der Abnahme.
func pruefeFassung(fassung int) error {
	if fassung != Handschlagfassung {
		return fmt.Errorf(
			"%w: dieser Handschlag hat die Fassung %d, angeboten wurde %d",
			domain.ErrValidation, Handschlagfassung, fassung)
	}

	return nil
}

// pruefeAngebot hält ein gelesenes Angebot gegen die eigene Bereitschaft.
func pruefeAngebot(
	huelle Huelle, angebot Angebot, bereit Bereitschaft,
) (string, error) {
	if fehler := pruefeFassung(angebot.Fassung); fehler != nil {
		return "", fehler
	}

	protokoll, gilt := dieZwecke[angebot.Zweck]
	if !gilt {
		return "", fmt.Errorf(
			"%w: %q ist kein Zweck dieses Programms", domain.ErrValidation,
			angebot.Zweck)
	}

	if protokoll != huelle.Protokoll {
		return "", fmt.Errorf(
			"%w: der Zweck %q gehört zu %s und nicht zu %s",
			domain.ErrValidation, angebot.Zweck, protokoll, huelle.Protokoll)
	}

	if angebot.Rolle == "" || bereit.Rolle == "" {
		return "", fmt.Errorf(
			"%w: beide Seiten nennen ihre Rolle", domain.ErrValidation)
	}

	if fehler := pruefeKanal(angebot.Kanal, bereit.Kanal); fehler != nil {
		return "", fehler
	}

	if fehler := pruefeZufallszahl(angebot.Nonce); fehler != nil {
		return "", fehler
	}

	// **Die Hülle nennt dieselbe Sitzung wie der Rumpf.** Solange es die
	// Sitzungskennung noch nicht gibt — sie entsteht erst aus beiden
	// Zufallszahlen —, ist die des Anbieters ihr Platzhalter. Zwei
	// verschiedene Angaben wären zwei Sitzungen in einer Nachricht.
	if huelle.Sitzung != angebot.Nonce {
		return "", fmt.Errorf(
			"%w: Kopf und Rumpf nennen zwei verschiedene Sitzungen",
			domain.ErrValidation)
	}

	return protokoll, nil
}

// pruefeKanal hält die Bindung des Transports gegen die eigene.
//
// **Ohne sie ließe sich ein Handschlag weiterreichen**: wer ihn auf einer
// Verbindung mitschneidet und auf einer anderen einspielt, hätte eine Sitzung,
// die zwei andere Peers gebaut haben.
func pruefeKanal(dort, hier string) error {
	if hier == "" {
		return fmt.Errorf(
			"%w: ohne Kanalbindung gibt es keinen Handschlag",
			domain.ErrValidation)
	}

	if dort != hier {
		return fmt.Errorf(
			"%w: dieser Handschlag gehört zu einem anderen Kanal",
			domain.ErrValidation)
	}

	return nil
}
