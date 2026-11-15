// **Das Zeichenwerk** — alles, was gebraucht wird, um eine Karte zu zeichnen
// (#886).
//
// # Warum es das gibt
//
// Der Weg von einer Kartenkennung zu einem Bild führt über sieben Teile:
// Paketwurzeln, Stilspeicher, Schriftsuche, Bildquelle, Planbauer, Rasterer
// und die Ausfuhr darüber. Sie hängen aneinander, und **jedes Teil zweimal
// zusammenzusetzen heißt, zwei Antworten auf dieselbe Frage zu haben** — zwei
// Schriftsuchen etwa gäben denselben Text an zwei Stellen.
//
// Fenster, Kommandozeile, MCP und REST-Dienst brauchen dasselbe.
//
// **Der Beleg, dass die Zusammensetzung eine Sache für sich ist**, steht schon
// im Quelltext, den sie ersetzt: „dieselben Teile wie die Vorschau, und das
// ist die Zusage aus #34 — eine Vorschau und ein ausgegebenes Bild, die
// auseinanderlaufen, tun das an genau einer Stelle, wenn es zwei Wege dorthin
// gibt."

package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// Zeichenleser ist, was das Zeichenwerk aus dem Bestand braucht.
//
// **Die Vereinigung dessen, was die vier Dienste darunter verlangen** — und
// nicht `*storage.SetLeser`: eine Schnittstelle sagt, was der Aufrufer
// braucht, nicht was das Repositorium anbietet (§2).
type Zeichenleser interface {
	service.Kartenquelle
	service.Kopfquelle
	service.Quellenleser
	service.Kartenstilnamen
	Setstand(ctx context.Context, setID int64) (int64, error)
}

// Zeichenwerk hält die Teile, die eine Karte zu einem Bild machen.
//
// **Ein Verbund und keine sieben Rückgabewerte**: wer eines davon braucht,
// braucht meist zwei, und eine Liste von sieben Werten in fester Reihenfolge
// ist eine Einladung zum Vertauschen.
type Zeichenwerk struct {
	// Wurzeln sind die Verzeichnisse, in denen Pakete gesucht werden.
	Wurzeln []string

	// Bildquelle liest Bilder und Symbolzeichnungen aus jedem Paket.
	Bildquelle *msepaket.Bildquelle

	// Bildabdruecke sagt, **was in einem Paketbild steht**, ohne es zu
	// entschlüsseln — der halbe Schlüssel einer Spieltextur (G2, #1459).
	Bildabdruecke *msepaket.Bildabdruecke

	// Spielbilder zeichnet Kartenbilder für den Spieltisch (G2, #1459) —
	// über **demselben** Planbauer und Rasterer wie die Vorschau.
	Spielbilder *service.Spieltexturen

	// Spielbildfach hält sie für die Assetleitung.
	Spielbildfach *vorschauweg.Texturfach

	// Schriften findet Schriftdateien auf dem Rechner und in den Paketen.
	Schriften *schriften.Suche

	// Stile hält den Stilbogen eines Sets samt seinen Bildern.
	Stile *service.Stilspeicher

	// Planbauer macht aus einer Karte einen Zeichenplan.
	Planbauer *service.Planbauer

	// Rasterer macht aus einem Zeichenplan ein Bild.
	Rasterer *render.Rasterer

	// Bildausfuhr schreibt Kartenbilder — Raster und Vektor.
	Bildausfuhr *service.Bildausfuhr

	// Bogenleger setzt Karten maßhaltig auf Papier.
	Bogenleger *service.Bogenleger

	// Vorlagen führt die Exportvorlagen aus.
	Vorlagen *service.Vorlagendienst

	// probleme ist die Prüfung über ganze Sets (PERF02, #2352).
	probleme problemteile

	// Leser liest Karten und Setköpfe — **seit dem 2026-09-02**
	// (#1391).
	//
	// Er kam bisher herein und blieb in den Diensten; die Symbolpalette
	// braucht ihn unmittelbar, weil sie eine **Karte** in den Planbauer gibt.
	Leser Zeichenleser
}

// NeuesZeichenwerk setzt die Teile über einer offenen Datenbank zusammen.
//
// **`wurzeln` kommt von außen**, weil das Fenster sie anders bestimmt als eine
// Fläche ohne Fenster: dort kann ein verwalteter Ort dazukommen, hier nicht.
//
// **Und der Leser auch**, obwohl beide Aufrufer denselben übergeben: dieses
// Paket darf `internal/storage` nur in seinen Composition Roots kennen (§2),
// und diese Datei ist keiner. Sie setzt **Dienste** zusammen; woher die Zeilen
// kommen, sagt der Aufrufer.
func NeuesZeichenwerk(leser Zeichenleser, wurzeln []string, bilderort string) *Zeichenwerk {
	ablage := service.NeueBildablage(bilderort)

	bildquelle := msepaket.NeueBildquelle(wurzeln...)

	// **Die Abdrücke neben der Bildquelle und nicht in ihr** (G2, #1459):
	// jene hält entschlüsselte Bilder à 780 KB, diese 32 Byte je Pfad. Ein
	// gemeinsamer Speicher verdrängte das eine für das andere.
	bildabdruecke := msepaket.NeueBildabdruecke(wurzeln...)

	// **Mit den Schriften der bezogenen Pakete** (#140): eine Vorlage nennt
	// Schriften, die sie mitbringt. Die Reihenfolge ist System vor Paket.
	//
	// # Aus den Wurzeln des Aufrufers und nicht aus `HOME`
	//
	// Hier stand `schriften.Paketwurzel()`, und das war eine **zweite Antwort
	// auf dieselbe Frage**: die Stile kommen aus `wurzeln`, die Schriften kamen
	// aus dem Zuhause dieses Rechners. Wer die Wurzeln umlenkt — ein Test, eine
	// Fläche ohne Fenster, ein anderes Zuhause — bekam Stile von hier und
	// Schriften von dort.
	//
	// **Der Rückfall auf `HOME` bleibt**: das Fenster reicht seine Wurzel
	// herein, und wer ohne eine ruft, meint die des Rechners.
	schriftsuche := schriften.Neue(schriften.Orte()).
		MitPaketen(append([]string{Schriftwurzel()}, wurzeln...)...)

	stile := service.NeuerStilspeicher(leser, wurzeln...).MitBildablage(ablage).MitStilwerten(leser).MitKartenstilen(leser)

	planbauer := service.NeuerPlanbauer().
		MitStilen(stile).
		MitSchriften(schriftsuche).
		MitBildern(bildquelle)

	// **Mit dem Ausweich für fehlende Zeichen** (#515): MSE weicht aus, weil
	// wxWidgets es tut — über die 23.168 kanonischen Karten fehlte `ǵ`
	// 159-mal.
	//
	// **Und mit der Pforte zum Haushalt** (#2020): jede Rasterung — Karte,
	// Bogen, Bildausfuhr, Vorschau, Tisch — belegt einen von zwei Plätzen.
	// Sie hängt hier und nicht in jedem Dienst, weil beide Wurzeln, das Fenster
	// und die Flächen ohne Fenster, diesen einen Rasterer bauen.
	rasterer := render.NeuerRasterer(bildquelle).
		MitSchriften(schriftsuche).
		MitErsatzschrift(schriftsuche.Ersatz).
		MitPforte(service.NeueRasterpforte())

	// **Ein geändertes Paket räumt die Bilder** (#2005): der Stilspeicher liest
	// die Vorlage neu, und Bildquelle und Ebenenspeicher halten, was sie aus
	// dem alten Stand entschlüsselt haben.
	stile.MitPaketwechsel(func() {
		bildquelle.Leeren()
		rasterer.EbenenLeeren()
	})

	vorlagen := service.NeuerVorlagendienst(leser, wurzeln...)
	vorlagen.MitSymbolen(stile, rasterer)

	// **Dieselben Teile für Raster und Vektor** (#852): zwei Schriftsuchen
	// wären zwei Antworten auf die Frage, wie breit ein Text ist — und dann
	// stünde derselbe Text im PNG und im SVG an verschiedenen Stellen.
	ausfuhr := service.NeueBildausfuhr(leser, planbauer, rasterer).
		MitVektor(svg.Quellen{
			Baeume:    rasterer,
			Raster:    rasterer,
			Schriften: schriftsuche,
			Symbole:   bildquelle,
		})

	// **Ein Fach mit einem Budget in Bytes** (G2, #1459): 512 Karten der
	// Tischgüte wiegen rund 128 MB, und ein Tisch zeigt sechzig — die Zahl
	// ist ein **Messstartwert** wie die Zielbreiten und wird gepinnt, wenn am
	// echten Tisch gemessen wurde.
	spielbildfach := vorschauweg.NeuesTexturfach(Spielbildbudget)

	return &Zeichenwerk{
		Leser:         leser,
		Wurzeln:       wurzeln,
		Bildquelle:    bildquelle,
		Bildabdruecke: bildabdruecke,
		Spielbildfach: spielbildfach,
		Spielbilder: service.NeueSpieltexturen(planbauer, rasterer,
			spielressourcen{bilder: bildabdruecke, schriften: schriftsuche},
			spielbildfach),
		Schriften:   schriftsuche,
		Stile:       stile,
		Planbauer:   planbauer,
		Rasterer:    rasterer,
		Bildausfuhr: ausfuhr,
		Bogenleger:  service.NeuerBogenleger(leser, planbauer, rasterer),
		Vorlagen:    vorlagen,
	}
}

// MitPruefer setzt den Riegel vor beide Wege, die Dateien schreiben —
// Kartenbild und Druckbogen (0078).
func (z *Zeichenwerk) MitPruefer(pruefer service.Pruefer) *Zeichenwerk {
	z.Bildausfuhr.MitPruefer(pruefer)
	z.Bogenleger.MitPruefer(pruefer)

	return z
}

// Bildschreibung sagt, was geschrieben wurde.
//
// **Ohne `service`-Typen**, damit eine Fläche ohne Fenster sie lesen kann,
// ohne `internal/service` zu kennen (§2): `internal/app` übersetzt zwischen
// den Flächen und den Diensten, und genau das ist hier zu tun.
type Bildschreibung struct {
	Pfad   string
	Breite int
	Hoehe  int

	// Befunde ist, was beim Zeichnen aufgefallen ist — als Sätze.
	//
	// **Sie fehlten bis zum 2026-08-25** (#1137).
	// Wer über die Kommandozeile, MCP oder REST zeichnete, bekam Pfad, Breite
	// und Höhe — und keinen einzigen Befund. Eine fehlende Schrift, ein
	// unlesbares Bild, eine Verknüpfung, die dieses Programm nicht kann: im
	// Fenster stand das alles neben der Karte, hier fiel es weg.
	//
	// **Als Sätze und nicht als `domain.Befund`**, aus demselben Grund wie die
	// drei Felder darüber: eine Fläche ohne Fenster soll sie lesen können,
	// ohne `internal/domain` zu kennen.
	Befunde []string
}

// SchreibeBild zeichnet eine Karte an einen genannten Ort.
//
// # Warum die Endung das Format bestimmt
//
// Ein eigener Schalter ließe `zeichnen 1 karte.svg --format png` zu — und
// schriebe ein PNG unter einem Namen, der SVG verspricht. Die Endung ist die
// Angabe, die ohnehin dasteht; zwei Angaben über dasselbe können sich
// widersprechen, eine nicht.
//
// **`ErrValidation`, wenn die Endung keine ist** — der Aufruf stimmt dann
// nicht, und eine Fläche kann das von „die Karte gibt es nicht" unterscheiden
// (§8).
func (z *Zeichenwerk) SchreibeBild(
	ctx context.Context, karteID int64, ziel string, guete render.Guete,
) (*Bildschreibung, error) {
	format, bekannt := bildformatAusEndung(ziel)
	if !bekannt {
		return nil, fmt.Errorf("%w: %q nennt kein bekanntes Bildformat — "+
			"erwartet ist .png, .jpg oder .svg",
			domain.ErrValidation, filepath.Base(ziel))
	}

	// **Verzeichnis und Datei getrennt**, weil die Ausfuhr den Dateinamen
	// prüft: er darf kein Pfad sein. Wer ihn erst dort zusammensetzte,
	// verließe sich darauf, dass jeder Aufrufer das tut.
	verzeichnis, datei := filepath.Split(ziel)
	if verzeichnis == "" {
		verzeichnis = "."
	}

	geschrieben, err := z.Bildausfuhr.Schreibe(
		ctx, karteID, verzeichnis, datei, format, 0, "ueberschreiben", guete)
	if err != nil {
		return nil, err
	}

	befunde := make([]string, 0, len(geschrieben.Befunde))
	for _, befund := range geschrieben.Befunde {
		befunde = append(befunde, befund.String())
	}

	return &Bildschreibung{
		Pfad: ziel, Breite: geschrieben.Breite, Hoehe: geschrieben.Hoehe,
		Befunde: befunde,
	}, nil
}

// Bogenlegung sagt, was auf Papier gelandet ist.
//
// **Ohne `service`-Typen**, aus demselben Grund wie [Bildschreibung]: eine
// Fläche ohne Fenster liest sie, ohne `internal/service` zu kennen (§2).
type Bogenlegung struct {
	Pfad   string
	Seiten int
	Karten int

	// Befunde ist, was beim Zeichnen der Karten aufgefallen ist — als Sätze.
	Befunde []string
}

// LegeBogen legt die Karten eines Sets als Druckbogen an einen genannten Ort.
//
// # Warum es diese Methode gibt, obwohl [App.BogenLegen] dasselbe tut
//
// Jene ist die **Bindung**: sie nimmt Zeichenketten entgegen, wählt ein
// Papierformat aus einer Liste, die die Oberfläche kennt, und gibt einen
// übersetzten Fehler zurück. Eine Fläche ohne Fenster braucht das Gegenteil —
// eine Kennung, einen Pfad und einen Go-Fehler
// (#1291).
//
// **Ohne Papierwahl und ohne Auswahl**: A4, neun Karten je Seite, 300 dpi, mit
// Schnittmarken — der Bogen, den man druckt. Erweitern geht später,
// zurücknehmen nicht
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
func (z *Zeichenwerk) LegeBogen(ctx context.Context, setID int64, ziel string) (
	*Bogenlegung, error,
) {
	vorgabe := bogenOhneFenster()
	masse, auswahl := vorgabe.masse, vorgabe.auswahl

	lauf, err := z.Bogenleger.Lege(ctx, setID, ziel, masse, auswahl)
	if err != nil {
		return nil, err
	}

	befunde := make([]string, 0, len(lauf.Befunde))
	for _, befund := range lauf.Befunde {
		befunde = append(befunde, befund.String())
	}

	return &Bogenlegung{
		Pfad: lauf.Pfad, Seiten: lauf.Seiten, Karten: lauf.Karten, Befunde: befunde,
	}, nil
}

// bildformatAusEndung nennt das Bildformat, das ein Zielname verlangt.
func bildformatAusEndung(ziel string) (service.Bildformat, bool) {
	endung := strings.ToLower(strings.TrimPrefix(filepath.Ext(ziel), "."))

	for _, format := range []service.Bildformat{
		service.FormatPNG, service.FormatJPG, service.FormatSVG,
	} {
		if string(format) == endung {
			return format, true
		}
	}

	return "", false
}

// schriftwurzel ist die Naht zum Ort der Paketschriften.
//
// **Eine eigene, und das ist der Befund** (#1170).
//
// Die Stilwurzeln sind **nicht** die Schriftwurzel. `msepaket.Paketwurzeln`
// gibt die Verzeichnisse, die Pakete *enthalten* — bei `M15-Magic-Pack` also
// `…/data`. Die Schriften liegen **daneben**, unter `Magic - Fonts/`, und
// fallen damit heraus. Im Betrieb fällt das nicht auf, weil beide unter
// derselben Wurzel im Zuhause liegen; in einem Lauf mit umgelenktem `HOME`
// schon.
//
// **Was das gekostet hat**: der Kanonlauf zeichnete seit dem 2026-08-26 jede
// Karte **ohne Text** und maß das als Abstand zu MSE2 — 28,362 statt 18,853,
// die Zerlegung 0 % `Text` statt 56 %. Das sah aus wie ein Rückschritt an der
// Rasterung und war einer an der Vorrichtung.
//
// **Und aus derselben Stelle wie die Pakete** — QA-10.NAT-01 (#1741). Hier
// stand `schriften.Paketwurzel`, die den Ort ein zweites Mal ausrechnete, in
// einem Paket, das von der Anwendung nichts weiß.
var schriftwurzel = paketschriftwurzel

// paketschriftwurzel gibt das Paketverzeichnis — **ohne es anzulegen**: die
// Schriftsuche fragt, sie richtet nicht ein.
func paketschriftwurzel() string {
	ort, fehler := paketablage(false)
	if fehler != nil {
		return ""
	}

	return ort
}

// Schriftwurzel gibt den Ort der Paketschriften.
func Schriftwurzel() string { return schriftwurzel() }

// MitSchriftwurzel setzt die Naht — **für Messläufe**.
//
// **Nicht für den Betrieb**: dort ist die Wurzel das Paketverzeichnis des
// Rechners, und wer sie verschöbe, suchte Schriften, die niemand dorthin
// gelegt hat.
func MitSchriftwurzel(ort func() string) (zurueck func()) {
	vorher := schriftwurzel
	schriftwurzel = ort

	return func() { schriftwurzel = vorher }
}
