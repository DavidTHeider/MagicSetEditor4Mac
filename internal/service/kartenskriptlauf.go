package service

// **`new_card` — Karten aus einem Skript** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was nachgelesen wurde (§15)
//
// `src/script/functions/construction.cpp`, Zeilen 23–52 (Fassung f9d9356 vom
// 2021-07-06). Die Funktion baut eine Karte **aus dem Spiel** und setzt darauf
// die übergebenen Werte:
//
//	IndexMap<FieldP,ValueP>::const_iterator value_it = new_card->data.find(name);
//	if (value_it == new_card->data.end()) {
//	  throw ScriptError(format_string(_("Card doesn't have a field named '%s'"),name));
//	}
//
// **Ein unbekannter Feldname ist ein Fehler**, kein stilles Übergehen. Und
// gesetzt wird nur, was sich setzen lässt — Text, Wahl, Paketwahl, Farbe;
// alles andere gibt „Can not set value '%s', it is not of the right type".
//
// # Warum das hier steht und nicht in `internal/msescript`
//
// **§2**: „`internal/msescript` importiert nichts aus dem Projekt, auch nicht
// `domain`. Es ist ein Sprachwerkzeug; eine Kartenentität darin wäre eine
// Vermischung, die den Interpreter unprüfbar machte."
//
// `new_card` braucht die **Felder des Spiels**, um einen Namen abzuweisen, und
// gibt eine **Karte** zurück. Beides ist Projektwissen. Sie gehört damit zu
// den Funktionen, die `service` in die Umgebung gibt — wie `symbol_variation`
// und `exists_in_package`.
//
// # Und warum die Karte als Abbildung reist
//
// Weil der Interpreter keine Kartenentität kennen darf. Was `new_card`
// zurückgibt, ist eine **Abbildung von Feldnamen auf Werte** — Daten, keine
// Entität. [KartenAusSkript] macht daraus `domain.Card`, und das ist die
// Anti-Korruptionsschicht an ihrer richtigen Stelle.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// setzbar sind die Feldarten, die `new_card` beschreiben kann.
//
// # Nachgelesen, nicht abgeleitet
//
// `construction.cpp` Zeilen 39–48 (Fassung f9d9356 vom 2021-07-06) prüft vier
// Typen: `TextValue`, `ChoiceValue`, `PackageChoiceValue`, `ColorValue`. Wer
// bei dieser Aufzählung stehen bleibt, liest sie falsch — **die Hierarchie
// entscheidet mit**, und `dynamic_cast<ChoiceValue*>` trifft auch die
// Abkömmlinge:
//
//	class MultipleChoiceValue : public ChoiceValue   // multiple_choice.hpp:50
//	class BooleanValue        : public ChoiceValue   // boolean.hpp:47
//
// beide Fassung f9d9356 vom 2021-07-06. **`multiple choice` und `boolean` sind
// also setzbar**, und zwar über `toString()` wie jede andere Wahl.
//
// **Nicht setzbar bleiben Bild und Symbol**: `ImageValue` und `SymbolValue`
// leiten unmittelbar von `Value` ab und fallen bei MSE2 in den `else`-Zweig
// mit „Can not set value '%s', it is not of the right type". Beide tragen
// einen Dateinamen, den es beim Anlegen noch nicht gibt.
var setzbar = map[domain.FieldKind]bool{
	domain.KindText:           true,
	domain.KindChoice:         true,
	domain.KindMultipleChoice: true,
	domain.KindBoolean:        true,
	domain.KindPackageChoice:  true,
	domain.KindColor:          true,

	// **Eine unbekannte Art gilt als Text.** So kommt sie auch sonst durch
	// dieses Programm (`KindUnknown` steht als Textfeld da); sie hier
	// abzuweisen hieße, ein Skript an einer Feldart scheitern zu lassen, die
	// die Anwendung im Übrigen bedient.
	domain.KindUnknown: true,

	// **`info` fehlt**, und das mit Grund: ein Informationsfeld trägt keinen
	// Wert (`InfoField() { editable = false; }`, `information.hpp` Zeilen
	// 26–34, Fassung f9d9356 vom 2021-07-06).
}

// kartenmarke kennzeichnet, was `new_card` gebaut hat.
const kartenmarke = "kartenskript:new_card"

// Kartenskriptgrenze ist, wie viele Karten ein Kartenskript in einem Zug
// anlegen darf. **Zehntausend**, und das ist großzügig: das größte kanonische
// Set (`White.mse-set`) trägt 3.761 Karten.
const Kartenskriptgrenze = 10000

// KartenAusSkript wertet ein `add cards script:` aus und gibt seine Karten.
//
// **Die Felder kommen vom Aufrufer**, denn nur er kennt das Spiel des Sets.
// Ohne sie wäre jeder Feldname gültig, und der Fehlerweg des Altcodes fiele
// weg.
func KartenAusSkript(
	ctx context.Context, quelle string, felder []domain.Field, spielskript string,
) (karten []*domain.Card, befunde []domain.Befund, fehler error) {
	// **Über der Bibliothek und nicht daneben.** Eine nackte Umgebung kennt
	// `length`, `for` und die übrigen 48 Funktionen **nicht** — gemessen am
	// 2026-08-31: `length([1,2,3])` gibt dort „der Name ist hier nicht
	// bekannt". `magic.mse-game` benutzt sie in jedem seiner Zyklen.
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("new_card", kartenbauer(felder))

	// **Das Spielskript zuerst** (#216 für die Vorlagen, hier aus demselben
	// Grund). `magic.mse-game` erklärt darin `skeleton_runner` und
	// `skeleton_script`; ohne sie scheiterten **sieben der zwölf** Skripte mit
	// „der Name ist hier nicht bekannt" — gemessen am 2026-08-31.
	//
	// **Und ein kaputtes Spielskript hält nicht auf** (§8): es gibt einen
	// Befund, und was ohne es geht, geht weiter.
	befunde = vorbelegt(ctx, spielskript, "init script des Spiels", umgebung)

	wert, err := msescript.AuswertenQuelle(ctx, quelle, umgebung)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: das Kartenskript: %w", domain.ErrValidation, err)
	}

	// **Eine Liste, und nichts anderes.** `AddCardsScript::perform` macht aus
	// dem Ergebnis einen Iterator; ein einzelner Wert wäre dort eine Liste aus
	// sich selbst, und ein `"Text"` gäbe seine Zeichen. Hier ist es ein
	// Fehler: ein Skript, das keine Karten gibt, hat keine gemeint.
	if wert.Art != msescript.ArtAbbildung {
		return nil, nil, fmt.Errorf(
			"%w: das Kartenskript gab keine Liste von Karten, sondern %s",
			domain.ErrValidation, wert.Art)
	}

	if befunde == nil {
		befunde = []domain.Befund{}
	}

	// **Eine Grenze, bevor angelegt wird** (QA-06c, #1976). MSE2 kennt keine;
	// hier geht ein Zug in den Bestand, und ein Skript mit einer Million
	// Karten wäre ein Bestand, den niemand mehr öffnet.
	if len(wert.Eintraege) > Kartenskriptgrenze {
		return nil, nil, fmt.Errorf("%w: das Kartenskript gab %d Karten — mehr als %d legt ein Zug nicht an",
			domain.ErrValidation, len(wert.Eintraege), Kartenskriptgrenze)
	}

	karten = make([]*domain.Card, 0, len(wert.Eintraege))

	for stelle, eintrag := range wert.Eintraege {
		karte, kartenfehler := alsKarte(eintrag.Wert, stelle)
		if kartenfehler != nil {
			return nil, nil, kartenfehler
		}

		karten = append(karten, karte)
	}

	return karten, befunde, nil
}

// alsKarte macht aus einer Abbildung eine Karte.
func alsKarte(wert msescript.Wert, stelle int) (*domain.Card, error) {
	// **Nur, was `new_card` gebaut hat** (QA-06c, #1976). MSE2 wandelt jedes
	// Glied mit `from_script<CardP>` und weist alles andere ab
	// (`AddCardsScript::perform`, `src/data/add_cards_script.cpp`, Zeile 33,
	// Fassung f9d9356 vom 2021-07-06). Eine bloße Abbildung ginge an der
	// Feldprüfung von `new_card` vorbei — ein Bildfeld, ein Feld, das es
	// nicht gibt. Die Marke kann ein Skript weder setzen noch fälschen
	// ([msescript.Wert.Marke]).
	if wert.Art == msescript.ArtAbbildung && wert.Marke != kartenmarke {
		return nil, fmt.Errorf(
			"%w: die %d. Stelle der Kartenliste kommt nicht aus new_card",
			domain.ErrValidation, stelle+1)
	}

	if wert.Art != msescript.ArtAbbildung {
		return nil, fmt.Errorf(
			"%w: die %d. Stelle der Kartenliste ist keine Karte, sondern %s",
			domain.ErrValidation, stelle+1, wert.Art)
	}

	karte := &domain.Card{Values: make([]domain.Value, 0, len(wert.Eintraege))}

	for _, eintrag := range wert.Eintraege {
		// **Der Feldname in der Dateischreibweise.** Das Skript schreibt
		// `casting_cost`, die `.mse-set` trägt `casting cost` — wer den
		// Skriptnamen ablegte, hätte eine Karte mit einem Feld, das die
		// Vorlage nicht kennt.
		karte.Values = append(karte.Values, domain.Value{
			Field: msepaket.Entkanonisch(eintrag.Name),
			Text:  eintrag.Wert.AlsText(),
		})
	}

	return karte, nil
}

// kartenbauer gibt die Funktion `new_card`.
func kartenbauer(felder []domain.Field) msescript.Wert {
	// **Unter beiden Schreibweisen.** Ein Feld heißt in der Datei `casting
	// cost`, im Skript `casting_cost` — MSE macht aus Unterstrichen Leerzeichen
	// und zurück (`msepaket.Entkanonisch`). Gemessen am 2026-08-31: **null von
	// zwölf** echten Skripten liefen ohne diese Zeile; sie scheiterten an
	// `super_type` und `casting_cost`.
	arten := make(map[string]domain.FieldKind, 2*len(felder))
	for _, feld := range felder {
		arten[feld.Name] = feld.Kind
		arten[msepaket.Kanonisch(feld.Name)] = feld.Kind
	}

	return msescript.Wert{
		Art: msescript.ArtFunktion,
		Funktion: &msescript.Funktionsteil{
			Eingebaut: func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
				return neueKarte(ruf, arten)
			},
		},
	}
}

// neueKarte baut die Abbildung, die eine Karte beschreibt.
func neueKarte(ruf msescript.Ruf, arten map[string]domain.FieldKind) (
	msescript.Wert, error,
) {
	// **Die Eingabe steht in der Stellung, nicht benannt.** `new_card([…])`
	// übergibt eine Abbildung als einziges Argument — so schreibt es
	// `magic.mse-game`, und so nimmt MSE2 es entgegen (`SCRIPT_PARAM(
	// ScriptValueP, input)`).
	//
	// **Ohne Argument eine leere Karte.** `#Many Empty Cards` in
	// `add_cards_scripts` ruft `new_card()` — auskommentiert, aber gemeint.
	if len(ruf.Stellung) == 0 {
		return msescript.Wert{Art: msescript.ArtAbbildung, Marke: kartenmarke}, nil
	}

	eingabe := ruf.Stellung[0]
	if eingabe.Art != msescript.ArtAbbildung {
		return msescript.Nichts, ruf.FehlerAn(
			"new_card nimmt eine Abbildung von Feldnamen auf Werte")
	}

	gebaut := msescript.Wert{
		Art:       msescript.ArtAbbildung,
		Eintraege: make([]msescript.Eintrag, 0, len(eingabe.Eintraege)),
		Marke:     kartenmarke,
	}

	for _, eintrag := range eingabe.Eintraege {
		art, bekannt := arten[eintrag.Name]
		if !bekannt {
			return msescript.Nichts, ruf.FehlerAn(fmt.Sprintf(
				"eine Karte hat kein Feld namens %q", eintrag.Name))
		}

		if !setzbar[art] {
			return msescript.Nichts, ruf.FehlerAn(fmt.Sprintf(
				"das Feld %q lässt sich hier nicht setzen — es ist ein %s",
				eintrag.Name, art))
		}

		gebaut.Eintraege = append(gebaut.Eintraege, eintrag)
	}

	return gebaut, nil
}
