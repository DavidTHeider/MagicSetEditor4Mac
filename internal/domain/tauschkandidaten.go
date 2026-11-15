package domain

// **Wer was von wem bekommen könnte** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Zwei Kennungen und kein Name
//
// **Ein genauer Wunsch trifft über den Druck, ein druckunabhängiger über das
// Konzept** — die Kennung, die mehrere Drucke derselben Karte teilen
// ([SemantikKonzept]). Der **Name** trifft nie: ein Alpha-Druck und ein
// Neudruck heißen gleich und sind nicht dasselbe Angebot, und wer aus der
// Namensgleichheit einen Treffer machte, böte jemandem etwas an, das er nie
// gewollt hat — es fiele erst am Tisch auf.
//
// # Warum eine fehlende Kennung ein Befund ist
//
// **Weil das Gegenteil eine Vermutung wäre** (§8). Eine Zeile ohne Druck kann
// einen genauen Wunsch nicht bedienen, und eine ohne Konzept keinen
// druckunabhängigen. Sie stillschweigend auszulassen hieße, einen möglichen
// Tausch zu verschweigen; sie zu raten hieße, einen falschen anzubieten.
//
// **Und ein Befund entsteht nur, wo er zählt.** Eine Zeile ohne Konzept neben
// lauter genauen Wünschen ist kein Befund: niemand hat danach gefragt. Ein
// Befund über alles ist ein Befund über nichts (§10).

import (
	"cmp"
	"slices"
)

// Angebotszeile ist, was einer hergeben kann — schon um Zusagen gekürzt.
//
// **Eine eigene Gestalt und nicht [Exemplar]**: dieselbe Rechnung läuft über
// den **eigenen** Bestand und über eine **freigegebene fremde Sicht**, und die
// fremde kennt weder Erwerbspreis noch Lagerort. Was hier steht, ist alles,
// was ein Abgleich braucht.
type Angebotszeile struct {
	// Druck ist die kanonische Druckidentität — leer heißt: keine.
	Druck string

	// Konzept ist die Kennung, die mehrere Drucke derselben Karte teilen —
	// leer heißt: keine.
	Konzept string

	Menge   int
	Zustand Zustand
	Sprache string
	Finish  Finish

	// Preis ist der Wert **je Stück** in der kleinsten Einheit — **0 heißt:
	// kein Preis**. Er kommt aus einer ausdrücklich gewählten Preisreihe
	// (A3) und wird hier nur getragen, nicht beschafft.
	Preis int64

	// Waehrung gehört zum Preis und gilt nur mit ihm — leer heißt: keine.
	Waehrung string

	Bezeichnung string
}

// Wunschzeile ist, was einer will.
type Wunschzeile struct {
	Druck   string
	Konzept string

	// Menge ist die **gewünschte** Zahl, nicht die vorhandene.
	Menge   int
	Zustand Zustand
	Sprache string
	Finish  Finish

	// Art sagt, ob genau dieser Druck gemeint ist — siehe [Wunschart].
	Art Wunschart

	// Rang ordnet die Wünsche untereinander — siehe [Wunschrang].
	Rang Wunschrang

	Bezeichnung string
}

// Tauschkandidat ist ein Treffer: was der eine hat und der andere will.
type Tauschkandidat struct {
	Wunsch  Wunschzeile
	Angebot Angebotszeile

	// Menge ist, wie viele Stücke dieser Treffer wirklich deckt.
	Menge int

	// Fehlend ist, wie viele dem Wunsch danach **noch** fehlen.
	//
	// **Sie steht dran, statt zu fehlen**: ein halb gedeckter Wunsch ist kein
	// erfüllter, und ein Vorschlag, der das verschweigt, verspricht mehr als
	// der Tisch hergibt.
	Fehlend int
}

// Tauschkandidaten sagt, welche Zeile welchen Wunsch bedienen könnte.
//
// **Ein Stück Papier liegt einmal da.** Eine Angebotszeile wird über alle
// Wünsche hinweg nur einmal vergeben; wer zuerst drankommt, entscheidet der
// Rang und danach eine feste Ordnung — nicht die Reihenfolge der Eingabe.
func Tauschkandidaten(
	angebot []Angebotszeile, wuensche []Wunschzeile,
) ([]Tauschkandidat, []Befund) {
	befunde := luecken(angebot, wuensche)

	zeilen := slices.Clone(angebot)
	slices.SortStableFunc(zeilen, func(links, rechts Angebotszeile) int {
		return cmp.Or(
			cmp.Compare(links.Druck, rechts.Druck),
			cmp.Compare(links.Bezeichnung, rechts.Bezeichnung),
		)
	})

	rest := make([]int, len(zeilen))
	for stelle, zeile := range zeilen {
		rest[stelle] = zeile.Menge
	}

	verzeichnis := verzeichnen(zeilen)
	treffer := make([]Tauschkandidat, 0, len(wuensche))

	for _, wunsch := range nachRang(wuensche) {
		treffer = append(treffer,
			fuerEinenWunsch(wunsch, zeilen, rest, verzeichnis)...)
	}

	return treffer, befunde
}

// tauschverzeichnis sagt, welche Zeilen für einen Schlüssel in Frage kommen.
//
// **Zwei Verzeichnisse und nicht eines**, weil es zwei Arten von Treffern
// gibt: über den Druck und über das Konzept. Ein gemeinsames müsste bei jeder
// Frage beide Schlüssel bilden und die Treffer wieder auseinandersortieren.
type tauschverzeichnis struct {
	nachDruck   map[string][]int
	nachKonzept map[string][]int
}

// verzeichnen legt die Zeilen unter ihre Schlüssel.
//
// **Einmal über die Angebote und danach je Wunsch ein Nachschlag.** Der erste
// Anlauf verglich jeden Wunsch mit jeder Zeile; gemessen am 2026-09-18 kostete
// das bei 20.000 gegen 20.000 **1,9 Sekunden** und wuchs mit dem Produkt —
// viermal so viele Zeilen, vierzehnmal so lange. Das Issue schließt genau das
// aus, und die Messung steht in `tauschmessung_test.go`.
//
// **Zeilen ohne Kennung stehen nicht darin.** Sie können nicht treffen, und
// dass sie es nicht können, sagt der Befund nebenan.
func verzeichnen(zeilen []Angebotszeile) tauschverzeichnis {
	verzeichnis := tauschverzeichnis{
		nachDruck:   make(map[string][]int, len(zeilen)),
		nachKonzept: make(map[string][]int, len(zeilen)),
	}

	for stelle, zeile := range zeilen {
		gestalt := gestaltVon(zeile.Zustand, zeile.Sprache, zeile.Finish)

		if zeile.Druck != "" {
			schluessel := zeile.Druck + gestalt
			verzeichnis.nachDruck[schluessel] = append(verzeichnis.nachDruck[schluessel], stelle)
		}

		if zeile.Konzept != "" {
			schluessel := zeile.Konzept + gestalt
			verzeichnis.nachKonzept[schluessel] = append(verzeichnis.nachKonzept[schluessel], stelle)
		}
	}

	return verzeichnis
}

// gestaltVon fasst zusammen, was neben der Kennung übereinstimmen muss.
//
// **Die Ausführung wird gelesen und nicht verglichen**: „leer" heißt
// `nonfoil`, und ein Vergleich der rohen Felder machte aus derselben Karte
// zwei (0119). Für den Zustand gilt dasselbe mit `nearMint`.
//
// **Der Nullbyte trennt die Teile**, damit „ab" + „c" und „a" + „bc" nicht
// denselben Schlüssel ergeben.
func gestaltVon(zustand Zustand, sprache string, finish Finish) string {
	return "\x00" + string(zustandOder(zustand)) +
		"\x00" + sprache +
		"\x00" + string(ausfuehrungOder(finish))
}

// stellenFuer sagt, welche Zeilen diesen Wunsch bedienen könnten.
//
// **Der Schlüssel ist die Regel.** Ein genauer Wunsch trifft über den Druck,
// ein druckunabhängiger über das Konzept — und in beiden Fällen müssen
// Zustand, Sprache und Ausführung übereinstimmen. Wer die Regel zusätzlich
// je Zeile prüfte, schriebe sie zweimal auf.
func (v tauschverzeichnis) stellenFuer(wunsch Wunschzeile) []int {
	gestalt := gestaltVon(wunsch.Zustand, wunsch.Sprache, wunsch.Finish)

	if wunsch.Art == WunschJederDruck {
		if wunsch.Konzept == "" {
			return nil
		}

		return v.nachKonzept[wunsch.Konzept+gestalt]
	}

	if wunsch.Druck == "" {
		return nil
	}

	return v.nachDruck[wunsch.Druck+gestalt]
}

// nachRang ordnet die Wünsche — **dringend zuerst**, dann fest.
//
// **Eine Kopie und keine Sortierung an Ort und Stelle.** Der Aufrufer hat
// seine Liste nicht zum Umstellen hergegeben, und eine Rechnung, die ihre
// Eingabe verändert, ist beim zweiten Ruf eine andere.
func nachRang(wuensche []Wunschzeile) []Wunschzeile {
	geordnet := slices.Clone(wuensche)

	slices.SortStableFunc(geordnet, func(links, rechts Wunschzeile) int {
		return cmp.Or(
			cmp.Compare(rechts.Rang.Gewicht(), links.Rang.Gewicht()),
			cmp.Compare(links.Bezeichnung, rechts.Bezeichnung),
			cmp.Compare(links.Druck, rechts.Druck),
		)
	})

	return geordnet
}

// fuerEinenWunsch nimmt aus dem Rest, was dieser Wunsch brauchen kann.
func fuerEinenWunsch(
	wunsch Wunschzeile, zeilen []Angebotszeile, rest []int,
	verzeichnis tauschverzeichnis,
) []Tauschkandidat {
	offen := wunsch.Menge
	gedeckt := []Tauschkandidat{}

	for _, stelle := range verzeichnis.stellenFuer(wunsch) {
		if offen <= 0 || rest[stelle] <= 0 {
			continue
		}

		nimm := min(offen, rest[stelle])
		offen -= nimm
		rest[stelle] -= nimm

		gedeckt = append(gedeckt, Tauschkandidat{
			Wunsch: wunsch, Angebot: zeilen[stelle], Menge: nimm,
		})
	}

	// **Der Rest steht an jedem Treffer dieses Wunsches.** Er ist eine
	// Aussage über den Wunsch und nicht über die einzelne Zeile — wer ihn nur
	// an die letzte schriebe, verlöre ihn beim Sortieren.
	for stelle := range gedeckt {
		gedeckt[stelle].Fehlend = offen
	}

	return gedeckt
}

// luecken sammelt die Kennungen, die für **diesen** Abgleich gefehlt haben.
func luecken(angebot []Angebotszeile, wuensche []Wunschzeile) []Befund {
	befunde := []Befund{}

	genau, jederDruck := false, false

	for _, wunsch := range wuensche {
		if wunsch.Art == WunschJederDruck {
			jederDruck = true

			if wunsch.Konzept == "" {
				befunde = append(befunde, fehltBefund(wunsch.Bezeichnung,
					"Dieser Wunsch gilt für jeden Druck, nennt aber keine Karte."))
			}

			continue
		}

		genau = true

		if wunsch.Druck == "" {
			befunde = append(befunde, fehltBefund(wunsch.Bezeichnung,
				"Dieser Wunsch gilt für genau einen Druck, nennt aber keinen."))
		}
	}

	for _, zeile := range angebot {
		if genau && zeile.Druck == "" {
			befunde = append(befunde, fehltBefund(zeile.Bezeichnung,
				"Diese Zeile nennt keinen Druck und kann keinen genauen Wunsch bedienen."))
		}

		if jederDruck && zeile.Konzept == "" {
			befunde = append(befunde, fehltBefund(zeile.Bezeichnung,
				"Diese Zeile nennt keine Karte und kann keinen druckunabhängigen Wunsch bedienen."))
		}
	}

	return befunde
}

// fehltBefund baut den Befund über eine fehlende Kennung.
//
// **[ArtEingabe] und nicht der Nullwert**: [ArtPortierung] sagte, dieses
// Programm könne etwas nicht, was MSE kann — und MSE handelt gar nicht.
func fehltBefund(ort, meldung string) Befund {
	return Befund{Ort: ort, Meldung: meldung, Art: ArtEingabe}
}

// ausfuehrungOder liest eine Ausführung — leer heißt `nonfoil`.
func ausfuehrungOder(finish Finish) Finish {
	if finish == "" {
		return FinishNonfoil
	}

	return finish
}

// zustandOder liest einen Zustand — leer heißt `nearMint`.
func zustandOder(zustand Zustand) Zustand {
	if zustand == "" {
		return ZustandNearMint
	}

	return zustand
}
