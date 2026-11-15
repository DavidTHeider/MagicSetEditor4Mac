package service

// **Die Parameterhilfen der Schlüsselwörter**
// (#1342,
// Werkstück D von Phase 4.11).
//
// # Was nachgelesen wurde (§15)
//
// `KeywordsPanel::onCommand` und `KeywordsPanel::runRefScript`
// (`src/gui/set/keywords_panel.cpp`, Zeilen 190–261, Fassung f9d9356 vom
// 2021-07-06). Zwei Hilfen, und sie zeigen auf **verschiedene Felder**:
//
//	// insert parameter → in `match`
//	String to_insert = "<atom-param>" + param->name + "</atom-param>";
//	match->insert(to_insert, "Insert parameter");
//
//	// refer parameter → in `reminder`
//	String to_insert = runRefScript(id - ID_PARAM_REF_MIN);
//	reminder->insert(to_insert, "Use parameter");
//
// # Warum der Text hier entsteht und nicht in der Oberfläche
//
// Weil er eine Regel des **Formats** ist und keine der Darstellung. Vier
// Flächen, die je selbst `<atom-param>` zusammensetzen, sind vier Fassungen
// derselben Regel — und die vierte weicht ab.

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Parametereinfuegung ist eine Parameterart, wie ein Menü sie anbietet.
type Parametereinfuegung struct {
	Name         string
	Beschreibung string

	// Einfuegen ist der Text, der ins **Muster** kommt.
	Einfuegen string
}

// Parameterverweis ist eine Form, in der ein Erinnerungstext auf einen
// Parameter zeigt.
type Parameterverweis struct {
	// Nummer ist die Stelle des Parameters im Muster, ab 1.
	//
	// **Sie zählt Parameter und nicht Einträge.** MSEs `++param` steht
	// außerhalb der Schleife über die Verweisskripte: zwei Skripte desselben
	// Parameters heißen beide `paramN`.
	Nummer int

	// Art ist der Name der Parameterart.
	Art string

	// Name ist der Name des Verweisskripts — leer bei der schlichten Form.
	Name         string
	Beschreibung string

	// Einfuegen ist der Text, der in den **Erinnerungstext** kommt.
	Einfuegen string
}

// Parameterangebot nennt die Parameterarten, die sich einfügen lassen.
func Parameterangebot(
	arten []msepaket.Schluesselwortparameterart,
) []Parametereinfuegung {
	angebot := make([]Parametereinfuegung, 0, len(arten))

	for _, art := range arten {
		angebot = append(angebot, Parametereinfuegung{
			Name:         art.Name,
			Beschreibung: art.Beschreibung,
			Einfuegen:    Parametermarke(art.Name),
		})
	}

	return angebot
}

// Parametermarke gibt die Marke, mit der ein Parameter im Muster steht.
func Parametermarke(art string) string {
	return "<atom-param>" + art + "</atom-param>"
}

// Parameterverweise nennt die Formen, in denen ein Erinnerungstext auf die
// Parameter dieses Musters zeigen kann.
//
// **Das Spielskript ist dabei**, weil ein Verweisskript rechnet:
// `\{cmc({input})\}` ruft `cmc` aus `magic.mse-game`. Ohne die Umgebung des
// Spiels scheiterte jedes zweite.
func Parameterverweise(
	ctx context.Context, muster string,
	arten []msepaket.Schluesselwortparameterart, spielskript string,
) ([]Parameterverweis, []domain.Befund) {
	folge, befunde := parameterfolge(muster, arten)

	verweise := make([]Parameterverweis, 0, len(folge))

	// **Die Umgebung steht einmal da**, nicht je Skript — dieselbe Überlegung
	// wie bei [zeilenAus].
	umgebung := spielumgebung(ctx, spielskript, nil, nil,
		Wortliste{}, &Schluesselwortverwendung{})

	for stelle, art := range folge {
		nummer := stelle + 1

		if len(art.Verweisskripte) == 0 {
			verweise = append(verweise, Parameterverweis{
				Nummer: nummer, Art: art.Name,
				Einfuegen: schlichterVerweis(nummer),
			})

			continue
		}

		gebaut, skriptbefunde := verweiseAusSkripten(ctx, umgebung, art, nummer)
		verweise = append(verweise, gebaut...)
		befunde = append(befunde, skriptbefunde...)
	}

	return verweise, befunde
}

// verweiseAusSkripten rechnet die Einträge einer Art mit Verweisskripten.
func verweiseAusSkripten(
	ctx context.Context, umgebung *msescript.Umgebung,
	art msepaket.Schluesselwortparameterart, nummer int,
) ([]Parameterverweis, []domain.Befund) {
	verweise := make([]Parameterverweis, 0, len(art.Verweisskripte))

	var befunde []domain.Befund

	for _, skript := range art.Verweisskripte {
		eigene := umgebung.Darueber()
		eigene.Setze("input", msescript.Text(parametername(nummer)))

		// **Ein Textskript und kein Ausdruck.** MSE führt es als
		// `StringScript script; ///< Code to insert into the reminder text
		// script` (`src/data/keyword.hpp`, Zeile 30, Fassung f9d9356 vom
		// 2021-07-06) — und deshalb steht im Pack `\{{input}\}`: eine
		// **maskierte** geschweifte Klammer um einen eingebetteten Ausdruck.
		// Als Ausdruck gelesen scheitert das am ersten `\`.
		wert, fehler := msescript.AuswertenTextskript(ctx, skript.Skript, eigene)
		if fehler != nil {
			// **Nur dieser Eintrag fällt aus, nicht das Menü** — MSE fängt
			// den Fehler je Skript (`catch (ScriptError const& e)`).
			befunde = append(befunde, domain.Befund{
				Ort: art.Name,
				Meldung: fmt.Sprintf(
					"das Verweisskript %q der Parameterart %q lief nicht: %v",
					skript.Name, art.Name, fehler),
			})

			continue
		}

		verweise = append(verweise, Parameterverweis{
			Nummer: nummer, Art: art.Name,
			Name: skript.Name, Beschreibung: skript.Beschreibung,
			Einfuegen: wert.AlsText(),
		})
	}

	return verweise, befunde
}

// parametername gibt den Namen, unter dem ein Parameter im Skript steht.
func parametername(nummer int) string {
	return fmt.Sprintf("param%d", nummer)
}

// schlichterVerweis ist die Form ohne Verweisskript — MSEs `"{" + param_s + "}"`.
func schlichterVerweis(nummer int) string {
	return "{" + parametername(nummer) + "}"
}

// parameterfolge gibt die Parameterarten eines Musters, in ihrer Reihenfolge.
//
// **Dieselbe Wanderung wie in [ausdruckFuer]**, und deshalb steht sie hier
// einmal: wer zweimal über `<atom-param>` liefe, hätte zwei Regeln darüber,
// was ein Parameter ist.
func parameterfolge(
	muster string, arten []msepaket.Schluesselwortparameterart,
) ([]msepaket.Schluesselwortparameterart, []domain.Befund) {
	nachName := make(map[string]msepaket.Schluesselwortparameterart, len(arten))
	for _, art := range arten {
		nachName[art.Name] = art
	}

	folge := []msepaket.Schluesselwortparameterart{}

	var befunde []domain.Befund

	rest := muster

	for {
		anfang := strings.Index(rest, "<atom-param>")
		if anfang < 0 {
			break
		}

		ende := strings.Index(rest[anfang:], "</atom-param>")
		if ende < 0 {
			break
		}

		name := rest[anfang+len("<atom-param>") : anfang+ende]

		art, bekannt := nachName[name]
		if !bekannt {
			befunde = append(befunde, domain.Befund{
				Ort: name,
				Meldung: fmt.Sprintf(
					"die Parameterart %q gibt es in diesem Spiel nicht", name),
			})
		} else {
			folge = append(folge, art)
		}

		rest = rest[anfang+ende+len("</atom-param>"):]
	}

	return folge, befunde
}
