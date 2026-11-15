// **Ein Feld auf vielen Karten setzen — über MCP**
// (#1390).
//
// # Warum das ein eigenes Werkzeug ist und keine Schleife
//
// **Weil die Atomarität die Fähigkeit ist.** Ein Modell, das ein Feld über
// zweihundert Karten ändert, ruft sonst zweihundertmal »schreiben.feld«; ein
// Fehler in der Mitte lässt das Set halb geändert zurück, und niemand weiß,
// welche Hälfte. Dieser Weg schreibt in **einer** Transaktion.
//
// # Warum es das bis heute nicht gab
//
// `docs/FAEHIGKEITEN.md` führte die Fähigkeit als *Viele Felder setzen* und
// begründete die Fensterbindung damit, das Fenster schreibe „die Eingaben
// **einer** Karte in einem Zug, weil eine Maske viele Felder hat". Die Bindung
// setzt **ein** Feld auf **vielen** Karten; die Begründung beschrieb eine
// andere Operation. Gefunden bei der Prüfung der 24 Flächenlücken (Werkstück E
// von #1324).

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// massenfeldwerkzeug gibt das Werkzeug der Massenänderung.
//
// **Es nimmt [Schreiber] und keine eigene Schnittstelle.** Eine zweite Naht
// für eine Methode zöge einen weiteren Parameter durch `werkzeugliste.go` und
// durch jedes Doppel — für eine Zusage, die dort ohnehin gilt: alles, was
// dieser Server schreibt, geht über dieselbe Aufsicht.
func massenfeldwerkzeug(schreiber Schreiber) []Werkzeug {
	return []Werkzeug{{
		Name: "schreiben.felder", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: setzt DASSELBE Feld auf MEHREREN Karten eines " +
			"Sets — in einer Transaktion. Entweder sind alle geändert oder keine; " +
			"mit »erwartet« (aus »lesen.felder.vorschau«) nur, wenn jede Karte noch " +
			"auf dem gezeigten Altwert steht — sonst ein Konflikt, der die " +
			"abweichenden Karten nennt. " +
			"eine Karte, die nicht zu diesem Set gehört, lässt den ganzen Aufruf " +
			"scheitern und ändert nichts. Berechnete und nicht editierbare Felder " +
			"werden abgewiesen. " +
			"ZUM ZURÜCKNEHMEN den zurückgegebenen »Vorgang« an " +
			"»schreiben.zuruecknehmen« geben — dann geht GENAU diese Änderung " +
			"zurück, auch wenn inzwischen jemand anderes in dasselbe Set " +
			"geschrieben hat. Ohne ihn nimmt »schreiben.zuruecknehmen« den " +
			"JÜNGSTEN Schritt des Sets.",
		Schema: objekt(map[string]any{
			"setID":    zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteIDs": zahlenliste("Die Kennungen der Karten aus »lesen.karten«.", "Eine Kartenkennung."),
			"feld":     text("Der Feldname, etwa »rarity« oder »artist«."),
			"wert":     text("Der neue Wert für alle genannten Karten. Eine leere Zeichenkette leert das Feld."),
			"erwartet": erwartungsschema(),
		}, "setID", "karteIDs", "feld", "wert"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl massenwahl
			if err := lies(argumente, &wahl); err != nil {
				return "", err
			}

			if err := wahl.pruefen(); err != nil {
				return "", err
			}

			var (
				vorgang string
				err     error
			)

			// **Mit Erwartung geht es den vergleichenden Weg** (U6, #1417).
			if len(wahl.Erwartet) == 0 {
				vorgang, err = schreiber.FelderSetzen(
					ctx, wahl.SetID, wahl.KarteIDs, wahl.Feld, wahl.Wert)
			} else {
				vorgang, err = schreiber.FelderSetzenMitErwartung(
					ctx, wahl.SetID, wahl.Erwartet, wahl.Feld, wahl.Wert)
			}

			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{
				"Geaendert": len(wahl.KarteIDs), "Vorgang": vorgang,
			})
		},
	}}
}

// massenwahl sind die Argumente von »schreiben.felder« und
// »lesen.felder.vorschau« — dieselben, bis auf die Erwartung.
type massenwahl struct {
	SetID    int64                `json:"setID"`
	KarteIDs []int64              `json:"karteIDs"`
	Feld     string               `json:"feld"`
	Wert     string               `json:"wert"`
	Erwartet []domain.Feldvorwert `json:"erwartet"`
}

// pruefen weist unbrauchbare Kennungen ab — **jede einzeln**, bevor
// irgendetwas geschrieben wird (§8) — und eine Erwartung für eine Karte, die
// nicht in `karteIDs` steht: sonst schriebe der Ruf Karten, die er gar nicht
// nennt.
func (w massenwahl) pruefen() error {
	if err := pruefeKennung("setID", w.SetID); err != nil {
		return err
	}

	genannt := make(map[int64]bool, len(w.KarteIDs))

	for _, karteID := range w.KarteIDs {
		if err := pruefeKennung("karteIDs", karteID); err != nil {
			return err
		}

		genannt[karteID] = true
	}

	for _, vorwert := range w.Erwartet {
		if !genannt[vorwert.KarteID] {
			return fmt.Errorf("%w: »erwartet« nennt die Karte %d, die nicht in »karteIDs« steht",
				domain.ErrValidation, vorwert.KarteID)
		}
	}

	return nil
}

// erwartungsschema beschreibt »erwartet«: je Karte der Altwert aus der
// Vorschau.
func erwartungsschema() map[string]any {
	return objektliste(
		"Wahlfrei: je Karte der Altwert, den »lesen.felder.vorschau« gezeigt hat. "+
			"Steht sie inzwischen anders da, wird NICHTS geschrieben.",
		objekt(map[string]any{
			"karteID": zahl("Die Kartenkennung — muss in »karteIDs« stehen."),
			"wert":    text("Der gezeigte Altwert."),
			"gabEs":   jaNein("Ob die Karte das Feld hatte — eine leere Zeile ist nicht dasselbe wie keine."),
		}, "karteID", "gabEs"),
	)
}

// massenvorschauwerkzeug gibt das lesende Werkzeug zur Massenänderung (U6,
// #1417).
func massenvorschauwerkzeug(schreiber Schreiber) []Werkzeug {
	return []Werkzeug{{
		Name: "lesen.felder.vorschau", Art: Lesend, Idempotent: true,
		Erklaerung: "Zeigt, was »schreiben.felder« mit demselben Feld und Wert auf den " +
			"genannten Karten ändern würde — je Karte Name, Altwert und Neuwert, und " +
			"einen »ausschluss«, wo nichts zu tun ist (»gleich«: die Karte steht schon " +
			"so). Schreibt nichts. Die Zeilen ohne Ausschluss sind die Erwartung für " +
			"»schreiben.felder«: Kennung, »alt« als »wert« und »gabAlt« als »gabEs«.",
		Schema: objekt(map[string]any{
			"setID":    zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteIDs": zahlenliste("Die Kennungen der Karten aus »lesen.karten«.", "Eine Kartenkennung."),
			"feld":     text("Der Feldname, etwa »rarity« oder »artist«."),
			"wert":     text("Der Wert, der gesetzt würde."),
		}, "setID", "karteIDs", "feld", "wert"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl massenwahl
			if err := lies(argumente, &wahl); err != nil {
				return "", err
			}

			if err := wahl.pruefen(); err != nil {
				return "", err
			}

			vorschau, err := schreiber.FelderVorschau(
				ctx, wahl.SetID, wahl.KarteIDs, wahl.Feld, wahl.Wert)
			if err != nil {
				return "", err
			}

			return alsJSON(vorschau)
		},
	}}
}
