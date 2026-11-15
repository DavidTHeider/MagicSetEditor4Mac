package bezug

// **Der ganze Baum eines Commits** — G04
// (#2251).
//
// # Wozu, wenn es den Vergleich gibt
//
// **Der Vergleich nennt höchstens 300 Dateien** — für die ganze Spanne, und
// weitere Seiten blättern die Commits, nicht die Dateien. Ein Vorlagenpack,
// das 63 Commits hinter seiner Spitze steht, ändert mehr; seine Vorschau
// nennte dann nur einen Teil der Pakete und der Sets, die sie nennen.
//
// `GET /repos/{owner}/{repo}/git/trees/{sha}?recursive=1` gibt dagegen
// **jede** Datei eines Commits mit Pfad und Kennung, bis 100.000 Einträge
// oder 7 MB; was darüber liegt, meldet `truncated`. Nachgelesen am
// 2026-10-11 in GitHubs REST-Dokumentation („Git trees" und „Commits", API-Fassung
// 2022-11-28) und gemessen am selben Tag: `Full-Magic-Pack` am geprüften
// Stand hat 36.571 Einträge in 10,7 MB Antwort, `M15-Magic-Pack` 31.820
// in 9,4 MB — beide ungekürzt.

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Baumstand sind die Dateien eines Commits — Pfad und Kennung des Inhalts.
type Baumstand struct {
	Dateien map[string]string

	// Gekuerzt sagt, dass die Gegenstelle nicht jede Datei genannt hat.
	Gekuerzt bool
}

// Baum fragt nach jeder Datei eines Commits.
//
// `adresse` ist die volle Adresse der Auskunft samt `recursive`; welche das
// ist, weiß der Aufrufer, wie bei [Spitze].
func Baum(ctx context.Context, adresse, kenner string) (Baumstand, error) {
	var antwort struct {
		Truncated bool `json:"truncated"`
		Tree      []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			SHA  string `json:"sha"`
		} `json:"tree"`
	}

	if fehler := auskunft(ctx, adresse, kenner, &antwort); fehler != nil {
		return Baumstand{}, fehler
	}

	// **Pflichtfelder streng** (§16): ohne Liste ist die Antwort keine
	// Auskunft über einen Baum, und eine Datei ohne Pfad keine Datei.
	if antwort.Tree == nil {
		return Baumstand{}, domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"%s nennt keinen Baum.", adresse)
	}

	baum := Baumstand{Dateien: make(map[string]string, len(antwort.Tree)), Gekuerzt: antwort.Truncated}

	for _, eintrag := range antwort.Tree {
		if eintrag.Path == "" {
			return Baumstand{}, domain.NewFieldError("adresse", domain.SchluesselValidierung,
				"%s nennt eine Datei ohne Pfad.", adresse)
		}

		// **Nur Dateien**: ein Verzeichnis ändert sich mit seinen Dateien,
		// und ein eingebundenes Repositorium (`commit`) bringt das Archiv
		// nicht mit.
		if eintrag.Type == "blob" {
			baum.Dateien[eintrag.Path] = eintrag.SHA
		}
	}

	return baum, nil
}

// Baumvergleich sagt, welche Dateien zwischen zwei Bäumen hinzukommen,
// wegfallen oder sich ändern — mit GitHubs Wörtern, nach Pfad geordnet.
func Baumvergleich(alt, neu Baumstand) []Aenderung {
	aenderungen := make([]Aenderung, 0)

	for pfad, kennung := range neu.Dateien {
		vorher, da := alt.Dateien[pfad]

		switch {
		case !da:
			aenderungen = append(aenderungen, Aenderung{Datei: pfad, Art: "added"})
		case vorher != kennung:
			aenderungen = append(aenderungen, Aenderung{Datei: pfad, Art: "modified"})
		}
	}

	for pfad := range alt.Dateien {
		if _, bleibt := neu.Dateien[pfad]; !bleibt {
			aenderungen = append(aenderungen, Aenderung{Datei: pfad, Art: "removed"})
		}
	}

	slices.SortFunc(aenderungen, func(x, y Aenderung) int { return strings.Compare(x.Datei, y.Datei) })

	return aenderungen
}
