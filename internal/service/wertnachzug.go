package service

// **Was MSE beim Bearbeiten nachzieht** — #2396
// (Issue (#2396)).
//
// # Was MSE tut, nachgelesen
//
// Eine Änderung an einem Kartenwert ist eine `ValueAction`, und
// `SetScriptManager::onAction` ruft `updateValue` (`src/script/script_manager.cpp`,
// Zeilen 142–147):
//
//  1. Der geänderte Wert rechnet sein Skript — `value` ist, was eingegeben
//     wurde (`TextValue::update`, `src/data/field/text.cpp`, Zeilen 174–183).
//  2. `alsoUpdate` reiht die Felder ein, deren Skript ihn liest
//     (`dependent_scripts`, Zeilen 343–399).
//  3. `updateRecursive` rechnet sie der Reihe nach — **jedes höchstens einmal
//     je Runde**, weil die Runde mit einem Alter beginnt, das jedes `update`
//     überholt — und reiht die Leser eines Felds **nur ein, wenn es sich
//     geändert hat** (Zeilen 309–341).
//
// Gespeichert wird danach, was im Speicher steht: der **gerechnete** Wert.
//
// Fassung f9d9356 vom 2021-07-06.
//
// # Was hier anders ist (adapted)
//
// **Wer wen liest, steht hier nicht in einem Graphen, den eine Auswertung
// erhebt**, sondern im Quelltext: `card.<feld>` in einem Wertskript
// ([kartenzugriff], dieselbe Suche, nach der [mitWertskripten] `card` neu
// baut). MSE erhebt `dependent_scripts` durch eine abstrakte Auswertung jedes
// Skripts (`Context::dependencies`, `src/script/dependency.cpp`) und folgt
// dabei auch den Funktionen der Initskripte. Was dadurch hier fehlt — ein
// Feld, das ein anderes nur über eine solche Funktion liest —, bleibt
// gespeichert, wie es war, und wird beim Zeichnen trotzdem richtig gerechnet:
// [mitWertskripten] rechnet jedes Feld in der Reihenfolge des Spiels.

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// nachziehen rechnet die geänderten Felder und, wer sie liest — siehe oben.
//
// `geaendert` nennt die Felder, wie die Einträge sie führen (Unterstriche
// statt Leerzeichen). Zurück kommen die Einträge und **die Namen der Felder,
// die gerechnet wurden und sich geändert haben** — die geänderten selbst
// immer, und zuerst.
func nachziehen(
	ctx context.Context, gefuellt []msescript.Eintrag, felder []domain.Field,
	umgebung *msescript.Umgebung, vermerk *Schluesselwortverwendung,
	karteNeu func([]msescript.Eintrag), geaendert []string,
) (stand []msescript.Eintrag, gerechnet []string, befunde []domain.Befund) {
	skripte := make(map[string]string, len(felder))
	for i := range felder {
		if roh := strings.TrimSpace(felder[i].Skript); roh != "" {
			skripte[strings.ReplaceAll(felder[i].Name, " ", "_")] = roh
		}
	}

	leser := leserJeFeld(felder, skripte)

	stellen := make(map[string]int, len(gefuellt))
	for i := range gefuellt {
		stellen[gefuellt[i].Name] = i
	}

	schonGerechnet := make(map[string]bool, len(gefuellt))

	// rechne rechnet ein Feld und sagt, ob es sich geändert hat.
	rechne := func(name string) bool {
		schonGerechnet[name] = true

		stelle, da := stellen[name]
		roh, mitSkript := skripte[name]

		if !da || !mitSkript {
			return false
		}

		geaendert, befund := wertskriptRechnen(ctx, &gefuellt[stelle], roh, umgebung, vermerk)
		if befund != nil {
			befunde = append(befunde, *befund)

			return false
		}

		// **`card` sieht den neuen Wert**, bevor der nächste Leser rechnet —
		// derselbe Grund wie in [mitWertskripten].
		if geaendert {
			karteNeu(gefuellt)
		}

		return geaendert
	}

	var reihe []string

	for _, name := range geaendert {
		rechne(name)

		gerechnet = append(gerechnet, name)
		reihe = append(reihe, leser[name]...)
	}

	for len(reihe) > 0 && ctx.Err() == nil {
		name := reihe[0]
		reihe = reihe[1:]

		if schonGerechnet[name] {
			continue
		}

		if rechne(name) {
			gerechnet = append(gerechnet, name)
			reihe = append(reihe, leser[name]...)
		}
	}

	return gefuellt, gerechnet, entdoppelt(befunde)
}

// leserJeFeld gibt je Feld, wessen Wertskript es über `card` liest — **in der
// Reihenfolge des Spiels**, wie MSE `dependent_scripts` beim Durchgang über
// `game.card_fields` füllt (`SetScriptManager::initDependencies`,
// `src/script/script_manager.cpp`, Zeilen 103–110, Fassung f9d9356 vom
// 2021-07-06).
func leserJeFeld(felder []domain.Field, skripte map[string]string) map[string][]string {
	leser := make(map[string][]string)

	for i := range felder {
		name := strings.ReplaceAll(felder[i].Name, " ", "_")

		for _, treffer := range kartenzugriff.FindAllStringSubmatch(skripte[name], -1) {
			if !slices.Contains(leser[treffer[1]], name) {
				leser[treffer[1]] = append(leser[treffer[1]], name)
			}
		}
	}

	return leser
}

// nachgezogen baut die Umgebung der Karte, mit der der Bauer entstand, wie
// [umgebungsbauer.fuerKarte] — Vorgaben, Vorgabeskripte, `card` — und zieht
// statt aller Wertskripte nur die geänderten Felder und ihre Leser nach.
//
// **Nur für diese eine Karte**: der Schreibweg ändert eine, und die
// Vorbereitung für sie steht seit dem Bau schon da.
func (b *umgebungsbauer) nachgezogen(
	ctx context.Context, geaendert []string,
) ([]msescript.Eintrag, []string, []domain.Befund) {
	umgebung := b.setweit.Darueber()

	b.vermerk.Leere()

	felder, befunde := mitVorgabeskripten(ctx, b.attrappenfelder, b.vorlage.Kartenfelder, umgebung)
	felder = normalisiereMehrfachwahlen(felder, b.vorlage.Kartenfelder)

	karteNeu := func(stand []msescript.Eintrag) {
		umgebung.Setze("card", kartenabbildung(stand, b.vorlage.Kartenfelder, b.vorlage.Setpaket))
	}
	karteNeu(felder)

	felder, gerechnet, wertbefunde := nachziehen(ctx, felder, b.vorlage.Kartenfelder,
		umgebung, b.vermerk, karteNeu, geaendert)

	return felder, gerechnet, append(befunde, wertbefunde...)
}
