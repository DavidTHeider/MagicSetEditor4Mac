package msescript

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Wert ist das Ergebnis einer Auswertung.
//
// Ein Verbund mit einer Art statt einer Schnittstelle: die Sprache rechnet mit
// wenigen Arten, und jede Verknüpfung muss ohnehin über beide Seiten
// entscheiden. Eine Schnittstelle verteilte diese Entscheidung auf die Typen und
// machte aus `+` sechs Methoden.
type Wert struct {
	Art WertArt

	Zahl     float64
	Text     string
	Wahrheit bool

	// Feld ist der Name des Setfelds, aus dem dieser Wert stammt — leer für
	// alles, was gerechnet wurde, und **heute auch für Kartenfelder**.
	//
	//	SymbolStyleP style = dynamic_pointer_cast<SymbolStyle>(
	//	    set->stylesheetForP(card)->styleFor(value->fieldP));
	//
	// `src/script/functions/image.cpp`, Zeile 161, Fassung f9d9356 vom
	// 2021-07-06. **Bei MSE ist ein Feldwert ein Objekt, das sein Feld kennt**;
	// hier ist er eine Zahl oder ein Text, und ohne diese Zeile ginge die
	// Herkunft verloren.
	//
	// **Wer sie braucht**: `symbol_variation` sucht seine Fassung am Stil des
	// Feldes, aus dem der Wert kommt — `magic.mse-game` erklärt `symbol:` und
	// `masterpiece symbol:` mit **gleichnamigen, verschieden gefärbten**
	// Fassungen (#548).
	//
	// **Sie überlebt keine Rechnung**, und das ist richtig: `set.symbol + "x"`
	// ist kein Feldwert mehr.
	//
	// **Kartenfelder tragen sie noch nicht**, und das ist eine benannte
	// Grenze: `magic.mse-game` erklärt Symbolfassungen ausschließlich unter
	// `set field:`. Ein Kartenwert kommt damit mit leerem Feld an, und die
	// Suche fällt auf „gleich welches" zurück — das bisherige Verhalten.
	Feld string

	// Eintraege einer Abbildung. Geordnet, weil `docs/FORMATE.md` die Reihenfolge
	// in der Datei als Teil des Rundlaufs führt — und weil eine Go-Abbildung
	// keine hat.
	Eintraege []Eintrag

	// Glieder sind die **benannten** Glieder, wenn sie nicht die Einträge sind.
	//
	// **Ein Wert kann als Liste etwas anderes sein als beim Zugriff über einen
	// Namen** (#653).
	// `set` ist der Fall, für den es das gibt — in MSE zwei Gesichter an einem
	// Objekt:
	//
	//	ScriptValueP make_iterator(const Set& set) {
	//	  return make_intrusive<ScriptCollectionIterator<vector<CardP>>>(&set.cards);
	//	}
	//
	//	// in scripts, set.something is read from the set_info
	//	void reflect_set_info_get_member(GetMember& handler, const IndexMap<FieldP, ValueP>& data) {
	//	  REFLECT_NAMELESS(data);
	//	}
	//
	// `src/data/set.cpp`, Zeilen 251–253 und 271–275, Fassung f9d9356 vom
	// 2021-07-06.
	//
	// **Warum herum**: die Einträge tragen die Karten, damit Durchlauf,
	// `length`, `sort_list`, `position` und der Index von selbst richtig sind
	// — vierzig Stellen lesen eine Abbildung als Liste. Der benannte Zugriff
	// geht durch **eine**, und die sieht hier nach.
	Glieder []Eintrag

	// Verbund sagt, dass diese Abbildung einen **reflektierten C++-Verbund**
	// von MSE nachbildet — und damit sein Glied über den **Präfix** findet.
	//
	//	template <typename T>
	//	void handle(const Char* name, const T& object) {
	//	  if (!gdm.result() && canonical_name_compare(target_name, name)) {
	//	    gdm.handle(object);
	//	  }
	//	}
	//
	// `src/util/io/get_member.hpp`, Zeilen 91–96, und
	//
	//	bool canonical_name_compare(StringView as, const Char* b) {
	//	  for (String::const_iterator a_it = as.begin(); a_it != as.end(); ++a_it, ++b) {
	//	    if (*a_it != *b && !(*a_it == '_' && *b == ' ')) return false;
	//	  }
	//	  return true;
	//	}
	//
	// `src/util/string.cpp`, Zeilen 339–347, Fassung f9d9356 vom 2021-07-06.
	// **Die Schleife läuft über den gesuchten Namen und gibt am Ende `true`**
	// — ohne zu prüfen, ob der Gliedname zu Ende ist.
	//
	// **Nicht für jede Abbildung.** MSE benutzt diesen Vergleich für
	// reflektierte Verbunde, nicht für Skriptabbildungen; dort vergleicht es
	// genau. Ein Kartenfeld `text` fände sonst `text_box_mana_symbols`
	// (#494).
	Verbund bool

	// Kennung sagt, **welcher Gegenstand** dieser Wert ist.
	//
	// Herkunft: `equal` fragt am Ende `compareAs`, und ein `ScriptObject<T>`
	// antwortet `COMPARE_AS_POINTER` (`src/script/value.cpp` Zeilen 96–132 und
	// `src/script/to_value.hpp` Zeilen 184–187, Fassung `f9d9356` vom
	// 2021-07-06): **zwei Karten sind gleich, wenn es dieselbe Karte ist** —
	// nicht, wenn sie dieselben Felder tragen.
	//
	// **Ohne sie zählte eine Auszählung still doppelt.** `magic.mse-game`
	// schreibt `if card != set.cards.0 then "" else (…)`, und zwei
	// strukturgleiche Karten gaben beide den setweiten Wert aus (Befund 1 des
	// Zweitblicks über Phase 4.11).
	//
	// **Leer heißt: kein Gegenstand, sondern ein Wert.** Dann entscheidet die
	// Gestalt, und das ist für zwei gerechnete Abbildungen richtig.
	//
	// **Sie überlebt keine Rechnung**, so wie [Wert.Feld]: `card + "x"` ist
	// keine Karte mehr.
	//
	// **Wo sie gesetzt wird**, seit
	// #1366:
	// an `card` im Auszählungsweg, in der Kartenliste und in den Exportdaten,
	// und an `set` und `stylesheet` in der Stilumgebung.
	//
	// **Nicht an `styling`**, und das ist nachgelesen: MSE bindet es als
	// `ScriptMap` mit `type() == SCRIPT_COLLECTION`, und `equal` vergleicht
	// zwei Sammlungen **gliedweise**, bevor es zu `compareAs` kommt
	// (`src/script/to_value.hpp` Zeilen 169–190 und `src/script/value.cpp`
	// Zeilen 107–117, Fassung `f9d9356` vom 2021-07-06). Das Issue zählte es
	// zu den Gegenständen; der Quelltext sagt etwas anderes.
	//
	// **Und nicht an `card_style`** — das steht in dieser Umgebung gar nicht
	// (#195).
	Kennung string

	// Funktion trägt, was **nur** eine Funktion braucht.
	//
	// **Als Zeiger, und das ist keine Sparsamkeit um ihrer selbst willen.** Mit
	// den vier Feldern im Verbund wog ein `Wert` 128 Byte, und jede Schleife
	// über eine Wertliste kopierte sie mit — `gocritic` zählt das nach und
	// bekam recht. Aus 128 werden 80, und die vier Felder stehen dort, wo sie
	// hingehören: an der Funktion, nicht an der Zahl.
	Funktion *Funktionsteil

	// Marke ist ein Griff des Aufrufers, den kein Skript sehen kann.
	//
	// # Wozu
	//
	// `write_image_file(card, file: "…")` muss **diese** Karte rastern
	// (#216).
	// Was ein Skript unter `card` sieht, ist aber eine Abbildung aus Feldname
	// und Text — sie trägt keine Identität, und einen `id`-Eintrag
	// hineinzulegen wäre ein Feld, das in keinem MSE-Paket steht und über das
	// jede Vorlage stolperte, die über die Felder einer Karte läuft.
	//
	// # Warum das §2 nicht bricht
	//
	// Es ist eine **Zeichenkette**, keine Entität. `msescript` weiß nicht, was
	// darin steht, liest sie nie und gibt sie nie aus: sie erscheint weder in
	// [Wert.AlsText] noch beim Durchgehen einer Abbildung. Wer sie setzt, weiß
	// was sie bedeutet; wer sie liest, ist derselbe.
	//
	// # Was sie nicht ist
	//
	// **Kein Weg für Skripte.** Es gibt keinen Bezeichner, der sie erreicht,
	// und keine Funktion der Bibliothek, die sie liest. Eine Vorlage kann sie
	// nicht setzen, nicht lesen und nicht fälschen.
	Marke string
}

// Funktionsteil ist, was eine Funktion über einen gewöhnlichen Wert hinaus hat.
type Funktionsteil struct {
	// Rumpf einer Funktion: der Baum einer geschweiften Klammer.
	Rumpf *Knoten

	// Gebunden sind die mit `@` vorab gesetzten Argumente.
	Gebunden []Eintrag

	// Eingebaut ist gesetzt, wenn die Funktion in Go steht statt in einer
	// Vorlage. Dann bleibt `Rumpf` leer.
	Eingebaut Eingebaut

	// Netzname ist gesetzt, wenn hinter dem Namen ein anderes Skript desselben
	// Netzes steht. Der Aufruf geht dann über das Netz, damit die Abhängigkeit
	// mitgeschrieben wird.
	Netzname string
}

// NeueFunktion baut einen Funktionswert über einem Rumpf.
func NeueFunktion(rumpf *Knoten) Wert {
	return Wert{Art: ArtFunktion, Funktion: &Funktionsteil{Rumpf: rumpf}}
}

// WertArt benennt, was ein Wert ist.
type WertArt string

const (
	ArtNichts        WertArt = "nichts"
	ArtZahlwert      WertArt = "Zahl"
	ArtTextwert      WertArt = "Zeichenkette"
	ArtWahrheitswert WertArt = "Wahrheitswert"
	ArtAbbildung     WertArt = "Abbildung"
	ArtFunktion      WertArt = "Funktion"
)

// Eintrag ist ein benannter Wert.
type Eintrag struct {
	Name string
	Wert Wert
}

// Nichts ist der leere Wert.
var Nichts = Wert{Art: ArtNichts}

// Zahl baut einen Zahlwert.
func Zahl(zahl float64) Wert { return Wert{Art: ArtZahlwert, Zahl: zahl} }

// Text baut einen Zeichenkettenwert.
func Text(text string) Wert { return Wert{Art: ArtTextwert, Text: text} }

// Wahrheit baut einen Wahrheitswert.
func Wahrheit(ja bool) Wert { return Wert{Art: ArtWahrheitswert, Wahrheit: ja} }

// AlsText gibt die Darstellung, die `+` und die Ausgabe verwenden.
//
// **Eine Zahl ohne überflüssige Nullen**: `17` und nicht `17.000000`. In den
// Vorlagen wird gerechnet und das Ergebnis in Pfade und Texte gehängt
// (`"attribute/earth_" + get_lang_ex() + ".png"`); eine Nachkommastelle, die
// niemand geschrieben hat, landete dort im Dateinamen.
func (w Wert) AlsText() string {
	switch w.Art {
	case ArtTextwert:
		return w.Text

	// **Ein Datum ist hier seine kanonische Schreibweise** — der Grund steht
	// in `bibliothek_datum.go`.
	case ArtDatum:
		return w.Text

	case ArtZahlwert:
		// **`inf`, `-inf` und `nan` schreibt MSE klein und ohne Vorzeichen
		// beim Plus.**
		//
		//	String toString() const override { return String() << value; }
		//
		// `src/script/value.cpp`, Zeile 290, Fassung f9d9356 vom
		// 2021-07-06 — das ist `wxString::operator<<(double)`. Am Orakel:
		// `"x" + 1/0` ergibt `"xinf"`.
		//
		// Go schreibt `+Inf`, `-Inf` und `NaN`, und das stünde so auf der
		// Karte (#555).
		if math.IsInf(w.Zahl, 1) {
			return "inf"
		}

		if math.IsInf(w.Zahl, -1) {
			return "-inf"
		}

		if math.IsNaN(w.Zahl) {
			return "nan"
		}

		return strconv.FormatFloat(w.Zahl, 'f', -1, 64)

	case ArtWahrheitswert:
		if w.Wahrheit {
			return "true"
		}

		return "false"

	case ArtNichts:
		return ""

	case ArtFarbe:
		return fmt.Sprintf("rgba(%d,%d,%d,%d)", int(w.Eintraege[0].Wert.Zahl),
			int(w.Eintraege[1].Wert.Zahl), int(w.Eintraege[2].Wert.Zahl),
			int(w.Eintraege[3].Wert.Zahl))

	// **Eine Abbildung wird aneinandergesetzt.**
	//
	// So benutzen es die Exportvorlagen: `to_text(for each card in cards do
	// zeile())` ergibt den Text aller Zeilen hintereinander
	// (#51).
	// Eine Liste **ist** eine Abbildung mit Zahlen als Namen
	// (`docs/FORMATE.md`), und deshalb gilt dieselbe Regel für beide.
	//
	// **Die Namen fallen dabei weg**, und das ist die Aussage: `to_text` gibt
	// den Inhalt, nicht die Gestalt. Wer die Gestalt braucht, hat eine
	// Abbildung und keinen Text.
	case ArtAbbildung:
		var gebaut strings.Builder

		for _, eintrag := range w.Eintraege {
			gebaut.WriteString(eintrag.Wert.AlsText())
		}

		return gebaut.String()

	case ArtFunktion:
	}

	return string(w.Art)
}

// AlsWahrheit sagt, ob ein Wert als Bedingung zutrifft.
//
// **Der kurze Weg auf [Wert.AlsWahrheitswert]**, für die Stellen, die nur
// wissen wollen *ob* — die Meldung setzen sie selbst aus [Wert.Art] zusammen.
//
// Getragen sind ein Wahrheitswert, `nil` und die fünf Wörter, die MSE kennt
// (`yes`, `true`, `no`, `false`, der leere Text). Alles übrige ist ein Fehler.
//
// # Auch eine Zahl — aber nur als Bedingung eines `if`
//
// Sie steht **nicht** hier, sondern in `auswerter_formen.go` bei der
// Bedingung. Der Grund ist die Reichweite: belegt ist
// `if to_number(substring(…)) then …` in zwei Exportvorlagen
// (#216),
// **nicht** `1 and true` oder `not 1`. Wer die Lockerung hierher legte,
// bekäme sie überall — und die drei Fälle in `auswerter_stellen_test.go`, die
// heute einen Fehler verlangen, fielen still um.
//
// Das ist der Unterschied zwischen dem Fall und der Klasse, andersherum
// gelesen: eine Klasse, für die nur ein Fall belegt ist, ist keine.
func (w Wert) AlsWahrheit() (trifftZu, brauchbar bool) {
	trifft, err := w.AlsWahrheitswert()

	return trifft, err == nil
}

// Eintrag sucht einen Namen in einer Abbildung.
//
// **Bei einem [Wert.Verbund] genügt der Präfix**, und der **erste** Treffer
// gewinnt — MSEs `if (!gdm.result() && …)` prüft jedes Glied der Reihe nach
// und nimmt das erste, das passt. Ein genauer Treffer ist auch ein Präfix,
// deshalb braucht es dafür keinen zweiten Durchgang.
func (w Wert) Eintrag(name string) (Wert, bool) {
	// **Sind Glieder gesetzt, gelten allein sie** (#653). `set.0` gibt damit
	// keine Karte, und das ist MSEs Verhalten: `getMember` geht ins
	// `set_info` und nirgends sonst.
	unter := w.Eintraege
	if w.Glieder != nil {
		unter = w.Glieder
	}

	for i := range unter {
		if w.Verbund {
			if gliedtrifft(name, unter[i].Name) {
				return unter[i].Wert, true
			}

			continue
		}

		if unter[i].Name == name {
			return unter[i].Wert, true
		}
	}

	// **Ein Name, der eine Zahl ist, meint eine Stelle** — und zwar erst
	// jetzt, nachdem der benannte Zugriff nichts gefunden hat. Die Regel und
	// ihre Herkunft stehen bei [anStelle].
	return anStelle(unter, name)
}
