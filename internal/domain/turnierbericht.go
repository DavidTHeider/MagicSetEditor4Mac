package domain

// **Der Abschlussbericht eines Turniers** — DOM-04.AK-03
// (#1699),
// Vermerk [0140](../../docs/entscheidungen/0140-ein-turnier-als-datei.md).
//
// # Warum Markdown
//
// **Weil ein Bericht gelesen und weitergegeben wird**, nicht wieder
// eingelesen — dafür gibt es die Turnierdatei. Markdown liest sich als Text,
// in jedem Editor und in einem Chat, und wird in einer Vorschau zu Tabellen.
//
// # Warum die Wörter von außen kommen
//
// **Weil ein Bericht Menschentext ist** (LOKALISIERUNG.md): er steht in der
// Sprache des Fensters, und die Domäne kennt keine. Die Zahlen stehen wie in
// der Rangliste des Fensters — Siege samt Byes, Quoten auf vier Stellen —,
// damit der Bericht nicht etwas anderes sagt als der Bildschirm.
//
// Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

import (
	"fmt"
	"strings"
)

// Berichttexte sind die Wörter eines Berichts.
type Berichttexte struct {
	Spiel, Format, Stand, Abgeschlossen, Laufend, Teilnehmende, Ausgestiegen, Runden string

	Rangliste, Platz, Name, Punkte, Bilanz, Gegnermatchquote, Spielquote, Gegnerspielquote string

	Runde, Tisch, SeiteA, SeiteB, Ergebnis, Bye, Offen string
}

// Turnierbericht schreibt den Bericht über ein Turnier — Kopf, Rangliste und
// jede Runde. **Auch über ein laufendes**: dann ist es ein Zwischenstand, und
// der Kopf sagt es.
func Turnierbericht(turnier Turnier, format string, texte Berichttexte) string {
	var bericht strings.Builder

	namen := make(map[int64]string, len(turnier.Teilnehmer))
	raus := 0

	for _, teilnehmer := range turnier.Teilnehmer {
		namen[teilnehmer.ID] = zelle(teilnehmer.Name)

		if teilnehmer.Ausgestiegen {
			namen[teilnehmer.ID] += " (" + texte.Ausgestiegen + ")"
			raus++
		}
	}

	fmt.Fprintf(&bericht, "# %s\n\n- %s: %s\n", turnier.Name, texte.Spiel, turnier.Spiel)

	if format != "" {
		fmt.Fprintf(&bericht, "- %s: %s\n", texte.Format, format)
	}

	stand := texte.Laufend
	if turnier.IstAbgeschlossen() {
		// **Der Tag, nicht die Sekunde** — und ohne Annahme über die Länge.
		tag, _, _ := strings.Cut(turnier.Abgeschlossen, "T")
		stand = texte.Abgeschlossen + " " + tag
	}

	fmt.Fprintf(&bericht, "- %s: %s\n- %s: %d", texte.Stand, stand, texte.Teilnehmende, len(turnier.Teilnehmer))

	if raus > 0 {
		fmt.Fprintf(&bericht, " (%s: %d)", texte.Ausgestiegen, raus)
	}

	fmt.Fprintf(&bericht, "\n- %s: %d\n", texte.Runden, len(turnier.Runden))

	ranglisteSchreiben(&bericht, turnier, namen, texte)

	for _, runde := range turnier.Runden {
		fmt.Fprintf(&bericht, "\n## %s %d\n\n| %s | %s | %s | %s |\n| ---: | --- | --- | --- |\n",
			texte.Runde, runde.Nummer, texte.Tisch, texte.SeiteA, texte.SeiteB, texte.Ergebnis)

		for i, paarung := range runde.Paarungen {
			fmt.Fprintf(&bericht, "| %d | %s | %s | %s |\n",
				i+1, namen[paarung.A], gegenseite(paarung, namen, texte), standVon(paarung, texte))
		}
	}

	return bericht.String()
}

// ranglisteSchreiben schreibt die Tabelle der Ränge — **ohne** Spiel- und
// Gegnerspielquote bei einem Einzelspielturnier: dort sind sie Kopien der
// Matchzahlen (MTR §3.1, Fassung 2026-02-27; Vermerk 0062, Entscheidung 5),
// wie im Fenster.
func ranglisteSchreiben(bericht *strings.Builder, turnier Turnier, namen map[int64]string, texte Berichttexte) {
	fmt.Fprintf(bericht, "\n## %s\n\n| %s | %s | %s | %s | %s |", texte.Rangliste,
		texte.Platz, texte.Name, texte.Punkte, texte.Bilanz, texte.Gegnermatchquote)

	trenner := "\n| ---: | --- | ---: | --- | ---: |"
	if !turnier.Einzelspiel {
		fmt.Fprintf(bericht, " %s | %s |", texte.Spielquote, texte.Gegnerspielquote)
		trenner += " ---: | ---: |"
	}

	bericht.WriteString(trenner + "\n")

	for _, rang := range Rangliste(turnier) {
		fmt.Fprintf(bericht, "| %d | %s | %d | %d–%d–%d | %.4f |", rang.Platz, namen[rang.TeilnehmerID],
			rang.Matchpunkte, rang.Bilanz.Siege+rang.Bilanz.Byes, rang.Bilanz.Niederlagen,
			rang.Bilanz.Unentschieden, rang.Gegnermatchquote)

		if !turnier.Einzelspiel {
			fmt.Fprintf(bericht, " %.4f | %.4f |", rang.Spielquote, rang.Gegnerspielquote)
		}

		bericht.WriteString("\n")
	}
}

// gegenseite nennt die zweite Seite einer Paarung — oder den Bye.
func gegenseite(paarung Paarung, namen map[int64]string, texte Berichttexte) string {
	if paarung.IstBye() {
		return texte.Bye
	}

	return namen[paarung.B]
}

// standVon schreibt den Spielstand — **a–b**, und mit Unentschieden a–b–u.
func standVon(paarung Paarung, texte Berichttexte) string {
	switch {
	case paarung.IstBye():
		return "—"
	case paarung.Ergebnis == nil:
		return texte.Offen
	case paarung.Ergebnis.Unentschieden > 0:
		return fmt.Sprintf("%d–%d–%d", paarung.Ergebnis.SpieleA, paarung.Ergebnis.SpieleB, paarung.Ergebnis.Unentschieden)
	}

	return fmt.Sprintf("%d–%d", paarung.Ergebnis.SpieleA, paarung.Ergebnis.SpieleB)
}

// zelle macht einen Namen tabellenfest — ein senkrechter Strich trennte
// sonst eine Spalte, und ein Zeilenumbruch eine Zeile.
func zelle(text string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ", "\r", " ").Replace(text)
}
