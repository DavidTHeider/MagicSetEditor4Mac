package service

// **Die portable Referenz kommt an** — Werkstück A2 von Phase 5.3
// (#1441, 0086).
//
// Eine CSV-Zeile nennt ihren Druck über die Referenz — System, Kennung,
// Rückfall — oder, wie vor 0.5.0, nur über den Alias. Hier wird daraus die
// **Zeile dieses Bestands**: gefunden über die Kennung, sonst über den
// vollständigen Rückfall; angelegt, wenn es sie nicht gibt; und ein Konflikt,
// wenn mehrere passen. Eine lokale Nummer des Senders spielt dabei nie eine
// Rolle — die Ausfuhr trägt keine.

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// Druckregister ist, was die Einfuhr von der Druckablage braucht.
type Druckregister interface {
	NachKennung(ctx context.Context, kennung domain.Druckkennung) ([]domain.Druck, error)
	Kandidaten(ctx context.Context, name, setcode, nummer string) ([]domain.Druck, error)
	Sichern(ctx context.Context, druck *domain.Druck) error
	Fuer(ctx context.Context, id int64) (*domain.Druck, error)
}

// Druckzuordnung ist das Ergebnis für eine Zeile: die Zeile des Drucks und
// sein Alias — oder die Meldung, warum es keine gibt.
type Druckzuordnung struct {
	// Neu sagt, dass der Druck beim Einlesen **angelegt würde** — nur die
	// Vorschau setzt es, denn sie legt nichts an (DOM-03.AK-01, #1698).
	Neu bool

	DruckID int64
	Alias   string
	Meldung string
}

// verweisAus liest die Referenz einer Zeile — mit dem Alias als Kennung, wo
// die Zeile keine eigene nennt.
func verweisAus(zeile sammelliste.Ausfuhrzeile) domain.Druckverweis {
	verweis, _ := verweisUndHerkunft(zeile)

	return verweis
}

// nurAlias sagt, ob die Kennung der Zeile allein aus `printing` stammt.
//
// **Dann bleibt die Einfuhr so nachsichtig wie vor 0.5.0**: ein unbekannter
// Alias wird nicht hier geprüft und angelegt, sondern von der Ablage beim
// Schreiben — sie kennt die Spalte, und sie hat ihn vor 0.5.0 auch genommen.
// Eine **ausdrückliche** Referenz dagegen wird geprüft und angelegt, und
// eine unbrauchbare ist ein Befund der Zeile.
func nurAlias(zeile sammelliste.Ausfuhrzeile) bool {
	_, ausAlias := verweisUndHerkunft(zeile)

	return ausAlias
}

func verweisUndHerkunft(zeile sammelliste.Ausfuhrzeile) (domain.Druckverweis, bool) {
	verweis := domain.Druckverweis{
		Fassung: zeile.Verweisfassung,
		Name:    zeile.Name, Setcode: zeile.Setcode, Nummer: zeile.Nummer, Sprache: zeile.Sprache,
	}

	switch {
	case zeile.Kennungssystem != "" && zeile.Kennung != "":
		verweis.System = domain.Kennungssystem(strings.ToLower(strings.TrimSpace(zeile.Kennungssystem)))
		verweis.Semantik = domain.SemantikDruck
		verweis.Wert = strings.TrimSpace(zeile.Kennung)

		// **Eine Referenz, die nur den Alias wiederholt, sagt nichts, was er
		// nicht schon sagt** — so schreibt die eigene Ausfuhr jede Zeile mit
		// Alias, und der Rundlauf bleibt so nachsichtig wie die Ablage.
		return verweis, verweis.System == domain.SystemScryfall && zeile.Printing != "" && verweis.Wert == zeile.Printing
	case zeile.Printing != "":
		verweis.System, verweis.Semantik, verweis.Wert = domain.SystemScryfall, domain.SemantikDruck, zeile.Printing

		return verweis, true
	}

	return verweis, false
}

// druckZuordnen findet oder legt den Druck einer Zeile an.
//
// **Ein Fehler des Registers ist ein Systemfehler** und kein Befund je Zeile —
// derselbe Schnitt wie bei [Listeneinfuhr.Einlesen].
func (s *Sammlungsausfuhr) druckZuordnen(
	ctx context.Context, zeile sammelliste.Ausfuhrzeile, anlegen bool,
) (Druckzuordnung, error) {
	verweis := verweisAus(zeile)

	if befunde := verweis.Validate(); len(befunde) > 0 {
		// **Die Fassung zuerst** — eine Zeile ohne alles ist kein Befund der
		// Referenz, sondern einer des Exemplars, und den gibt `vorbereiten`.
		if befunde[0].Field == "fassung" {
			return Druckzuordnung{Meldung: befunde[0].Message}, nil
		}

		return Druckzuordnung{}, nil
	}

	kandidaten, err := s.kandidaten(ctx, verweis)
	if err != nil {
		return Druckzuordnung{}, err
	}

	aufloesung := domain.LoeseAuf(kandidaten, verweis)

	switch aufloesung.Art {
	case domain.AufloesungEindeutig:
		druck, fehler := s.drucke.Fuer(ctx, aufloesung.DruckIDs[0])
		if fehler != nil {
			return Druckzuordnung{}, fehler
		}

		return Druckzuordnung{DruckID: druck.ID, Alias: druck.Scryfall()}, nil

	case domain.AufloesungMehrdeutig:
		return Druckzuordnung{Meldung: fmt.Sprintf(
			"die Referenz trifft %d Drucke — welcher gemeint ist, entscheidet niemand von selbst",
			len(aufloesung.DruckIDs))}, nil
	}

	if nurAlias(zeile) {
		return Druckzuordnung{}, nil
	}

	return s.druckAnlegen(ctx, verweis, anlegen)
}

// kandidaten holt, was die Kennung und was der Rückfall treffen — ohne Doppel.
func (s *Sammlungsausfuhr) kandidaten(
	ctx context.Context, verweis domain.Druckverweis,
) ([]domain.Druck, error) {
	var gefunden []domain.Druck

	if kennung, hatKennung := verweis.Kennung(); hatKennung {
		ueberKennung, err := s.drucke.NachKennung(ctx, kennung)
		if err != nil {
			return nil, err
		}

		gefunden = append(gefunden, ueberKennung...)
	}

	ueberRueckfall, err := s.drucke.Kandidaten(ctx, verweis.Name, verweis.Setcode, verweis.Nummer)
	if err != nil {
		return nil, err
	}

	bekannt := make(map[int64]bool, len(gefunden))
	for _, druck := range gefunden {
		bekannt[druck.ID] = true
	}

	for _, druck := range ueberRueckfall {
		if !bekannt[druck.ID] {
			gefunden = append(gefunden, druck)
		}
	}

	return gefunden, nil
}

// druckAnlegen legt den unbekannten Druck an — **mit seiner Kennung**, damit
// die nächste Einfuhr ihn findet, und mit dem Rückfall als Anzeige.
//
// **Ohne Kennung und ohne vollständigen Rückfall entsteht nichts**: eine Zeile
// mit bloßem Namen kann dreißig Drucke meinen, und ein Druck, der nur einen
// Namen trägt, wäre keiner.
//
// **Ohne `anlegen` sagt es nur, dass es einen anlegte** — die Vorschau legt
// nichts an (DOM-03.AK-01, #1698), und ein Druck, den sie anlegte, stünde
// danach im Register, auch wenn niemand einliest.
func (s *Sammlungsausfuhr) druckAnlegen(
	ctx context.Context, verweis domain.Druckverweis, anlegen bool,
) (Druckzuordnung, error) {
	kennung, hatKennung := verweis.Kennung()
	rueckfall := strings.TrimSpace(verweis.Name) != "" && strings.TrimSpace(verweis.Setcode) != "" &&
		strings.TrimSpace(verweis.Nummer) != ""

	if !hatKennung && !rueckfall {
		return Druckzuordnung{}, nil
	}

	if !anlegen {
		return Druckzuordnung{Neu: true}, nil
	}

	druck := &domain.Druck{
		Name: verweis.Name, Setcode: verweis.Setcode, Nummer: verweis.Nummer, Sprache: verweis.Sprache,
	}

	if hatKennung {
		druck.Kennungen = []domain.Druckkennung{kennung}
	}

	if err := s.drucke.Sichern(ctx, druck); err != nil {
		return Druckzuordnung{}, err
	}

	return Druckzuordnung{DruckID: druck.ID, Alias: druck.Scryfall()}, nil
}
