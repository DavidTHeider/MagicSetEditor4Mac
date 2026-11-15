package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Vier Namen, die jedes Skript eines Stils erwartet** (#214, #240).
//
// `styling`, `stylesheet`, `set` und `card_style` stehen in den Stilen von
// `M15-Magic-Pack` in Ausdrücken wie `styling.popout_image_style` oder
// `set.title`. Fehlt der Name, bricht der **ganze** Ausdruck ab — und mit ihm
// die Ebene, um die es ging.
//
// # `styling` und `stylesheet` tragen die Vorgaben des Stils
//
// Nicht eine leere Abbildung: der Stil erklärt seine `styling field:`-Blöcke
// samt `initial:`, und ohne sie rechnet ein `styling.imagetop * 2` mit nichts.
// Genau das war elfmal der Grund, warum eine Ebene von `magic-m15` ausfiel —
// „`*` rechnet mit Zahlen, hier steht Zahl und nichts".
//
// **`stylesheet` ist *fast* derselbe Wert unter zweitem Namen.** In den Stilen stehen
// beide für die Einstellungen der Vorlage; welcher benutzt wird, wechselt von
// Datei zu Datei.
//
// # `set` trägt die Angaben des Sets, `card_style` bleibt hier leer
//
// Gezählt über `M15-Magic-Pack` sind es **420 Zugriffe auf `set.symbol`**, 345
// auf `set.border_color` und ein langer Schwanz; nach `stylesheet.card_width`
// der meistgelesene Name des Packs.
//
// `card_style` wird nicht hier gesetzt, sondern in [stilplan] nach den
// Vorgaben — es braucht die gemessenen Inhaltsmaße.
func belegeVorlage(
	ctx context.Context, umgebung *msescript.Umgebung, vorlage Vorlage,
	bilder render.Bildquelle, karte *domain.CardRow,
) {
	stil := vorlage.Stil
	optionen := stil.Stiloptionen

	vorgaben := make([]msescript.Eintrag, 0, len(optionen))

	farben := farbstilfelder(optionen)

	// **Erst die Anfangswerte, dann die Vorgaben** — MSEs zwei Schritte
	// (#789). Ein `default:`, das `styling.andere_option` liest, sieht damit
	// deren `initial:` beziehungsweise erste Wahl.
	anfang := anfangsstyling(optionen)

	for _, option := range optionen {
		name := strings.ReplaceAll(option.Name, " ", "_")

		// **Ein Farbfeld trägt eine Farbe** — `styling.text_color` wird
		// gelesen, wo eine Farbe hingehört: `font_color(styling.text_color)`,
		// 72-mal im Pack (`farbfelder.go`).
		if feld, ist := farben[name]; ist {
			vorgaben = append(vorgaben,
				msescript.Eintrag{
					Name: name,
					Wert: farbfeldwert(stiloptionsvorgabeMit(ctx, option, anfang), feld),
				})

			continue
		}

		// **Ohne `initial:` gilt die erste Wahl** — dieselbe Regel wie beim
		// Kartenfeld ([stiloptionsvorgabe], #409).
		vorgaben = append(vorgaben, msescript.Eintrag{
			Name: name,
			Wert: msescript.Text(stiloptionsvorgabeMit(ctx, option, anfang)),
		})
	}

	// **Und darüber, was die Karte selbst nennt**
	// (#547).
	//
	//	IndexMap<FieldP,ValueP>& Set::stylingDataFor(const CardP& card) {
	//	  if (card && card->has_styling) return card->styling_data;
	//	  else                           return stylingDataFor(stylesheetFor(card));
	//	}
	//
	// `src/data/set.cpp`, Zeilen 102–105, Fassung f9d9356 vom 2021-07-06.
	//
	// **Alle 25.148 kanonischen Karten tragen `has_styling: true`**, und die
	// Werte entscheiden über die Gestalt: `magic-m15-mainframe-dfc` liest
	// `styling.front_style` und zeichnet danach eine Saga oder eine
	// gewöhnliche Karte. Ohne diese Zeilen galt immer das `initial:` — und die
	// 23 doppelseitigen Sagas des Kanons waren die weitesten Karten überhaupt.
	// **Dazwischen, was das Set setweit nennt** — der zweite Zweig
	// (#557):
	//
	//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
	//	  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
	//	}
	//
	// `src/data/set.cpp`, Zeilen 99–101. **Die Reihenfolge ist die Aussage**:
	// die Karte schlägt das Set, das Set schlägt die Vorgabe des Stils. Bei
	// MSE steht sie im `if`/`else` — hier in der Reihenfolge der beiden
	// Überlagerungen.
	//
	// **Der Schlüssel ist der präfixierte Paketname**, `magic-m15-saga` und
	// nicht `m15-saga`. So schreibt es `magic-set-generator` (§11), und so
	// steht es in jedem kanonischen Set.
	vorgaben = mitKartenstil(vorgaben, setstilwerte(vorlage.Setangaben, stilname(stil)), farben)
	vorgaben = mitKartenstil(vorgaben, stilwerteNachName(vorlage.Stilwerte), farben)
	vorgaben = mitKartenstil(vorgaben, kartenstilwerte(karte), farben)

	einstellungen := msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: vorgaben}

	// **Zwei Gegenstände — und `styling` ist keiner** (#1366).
	//
	// # Was der Altcode bindet
	//
	//	ctx.setVariable(SCRIPT_VAR_set,        make_intrusive<ScriptObject<Set*>>(&set));
	//	ctx.setVariable(SCRIPT_VAR_stylesheet, to_script(stylesheet));
	//	ctx.setVariable(SCRIPT_VAR_styling,    to_script(&set.stylingDataFor(*stylesheet)));
	//
	// `src/script/script_manager.cpp`, Zeilen 39–44, Fassung `f9d9356` vom
	// 2021-07-06.
	//
	// # Warum das zwei und nicht drei sind
	//
	// `set` und `stylesheet` werden `ScriptObject<T>`; ihr `type()` ist
	// `SCRIPT_OBJECT`, und `equal` landet für sie bei `compareAs` →
	// `COMPARE_AS_POINTER`. **Der Gegenstand entscheidet.**
	//
	// `stylingDataFor` gibt eine `IndexMap`, und daraus wird ein `ScriptMap`
	// mit `type() == SCRIPT_COLLECTION` (`src/script/to_value.hpp`, Zeilen
	// 169–190). In `equal` steht der Zweig für **zwei Sammlungen vor** dem
	// Zeigervergleich (`value.cpp`, Zeilen 107–117): sie werden **gliedweise**
	// verglichen. Eine Kennung an `styling` machte hier also etwas anderes als
	// MSE.
	//
	// **Das ist eine Berichtigung von #1366**, der es zu den Gegenständen
	// zählte. Nachgelesen am 2026-09-01.
	umgebung.Setze("styling", einstellungen)

	stilwert := stilangaben(einstellungen, stil)

	// **Je Stilpaket einer** — MSE führt genau einen `StyleSheet` je Name und
	// legt seinen Kontext danach an (`contexts.try_emplace(stylesheet.get())`,
	// dieselbe Datei, Zeile 31).
	stilwert.Kennung = "stil:" + stilname(stil)

	umgebung.Setze("stylesheet", stilwert)

	setwert := setangaben(ctx, umgebung, vorlage.Setfelder, vorlage.Setangaben)

	// **Ohne Setkennung im Namen**, und das ist ausgesprochen: eine Umgebung
	// trägt genau ein Set, und zwei Umgebungen begegnen einander in keinem
	// Ausdruck. Käme das je vor, gehörte sie hierher.
	setwert.Kennung = "set"

	umgebung.Setze("set", setwert)
	umgebung.Setze("exists_in_package", dateifrage(bilder))

	// **Und die Schlüsselwörter** (#1061). `msescript` bringt
	// `expand_keywords` als Stummel mit — er gibt den Text zurück und setzt
	// nichts ein, weil die Sprache keine Wortliste kennt. Hier steht sie zur
	// Verfügung, und der Name wird überschrieben.
	//
	// **Auch mit leerer Liste**: dann tut die Funktion, was der Stummel tat,
	// und der Weg ist derselbe. Zwei Wege wären zwei Verhalten.
	// **Die Schlüsselwortnaht steht hier nicht mehr**, seit Befund 5 des
	// Zweitblicks über Phase 4.11: `text_filter` friert die Funktion beim
	// Initialisieren des **Spiels** ein, und was danach auf der Umgebung der
	// Karte steht, erreicht die Kette nie. Sie steht jetzt eine Ebene höher —
	// `schluesselwortnaht.go`.

	// **Die Fassungen kommen aus zwei Quellen** — das Spiel erklärt sie am
	// `set field:` (#323), die Vorlage ergänzt sie unter `set info style:`.
	// **Die des Spiels steht vorn und gewinnt**; die Begründung steht bei
	// [ErsteFassung] (#548).
	umgebung.Setze("symbol_variation", symbolfassungsfrage(
		append(append([]msepaket.Symbolfassung{}, vorlage.Symbolfassungen...),
			stil.Symbolfassungen...)))
}

// feldstile baut `card_style` — je Feld seine Maße.
//
// # Warum das nicht leer bleiben darf
//
// Elf Ebenen von `magic-m15` fielen mit „`*` rechnet mit Zahlen, hier steht
// Zahl und nichts" aus: die Stile rechnen mit `card_style.<feld>.content_width`
// und multiplizieren das Ergebnis. Eine leere Abbildung gibt dort `nichts`,
// und `nichts * 2` ist kein Fehler der Vorlage, sondern einer der Umgebung.
//
// # Was hier steht und was nicht
//
// **`content_width` und `content_height` sind die Maße des *Inhalts*, wo er
// sich messen lässt** — bei einem Symbolzug tut er das, siehe [inhaltsmasse].
// Sonst stehen dort die Maße des Feldes, wie der Stil sie erklärt.
//
// Das ist der zweistufige Satz aus
// #291.
//
// # Jedes Kartenfeld des Spiels steht darin
//
// Die Abbildung entsteht bei MSE aus den Feldern des **Spiels**, nicht aus
// denen der Vorlage:
//
//	REFLECT_IF_READING {
//	  card_style.init(game->card_fields);
//	  …
//	}
//	REFLECT(card_style);
//
// `src/data/stylesheet.cpp`, Zeilen 106–113, Fassung f9d9356 vom 2021-07-06;
// `IndexMap::init` legt für **jeden** Schlüssel einen Eintrag an
// (`src/util/index_map.hpp`, Zeilen 48–61). Was die Vorlage nicht beschreibt,
// bekommt einen Stil mit Nullmaßen — erst danach überschreibt `REFLECT`, was
// dasteht. Ohne diesen Schritt ist `card_style.<feld>` **nichts**, und
// `content_width` daran ein Fehler statt einer Null.
//
// **Ein Feldstil ohne Kartenfeld bleibt trotzdem stehen.** MSE ließe ihn
// fallen, weil seine Abbildung über den Feldindex läuft; hier wäre das ein
// stiller Verlust in jeder Vorrichtung, die kein Spiel mitbringt.
func feldstile(
	ctx context.Context, spielfelder []domain.Field, felder []msepaket.Feldstil,
	masse map[string]Inhaltsmass, umgebung *msescript.Umgebung,
) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(spielfelder)+len(felder))

	// Über den Index: ein `Feldstil` wiegt 296 Byte, und `gocritic` zählt jede
	// Kopie in einer Schleife nach.
	alle := ungestylteErgaenzt(spielfelder, felder)

	for i := range alle {
		feld := &alle[i]

		inhaltBreite, inhaltHoehe := feldmass(feld.Breite), feldmass(feld.Hoehe)
		zeilen := msescript.Zahl(0)
		gemessen := masse[feld.Feld]

		if _, da := masse[feld.Feld]; da {
			inhaltBreite = msescript.Zahl(gemessen.Breite)
			inhaltHoehe = msescript.Zahl(gemessen.Hoehe)
			zeilen = msescript.Zahl(float64(gemessen.Zeilen))
		}

		eintraege = append(eintraege, msescript.Eintrag{
			Name: strings.ReplaceAll(feld.Feld, " ", "_"),
			Wert: msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: append([]msescript.Eintrag{
				{Name: "content_width", Wert: inhaltBreite},
				{Name: "content_height", Wert: inhaltHoehe},

				// **`content_lines` steht 221-mal im Pack**, 174-mal für das
				// Feld `text`: `magic-m15` entscheidet daran, ob ein kurzer
				// Regeltext mittig steht (#291).
				{Name: "content_lines", Wert: zeilen},

				// **Und das ganze Layout** (#338).
				{Name: "layout", Wert: layoutAus(gemessen)},
			}, append(kanten(ctx, feld, umgebung), textangaben(feld)...)...)},
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// ungestylteErgaenzt stellt die Feldstile in der Reihenfolge des Spiels auf und
// legt für jedes ungestylte Kartenfeld einen leeren an.
//
// **Der leere trägt Nullmaße** — das ist der frisch angelegte `Style` von MSE
// und keine Erfindung: eine Vorlage, die ein Feld nicht beschreibt, zeichnet
// es nicht, und `content_width` daran ist null.
func ungestylteErgaenzt(spielfelder []domain.Field, felder []msepaket.Feldstil) []msepaket.Feldstil {
	alle := make([]msepaket.Feldstil, 0, len(spielfelder)+len(felder))
	gestylt := make(map[string]bool, len(felder))

	for i := range felder {
		gestylt[felder[i].Feld] = true
	}

	alle = append(alle, felder...)

	for i := range spielfelder {
		if !gestylt[spielfelder[i].Name] {
			alle = append(alle, msepaket.Feldstil{Feld: spielfelder[i].Name})
		}
	}

	return alle
}

// masszahl deutet eine rohe Maßangabe des Stils als Zahl.
//
// **Was kein fester Wert ist, wird null.** Eine Angabe kann ein Skript sein
// (`{ … }`), und das hier auszurechnen hieße, den Planbauer zweimal zu bauen.
// Null ist an dieser Stelle kein erfundener Wert, sondern der Anfang des
// Koordinatensystems — und sie lässt die Rechnung laufen, statt sie
// abzubrechen.
func feldmass(roh string) msescript.Wert {
	zahl, err := strconv.ParseFloat(strings.TrimSpace(roh), 64)
	if err != nil {
		return msescript.Zahl(0)
	}

	return msescript.Zahl(zahl)
}

// stilangaben ist `stylesheet` — die Optionen **und** die Maße der Vorlage.
//
// # Die Maßproblematik, und woran sie hing
//
// Bis zum 2026-08-16 war `stylesheet` derselbe Wert wie `styling`: die
// Styling-Optionen und sonst nichts. Über die 190 Stile von `M15-Magic-Pack`
// scheiterten dadurch **464 Maßangaben** mit „die Angabe ergab keine Zahl" —
// und der Wert dahinter war jedes Mal leer.
//
// Die Spur führt über drei Ecken:
//
//	corners:  width: { face_coordinates_map(1).width }
//	magic.mse-game/script:2905  face_coordinates_map := { fc[input-1] or else […] }
//	magic.mse-game/script:2891  faces_coordinates := { [[ … width: stylesheet.card_width … ]] }
//
// **`card_width` stand nicht darin**, also war die Breite nichts, also die
// Abbildung nichts, also die Maßangabe leer.
//
// # Warum das der meistgelesene Wert des Packs ist
//
// Gezählt am 2026-08-16: **499 Zugriffe auf `stylesheet.card_width`, 506 auf
// `.card_height`**. Kein anderer Name kommt so oft vor.
//
// # Was noch fehlt, und es steht hier statt nirgends
//
// `stylesheet.card_style` (36), `.png` (24), `.short_name` (20) und
// `.folder_name` (1). Sie sind **nicht** gesetzt: was sie bedeuten, ist nicht
// erschlossen, und ein geratener Wert sähe aus wie ein gelesener. Wer sie
// braucht, findet hier den Grund statt einer Lücke.
func stilangaben(einstellungen msescript.Wert, stil msepaket.Stil) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(einstellungen.Eintraege)+2)
	eintraege = append(eintraege, einstellungen.Eintraege...)
	eintraege = append(eintraege,
		msescript.Eintrag{Name: "card_width", Wert: msescript.Zahl(float64(stil.Breite))},
		msescript.Eintrag{Name: "card_height", Wert: msescript.Zahl(float64(stil.Hoehe))},

		// **Der kurze Name eines Pakets, mit MSEs Rückfall** — die Regel
		// steht in [Kurzname]. 190 Befunde über die 190 Stile, einer je Stil
		// (#270).
		msescript.Eintrag{Name: "short_name", Wert: msescript.Text(Kurzname(stil.KurzerName, stil.Paket))},
		msescript.Eintrag{Name: "full_name", Wert: msescript.Text(stil.VollerName)},
	)

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// stilname gibt den Namen, unter dem die setweiten Stiloptionen liegen.
//
//	return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
//
// `src/data/set.cpp`, Zeile 100, Fassung f9d9356 vom 2021-07-06 —
// `Packaged::name()` ist der **Dateiname ohne Endung**, also `magic-m15-saga`
// und nicht die Kurzform `m15-saga`.
//
// **Nachgesehen an der dritten Quelle** (§11): `magic-set-generator` schreibt
// den präfixierten Namen (`src/mse.rs`, Zeilen 151–174), und jedes kanonische
// Set führt ihn so.
func stilname(stil msepaket.Stil) string {
	return strings.TrimSuffix(stil.Paket, ".mse-style")
}
