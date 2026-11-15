package service

// **Der Riegel vor den verbindlichen Schritten eines Handels** — Werkstück
// `T10` von Phase 5.6
// (#1475).
//
// **Eine eigene Datei neben `handelsverbuchung.go`**, damit jene unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, **was**
// eine Verbuchung tut, hier, **wer sie tun darf**.
//
// # Warum es dafür einen Nachweis gibt und nicht nur eine Rolle
//
// **Weil eine Rolle sagt, was jemand grundsätzlich darf, und ein Nachweis,
// dass er dieses eine wollte.** Eine Verbuchung bewegt echtes Papier; sie
// hängt an einem Vorgang und an einem Stand, und beide reisen in der Bindung
// mit ([domain.Nachweis]). Ein Feld `bestaetigt`, das der Aufrufer selbst
// setzt, beweist nur, dass er das Feld kennt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// HandelsverbuchungZweck ist der Zweck, unter dem ein Nachweis für eine
// Verbuchung ausgestellt und eingelöst wird (T10, #1475).
//
// **Zwei Zwecke und nicht einer**: wer das Verbuchen freigibt, hat den
// Bestandsübergang gewollt — nicht die zweite Unterschrift, die danach kommt.
// Ein gemeinsamer Zweck ließe das eine Ja für beides gelten.
const (
	HandelsverbuchungZweck = "handel.verbuchen"
	HandelsabschlussZweck  = "handel.bestaetigen"
)

// Verbuchungswache fragt, ob dieser Aufruf den Schritt tun darf.
//
// **Erfüllt von [Autorisierung]** — die Schnittstelle steht hier, weil dieser
// Anwendungsfall sie braucht (§2), und sie hat genau **ein** Verfahren: die
// Frage „darfst du das, und hast du es gewollt?". Den Akteur braucht diese
// Stelle nicht; wem der Bestand gehört, entscheidet die Verbindung.
type Verbuchungswache interface {
	PruefeMitNachweis(ctx context.Context, wirkung domain.Wirkung,
		kennung string, bindung Nachweisbindung) error
}

// MitWache setzt den Riegel vor die verbindlichen Schritte (T10, #1475).
//
// **Wahlweise im Bau und nicht im Ruf**: die Stelle entsteht an vier Flächen
// und in vielen Läufen; ein Pflichtargument hieße, jede Zeile anzufassen, die
// den Riegel gar nicht misst. **Ohne Wache wird nicht verbucht** — der
// Rückfall ist das Nein, und ein Lauf betritt ihn.
func (v *Verbuchungsstelle) MitWache(wache Verbuchungswache) *Verbuchungsstelle {
	v.wache = wache

	return v
}

// darf fragt die Wache — oder sagt nein, wenn es keine gibt.
func (v *Verbuchungsstelle) darf(
	ctx context.Context, wirkung domain.Wirkung, nachweis, zweck, ziel string,
	revision int,
) error {
	if v.wache == nil {
		return domain.MitGrund(domain.ErrErlaubnis, nil,
			"diese Fläche hat keinen Riegel vor der Verbuchung — ohne ihn "+
				"wird nicht gebucht")
	}

	return v.wache.PruefeMitNachweis(ctx, wirkung, nachweis, Nachweisbindung{
		Zweck: zweck, Ziel: ziel, Revision: int64(revision),
	})
}

// NeueVerbuchungsstelle baut sie.
//
// **Die ganze [Handelsablage] und nicht nur der Stand**: Abgleichen und
// Bestätigen bewegen den Bestand nicht und müssen ihren Zustand trotzdem
// ablegen — sonst stünde der Vorgang nach einem Neustart wieder dort, wo er
// vor dem Schritt stand.
func NeueVerbuchungsstelle(
	stand Handelsablage, bestand Verbuchungsbestand, ausweis Ausweisstelle,
	drucke Druckbezeichnungen,
) *Verbuchungsstelle {
	return &Verbuchungsstelle{
		stand: stand, bestand: bestand, ausweis: ausweis, drucke: drucke,
	}
}

// Handelsfreigabeziel benennt, wofür ein Nachweis gilt.
//
// **Der Vorgang allein reicht für `bestaetigen` nicht** (Befund 5 des
// Zweitblicks über Phase 5.6): dieser Schritt setzt die Stimme einer
// **benannten** Partei, und ein Nachweis, der nur den Vorgang nennt, gilt
// für jede. Ein Mensch gäbe einmal frei, und derselbe Nachweis setzte die
// Unterschrift des Gegenübers.
//
// **Ohne Partei bleibt es der Vorgang** — `verbuchen` und `abgleichen`
// setzen keine Stimme.
func Handelsfreigabeziel(vorgang, wer string) string {
	if wer == "" {
		return vorgang
	}

	return vorgang + "#" + wer
}
