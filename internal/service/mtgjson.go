package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// MTGJSON liest eine MTG-JSON-Quelle in den Bestand.
//
// # Warum ein eigener Weg und nicht `Set`
//
// `Set` öffnet ein **MSE-Paket**: es fragt nach der Paketart, deutet einen
// Datenbaum und führt unverstandene Blöcke wörtlich mit. Nichts davon gilt
// hier. Beides in eine Methode zu legen hieße, in ihr zuerst zu raten, was für
// eine Datei vorliegt — und wer sich beim Raten irrt, bekommt eine Meldung über
// das falsche Format.
//
// Was **gleich** ist, ist das Ergebnis: dieselbe Kennung, dieselbe Kartenzahl,
// dieselben Befunde. Die Oberfläche unterscheidet die beiden Wege deshalb nur
// beim Aufruf.
//
// # Warum das Spiel von außen kommt
//
// MTG JSON beschreibt Magic-Karten, nicht MSE-Kartenfelder. Zwischen beidem
// liegt eine Zuordnung, und die gehört zum **Spiel** — ein MSE-Spiel bestimmt
// seine Felder selbst. Ein Spiel ohne Zuordnung wird abgewiesen und nicht
// geraten.
func (i *Importeur) MTGJSON(ctx context.Context, pfad, spiel string) (*Ergebnis, error) {
	freigeben, platzfehler := i.platz(ctx, domain.RessourceEinfuhr)
	if platzfehler != nil {
		return nil, platzfehler
	}

	defer freigeben()

	// **Eine Quelle, ein Set** — dieselbe Regel wie in [Importeur.Set], und
	// dieselbe Stelle (#886): in einer Fassade gälte sie nur dort.
	if schon, err := i.schonOffen(ctx, pfad); err != nil || schon != nil {
		return schon, err
	}

	zuordnung, bekannt := mtgjson.ZuordnungFuer(spiel)
	if !bekannt {
		return nil, domain.NewFieldError("spiel", domain.SchluesselValidierung,
			"Für das Spiel %q gibt es keine Zuordnung von MTG-JSON-Feldern.", spiel)
	}

	quelle, err := mtgjson.Lies(pfad, zuordnung)
	if err != nil {
		if eingeordnet := unzugaenglich(pfad, err); eingeordnet != nil {
			return nil, eingeordnet
		}

		if eingeordnet := mtgjsonUnlesbar(pfad, err); eingeordnet != nil {
			return nil, eingeordnet
		}

		return nil, fmt.Errorf("MTG JSON öffnen: %w", err)
	}

	ergebnis := &Ergebnis{Befunde: quelle.Befunde}

	for nummer, gelesen := range quelle.Sets {
		// **Zwischen zwei Sets wird gefragt** (FMT-03): jedes ist eine
		// eigene Transaktion, und ein Abbruch hier lässt nur Ganzes stehen —
		// benannt im Ergebnis, wie bei einem Fehler.
		if fehler := Schritt(ctx, domain.Fortschritt{
			Erledigt: nummer, Gesamt: len(quelle.Sets), Einheit: domain.EinheitSet,
		}); fehler != nil {
			return ergebnis, fehler
		}

		set := setAus(gelesen, spiel, pfad)

		// **Ein Abbruch gibt das Ergebnis mit.** Jedes Set liegt in einer
		// eigenen Transaktion; was vor dem Fehler durchkam, steht im Bestand
		// und gehört benannt, nicht verschwiegen. Eine Transaktion über die
		// ganze Datei wäre die Alternative — eine AllSets-Datei führt Hunderte
		// Sets, und ein Fehler am Ende nähme alles mit.
		if befund := set.ValidateBasics(); befund != nil {
			return ergebnis, befund
		}

		kennung, err := i.speicher.SetAnlegen(ctx, set, nil)
		if err != nil {
			return ergebnis, fmt.Errorf("Set ablegen: %w", err)
		}

		ergebnis.Angelegt = append(ergebnis.Angelegt, kennung)

		// **Die Kennung des zuletzt angelegten Sets gewinnt.** Eine
		// AllSets-Datei führt viele Sets; die Oberfläche kann eines zeigen, und
		// welches, ist ihre Sache. Was hier zählt, ist, dass keines
		// verlorengeht.
		ergebnis.SetID = kennung
		ergebnis.Karten += len(set.Cards)
	}

	Melde(ctx, domain.Fortschritt{
		Erledigt: len(quelle.Sets), Gesamt: len(quelle.Sets), Einheit: domain.EinheitSet,
	})

	return ergebnis, nil
}

// setAus macht aus einem gelesenen MTG-JSON-Set ein Set der Domäne.
func setAus(gelesen *mtgjson.Set, spiel, pfad string) *domain.Set {
	set := &domain.Set{
		Name:   gelesen.Name,
		Game:   spiel,
		Origin: pfad,
		Values: []domain.Value{{Field: "code", Text: gelesen.Code}},
	}

	for _, gelesenKarte := range gelesen.Karten {
		karte := &domain.Card{Values: gelesenKarte.Werte}

		// **Der Rest reist an der Karte mit, nicht als Rohblock.** Rohblöcke
		// hängen am Set und kennen keine Karte; ein mitgeführter JSON-Schlüssel
		// gehört aber zu genau einer. Die Domäne erlaubt das ausdrücklich: ein
		// unbekanntes Feld ist kein Fehler, es wird mitgeführt.
		karte.Values = append(karte.Values, gelesenKarte.Rest...)

		set.Cards = append(set.Cards, karte)
	}

	return set
}

// mtgjsonUnlesbar ordnet ein, was am **Inhalt** einer MTG-JSON-Datei
// scheitert — FMT-03.AK-01 (#1692). Bis dahin ging abgeschnittenes oder
// falsch geformtes JSON als unbekannter Fehler hinaus, und der Dienst machte
// daraus `500`; es ist aber eine Eigenschaft der Eingabe, wie ein kaputtes
// Setpaket ([unlesbar]).
func mtgjsonUnlesbar(pfad string, fehler error) error {
	name := filepath.Base(pfad)

	var (
		syntax *json.SyntaxError
		typ    *json.UnmarshalTypeError
	)

	// **Der Typ des Dekodierers geht ins Protokoll, nicht in den Satz** (G01,
	// `g01fan`): »erwartet war mtgjson.rohSet« sagt niemandem etwas, der eine
	// Datei gewählt hat. Der Satz nennt Datei und Stelle, und beide gehen als
	// Werte mit, damit die Oberfläche ihn übersetzt.
	switch {
	case errors.As(fehler, &syntax):
		return domain.NewFieldError("datei", domain.SchluesselMTGJSONKeinJSON,
			"%s ist kein lesbares MTG JSON — an Byte %d bricht die Datei ab oder ist falsch geformt", name, syntax.Offset).
			MitWerten(map[string]string{"datei": name, "stelle": strconv.FormatInt(syntax.Offset, 10)}).MitGrund(fehler)
	case errors.As(fehler, &typ):
		return domain.NewFieldError("datei", domain.SchluesselMTGJSONKeinSet,
			"%s ist kein lesbares MTG JSON — sie ist JSON, enthält aber kein Set (an Byte %d weicht sie ab)", name, typ.Offset).
			MitWerten(map[string]string{"datei": name, "stelle": strconv.FormatInt(typ.Offset, 10)}).MitGrund(fehler)
	}

	return nil
}
