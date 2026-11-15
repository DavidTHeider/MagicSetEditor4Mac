package sammelliste

// **Eine Sammlung ausgeben** — Werkstück 4 von Phase 4.5
// (#1292).
//
// # Zwei Formate, weil es zwei Fragen sind
//
//   - [SchreibeListe] gibt die Gestalt, die [Lies] versteht: einfügbar,
//     lesbar, und **verlustbehaftet**. Zustand, Erwerbspreis und Notiz stehen
//     nicht darin, weil das Format sie nicht kennt.
//   - [SchreibeCSV] gibt jede Spalte. Es ist die Gestalt, in der eine Sammlung
//     dieses Haus verlässt und wieder hereinkommt, ohne dass etwas fehlt.
//
// **Was die Liste verliert, steht ausgesprochen**, statt dass es jemand beim
// Zurücklesen merkt (§10).
//
// # Warum hier und nicht in `service`
//
// Weil es dieselbe Frage ist wie beim Lesen: *wie sieht eine Kartenliste aus?*
// Ein Schreiber in einer anderen Schicht liefe vom Leser weg, und der Rundlauf
// — schreiben, lesen, dasselbe bekommen — wäre eine Zusage über zwei Pakete.

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ausfuhrzeile ist ein Exemplar, wie es hinausgeht.
//
// **Ein eigener Typ und nicht [domain.Exemplar]**: dieses Paket ist ein Blatt
// und kennt keine Kennungen und keine Fassungen — was hier steht, ist der
// **Inhalt** einer Zeile, nicht ihr Platz im Bestand.
type Ausfuhrzeile struct {
	Printing string
	Menge    int
	Zustand  string
	Sprache  string

	// Finish ist die Ausführung — **leer heißt `nonfoil`** (T2, #1467).
	Finish domain.Finish

	Name    string
	Setcode string
	Nummer  string

	Erwerbsdatum    string
	Erwerbspreis    string
	Erwerbswaehrung string
	Notiz           string

	// Lagerort und Wunsch sind die Ordnung
	// ([0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
	//
	// **Sie gehören ins CSV und nicht in die Liste.** Das Listenformat kennt
	// sie nicht — und ein CSV, das sie verlöre, wäre keine Sicherung.
	Lagerort string

	// Wunschrang und Wunschart gelten **nur an einem Wunsch** (T4, #1469) —
	// leer heißt `normal` und `genau`.
	//
	// **Sie gehören ins CSV**, aus demselben Grund wie Lagerort und Wunsch:
	// eine Sicherung, die sie verlöre, käme ohne die Ordnung zurück, die
	// jemand von Hand gesetzt hat.
	Wunschrang string
	Wunschart  string
	Wunsch     bool

	// Verweisfassung, Kennungssystem und Kennung sind die **portable
	// Referenz** auf den Druck — Werkstück A2 von Phase 5.3 (#1441, 0086).
	//
	// **Neben `printing` und nicht an seiner Stelle**: der Alias bleibt, was
	// ein CSV von vor 0.5.0 trägt und was jeder Marktplatz versteht. Die
	// Referenz sagt zusätzlich, **welches System** die Kennung vergeben hat und
	// in welcher Fassung sie geschrieben wurde; Name, Setcode, Nummer und
	// Sprache daneben sind ihr Rückfall. Eine lokale Zeilennummer steht hier
	// nie.
	Verweisfassung int
	Kennungssystem string
	Kennung        string
}

// spalten sind die Namen der CSV-Kopfzeile, in ihrer Reihenfolge.
//
// **Sie stehen an einer Stelle**, und Schreiber wie Leser nehmen sie von
// hier — zwei Aufzählungen liefen an dem Tag auseinander, an dem eine Spalte
// dazukommt.
var spalten = []string{
	"printing", "menge", "zustand", "sprache", "finish",

	// **`foil` steht weiter da, obwohl `finish` es enthält** — Werkstück `T2`
	// (#1467). Eine Ausfuhr ist eine **Austauschdatei**: sie geht in ein
	// fremdes Werkzeug und in eine ältere Fassung dieses Programms zurück, und
	// beide lesen die Spalte, die sie kennen. Fiele sie weg, würde jede Zeile
	// beim Einlesen dort still gewöhnlich.
	//
	// **Für `etched` steht `false`** — was diese Spalte nicht sagen kann, sagt
	// sie nicht. Der Preis steht im Vermerk
	// [0119](../../docs/entscheidungen/0119-eine-ausfuehrung-statt-eines-wahrheitswerts.md).
	"foil",

	"name", "setcode", "nummer",
	"erwerbsdatum", "erwerbspreis", "erwerbswaehrung", "notiz",
	"lagerort", "wunsch", "wunschrang", "wunschart",
	"verweisfassung", "kennungssystem", "kennung",
}

// SchreibeListe gibt die Sammlung in der Gestalt, die [Lies] versteht.
//
// **Verlustbehaftet, und das ist der Zweck.** Wer eine Liste in einen
// Deckbaukasten bringt, will „4 Lightning Bolt (M11) 149" und nicht zwölf
// Spalten. Was verloren geht: Zustand, Sprache, Erwerbsdatum, Erwerbspreis,
// Währung und Notiz.
func SchreibeListe(zeilen []Ausfuhrzeile) string {
	var gebaut strings.Builder

	for _, zeile := range zeilen {
		gebaut.WriteString(strconv.Itoa(zeile.Menge))
		gebaut.WriteString(" ")

		// **Ohne lesbaren Namen die Kennung** — dieselbe Entscheidung wie bei
		// [domain.Exemplar.Bezeichnung]. Eine Zeile ohne beides fiele beim
		// Rundlauf heraus.
		if zeile.Name != "" {
			gebaut.WriteString(zeile.Name)
		} else {
			gebaut.WriteString(zeile.Printing)
		}

		if zeile.Setcode != "" {
			gebaut.WriteString(" (" + strings.ToUpper(zeile.Setcode) + ")")

			if zeile.Nummer != "" {
				gebaut.WriteString(" " + zeile.Nummer)
			}
		}

		// **Die Marke sagt, welche Ausführung** — `*F*` wie bisher, und
		// `*etched*`, weil `*F*` dafür falsch wäre.
		switch zeile.Finish {
		case domain.FinishFoil:
			gebaut.WriteString(" *F*")

		case domain.FinishEtched:
			gebaut.WriteString(" *etched*")
		}

		gebaut.WriteString("\n")
	}

	return gebaut.String()
}

// SchreibeCSV gibt die Sammlung mit jeder Spalte.
//
// **Über `encoding/csv` und nicht über `strings.Join`**: eine Notiz mit einem
// Komma oder einem Anführungszeichen bräche ein selbst gebautes Format, und
// „getauscht, 2026" ist eine Notiz, die jemand schreibt.
func SchreibeCSV(zeilen []Ausfuhrzeile) string {
	var gebaut strings.Builder

	schreiber := csv.NewWriter(&gebaut)

	// **Der Fehler kann hier nicht auftreten**: ein `strings.Builder` schlägt
	// beim Schreiben nie fehl. Ihn zu prüfen hieße, einen Zweig zu bauen, den
	// kein Fall betritt (§3) — deshalb steht der Grund hier statt eines `if`.
	//nolint:errcheck // ein strings.Builder scheitert nicht
	_ = schreiber.Write(spalten)

	for _, zeile := range zeilen {
		//nolint:errcheck // derselbe Grund wie darüber
		_ = schreiber.Write([]string{
			zeile.Printing, strconv.Itoa(zeile.Menge), zeile.Zustand,
			zeile.Sprache, string(ausfuehrungOder(zeile.Finish)),
			strconv.FormatBool(ausfuehrungOder(zeile.Finish) == domain.FinishFoil),
			zeile.Name, zeile.Setcode, zeile.Nummer,
			zeile.Erwerbsdatum, zeile.Erwerbspreis, zeile.Erwerbswaehrung,
			zeile.Notiz,
			zeile.Lagerort, strconv.FormatBool(zeile.Wunsch),
			zeile.Wunschrang, zeile.Wunschart,
			fassungstext(zeile.Verweisfassung), zeile.Kennungssystem, zeile.Kennung,
		})
	}

	schreiber.Flush()

	return gebaut.String()
}

// LiesCSV liest eine Sammlung aus CSV.
//
// **Pflichtfelder streng, Zusatzfelder tolerant** (§16): `printing` und
// `menge` müssen da sein; eine Spalte, die dieses Programm nicht kennt, wird
// übergangen statt bemängelt — wer aus einem Marktplatz einliest, bringt
// welche mit.
func LiesCSV(text string) ([]Ausfuhrzeile, []domain.Befund) {
	if strings.TrimSpace(text) == "" {
		return []Ausfuhrzeile{}, []domain.Befund{}
	}

	leser := csv.NewReader(strings.NewReader(text))

	// **Eine wechselnde Spaltenzahl ist kein Abbruch.** Der Leser wiese sie
	// sonst als Fehler ab, und eine einzige kurze Zeile machte die ganze Datei
	// unlesbar — §8 sagt das Gegenteil.
	leser.FieldsPerRecord = -1

	roh, fehler := leser.ReadAll()
	if fehler != nil {
		return []Ausfuhrzeile{}, []domain.Befund{{
			Meldung: "die Datei ist kein lesbares CSV: " + fehler.Error(),
			Art:     domain.ArtEingabe,
		}}
	}

	stellen, kopffehler := kopfLesen(roh[0])
	if kopffehler != nil {
		return []Ausfuhrzeile{}, []domain.Befund{{
			Ort: "Zeile 1", Meldung: kopffehler.Error(), Art: domain.ArtEingabe,
		}}
	}

	return zeilenLesen(roh[1:], stellen)
}

// ohneWagenruecklauf macht aus jedem Wagenrücklauf einen Zeilenumbruch.
//
// **Weil ein CSV ihn nicht heil zurückgibt** — Werkstück O11 von Phase 5.2
// (#1435),
// gefunden vom Fuzzlauf über den Rundlauf. `encoding/csv` liest `\r\n`
// innerhalb eines Feldes als `\n`; ein einzelnes `\r` bleibt dagegen stehen
// und wird beim **nächsten** Lesen zu einem Zeilenumbruch. Damit gab dieselbe
// Notiz beim ersten und beim zweiten Rundlauf zwei verschiedene Werte — und
// `docs/FAEHIGKEITEN.md` sagt zu, dass sich ein CSV **vollständig**
// zurücklesen lässt.
//
// **Vereinheitlicht wird beim Lesen und nicht beim Schreiben**: was aus einem
// Marktplatz kommt, trägt die Zeilenenden seines Rechners, und eine Notiz
// soll auf beiden Seiten dieselbe sein. Was dabei verloren geht, ist die
// **Art** des Umbruchs, nicht der Umbruch.
func ohneWagenruecklauf(text string) string {
	if !strings.ContainsRune(text, '\r') {
		return text
	}

	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
}

// kopfLesen ordnet die bekannten Spalten ihren Stellen zu.
func kopfLesen(kopf []string) (map[string]int, error) {
	stellen := map[string]int{}

	for stelle, name := range kopf {
		stellen[strings.ToLower(strings.TrimSpace(name))] = stelle
	}

	for _, pflicht := range []string{"printing", "menge"} {
		if _, da := stellen[pflicht]; !da {
			return nil, fmt.Errorf(
				"%w: die Spalte %q fehlt — ohne sie ist eine Zeile kein Exemplar",
				domain.ErrValidation, pflicht)
		}
	}

	return stellen, nil
}

// zeilenLesen baut die Ausfuhrzeilen — und liest weiter, wo eine nicht aufgeht.
func zeilenLesen(roh [][]string, stellen map[string]int) (
	[]Ausfuhrzeile, []domain.Befund,
) {
	zeilen := []Ausfuhrzeile{}
	befunde := []domain.Befund{}

	for nummer, felder := range roh {
		hole := func(name string) string {
			stelle, da := stellen[name]
			if !da || stelle >= len(felder) {
				return ""
			}

			return ohneWagenruecklauf(strings.TrimSpace(felder[stelle]))
		}

		menge, mengenfehler := strconv.Atoi(hole("menge"))
		if mengenfehler != nil || menge < 1 {
			befunde = append(befunde, domain.Befund{
				// **`nummer+2`**: die Kopfzeile ist Zeile 1.
				Ort: fmt.Sprintf("Zeile %d", nummer+2),
				Meldung: fmt.Sprintf("%q ist keine Stückzahl",
					hole("menge")),
				Art: domain.ArtEingabe,
			})

			continue
		}

		zeilen = append(zeilen, Ausfuhrzeile{
			Printing: hole("printing"), Menge: menge,
			Zustand: hole("zustand"), Sprache: hole("sprache"),
			Finish: ausfuehrungAus(hole("finish"), hole("foil")),
			Name:   hole("name"), Setcode: hole("setcode"), Nummer: hole("nummer"),
			Erwerbsdatum: hole("erwerbsdatum"), Erwerbspreis: hole("erwerbspreis"),
			Erwerbswaehrung: hole("erwerbswaehrung"), Notiz: hole("notiz"),
			Lagerort: hole("lagerort"), Wunsch: hole("wunsch") == "true",
			Wunschrang: hole("wunschrang"), Wunschart: hole("wunschart"),
			Verweisfassung: fassungAus(hole("verweisfassung")),
			Kennungssystem: hole("kennungssystem"), Kennung: hole("kennung"),
		})
	}

	return zeilen, befunde
}

// fassungstext schreibt die Fassung der Referenz — leer, wo es keine gibt.
//
// **Leer und nicht „0"**: eine Null läse sich wie eine Fassung, und es gibt
// keine nullte.
func fassungstext(fassung int) string {
	if fassung == 0 {
		return ""
	}

	return strconv.Itoa(fassung)
}

// fassungAus liest die Fassung der Referenz — 0, wo keine oder keine Zahl steht.
//
// **Ohne Befund**, aus demselben Grund wie beim Preis: ein CSV von vor 0.5.0
// hat die Spalte nicht, und eine fremde Datei mit einem Wort darin soll an
// ihren Karten gemessen werden, nicht an einer Spalte, die sie nicht kennt.
func fassungAus(text string) int {
	fassung, fehler := strconv.Atoi(strings.TrimSpace(text))
	if fehler != nil || fassung < 0 {
		return 0
	}

	return fassung
}

// ausfuehrungOder gibt die Ausführung — leer heißt `nonfoil`.
func ausfuehrungOder(finish domain.Finish) domain.Finish {
	if finish == "" {
		return domain.FinishNonfoil
	}

	return finish
}

// ausfuehrungAus liest die Ausführung aus einer Zeile — **und kennt die alte
// Spalte**.
//
// # Warum beide Spalten gelesen werden
//
// **Weil eine Ausfuhr von gestern `foil` heißt.** Bis `T2` schrieb dieser
// Schreiber `foil` mit `true`/`false`; eine Datei, die jemand aufgehoben hat,
// muss weiter einzulesen sein. Die Änderung ist damit **verträglich** im Sinn
// von [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
// geschrieben wird die neue Spalte, gelesen werden beide.
//
// **Und `finish` schlägt `foil`**: wer beide setzt, hat die genauere Angabe
// gemacht, und eine `1` daneben sagt nichts, was `etched` nicht schon sagte.
//
// **Was keine Ausführung ist, ist `nonfoil`** — nicht `etched`, und auch kein
// Fehler: eine Zeile aus fremder Hand darf eine Einfuhr nicht anhalten (§8).
func ausfuehrungAus(finish, foil string) domain.Finish {
	if gelesen := domain.Finish(finish); domain.IstFinish(gelesen) {
		return gelesen
	}

	if foil == "true" {
		return domain.FinishFoil
	}

	return domain.FinishNonfoil
}
