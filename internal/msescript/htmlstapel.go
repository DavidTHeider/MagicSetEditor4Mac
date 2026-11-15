package msescript

import "strings"

// **Der Auszeichnungsstapel von `to_html`** —
// #723.
//
// # Wofür es ihn gibt
//
// Ein Kartentext ist nicht wohlgeformt. `a<b>b<i>c</b>d</i>e` steht so in den
// Paketen, und HTML verträgt es nicht: `<b>` müsste vor `<i>` schließen. MSE
// sortiert deshalb um, statt durchzureichen.
//
//	class TagStack { … };
//
// `src/script/functions/export.cpp`, Zeilen 60–157, Fassung f9d9356 vom
// 2021-07-06.
//
// # Zwei Eigenheiten, die hier mitgenommen werden
//
// **Das Wiederöffnen findet nicht statt.** Der Kommentar im Original sagt, aus
// `<a><b><c>text` werde `<a><b><c>text</c></b><c>` — die Schleife dafür läuft
// über `tags`, und `tags` ist zu diesem Zeitpunkt schon verkürzt. Gemessen am
// Orakel (§15): `<b><i>x</b>y</i>` gibt `<b><i>x</i></b>y`, das `y` steht ohne
// Kursivierung da.
//
// **Ein Zähler darf unter null fallen.** Eine schließende Marke ohne offenen
// Stapel wird vorgemerkt und geschrieben; `--opened` geht dann auf -1, und die
// Marke erscheint nicht. Wer hier bei null abschnitte, bekäme ein `</b>` zu
// viel.
//
// Beides ist **belegtes Altverhalten**, kein Versehen dieser Umsetzung — und
// es zu „verbessern" wäre die stillschweigende Änderung, die §15 Satz 4
// verbietet.

// htmlmarke ist ein Paar aus öffnender und schließender HTML-Marke.
type htmlmarke struct {
	auf, zu string

	// offen zählt, wie oft die Marke in der **Eingabe** offen ist.
	offen int
}

// schreibe setzt die Marke, wenn sich die Schachtelungstiefe ändert.
func (m *htmlmarke) schreibe(gebaut *strings.Builder, schliessend bool) {
	if schliessend {
		m.offen--

		if m.offen == 0 {
			gebaut.WriteString(m.zu)
		}

		return
	}

	m.offen++

	if m.offen == 1 {
		gebaut.WriteString(m.auf)
	}
}

// vormerkung ist eine Marke samt der Richtung, in der sie wirkt.
type vormerkung struct {
	marke       *htmlmarke
	schliessend bool
}

// markenstapel hält, was offen ist, und was noch geschrieben werden soll.
type markenstapel struct {
	// offen sind die Marken, die in der **Ausgabe** offen stehen.
	offen []*htmlmarke

	// vorgemerkt sind Marken der Eingabe, die noch nicht in der Ausgabe stehen.
	//
	// **Vorgemerkt und nicht geschrieben**: erst wenn Text kommt, ist klar, ob
	// die Marke überhaupt etwas umschließt. `<b></b>` schreibt so gar nichts.
	vorgemerkt []vormerkung
}

func (s *markenstapel) oeffne(marke *htmlmarke) { s.trage(vormerkung{marke: marke}) }
func (s *markenstapel) schliesse(marke *htmlmarke) {
	s.trage(vormerkung{marke: marke, schliessend: true})
}

// schreibeVorgemerkte setzt alle Vormerkungen — vor jeder Ausgabe, die keine
// Marke ist.
func (s *markenstapel) schreibeVorgemerkte(gebaut *strings.Builder) {
	for _, eine := range s.vorgemerkt {
		eine.marke.schreibe(gebaut, eine.schliessend)

		if !eine.schliessend {
			s.offen = append(s.offen, eine.marke)
		}
	}

	s.vorgemerkt = s.vorgemerkt[:0]
}

// schliesseAlles beendet die Eingabe.
func (s *markenstapel) schliesseAlles(gebaut *strings.Builder) {
	s.schreibeVorgemerkte(gebaut)

	for stelle := len(s.offen) - 1; stelle >= 0; stelle-- {
		s.offen[stelle].schreibe(gebaut, true)
	}

	s.offen = s.offen[:0]
}

// trage nimmt eine Marke auf — `TagStack::add`.
func (s *markenstapel) trage(neue vormerkung) {
	if s.hebtVormerkungAuf(neue) {
		return
	}

	if neue.schliessend && s.schliesstOffene(neue) {
		return
	}

	s.vorgemerkt = append(s.vorgemerkt, neue)
}

// hebtVormerkungAuf sagt, ob die neue Marke eine Vormerkung aufhebt.
//
// **Von oben gesucht, und beim ersten Treffer derselben Marke ist Schluss** —
// gleich, ob er aufhebt oder nicht. `<b>a</b><b>b</b>` wird dadurch zu
// `<b>ab</b>`: die Schließung steht noch vor, die nächste Öffnung hebt sie auf.
func (s *markenstapel) hebtVormerkungAuf(neue vormerkung) bool {
	for stelle := len(s.vorgemerkt) - 1; stelle >= 0; stelle-- {
		if s.vorgemerkt[stelle].marke != neue.marke {
			continue
		}

		if s.vorgemerkt[stelle].schliessend == neue.schliessend {
			return false
		}

		s.vorgemerkt = append(s.vorgemerkt[:stelle], s.vorgemerkt[stelle+1:]...)

		return true
	}

	return false
}

// schliesstOffene schließt eine Marke, die nicht obenauf liegt.
//
// Alles darüber wird zuerst geschlossen — **und nicht wieder geöffnet**, siehe
// den Kopf dieser Datei.
func (s *markenstapel) schliesstOffene(neue vormerkung) bool {
	for stelle := len(s.offen) - 1; stelle >= 0; stelle-- {
		if s.offen[stelle] != neue.marke {
			continue
		}

		for hoeher := len(s.offen) - 1; hoeher > stelle; hoeher-- {
			s.vorgemerkt = append(s.vorgemerkt,
				vormerkung{marke: s.offen[hoeher], schliessend: true})
		}

		s.vorgemerkt = append(s.vorgemerkt, neue)
		s.offen = s.offen[:stelle]

		return true
	}

	return false
}
