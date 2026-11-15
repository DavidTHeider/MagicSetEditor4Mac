package domain

// **Schlüsselwortfähigkeiten** — Werkstück M14 von Phase 5.5
// (#1616).
//
// # Eine eigene Datei, und der Name ist der Unterschied
//
// **Nebenan steht [Keyword]** — das Schlüsselwort, das ein **Set** mitbringt
// (#1062):
// ein Eintrag im Bestand mit Name, Muster und Erinnerungstext. Hier steht die
// **Fähigkeit**, die eine Karte damit trägt (CR 702) — dasselbe Wort, zwei
// Gegenstände.
//
// # Der Schnitt ist gezählt und nicht gefühlt
//
// **CR 702 trägt 771 Regeln** — ein Viertel des Regelwerks. Welche zwölf
// dieses Programm liest, folgt einer **Zählung** über die zehn kanonischen
// Sets: 24.422 Karten mit Regeltext, und je Schlüsselwort, auf wie vielen es
// **steht**. Die Zahlen, ihre Stichprobe und ihre Grenzen stehen in
// `docs/SCHLUESSELWOERTER.md`; erhoben hat sie
// `internal/werkzeuge/schluesselwoerterzaehlen_test.go` hinter
// `fremdpakete`.
//
// **Elf der zwölf sind die häufigsten**, die dieses Programm auch
// durchsetzen kann. Der zwölfte ist `indestructible` (75 Karten) — er steht
// dabei, weil er eine **zustandsbasierte Aktion** aufhebt (CR 704.5g), und
// eine Regel, die für ihn nicht gälte, wäre falsch statt unvollständig.
//
// # Und warum die Zeile zählt und nicht das Wort
//
// **CR 702.1**: ein Objekt nennt nur den **Namen** der Fähigkeit, und ein
// Erinnerungstext fasst die Regel zusammen. Sie steht deshalb auf einer eigenen Zeile,
// mehrere durch Kommas getrennt. **CR 207.2a** erlaubt dahinter
// Erinnerungstext in Klammern.
//
// **`M6` las als Teilzeichenkette**, und der Preis stand als Grenze in der
// Matrix. Jetzt ist er **gemessen**: bei `flying` 3.349 gegen 2.315 Karten —
// 45 Prozent zu viel. „Kann nicht von Kreaturen ohne Flugfähigkeit geblockt
// werden" nennt das Wort und gibt es niemandem.
//
// # Was dieses Lesen weiterhin nicht kann
//
// **Es liest den gedruckten Text und nichts sonst.** Ein Effekt, der einer
// Kreatur Flugfähigkeit **gibt**, steht in einem anderen Kartentext und ist
// `M8`/`M10`. Und es liest **englisch**, so wie es im Paket steht (§9); ein
// deutsches Paket bringt das Wort in seiner Sprache mit. Beides steht als
// Grenze in der Regelmatrix.

import "strings"

// Die zwölf Schlüsselwörter, die dieses Programm liest — CR 702.
const (
	// SchluesselFlug — CR 702.9, auf 2.315 der 24.422 Karten.
	SchluesselFlug = "flying"

	// SchluesselTrampelschaden — CR 702.19, auf 622.
	SchluesselTrampelschaden = "trample"

	// SchluesselEile — CR 702.10, auf 432.
	SchluesselEile = "haste"

	// SchluesselWachsamkeit — CR 702.20, auf 358.
	SchluesselWachsamkeit = "vigilance"

	// SchluesselErstschlag — CR 702.7, auf 288.
	SchluesselErstschlag = "first strike"

	// SchluesselVerteidiger — CR 702.3, auf 268.
	SchluesselVerteidiger = "defender"

	// SchluesselReichweite — CR 702.17, auf 207.
	SchluesselReichweite = "reach"

	// SchluesselLebensverknuepfung — CR 702.15, auf 198.
	SchluesselLebensverknuepfung = "lifelink"

	// SchluesselTodesberuehrung — CR 702.2, auf 185.
	SchluesselTodesberuehrung = "deathtouch"

	// SchluesselBedrohlich — CR 702.111, auf 165.
	SchluesselBedrohlich = "menace"

	// SchluesselDoppelschlag — CR 702.4, auf 75.
	SchluesselDoppelschlag = "double strike"

	// SchluesselUnzerstoerbar — CR 702.12, auf 75.
	SchluesselUnzerstoerbar = "indestructible"
)

// Schluesselwoerter gibt die zwölf, nach Häufigkeit.
//
// **Exportiert, weil ein Wächter sie zählt** und weil `docs/` sie nennt —
// eine Liste im Code und eine in der Datei liefen auseinander (§9).
func Schluesselwoerter() []string {
	return []string{
		SchluesselFlug, SchluesselTrampelschaden, SchluesselEile,
		SchluesselWachsamkeit, SchluesselErstschlag, SchluesselVerteidiger,
		SchluesselReichweite, SchluesselLebensverknuepfung,
		SchluesselTodesberuehrung, SchluesselBedrohlich,
		SchluesselDoppelschlag, SchluesselUnzerstoerbar,
	}
}

// SchluesselwoerterAus liest, welche Schlüsselwörter ein Regeltext gibt.
//
// **Exportiert, weil die Zählung sie braucht** — dieselbe Lesart für die
// Messung und für das Spiel, sonst misst die eine etwas anderes, als die
// andere entscheidet (§10).
func SchluesselwoerterAus(text string) map[string]bool {
	woerter := map[string]bool{}

	for _, zeile := range strings.Split(text, "\n") {
		for _, wort := range schluesselzeile(zeile) {
			woerter[wort] = true
		}
	}

	return woerter
}

// schluesselzeile gibt die Schlüsselwörter einer Zeile — oder nichts.
//
// **Alles oder nichts**: steht in der Aufzählung etwas, das kein bekanntes
// Schlüsselwort ist, ist die Zeile ein **Satz** und keine Aufzählung. Genau
// daran scheitert die Teilzeichenkette.
//
// **Was das kostet, gehört dazu** (§10): eine Zeile mit einem Schlüsselwort,
// das dieses Programm **nicht** führt — „Flying, ward {2}" — gibt gar
// nichts. Das ist die richtige Richtung: lieber eine Fähigkeit zu wenig als
// eine, die niemand hat.
func schluesselzeile(zeile string) []string {
	ohne := ohneEingeklammertes(strings.ToLower(zeile), '(', ')')

	// **Die Auszeichnungen des Pakets fallen weg** (§8): MSE schreibt
	// `<i-flavor>` und `<soft-line>` in denselben Text.
	ohne = ohneEingeklammertes(ohne, '<', '>')

	teile := strings.Split(ohne, ",")
	woerter := make([]string, 0, len(teile))

	for _, teil := range teile {
		teil = strings.TrimSpace(strings.Trim(strings.TrimSpace(teil), "."))
		if teil == "" {
			continue
		}

		if !istSchluesselwort(teil) {
			return nil
		}

		woerter = append(woerter, teil)
	}

	return woerter
}

// istSchluesselwort sagt, ob dieses Wort eines der zwölf ist.
func istSchluesselwort(teil string) bool {
	for _, wort := range Schluesselwoerter() {
		if teil == wort {
			return true
		}
	}

	return false
}

// ohneEingeklammertes schneidet alles zwischen zwei Zeichen weg.
func ohneEingeklammertes(zeile string, auf, zu rune) string {
	var gebaut strings.Builder

	tiefe := 0

	for _, zeichen := range zeile {
		switch {
		case zeichen == auf:
			tiefe++
		case zeichen == zu && tiefe > 0:
			tiefe--
		case tiefe == 0:
			gebaut.WriteRune(zeichen)
		}
	}

	return gebaut.String()
}

// hatSchluesselwort sagt, ob dieses Objekt die Fähigkeit **gedruckt** trägt.
//
// **Nicht die Frage, die eine Regel stellt** — das ist seit `M10`
// (#1612)
// [Partie.HatFaehigkeit]: sie nimmt die sechste Schicht dazu (CR 613.1f).
// Diese hier liest den Text, und sonst nichts.
func (p Partie) hatSchluesselwort(kennung, wort string) bool {
	return SchluesselwoerterAus(p.regeltext(kennung))[wort]
}
