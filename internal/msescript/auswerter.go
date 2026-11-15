package msescript

import (
	"context"
	"strconv"
)

// Tiefengrenze begrenzt, wie tief eine Auswertung sich selbst rufen darf.
//
// **Eine Vorlage darf die Anwendung nicht beenden.** Fremde Daten sind der
// Normalfall (AGENTS.md §8), und `a := {a()}` gefolgt von `a()` ist in zwei
// Zeilen geschrieben. Ohne Grenze frisst das den Stapel, und Go beendet den
// Prozess — nicht als Fehler, den jemand anzeigen könnte, sondern sofort.
//
// Die Zahl ist großzügig: die tiefste gesichtete Verschachtelung ist eine
// else-Kette über rund 20 Stufen.
const Tiefengrenze = 200

// Auswerten rechnet einen Ausdruck zu einem Wert.
//
// Der `context` wird an jedem Knoten geprüft (AGENTS.md §7): eine Kartenvorschau
// wird **abgebrochen**, nicht abgewartet. Wer während des Tippens vier Läufe
// anstößt, zeigt sonst den drittletzten Stand.
func Auswerten(ctx context.Context, knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	if umgebung == nil {
		umgebung = NeueUmgebung(nil)
	}

	a := &auswerter{ctx: ctx}

	return a.werte(knoten, umgebung.Darueber())
}

// AuswertenMitVorbehalten ist [Auswerten] und gibt zusätzlich die Stellen an,
// an denen weitergerechnet wurde, obwohl etwas fehlte.
//
// **Als eigene Funktion und nicht als dritter Rückgabewert von [Auswerten]**:
// die meisten Aufrufer haben keine Verwendung dafür, und ein `_` an fünfzig
// Stellen versteckt die Auskunft besser als ein eigener Name sie zeigt.
func AuswertenMitVorbehalten(ctx context.Context, knoten *Knoten, umgebung *Umgebung) (
	Wert, []Vorbehalt, error,
) {
	if umgebung == nil {
		umgebung = NeueUmgebung(nil)
	}

	a := &auswerter{ctx: ctx}

	wert, err := a.werte(knoten, umgebung.Darueber())

	return wert, a.vorbehalte, err
}

// Vorbelegen rechnet eine Quelle **in** der übergebenen Umgebung.
//
// # Warum das nicht [Auswerten] ist
//
// `Auswerten` rechnet in einer Kindebene, und der Grund steht an [Umgebung]:
// eine Zuweisung in einem Feldskript darf die Umgebung des Aufrufers nicht
// verändern, sonst beeinflussen sich zwei Felder derselben Vorlage.
//
// Ein `init script:` ist der eine Fall, in dem genau das gewollt ist — es
// **ist** die Vorbelegung. `magic-m15.mse-style` legt dort `template`,
// `mana_sort`, `module_card` und ein Dutzend weitere an, und ohne sie ließ sich
// von 356 `image:`-Angaben des Packs keine einzige ausrechnen
// (#186).
//
// **Der Rückgabewert fällt weg**, denn er ist der Wert der letzten Anweisung
// und sagt nichts. Was zählt, sind die Namen, die danach dastehen.
func Vorbelegen(ctx context.Context, quelle string, umgebung *Umgebung) error {
	_, fehler := VorbelegenMitVorbehalten(ctx, quelle, umgebung)

	return fehler
}

// VorbelegenMitVorbehalten ist [Vorbelegen] und gibt zusätzlich die Stellen
// an, an denen weitergerechnet wurde, obwohl etwas fehlte.
//
// **Für das Initskript ist das der wichtige Weg** (#216): ein einziges
// unübersetzbares Muster in Zeile 8425 von `magic.mse-game` nahm vorher die
// 11.000 Zeilen danach mit — und mit ihnen `cmc`, an dem fünf Exportvorlagen
// hängen.
func VorbelegenMitVorbehalten(ctx context.Context, quelle string, umgebung *Umgebung) (
	[]Vorbehalt, error,
) {
	knoten, err := Zerteile(quelle)
	if err != nil {
		return nil, err
	}

	a := &auswerter{ctx: ctx}

	// **Ein Initskript ist eine Folge von Zuweisungen, und jede steht für
	// sich** (#240).
	//
	// Deshalb wird hier Glied für Glied gerechnet und nicht der Baum am Stück:
	// scheitert eines, bleibt sein Name ungesetzt, und die übrigen entstehen
	// trotzdem. `magic.mse-game` rief an Zeile 14741 ein `expand_keywords`,
	// das dieses Programm nicht kennt — und verlor damit die **4.000 Zeilen
	// danach**, unter ihnen `module_card`, `card_style` und
	// `color_background`. Genau die drei, an denen jede `image:`-Angabe
	// scheiterte und die Karte weiß blieb.
	//
	// Das ist §8 für das Vorbelegen: ein Set mit 300 lesbaren und zwei
	// unlesbaren Karten wird mit 300 Karten geöffnet.
	//
	// # Warum das nur hier gilt und nicht bei [Auswerten]
	//
	// Ein Feldskript **ist** sein Ergebnis; einen Fehler darin zu übergehen
	// hieße, einen Wert zu erfinden. Ein Initskript ist eine Liste von
	// Vereinbarungen, und eine ausgefallene Vereinbarung meldet sich beim
	// ersten, der sie braucht — mit dem Namen, der fehlt.
	// Ohne Prüfung auf `nil`: `Zerteile` gibt entweder einen Baum oder einen
	// Fehler, und der ist oben abgefangen. Ein Zweig, den kein Test betreten
	// kann, ist eine Behauptung darüber, was im Ernstfall geschieht.
	if knoten.Form == FormFolge {
		for _, glied := range knoten.Kinder {
			if _, fehler := a.werte(glied, umgebung); fehler != nil {
				a.vorbehalte = append(a.vorbehalte, Vorbehalt{
					Zeile: glied.Zeile, Spalte: glied.Spalte, Meldung: fehler.Error(),
				})
			}
		}

		return a.vorbehalte, nil
	}

	_, fehler := a.werte(knoten, umgebung)

	return a.vorbehalte, fehler
}

// AuswertenQuelle zerteilt und rechnet in einem Zug.
func AuswertenQuelle(ctx context.Context, quelle string, umgebung *Umgebung) (Wert, error) {
	knoten, err := Zerteile(quelle)
	if err != nil {
		return Nichts, err
	}

	return Auswerten(ctx, knoten, umgebung)
}

type auswerter struct {
	ctx   context.Context //nolint:containedctx // trägt Abbruch durch die Rekursion
	tiefe int

	// netz und gelesen sind nur gesetzt, wenn die Auswertung aus einem `Netz`
	// kommt. Ohne sie verhält sich der Auswerter wie zuvor — ein Skript, das
	// niemand beobachtet, kostet nichts.
	netz    *Netz
	gelesen map[string]bool

	// vorbehalte sind die Stellen, an denen weitergerechnet wurde, obwohl
	// etwas fehlte — siehe [Vorbehalt].
	vorbehalte []Vorbehalt

	// fertig ist `ctx.Done()`, einmal geholt — siehe [auswerter.abgebrochen].
	fertig       <-chan struct{}
	fertigGeholt bool
}

// abgebrochen sagt, ob der Kontext beendet ist — **über seinen `Done`-Kanal**,
// einmal je Auswerter geholt. E04
// (#2241):
// `ctx.Err()` bei jedem Schritt lief jedes Mal die Kette der Kontexte mit
// Werten hinauf und kostete beim Export über Black 8 % der Rechenzeit. Ein
// Kontext, der nie endet, hat keinen Kanal — dann bleibt `select` beim
// `default`.
func (a *auswerter) abgebrochen() error {
	if !a.fertigGeholt {
		a.fertig, a.fertigGeholt = a.ctx.Done(), true
	}

	select {
	case <-a.fertig:
		return a.ctx.Err()
	default:
		return nil
	}
}

// merke schreibt einen gelesenen Namen ins Protokoll, falls eines läuft.
func (a *auswerter) merke(name string) {
	if a.gelesen != nil {
		a.gelesen[name] = true
	}
}

// werte ist die Rekursion. Jeder Eintritt kostet eine Stufe.
func (a *auswerter) werte(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	if err := a.abgebrochen(); err != nil {
		return Nichts, err
	}

	a.tiefe++
	defer func() { a.tiefe-- }()

	if a.tiefe > Tiefengrenze {
		return Nichts, a.fehler(knoten, SchluesselZuTief,
			strconv.Itoa(Tiefengrenze))
	}

	return a.nachForm(knoten, umgebung)
}

// fehler nennt die Stelle im Skript, nicht nur die Ursache.
func (a *auswerter) fehler(
	knoten *Knoten, schluessel Meldungsschluessel, werte ...string,
) *Fehler {
	return NeuerFehler(knoten.Zeile, knoten.Spalte, schluessel, werte...)
}
