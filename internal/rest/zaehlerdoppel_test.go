// **Das mitzählende Doppel.**
//
// Aus `vorrichtung_test.go` herausgelöst, als die Datei über 400 Zeilen wuchs
// (§2) — der Anlass war das Löschen eines Sets
// (#1218).
// Der Schnitt folgt der Frage: dort steht, **wie** gerufen wird, hier,
// **wogegen**.

package rest_test

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// zaehlender ist ein `williger`, der mitzählt, wie oft er gerufen wurde.
//
// **Damit „gültig" den Weg bis zum Anwendungsfall belegt** und nicht nur einen
// Status. Gezählt wird über **alle** Verfahren zusammen: je Endpunkt läuft
// genau eines, und welches, sagt schon der Weg.
type zaehlender struct {
	williger

	schranke sync.Mutex
	rufe     int

	// kontext ist der, mit dem [zaehlender.Setkoepfe] zuletzt gerufen wurde.
	kontext context.Context //nolint:containedctx // siehe dort

	// werte sind die, mit denen [zaehlender.VorgabeSetzen] zuletzt gerufen
	// wurde — siehe dort.
	werte []string

	// sitze sind die, mit denen [zaehlender.PartieAnlegen] zuletzt gerufen
	// wurde (UXR-F21, #2041): geprüft wird, was ankommt, und nicht, dass
	// gerufen wurde.
	sitze []rest.Sitzwunsch

	// abgewaehlt ist die Abwahl, mit der [zaehlender.PartieAnlegen] zuletzt
	// gerufen wurde (P1a, #2335).
	abgewaehlt []string
}

// Anlegen legt eine Karte aus einem Scryfall-Namen an (#1032).
func (z *zaehlender) Anlegen(context.Context, int64, string) (rest.Anlage, error) {
	z.rufe++

	return rest.Anlage{KarteID: 7, Felder: []string{"name"}, Befunde: []string{}}, nil
}

// ListeDeuten deutet eine Kartenliste und zählt mit (#1292).
func (z *zaehlender) ListeDeuten(context.Context, string) (any, error) {
	z.rufe++

	return map[string]any{"zeilen": []any{}, "befunde": []string{}}, nil
}

// ListeEinlesen nimmt Zeilen in die Sammlung und zählt mit (#1292).
func (z *zaehlender) ListeEinlesen(context.Context, []rest.Listeneinfuhrzeile, string) (any, error) {
	z.rufe++

	return map[string]any{"angelegt": []int64{1}, "erhoeht": 0, "stueck": 1}, nil
}

// DecklisteVorschau deutet eine Deckliste und zählt mit (T8, #1473).
func (z *zaehlender) DecklisteVorschau(context.Context, int64, string) (any, error) {
	z.rufe++

	return map[string]any{
		"deckID": 1, "stand": "abc", "zeilen": []any{}, "befunde": []string{},
	}, nil
}

// DecklisteUebernehmen legt Zeilen ins Deck und zählt mit (T8, #1473).
func (z *zaehlender) DecklisteUebernehmen(
	context.Context, int64, string, []rest.Deckeinfuhrzeile,
) (any, error) {
	z.rufe++

	return map[string]any{"gelegt": 1, "stueck": 4, "stand": "cde"}, nil
}

// SammlungAusgeben gibt die Sammlung und zählt mit (#1292).
func (z *zaehlender) SammlungAusgeben(context.Context, string) (string, error) {
	z.rufe++

	return "printing,menge\nprobe-druck,1\n", nil
}

// SammlungCSVEinlesen liest ein CSV zurück und zählt mit (#1292).
func (z *zaehlender) SammlungCSVEinlesen(context.Context, string, string, string) (any, error) {
	z.rufe++

	return map[string]any{"angelegt": []int64{1}, "erhoeht": 0, "stueck": 1}, nil
}

// Auswerten rechnet einen Ausdruck aus und zählt mit (#1070).
func (z *zaehlender) Auswerten(_ context.Context, _ int64, ausdruck string) (
	rest.Skriptergebnis, error,
) {
	z.rufe++

	return rest.Skriptergebnis{
		Quelltext: `"` + ausdruck + `"`, Text: ausdruck,
		Art: "Zeichenkette", Befunde: []string{},
	}, nil
}

// Die drei Züge der Paketwerkstatt — sie zählen mit (#1164).
func (z *zaehlender) DateienUndBeiwerk(context.Context, string) (dateien, beiwerk []string, fehler error) {
	z.rufe++

	return []string{"style"}, []string{}, nil
}

// Der Dateibaum (FMT-02, #1691).
func (z *zaehlender) DateiAnlegen(context.Context, string, string, string) (string, error) {
	z.rufe++

	return "neu", nil
}

func (z *zaehlender) DateiUmbenennen(context.Context, string, string, string, string) error {
	z.rufe++

	return nil
}

func (z *zaehlender) DateiEntfernen(context.Context, string, string, string) error {
	z.rufe++

	return nil
}

func (z *zaehlender) Lesen(context.Context, string, string) (inhalt, stand string, fehler error) {
	z.rufe++

	return "mse version: 2.0.0\n", "s", nil
}

func (z *zaehlender) Schreiben(context.Context, string, string, string, string) (string, error) {
	z.rufe++

	return "neu", nil
}

// Namen zählt auf, was die Umgebung kennt, und zählt mit (#1164).
func (z *zaehlender) Namen(context.Context, int64) ([]string, error) {
	z.rufe++

	return []string{"card", "length", "set"}, nil
}

// Einlesen liest einen ganzen Satz als Set ein (#1033).
func (z *zaehlender) Einlesen(context.Context, string, string) (rest.Satz, error) {
	z.rufe++

	return rest.Satz{SetID: 9, Karten: 103, Befunde: []string{}}, nil
}

func (z *zaehlender) wieOft() int {
	z.schranke.Lock()
	defer z.schranke.Unlock()

	return z.rufe
}

func (z *zaehlender) zaehle() {
	z.schranke.Lock()
	defer z.schranke.Unlock()

	z.rufe++
}

// Setkoepfe zählt mit und **merkt sich den Kontext**.
//
// **Der Kontext ist ein eigener Prüfgegenstand** und nicht nur ein Parameter:
// an ihm hängt das Profil, in dem gearbeitet wird, und ein Zugriffstoken
// bringt ein anderes mit als der Bestand (Werkstück 10, #1351). Gemerkt wird
// er hier und nicht an einem Zeugen zwischen Wache und Wegetafel — der
// bezeugte nur die Wache, nicht den Weg bis zum Anwendungsfall.
func (z *zaehlender) Setkoepfe(ctx context.Context) ([]domain.SetHead, error) {
	z.zaehle()
	z.kontext = ctx

	return z.williger.Setkoepfe(ctx)
}

func (z *zaehlender) Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error) {
	z.zaehle()

	return z.williger.Karten(ctx, setID)
}

func (z *zaehlender) FeldSetzenMitVorwert(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	z.zaehle()

	return z.williger.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)
}

func (z *zaehlender) Einsetzen(
	ctx context.Context, setID, karteID int64, feld, name, art string,
) (rest.Bildeinsatz, error) {
	z.zaehle()

	return z.williger.Einsetzen(ctx, setID, karteID, feld, name, art)
}

func (z *zaehlender) Stand(ctx context.Context) (rest.Sammelstand, error) {
	z.zaehle()

	return z.williger.Stand(ctx)
}

func (z *zaehlender) Holen(ctx context.Context) (rest.Sammelstand, error) {
	z.zaehle()

	return z.williger.Holen(ctx)
}

func (z *zaehlender) Regelhinweise(
	ctx context.Context, name string,
) (rest.Regelauskunft, error) {
	z.zaehle()

	return z.williger.Regelhinweise(ctx, name)
}

func (z *zaehlender) Manakosten(ctx context.Context, kosten string) ([]string, error) {
	z.zaehle()

	return z.williger.Manakosten(ctx, kosten)
}

func (z *zaehlender) Vorschlag(
	ctx context.Context, setID, karteID int64, name string,
) (rest.Vorschlag, error) {
	z.zaehle()

	return z.williger.Vorschlag(ctx, setID, karteID, name)
}

func (z *zaehlender) Uebernehmen(
	ctx context.Context, setID, karteID int64, name string, felder []string,
) (vorher []rest.Aenderung, fehler error) {
	z.zaehle()

	return z.williger.Uebernehmen(ctx, setID, karteID, name, felder)
}

// **Die Sammlung** (#1165).
func (z *zaehlender) Sammlung(context.Context) ([]rest.Exemplarsicht, error) {
	z.rufe++

	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}}, nil
}

func (z *zaehlender) Sammlungsfenster(
	context.Context, string, string, string, int, int,
) (fenster []rest.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	z.rufe++

	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}},
		domain.Fensterzahlen{Gesamt: 1, Insgesamt: 1, Stueck: 1}, nil
}

func (z *zaehlender) SammlungFuer(context.Context, string) ([]rest.Exemplarsicht, error) {
	z.rufe++

	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}}, nil
}

func (z *zaehlender) ExemplarAnlegen(context.Context, rest.Exemplarsicht) (int64, error) {
	z.rufe++

	return 1, nil
}

func (z *zaehlender) ExemplarAendern(context.Context, rest.Exemplarsicht) error {
	z.rufe++

	return nil
}

func (z *zaehlender) ExemplarLoeschen(context.Context, int64, int64) error {
	z.rufe++

	return nil
}
