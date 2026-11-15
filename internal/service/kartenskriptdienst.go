package service

// **Ein Kartenskript ausführen** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was nachgelesen wurde (§15)
//
// `AddCardsScript::perform` (`src/data/add_cards_script.cpp`, Zeilen 20–46,
// Fassung f9d9356 vom 2021-07-06) tut zwei Dinge: das Skript ausführen und die
// Karten **in einer Handlung** ans Set hängen —
//
//	if (!cards.empty()) {
//	  set.actions.addAction(make_unique<AddCardAction>(ADD, set, cards));
//	}
//
// **Ein ⌘Z holt sie zusammen zurück**, und `storage.KartenAnlegen` löst das
// über eine gemeinsame Verlaufsgruppe ein.
//
// # Und eine Abweichung, die benannt gehört (§15, Satz 7)
//
// MSE2 **kopiert** eine Karte, die schon im Set liegt:
//
//	if (contains(set.cards,card) || contains(out,card)) {
//	  card = make_intrusive<Card>(*card);
//	}
//
// Das kann hier nicht eintreten: `new_card` gibt jedes Mal eine **neue**
// Abbildung, und ein Skript hat keinen Griff auf die Karten des Sets. Der
// Zweig hat also keinen Aufrufer — und einen, den kein Test betreten kann,
// lässt §3 nicht zu. **Die Zusage bleibt trotzdem eingelöst**: keine zwei
// angelegten Karten teilen sich einen Zustand, weil es keinen geteilten gibt.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Reihenanleger legt mehrere Karten in **einem** Zug an.
//
// **Nicht [Kartenanleger]**, der eine einzelne leere Karte anlegt: die
// Zusage hier ist eine andere — ganz oder gar nicht, und ein ⌘Z für alle.
// Die Schnittstelle steht hier, erfüllt wird sie in `internal/storage` (§2).
type Reihenanleger interface {
	KartenAnlegen(ctx context.Context, setID int64, karten [][]domain.Feldstand) (
		[]int64, error)
}

// Kartenskriptangabe ist ein Kartenskript, wie der Dienst es braucht.
//
// **Die Gestalt der Domäne und nicht die des Pakets.**
// `msepaket.Kartenskript` trägt das wxWidgets-Kürzel und den rohen Skripttext;
// was hier zählt, ist Name, Beschreibung und Skript. Das ist die
// Anti-Korruptionsschicht an ihrer Stelle (§2, §16) — ändert MSE das Format
// seiner Blöcke, ändert sich `msepaket` und sonst nichts.
type Kartenskriptangabe struct {
	Name         string
	Kuerzel      string
	Beschreibung string
	Skript       string
}

// Spielauskunft gibt, was ein Kartenskript zum Laufen braucht.
//
// **Drei Dinge, und jedes hat seinen Grund**: die **Felder**, damit ein
// unbekannter Feldname auffällt; das **Spielskript**, weil `magic.mse-game`
// darin `skeleton_runner` erklärt; die **Skripte** selbst.
type Spielauskunft interface {
	Kartenskripte(ctx context.Context, setID int64) (
		skripte []Kartenskriptangabe, felder []domain.Field,
		spielskript string, fehler error)
}

// Kartenskriptdienst führt die `add cards script:`-Blöcke eines Spielpakets.
type Kartenskriptdienst struct {
	anleger Reihenanleger
	spiel   Spielauskunft
	frist   time.Duration

	// rechnung rechnet die Karten vor dem Anlegen (#2516) — `nil` legt an,
	// was das Skript gibt.
	rechnung *feldnachzug
}

// kartenskriptfrist ist, wie lange ein Kartenskript rechnen darf (QA-06c,
// #1976). **Der Kontext einer Bindung hat keine Frist**, und eine Schleife
// im Skript eines fremden Spielpakets hielte sie sonst fest. Die zwölf
// Skripte von `magic.mse-game` brauchen zusammen Sekundenbruchteile.
const kartenskriptfrist = 30 * time.Second

// NeuerKartenskriptdienst baut ihn.
func NeuerKartenskriptdienst(
	anleger Reihenanleger, spiel Spielauskunft,
) *Kartenskriptdienst {
	return &Kartenskriptdienst{anleger: anleger, spiel: spiel, frist: kartenskriptfrist}
}

// MitRechnung lässt jede Karte des Skripts ihre Wertskripte rechnen, bevor
// sie angelegt wird — wie MSE bei jeder `AddCardAction` (#2516,
// [anfangswerte]).
func (d *Kartenskriptdienst) MitRechnung(
	stile Kartenstilquelle, schriften render.Schriftquelle, bilder render.Bildquelle,
) *Kartenskriptdienst {
	d.rechnung = &feldnachzug{stile: stile, schriften: schriften, bilder: bilder}

	return d
}

// MitFrist setzt eine andere Frist — für Läufe, die den Abbruch messen.
func (d *Kartenskriptdienst) MitFrist(frist time.Duration) *Kartenskriptdienst {
	d.frist = frist

	return d
}

// Angebot nennt die Kartenskripte, die zu diesem Set gehören.
//
// **Ohne sie auszuführen.** Ein Menü fragt danach; auszuführen wäre eine
// Schreibhandlung als Antwort auf eine Lesefrage.
func (d *Kartenskriptdienst) Angebot(ctx context.Context, setID int64) (
	[]Kartenskriptangabe, error,
) {
	skripte, _, _, fehler := d.spiel.Kartenskripte(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	if skripte == nil {
		skripte = []Kartenskriptangabe{}
	}

	return skripte, nil
}

// Kartenskriptlauf sagt, was ein Lauf angelegt hat.
type Kartenskriptlauf struct {
	// Karten sind die Kennungen der angelegten Karten, in ihrer Reihenfolge.
	Karten []int64

	// Befunde ist, was unterwegs auffiel — ein Spielskript, das nicht lief,
	// etwa.
	Befunde []domain.Befund
}

// Ausfuehren führt das benannte Kartenskript und hängt seine Karten ans Set.
//
// **Erst rechnen, dann schreiben** (§7, Regel 3). Scheitert das Skript, ist
// nichts angelegt; scheitert das Anlegen, sind **alle** nicht angelegt — das
// hält die Transaktion darunter.
func (d *Kartenskriptdienst) Ausfuehren(ctx context.Context, setID int64, name string) (
	*Kartenskriptlauf, error,
) {
	skripte, felder, spielskript, fehler := d.spiel.Kartenskripte(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	gewaehlt, gefunden := nachNamen(skripte, name)
	if !gefunden {
		return nil, fmt.Errorf("%w: dieses Spiel kennt kein Kartenskript %q",
			domain.ErrNotFound, name)
	}

	befristet, ende := context.WithTimeout(ctx, d.frist)
	defer ende()

	karten, befunde, laufFehler := KartenAusSkript(
		befristet, gewaehlt.Skript, felder, spielskript)
	if laufFehler != nil {
		return nil, laufFehler
	}

	// **Ein Skript, das nichts gibt, legt nichts an** — so auch MSE2:
	// `if (!cards.empty())`. Ein leerer Zug im Verlauf wäre ein ⌘Z, das
	// nichts tut.
	if len(karten) == 0 {
		return &Kartenskriptlauf{Karten: []int64{}, Befunde: befunde}, nil
	}

	for _, karte := range karten {
		karte.Values = anfangswerte(ctx, d.rechnung, setID, karte.Values)
	}

	kennungen, anlegefehler := d.anleger.KartenAnlegen(ctx, setID, alsFeldstaende(karten))
	if anlegefehler != nil {
		return nil, fmt.Errorf("die Karten anlegen: %w", anlegefehler)
	}

	return &Kartenskriptlauf{Karten: kennungen, Befunde: befunde}, nil
}

// alsFeldstaende macht aus den Karten die Gestalt, die die Ablage nimmt.
func alsFeldstaende(karten []*domain.Card) [][]domain.Feldstand {
	gebaut := make([][]domain.Feldstand, 0, len(karten))

	for _, karte := range karten {
		felder := make([]domain.Feldstand, 0, len(karte.Values))

		for stelle, wert := range karte.Values {
			felder = append(felder, domain.Feldstand{
				Feld: wert.Field, Wert: wert.Text, Reihenfolge: int64(stelle),
			})
		}

		gebaut = append(gebaut, felder)
	}

	return gebaut
}

// nachNamen sucht ein Skript.
func nachNamen(skripte []Kartenskriptangabe, name string) (Kartenskriptangabe, bool) {
	for _, skript := range skripte {
		if skript.Name == name {
			return skript, true
		}
	}

	return Kartenskriptangabe{}, false
}

// alsKartenskriptangaben übersetzt die Paketgestalt in die der Domäne.
//
// **Feld für Feld und nicht durch Weiterreichen** (§2, §16): `msepaket`
// beschreibt eine **Datei**, dieser Dienst eine Fähigkeit. Wer den fremden
// Verbund durchreichte, hätte MSEs Dateiformat in seiner Fachlogik — und
// änderte MSE es, änderte sich hier alles.
func alsKartenskriptangaben(skripte []msepaket.Kartenskript) []Kartenskriptangabe {
	angaben := make([]Kartenskriptangabe, 0, len(skripte))

	for _, skript := range skripte {
		angaben = append(angaben, Kartenskriptangabe{
			Name:         skript.Name,
			Kuerzel:      skript.Kuerzel,
			Beschreibung: skript.Beschreibung,
			Skript:       skript.Skript,
		})
	}

	return angaben
}

// Vorlagenauskunft gibt die Kartenskripte über den Stilspeicher.
//
// **Er kennt den Weg vom Set zum Spielpaket** — und hat es beim Bauen der
// Vorlage ohnehin geöffnet. Ein zweiter Weg dorthin öffnete es ein zweites
// Mal, samt seinen 19.000 Zeilen Skript.
type Vorlagenauskunft struct{ stile *Stilspeicher }

// NeueVorlagenauskunft baut sie.
func NeueVorlagenauskunft(stile *Stilspeicher) *Vorlagenauskunft {
	return &Vorlagenauskunft{stile: stile}
}

// Kartenskripte gibt, was ein Kartenskript zum Laufen braucht.
//
// **Ohne Vorlage keine Skripte, aber auch kein Fehler.** Ein Set ohne Stil ist
// eines, dessen Spielpaket fehlt; ein Menü mit null Einträgen ist die richtige
// Antwort darauf, ein Abbruch nicht (§8).
func (v *Vorlagenauskunft) Kartenskripte(ctx context.Context, setID int64) (
	[]Kartenskriptangabe, []domain.Field, string, error,
) {
	vorlage, _, da := v.stile.StilFuerSet(ctx, setID)
	if !da {
		return nil, nil, "", nil
	}

	return vorlage.Kartenskripte, vorlage.Kartenfelder, vorlage.Spielskript, nil
}
