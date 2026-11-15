// Die Vorgabe einer Stiloption: `default:`, `initial:` und die erste Wahl.
//
// **Eine eigene Datei**, weil `stiloptionen.go` sonst über die 400 Zeilen aus
// §2 liefe. Der Schnitt folgt einer Frage: dort steht, wie eine Option gelesen
// und geschrieben wird — hier, welchen Wert sie hat, wenn niemand sie angefasst
// hat.
//
// Getrennt am 2026-08-22 mit
// #789.

package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// stiloptionsvorgabe gibt den Wert, den eine Option ohne eigene Angabe hat.
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	         :                           _("") , true)
//
// `src/data/field/choice.cpp`, Zeilen 279–284; der Schalter steht in
// `choice.hpp:192` auf **`true`**. Fassung f9d9356 vom 2021-07-06.
//
// # Dieselbe Regel wie beim Kartenfeld, andere Stelle
//
// Für Kartenfelder steht sie seit
// #341 in
// [vorgabewert]; für Stiloptionen fehlte sie. `magic-archenemy` setzt die
// Schriftfarbe seines Fußes über `{styling.artist_font_color}` — ein
// `choice`-Feld mit `white` und `black` und **ohne** `initial:`. Bei uns war
// der Wert leer und die Farbe fiel auf Schwarz zurück; MSE2 schreibt dort weiß
// (#409).
//
// Über die 376 Stile des vollen Packs nennen **194** `styling field:` der Art
// `choice` kein `initial:` — gezählt am Text der Stildateien (§10).
func stiloptionsvorgabe(ctx context.Context, feld domain.StyleField) string {
	return stiloptionsvorgabeMit(ctx, feld, msescript.Nichts)
}

// stiloptionsanfang ist der Wert **ohne** `default:` — `initial:` oder die
// erste Wahl.
//
// **Er ist die erste Hälfte von MSEs zwei Schritten** (#789): der Konstruktor
// setzt `initial` und markiert den Wert als Vorgabe
// (`src/data/field/color.cpp`, Zeilen 82–101, Fassung f9d9356 vom
// 2021-07-06), und erst `update` ruft das `default:`-Skript
// (`invokeOnDefault`, `src/script/scriptable.hpp`, Zeilen 71–78, dieselbe
// Fassung).
//
// **Damit löst sich der Kreis**: ein `default:`, das `styling.andere_option`
// liest, sieht deren Anfangswert — nicht das Ergebnis ihres eigenen
// `default:`. Genau so verhält sich MSE beim ersten Durchlauf.
func stiloptionsanfang(feld domain.StyleField) string {
	if anfang := strings.TrimSpace(feld.Initial); anfang != "" {
		return anfang
	}

	if (feld.Kind == domain.KindChoice || feld.Kind == domain.KindBoolean) &&
		len(feld.Choices) > 0 {
		return feld.Choices[0]
	}

	return ""
}

// stiloptionsvorgabeMit ist [stiloptionsvorgabe] mit einer Umgebung für das
// `default:`-Skript.
func stiloptionsvorgabeMit(
	ctx context.Context, feld domain.StyleField, styling msescript.Wert,
) string {
	// **`default:` schlägt `initial:`** — und das ist nachgelesen, nicht aus
	// der Reihenfolge der Schlüssel erschlossen
	// (#590):
	//
	//	ColorValue::ColorValue(const ColorFieldP& field)
	//	  : Value(field)
	//	  , value( !field->initial.isDefault() ? field->initial()
	//	         : !field->choices.empty()     ? field->choices[0]->color
	//	         :                               Color()
	//	         , true)
	//	{}
	//
	//	bool ColorValue::update(Context& ctx) {
	//	  bool change = field().default_script.invokeOnDefault(ctx, value)
	//
	// `src/data/field/color.cpp`, Zeilen 82–101, Fassung f9d9356 vom
	// 2021-07-06. **Das zweite Argument von `Defaultable` ist `true`**: der
	// Konstruktor setzt `initial` und markiert den Wert zugleich als Vorgabe.
	// `invokeOnDefault` läuft genau dann (`src/script/scriptable.hpp`, Zeilen
	// 71–78) und überschreibt ihn — solange niemand die Option angefasst hat.
	if wert, ist := vorgabeAusAusdruck(ctx, feld.Default, styling); ist {
		return wert
	}

	if anfang := strings.TrimSpace(feld.Initial); anfang != "" {
		return anfang
	}

	// **Bei einer Auswahl und bei einem Wahrheitswert.** Ein Text- oder
	// Zahlenfeld hat keine Wahlen, und ein leerer Wert ist dort die richtige
	// Antwort.
	//
	// **Auch bei `KindBoolean`.** Ein `boolean` ist bei MSE ein Auswahlfeld mit
	// den festen Wahlen `yes` und `no`
	// (`src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom
	// 2021-07-06) — und `BooleanValue` ruft `ChoiceValue(field)` **ohne**
	// zweites Argument (`boolean.hpp`, Zeile 49), womit
	// `initial_first_choice` auf `true` steht (`choice.hpp`, Zeile 192).
	// Ohne `initial:` gilt damit `yes`.
	//
	// Über das Pack sind **378** von 1.489 boolean-Stiloptionen ohne
	// `initial:`, in **178 von 365** Stilpaketen
	// (#531).
	if (feld.Kind == domain.KindChoice || feld.Kind == domain.KindBoolean) &&
		len(feld.Choices) > 0 {
		return feld.Choices[0]
	}

	return ""
}

// gespeicherteForm bringt einen Wert in die Form, in der er im Bestand steht.
//
// **Nur eine Mehrfachwahl wird angefasst.** Ein Text mit Komma ist ein Text;
// ihn zu ordnen hieße, eine Eingabe umzuschreiben, die niemand darum gebeten
// hat.

// vorgabeAusAusdruck wertet `default:` aus und sagt, ob dabei etwas herauskam.
//
// # `default:` ist ein Skript, keine Zeichenkette
//
// `OptionalScript default_script` (`src/data/field/color.hpp`, Zeile 33,
// Fassung f9d9356 vom 2021-07-06). Am Orakel nachgefragt
// (`magicseteditor --cli … -q --raw`) — und unser Auswerter gibt dasselbe:
//
//	"black"      ->  "black"
//	rgb(0,0,0)   ->  rgb(0,0,0)
//	black        ->  ERROR: Variable not set: black
//
// **Ein Fehler ist kein Abbruch, sondern ein Rückfall.** MSE fängt Skriptfehler
// im Wertumlauf ab und lässt den Wert stehen; hier heißt das: es gilt weiter,
// was `initial:` sagt. Genau dieser Fall steht im Pack — `magic-m15-bigtext`
// schreibt `default: black` **ohne** Anführungszeichen und daneben
// `initial: black`, das ihn auffängt.
//
// # Die Reichweite, und was sie auslässt
//
// **Ausgewertet wird gegen die Standardbibliothek, ohne Karte und ohne Set.**
// Über die 376 Stilpakete von `full-magic-pack` tragen **zehn** von **105**
// Farboptionen ein `default:`, und alle zehn sind Konstanten: achtmal
// `"black"`, einmal `rgb(0,0,0)`, einmal das nackte `black`.
//
// **Käme eine Vorlage mit `default: card.name`**, liefe die Auswertung in einen
// Fehler und es gälte `initial:`. Das ist der benannte Preis dieser Umgebung —
// keine stille Lücke, sondern eine Grenze, die hier steht, weil im Pack kein
// Fall dahinterliegt.
// # Womit ausgewertet wird
//
// **Mit `styling`** — seit
// #789.
// MSE legt in die Umgebung eines `default:`-Skripts `set`, `game`,
// `stylesheet`, `card_style` und `styling`
// (`src/script/script_manager.cpp`, Zeilen 30–46, Fassung f9d9356 vom
// 2021-07-06).
//
// **`card` steht dabei auf `nil`, und das ist keine Lücke** — es ist, was MSE
// tut: die setweite Stilseite ruft `getContext(CardP())`, und der Zweig setzt
// `card` ausdrücklich leer (`script_manager.cpp`, Zeilen 64–74). Das Issue
// nannte „welche Karte den Kontext stellt" die eigentliche Entwurfsfrage; die
// Antwort steht im Quelltext, und sie lautet: **gar keine**.
//
// **`set`, `game` und `stylesheet` fehlen hier noch.** Sie kosten eine neue
// Naht: `Stilwaehler` führt `koepfe`, `katalog`, `ablage` und `felder` — keine
// davon gibt Setwerte her. Gemessen über beide Paketstände braucht es heute
// kein Fall: von 105 Farboptionen tragen zehn ein `default:`, und alle zehn
// sind Konstanten. Wer den ersten echten Fall findet, hat damit das Werkstück.
func vorgabeAusAusdruck(
	ctx context.Context, quelle string, styling msescript.Wert,
) (string, bool) {
	if strings.TrimSpace(quelle) == "" {
		return "", false
	}

	umgebung := msescript.Bibliothek()
	if styling.Art != msescript.ArtNichts {
		umgebung.Setze("styling", styling)
	}

	// **Ein `{ … }` ist der Rumpf, nicht eine Funktion** (#789).
	//
	// Bei MSE ist `default:` ein `Scriptable<Color>`; steht dort ein Block,
	// **ruft** `invokeOnDefault` ihn (`src/script/scriptable.hpp`, Zeilen
	// 71–78, Fassung f9d9356 vom 2021-07-06). Hier stand nur
	// `AuswertenQuelle`, und das gibt für einen Block
	// eine **Funktion** zurück — die der Zweig unten dann als „keine Vorgabe"
	// abwies. Damit war jedes gerechnete `default:` still wirkungslos; dass es
	// niemandem auffiel, lag daran, dass alle zehn im Pack Konstanten sind.
	rumpf, geklammert := skriptrumpf(quelle)
	if !geklammert {
		rumpf = quelle
	}

	wert, err := msescript.AuswertenQuelle(ctx, rumpf, umgebung)
	if err != nil {
		return "", false
	}

	// **Angenommen werden genau die beiden Formen, die im Pack stehen** — eine
	// Zeichenkette und eine Farbe. Das ist enger als „alles außer nichts", und
	// die Enge ist Absicht: `nil` wertet fehlerfrei zu *nichts* aus, und ein
	// `{ … }` ergibt eine **Funktion**. Beides als Vorgabe zu nehmen hieße,
	// aus einer leeren oder unaufgerufenen Angabe einen Wert zu machen.
	//
	// **Die Textform ist die des Auswerters und keine zweite daneben.** Eine
	// Farbe wird dabei `rgba(r,g,b,a)`, und `farbeAus` liest genau das — zwei
	// Schreibweisen für dieselbe Farbe liefen früher oder später auseinander
	// (`farbdeutung.go`).
	switch wert.Art {
	case msescript.ArtTextwert, msescript.ArtFarbe:
		return wert.AlsText(), true
	}

	return "", false
}

// OptionZuruecksetzen entfernt den gespeicherten Wert einer Stiloption.
//
//	void DropDownColorList::select(size_t item) {
//	  if (isDefault(item)) {
//	    cve.change( Defaultable<Color>());
//
// `src/gui/value/color.cpp`, Zeilen 118–120, Fassung f9d9356 vom 2021-07-06:
// die Wahl des Vorgabeeintrags setzt einen **leeren** `Defaultable` ein. Bei
// uns heißt das, die Zeile aus dem Bestand zu nehmen — ab dann rechnet
// [stiloptionsvorgabe] den Wert wieder aus der Vorlage
// (#764).
//
// **Hier wird geräumt, nicht entschieden.** Ob eine Option den Eintrag
// überhaupt anbietet, hängt bei MSE an `hasDefault()` und damit am
// `default_script` (`color.cpp`, Zeile 38) — das ist eine Frage der
// Oberfläche, und sie wird dort beantwortet.
//
// **Eine Option, die der Stil nicht kennt, wird abgewiesen** — dieselbe
// Prüfung wie in [Stilwaehler.OptionSetzen]. Ohne sie räumte ein Tippfehler
// still nichts und meldete Erfolg.

// anfangsstyling baut `styling` aus den Anfangswerten aller Optionen.
//
// **Ohne ihre `default:`-Skripte**, denn genau die werden damit ausgewertet.
// Bei MSE ist das dieselbe Reihenfolge: der Konstruktor setzt `initial`, und
// `update` ruft danach die Skripte (#789).
//
// **Was das nicht leistet**: eine Kette. Liest ein `default:` eine Option, die
// ihrerseits ein `default:` hat, sieht es deren **Anfangswert** und nicht das
// Ergebnis. MSE löst das über eine Abhängigkeitsordnung
// (`script_manager.cpp`); hier steht ein Durchlauf, und über beide
// Paketstände gibt es keinen Fall, der mehr bräuchte — von 105 Farboptionen
// tragen zehn ein `default:`, und alle zehn sind Konstanten.
func anfangsstyling(optionen []domain.StyleField) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(optionen))

	for _, option := range optionen {
		eintraege = append(eintraege, msescript.Eintrag{
			Name: strings.ReplaceAll(option.Name, " ", "_"),
			Wert: msescript.Text(stiloptionsanfang(option)),
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// mitGespeicherten legt die gespeicherten Werte über die Anfangswerte.
//
// # Warum das nötig ist
//
// **MSE bindet die tatsächlichen Stilwerte**, nicht die gerechneten:
//
//	ctx.setVariable(SCRIPT_VAR_styling, to_script(&set.stylingDataFor(*stylesheet)));
//
// `src/script/script_manager.cpp`, Zeilen 30–61, und `src/data/set.cpp`,
// Zeilen 99–105, Fassung f9d9356 vom 2021-07-06.
//
// Bis zum 2026-08-22 baute [Stilwaehler.Optionen] die Umgebung **nur** aus
// [anfangsstyling] und legte die gespeicherten Werte erst danach auf das
// Ausgabe-DTO. Ein `default:`, das eine Nachbaroption liest, sah damit immer
// den Anfangswert — bei `hintergrund: dunkel` weiterhin Schwarz. Angestrichen
// vom Nachlauf zum Zweitblick über Phase P
// (#860).
//
// **Die Anfangswerte bleiben darunter**: eine Option, für die nichts
// gespeichert ist, hat trotzdem einen Wert — das ist der Grund für #789, und
// er gilt weiter.
func mitGespeicherten(styling msescript.Wert, werte []domain.Value) msescript.Wert {
	if len(werte) == 0 {
		return styling
	}

	eintraege := make([]msescript.Eintrag, len(styling.Eintraege))
	copy(eintraege, styling.Eintraege)

	for _, wert := range werte {
		name := strings.ReplaceAll(wert.Field, " ", "_")

		gefunden := false

		for stelle := range eintraege {
			if eintraege[stelle].Name == name {
				eintraege[stelle].Wert = msescript.Text(wert.Text)
				gefunden = true

				break
			}
		}

		// **Auch ein Wert ohne Feld kommt hinein.** Ein Set kann einen Wert
		// tragen, den der heutige Stil nicht mehr führt; MSEs `styling_data`
		// hält ihn ebenso, und ein Skript, das ihn liest, soll ihn sehen.
		if !gefunden {
			eintraege = append(eintraege, msescript.Eintrag{
				Name: name, Wert: msescript.Text(wert.Text),
			})
		}
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}
