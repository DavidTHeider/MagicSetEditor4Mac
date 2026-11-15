package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Tippfehler bleibt ein Abbruch.**
//
// Die Gegenprobe: ein Vorbehalt entsteht nur, wo die Vorlage **richtig** ist
// und dieses Programm sie nicht vollständig ausführen kann. Ein fehlendes
// `match:` ist keins von beidem.
func TestEinFehlendesArgumentBleibtEinAbbruch(t *testing.T) {
	if _, err := msescript.VorbelegenMitVorbehalten(t.Context(),
		`x := replace("a", replace: "")`, msescript.Bibliothek().Darueber()); err == nil {
		t.Error("kein Fehler — ein fehlendes match: ging als Vorbehalt durch")
	}
}

// **Ohne Vorbehalt ist die Liste leer**, und ein Abbruch bleibt ein Abbruch.
func TestOhneVorbehaltBleibtDieListeLeer(t *testing.T) {
	vorbehalte, err := msescript.VorbelegenMitVorbehalten(t.Context(),
		`x := replace("ab", match: "b", replace: "")`, msescript.Bibliothek().Darueber())
	if err != nil || len(vorbehalte) != 0 {
		t.Errorf("%d Vorbehalte, %v", len(vorbehalte), err)
	}

	if _, fehler := msescript.VorbelegenMitVorbehalten(t.Context(),
		`x := unbekannt()`, msescript.Bibliothek().Darueber()); fehler == nil {
		t.Error("ein unbekannter Name ging durch")
	}
}

// `tag_contents` ersetzt den Inhalt einer Auszeichnung, nicht sie selbst (#216).
func TestTagContentsErsetztDenInhalt(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{
			"ein Vorkommen",
			`tag_contents("<soft>x</soft>", tag: "<soft>", contents: {""})`,
			"<soft></soft>",
		},

		// **Nicht gierig**: zwei Vorkommen sind zwei Ersetzungen.
		{
			"zwei Vorkommen",
			`tag_contents("<b>a</b>-<b>c</b>", tag: "<b>", contents: {"!"})`,
			"<b>!</b>-<b>!</b>",
		},

		// **Der bisherige Inhalt kommt als `input`.**
		{
			"der Inhalt kommt an",
			`tag_contents("<b>a</b>", tag: "<b>", contents: { input + input })`,
			"<b>aa</b>",
		},

		{
			"keine solche Auszeichnung",
			`tag_contents("<i>a</i>", tag: "<b>", contents: {""})`, "<i>a</i>",
		},

		{
			"ein Name aus lauter Klammern",
			`tag_contents("<b>a</b>", tag: "<>", contents: {""})`, "<b>a</b>",
		},

		// **Ein Wert, der keine Funktion ist, meint sich selbst.**
		{
			"ein Text als contents",
			`tag_contents("<b>a</b>", tag: "<b>", contents: "fest")`, "<b>fest</b>",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **`tag_contents` reicht einen Fehler seines Blocks durch.**
func TestTagContentsReichtEinenFehlerDurch(t *testing.T) {
	for _, quelle := range []string{
		`tag_contents("<b>a</b>", tag: "<b>", contents: { unbekannt })`,
		`tag_contents("a", tag: "<b>")`,
		`tag_contents("a", contents: {""})`,
	} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Auch eine eingebaute Funktion darf bei `contents:` stehen.**
func TestTagContentsNimmtEineEingebauteFunktion(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`tag_contents("<b>ab</b>", tag: "<b>", contents: to_upper)`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "<b>AB</b>" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Auch [msescript.AuswertenMitVorbehalten] gibt einen Abbruch weiter.**
func TestAuswertenMitVorbehaltenReichtEinenAbbruchDurch(t *testing.T) {
	knoten, err := msescript.Zerteile(`unbekannt()`)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	if _, _, fehler := msescript.AuswertenMitVorbehalten(t.Context(), knoten, nil); fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Eine Funktion mit gebundenen Argumenten bei `contents:`.**
func TestTagContentsNimmtEineGebundeneFunktion(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`tag_contents("<b>a</b>", tag: "<b>", contents: ({ input + zusatz })@(zusatz: "!"))`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "<b>a!</b>" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Ein Skript ohne Folge geht am Stück, und sein Fehler bleibt ein Fehler.**
//
// Die Gegenprobe zur satzweisen Regel: sie greift nur, wo es Sätze gibt.
func TestEinEinzelnerAusdruckBleibtEinAbbruch(t *testing.T) {
	if _, err := msescript.VorbelegenMitVorbehalten(t.Context(), `unbekannt()`,
		msescript.Bibliothek().Darueber()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein unzerteilbares Skript bleibt ein Fehler und wird kein Vorbehalt.**
func TestEinUnzerteilbaresSkriptBleibtEinFehler(t *testing.T) {
	if _, err := msescript.VorbelegenMitVorbehalten(t.Context(), `x := (`,
		msescript.Bibliothek().Darueber()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein gescheiterter Satz einer Folge nimmt die übrigen nicht mit** (#240).
//
// `magic.mse-game` verlor an Zeile 14741 die 4.000 Zeilen danach — und in
// ihnen stand, was den Kartenrahmen zeichnet.
func TestEinGescheiterterSatzWirdEinVorbehalt(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()

	vorbehalte, err := msescript.VorbelegenMitVorbehalten(t.Context(),
		"vorher := \"da\"\nmitten := gibtesnicht()\nnachher := \"auch da\"\n", umgebung)
	if err != nil {
		t.Fatalf("das Skript brach ab: %v", err)
	}

	if len(vorbehalte) != 1 || !strings.Contains(vorbehalte[0].Meldung, "gibtesnicht") {
		t.Fatalf("%+v", vorbehalte)
	}

	// **Beide Nachbarn sind entstanden** — der davor und der danach.
	for _, name := range []string{"vorher", "nachher"} {
		if wert, da := umgebung.Suche(name); !da || wert.AlsText() == "" {
			t.Errorf("%q fehlt", name)
		}
	}

	// **Und der ausgefallene Name bleibt ungesetzt.** Ein erfundener Wert wäre
	// schlimmer als ein fehlender: er sähe aus wie ein Ergebnis.
	if _, da := umgebung.Suche("mitten"); da {
		t.Error("der ausgefallene Name ist gesetzt")
	}
}
