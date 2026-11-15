// **Die Sammlung am Fenster** — Werkstück 2 von Phase 3.4
// (#1165).
//
// Fünf Bindungen: aufzählen, nachschlagen, anlegen, ändern, löschen. Die
// Fachlichkeit steht in `service.Sammlungspflege`; hier wird übersetzt (§2).

package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungsquelle ist, was die Sammlungsseite braucht.
type Sammlungsquelle interface {
	Liste(ctx context.Context) ([]*domain.Exemplar, error)
	Fuer(ctx context.Context, printing string) ([]*domain.Exemplar, error)
	Anlegen(ctx context.Context, exemplar domain.Exemplar) (int64, error)
	Aendern(ctx context.Context, exemplar domain.Exemplar) error
	Loeschen(ctx context.Context, id, fassung int64) error
}

// ExemplarDTO ist ein Exemplar, wie das Fenster es sieht.
//
// **Der Erwerbspreis reist als Zeichenkette.** Im Bestand steht er in
// der kleinsten Einheit als Ganzzahl; JSON kennt nur `number`, und eine Zahl mit
// Nachkommastellen ist dort ein `float64` — genau das, was
// [domain.GeldAus] verhindert. Was hier steht, ist die Form, in der
// jemand ihn tippt: `"4.17"`.
type ExemplarDTO struct {
	ID int64 `json:"id"`

	// Fassung ist die Fassung, die gelesen wurde.
	//
	// **Sie geht beim Ändern zurück** — stimmt sie dann nicht mehr, hat jemand
	// anderes geschrieben, und das Fenster erfährt es, statt die fremde
	// Änderung zu überschreiben (§7, Regel 2).
	Fassung int64 `json:"fassung"`

	Printing        string `json:"printing"`
	Menge           int    `json:"menge"`
	Zustand         string `json:"zustand"`
	Sprache         string `json:"sprache"`
	Finish          string `json:"finish"`
	Erwerbsdatum    string `json:"erwerbsdatum"`
	Erwerbspreis    string `json:"erwerbspreis"`
	Erwerbswaehrung string `json:"erwerbswaehrung"`
	Notiz           string `json:"notiz"`

	// DruckID ist die Zeile des Drucks in **diesem** Bestand — Werkstück A2
	// von Phase 5.3 (#1441, 0086). **Als Zeichenkette wie jede Kennung an
	// dieser Naht**, und leer für „keine": ein `int64` verlöre in JavaScript
	// oberhalb von 2⁵³ Stellen. Sie verlässt den Bestand nie — eine Ausfuhr
	// trägt die Referenz, nicht die Nummer.
	DruckID string `json:"druckID"`

	// Vorlaeufig sagt, dass der Druck **keinen Alias** hat: die Zeile kam aus
	// einer Liste ohne Scryfall-Kennung, und Preise wie Bilder bleiben aus,
	// bis jemand sie auflöst. Ein Zustand, kein Fehler.
	Vorlaeufig bool `json:"vorlaeufig"`

	// Name, Setcode und Nummer sind der **lesbare Bezug** des Drucks
	// (#1200).
	//
	// **Der letzte gute Stand und nicht die Identität** — die ist
	// [ExemplarDTO.Printing]. Ohne Netz steht hier, was zuletzt bekannt war
	// (§16, Regel 5); ohne Angabe bleibt die Kennung.
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`

	// Lagerort ist, wo das Stück Papier liegt — Freitext
	// ([0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
	Lagerort string `json:"lagerort"`

	// Wunsch sagt, dass diese Zeile **kein Besitz** ist (0057).
	Wunsch bool `json:"wunsch"`

	// Wunschrang ist `niedrig`, `normal`, `hoch` oder `sehrhoch` — **leer
	// heißt `normal`**, und an einem Besitz bedeutet er nichts (T4, #1469).
	Wunschrang string `json:"wunschrang"`

	// Wunschart ist `genau` oder `jederdruck` — **leer heißt `genau`**.
	Wunschart string `json:"wunschart"`

	// Bezeichnung ist der Bezug, wie eine Fläche ihn zeigt — oder die Kennung.
	//
	// **Gerechnet und nicht im Frontend zusammengesetzt**: „Lightning Bolt ·
	// M11 #149" ist eine fachliche Zusammenfassung, und drei Flächen, die sie
	// je für sich bauen, liefen auseinander (§16).
	Bezeichnung string `json:"bezeichnung"`

	// Herkunft sagt, woher diese Zeile kam — Werkstück 12 (#1351).
	//
	// **Sie reist mit jeder Zeile**, weil sie über einen `LEFT JOIN` kommt und
	// nichts kostet; ein zweiter Ruf je Zeile wäre ein Ruf je Zeile.
	Herkunft HerkunftDTO `json:"herkunft"`
}

// Sammlungsantwort trägt die Exemplare.
type Sammlungsantwort struct {
	Exemplare []ExemplarDTO `json:"exemplare"`
	Fehler    *Fehler       `json:"fehler"`
}

// Exemplarantwort trägt die Kennung eines angelegten Exemplars.
type Exemplarantwort struct {
	ID     int64   `json:"id"`
	Fehler *Fehler `json:"fehler"`

	// Vorhanden ist die Zeile, die es schon gibt, wenn die Aufnahme an ihr
	// scheiterte — DOM-03, S04.NAT-01 und 03 (#1698). **Neben dem Fehler und
	// nicht an seiner Stelle**: wer sie nicht kennt, sieht die Absage wie
	// vorher; das Fenster bietet an, ihre Menge zu erhöhen oder zu ihr zu
	// springen.
	Vorhanden *ExemplarDTO `json:"vorhanden"`
}

// MitSammlung setzt die Naht.
func (a *App) MitSammlung(quelle Sammlungsquelle) { a.sammlung = quelle }

// Sammlung gibt die ganze Sammlung.
func (a *App) Sammlung(ctx context.Context) Sammlungsantwort {
	ctx = a.imProfil(ctx)

	return a.sammlungsliste(ctx, func() ([]*domain.Exemplar, error) {
		return a.sammlung.Liste(ctx)
	})
}

// SammlungFuer gibt die Exemplare eines Drucks.
func (a *App) SammlungFuer(ctx context.Context, printing string) Sammlungsantwort {
	ctx = a.imProfil(ctx)

	return a.sammlungsliste(ctx, func() ([]*domain.Exemplar, error) {
		return a.sammlung.Fuer(ctx, printing)
	})
}

// sammlungsliste beantwortet beide Listenwege.
//
// **Einmal geschrieben.** Zwei Fassungen derselben acht Zeilen liefen an dem
// Tag auseinander, an dem eine von ihnen etwas dazulernt.
func (a *App) sammlungsliste(ctx context.Context, holen func() ([]*domain.Exemplar, error)) Sammlungsantwort {
	if a.sammlung == nil {
		return Sammlungsantwort{
			Exemplare: []ExemplarDTO{}, Fehler: Uebersetze(ctx, keineSammlung()),
		}
	}

	exemplare, err := holen()
	if err != nil {
		return Sammlungsantwort{Exemplare: []ExemplarDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return Sammlungsantwort{Exemplare: zuExemplarDTOs(exemplare)}
}

// ExemplarAnlegen nimmt ein Exemplar auf.
func (a *App) ExemplarAnlegen(ctx context.Context, dto ExemplarDTO) Exemplarantwort {
	ctx = a.imProfil(ctx)

	if a.sammlung == nil {
		return Exemplarantwort{Fehler: Uebersetze(ctx, keineSammlung())}
	}

	exemplar, err := ausExemplarDTO(dto)
	if err != nil {
		return Exemplarantwort{Fehler: Uebersetze(ctx, err)}
	}

	id, anlegefehler := a.sammlung.Anlegen(ctx, exemplar)
	if anlegefehler != nil {
		antwort := Exemplarantwort{Fehler: Uebersetze(ctx, anlegefehler)}

		var dublette *domain.Dublette
		if errors.As(anlegefehler, &dublette) {
			vorhanden := zuExemplarDTO(&dublette.Vorhanden)
			antwort.Vorhanden = &vorhanden
		}

		return antwort
	}

	return Exemplarantwort{ID: id}
}

// ExemplarAendern schreibt ein Exemplar zurück.
func (a *App) ExemplarAendern(ctx context.Context, dto ExemplarDTO) *Fehler {
	ctx = a.imProfil(ctx)

	if a.sammlung == nil {
		return Uebersetze(ctx, keineSammlung())
	}

	exemplar, err := ausExemplarDTO(dto)
	if err != nil {
		return Uebersetze(ctx, err)
	}

	return Uebersetze(ctx, a.sammlung.Aendern(ctx, exemplar))
}

// ExemplarLoeschen nimmt ein Exemplar aus der Sammlung.
func (a *App) ExemplarLoeschen(ctx context.Context, id, fassung int64) *Fehler {
	ctx = a.imProfil(ctx)

	if a.sammlung == nil {
		return Uebersetze(ctx, keineSammlung())
	}

	return Uebersetze(ctx, a.sammlung.Loeschen(ctx, id, fassung))
}

// keineSammlung meldet die fehlende Naht.
func keineSammlung() error {
	return fmt.Errorf("%w: die Sammlung ist nicht eingerichtet", domain.ErrValidation)
}
