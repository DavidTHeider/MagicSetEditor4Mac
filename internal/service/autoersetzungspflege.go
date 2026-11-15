package service

// **Die Autoersetzungen eines Spiels sehen und pflegen** — Werkstück E von
// Phase 4.11
// (#1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
//
// # Was nachgelesen wurde (§15)
//
// `GameSettings::initDefaults` (`src/data/settings.cpp`, Zeilen 88–104,
// Fassung f9d9356 vom 2021-07-06):
//
//	FOR_EACH_CONST(ar, game.auto_replaces) {
//	  bool already_have = false;
//	  FOR_EACH(ar2, auto_replaces) {
//	    if (ar->match == ar2->match) { ar2->custom = false; already_have = true; break; }
//	  }
//	  if (!already_have) { ar->custom = false; auto_replaces.push_back(ar); }
//	}
//
// **Drei Sätze stehen darin**, und alle drei gelten hier:
//
//  1. verglichen wird über das **Muster**;
//  2. eine Regel des Pakets ist **nicht eigen** — ihr Muster ist fest;
//  3. was jemand angelegt hat, **bleibt eigen** und überlebt jede
//     Paketaktualisierung.
//
// # Die eine Abweichung, und sie ist begründet
//
// **MSE kopiert die Regeln des Pakets in die Einstellungen; hier werden sie
// gemischt.** MSE notiert selbst, warum es kopiert:
// `// TODO: when we start saving games, clone here` — seine Liste hält Zeiger
// auf die Regeln des Spiels. Eine Kopie im Bestand hieße, dass eine
// Paketaktualisierung nichts mehr ändert; ein Muster, das der Urheber
// berichtigt, bliebe hier falsch. Der Vermerk führt es aus.

import (
	"context"
	"fmt"
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Autoersetzungsablage ist, was die Pflege vom Bestand braucht.
//
// **Die Schnittstelle steht hier, erfüllt wird sie in `internal/storage`**
// (§2).
type Autoersetzungsablage interface {
	Regeln(ctx context.Context, spiel string) ([]domain.Autoersetzungsregel, error)
	Setzen(ctx context.Context, spiel string, regel domain.Autoersetzungsregel) error
	Loeschen(ctx context.Context, spiel, muster string) error
	Zuruecksetzen(ctx context.Context, spiel string) error
	Schalter(ctx context.Context, spiel string) (bool, error)
	SchalterSetzen(ctx context.Context, spiel string, an bool) error
}

// Autoersetzungsliste ist die Antwort auf »welche Regeln gelten hier?«.
type Autoersetzungsliste struct {
	Regeln []domain.Autoersetzungsregel

	// An ist der globale Schalter dieses Spiels — MSEs `use_auto_replace`.
	An bool
}

// Autoersetzungspflege zeigt und pflegt die Regeln eines Spiels.
type Autoersetzungspflege struct {
	ablage Autoersetzungsablage
	spiele Spielquelle
}

// Spielquelle sagt, zu welchem Spiel ein Set gehört und was dessen Paket an
// Regeln mitbringt.
type Spielquelle interface {
	SpielVon(ctx context.Context, setID int64) (string, error)
	Paketregeln(ctx context.Context, setID int64) ([]msepaket.AutoErsetzung, error)
}

// Paketregelquelle gibt die `auto replace:`-Blöcke des Spielpakets eines Sets.
//
// **Der Stilwähler erfüllt sie schon** (#897): er öffnet das Spielpaket
// ohnehin. Ein zweiter Weg dorthin öffnete es ein zweites Mal.
type Paketregelquelle interface {
	Autoersetzungen(ctx context.Context, setID int64) ([]msepaket.AutoErsetzung, error)
}

// Spielregeln erfüllt [Spielquelle] aus den zwei Quellen, die es dafür gibt.
//
// **Zwei und nicht eine**: welchem Spiel ein Set gehört, weiß der Bestand;
// was dessen Paket mitbringt, weiß der Stilwähler. Ein Typ, der beides
// könnte, wäre ein Typ zu viel.
type Spielregeln struct {
	koepfe Kopfquelle
	paket  Paketregelquelle
}

// NeueSpielregeln baut sie.
func NeueSpielregeln(koepfe Kopfquelle, paket Paketregelquelle) *Spielregeln {
	return &Spielregeln{koepfe: koepfe, paket: paket}
}

// SpielVon gibt das Spiel eines Sets.
func (s *Spielregeln) SpielVon(ctx context.Context, setID int64) (string, error) {
	koepfe, err := s.koepfe.Setkoepfe(ctx)
	if err != nil {
		return "", fmt.Errorf("die Sets lesen: %w", err)
	}

	for i := range koepfe {
		if koepfe[i].ID == setID {
			return koepfe[i].Game, nil
		}
	}

	return "", fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID)
}

// Paketregeln reicht die Regeln des Spielpakets durch.
func (s *Spielregeln) Paketregeln(ctx context.Context, setID int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return s.paket.Autoersetzungen(ctx, setID)
}

// NeueAutoersetzungspflege baut sie.
func NeueAutoersetzungspflege(
	ablage Autoersetzungsablage, spiele Spielquelle,
) *Autoersetzungspflege {
	return &Autoersetzungspflege{ablage: ablage, spiele: spiele}
}

// Liste gibt die Regeln des Pakets, überlagert von denen des Bestands.
//
// **Nach dem Muster geordnet**, wie MSE sie zeigt (`sortBy(1, true)`,
// `src/gui/auto_replace_window.cpp`, Zeile 90, Fassung f9d9356 vom 2021-07-06) — und wie der Bestand sie
// ohnehin liefert.
func (p *Autoersetzungspflege) Liste(ctx context.Context, setID int64) (
	Autoersetzungsliste, error,
) {
	spiel, err := p.spiele.SpielVon(ctx, setID)
	if err != nil {
		return Autoersetzungsliste{}, err
	}

	an, schalterfehler := p.ablage.Schalter(ctx, spiel)
	if schalterfehler != nil {
		return Autoersetzungsliste{}, schalterfehler
	}

	eigene, lesefehler := p.ablage.Regeln(ctx, spiel)
	if lesefehler != nil {
		return Autoersetzungsliste{}, lesefehler
	}

	// **Ein Fehler am Paket hält die Liste nicht auf** (§8): die eigenen
	// Regeln sind der Grund, warum jemand hier ist, und ein Set ohne
	// Spielpaket ist auf einem frischen Rechner der Normalfall.
	//nolint:errcheck // siehe darüber
	ausDemPaket, _ := p.spiele.Paketregeln(ctx, setID)

	return Autoersetzungsliste{Regeln: gemischt(ausDemPaket, eigene), An: an}, nil
}

// gemischt legt die Zeilen des Bestands über die des Pakets.
//
// **Über das Muster**, siehe den Kopf. Eine Zeile des Bestands zu einem
// Muster, das im Paket steht, ist **nicht eigen**: sie überlagert nur, und
// ihr Muster gehört dem Paket.
func gemischt(
	ausDemPaket []msepaket.AutoErsetzung, eigene []domain.Autoersetzungsregel,
) []domain.Autoersetzungsregel {
	nachMuster := make(map[string]domain.Autoersetzungsregel, len(eigene))
	for _, regel := range eigene {
		nachMuster[regel.Muster] = regel
	}

	gebaut := make([]domain.Autoersetzungsregel, 0, len(ausDemPaket)+len(eigene))
	ausDemPaketGesehen := map[string]bool{}

	for _, regel := range ausDemPaket {
		ausDemPaketGesehen[regel.Muster] = true

		if ueberlagert, da := nachMuster[regel.Muster]; da {
			ueberlagert.Eigen = false
			gebaut = append(gebaut, ueberlagert)

			continue
		}

		gebaut = append(gebaut, domain.Autoersetzungsregel{
			Muster: regel.Muster, Ersatz: regel.Ersatz,
			GanzesWort: regel.GanzesWort, Aktiv: true, Eigen: false,
		})
	}

	for _, regel := range eigene {
		if !ausDemPaketGesehen[regel.Muster] {
			regel.Eigen = true
			gebaut = append(gebaut, regel)
		}
	}

	sort.Slice(gebaut, func(a, b int) bool {
		return gebaut[a].Muster < gebaut[b].Muster
	})

	return gebaut
}

// Setzen legt eine Regel an oder schreibt sie fort.
//
// **Das Muster einer Paketregel lässt sich nicht ändern** — das ist MSEs
// `match->Enable(ar && ar->custom)`. Wer es doch versucht, legt in Wahrheit
// eine **neue** Regel an; das geht, und dafür gibt es diesen Weg. Was hier
// abgewiesen wird, ist nichts — die Wache sitzt beim Löschen.
func (p *Autoersetzungspflege) Setzen(
	ctx context.Context, setID int64, regel domain.Autoersetzungsregel,
) error {
	spiel, err := p.spiele.SpielVon(ctx, setID)
	if err != nil {
		return err
	}

	// **`Eigen` entscheidet der Dienst und nicht der Aufrufer.** Ob eine
	// Regel dem Paket gehört, steht im Paket; eine Fläche, die es
	// mitschickte, könnte es falsch behaupten.
	//nolint:errcheck // ohne Paket ist keine Regel eine Paketregel
	ausDemPaket, _ := p.spiele.Paketregeln(ctx, setID)

	regel.Eigen = !imPaket(ausDemPaket, regel.Muster)

	return p.ablage.Setzen(ctx, spiel, regel)
}

// Loeschen nimmt eine **eigene** Regel heraus.
//
// **Eine Regel des Pakets lässt sich nicht löschen** — MSEs
// `remove->Enable(ar && ar->custom)`. Sie gehört dem Paket; was hier ginge,
// wäre sie abzuschalten, und dafür gibt es `Aktiv`.
func (p *Autoersetzungspflege) Loeschen(
	ctx context.Context, setID int64, muster string,
) error {
	spiel, err := p.spiele.SpielVon(ctx, setID)
	if err != nil {
		return err
	}

	//nolint:errcheck // ohne Paket ist keine Regel eine Paketregel
	ausDemPaket, _ := p.spiele.Paketregeln(ctx, setID)

	if imPaket(ausDemPaket, muster) {
		return fmt.Errorf(
			"%w: die Regel %q gehört dem Spielpaket — sie lässt sich abschalten, "+
				"nicht löschen", domain.ErrValidation, muster)
	}

	return p.ablage.Loeschen(ctx, spiel, muster)
}

// Zuruecksetzen vergisst alle Abweichungen dieses Spiels — MSEs „defaults".
func (p *Autoersetzungspflege) Zuruecksetzen(ctx context.Context, setID int64) error {
	spiel, err := p.spiele.SpielVon(ctx, setID)
	if err != nil {
		return err
	}

	return p.ablage.Zuruecksetzen(ctx, spiel)
}

// SchalterSetzen legt den globalen Schalter dieses Spiels um.
func (p *Autoersetzungspflege) SchalterSetzen(
	ctx context.Context, setID int64, an bool,
) error {
	spiel, err := p.spiele.SpielVon(ctx, setID)
	if err != nil {
		return err
	}

	return p.ablage.SchalterSetzen(ctx, spiel, an)
}

// imPaket sagt, ob ein Muster aus dem Spielpaket kommt.
func imPaket(ausDemPaket []msepaket.AutoErsetzung, muster string) bool {
	for _, regel := range ausDemPaket {
		if regel.Muster == muster {
			return true
		}
	}

	return false
}

// Wirksame gibt die Regeln, die beim Tippen wirklich greifen.
//
// # Drei Filter, und jeder steht in MSE
//
//  1. **Der globale Schalter.** `tryAutoReplace` beginnt mit
//     `if (!gs.use_auto_replace) return false;`
//     (`src/gui/value/text.cpp`, Zeile 1034, Fassung f9d9356 vom
//     2021-07-06). Aus heißt: keine Regel greift, auch keine des Pakets.
//  2. **`enabled` je Regel** — ebenda Zeile 1038: `if (!ar->enabled) continue;`.
//  3. **Die Reihenfolge** bleibt die der Liste, und die ist nach dem Muster
//     geordnet. **Das ist eine Abweichung**: MSE läuft seine Liste in der
//     Reihenfolge der Datei ab, und diese hier ist die des Bestands. Sie
//     wird sichtbar, wenn zwei Regeln auf dieselbe Stelle passen — im Pack
//     tut das keine (gemessen am 2026-08-31: 615 Blöcke, kein Muster
//     zweimal).
func (p *Autoersetzungspflege) Wirksame(ctx context.Context, setID int64) (
	[]msepaket.AutoErsetzung, error,
) {
	liste, err := p.Liste(ctx, setID)
	if err != nil {
		return nil, err
	}

	if !liste.An {
		return nil, nil
	}

	wirksam := make([]msepaket.AutoErsetzung, 0, len(liste.Regeln))

	for _, regel := range liste.Regeln {
		if !regel.Aktiv {
			continue
		}

		wirksam = append(wirksam, msepaket.AutoErsetzung{
			Muster: regel.Muster, Ersatz: regel.Ersatz, GanzesWort: regel.GanzesWort,
		})
	}

	return wirksam, nil
}
