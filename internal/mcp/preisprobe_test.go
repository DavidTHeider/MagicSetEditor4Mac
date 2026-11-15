// **Die Vorrichtung für Preise, Markt und KI** — herausgezogen für Werkstück
// A13.3 von Phase 5.3
// (#1452),
// als `sammlungsprobe_test.go` an der Zeilengrenze aus §2 stand.
//
// **Der Schnitt folgt der Frage**: dort steht, was eine **Sammlung** hergibt
// — Exemplare, die man anlegt, ändert und löscht. Hier steht, was ein
// **Dienst** hergibt: Preise, Angebote, Modelle und was ein Modell antwortet.
// Dass beide hinter derselben Naht hängen, ist eine Eigenschaft der Naht und
// keine der Vorrichtung.
package mcp_test

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// preisprobe gibt feste Preisstände.
type preisprobe struct{ geholt int }

// neuePreisprobe gibt eine Probe mit einem Stand.
func neuePreisprobe() *preisprobe { return &preisprobe{} }

// probesicht ist der Stand, den die Probe herausgibt.
func probesicht() mcp.Preissicht {
	return mcp.Preissicht{
		Printing: "probe-druck", Quelle: "scryfall", Waehrung: "EUR",
		Finish: "nonfoil", Wert: "4.17", Erhoben: "2026-08-27T12:00:00Z",
		AlterStunden: 3,
	}
}

func (p *preisprobe) PreiseHolen(context.Context, string) (
	[]mcp.Preissicht, []string, error,
) {
	// **Gezählt, damit ein Lauf sieht, dass wirklich geholt wurde.**
	p.geholt++

	return []mcp.Preissicht{probesicht()}, []string{}, nil
}

func (p *preisprobe) Preise(context.Context, string) ([]mcp.Preissicht, error) {
	return []mcp.Preissicht{probesicht()}, nil
}

func (p *preisprobe) Preisverlauf(context.Context, mcp.Preisreihe) (
	[]mcp.Preissicht, error,
) {
	return []mcp.Preissicht{probesicht()}, nil
}

func (p *preisprobe) Preisbulk(_ context.Context, quelle, umfang string) (
	mcp.Bulkbilanzsicht, error,
) {
	return mcp.Bulkbilanzsicht{
		Quelle: quelle, Umfang: umfang, Drucke: 1, Staende: 1,
	}, nil
}

func (p *preisprobe) Marktangebote(_ context.Context, wunsch mcp.Marktwunsch) (
	mcp.Marktsicht, error,
) {
	return mcp.Marktsicht{
		Quelle: wunsch.Quelle, Fassung: "89", Zugeordnet: true,
		Angebote: []mcp.Angebotssicht{{
			Quelle: wunsch.Quelle, Kennung: "101862104", Wert: "2.17",
			Waehrung: "EUR", Menge: 3, Buendel: 1, Finish: "nonfoil",
			Zustand: "Moderately Played", Sprache: "en", Verkaeufer: "einer",
			Land: "FI",
		}},
	}, nil
}

func (p *preisprobe) Anbieterstand(_ context.Context, dienst string) (
	mcp.Anbietersicht, error,
) {
	return mcp.Anbietersicht{
		Dienst: dienst, Gesamt: 100000, Rest: 84213, Verbraucht: 15787,
		Ueberzugssatz: "0.0004", HoechstensAltMinuten: 30,
	}, nil
}

func (p *preisprobe) Fremdpreisverlauf(_ context.Context, wunsch mcp.Verlaufsbezug) (
	mcp.Verlaufssicht, error,
) {
	return mcp.Verlaufssicht{
		Quelle: wunsch.Quelle, Fassung: "mtg-m11-149", Zugeordnet: true, Staende: 4,
	}, nil
}

func (p *preisprobe) KIModelle(context.Context) (mcp.KIKatalogsicht, error) {
	return mcp.KIKatalogsicht{
		NurLokal: true, Modelle: []mcp.KIModellsicht{einKIModell()},
	}, nil
}

// KIVorschlag gibt einen Vorschlag samt Kennung (A7, #1446).
//
// **`uebernehmen` kommt durch**: ohne das Feld sähe ein Lauf über das
// Werkzeug nicht, ob die Flagge überhaupt ankommt. Der Fehlerweg steht bei
// `stolpernderPreis`.
func (p *preisprobe) KIVorschlag(_ context.Context, bezug mcp.KIVorschlagsbezug) (
	mcp.KIVorschlagssicht, error,
) {
	sicht := mcp.KIVorschlagssicht{
		Kennung: "ki-4-1", Modell: bezug.Alias,
		Felder: []mcp.KIFeldwertsicht{{Feld: "name", Werte: []string{"Probe"}}},
		Ablauf: "2026-09-12T12:00:00Z",
	}

	if bezug.Uebernehmen {
		sicht.Uebernommen = 1
	}

	return sicht, nil
}

func einKIModell() mcp.KIModellsicht {
	return mcp.KIModellsicht{
		Anbieter: "ollama", Alias: "beispiel:3b", Laufzeit: "ollama 0.12.3",
		Ort: "lokal",
		Faehigkeiten: []mcp.KIBelegsicht{{
			Faehigkeit: "einbettung", Datum: "2026-09-11",
			Beleg: "eigener Lauf gegen die Laufzeit",
		}},
		Herkunft: []mcp.KIGliedsicht{{
			Art: "artefakt", Kennung: "ausgefuehrt",
			Quelle: "ollama:beispiel", Revision: "sha256:dddd",
			Lizenz: "Apache-2.0",
		}},
		Bindung: "abcdef",
	}
}

func (p *preisprobe) Portfolio(context.Context) (mcp.Portfoliosicht, error) {
	return mcp.Portfoliosicht{
		Zeilen: []mcp.Portfoliozeilensicht{{
			Waehrung: "EUR", Wert: "12.51", Exemplare: 3,
			AeltesterPreis: "2026-08-27T12:00:00Z", AlterStunden: 3,
			Quelle: "scryfall",
		}},
		OhnePreis: 0,
	}, nil
}

// stolperndeSammlung scheitert auf jedem Weg.
//
// **Für die Fehlerwege der Werkzeuge**: dass eine Meldung durchkommt, ist eine
// eigene Zusage — und ohne einen Lauf dafür bliebe sie eine Behauptung (§3).
type stolperndeSammlung struct{}

// errGestolpertMCP ist der Fehler, den die stolpernden Doubles geben.
var errGestolpertMCP = errors.New("die Naht stolpert")

func (stolperndeSammlung) Sammlung(context.Context) ([]mcp.Exemplarsicht, error) {
	return nil, errGestolpertMCP
}

func (stolperndeSammlung) Sammlungsfenster(
	context.Context, string, string, string, int, int,
) (fenster []mcp.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	return nil, domain.Fensterzahlen{}, errGestolpertMCP
}

func (stolperndeSammlung) SammlungFuer(context.Context, string) (
	[]mcp.Exemplarsicht, error,
) {
	return nil, errGestolpertMCP
}

func (stolperndeSammlung) ExemplarAnlegen(context.Context, mcp.Exemplarsicht) (
	int64, error,
) {
	return 0, errGestolpertMCP
}

func (stolperndeSammlung) ExemplarAendern(context.Context, mcp.Exemplarsicht) error {
	return errGestolpertMCP
}

func (stolperndeSammlung) ExemplarLoeschen(context.Context, int64, int64) error {
	return errGestolpertMCP
}

// stolpernderPreis scheitert auf jedem Weg.
type stolpernderPreis struct{}

func (stolpernderPreis) PreiseHolen(context.Context, string) (
	[]mcp.Preissicht, []string, error,
) {
	return nil, nil, errGestolpertMCP
}

func (stolpernderPreis) Preise(context.Context, string) ([]mcp.Preissicht, error) {
	return nil, errGestolpertMCP
}

func (stolpernderPreis) Preisverlauf(context.Context, mcp.Preisreihe) (
	[]mcp.Preissicht, error,
) {
	return nil, errGestolpertMCP
}

func (stolpernderPreis) Preisbulk(context.Context, string, string) (
	mcp.Bulkbilanzsicht, error,
) {
	return mcp.Bulkbilanzsicht{}, errGestolpertMCP
}

func (stolpernderPreis) Marktangebote(context.Context, mcp.Marktwunsch) (
	mcp.Marktsicht, error,
) {
	return mcp.Marktsicht{}, errGestolpertMCP
}

func (stolpernderPreis) Anbieterstand(context.Context, string) (mcp.Anbietersicht, error) {
	return mcp.Anbietersicht{}, errGestolpertMCP
}

func (stolpernderPreis) Fremdpreisverlauf(context.Context, mcp.Verlaufsbezug) (
	mcp.Verlaufssicht, error,
) {
	return mcp.Verlaufssicht{}, errGestolpertMCP
}

func (stolpernderPreis) KIModelle(context.Context) (mcp.KIKatalogsicht, error) {
	return mcp.KIKatalogsicht{}, errGestolpertMCP
}

func (stolpernderPreis) KIVorschlag(context.Context, mcp.KIVorschlagsbezug) (
	mcp.KIVorschlagssicht, error,
) {
	return mcp.KIVorschlagssicht{}, errGestolpertMCP
}

func (stolpernderPreis) KIBatch(context.Context, mcp.KIBatchbezug) (
	mcp.KIBatchsicht, error,
) {
	return mcp.KIBatchsicht{}, errGestolpertMCP
}

func (stolpernderPreis) Portfolio(context.Context) (mcp.Portfoliosicht, error) {
	return mcp.Portfoliosicht{}, errGestolpertMCP
}

// **Und die zwei Wege der Messung** — A9, #1448.
func (p *preisprobe) KIModellPruefen(_ context.Context, bezug mcp.KIProbenbezug) (
	mcp.KIProbensicht, error,
) {
	return mcp.KIProbensicht{Proben: []mcp.KIProbesicht{{
		Anbieter: bezug.Anbieter, Alias: bezug.Alias,
		Faehigkeit: "strukturiert", Datum: "2026-09-12",
		Beleg: "gemessen", Modell: bezug.Alias,
	}}}, nil
}

func (p *preisprobe) KIModellVergessen(context.Context, mcp.KIProbenbezug) error {
	return nil
}

// KIGeraetemodellFuerSitzung merkt den Wunsch — AI-05 (#2042).
func (p *preisprobe) KIGeraetemodellFuerSitzung(_ context.Context, an bool) (bool, error) {
	return an, nil
}

func (stolpernderPreis) KIModellPruefen(context.Context, mcp.KIProbenbezug) (
	mcp.KIProbensicht, error,
) {
	return mcp.KIProbensicht{}, errGestolpertMCP
}

func (stolpernderPreis) KIModellVergessen(context.Context, mcp.KIProbenbezug) error {
	return errGestolpertMCP
}

func (stolpernderPreis) KIGeraetemodellFuerSitzung(context.Context, bool) (bool, error) {
	return true, errGestolpertMCP
}

// KIErkennung liest ein Kartenfoto — A13.3 (#1452).
//
// **Es antwortet und scheitert nicht**: dieses Doppel ist der Weg, auf dem
// das Werkzeug seinen Erfolgsfall zeigt.
func (*preisprobe) KIErkennung(_ context.Context, bezug mcp.KIErkennungsbezug) (
	mcp.KIErkennungssicht, error,
) {
	return mcp.KIErkennungssicht{
		Beobachtungen: []mcp.KIBeobachtungsicht{
			{Art: "name", Wert: "Shivan Dragon", Lesbar: true},
			{Art: "nummer", Wert: "", Lesbar: false},
		},
		NichtGelesen: []string{"typzeile"},
		Kandidaten: []mcp.KIBildkandidatsicht{{
			Kennung: "lea-174", Name: "Shivan Dragon", Setcode: "lea",
			Setname: "Limited Edition Alpha", Nummer: "174",
			Gestuetzt: []string{"name"},
		}},
		Drucke: 1, Eindeutig: true, Alter: "1h0m0s",
		Anbieter: bezug.Anbieter, Modell: bezug.Alias,
	}, nil
}

// KIErkennung stolpert — A13.3 (#1452), mit **dem** Fehler dieses Doppels:
// der Lauf darüber prüft, dass er durchkommt.
func (stolpernderPreis) KIErkennung(context.Context, mcp.KIErkennungsbezug) (
	mcp.KIErkennungssicht, error,
) {
	return mcp.KIErkennungssicht{}, errGestolpertMCP
}
