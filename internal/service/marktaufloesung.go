package service

// **Welcher Blueprint zu welchem Druck gehört** (A4, #1443).
//
// # Die Regel kommt aus A2
//
// „Mehrdeutige oder widersprüchliche Cross-IDs als **Auflösungskonflikt**,
// kein Namens-Automerge"
// (#1441).
// Zwei Fassungen mit derselben Scryfall-Kennung heißen: der Anbieter führt
// etwas doppelt, oder seine Kennung meint etwas anderes als hier. Beides ist
// eine Frage an einen Menschen und keine, die ein Programm still beantwortet.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fassungFuer findet die Kartenfassung eines Drucks beim Anbieter.
//
// **Die gemerkte zuerst**: die Blueprints einer Erweiterung sind einige
// Megabyte, und die Zuordnung ändert sich nicht.
func (m *Marktabruf) fassungFuer(ctx context.Context, druck *domain.Druck) (
	Marktergebnis, error,
) {
	if gemerkt, gibts := druck.Kennung(
		m.quelle.System(), domain.SemantikDruck); gibts {
		return Marktergebnis{Fassung: gemerkt}, nil
	}

	// **Der Alias steht fest**: gefunden wurde der Druck über ihn
	// ([Marktregister]). Ein Druck ohne ihn erreicht diese Stelle nicht, und
	// ein Zweig dafür wäre einer, den keine Fläche betritt (§3).
	alias := druck.Scryfall()

	gesucht := domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: alias,
	}

	fassungen, befunde, fehler := m.quelle.Fassungen(ctx, druck.Setcode)
	if fehler != nil {
		return Marktergebnis{}, fehler
	}

	var treffer []domain.Fremdfassung

	for _, fassung := range fassungen {
		if fassung.Nennt(gesucht) {
			treffer = append(treffer, fassung)
		}
	}

	switch len(treffer) {
	case 0:
		return Marktergebnis{}, fmt.Errorf(
			"%w: %s führt in %q keine Fassung mit der Kennung %s",
			domain.ErrNotFound, m.quelle.Quelle(), druck.Setcode, alias)
	case 1:
	default:
		return Marktergebnis{}, fmt.Errorf(
			"%w: %s führt in %q %d Fassungen mit der Kennung %s — welche "+
				"gemeint ist, entscheidet dieses Programm nicht",
			domain.ErrConflict, m.quelle.Quelle(), druck.Setcode, len(treffer), alias)
	}

	ergebnis := Marktergebnis{
		Fassung: treffer[0].Kennung, Zugeordnet: true, Befunde: befunde,
	}

	// **Geschrieben ist geschrieben, und gemerkt ist Bestenfalls-Arbeit**
	// (§7, Regel 3): die Zuordnung spart den nächsten Abruf. Wer sie zur
	// Bedingung machte, ließe einen Schreibfehler eine Auskunft verhindern,
	// die schon dasteht.
	if fehler := m.merken(ctx, druck, treffer[0].Kennung); fehler != nil {
		ergebnis.Befunde = append(ergebnis.Befunde, domain.Befund{
			Datei: druck.Name, Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Die Zuordnung zu %s ließ sich nicht merken: %v. "+
					"Der nächste Abruf sucht sie wieder.",
				m.quelle.Quelle(), fehler),
		})
	}

	return ergebnis, nil
}

// merken schreibt die gefundene Kennung an den Druck.
func (m *Marktabruf) merken(ctx context.Context, druck *domain.Druck, fassung string) error {
	druck.Kennungen = append(druck.Kennungen, domain.Druckkennung{
		System:   m.quelle.System(),
		Semantik: domain.SemantikDruck,
		Wert:     fassung,
	})

	return m.drucke.Sichern(ctx, druck)
}
