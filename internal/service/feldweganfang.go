package service

// **Eine neue Karte rechnet ihre Werte** — #2516 (Rest aus #2396).
//
// MSE rechnet beim Anlegen jeden Wert der neuen Karte, in der Reihenfolge der
// Kartenfelder und je einmal (`AddCardAction` in `SetScriptManager::onAction`,
// `src/script/script_manager.cpp`, Zeilen 168–178, Fassung f9d9356 vom
// 2021-07-06: `FOR_EACH(v, card->data) v->update(ctx)`) — für eine angelegte
// Karte wie für die Karten eines Kartenskripts. Dasselbe tut hier der Nachzug
// mit allen Feldern, die ein Wertskript tragen: jedes rechnet einmal, in der
// Reihenfolge des Spiels, und wer schon gerechnet hat, rechnet nicht wieder
// ([nachziehen]).
//
// # Was anders ist (adapted)
//
// **Ein leeres Ergebnis wird nicht gespeichert.** MSE hält zu jedem Feld einen
// Wert, auch einen leeren; der Bestand hier führt nur Zeilen, die etwas sagen,
// und ein leeres Feld zeichnet wie ein fehlendes. **Und die übrigen Karten
// rechnen nicht nach**: MSE ruft danach `updateAllDependend` über die
// Kartenliste (Zeilen 180–188); was davon abhängt, rechnet hier das Zeichnen
// (Vermerk 0203).

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Anfangswerte gibt die Werte, mit denen eine neue Karte dieses Sets in den
// Bestand geht — `werte` sind, was sie schon trägt (ein Kartenskript gibt
// welche), gerechnet durch ihre Skripte, dahinter, was die übrigen Skripte
// liefern. Ohne Nachzug oder Vorlage bleibt, was gegeben ist.
func (f *Feldweg) Anfangswerte(ctx context.Context, setID int64, werte []domain.Value) []domain.Value {
	return anfangswerte(ctx, f.nachzug, setID, werte)
}

func anfangswerte(ctx context.Context, n *feldnachzug, setID int64, werte []domain.Value) []domain.Value {
	if n == nil {
		return werte
	}

	karte := &domain.CardRow{SetID: setID, Values: werte}

	vorlage, _, gefunden := n.stile.StilFuerKarte(ctx, karte)
	if !gefunden {
		return werte
	}

	// **Jedes Feld hat einen Wert, bevor gerechnet wird** — in MSE ein
	// `TextValue` im Vorgabezustand (`src/data/field/text.hpp`, Zeile 109,
	// Fassung f9d9356 vom 2021-07-06),
	// erst `default_script`, dann `script` (`TextValue::update`,
	// `src/data/field/text.cpp`, Zeilen 173–183). Hier legt [mitVorgaben]
	// den Eintrag jedes fehlenden Felds mit seiner Vorgabe an; eine eigene
	// Ergänzung mit leeren Werten überdeckte sie (gemessen: ohne sie rechnet
	// die neue Karte dasselbe).

	var namen []string

	for _, feld := range vorlage.Kartenfelder {
		if strings.TrimSpace(feld.Skript) != "" {
			namen = append(namen, eintragsname(feld.Name))
		}
	}

	eintraege, gerechnet := n.rechne(ctx, vorlage, karte, namen)

	return angelegte(eintraege, gerechnet, vorlage.Kartenfelder, werte)
}

// angelegte setzt die gegebenen Werte auf ihr Gerechnetes und hängt an, was
// die übrigen Skripte liefern — mit `save value` und nicht leer.
func angelegte(
	eintraege []msescript.Eintrag, gerechnet []string, felder []domain.Field, werte []domain.Value,
) []domain.Value {
	text := make(map[string]string, len(eintraege))
	for _, eintrag := range eintraege {
		text[eintrag.Name] = eintrag.Wert.AlsText()
	}

	gerechnete := make(map[string]bool, len(gerechnet))
	for _, name := range gerechnet {
		gerechnete[name] = true
	}

	ergebnis := make([]domain.Value, 0, len(werte)+len(gerechnet))
	gegeben := make(map[string]bool, len(werte))

	for _, wert := range werte {
		if name := eintragsname(wert.Field); gerechnete[name] {
			wert.Text = text[name]
		}

		gegeben[wert.Field] = true
		ergebnis = append(ergebnis, wert)
	}

	for _, feld := range felder {
		name := eintragsname(feld.Name)
		if !gegeben[feld.Name] && gerechnete[name] && feld.Gespeichert && text[name] != "" {
			ergebnis = append(ergebnis, domain.Value{Field: feld.Name, Text: text[name]})
		}
	}

	return ergebnis
}

// eintragsname ist der Name eines Felds, wie die Einträge ihn führen.
func eintragsname(feld string) string { return strings.ReplaceAll(feld, " ", "_") }
