// **Die Werkzeuge, die eine ganze Datei anfassen** — die Sicherung des
// Bestands (#1169)
// und der portable Umschlag
// (#1457).
//
// **Eine eigene Datei, weil `werkzeugliste.go` sonst über die Zeilengrenze
// aus §2 ginge** — und der Schnitt folgt der Frage: die beiden gehören
// zusammen, weil sie als einzige nicht an einer Karte oder einem Set hängen,
// sondern an einer **Datei auf der Platte**.

package mcp

// werkzeugeAnGanzenDateien gibt beide Gruppen in einem Zug.
func werkzeugeAnGanzenDateien(naehte Naehte) (lesend, schreibend []Werkzeug) {
	sicherunglesend, sicherungschreibend := sicherungswerkzeuge(naehte.Sicherung)
	umschlaglesend, umschlagschreibend := umschlagwerkzeuge(naehte.Umschlag)

	// **Die Partien stehen hier, obwohl sie an keiner Datei hängen** —
	// und das ist eine Abweichung mit Grund: `werkzeugliste.go` steht an der
	// Zeilengrenze aus §2, und eine dritte Datei für vier Werkzeuge wäre
	// eine Datei mehr für dieselbe Frage. Wer hier etwas sucht, sucht „was
	// hängt an keiner Karte und keinem Set".
	partienlesend, partienschreibend := partienwerkzeuge(naehte.Partien)

	// **Und der Verlauf** — seit Werkstück G5 von Phase 5.5
	// (#1462).
	// Er hängt an derselben Frage wie die Partien und steht deshalb hier.
	partienlesend = append(partienlesend, partieverlaufswerkzeug(naehte.Partienverlauf)...)

	// **Und die Ansicht einer Datei** (UX-13, #1685) — sie hängt an einer
	// Datei auf der Platte wie die Sicherung.
	sicherunglesend = append(sicherunglesend, dateiwerkzeuge(naehte.Dateien)...)

	// **Und die Vorschau vor dem Einfahren** (FMT-03, #1692) — sie liest eine
	// Setdatei, ohne zu schreiben.
	sicherunglesend = append(sicherunglesend, setvorschauwerkzeuge(naehte.Setvorschau)...)

	// **Und die Marken an jedem Fachobjekt** (UX-12, #1684) — sie hängen an
	// keiner Art allein.
	objektlesend, objektschreibend := objektmarkenwerkzeuge(naehte.Objektmarken)
	sicherunglesend = append(sicherunglesend, objektlesend...)
	sicherungschreibend = append(sicherungschreibend, objektschreibend...)

	// **Und die Symbolbibliothek** (DOM-02, #1697) — sie gehört dem Profil
	// und hängt an keiner Karte und keinem Set.
	bibliothekslesend, bibliotheksschreibend := symbolbibliothekswerkzeuge(naehte.Symbolbibliothek)
	sicherunglesend = append(sicherunglesend, bibliothekslesend...)
	sicherungschreibend = append(sicherungschreibend, bibliotheksschreibend...)

	// **Und die eigenen Übersetzungen** (LOC-02, #1688) — sie gehören dem
	// Rechner und hängen an keinem Profil.
	sprachlesend, sprachschreibend := sprachwerkzeuge(naehte.Sprachpakete)
	sicherunglesend = append(sicherunglesend, sprachlesend...)
	sicherungschreibend = append(sicherungschreibend, sprachschreibend...)

	return append(append(sicherunglesend, umschlaglesend...), partienlesend...),
		append(append(sicherungschreibend, umschlagschreibend...),
			partienschreibend...)
}
