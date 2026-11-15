package msescript

import (
	"context"
	"regexp"
	"strings"
)

// `remove_tag` — eine **benannte** Auszeichnung entfernen.
//
// # Der Unterschied zu `remove_tags`
//
// `remove_tags` nimmt **alle** Auszeichnungen heraus; `remove_tag(tag: "<sym")`
// nimmt eine bestimmte samt ihrem Abschluss und lässt den Inhalt stehen. Beide
// stehen in `magic.mse-game`
// (#216):
//
//	remove_tag(input, tag: "<word-list")
//	remove_tag(input, tag:"<sym-auto>")
//	remove_tag(subtype_separator, tag: "<atom-sep")
//
// **Die Angabe ist ein Anfang, kein vollständiger Name.** `"<word-list"` trifft
// `<word-list-land>` ebenso wie `<word-list-type>`; so steht es in zwei der
// drei Fundstellen, und die dritte nennt den Namen vollständig samt `>`. Beides
// geht auf, wenn der Anfang zählt.
func init() {
	eintragen("remove_tag", entferneAuszeichnung)
}

func entferneAuszeichnung(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("tag")
	if !benannt {
		return Nichts, ruf.fehlerAn("remove_tag braucht ein Argument `tag:`")
	}

	if err := ruf.eineEingabe("remove_tag"); err != nil {
		return Nichts, err
	}

	name := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(roh.AlsText()), "<"), ">")
	if name == "" {
		return ruf.Stellung[0], nil
	}

	// Die öffnende und die schließende Form derselben Auszeichnung — und nur
	// sie. Ein `<word-listen>` bliebe stehen, wenn jemand `<word-list>`
	// meinte; das ist der Preis dafür, dass der Anfang zählt, und er steht am
	// Kopf dieser Datei.
	// **`MustCompile` und kein Fehlerzweig.** Der Name geht durch
	// `QuoteMeta`; was danach herauskommt, ist immer ein gültiges Muster. Ein
	// Fehlerpfad, den man beseitigen kann, ist besser als einer, den man
	// nachstellt — das ist Punkt 1 aus „Was Phase 2 gelehrt hat".
	muster := gemerkterAusdruck(`</?` + regexp.QuoteMeta(name) + `[^>]*>`)

	return Text(muster.ReplaceAllLiteralString(ruf.Stellung[0].AlsText(), "")), nil
}

// `regex_escape` macht aus einem Text ein Muster, das genau ihn trifft.
//
// `magic.mse-game` baut damit Suchmuster aus Kartennamen — ein Name mit `(`
// oder `+` darin wäre sonst ein Muster und träfe etwas anderes oder nichts
// (#216).
func init() {
	eintragen("regex_escape", musterschutz)
}

func musterschutz(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("regex_escape"); err != nil {
		return Nichts, err
	}

	return Text(regexp.QuoteMeta(ruf.Stellung[0].AlsText())), nil
}

// `tag_contents` ersetzt den **Inhalt** einer Auszeichnung, nicht sie selbst.
//
// `tag_contents@(tag: "<soft>", contents: {""})` aus `magic.mse-game`: die
// Auszeichnung bleibt stehen, was zwischen ihr steht, wird durch das Ergebnis
// von `contents` ersetzt. `contents` ist eine **Funktion** und bekommt den
// bisherigen Inhalt als `input`
// (#216).
func init() {
	eintragen("tag_contents", auszeichnungsinhalt)
}

func auszeichnungsinhalt(ctx context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("tag")
	if !benannt {
		return Nichts, ruf.fehlerAn("tag_contents braucht ein Argument `tag:`")
	}

	inhalt, benannt := ruf.Benannter("contents")
	if !benannt {
		return Nichts, ruf.fehlerAn("tag_contents braucht ein Argument `contents:`")
	}

	if err := ruf.eineEingabe("tag_contents"); err != nil {
		return Nichts, err
	}

	name := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(roh.AlsText()), "<"), ">")
	if name == "" {
		return ruf.Stellung[0], nil
	}

	// Öffnende Auszeichnung, Inhalt, schließende — die Gruppe in der Mitte ist,
	// was ersetzt wird. **Nicht gierig**, damit zwei Vorkommen nacheinander
	// zwei Ersetzungen sind und nicht eine über alles.
	muster := gemerkterAusdruck(`(<` + regexp.QuoteMeta(name) + `[^>]*>)(.*?)(</` +
		regexp.QuoteMeta(name) + `[^>]*>)`)

	var fehler error

	ergebnis := muster.ReplaceAllStringFunc(ruf.Stellung[0].AlsText(), func(treffer string) string {
		teile := muster.FindStringSubmatch(treffer)

		neu, err := anwenden(ctx, inhalt, Text(teile[2]), ruf.Umgebung)
		if err != nil {
			fehler = err

			return treffer
		}

		return teile[1] + neu.AlsText() + teile[3]
	})

	if fehler != nil {
		return Nichts, fehler
	}

	return Text(ergebnis), nil
}

// anwenden ruft einen Funktionswert mit einem einzigen Stellungsargument.
//
// **Ein Wert, der keine Funktion ist, wird genommen, wie er ist.** In den
// Vorlagen steht bei `contents:` fast immer ein Block; eine Zeichenkette dort
// meint sie selbst.
//
// **Der Rumpf läuft in der Umgebung des Aufrufers**, nicht über der Bibliothek
// — MSE bindet dynamisch (`src/script/context.cpp`, Zeilen 466–472, Fassung
// f9d9356 vom 2021-07-06). Ohne `card` und `set` scheitert jeder Block, der
// mehr tut als seine Eingabe umzudrehen.
func anwenden(ctx context.Context, funktion, eingabe Wert, aussen *Umgebung) (Wert, error) {
	if funktion.Art != ArtFunktion || funktion.Funktion == nil {
		return funktion, nil
	}

	teil := funktion.Funktion
	ruf := Ruf{Stellung: []Wert{eingabe}, Benannt: teil.Gebunden, Umgebung: aussen}

	if teil.Eingebaut != nil {
		return teil.Eingebaut(ctx, ruf.mitEingabe())
	}

	if aussen == nil {
		aussen = Bibliothek()
	}

	umgebung := aussen.Darueber()
	umgebung.Setze("input", eingabe)

	for i := range teil.Gebunden {
		umgebung.Setze(teil.Gebunden[i].Name, teil.Gebunden[i].Wert)
	}

	return Auswerten(ctx, teil.Rumpf, umgebung)
}
