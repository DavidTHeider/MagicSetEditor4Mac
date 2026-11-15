package rest_test

// **Die Deckwerkstatt in den Vorrichtungen** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Fünfzehn Methoden an zwei Doppeln**: das willige antwortet, das unwillige
// gibt seinen Fehler. Beides in einer Datei, damit der Zusammenhang sichtbar
// bleibt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (w williger) Decks(context.Context, string) (any, error) {
	return []any{}, nil
}

func (w williger) Deck(context.Context, int64) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) DeckAnlegen(context.Context, string, string, string) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) DeckUmbenennen(context.Context, int64, int64, string, string) (
	any, error,
) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) DeckLoeschen(context.Context, int64) error { return nil }

func (w williger) DeckKopieren(context.Context, int64, string) (any, error) {
	return map[string]any{"id": "2"}, nil
}

func (w williger) DeckAusSet(context.Context, int64, string, int) (any, error) {
	return map[string]any{"id": "3"}, nil
}

func (w williger) DeckFehlkartenWuenschen(context.Context, int64) (any, error) {
	return map[string]any{"neu": []any{}}, nil
}

func (w williger) DeckplatzLegen(
	context.Context, int64, int64, string, int, string,
) (any, error) {
	return map[string]any{"id": "2"}, nil
}

func (w williger) DeckplatzNehmen(context.Context, int64, int64) error { return nil }

// **Der Willige ist bereit** — der Betriebszustand hängt an keinem Deck.
func (w williger) Gesundheit(context.Context) (
	zustand any, bereit bool, fehler error,
) {
	return map[string]any{"lebt": true, "bereit": true}, true, nil
}

func (w williger) DeckplatzVerschieben(
	_ context.Context, _, platzID int64, bereich string,
) (any, error) {
	return map[string]any{"id": platzID, "bereich": bereich}, nil
}

func (w williger) DeckFormatWaehlen(context.Context, int64, int64) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) DeckPruefen(context.Context, int64) (any, error) {
	return map[string]any{"urteil": map[string]any{}}, nil
}

func (w williger) DeckAbgleichen(context.Context, int64) (any, error) {
	return map[string]any{"zeilen": []any{}}, nil
}

// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
func (w williger) DeckZuordnungen(context.Context, int64) (any, error) {
	return map[string]any{"zeilen": []any{}, "fehlend": 0, "unbekannt": 0}, nil
}

func (w williger) DeckZuordnen(_ context.Context, _, platzID, exemplarID int64, menge int) (any, error) {
	return map[string]any{
		"id": "1", "platzID": platzID, "exemplarID": exemplarID, "menge": menge,
	}, nil
}

func (w williger) DeckZuordnungLoesen(context.Context, int64, int64) error { return nil }

func (w williger) DeckFehlkarten(context.Context, int64) (any, error) {
	return map[string]any{"liste": "3 Lightning Bolt (M11) 149\n"}, nil
}

func (w williger) Deckliste(context.Context, int64) (any, error) {
	return map[string]any{"liste": "Deck\n4 Lightning Bolt (M11) 149\n"}, nil
}

func (w williger) ExemplarDecks(context.Context, int64) (any, error) {
	return []any{map[string]any{"deckID": "1", "name": "Aggro", "menge": 2}}, nil
}

func (w williger) Bauarten(context.Context) (any, error) { return []any{}, nil }

func (w williger) Formate(context.Context, string) (any, error) { return []any{}, nil }

func (w williger) FormatAusBauart(context.Context, string, string, string, string) (
	any, error,
) {
	return map[string]any{"id": "5"}, nil
}

func (w williger) FormatSetzen(context.Context, map[string]any) (any, error) {
	return map[string]any{"id": "5"}, nil
}

func (w williger) FormatLoeschen(context.Context, int64) error { return nil }

func (u unwilliger) Decks(context.Context, string) (any, error) { return nil, u.fehler }

func (u unwilliger) Deck(context.Context, int64) (any, error) { return nil, u.fehler }

func (u unwilliger) DeckAnlegen(context.Context, string, string, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckUmbenennen(context.Context, int64, int64, string, string) (
	any, error,
) {
	return nil, u.fehler
}

func (u unwilliger) DeckLoeschen(context.Context, int64) error { return u.fehler }

func (u unwilliger) DeckKopieren(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckAusSet(context.Context, int64, string, int) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckFehlkartenWuenschen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckplatzLegen(
	context.Context, int64, int64, string, int, string,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckplatzNehmen(context.Context, int64, int64) error { return u.fehler }

func (u unwilliger) DeckplatzVerschieben(
	context.Context, int64, int64, string,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckFormatWaehlen(context.Context, int64, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckPruefen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckAbgleichen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckZuordnungen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckZuordnen(context.Context, int64, int64, int64, int) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DeckZuordnungLoesen(context.Context, int64, int64) error { return u.fehler }

func (u unwilliger) DeckFehlkarten(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Deckliste(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) ExemplarDecks(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Bauarten(context.Context) (any, error) { return nil, u.fehler }

func (u unwilliger) Formate(context.Context, string) (any, error) { return nil, u.fehler }

func (u unwilliger) FormatAusBauart(context.Context, string, string, string, string) (
	any, error,
) {
	return nil, u.fehler
}

func (u unwilliger) FormatSetzen(context.Context, map[string]any) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) FormatLoeschen(context.Context, int64) error { return u.fehler }

// **Und er nennt seinen Zählstand** — die Metriken hängen an keinem Deck.
func (w williger) Metriken(context.Context) (any, error) {
	return map[string]any{"reihen": []any{}, "grenze": 40}, nil
}

// **Und er sagt, wer er ist** — auch das hängt an keinem Deck.
func (w williger) Meta(context.Context) (any, error) {
	return map[string]any{"fassung": "0.4.0", "flaeche": "dienst"}, nil
}

// **Und er sagt, ob die neueste Freigabe passt** — auch das hängt an keinem
// Deck.
func (w williger) Vertraeglichkeit(context.Context) (any, error) {
	return map[string]any{"urteil": "vertraeglich", "hindernisse": []any{}}, nil
}

// **Und was zurückgehen kann** (O7, #1431) — vier Wege, die an keinem Deck
// hängen.
func (w williger) Ruecksprungarten(context.Context) (any, error) {
	return map[string]any{"arten": []any{}}, nil
}

func (w williger) Ruecksprunglage(context.Context, string, string) (any, error) {
	return map[string]any{"zeilen": []any{}, "unklar": []any{}}, nil
}

func (w williger) Rueckkehren(context.Context) (any, error) {
	return map[string]any{"rueckkehren": []any{}}, nil
}

func (w williger) RueckkehrQuittieren(context.Context, int64) error { return nil }

// **Und was im Bestand herumliegt** (O8, #1432) — auch das hängt an keinem
// Deck.
func (w williger) Pflegevorschau(context.Context) (any, error) {
	return map[string]any{"befunde": []any{}}, nil
}

func (w williger) Pflegen(context.Context, string, []rest.Pflegewahl) (any, error) {
	return map[string]any{"zeilen": []any{}}, nil
}

// **Und das Kostenjournal** (O12, #1439).
func (w williger) Kostenstand(context.Context, string, string) (any, error) {
	return map[string]any{"verfuegbar": map[string]any{"wert": 0}}, nil
}

func (w williger) Kostenposten(context.Context, string, string) (any, error) {
	return map[string]any{"posten": []any{}}, nil
}

func (w williger) KostenbudgetSetzen(
	context.Context, string, string, int64, string,
) (any, error) {
	return map[string]any{"fehler": nil}, nil
}

// **Und was eine Rücksetzung anfassen würde** (O4, #1428).
func (w williger) Ruecksetzungsarten(context.Context) (any, error) {
	return []string{"layout"}, nil
}

func (w williger) Ruecksetzungsvorschau(context.Context, string, string, string) (any, error) {
	return map[string]any{"betroffen": []any{}, "hindernisse": []any{}}, nil
}

func (w williger) Zuruecksetzen(
	context.Context, string, string, string, string, string, string,
) (any, error) {
	return map[string]any{"zurueckgesetzt": 0}, nil
}
