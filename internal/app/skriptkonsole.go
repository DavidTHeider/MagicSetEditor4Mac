package app

// Die **Skriptkonsole** im Fenster
// (#1070,
// Werkstück 3 von Phase W).
//
// # Warum es sie im Fenster gibt
//
// MSE-Script ist vollständig portiert — und wer einen Ausdruck ausprobieren
// wollte, musste die Kommandozeile bemühen. MSE2 hat dafür einen eigenen
// Bereich (`src/gui/set/console_panel.cpp`, Fassung f9d9356 vom 2021-07-06),
// und §15 nennt seine Kommandozeile als **Orakel** dieses Projekts: eine
// eigene Konsole beantwortet dieselben Fragen über **dieses** Programm und
// macht die beiden vergleichbar.
//
// # Sie ändert nichts
//
// Die Sprache kann es nicht — `internal/schichten` weist jeden Zugriff auf das
// Dateisystem aus `msescript` ab —, und diese Bindung legt nichts dazu.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// KonsolenAntwort trägt das Ergebnis eines Ausdrucks oder einen Fehler.
type KonsolenAntwort struct {
	// Quelltext ist das Ergebnis als **Quelltext** — `"a"` und nicht `a`.
	//
	// So zeigt MSE es (`src/cli/cli_main.cpp`, Zeile 234, Fassung f9d9356 vom
	// 2021-07-06), und der Unterschied ist der Punkt: wer eine Konsole
	// benutzt, will eine Zeichenkette von einem Namen unterscheiden.
	Quelltext string `json:"quelltext"`

	// Text ist dasselbe Ergebnis in gewöhnlicher Form.
	Text string `json:"text"`

	// Art ist die Art des Werts — `text`, `zahl`, `abbildung`.
	Art string `json:"art"`

	// Befunde sind die Vorbehalte beim Aufbau der Umgebung. **Sie halten die
	// Auswertung nicht auf** (§8): ein Ausdruck, der die Vorlage nicht
	// braucht, läuft trotzdem.
	Befunde []string `json:"befunde"`

	// Stelle ist der Ort eines Syntaxfehlers — **nil**, wenn er keinen hat.
	//
	// # Warum sie neben dem Fehler steht und nicht darin
	//
	// [Fehler] ist die Gestalt, die **jede** Bindung liefert; eine Zeile und
	// eine Spalte gibt es nur hier. Sie dort einzubauen hieße, jede Antwort
	// dieser Anwendung um zwei Felder zu erweitern, die fast immer leer sind.
	//
	// **Und nicht jeder Fehler hat eine.** Eine Karte, die es nicht gibt, hat
	// keine Zeile — dann bleibt sie leer, und das Fenster streicht nichts an
	// (#1164).
	Stelle *Skriptstelle `json:"stelle"`

	// DauerMs ist, wie lange die Auswertung gedauert hat — Werkstück U13 von
	// Phase 5.1 (#1424).
	DauerMs int64 `json:"dauerMs"`

	// Abgebrochen sagt, dass der Lauf endete, bevor ein Ergebnis da war —
	// über [App.SkriptAbbrechen] oder einen endenden Kontext. **Kein
	// Fehler**: es ist, was jemand wollte.
	Abgebrochen bool `json:"abgebrochen"`

	// Eintraege sind die Einträge einer Abbildung, rekursiv — leer für jeden
	// anderen Wert. Das Fenster zeigt sie als Baum.
	Eintraege []KonsolenEintrag `json:"eintraege"`

	Fehler *Fehler `json:"fehler"`
}

// Skriptstelle ist ein Ort im Ausdruck — **eins-basiert**, wie der Zerteiler
// ihn meldet.
type Skriptstelle struct {
	Zeile  int `json:"zeile"`
	Spalte int `json:"spalte"`
}

// stelleAus holt Zeile und Spalte aus einem Skriptfehler.
//
// **`errors.As` und kein Zeichenkettenvergleich** (§8). Der Zerteiler wirft
// `*msescript.Fehler`; der Dienst umschließt ihn mit `%w`, und hier wird er
// wieder ausgepackt.
func stelleAus(fehler error) *Skriptstelle {
	var skript *msescript.Fehler
	if !errors.As(fehler, &skript) {
		return nil
	}

	return &Skriptstelle{Zeile: skript.Zeile, Spalte: skript.Spalte}
}

// Skriptkonsolenquelle ist die Naht zur Auswertung.
type Skriptkonsolenquelle interface {
	Auswerten(ctx context.Context, karteID int64, ausdruck string) (KonsolenDTO, error)

	// Namen gibt, was ein Ausdruck an dieser Karte nennen darf
	// (#1164).
	Namen(ctx context.Context, karteID int64) ([]string, error)
}

// Namensantwort trägt die Namen der Umgebung oder einen Fehler.
type Namensantwort struct {
	// Namen sind sortiert und ohne Doppel — die Bibliothek **und** das, was
	// die Vorlage dazulegt.
	Namen []string `json:"namen"`

	Fehler *Fehler `json:"fehler"`
}

// Skriptnamen gibt, was ein Ausdruck an dieser Karte nennen darf.
//
// **Für Vervollständigung und die Liste in der Konsole.** Für jemanden, der
// MSE-Script nicht kennt, ist ein leeres Feld keine Einladung — das steht so
// im Issue, und es ist der Grund für diese Bindung.
func (a *App) Skriptnamen(ctx context.Context, karteID string) Namensantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return Namensantwort{Namen: []string{}, Fehler: befund}
	}

	if a.skriptkonsole == nil {
		return Namensantwort{Namen: []string{}, Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Skriptkonsole ist nicht eingerichtet", domain.ErrValidation))}
	}

	namen, err := a.skriptkonsole.Namen(ctx, kennung)
	if err != nil {
		return Namensantwort{Namen: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	// **Eine leere Liste und kein `nil`** — dieselbe Überlegung wie bei den
	// Befunden: über die Wails-Bindung stünde dort sonst `null`.
	if namen == nil {
		namen = []string{}
	}

	return Namensantwort{Namen: namen}
}

// MitSkriptkonsole setzt die Naht.
func (a *App) MitSkriptkonsole(quelle Skriptkonsolenquelle) { a.skriptkonsole = quelle }

// Skript wertet einen Ausdruck im Kontext einer Karte aus.
//
// **Eine Karte ist Pflicht.** MSEs Konsole arbeitet an einem Set und kennt
// „die aufgeschlagene Karte" aus dem Fenster; ohne eine gibt es kein `card`,
// und ein erfundenes wäre eine Umgebung, die niemand nachvollziehen kann.
func (a *App) Skript(ctx context.Context, karteID, ausdruck string) KonsolenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return KonsolenAntwort{Befunde: []string{}, Eintraege: []KonsolenEintrag{}, Fehler: befund}
	}

	if a.skriptkonsole == nil {
		return KonsolenAntwort{Befunde: []string{}, Eintraege: []KonsolenEintrag{}, Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Skriptkonsole ist nicht eingerichtet", domain.ErrValidation))}
	}

	// **Ein Lauf je Karte, mit Abbruch** (§7, Regel 4; #1424). Ein zweiter an
	// derselben Karte bekommt `konflikt`; eine andere Karte ist ein anderer
	// Datensatz und läuft.
	lauf, beenden, befund := a.skriptBeginnen(ctx, kennung)
	if befund != nil {
		return KonsolenAntwort{Befunde: []string{}, Eintraege: []KonsolenEintrag{}, Fehler: befund}
	}

	defer beenden()

	begonnen := time.Now()
	ergebnis, err := a.skriptkonsole.Auswerten(lauf, kennung, ausdruck)
	dauer := time.Since(begonnen).Milliseconds()

	if err != nil {
		// **Abgebrochen ist kein Fehler.** Der Auswerter gibt den Grund des
		// Kontexts zurück, der Service umschließt ihn — `errors.Is` findet ihn.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return KonsolenAntwort{
				Befunde: []string{}, Eintraege: []KonsolenEintrag{}, DauerMs: dauer, Abgebrochen: true,
			}
		}

		return KonsolenAntwort{
			Befunde: []string{}, Eintraege: []KonsolenEintrag{}, DauerMs: dauer,
			Stelle: stelleAus(err), Fehler: Uebersetze(ctx, err),
		}
	}

	// **Eine leere Liste und kein `nil`**: über die Wails-Bindung stünde dort
	// sonst `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
	befunde := ergebnis.Befunde
	if befunde == nil {
		befunde = []string{}
	}

	eintraege := ergebnis.Eintraege
	if eintraege == nil {
		eintraege = []KonsolenEintrag{}
	}

	return KonsolenAntwort{
		Quelltext: ergebnis.Quelltext, Text: ergebnis.Text,
		Art: ergebnis.Art, Befunde: befunde, Eintraege: eintraege, DauerMs: dauer,
	}
}
