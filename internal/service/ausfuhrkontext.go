package service

// **Ein Kontext je Stilbogen in der Ausfuhr** — VK-01b
// (#2182).
//
// # Was MSE2 tut, nachgelesen
//
// `SetScriptContext::getContext(card)` wählt zuerst `stylesheetForP(card)` —
// den eigenen Stil der Karte, sonst den des Sets — und legt für jeden
// Stilbogen **einmal** einen Kontext an: `set`, `game`, `stylesheet`,
// `card_style`, eine Attrappe als `card` und `styling`, **dann** das
// `init script` des Spiels und danach das des Stils
// (`src/script/script_manager.cpp`, Zeilen 29–74, Fassung f9d9356 vom
// 2021-07-06). `updateAll` rechnet jede Karte in diesem Kontext
// (dieselbe Datei, Zeilen 265–300).
//
// Gemessen in VK-01 (`docs/belege/vk01-kartenstil.md`): `shape` von
// `magic.mse-game` rechnet `stylesheet; card_shape()`, und der Stil
// überschreibt `card_shape`. Ohne ihn gab die Ausfuhr den Wert des Spiels.
//
// # Was hier davon steht
//
// Dieselbe Reihenfolge wie im Zeichenpfad ([neuerUmgebungsbauer]): erst die
// Vorlage, dann das Spiel, dann der Stil. **Ein Kontext je Stilbogen und nicht
// je Karte**, wie dort ([stilbauer]) — die Initskripte laufen einmal je Stil.
//
// **Die Vorbehalte der Initskripte werden nicht gemeldet** — dieselbe Haltung
// wie [spielumgebung], aus der die Ausfuhr bis dahin rechnete: sie ist eine
// Ausgabe der Werte und keine Diagnose der Vorlage; wer sie sehen will, öffnet
// die Karte, und der Zeichenpfad meldet sie dort.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// ausfuhrkontext ist der Kontext eines Stilbogens.
type ausfuhrkontext struct {
	spiel   *msescript.Umgebung
	modell  Vorlage
	vermerk *Schluesselwortverwendung
}

// ausfuhrkontexte hält die Kontexte eines Laufs, nach dem Namen des Stils.
type ausfuhrkontexte struct {
	modell Vorlage
	stile  Ausfuhrstile
	je     map[string]*ausfuhrkontext

	// erste ist die erste Karte des Sets — MSEs Attrappe für `card`, ehe
	// eine Karte gerechnet wird (`set.cards.front()`).
	erste *domain.CardRow
}

func neueAusfuhrkontexte(modell Vorlage, stile Ausfuhrstile, karten []*domain.CardRow) *ausfuhrkontexte {
	// **Gemerkt** — Befund 3 des Nachlaufs über Phase G: die Quelle öffnet ihr
	// Paket bei jedem Ruf, und gefragt wurde zweimal je Karte. Seit E04
	// (#2241) mit demselben Mittel wie die übrige Ausfuhr ([Ausfuhrstile.gemerkt]).
	kontexte := &ausfuhrkontexte{
		modell: modell, stile: stile.gemerkt(),
		je: map[string]*ausfuhrkontext{},
	}
	if len(karten) > 0 {
		kontexte.erste = karten[0]
	}

	return kontexte
}

// fuer gibt den Kontext dieser Karte und legt ihn beim ersten Mal an.
//
// **Angelegt wie bei MSE**, und das ist nachgelesen (Befund 1 des Nachlaufs
// über Phase G): vor beiden Initskripten stehen `stylesheet`, `card_style`,
// als `card` die **erste Karte des Sets** und als `styling` die
// **setweiten** Stiloptionen; die Optionen einer einzelnen Karte setzt erst
// [auszaehlumgebung] je Karte (`SetScriptContext::getContext`,
// `src/script/script_manager.cpp`, Zeilen 39–44 und 64–73, Fassung f9d9356
// vom 2021-07-06). Hier standen die Optionen der ersten Karte **dieses
// Stils** da, und ein Initskript, das `styling` las, gab je nach Reihenfolge
// der Karten anderes.
func (k *ausfuhrkontexte) fuer(ctx context.Context, karte *domain.CardRow) *ausfuhrkontext {
	modell := k.vorlageFuer(karte)

	name := stilname(modell.Stil)
	if kontext, da := k.je[name]; da {
		return kontext
	}

	kontext := &ausfuhrkontext{modell: modell, vermerk: &Schluesselwortverwendung{}}
	kontext.spiel = msescript.Bibliothek().Darueber()

	belegeSchluesselwortnaht(kontext.spiel, modell.Schluesselwoerter, kontext.vermerk)

	// **`set` und `card` stehen in voller Gestalt da**, bevor ein Skript sie
	// liest — die Gründe stehen bei [spielumgebung].
	kontext.spiel.Setze("set", leereAbbildung(modell.Setfelder))
	kontext.spiel.Setze("card", leereAbbildung(modell.Kartenfelder))

	belegeVorlage(ctx, kontext.spiel, modell, nil, nil)

	if k.erste != nil {
		felder := make([]msescript.Eintrag, 0, len(k.erste.Values)+1)
		for _, wert := range k.erste.Values {
			felder = append(felder, msescript.Eintrag{
				Name: umgebungsname(wert.Field), Wert: msescript.Text(wert.Text),
			})
		}

		kontext.spiel.Setze("card", kartenabbildungMit(mitNotizen(felder, k.erste),
			modell.Kartenfelder, modell.Setpaket, Skriptkartenkennung(k.erste)))
	}

	// **`card_style` ohne Inhaltsmaße** — die Ausfuhr misst keinen Text, und
	// MSE bindet hier den Stil der Felder, wie der Stil ihn erklärt.
	kontext.spiel.Setze("card_style",
		feldstile(ctx, modell.Kartenfelder, modell.Stil.Felder, nil, kontext.spiel))

	_ = vorbelegt(ctx, modell.Spielskript, "init script des Spiels", kontext.spiel)
	_ = vorbelegt(ctx, modell.Stil.Initskript, "init script", kontext.spiel)

	// **Der Speicher der Ordnungen gehört dem Kontext**, damit ihn alle seine
	// Karten teilen: `card_number` ordnet das ganze Set, und das einmal je
	// Filter.
	kontext.spiel.OrdnungenMerken()

	k.je[name] = kontext

	return kontext
}

// vorlageFuer gibt das Modell mit dem Stil dieser Karte — ihrem eigenen, sonst
// dem des Sets (`Set::stylesheetFor`, [Stilspeicher.StilFuerKarte]).
//
// **Der Rückfall ist der Stil des Sets** und, fehlt auch der, die Vorlage
// ohne Stil — so rechnete die Ausfuhr vorher. Dass ein Stil fehlt, meldet
// [mitStilen]; hier zöge ein zweiter Befund nur dieselbe Zeile nach.
func (k *ausfuhrkontexte) vorlageFuer(karte *domain.CardRow) Vorlage {
	modell := k.modell
	if k.stile == nil {
		return modell
	}

	eigener, _ := karte.Value(Kartenstilfeld)
	if stil, da := k.stile(strings.TrimSpace(eigener)); da {
		modell.Stil = stil

		return modell
	}

	if stil, da := k.stile(""); da {
		modell.Stil = stil
	}

	return modell
}

// ordnungenVergessen leert den Speicher der Ordnungen in jedem Kontext — MSEs
// `clearOrderCache` am Anfang einer Runde.
func (k *ausfuhrkontexte) ordnungenVergessen() {
	for _, kontext := range k.je {
		kontext.spiel.OrdnungenVergessen()
	}
}
