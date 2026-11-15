package service

import (
	"cmp"
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Stilablage führt Stil und Stiloptionen eines Sets — die Naht zu `storage`.
type Stilablage interface {
	Stilwerte(ctx context.Context, setID int64) ([]domain.Value, error)
	StilSetzen(ctx context.Context, setID int64, spiel, stil string, werte []domain.Value) error

	// **Und eine einzelne Option, für sich** — aus dem Zweitblick über Phase L.
	//
	// [Stilwaehler.OptionSetzen] las bis zum 2026-08-22 die vollständige Liste,
	// änderte einen Eintrag und schrieb alles zurück, samt Spiel und Stil. Zwei
	// Optionsschreibungen verloren einander, und ein alter Lauf konnte einen
	// Stilwechsel zurückdrehen — beides im Zwei-Schreiber-Betrieb der
	// Normalfall (§7 Regel 4).
	StiloptionSetzen(ctx context.Context, setID int64, feld, wert string) error
	StiloptionEntfernen(ctx context.Context, setID int64, feld string) error
}

// Feldquelle nennt die Feldnamen, die in einem Set vorkommen.
type Feldquelle interface {
	Felder(ctx context.Context, setID int64) ([]string, error)
}

// Stilwaehler beantwortet die drei Fragen der Stilauswahl.
//
// **Welche Stile gibt es, was kostet ein Wechsel, und welche Optionen hat der
// Stil?** Die erste beantwortet der Paketkatalog (#54), die zweite ein
// Vergleich zweier Spielpakete, die dritte das Stilpaket selbst — und nicht
// eine Liste bei uns
// (#41).
type Stilwaehler struct {
	koepfe  Kopfquelle
	katalog Paketablage
	ablage  Stilablage
	felder  Feldquelle
	wurzeln []string

	// vergessen sagt dem Stilspeicher, dass ein Set seinen Stil gewechselt hat
	// (#693).
	//
	// **Als Naht und nicht als Aufgabe des Aufrufers.** Wer wechselt, muss
	// nicht wissen, dass irgendwo eine Vorlage gemerkt liegt — sonst hinge die
	// Zusage an jedem einzelnen Aufrufer, und der MCP-Server oder ein Test
	// vergäße sie.
	//
	// **Nil ist erlaubt** und heißt „niemand merkt sich etwas": ein Test, der
	// nur den Wechsel prüft, braucht keinen Speicher.
	vergessen func(setID int64)
}

// NeuerStilwaehler baut den Wähler über seinen vier Nähten und den Paketwurzeln.
func NeuerStilwaehler(
	koepfe Kopfquelle, katalog Paketablage, ablage Stilablage, felder Feldquelle, wurzeln ...string,
) *Stilwaehler {
	return &Stilwaehler{koepfe: koepfe, katalog: katalog, ablage: ablage, felder: felder, wurzeln: wurzeln}
}

// MitSpeicher sagt dem Wähler, wessen Gedächtnis ein Wechsel ungültig macht.
func (w *Stilwaehler) MitSpeicher(speicher *Stilspeicher) *Stilwaehler {
	return w.MitVergessen(speicher.Vergiss)
}

// MitVergessen setzt die Naht unmittelbar.
//
// **Für den Test, der die Zusage prüft.** `MitSpeicher` verlangt einen echten
// Stilspeicher samt Paketwurzeln; die Frage hier ist nur, **ob** und **wann**
// gemeldet wird.
func (w *Stilwaehler) MitVergessen(vergessen func(setID int64)) *Stilwaehler {
	w.vergessen = vergessen

	return w
}

// Stilangebot ist ein installierter Stil, wie die Auswahl ihn zeigt.
type Stilangebot struct {
	// Paket ist der Name **ohne** Endung — `probe-schlicht`. So steht er im
	// Set (`stylesheet:`), und so sucht ihn der Stilspeicher.
	Paket string

	// Anzeigename ist, was über dem Stil steht — [domain.Paket.Anzeigename]
	// (UX-05, S07.NAT-02, #1677).
	Anzeigename string

	Spiel   string
	Version string

	// EigenesSpiel sagt, ob dieser Stil zum Spiel des Sets gehört.
	//
	// **Angezeigt und nicht weggefiltert**: ein Stil eines anderen Spiels ist
	// eine Möglichkeit mit Preis, und der Preis steht in [Folgen]. Wer ihn
	// verstecken wollte, nähme eine Entscheidung vorweg, die den Nutzenden
	// gehört.
	EigenesSpiel bool

	// Gewaehlt sagt, ob das Set gerade diesen Stil führt.
	Gewaehlt bool
}

// Stile nennt die installierten Stile — den eigenen zuerst.
func (w *Stilwaehler) Stile(ctx context.Context, setID int64) ([]Stilangebot, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	pakete, err := w.katalog.Pakete(ctx)
	if err != nil {
		return nil, err
	}

	angebote := make([]Stilangebot, 0, len(pakete))

	// **In MSE2s Ordnung** — `position hint:` und Name
	// ([domain.VergleicheNachRang]; UX-05, #1677).
	for _, paket := range slices.SortedStableFunc(slices.Values(pakete), domain.VergleicheNachRang) {
		if paket.Art != string(msepaket.ArtStil) {
			continue
		}

		angebote = append(angebote, Stilangebot{
			Paket:        paket.Name,
			Anzeigename:  Anzeigename(paket),
			Spiel:        spielAus(paket),
			Version:      paket.Version,
			EigenesSpiel: spielAus(paket) == kopf.Game,
			// **Gegen beide Formen** (#951): das Set nennt die Kurzform,
			// der Katalog führt den Paketnamen. Ein Vergleich der rohen
			// Werte fand nie einen Treffer, und die Auswahl stand auf
			// „kein installierter Stil", während die Karte gezeichnet wurde.
			Gewaehlt: slices.Contains(stilpaketnamen(kopf.Style, kopf.Game), paket.Name),
		})
	}

	// **Der eigene zuerst** — stabil, damit die Ordnung darüber bleibt. Eine
	// Liste von 194 Stilen, in der die drei brauchbaren irgendwo stehen, ist
	// keine Auswahl; MSE2 zeigt die fremden gar nicht (`StylePanel`,
	// `src/gui/set/style_panel.cpp`, Zeile 87, Fassung f9d9356 vom 2021-07-06).
	slices.SortStableFunc(angebote, func(a, b Stilangebot) int {
		return cmp.Compare(fremdheit(a), fremdheit(b))
	})

	return angebote, nil
}

// fremdheit stellt den eigenen Stil vor den fremden.
func fremdheit(angebot Stilangebot) int {
	if angebot.EigenesSpiel {
		return 0
	}

	return 1
}

// spielAus liest das Spiel einer Abhängigkeit — `depends on: magic.mse-game`.
//
// **Aus dem Katalog und nicht aus dem Paket.** Das Paket zu öffnen hieße, für
// eine Liste von 194 Stilen 194 Dateien zu lesen; der Katalog hat die Angabe
// schon. Der Preis steht daneben: ein Stil ohne `depends on:` nennt hier kein
// Spiel und gilt damit als fremd — sichtbar bleibt er trotzdem.
func spielAus(paket domain.Paket) string {
	for _, bedarf := range paket.Braucht {
		if name, ist := strings.CutSuffix(bedarf.Paket, "."+string(msepaket.ArtSpiel)); ist {
			return name
		}
	}

	return ""
}

// Folgen sagt, was ein Stilwechsel kostet — **bevor** er stattfindet.
type Folgen struct {
	// UngedeuteteFelder sind Kartenfelder, die das neue Spiel nicht kennt.
	//
	// **Sie gehen nicht verloren**, und das ist der Unterschied zu einer
	// Warnung, die es behauptete: der Bestand führt jeden Feldwert mit, auch
	// einen, den kein Spiel kennt (§8, und die Rundlaufzusage aus #48). Was
	// verlorengeht, ist ihre **Darstellung** — sie werden nicht mehr gezeichnet.
	UngedeuteteFelder []string

	// VerloreneOptionen sind Stiloptionen, die der neue Stil nicht führt.
	//
	// **Diese gehen wirklich verloren.** Ein Wert ohne Option wäre einer, den
	// kein Skript liest und den niemand ändern kann; der Wechsel räumt ihn.
	VerloreneOptionen []string

	// Spielwechsel sagt, dass das neue Spiel ein anderes ist.
	Spielwechsel bool
}

// Folgen rechnet aus, was ein Wechsel auf diesen Stil kostet.
func (w *Stilwaehler) Folgen(ctx context.Context, setID int64, paket string) (*Folgen, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	neu, err := w.stil(paket, kopf.Game)
	if err != nil {
		return nil, err
	}

	folgen := &Folgen{
		UngedeuteteFelder: []string{},
		VerloreneOptionen: []string{},
		Spielwechsel:      neu.Spiel != kopf.Game,
	}

	if folgen.Spielwechsel {
		ungedeutet, fehler := w.ungedeutet(ctx, setID, neu.Spiel)
		if fehler != nil {
			return nil, fehler
		}

		folgen.UngedeuteteFelder = ungedeutet
	}

	werte, err := w.ablage.Stilwerte(ctx, setID)
	if err != nil {
		return nil, err
	}

	for _, wert := range werte {
		if !kenntOption(neu.Stiloptionen, wert.Field) {
			folgen.VerloreneOptionen = append(folgen.VerloreneOptionen, wert.Field)
		}
	}

	return folgen, nil
}

// Wechseln setzt den Stil eines Sets und behält, was der neue Stil kennt.
func (w *Stilwaehler) Wechseln(ctx context.Context, setID int64, paket string) error {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return err
	}

	neu, err := w.stil(paket, kopf.Game)
	if err != nil {
		return err
	}

	werte, err := w.ablage.Stilwerte(ctx, setID)
	if err != nil {
		return err
	}

	bleibt := make([]domain.Value, 0, len(werte))

	for _, wert := range werte {
		if kenntOption(neu.Stiloptionen, wert.Field) {
			bleibt = append(bleibt, wert)
		}
	}

	if err := w.ablage.StilSetzen(ctx, setID, neu.Spiel, paket, bleibt); err != nil {
		return err
	}

	// **Erst nach dem Schreiben** (§7, Regel 3): was gemerkt ist, gilt so
	// lange, wie der Bestand es sagt. Ein Vergessen vor einem gescheiterten
	// Schreiben ließe den nächsten Aufruf neu lesen — richtig, aber
	// unbegründet.
	if w.vergessen != nil {
		w.vergessen(setID)
	}

	return nil
}

// kopf sucht den Setkopf.
func (w *Stilwaehler) kopf(ctx context.Context, setID int64) (domain.SetHead, error) {
	koepfe, err := w.koepfe.Setkoepfe(ctx)
	if err != nil {
		return domain.SetHead{}, err
	}

	kopf, da := kopfMit(koepfe, setID)
	if !da {
		return domain.SetHead{}, fmt.Errorf("%w: das Set %d führt der Bestand nicht",
			domain.ErrNotFound, setID)
	}

	return kopf, nil
}

// ungedeutet nennt die Felder des Sets, die ein Spiel nicht kennt.
func (w *Stilwaehler) ungedeutet(ctx context.Context, setID int64, spiel string) ([]string, error) {
	felder, err := w.felder.Felder(ctx, setID)
	if err != nil {
		return nil, err
	}

	bekannt, err := w.spielfelder(spiel)
	if err != nil {
		return nil, err
	}

	ungedeutet := make([]string, 0)

	for _, feld := range felder {
		if !bekannt[feld] {
			ungedeutet = append(ungedeutet, feld)
		}
	}

	return ungedeutet, nil
}

// spielmodell öffnet das Spielpaket und gibt sein Modell.
//
// **Die Module zuerst** — der Grund steht bei [Stilwaehler.spielfelder].
func (w *Stilwaehler) spielmodell(spiel string) (domain.Game, error) {
	name := spiel + "." + string(msepaket.ArtSpiel)

	for _, wurzel := range w.wurzeln {
		geoeffnet, err := msepaket.Oeffne(filepath.Join(wurzel, name))
		if err != nil {
			continue
		}

		_ = geoeffnet.ModuleEinbinden(msepaket.NeuerPaketleser(w.wurzeln...)) //nolint:errcheck // ein fehlendes Modul ist ein Befund, kein Abbruch

		modell, _ := geoeffnet.Spielmodell()
		_ = geoeffnet.Schliessen() //nolint:errcheck // das Modell ist vollständig kopiert

		return modell, nil
	}

	return domain.Game{}, fmt.Errorf("%w: das Spielpaket %q liegt in keinem Paketverzeichnis",
		domain.ErrNotFound, name)
}

// spielfelder gibt die Feldnamen eines Spielpakets.
func (w *Stilwaehler) spielfelder(spiel string) (map[string]bool, error) {
	name := spiel + "." + string(msepaket.ArtSpiel)

	for _, wurzel := range w.wurzeln {
		geoeffnet, err := msepaket.Oeffne(filepath.Join(wurzel, name))
		if err != nil {
			continue
		}

		// **Die Module zuerst.** `magic.mse-game` führt **kein** `card field:`
		// im Kopf — seine Felder stehen in `include file: card_fields`. Ohne
		// diesen Schritt hätte das Spiel null Felder, und ein Wechsel meldete
		// jedes Feld des Sets als ungedeutet. Gesehen im laufenden Fenster:
		// „name" und „text" galten als unbekannt, obwohl Magic sie führt.
		_ = geoeffnet.ModuleEinbinden(msepaket.NeuerPaketleser(w.wurzeln...)) //nolint:errcheck // ein fehlendes Modul ist ein Befund, kein Abbruch

		modell, _ := geoeffnet.Spielmodell()
		_ = geoeffnet.Schliessen() //nolint:errcheck // das Modell ist vollständig kopiert

		bekannt := make(map[string]bool, len(modell.Fields))
		for i := range modell.Fields {
			bekannt[modell.Fields[i].Name] = true
		}

		return bekannt, nil
	}

	return nil, fmt.Errorf("%w: das Spielpaket %q liegt in keinem Paketverzeichnis",
		domain.ErrNotFound, name)
}

// kenntOption sagt, ob eine Optionsliste einen Namen führt.
func kenntOption(optionen []domain.StyleField, name string) bool {
	for _, option := range optionen {
		if option.Name == name {
			return true
		}
	}

	return false
}
