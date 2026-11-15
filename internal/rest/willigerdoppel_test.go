// **Das willige Double des Dienstes** — abgetrennt am 2026-08-26.
//
// # Warum es eine eigene Datei ist
//
// `vorrichtung_test.go` war auf 410 Zeilen gewachsen und trug zwei Dinge: die
// **Vorrichtung** — was ein Endpunkt ist, wie er gerufen wird, wie ein Dienst
// entsteht — und das **Double**, das zu allem ja sagt. Der Schnitt folgt der
// Frage, die eine Datei beantwortet (§2), nicht der Zeilenzahl: das Double
// wächst mit jeder Schnittstelle, die der Dienst bekommt, die Vorrichtung
// nicht.
package rest_test

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// williger erfüllt alle Schnittstellen und sagt zu allem ja.
type williger struct{}

// Packarten und PackZiehen antworten willig (#889).
//
// **Mit einer Karte und nicht mit nichts**: ein leeres Pack sähe aus wie ein
// gültiger Erfolg, und der Wächter über die vier Klassen prüfte dann eine
// Antwort, die nichts sagt.
// SetAnlegen legt ein leeres Set an (#891).
func (williger) SetAnlegen(context.Context, string, string, string) (int64, error) {
	return 1, nil
}

func (williger) Packarten(context.Context, int64) (rest.Packangebot, error) {
	return rest.Packangebot{
		Arten:   []rest.Packart{{Name: "alles", Waehlbar: true, Karten: 1}},
		Befunde: []string{},
	}, nil
}

func (williger) PackZiehen(context.Context, int64, string, string) (
	rest.Packziehung, error,
) {
	return rest.Packziehung{
		Karten:  []rest.Packkarte{{ID: "1", Name: "Probe"}},
		Befunde: []string{},
	}, nil
}

// Auswerten rechnet einen Ausdruck aus (#1070).
func (williger) Auswerten(_ context.Context, _ int64, ausdruck string) (
	rest.Skriptergebnis, error,
) {
	return rest.Skriptergebnis{
		Quelltext: `"` + ausdruck + `"`, Text: ausdruck,
		Art: "Zeichenkette", Befunde: []string{},
	}, nil
}

// Die drei Züge der Paketwerkstatt (#1164); der Dateibaum steht in
// `dateibaumdoppel_test.go`.
func (williger) DateienUndBeiwerk(context.Context, string) (dateien, beiwerk []string, fehler error) {
	return []string{"style"}, []string{"rahmen.png"}, nil
}

func (williger) Lesen(context.Context, string, string) (inhalt, stand string, fehler error) {
	return "mse version: 2.0.0\n", "s", nil
}

func (williger) Schreiben(context.Context, string, string, string, string) (string, error) {
	return "neu", nil
}

// Namen zählt auf, was die Umgebung kennt (#1164).
func (williger) Namen(context.Context, int64) ([]string, error) {
	return []string{"card", "length", "set"}, nil
}

// Einlesen liest einen ganzen Satz als Set ein (#1033).
func (williger) Einlesen(context.Context, string, string) (rest.Satz, error) {
	return rest.Satz{SetID: 9, Karten: 103, Befunde: []string{}}, nil
}

// Anlegen legt eine Karte aus einem Scryfall-Namen an (#1032).
func (williger) Anlegen(context.Context, int64, string) (rest.Anlage, error) {
	return rest.Anlage{
		KarteID:  7,
		Felder:   []string{"name", "rule text"},
		Befunde:  []string{},
		Herkunft: "https://scryfall.com/card/lea/161",
	}, nil
}

// Regelhinweise gibt zwei Hinweise — einen offiziellen und eine Anmerkung.
func (williger) Regelhinweise(context.Context, string) (rest.Regelauskunft, error) {
	return rest.Regelauskunft{
		Kartenname: "Prüfstein",
		Hinweise: []rest.Regelhinweis{
			{Offiziell: true, Quelle: "wotc", Datum: "2004-10-04", Text: "Offiziell."},
			{Offiziell: false, Quelle: "scryfall", Datum: "2021-07-23", Text: "Anmerkung."},
		},
		Adresse: "https://beispiel.invalid/karte/1",
	}, nil
}

// Stand meldet eine Sammeldatei, die bereitliegt.
func (williger) Stand(context.Context) (rest.Sammelstand, error) {
	return rest.Sammelstand{
		Art: "oracle_cards", Pfad: "/probe/oracle_cards.jsonl.gz",
		Vorhanden: true, Bytes: 24531600, Angeboten: 24531600,
		Stand: "2026-08-23T21:01:56.596+00:00",
	}, nil
}

// Holen tut dasselbe — das Doppel lädt nichts.
func (w williger) Holen(ctx context.Context) (rest.Sammelstand, error) {
	return w.Stand(ctx)
}

// Manakosten meldet einen Kosten ohne Beanstandung.
func (williger) Manakosten(context.Context, string) ([]string, error) { return nil, nil }

// Einsetzen meldet ein eingesetztes Bild samt Nennung.
func (williger) Einsetzen(context.Context, int64, int64, string, string, string) (
	rest.Bildeinsatz, error,
) {
	return rest.Bildeinsatz{
		Feld: "image", Kuenstler: "Christopher Rush",
		Nennung:            "Illustration: Christopher Rush. ™ & © Wizards of the Coast",
		Adresse:            "https://scryfall.com/card/lea/161",
		IllustratorGesetzt: true,
	}, nil
}

// **Die Sammlung** (#1165).
func (williger) Sammlungsfenster(
	context.Context, string, string, string, int, int,
) (fenster []rest.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	return []rest.Exemplarsicht{}, domain.Fensterzahlen{}, nil
}

func (williger) Sammlung(context.Context) ([]rest.Exemplarsicht, error) {
	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}}, nil
}

func (williger) SammlungFuer(context.Context, string) ([]rest.Exemplarsicht, error) {
	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}}, nil
}

func (williger) ExemplarAnlegen(context.Context, rest.Exemplarsicht) (int64, error) {
	return 1, nil
}

func (williger) ExemplarAendern(context.Context, rest.Exemplarsicht) error {
	return nil
}

func (williger) ExemplarLoeschen(context.Context, int64, int64) error {
	return nil
}

// **Und die Preise** (#1165).
func (williger) PreiseHolen(context.Context, string) ([]rest.Preissicht, []string, error) {
	return []rest.Preissicht{probepreis()}, []string{}, nil
}

// **Und das Löschen eines Sets** (#1218).
func (williger) SetLoeschen(context.Context, int64, string) ([]string, error) {
	return nil, nil
}

func (williger) Preise(context.Context, string) ([]rest.Preissicht, error) {
	return []rest.Preissicht{probepreis()}, nil
}

func (williger) Preisverlauf(context.Context, rest.Preisreihe) (
	[]rest.Preissicht, error,
) {
	return []rest.Preissicht{probepreis()}, nil
}

func (williger) Preisbulk(context.Context, string, string) (
	rest.Bulkbilanzsicht, error,
) {
	return rest.Bulkbilanzsicht{Quelle: "mtgjson", Umfang: "heute"}, nil
}

func (williger) Marktangebote(_ context.Context, wunsch rest.Marktwunsch) (
	rest.Marktsicht, error,
) {
	return rest.Marktsicht{
		Quelle: wunsch.Quelle, Fassung: "89", Zugeordnet: true,
		Angebote: []rest.Angebotssicht{{
			Quelle: wunsch.Quelle, Kennung: "101862104", Wert: "2.17",
			Waehrung: "EUR", Menge: 3, Buendel: 1, Finish: "nonfoil",
			Zustand: "Moderately Played", Sprache: "en", Verkaeufer: "einer",
			Land: "FI",
		}},
	}, nil
}

func (williger) Anbieterstand(_ context.Context, dienst string) (rest.Anbietersicht, error) {
	return rest.Anbietersicht{
		Dienst: dienst, Gesamt: 100000, Rest: 84213, Verbraucht: 15787,
		Ueberzugssatz: "0.0004", HoechstensAltMinuten: 30,
	}, nil
}

func (williger) Fremdpreisverlauf(_ context.Context, wunsch rest.Verlaufsbezug) (
	rest.Verlaufssicht, error,
) {
	return rest.Verlaufssicht{
		Quelle: wunsch.Quelle, Fassung: "mtg-m11-149", Zugeordnet: true, Staende: 4,
	}, nil
}

func (williger) KIModelle(context.Context) (rest.KIKatalogsicht, error) {
	return rest.KIKatalogsicht{NurLokal: true, Modelle: []rest.KIModellsicht{{
		Anbieter: "ollama", Alias: "beispiel:3b", Ort: "lokal",
	}}}, nil
}

func (williger) Portfolio(context.Context) ([]rest.Portfoliozeilensicht, int, error) {
	return []rest.Portfoliozeilensicht{{
		Waehrung: "EUR", Wert: "4.17", Exemplare: 1,
		AeltesterPreis: "2026-08-27T12:00:00Z", Quelle: "scryfall",
	}}, 0, nil
}

// probepreis ist der Preisstand, den die Doubles herausgeben.
func probepreis() rest.Preissicht {
	return rest.Preissicht{
		Printing: "probe-druck", Quelle: "scryfall", Waehrung: "EUR",
		Finish: "nonfoil", Wert: "4.17", Erhoben: "2026-08-27T12:00:00Z",
	}
}

// **Die Listeneinfuhr** (#1292).
func (williger) ListeDeuten(context.Context, string) (any, error) {
	return map[string]any{"zeilen": []any{}, "befunde": []string{}}, nil
}

func (williger) ListeEinlesen(context.Context, []rest.Listeneinfuhrzeile, string) (any, error) {
	return map[string]any{"angelegt": []int64{1}, "erhoeht": 0, "stueck": 1}, nil
}

// **Die Sammlungsausfuhr** (#1292).
func (williger) SammlungAusgeben(context.Context, string) (string, error) {
	return "printing,menge\nprobe-druck,1\n", nil
}

func (williger) SammlungCSVEinlesen(context.Context, string, string, string) (any, error) {
	return map[string]any{"angelegt": []int64{1}, "erhoeht": 0, "stueck": 1}, nil
}

// **Die Sammlungssichten** (#1292).
func (williger) SammlungBesitz(context.Context, string) (any, error) {
	return map[string]any{"stueck": 1, "gewuenscht": 0, "besitzt": true}, nil
}

func (williger) SammlungLagerorte(context.Context) (any, error) {
	return []any{}, nil
}

func (williger) SammlungDubletten(context.Context, int) (any, error) {
	return []any{}, nil
}

func (williger) SammlungWuensche(context.Context) (any, error) {
	return []any{}, nil
}

// **Die Massenvorgänge** (#1292).
func (williger) MasseAendern(context.Context, rest.Massenaenderung) (any, error) {
	return map[string]any{"geaendert": 1, "befunde": []string{}}, nil
}

func (williger) MasseLoeschen(context.Context, []rest.Massenzeile) (any, error) {
	return map[string]any{"geaendert": 1, "befunde": []string{}}, nil
}

func (williger) MassePreise(context.Context, int) (any, error) {
	return map[string]any{"geholt": 0, "offen": 0, "fertig": true}, nil
}

// **Die Kartenskripte** (#1337).
func (williger) Kartenskripte(context.Context, int64) (any, error) {
	return []map[string]any{{"name": "Zwei Farben"}}, nil
}

func (williger) KartenskriptAusfuehren(context.Context, int64, string) (any, error) {
	return map[string]any{"karten": []string{"1", "2"}, "befunde": []string{}}, nil
}

// Die beiden Wege der Auszählung (#1338).
func (williger) Auszaehldimensionen(context.Context, int64) (any, error) {
	return []map[string]any{{"name": "farbe", "beschreibung": "je Farbe"}}, nil
}

// Die fünf Wege der Autoersetzung (#1341).
func (williger) Autoersetzungen(context.Context, int64) (any, error) {
	return map[string]any{
		"regeln": []map[string]any{{"muster": "(W)", "eigen": false}}, "an": true,
	}, nil
}

func (williger) AutoersetzungSetzen(context.Context, int64, string, string, bool, bool) error {
	return nil
}

func (williger) AutoersetzungLoeschen(context.Context, int64, string) error { return nil }

func (williger) AutoersetzungenZuruecksetzen(context.Context, int64) error { return nil }

func (williger) AutoersetzungSchalten(context.Context, int64, bool) error { return nil }

// Die Parameterhilfen (#1342).
func (williger) Parameterhilfen(context.Context, int64, int64) (any, error) {
	return map[string]any{
		"arten":    []map[string]any{{"name": "zahl", "einfuegen": "<atom-param>zahl</atom-param>"}},
		"verweise": []map[string]any{{"nummer": 1, "art": "zahl", "einfuegen": "{param1}"}},
		"befunde":  []string{},
	}, nil
}

func (williger) Auszaehlen(context.Context, int64, string) (any, error) {
	return map[string]any{
		"dimension": "farbe",
		"gruppen":   []map[string]any{{"name": "weiss", "karten": 2}},
		"befunde":   []string{},
	}, nil
}

// **Und das Nachzeichnen** (#1340).
func (williger) SymbolNachzeichnen(context.Context, int64, string) (any, error) {
	return map[string]any{"name": "symbol.mse-symbol", "teile": 1, "punkte": 12}, nil
}

// Pruefen meldet einen Befund in Zeile 3 — die Prüfung hat keinen Fehlerweg.
func (williger) Pruefen(string) (zeile, spalte int, meldung string) {
	return 3, 7, "hier fehlt ein Wert"
}

// FelderSetzen geht durch (#1390).
func (williger) FelderSetzen(context.Context, int64, []int64, string, string) (
	string, error,
) {
	return "vorgang7", nil
}

// FelderVorschau gibt eine Zeile je Karte; die zweite steht schon so (U6).
func (williger) FelderVorschau(_ context.Context, _ int64, karteIDs []int64, feld, wert string) (
	domain.Massenvorschau, error,
) {
	zeilen := make([]domain.Massenvorschauzeile, 0, len(karteIDs))
	for i, karteID := range karteIDs {
		zeile := domain.Massenvorschauzeile{KarteID: karteID, Name: "Karte", Alt: "Alt", GabAlt: true, Neu: wert}
		if i == 1 {
			zeile.Alt, zeile.Ausschluss = wert, domain.AusschlussGleich
		}

		zeilen = append(zeilen, zeile)
	}

	return domain.Massenvorschau{Feld: feld, Wert: wert, Zeilen: zeilen}, nil
}

func (williger) FelderSetzenMitErwartung(
	context.Context, int64, []domain.Feldvorwert, string, string,
) (string, error) {
	return "vorgang8", nil
}

// **Der Symboleditor** (#1391) — alles geht durch.
func (williger) SymbolLesen(context.Context, int64, string) (any, error) {
	return []map[string]any{{"art": "Form", "name": "probe"}}, nil
}

func (williger) SymbolAblegen(context.Context, int64, json.RawMessage, string) (string, error) {
	return "probe.mse-symbol", nil
}

func (williger) SymbolRechnen(string, json.RawMessage) (any, error) {
	return []map[string]any{{"art": "Form", "name": "gerechnet"}}, nil
}

func (williger) Palette(context.Context, int64, string) (any, error) {
	return []map[string]any{{"art": "code", "code": "W", "beschriftung": "Weiß"}}, nil
}

// Verschluesselungsstand meldet einen Klartextbestand.
func (williger) Verschluesselungsstand() (bool, error) { return false, nil }

func (williger) Sicherungsstand(context.Context) (any, error) {
	return map[string]any{"stufe": "frisch", "tage": 2}, nil
}
