package service

// **Der Katalog der KI-Modelle** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Was er beantwortet
//
// „Wer kann diese Fähigkeit" — und **nur mit Beleg**. Das Issue sagt es
// scharf: „Unbekannt bleibt unbekannt; Modelllisten garantieren keine
// Capability." Ein Name in einem Katalog heißt, dass es das Modell gibt.
//
// # Und was er bemerkt
//
// **Wenn ein Alias hinter demselben Namen etwas anderes geworden ist.** Das
// ist der Fall, den man sonst nicht sieht: der Aufruf bleibt gleich, die
// Antwort ändert sich. Verglichen werden [domain.KIModell.Bindung]en und
// nicht Namen.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Modellkatalog führt die Modelle, die feststehen.
type Modellkatalog struct{ modelle []domain.KIModell }

// NeuerModellkatalog baut ihn — und weist ab, was nicht vollständig ist.
//
// **Beim Bauen und nicht beim Fragen**: ein Katalog, der ungültige Einträge
// führt und sie erst beim Zugriff prüft, ist eine Liste mit einer Falle.
func NeuerModellkatalog(modelle ...domain.KIModell) (*Modellkatalog, error) {
	gesehen := map[string]bool{}

	for _, modell := range modelle {
		if befunde := modell.Validate(); len(befunde) > 0 {
			return nil, fmt.Errorf("%w: das Modell %q ist unvollständig: %s",
				domain.ErrValidation, modell.Alias, befunde[0].Error())
		}

		schluessel := schluesselVon(modell)

		if gesehen[schluessel] {
			return nil, fmt.Errorf("%w: %q steht zweimal im Katalog",
				domain.ErrConflict, schluessel)
		}

		gesehen[schluessel] = true
	}

	kopie := make([]domain.KIModell, len(modelle))
	copy(kopie, modelle)

	return &Modellkatalog{modelle: kopie}, nil
}

// Alle gibt die Modelle in Katalogreihenfolge.
func (k *Modellkatalog) Alle() []domain.KIModell {
	kopie := make([]domain.KIModell, len(k.modelle))
	copy(kopie, k.modelle)

	return kopie
}

// Fuer gibt die Modelle, die eine Fähigkeit **belegt** haben.
//
// **Eine leere Liste ist eine Auskunft** und kein Fehler: „das kann gerade
// keiner" ist etwas anderes als „da ging etwas schief".
func (k *Modellkatalog) Fuer(faehigkeit domain.KIFaehigkeit) []domain.KIModell {
	var passende []domain.KIModell

	for _, modell := range k.modelle {
		if modell.Kann(faehigkeit) {
			passende = append(passende, modell)
		}
	}

	return passende
}

// Bindungen gibt je Modell seine Bindung — zum Aufheben für das nächste Mal.
func (k *Modellkatalog) Bindungen() map[string]string {
	bindungen := make(map[string]string, len(k.modelle))

	for _, modell := range k.modelle {
		bindungen[schluesselVon(modell)] = modell.Bindung()
	}

	return bindungen
}

// Wechsel nennt die Modelle, deren Bindung sich geändert hat.
//
// **Ein Modell, das vorher nicht dastand, ist kein Wechsel**, sondern ein
// Zugang — und der Unterschied zählt: wer ein Modell aufnimmt, weiß es; wer
// eines unter der Hand getauscht bekommt, nicht.
func (k *Modellkatalog) Wechsel(vorher map[string]string) []domain.Befund {
	var befunde []domain.Befund

	for _, modell := range k.modelle {
		schluessel := schluesselVon(modell)

		alte, gabs := vorher[schluessel]
		if !gabs || alte == modell.Bindung() {
			continue
		}

		befunde = append(befunde, domain.Befund{
			Datei: schluessel, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Hinter %q liegt jetzt etwas anderes: der Name ist derselbe, "+
					"das ausgeführte Artefakt nicht. Was zuletzt geprüft wurde, "+
					"gilt für dieses Modell nicht mehr.", modell.Alias),
		})
	}

	return befunde
}

// schluesselVon macht aus Anbieter und Alias den Katalogschlüssel.
//
// **Beides zusammen**: zwei Anbieter dürfen denselben Alias führen, und sie
// meinen dann verschiedene Modelle.
func schluesselVon(modell domain.KIModell) string {
	return modell.Anbieter + "/" + modell.Alias
}
