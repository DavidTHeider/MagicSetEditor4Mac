package msescript_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Der Vorrang wird an Stellen belegt, an denen die Pakete ihn erzwingen.
//
// Jeder Fall nennt seine Quelle. Wo eine Zeile den Vorrang **nicht** entscheidet,
// steht sie nicht hier, sondern in `docs/FORMATE.md` unter „noch offen" — eine
// Tabellenzeile, die eine Vermutung festschreibt, sieht aus wie ein Beleg.
func TestVorrangAnBelegtenBeispielen(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		baum   string
	}{
		{
			// `card.rarity=="a" or card.rarity=="b"` aus yugioh Zeile 116.
			//
			// **Das entscheidet den Vorrang wirklich**: bände `or` stärker,
			// stünde da `card.rarity == ("a" or card.rarity) == "b"` — ein
			// Ausdruck, den niemand schreiben wollte.
			name:   "Vergleich bindet stärker als oder",
			quelle: `a=="x" or a=="y"`,
			baum:   `(Binär or (Binär == Name(a) Zeichenkette(x)) (Binär == Name(a) Zeichenkette(y)))`,
		},
		{
			// `not hide_foil() and card.rarity=="rare"` aus yugioh Zeile 224.
			//
			// `not` gilt dem Aufruf, nicht der ganzen Verknüpfung: sonst hieße
			// die Zeile „nicht (verborgen und selten)", und die Nachbarzeilen,
			// die sich nur in der Seltenheit unterscheiden, ergäben keinen Sinn.
			name:   "nicht bindet stärker als und",
			quelle: `not f() and a=="r"`,
			baum:   `(Binär and (Unär not (Aufruf Name(f))) (Binär == Name(a) Zeichenkette(r)))`,
		},
		{
			// Aus derselben Zeile: `and` verkettet nach links.
			name:   "und verkettet nach links",
			quelle: `a and b and c`,
			baum:   `(Binär and (Binär and Name(a) Name(b)) Name(c))`,
		},
		{
			// `length(remove_tags(card.type_1))*17 + 2` aus yugioh Zeile 732.
			name:   "Malnehmen bindet stärker als Zusammenzählen",
			quelle: `f(a)*17 + 2`,
			baum:   `(Binär + (Binär * (Aufruf Name(f) Name(a)) Zahl(17)) Zahl(2))`,
		},
		{
			// `card_style.text.content_lines` aus magic-m15 Zeile 291.
			name:   "Punktzugriff verkettet nach links",
			quelle: `a.b.c`,
			baum:   `(Zugriff c (Zugriff b Name(a)))`,
		},
		{
			// `if … then 0 else if … then 14 else 29` aus yugioh ab Zeile 153.
			name:   "eine else-Kette hängt rechts",
			quelle: `if a then 0 else if b then 14 else 29`,
			baum: `(Bedingung Name(a) Zahl(0) ` +
				`(Bedingung Name(b) Zahl(14) Zahl(29)))`,
		},
		{
			// `sort_text@(order: "XY")()` aus magic-m15 Zeile 75.
			name:   "die Bindung ergibt eine Funktion, die gerufen wird",
			quelle: `f@(order: "XY")()`,
			baum:   `(Aufruf (Bindung Name(f) Zeichenkette(XY)))`,
		},
		{
			// `{not is_nopendulum()}` aus yugioh Zeile 110.
			name:   "eine geschweifte Klammer ist ein Wert",
			quelle: `{not f()}`,
			baum:   `(Klammer (Unär not (Aufruf Name(f))))`,
		},
		{
			// Aus `game_script`: mehrere Zuweisungen, der letzte Ausdruck zählt.
			name:   "eine runde Klammer fasst eine Folge",
			quelle: "(a := 1\n b := 2\n a)",
			baum: `(Folge (Zuweisung a Zahl(1)) ` +
				`(Zuweisung b Zahl(2)) Name(a))`,
		},
		{
			// Aus magic-m15 Zeile 288: die Klammer hebt den Vorrang auf.
			name:   "eine Klammer schlägt den Vorrang",
			quelle: `(a or b) and c`,
			baum:   `(Binär and (Binär or Name(a) Name(b)) Name(c))`,
		},
		{
			name:   "eine Zuweisung nimmt den ganzen Ausdruck rechts",
			quelle: `a := b and c`,
			baum:   `(Zuweisung a (Binär and Name(b) Name(c)))`,
		},
		{
			name:   "nil ist ein Wert",
			quelle: `a == nil`,
			baum:   `(Binär == Name(a) Nichts)`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			knoten, err := msescript.Zerteile(fall.quelle)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if knoten.String() != fall.baum {
				t.Errorf("Baum\n  %s\nerwartet war\n  %s", knoten.String(), fall.baum)
			}
		})
	}
}

// Mehrere Zuweisungen untereinander sind eine Folge.
//
// So sieht jedes `init script:` aus. Der Umbruch trennt, mehr tut er nicht —
// und leere Zeilen dazwischen ändern nichts.
func TestEinSkriptblockIstEineFolge(t *testing.T) {
	knoten, err := msescript.Zerteile("a := 1\n\n\nb := 2\n")
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	erwartet := `(Folge (Zuweisung a Zahl(1)) (Zuweisung b Zahl(2)))`
	if knoten.String() != erwartet {
		t.Errorf("Baum\n  %s\nerwartet war\n  %s", knoten.String(), erwartet)
	}
}

// Ein Ausdruck darf über mehrere Zeilen laufen.
//
// In `yugioh` steht die else-Kette so, und in `magic-m15` die and-Kette: der
// Umbruch fällt mitten in den Ausdruck. Ein Zerteiler, der ihn dort für ein
// Ende hält, liest die Hälfte jeder Vorlage falsch.
func TestEinAusdruckLaeuftUeberMehrereZeilen(t *testing.T) {
	knoten, err := msescript.Zerteile("if a then\n  0\nelse\n  1")
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	erwartet := `(Bedingung Name(a) Zahl(0) Zahl(1))`
	if knoten.String() != erwartet {
		t.Errorf("Baum\n  %s\nerwartet war\n  %s", knoten.String(), erwartet)
	}
}

func TestEinUnvollstaendigerAusdruckNenntSeineStelle(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		zeile  int
		spalte int
	}{
		{name: "eine Zuweisung ohne Wert", quelle: "a :=", zeile: 1, spalte: 5},
		{name: "ein Vergleich ohne rechte Seite", quelle: "a ==", zeile: 1, spalte: 5},
		// **`f(a` und `{a` stehen jetzt in `klammerstelle_test.go`.** Ihre
		// Stelle ist seit Werkstück B von Phase 4.9 die **öffnende**
		// Klammer, und sie gehören damit zu einer anderen Frage als die
		// übrigen hier: nicht „wo hört der Ausdruck auf", sondern „welche
		// Klammer hat niemand geschlossen".
		{name: "ein if ohne then", quelle: "if a else 1", zeile: 1, spalte: 6},
		{name: "nichts", quelle: "", zeile: 1, spalte: 1},
		{name: "ein Operator ohne alles", quelle: "* 2", zeile: 1, spalte: 1},
		{name: "eine Zuweisung an keinen Namen", quelle: "1 := 2", zeile: 1, spalte: 3},
		{name: "ein Rest, den niemand liest", quelle: "a b", zeile: 1, spalte: 3},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerteile(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Zeile != fall.zeile || befund.Spalte != fall.spalte {
				t.Errorf("Stelle %d:%d, erwartet war %d:%d — %s",
					befund.Zeile, befund.Spalte, fall.zeile, fall.spalte, befund.Meldung)
			}
		})
	}
}

// Ein Fehler des Lexers kommt unverändert durch.
func TestZerteileReichtDenLexerfehlerDurch(t *testing.T) {
	_, err := msescript.Zerteile("a = 1")

	var befund *msescript.Fehler
	if !errors.As(err, &befund) || befund.Spalte != 3 {
		t.Errorf("Fehler %v, erwartet war die Stelle 1:3", err)
	}
}

// Die Darstellung eines Knotens zeigt Form, Text und Kinder.
func TestDieDarstellungEinesKnotens(t *testing.T) {
	blatt := &msescript.Knoten{Form: msescript.FormNichts}
	if blatt.String() != "Nichts" {
		t.Errorf("Darstellung %q", blatt.String())
	}
}

// Ein Komma trennt Argumente, und weglassen darf man es auch.
//
// Beides kommt vor: `rgb(156, 130, 90)` mit Komma, `chosen(choice: "artifact")`
// mit einem einzigen. Ein Zerteiler, der auf dem Komma besteht, weist die
// Hälfte der Aufrufe ab.
func TestArgumenteMitUndOhneKomma(t *testing.T) {
	for _, quelle := range []string{`f(a, b)`, `f(a b)`} {
		t.Run(quelle, func(t *testing.T) {
			knoten, err := msescript.Zerteile(quelle)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			erwartet := `(Aufruf Name(f) Name(a) Name(b))`
			if knoten.String() != erwartet {
				t.Errorf("Baum %s, erwartet war %s", knoten.String(), erwartet)
			}
		})
	}
}

// Abbildung und Index — beim Bau des Auswerters nachgetragen.
//
// `template_prefix[type]` aus `magic-m15` Zeile 54 ließ sich vom Zerteiler aus
// #22 gar nicht lesen: `[` war weder Wert noch Nachsatz. Aufgefallen ist es
// erst, als der Auswerter einen Abbildungswert brauchte — die Grammatik in
// `docs/FORMATE.md` führte die Form die ganze Zeit, der Zerteiler nicht.
func TestAbbildungUndIndex(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		baum   string
	}{
		{
			name:   "eine Abbildung mit zwei Einträgen",
			quelle: `[card: "c.jpg", pt: "p.png"]`,
			baum:   `(Abbildung Zeichenkette(c.jpg) Zeichenkette(p.png))`,
		},
		{
			name:   "eine leere Abbildung",
			quelle: `[]`,
			baum:   `Abbildung`,
		},
		{
			name:   "ein Zugriff über den Index",
			quelle: `template_prefix[type]`,
			baum:   `(Index Name(template_prefix) Name(type))`,
		},
		{
			name:   "Index und Verkettung, wie in magic-m15 Zeile 54",
			quelle: `a[t] + input + b[t]`,
			baum: `(Binär + (Binär + (Index Name(a) Name(t)) Name(input)) ` +
				`(Index Name(b) Name(t)))`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			knoten, err := msescript.Zerteile(fall.quelle)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if knoten.String() != fall.baum {
				t.Errorf("Baum\n  %s\nerwartet war\n  %s", knoten.String(), fall.baum)
			}
		})
	}
}

func TestAbbildungUndIndexMeldenIhreStelle(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		spalte int
	}{
		// **`["a"]` war hier einmal ein Fehler und ist jetzt eine Liste.**
		// Eine Liste ist eine Abbildung mit Zahlen als Namen
		// (#186).
		// Was weiterhin nicht geht, ist ein Name ohne Wert dahinter.
		{name: "ein Name ohne Wert", quelle: `[a: ]`, spalte: 5},
		// `[a: 1` und `f[a` stehen jetzt in `klammerstelle_test.go` — siehe
		// dort, und den Grund in `TestEinUnvollstaendigerAusdruckNenntSeineStelle`.
		{name: "ein Index ohne Ausdruck", quelle: `f[]`, spalte: 3},
		{name: "ein Wert in einer Abbildung, der keiner ist", quelle: `[a: *]`, spalte: 5},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerteile(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Spalte != fall.spalte {
				t.Errorf("Spalte %d, erwartet war %d — %s",
					befund.Spalte, fall.spalte, befund.Meldung)
			}
		})
	}
}

// Beide Schreibweisen eines Skripts ergeben denselben Baum.
//
// **Das ist die eigentliche Feststellung**, und der zuvor genannte Testfall
// deckte nur die eingerückte Form. `docs/FORMATE.md` sagt „geklammert **oder**
// eingerückt", also gehört die Gleichwertigkeit geprüft, nicht eine der beiden.
// Gefunden vom Nachlauf des Zweitblicks über Phase 3.
func TestGeklammertUndEingerueckt(t *testing.T) {
	geklammert, err := msescript.Zerteile(`{ if a then 1 else 2 }`)
	if err != nil {
		t.Fatalf("geklammert: %v", err)
	}

	eingerueckt, err := msescript.Zerteile("if a then\n  1\nelse\n  2")
	if err != nil {
		t.Fatalf("eingerückt: %v", err)
	}

	// Die geklammerte Form ist ein Wert, der gerufen wird; die eingerückte ist
	// der Ausdruck selbst. Gleich ist, was **darin** steht.
	if geklammert.Kinder[0].String() != eingerueckt.String() {
		t.Errorf("geklammert %s, eingerückt %s",
			geklammert.Kinder[0].String(), eingerueckt.String())
	}
}
