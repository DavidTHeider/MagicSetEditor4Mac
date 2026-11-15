// **Die Roadmap führt nur Offenes** — §9, und
// #1169,
// Befund 7 des Zweitblicks.
//
// # Der Befund
//
// `docs/ROADMAP.md` erklärte Phase 3.8 im selben Dokument für **abgeschlossen**
// und führte sie zugleich in der Tabelle, die laut §9 nur Offenes führt. Ein
// Leser konnte aus derselben Datei beides ableiten.
//
// **Es ist zweimal passiert.** Beim ersten Mal war es Phase 3.4, und die
// Chronik hat sie nachgetragen; beim zweiten Mal war es 3.8, und der Grund war
// banal: die Zeile sollte entfernt werden, und das Muster traf nicht, weil
// `prettier` die Tabellenspalten inzwischen ausgerichtet hatte. Ein Fehler,
// der zweimal passiert, ist keine Unachtsamkeit mehr.
//
// # Was er hält
//
// Keine Phase, die im Fließtext als **abgeschlossen** genannt wird, steht noch
// in der Phasentabelle.
//
// # Was er nicht hält
//
// **Ob eine Phase, die dort steht, wirklich offen ist.** Er sieht nur den
// Widerspruch, nicht das Versäumnis: eine fertige Phase, die niemand irgendwo
// abgeschlossen nennt, bleibt unbemerkt in der Tabelle stehen. Was sie hält,
// ist der Phasenabschluss nach §13 — und der ist Handwerk.
//
// # Warum die Leserprobe an einer Probe hängt und nicht am Dokument
//
// Hier stand `t.Fatal("die Roadmap nennt keine abgeschlossene Phase")`, und der
// Gedanke war richtig: ein Wächter, der nichts findet, hält keine Zusage. Am
// 2026-08-29 wurde die Roadmap von den Berichten der Freigabe 0.3.0 befreit —
// §9, sie führt **nur Offenes** — und der Wächter wurde rot, obwohl das
// Dokument zum ersten Mal genau das tat, was die Regel verlangt.
//
// **Ein Anhalter, der eine Dokumentgestalt erzwingt, prüft die Gestalt und
// nicht den Leser.** Beide Leser haben deshalb eine eigene Probe: sie zeigt an
// gebautem Text, dass sie greifen — unabhängig davon, was heute in der Datei
// steht. Findet der eine im Dokument nichts, ist das eine Aussage über das
// Dokument und keine über ihn
// (#1287).
package werkzeuge_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

// abgeschlosseneImText findet jede Form, in der eine Phase für fertig erklärt
// wird — mit und ohne Fettschrift, mit und ohne Einschub, einzeln und im Paar.
//
// **Der erste Anlauf kannte genau eine Form**: `**Phase X ist abgeschlossen**`.
// „Phase X ist damit abgeschlossen", „Die Phasen X und Y sind abgeschlossen"
// oder eine anders gesetzte Zelle gingen daran vorbei — angestrichen vom
// Nachlauf über Phase 3.8, Befund 7
// (#1253).
//
// **Der Einschub war einmal eine Wortzahl**, und die war zweimal falsch: erst
// drei Wörter, dann fünf, und `\w` ist in Go **ASCII** — „Durchgängen" war für
// den Wächter kein Wort. Beides hat die Probe unten gefunden, das Zweite erst
// nach der ersten Korrektur, und der Zweitblick über Phase 4.0 fand die
// nächste Länge sofort wieder (Befund 6). Eine gewürfelte Zahl in einem Regex
// ist dieselbe Behauptung wie eine gewürfelte Zahl im Fließtext.
//
// **Deshalb zählt jetzt kein Wort mehr, sondern eine Grenze.** Zwischen der
// Nummer und dem Wort „abgeschlossen" darf alles stehen, was **innerhalb
// eines Satzteils** bleibt: kein Punkt, kein Frage- oder Ausrufezeichen, kein
// Komma, kein Strichpunkt, kein Zellentrenner, kein Zeilenumbruch. So lang der
// Satzteil ist, so weit reicht der Wächter.
//
// **Und ein zweites „Phase" im Einschub beendet ihn** — geprüft in
// `abgeschlossenePhasen`, nicht im Muster: RE2 kennt keine Vorschau, und die
// Bedingung ist zu wichtig, um sie deshalb wegzulassen. Ohne sie läse „Phase
// 4.1 ist offen und Phase 3.8 ist abgeschlossen" die **falsche** Nummer.
//
// **Was er weiterhin nicht sieht**: eine Abschlussangabe, die das Wort
// „abgeschlossen" nicht benutzt — „Phase X ist durch", „Phase X liegt hinter
// uns". Ein Wächter über Bedeutungen wäre keiner. Und eine, die über einen
// Punkt oder ein Komma hinweggeht: „Phase 3.8, seit gestern, ist
// abgeschlossen". Das ist der Preis dafür, den Satzteil als Grenze zu nehmen,
// und er ist kleiner als eine gewürfelte Wortzahl.
var abgeschlosseneImText = regexp.MustCompile(
	`\*{0,2}(?:Die )?Phasen? \*{0,2}(\d+(?:\.\d+)?[a-z]?)\*{0,2}` +
		`(?: und \*{0,2}(\d+(?:\.\d+)?[a-z]?)\*{0,2})?` +
		`([^.!?|,;\n]{0,120}?) abgeschlossen`)

// abgeschlossenePhasen gibt die Nummern, die ein Text für fertig erklärt.
//
// **Der Text wird zuerst absatzweise verbunden.** Ein Markdown-Umbruch beendet
// weder Satz noch Satzteil — er steht dort, wo die Zeile zu lang wurde. „Phase
// 4.0 ist nach dem zweiten\nDurchgang abgeschlossen." ging deshalb lautlos
// durch, und die Roadmap ist auf 80 Zeichen umbrochen: die Gestalt ist nicht
// die Ausnahme, sondern der Normalfall (Nachlauf über Phase 4.0, Befund 6).
func abgeschlossenePhasen(text string) []string {
	var nummern []string

	for _, treffer := range abgeschlosseneImText.FindAllStringSubmatch(
		absatzweiseVerbunden(text), -1) {
		// **Der Einschub darf keine zweite Phase nennen.** Sonst gehörte das
		// „abgeschlossen" zu ihr und nicht zu der Nummer, die der Wächter
		// gerade gefangen hat.
		if strings.Contains(treffer[3], "Phase") {
			continue
		}

		for _, nummer := range treffer[1:3] {
			if nummer != "" {
				nummern = append(nummern, nummer)
			}
		}
	}

	return nummern
}

// absatzweiseVerbunden macht aus den Zeilen eines Absatzes eine.
//
// **Eine Leerzeile trennt weiter**, damit ein Satz nicht in den nächsten
// Absatz hineinliest. Und der Zellentrenner `|` steht in der Zeichenklasse des
// Musters, damit zwei Tabellenzellen nicht zu einem Satz verschmelzen.
func absatzweiseVerbunden(text string) string {
	var (
		gebaut  strings.Builder
		laufend []string
	)

	schliessen := func() {
		if len(laufend) > 0 {
			gebaut.WriteString(strings.Join(laufend, " "))
			gebaut.WriteString("\n\n")

			laufend = nil
		}
	}

	for _, zeile := range strings.Split(text, "\n") {
		if strings.TrimSpace(zeile) == "" {
			schliessen()

			continue
		}

		laufend = append(laufend, strings.TrimSpace(zeile))
	}

	schliessen()

	return gebaut.String()
}

// phaseInDerTabelle findet „| **3.8** |" am Zeilenanfang.
var phaseInDerTabelle = regexp.MustCompile(`(?m)^\|\s*\*\*(\d+(?:\.\d+)?[a-z]?)\*\*\s*\|`)

func TestKeineAbgeschlossenePhaseStehtNochInDerTabelle(t *testing.T) {
	roadmap := lies(t, "docs/ROADMAP.md")

	// **Beide Nummern je Fund**, denn »Die Phasen 3.6 und 3.7 sind
	// abgeschlossen« nennt zwei.
	abgeschlossen := map[string]bool{}
	for _, nummer := range abgeschlossenePhasen(roadmap) {
		abgeschlossen[nummer] = true
	}

	// **Keine Zahl ist hier ein zulässiger Befund.** Eine Roadmap, die nur
	// Offenes führt, nennt keine abgeschlossene Phase — dann gibt es keinen
	// Widerspruch, und das ist der Zielzustand. Dass die Leser trotzdem greifen,
	// zeigen die Proben unten.
	inDerTabelle := phaseInDerTabelle.FindAllStringSubmatch(roadmap, -1)

	for _, treffer := range inDerTabelle {
		if abgeschlossen[treffer[1]] {
			t.Errorf("Phase %s steht in der Phasentabelle und heißt im Text zugleich "+
				"abgeschlossen — die Roadmap führt nur Offenes (AGENTS.md §9)", treffer[1])
		}
	}

	// **Und die Chronik nennt jede von ihnen.** Eine Phase, die im Text
	// abgeschlossen heißt und nirgends steht, ist erledigte Arbeit, die
	// niemand wiederfindet — genau der Fall von Phase 3.4.
	chronik := lies(t, "docs/CHRONIK.md")

	for phase := range abgeschlossen {
		if !strings.Contains(chronik, "Phase "+phase+":") {
			t.Errorf("Phase %s heißt in der Roadmap abgeschlossen, steht aber in "+
				"keinem Chronikeintrag (AGENTS.md §9)", phase)
		}
	}
}

// **Die Leserproben.** Sie stehen hier, weil ein Anhalter am Dokument die
// falsche Frage stellte — siehe den Kopf dieser Datei.
func TestDieLeserDerRoadmapGreifenWirklich(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		text string
		will []string
	}{
		{"fett und mit Nachsatz", "**Phase 3.8 ist abgeschlossen** und steht in der Chronik.", []string{"3.8"}},
		{"Einschub mit einem Wort", "Phase 3.8b ist damit abgeschlossen.", []string{"3.8b"}},
		{"zwei Phasen in einem Satz", "Die Phasen 3.6 und 3.7 sind abgeschlossen.", []string{"3.6", "3.7"}},
		{"Einschub mit vier Wörtern", "Phase 4 ist nach zwei Durchgängen abgeschlossen.", []string{"4"}},
		{"offen ist nicht abgeschlossen", "Phase 4.1 ist offen.", nil},
		{
			"über den Punkt hinweg greift er nicht",
			"Phase 4.1 ist offen. Phase 3.8 ist abgeschlossen.",
			[]string{"3.8"},
		},
		{
			"langer Einschub — der Befund des Zweitblicks über Phase 4.0",
			"Phase 4.0 ist nun nach diesen insgesamt sechs Durchgängen abgeschlossen.",
			[]string{"4.0"},
		},
		{
			"über das Komma hinweg greift er nicht",
			"Phase 4.1 ist offen, Phase 3.8 ist abgeschlossen.",
			[]string{"3.8"},
		},
		{
			"ein zweites »Phase« im Einschub beendet ihn",
			"Phase 4.1 ist offen und Phase 3.8 ist abgeschlossen.",
			[]string{"3.8"},
		},
		{
			"über einen Markdown-Umbruch hinweg — der Befund des Nachlaufs",
			"Phase 4.0 ist nach dem zweiten\nDurchgang abgeschlossen.",
			[]string{"4.0"},
		},
		{
			"über eine Leerzeile hinweg greift er nicht",
			"Phase 4.0 ist offen\n\nDurchgang abgeschlossen.",
			nil,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			hat := abgeschlossenePhasen(fall.text)

			if !slices.Equal(hat, fall.will) {
				t.Errorf("%q: gelesen %v — erwartet %v", fall.text, hat, fall.will)
			}
		})
	}

	t.Run("Phase in der Tabelle", func(t *testing.T) {
		t.Parallel()

		const probe = "| **4.0** | Die Baseline |\n" +
			"|   **4.7**   | Turniere |\n" +
			"| **3.8b** | Der Sicherungsweg |\n" +
			"| Phase | Was |\n"

		treffer := phaseInDerTabelle.FindAllStringSubmatch(probe, -1)

		hat := make([]string, 0, len(treffer))
		for _, fund := range treffer {
			hat = append(hat, fund[1])
		}

		if will := []string{"4.0", "4.7", "3.8b"}; !slices.Equal(hat, will) {
			t.Errorf("gelesen %v — erwartet %v", hat, will)
		}
	})
}
