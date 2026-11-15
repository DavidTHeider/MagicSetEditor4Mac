package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// **Der Lebenszyklus einer Karte über MCP** —
// #591.
//
// # Warum es diese Werkzeuge gibt
//
// `AGENTS.md` §2 sagt, wofür es dieses Paket gibt: „Werkzeuge, mit denen ein
// Sprachmodell Karten **liest und ändert**". Bis zum 2026-08-22 gab es fünf
// Werkzeuge, davon **eines** schreibendes — `schreiben.feld`. Ein Sprachmodell
// konnte ein Feld setzen und sonst nichts ändern.
//
// **Das war der Befund, an dem
// [0039](../../docs/entscheidungen/0039-kein-dienst-auf-der-rueckschleife.md)
// hing**: die Lücke lag im **Umfang** und nicht im Protokoll, und deshalb hätte
// ein zweiter Zugang sie nicht geschlossen.
//
// # Wie man sie zurücknimmt
//
// [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md) verlangt,
// dass eine Änderung sagt, wie sie sich zurücknehmen lässt. Bei einem Feld ist
// das der Vorwert; hier ist es der **Verlauf**: jede dieser Änderungen schreibt
// einen Kartenverlaufsschritt, und der Schreiber des Servers trägt
// `domain.UrheberServer`. Die Antwort sagt es, statt es vorauszusetzen — wer
// über MCP löscht, holt es mit ⌘Z im Fenster zurück, samt derselben Kennung
// (#621).
//
// **Zwei der vier nennen zusätzlich den unmittelbaren Rückweg**: eine neu
// angelegte oder kopierte Karte lässt sich ohne Verlauf wieder entfernen, und
// das ist die kürzere Antwort auf „wie mache ich das rückgängig".

// Kartenschreiber ist der Lebenszyklus, den die Werkzeuge hier brauchen.
//
// **Eine eigene Schnittstelle neben [Schreiber]**, weil sie eine andere Frage
// beantwortet: dort geht es um ein Feld einer Karte, hier um die Karte selbst.
// Wer nur Felder schreiben lassen will, reicht nur [Schreiber] herein.
//
// **Jeder Vorgang nennt das Set** — seit
// #801. Bis
// dahin nahm er eine blanke Kartenkennung, und eine verwechselte löschte eine
// Karte in einem Set, das niemand geöffnet hatte. Geprüft wird die
// Zugehörigkeit in `service`, nicht hier
// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
type Kartenschreiber interface {
	KarteAnlegen(ctx context.Context, setID int64) (int64, error)
	KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error)
	KarteLoeschen(ctx context.Context, setID, karteID int64, trotzdem bool) error
	KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error
}

// pruefeKennungen prüft Set und Karte in einem Zug.
//
// **In dieser Reihenfolge**: wer eine unbrauchbare Setkennung schickt, soll das
// erfahren, bevor er über die Karte nachdenkt.
func pruefeKennungen(setID, karteID int64) error {
	if err := pruefeKennung("setID", setID); err != nil {
		return err
	}

	return pruefeKennung("karteID", karteID)
}

// lebenszyklus gibt die Werkzeuge, die eine Karte anlegen, kopieren, löschen
// oder umordnen.
// **Keines dieser vier Werkzeuge sagt Idempotenz zu** (#895), und die vier
// Gründe sind vier verschiedene:
//
//   - `anlegen` und `kopieren` geben bei jedem Ruf eine **neue Kennung**.
//   - `loeschen` glückt einmal; der zweite Ruf findet die Karte nicht mehr.
//   - `tauschen` stellt beim zweiten Ruf die **alte** Reihenfolge wieder her —
//     es ist seine eigene Umkehrung, und das ist das Gegenteil von idempotent.
//
// Bis zum 2026-08-23 sagten alle vier sie zu, weil der Hinweis an der **Art**
// hing statt am Werkzeug. `idempotenz_test.go` führt die Zusage seither aus.
func lebenszyklus(schreiber Kartenschreiber) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.karte.anlegen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt eine leere Karte im Set an und gibt ihre Kennung zurück.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			}, "setID"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				var wahl struct {
					SetID int64 `json:"setID"`
				}

				if err := lies(argumente, &wahl); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", wahl.SetID); err != nil {
					return "", err
				}

				neue, err := schreiber.KarteAnlegen(ctx, wahl.SetID)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"setID":   wahl.SetID,
					"karteID": neue,
					"zuruecknahme": rueckweg(
						"schreiben.karte.loeschen mit setID %d und karteID %d",
						wahl.SetID, neue),
				})
			},
		},
		{
			Name: "schreiben.karte.kopieren", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt eine Kopie der Karte an und gibt die Kennung der Kopie zurück.",
			Schema: objekt(map[string]any{
				"setID":   zahl("Die Kennung des Sets, in dem die Karte liegt."),
				"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			}, "setID", "karteID"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				var wahl struct {
					SetID   int64 `json:"setID"`
					KarteID int64 `json:"karteID"`
				}

				if err := lies(argumente, &wahl); err != nil {
					return "", err
				}

				if err := pruefeKennungen(wahl.SetID, wahl.KarteID); err != nil {
					return "", err
				}

				kopie, err := schreiber.KarteKopieren(ctx, wahl.SetID, wahl.KarteID)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"vorlageID": wahl.KarteID,
					"karteID":   kopie,
					"zuruecknahme": rueckweg(
						"schreiben.karte.loeschen mit setID %d und karteID %d",
						wahl.SetID, kopie),
				})
			},
		},
		{
			Name: "schreiben.karte.loeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: nimmt die Karte aus dem Set. " +
				"Rückgängig nur über den Verlauf des Fensters — die Karte kommt mit derselben Kennung zurück. " +
				"Liegt die Karte in einem Deck, wird der Ruf abgewiesen und nennt die Decks; " +
				"mit trotzdem=true wird sie auch dort herausgenommen.",
			Schema: objekt(map[string]any{
				"setID":   zahl("Die Kennung des Sets, in dem die Karte liegt."),
				"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
				"trotzdem": jaNein("Bestätigt ein Löschen, das einen Deckplatz mitnimmt. " +
					"Ohne diese Angabe wird eine Karte, die in einem Deck liegt, abgewiesen."),
			}, "setID", "karteID"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				var wahl struct {
					SetID    int64 `json:"setID"`
					KarteID  int64 `json:"karteID"`
					Trotzdem bool  `json:"trotzdem"`
				}

				if err := lies(argumente, &wahl); err != nil {
					return "", err
				}

				if err := pruefeKennungen(wahl.SetID, wahl.KarteID); err != nil {
					return "", err
				}

				if err := schreiber.KarteLoeschen(
					ctx, wahl.SetID, wahl.KarteID, wahl.Trotzdem); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"setID":        wahl.SetID,
					"karteID":      wahl.KarteID,
					"geloescht":    true,
					"zuruecknahme": ueberDenVerlauf,
				})
			},
		},
		{
			Name: "schreiben.karten.tauschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: tauscht die Stellen zweier Karten im Set.",
			Schema: objekt(map[string]any{
				"setID":    zahl("Die Kennung des Sets, in dem beide Karten liegen."),
				"eineID":   zahl("Die Kennung der einen Karte aus »lesen.karten«."),
				"andereID": zahl("Die Kennung der anderen Karte. Muss eine andere sein."),
			}, "setID", "eineID", "andereID"),
			Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
				var wahl struct {
					SetID    int64 `json:"setID"`
					EineID   int64 `json:"eineID"`
					AndereID int64 `json:"andereID"`
				}

				if err := lies(argumente, &wahl); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", wahl.SetID); err != nil {
					return "", err
				}

				if err := pruefeKennung("eineID", wahl.EineID); err != nil {
					return "", err
				}

				if err := pruefeKennung("andereID", wahl.AndereID); err != nil {
					return "", err
				}

				// **Eine Karte mit sich selbst zu tauschen ist keine
				// Änderung, sondern ein Tippfehler.** Der Bestand käme damit
				// zurecht und schriebe einen Verlaufsschritt, der nichts tut —
				// und der Aufrufer glaubte, er habe umgeordnet.
				if wahl.EineID == wahl.AndereID {
					return "", fmt.Errorf(
						"andereID: dieselbe Karte wie eineID (%d) — das tauscht nichts",
						wahl.EineID)
				}

				if err := schreiber.KartenTauschen(ctx, wahl.SetID, wahl.EineID,
					wahl.AndereID); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"eineID":       wahl.EineID,
					"andereID":     wahl.AndereID,
					"zuruecknahme": rueckweg("schreiben.karten.tauschen mit denselben Kennungen"),
				})
			},
		},
	}
}

// ueberDenVerlauf ist der Rückweg, den nur das Fenster gehen kann.
const ueberDenVerlauf = "über den Verlauf des Fensters (⌘Z) — die Karte kommt " +
	"mit derselben Kennung zurück"

// lies zerlegt die Argumente und gibt einen Fehler, der sagt, was fehlte.
func lies(argumente json.RawMessage, ziel any) error {
	if err := json.Unmarshal(argumente, ziel); err != nil {
		return fmt.Errorf("die Argumente sind kein JSON: %w", err)
	}

	return nil
}

// pruefeKennung weist eine Kennung ab, die keine ist — **und nennt das Feld** (§8).
//
// **Der Name gehört in die Meldung.** Ein Werkzeug mit zwei Kennungen hat sonst
// zwei gleich lautende Fehler, und der Aufrufer sucht in beiden Feldern.
//
// **Sie ist die eine Stelle**, an der "ist das eine Kennung" beantwortet wird:
// [kennung] ruft sie für die lesenden Werkzeuge, die schreibenden rufen sie
// unmittelbar. Zwei Rechnungen über dieselbe Sache liefen früher oder später
// auseinander.
func pruefeKennung(feld string, wert int64) error {
	if wert <= 0 {
		return fmt.Errorf("%s: %d ist keine Kennung — sie kommt aus einem »lesen«-Werkzeug",
			feld, wert)
	}

	return nil
}

// rueckweg formt den Hinweis, wie sich eine Änderung zurücknehmen lässt.
func rueckweg(muster string, teile ...any) string {
	return fmt.Sprintf(muster, teile...)
}
