package domain

// **Was beide hergeben und bekommen** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Warum „möglichst gleicher Wert" allein nicht geht
//
// **Weil diese Formel ihr Optimum bei null zu null hat.** Wer nur den
// Wertunterschied minimiert, bekommt den leeren Tausch empfohlen — er ist
// perfekt ausgeglichen und nutzlos. Vor dem Wert steht deshalb eine
// **Nichtleer-Bedingung**: ohne ein Stück auf jeder Seite gibt es keinen
// Vorschlag.
//
// # Und warum der Wert erst danach kommt
//
// **Weil ein Tausch Wünsche deckt und nicht Beträge ausgleicht.** Geordnet
// wird nach Dringlichkeit; der Wert ist eine **Auskunft** über den fertigen
// Vorschlag und keine Zielgröße. Wer ihn zur Zielgröße machte, böte teure
// Karten an, die niemand haben will.
//
// # Zwei Währungen werden nicht addiert
//
// **Eine Summe über EUR und USD ist eine Zahl, die es nicht gibt.** Fehlt an
// einer Zeile der Preis oder steht dort eine zweite Währung, wiegt die Seite
// gar nicht — sie wiegt nicht falsch.

import (
	"cmp"
	"slices"
)

// VorgabeHoechstzahl ist der Deckel je Seite, wenn keiner genannt ist.
//
// **Ein Deckel und keine Suche über alles.** Der Vorschlag ist begrenzt, weil
// er auf Tastendruck entsteht und weil zwanzig Karten je Seite mehr sind, als
// an einem Tisch je den Besitzer wechseln.
const VorgabeHoechstzahl = 20

// Tauschumfang begrenzt, was ein Vorschlag umfassen darf.
type Tauschumfang struct {
	// Hoechstzahl ist die Zahl der Stücke je Seite — **0 heißt die Vorgabe**.
	Hoechstzahl int
}

// Tauschseite ist, was eine Seite gibt — samt dem, was offen bleibt.
type Tauschseite struct {
	Zeilen []Tauschkandidat

	// Stueck ist die Zahl der Stücke auf dieser Seite.
	Stueck int

	// Gedeckt ist, wie viele gewünschte Stücke dieser Vorschlag deckt.
	Gedeckt int

	// Offen ist, wie viele danach **noch** offen sind — der Rest der
	// bedienten Wünsche und das, was der Deckel abgeschnitten hat.
	Offen int

	// Wert gilt nur mit [Tauschseite.Bepreist] — sonst ist er 0.
	Wert int64

	// Waehrung gehört zum Wert und gilt nur mit ihm.
	Waehrung string

	// Bepreist sagt, ob **jede** Zeile einen Preis in **derselben** Währung
	// trug. Sonst wiegt diese Seite nicht.
	Bepreist bool
}

// Tauschvorschlag ist ein Vorschlag für beide Seiten.
type Tauschvorschlag struct {
	// Gebe ist, was dieses Profil hergibt.
	Gebe Tauschseite

	// Bekomme ist, was es dafür bekommt.
	Bekomme Tauschseite
}

// Tauschvorschlagen baut aus den Kandidaten beider Richtungen einen Vorschlag.
//
// **Und gibt `false`, wenn eine Seite leer bliebe.** Ein Tausch, bei dem einer
// nichts gibt, ist ein Geschenk; dass er im Wert perfekt ausgeglichen wäre,
// macht ihn nicht zum Vorschlag.
func Tauschvorschlagen(
	gebe, bekomme []Tauschkandidat, umfang Tauschumfang,
) (Tauschvorschlag, bool) {
	hoechst := umfang.Hoechstzahl
	if hoechst <= 0 {
		hoechst = VorgabeHoechstzahl
	}

	meine, fremde := seiteBauen(gebe, hoechst), seiteBauen(bekomme, hoechst)

	if meine.Stueck == 0 || fremde.Stueck == 0 {
		return Tauschvorschlag{}, false
	}

	return Tauschvorschlag{Gebe: meine, Bekomme: fremde}, true
}

// seiteBauen nimmt vom Dringendsten so viel, wie der Deckel zulässt.
func seiteBauen(kandidaten []Tauschkandidat, hoechst int) Tauschseite {
	geordnet := slices.Clone(kandidaten)

	slices.SortStableFunc(geordnet, func(links, rechts Tauschkandidat) int {
		return cmp.Or(
			cmp.Compare(rechts.Wunsch.Rang.Gewicht(), links.Wunsch.Rang.Gewicht()),
			cmp.Compare(links.Wunsch.Bezeichnung, rechts.Wunsch.Bezeichnung),
			cmp.Compare(links.Angebot.Druck, rechts.Angebot.Druck),
		)
	})

	seite := Tauschseite{Zeilen: make([]Tauschkandidat, 0, len(geordnet))}
	gezaehlt := map[Wunschzeile]bool{}

	for _, kandidat := range geordnet {
		frei := hoechst - seite.Stueck
		if frei <= 0 {
			// **Was der Deckel abschneidet, bleibt offen** und verschwindet
			// nicht: ein Vorschlag sagt auch, was er nicht deckt.
			seite.Offen += kandidat.Menge

			continue
		}

		genommen := kandidat
		genommen.Menge = min(kandidat.Menge, frei)

		seite.Offen += kandidat.Menge - genommen.Menge
		seite.Stueck += genommen.Menge
		seite.Gedeckt += genommen.Menge
		seite.Zeilen = append(seite.Zeilen, genommen)

		// **Der Rest eines Wunsches zählt einmal.** Er steht an jedem
		// Treffer dieses Wunsches; wer ihn je Zeile addierte, zählte ihn
		// doppelt, sobald zwei Zeilen denselben Wunsch bedienen.
		if !gezaehlt[kandidat.Wunsch] {
			gezaehlt[kandidat.Wunsch] = true
			seite.Offen += kandidat.Fehlend
		}
	}

	seite.wiegen()

	return seite
}

// wiegen rechnet den Wert — oder sagt, dass diese Seite nicht wiegt.
func (s *Tauschseite) wiegen() {
	if len(s.Zeilen) == 0 {
		return
	}

	waehrung := s.Zeilen[0].Angebot.Waehrung
	summe := int64(0)

	for _, zeile := range s.Zeilen {
		if zeile.Angebot.Preis <= 0 || zeile.Angebot.Waehrung != waehrung {
			return
		}

		summe += zeile.Angebot.Preis * int64(zeile.Menge)
	}

	if waehrung == "" {
		return
	}

	s.Wert, s.Waehrung, s.Bepreist = summe, waehrung, true
}
