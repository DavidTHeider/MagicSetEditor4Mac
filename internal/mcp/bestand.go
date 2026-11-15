package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Leser ist, was die lesenden Werkzeuge brauchen.
//
// Die Schnittstelle sagt, was `mcp` **braucht** — nicht, was `storage` anzubieten
// hat (§2).
type Leser interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
	Felder(ctx context.Context, setID int64) ([]string, error)
	Suche(ctx context.Context, frage domain.SearchQuery) (domain.SearchResult, error)
}

// Schreiber ist, was die schreibenden Werkzeuge brauchen.
//
// # Warum beide Methoden einen Vorwert geben
//
// Eine Änderung gibt den vorigen Wert zurück, damit sie sich zurücknehmen lässt
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). **Lesen
// und Schreiben müssen dafür ein Zug sein**: der erste Anlauf las erst und
// schrieb danach, und im ausdrücklich normalen Zwei-Schreiber-Betrieb konnte
// das Fenster dazwischen denselben Wert ändern. Der Zweitblick über Phase 9 hat
// es gefunden.
//
// **`FeldEntfernen` gehört dazu, weil eine halbe Rücknahme keine ist.** Ein
// Feld, das es vorher nicht gab, lässt sich nicht durch Leeren wiederherstellen:
// eine leere Zeile ist nicht dasselbe wie keine Zeile.
type Schreiber interface {
	// Beide geben: den vorigen Wert, ob es ihn gab, und einen Fehler.
	FeldSetzenMitVorwert(ctx context.Context, setID, karteID int64, feld, wert string) (vorher string, hatte bool, fehler error)
	FeldEntfernen(ctx context.Context, setID, karteID int64, feld string) (vorher string, hatte bool, stelle *int64, fehler error)

	// FeldSetzenAnStelle setzt ein Feld an die Stelle, an der es stand — der
	// zweite Teil einer vollständigen Rücknahme
	// (#160).
	FeldSetzenAnStelle(ctx context.Context, setID, karteID int64, feld, wert string, stelle int64) (vorher string, hatte bool, fehler error)

	// FelderSetzen setzt **ein** Feld auf **mehreren** Karten, in einer
	// Transaktion (#1390).
	//
	// **Die Atomarität ist die Fähigkeit**, nicht eine Bequemlichkeit: eine
	// Schleife über `FeldSetzenMitVorwert` kann halb durch sein, dieser Weg
	// nicht.
	//
	// **Sie gibt den Vorgang zurück** (#1393) — mit ihm nimmt
	// »schreiben.zuruecknehmen« genau diese Änderung zurück.
	FelderSetzen(ctx context.Context, setID int64, karteIDs []int64, feld, wert string) (
		vorgang string, fehler error)

	// FelderVorschau sagt, was der Wert auf den Karten hieße — ohne zu
	// schreiben (Werkstück U6, #1417).
	FelderVorschau(ctx context.Context, setID int64, karteIDs []int64, feld, wert string) (
		domain.Massenvorschau, error)

	// FelderSetzenMitErwartung schreibt nur, wenn jede Karte noch auf dem
	// Altwert steht, den die Vorschau gezeigt hat — sonst ein Konflikt, der
	// die abweichenden Karten nennt.
	FelderSetzenMitErwartung(ctx context.Context, setID int64, erwartet []domain.Feldvorwert, feld, wert string) (
		vorgang string, fehler error)

	// **Und der Lebenszyklus der Karte selbst** — anlegen, kopieren, löschen,
	// umordnen. Er steht hier und nicht in einer eigenen Schnittstelle, weil
	// eine Zusicherung zur Laufzeit ihn stillschweigend verlieren ließe
	// (#591).
	Kartenschreiber
}

// lesende gibt die Werkzeuge, die nichts ändern.
func lesende(leser Leser) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.sets", Art: Lesend, Idempotent: true,
			Erklaerung: "Zählt die geöffneten Sets auf: Kennung, Name, Spiel, Kartenzahl, Herkunft.",
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				koepfe, err := leser.Setkoepfe(ctx)
				if err != nil {
					return "", err
				}

				return alsJSON(alsSetzeilen(koepfe))
			},
		},
		{
			Name: "lesen.karten", Art: Lesend, Idempotent: true,
			Erklaerung: "Gibt die Karten eines Sets mit allen Feldwerten.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			}, "setID"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				wahl, err := kennung(argumente)
				if err != nil {
					return "", err
				}

				karten, err := leser.Karten(ctx, wahl)
				if err != nil {
					return "", err
				}

				return alsJSON(alsKartenzeilen(karten))
			},
		},
		{
			Name: "lesen.felder", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die Feldnamen, die in einem Set vorkommen.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			}, "setID"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				wahl, err := kennung(argumente)
				if err != nil {
					return "", err
				}

				felder, err := leser.Felder(ctx, wahl)
				if err != nil {
					return "", err
				}

				return alsJSON(liste(felder))
			},
		},
		{
			Name: "lesen.suche", Art: Lesend, Idempotent: true,
			Erklaerung: "Sucht im Volltext über alle Kartenfelder. »feld« und »setID« schränken ein.",
			Schema: objekt(map[string]any{
				"text":  text("Wonach gesucht wird."),
				"feld":  text("Nur in diesem Feld suchen. Leer: in allen."),
				"setID": zahl("Nur in diesem Set suchen. 0: in allen."),
			}, "text"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				var frage struct {
					Text  string `json:"text"`
					Feld  string `json:"feld"`
					SetID int64  `json:"setID"`
				}

				if err := json.Unmarshal(argumente, &frage); err != nil {
					return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
				}

				ergebnis, err := leser.Suche(ctx, domain.SearchQuery{
					// **`setID` wird hier nicht geprüft, und das ist der
					// Unterschied**: sie steht nicht in `required`. Ohne sie
					// sucht dieses Werkzeug über **alle** Sets, und eine Null
					// heißt „nicht eingeschränkt" statt „keine Kennung".
					Text: frage.Text, Field: frage.Feld, SetID: frage.SetID,
				})
				if err != nil {
					return "", err
				}

				ergebnis.Hits = liste(ergebnis.Hits)

				return alsJSON(alsSuchantwort(ergebnis))
			},
		},
	}
}

// schreibende gibt die Werkzeuge, die ändern.
//
// **Genau eines**, und es betrifft **ein** Feld **einer** Karte. Es gibt keines,
// das ein Set überschreibt, keines, das eine Karte löscht, und keines für
// mehrere Karten — obwohl der Bestand das könnte. Der Grund steht in
// [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md).
//
// Ein **Feld** wegnehmen kann es: das ist der Weg, eine Änderung zurückzunehmen,
// die es erst angelegt hat.
func schreibende(schreiber Schreiber) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.feld", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: setzt ein Feld einer Karte, oder nimmt es weg. " +
				"Gibt den vorigen Wert zurück und sagt, wie sich die Änderung zurücknehmen lässt.",
			Schema: objekt(map[string]any{
				"setID":       zahl("Die Kennung des Sets, in dem die Karte liegt."),
				"karteID":     zahl("Die Kennung der Karte aus »lesen.karten«."),
				"feld":        text("Der Feldname, etwa »name« oder »rule text«."),
				"wert":        text("Der neue Wert. Eine leere Zeichenkette leert das Feld — das ist nicht dasselbe wie »entfernen«. Fehlt »wert« und ist »entfernen« nicht gesetzt, geschieht nichts."),
				"entfernen":   jaNein("Nimmt das Feld von der Karte. Der einzige Weg, eine Änderung zurückzunehmen, die es erst angelegt hat."),
				"reihenfolge": zahl("Die Stelle, an die das Feld gehört — aus »reihenfolge« einer früheren Entfernung. Ohne sie kommt ein neu gesetztes Feld ans Ende der Karte."),
			}, "setID", "karteID", "feld"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				// **`Wert` als Zeiger, damit „fehlt" von „leer" zu
				// unterscheiden ist.** Seit `entfernen` dazukam, ist `wert`
				// kein Pflichtfeld mehr — und ein Aufruf, der beides weglässt,
				// leerte das Feld. Der Nachlauf zum Zweitblick über Phase 9 hat
				// es gefunden. Dieselbe Unterscheidung wie überall sonst hier:
				// „ist leer" und „gibt es nicht" sind zwei Aussagen.
				var wahl struct {
					SetID     int64   `json:"setID"`
					KarteID   int64   `json:"karteID"`
					Feld      string  `json:"feld"`
					Wert      *string `json:"wert"`
					Entfernen bool    `json:"entfernen"`

					// **Als Zeiger**, denn die 0 ist die erste Stelle und
					// nicht „keine" (#160).
					Reihenfolge *int64 `json:"reihenfolge"`
				}

				if err := json.Unmarshal(argumente, &wahl); err != nil {
					return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
				}

				if strings.TrimSpace(wahl.Feld) == "" {
					return "", domain.NewFieldError("feld", domain.SchluesselPflichtfeld,
						"Ohne Feldnamen ist nicht zu sagen, was zu ändern wäre.")
				}

				if wahl.Wert == nil && !wahl.Entfernen {
					return "", domain.NewFieldError("wert", domain.SchluesselPflichtfeld,
						"Es fehlt »wert«. Wer ein Feld leeren will, gibt »wert« als leere "+
							"Zeichenkette an; wer es wegnehmen will, »entfernen«.")
				}

				wert := ""
				if wahl.Wert != nil {
					wert = *wahl.Wert
				}

				// **Die Kennung wird geprüft, bevor der Bestand sie sieht**
				// (§8). Vorher lief eine Null durch bis zum Schreiber und kam
				// als „nicht gefunden" zurück — richtig abgewiesen, aber am
				// falschen Ort und ohne den Feldnamen.
				if err := pruefeKennungen(wahl.SetID, wahl.KarteID); err != nil {
					return "", err
				}

				// **Lesen und Ändern in einem Zug** — sonst gehörte der Vorwert
				// zu einem Stand, den ein zweiter Schreiber schon überholt hat.
				vorher, hatte, stelle, err := aendern(ctx, schreiber, wahl.SetID,
					wahl.KarteID, wahl.Feld, wert, wahl.Entfernen, wahl.Reihenfolge)
				if err != nil {
					return "", err
				}

				antwort := map[string]any{
					"setID":        wahl.SetID,
					"karteID":      wahl.KarteID,
					"feld":         wahl.Feld,
					"entfernt":     wahl.Entfernen,
					"wert":         wert,
					"vorher":       vorher,
					"vorherDa":     hatte,
					"zuruecknahme": zuruecknahme(hatte, stelle),
				}

				// **Die Stelle nur, wenn es eine gab.** Ein `null` im JSON
				// hieße „hatte keine" und läse sich wie „unbekannt"; wer sie
				// weglässt, sagt dasselbe ohne die zweite Lesart.
				if stelle != nil {
					antwort["reihenfolge"] = *stelle
				}

				return alsJSON(antwort)
			},
		},
	}
}

// aendern führt die gewählte Änderung aus.
func aendern(
	ctx context.Context, schreiber Schreiber, setID, karteID int64, feld, wert string,
	entfernen bool, stelle *int64,
) (vorher string, hatte bool, gestanden *int64, fehler error) {
	if entfernen {
		return schreiber.FeldEntfernen(ctx, setID, karteID, feld)
	}

	// **Mit Stelle heißt: an ihre Stelle zurück** (#160). Ohne sie hängt ein
	// neu gesetztes Feld hinten an, und das ist für eine Rücknahme falsch.
	if stelle != nil {
		vorher, hatte, fehler = schreiber.FeldSetzenAnStelle(
			ctx, setID, karteID, feld, wert, *stelle)

		return vorher, hatte, nil, fehler
	}

	vorher, hatte, fehler = schreiber.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)

	return vorher, hatte, nil, fehler
}

// zuruecknahme sagt, wie sich die eben gemachte Änderung umkehren lässt.
//
// **Zwei Wege, und der Unterschied zählt.** Gab es das Feld vorher nicht, führt
// kein Wert zurück — nur das Entfernen tut es. Der erste Anlauf nannte nur den
// einen und ließ die halbe Zusage stehen.
func zuruecknahme(hatte bool, stelle *int64) string {
	if hatte {
		if stelle != nil {
			return "schreiben.feld mit »wert« = »vorher« und »reihenfolge« = »reihenfolge«"
		}

		return "schreiben.feld mit »wert« = »vorher«"
	}

	return "schreiben.feld mit »entfernen« = true — das Feld gab es vorher nicht"
}

// kennung liest die Setkennung aus den Argumenten.
//
// **Ohne Feldnamen als Parameter**, seit alle vier Aufrufer denselben
// übergeben (#1107). Ein Parameter, der nur einen Wert annimmt, sieht aus wie
// eine Wahl und ist keine — wer eine Kartenkennung so lesen will, nimmt
// [pruefeKennungen], das beide auf einmal prüft.
func kennung(argumente json.RawMessage) (int64, error) {
	const name = "setID"

	var werte map[string]json.Number

	if err := json.Unmarshal(argumente, &werte); err != nil {
		return 0, fmt.Errorf("die Argumente sind kein JSON: %w", err)
	}

	roh, da := werte[name]
	if !da {
		return 0, domain.NewFieldError(name, domain.SchluesselPflichtfeld,
			"Es fehlt %q.", name)
	}

	wert, err := roh.Int64()
	if err != nil {
		// **Ohne den `strconv`-Text** (Befund 3 des Zweitblicks über Phase
		// 4.13): „invalid syntax" sagt dem Rufer nichts, was er nicht schon
		// weiß — der Name des Feldes und die Erwartung sagen alles.
		return 0, domain.NewFieldError(name, domain.SchluesselValidierung,
			"%q ist keine Kennung — erwartet wird eine ganze Zahl", name)
	}

	// **Und null ist keine.** Bis zum 2026-08-22 fehlte diese Zeile, und die
	// lesenden Werkzeuge gaben für `setID: 0` ein leeres `[]` zurück — ein
	// Gegenüber schloss daraus „das Set ist leer" statt „die Kennung ist
	// keine" (#591).
	if err := pruefeKennung(name, wert); err != nil {
		return 0, err
	}

	return wert, nil
}

// liste macht aus einer fehlenden Liste eine leere.
//
// **Immer `[]`, nie `null`** — dieselbe Zusage, die `internal/app` der
// Oberfläche gibt, und aus demselben Grund: ein Gegenüber darf darüber laufen,
// ohne vorher zu prüfen. Ein Modell, das `null` bekommt, hat die Wahl zwischen
// einem Abbruch und einer Vermutung.
//
// Gefunden hat das die Sitzung von Anfang bis Ende gegen einen **leeren**
// Bestand: `lesen.sets` gab `null`.
func liste[T any](werte []T) []T {
	if werte == nil {
		return []T{}
	}

	return werte
}

// alsJSON gibt einen Wert als eingerückten Text.
//
// **Eingerückt, weil ein Modell es liest.** Eine Zeile mit 3.000 Karten ist
// gültig und unlesbar.
//
// **Ohne Fehlerweg**: die Werte hier sind Domänenstrukturen, Listen und
// Abbildungen aus Zeichenketten — nichts davon kann `MarshalIndent` scheitern
// lassen. Ein Zweig, den kein Aufruf betreten kann, wäre eine Behauptung
// darüber, was im Ernstfall geschieht (§3).
func alsJSON(wert any) (string, error) {
	roh, _ := json.MarshalIndent(wert, "", "  ") //nolint:errcheck // siehe darüber

	return string(roh), nil
}
