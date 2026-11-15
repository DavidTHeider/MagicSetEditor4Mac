package service

import (
	"context"
	"errors"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Vorgabewerte der Kartenfelder eines Spiels.
//
// # Der Befund
//
// `magic-m15.mse-style` legt den Kartennamen so an:
//
//	name:
//	    left: { if card.card_symbol == "none" then 32 else 50 }
//
// Eine Karte ohne Symbol führt kein Feld `card symbol`. Damit war
// `card.card_symbol` gleich `nil`, `nil == "none"` falsch, und der Name begann
// bei **50** statt bei **32** — verglichen an einer echten MSE-Ausgabe
// derselben Auflösung
// (#276).
//
// **Das ist keine Feinheit an einem Feld.** Jede Stelle, die `card.<feld>` gegen
// eine Zeichenkette prüft, entscheidet ohne Vorgabewerte falsch herum, und
// `magic.mse-game` führt 249 Kartenfelder.

// vorgabewert gibt, was ein Feld ohne Wert der Karte trägt.
//
// Die Reihenfolge ist die Aussage:
//
//  1. `default:`, wenn es dasteht — 142 der 249 Felder von `magic.mse-game`
//     haben eines, und **alle davon sind Skript**,
//  2. sonst `initial:`, der Anfangswert eines Auswahlfelds — 17-mal im Pack,
//     darunter `rarity: initial: common`,
//  3. sonst die **erste** `choice:` eines Auswahlfelds — `card symbol` bekommt
//     so `none`, und genau danach fragt der Stil,
//  4. sonst der leere Text.
//
// **Ohne Stufe 2** bekäme `rarity` `basic land` statt `common` — auf jeder
// Karte, die keine Seltenheit führt
// (#341).
// MSE ordnet die Stufen genauso (`src/data/field/choice.cpp`, Zeilen 279–283,
// Fassung f9d9356 vom 2021-07-06).
//
// **Ein Vorgabewert darf ein Skript sein**, und ein Skript, das scheitert, wird
// zum leeren Text: ein Spielpaket mit 249 Feldern, von denen eines eine
// Funktion ruft, die dieses Programm nicht kennt, muss die anderen 248 liefern
// (§8).
//
// **`value` steht dem Skript zur Verfügung** — MSE gibt den bisherigen Wert
// mit (`invokeOnDefault(ctx, value)`), und die Vorgabewerte des Packs rechnen
// damit: `separator_default(value, margin_code: "sep1")`. Ohne diese Bindung
// scheiterten **21.280** von ihnen über die 190 Stile
// (#295).
func vorgabewert(
	ctx context.Context, feld domain.Field, umgebung *msescript.Umgebung, bisher string,
) (string, error) {
	roh := strings.TrimSpace(feld.Default)

	if roh == "" {
		return Anfangswert(feld), nil
	}

	// **Ein `default:` ist ein Skript, auch ohne geschweifte Klammern.**
	//
	// `magic.mse-game/card_fields` führt 142 Vorgabewerte, und **alle 142**
	// beginnen mit `stylesheet; `:
	//
	//	default: stylesheet; card_color(casting_cost: card.casting_cost, …)
	//
	// `a; b` ist eine Folge — `stylesheet` wird ausgewertet und verworfen,
	// `card_color(…)` ist das Ergebnis. Das vorangestellte `stylesheet` ist
	// eine Abhängigkeitsangabe: es sagt MSE, wann neu zu rechnen ist.
	//
	// **Keines ist geklammert**, und bis zum 2026-08-16 verlangte dieser Weg
	// genau das. Er ließ damit **jeden einzelnen** fallen und reichte den rohen
	// Text durch — eine Karte ohne `card color` bekam die Zeichenkette
	// `stylesheet; card_color(…)` als Farbe
	// (#298).
	//
	// **Die Gegenprobe steht in `set_fields`**: dort heißt es
	// `default: "Type custom types here; …"` — in Anführungszeichen, also eine
	// Zeichenkette **in Skriptsyntax**. Beide Formen sind Skript.
	rumpf, geklammert := skriptrumpf(roh)
	if !geklammert {
		rumpf = roh
	}

	eigene := umgebung.Darueber()
	eigene.Setze("value", msescript.Text(bisher))

	wert, err := msescript.AuswertenQuelle(ctx, rumpf, eigene)
	if err != nil {
		return "", err
	}

	// **Ein Fehlerwert ist kein Vorgabewert** — derselbe Grund wie beim
	// Wertskript: MSEs Zuweisung ruft `toString()`, und das wirft.
	if wert.IstFehler() {
		return "", errors.New(wert.Text)
	}

	return wert.AlsText(), nil
}

// mitVorgaben ergänzt die Felder einer Karte um die, die sie nicht führt.
//
// **Der Wert der Karte schlägt die Vorgabe.** Andernfalls wäre die Vorgabe
// keine Vorgabe, sondern eine Festlegung.
//
// # Warum in einer eigenen Ebene ausgewertet
//
// Ein Vorgabewert darf ein Skript sein, und dieses Skript läuft, bevor `card`
// steht — es kann sich also nicht auf die Karte beziehen. Das ist kein Mangel,
// sondern die Lage: ein Vorgabewert gilt für **jede** Karte des Spiels.
// **Zwei Durchgänge, und der Grund steht in MSE2.**
//
//	value = field().default_script.invokeOnDefault(ctx, value)
//	      | field().        script.invokeOn(ctx, value);
//
// `TextValue::update`, `src/data/field/text.cpp`, Zeilen 178–179, Fassung
// f9d9356 vom 2021-07-06. **Beides läuft in `update`**, also nachdem die
// `init script:` von Spiel und Stil gelaufen sind — und die legen die
// Funktionen an, die ein Vorgabewert ruft.
//
// **Trotzdem läuft [mitVorgaben] vorher**, denn das Initskript liest die
// Kartenfelder: `magic-m15` fragt `card.card_symbol == "none"`
// (#341).
// Was dabei scheitert, bleibt leer und **still** — [mitVorgabeskripten] nimmt
// sich danach genau diese Felder noch einmal vor, wenn die Funktionen des
// Spiels dastehen, und meldet erst dann.
//
// Zwei Durchgänge statt einem, und der Preis steht dabei: ein Vorgabeskript,
// das schon früh gelingt, sieht die Funktionen des Initskripts **nicht** —
// aber es braucht sie auch nicht, sonst wäre es gescheitert.
func mitVorgaben(
	ctx context.Context, gefuellt []msescript.Eintrag, felder []domain.Field,
	umgebung *msescript.Umgebung,
) []msescript.Eintrag {
	da := make(map[string]bool, len(gefuellt))
	for i := range gefuellt {
		da[gefuellt[i].Name] = true
	}

	for i := range felder {
		feld := felder[i]

		// **Der Abbruch wirkt hier und nicht erst danach.**
		//
		// `magic.mse-game` führt 249 Kartenfelder, und 142 davon rechnen ein
		// Skript aus. Das ist die längste Strecke zwischen zwei
		// Kontextprüfungen im ganzen Vorschaulauf — wer während des Tippens
		// die Karte wechselt, wartet sonst darauf, dass die alte fertig
		// gerechnet wird. Angestrichen vom Zweitblick über die Phasen B bis D.
		if ctx.Err() != nil {
			return gefuellt
		}

		name := strings.ReplaceAll(feld.Name, " ", "_")
		if da[name] {
			continue
		}

		// **Ein gescheiterter Vorgabewert wird gemeldet, nicht verschwiegen.**
		//
		// Hier stand ein `return ""`, und es hat gekostet: `separator`
		// bekommt seinen Wert aus `separator_default(…)`, und ohne ihn baute
		// die Vorlage den Dateinamen `…/separators/.png`. Der Befund kam erst
		// beim Rastern und sprach über eine fehlende Datei statt über den
		// Namen, den die Umgebung nicht führt
		// (#295).
		//
		// **Der Wert bleibt leer.** Erfunden wird keiner.
		// **Der Fehler fällt hier weg** — [mitVorgabeskripten] geht die noch
		// leeren Felder nach den Initskripten erneut an und meldet dort.
		//nolint:errcheck // der zweite Durchgang meldet, siehe [mitVorgabeskripten]
		wert, _ := vorgabewert(ctx, feld, umgebung, "")

		gefuellt = append(gefuellt, msescript.Eintrag{Name: name, Wert: msescript.Text(wert)})
		da[name] = true
	}

	return gefuellt
}

// mitVorgabeskripten rechnet die `default:`-Skripte — nach den Initskripten.
//
// **Der zweite Durchgang.** Nur für ein leeres Feld: MSEs `invokeOnDefault` wendet das Skript an,
// solange der Wert der vorgegebene ist; sobald jemand etwas eingetragen hat,
// bleibt es stehen (`OptionalScript::invokeOnDefault`,
// `src/script/scriptable.hpp`). „Leer" ist hier die Näherung dafür, und sie
// steht so da, damit niemand sie für die Regel hält
// (#295).
//
// Fassung f9d9356 vom 2021-07-06.
func mitVorgabeskripten(
	ctx context.Context, felder []msescript.Eintrag, beschreibungen []domain.Field,
	umgebung *msescript.Umgebung,
) ([]msescript.Eintrag, []domain.Befund) {
	var befunde []domain.Befund

	stelle := make(map[string]int, len(felder))
	for i := range felder {
		stelle[felder[i].Name] = i
	}

	for beschrieben := range beschreibungen {
		feld := beschreibungen[beschrieben]

		if ctx.Err() != nil {
			return felder, befunde
		}

		if strings.TrimSpace(feld.Default) == "" {
			continue
		}

		i, da := stelle[strings.ReplaceAll(feld.Name, " ", "_")]
		if !da || felder[i].Wert.AlsText() != "" {
			continue
		}

		wert, err := vorgabewert(ctx, feld, umgebung, felder[i].Wert.AlsText())
		if err != nil {
			befunde = append(befunde, domain.Befund{
				Ort:     feld.Name,
				Meldung: "der Vorgabewert ließ sich nicht ausrechnen: " + err.Error(),
			})

			continue
		}

		felder[i].Wert = msescript.Text(wert)
	}

	return felder, befunde
}
