package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kartenschreiber ist, was `app` zum Ändern eines Feldwerts braucht.
//
// # Warum `FeldSetzenMitVorwert` und nicht das schlichte `FeldSetzen`
//
// **Weil sonst der Verlauf des Fensters fehlt.** Nur dieser Weg schreibt seinen
// Verlaufseintrag in derselben Transaktion (#52); das schlichte Schreiben tut es
// nicht. Bis hierher ging das Fenster daran vorbei — und ⌘Z hätte alles
// zurücknehmen können außer dem, was jemand getippt hat. Der MCP-Server schrieb
// den Verlauf, das Fenster nicht.
//
// Den Vorwert braucht `app` nicht; er kommt zurück, weil der Bestand ihn in
// derselben Transaktion ohnehin liest.
type Kartenschreiber interface {
	FeldSetzenMitVorwert(ctx context.Context, setID, karteID int64, feld, wert string) (
		vorher string, hatte bool, fehler error,
	)

	// FeldSetzenVonAussen hält zusätzlich fest, **woher der Wert kam**
	// (#888).
	//
	// **Leer heißt: von hier** — dann verhält sie sich wie die Methode
	// darüber. Sie steht daneben und ersetzt sie nicht: eine Tipparbeit hat
	// keine Herkunft, und ein Aufruf, der eine leere Zeichenkette
	// mitschleifte, wäre eine Einladung, versehentlich eine falsche zu setzen.
	FeldSetzenVonAussen(
		ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
	) (vorher string, hatte bool, fehler error)

	// FelderVonAussen schreibt mehrere übernommene Werte in **einem** Zug
	// (#1163,
	// Befund Z7-02).
	//
	// **Ein Zug und nicht zwei Aufrufe.** Ein Bild von Scryfall braucht sein
	// Bildfeld **und** seinen Künstler; zwischen zwei Aufrufen liegt ein
	// Stand, in dem die Karte fremde Kunst ohne ihre Nennung trägt — genau
	// den, den die Bedingung von Scryfall ausschließt.
	// FelderVonAussenMitHerkunft schreibt zusätzlich, **woher jeder Wert
	// kommt** — Werkstück A11 (#1450).
	//
	// **Neben `FelderVonAussen` und nicht statt seiner**: ein Bild von
	// Scryfall hat eine Herkunft, die in ein Wort passt; ein übernommener
	// KI-Vorschlag hat fünf Dinge zu sagen, und sie gehören in dieselbe
	// Transaktion wie die Werte.
	FelderVonAussenMitHerkunft(
		ctx context.Context, setID, karteID int64, werte []domain.Value,
		herkunft string, herkuenfte []domain.KIFeldherkunft,
		angaben map[string][]domain.KIBildparameter, erwartet string,
	) (vorher string, hatte bool, fehler error)

	// GruppeVonAussenMitHerkunft schreibt **mehrere Karten in einem Zug** —
	// Werkstück A13.1b (#1452).
	//
	// **Und nicht die Methode darüber in einer Schleife**: die wären N
	// Transaktionen, und zwischen der dritten und der vierten stünde ein
	// Bestand, in dem die halbe Gruppe den Vorschlag trägt. Das Issue nennt
	// es „klarer Konflikt **ohne Teilbestand**".
	GruppeVonAussenMitHerkunft(
		ctx context.Context, schritte []domain.Kartenschritt) error

	FelderVonAussen(
		ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
	) (vorher string, hatte bool, fehler error)
}

// Kartenleser gibt eine einzelne Karte.
type Kartenleser interface {
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)

	// Herkuenfte gibt je Feld die Adresse, aus der sein **aktueller** Wert
	// stammt (#1025).
	//
	// **An derselben Naht wie die Karte**, und das ist Absicht: wer eine Karte
	// liest, liest damit, woher ihre Werte kommen. Eine eigene, nachträglich
	// gesetzte Naht wäre bei jedem Aufrufer, der sie vergisst, still aus — und
	// dann stünde ein Bildausschnitt ohne seine Nennung da.
	Herkuenfte(ctx context.Context, karteID int64) (map[string]string, error)
}

// Planbauer macht aus einer Karte einen Zeichenplan.
//
// **`Kartenplan` und nicht `Behelfsplan`** (#206): der Behelf ist der Rückfall,
// nicht der Weg. Wer hier den Behelf anmeldete, bekäme ihn auch dann, wenn ein
// Stil dagewesen wäre — und genau so war es bis zum Zweitblick über Phase 6.
type Planbauer interface {
	Kartenplan(ctx context.Context, karte *domain.CardRow) (*render.Plan, []domain.Befund, error)

	// Kartenpalette gibt, was ein Feld zum Einfügen anbietet — samt der
	// Auskunft, ob ein Code roh oder in `<sym>…</sym>` eingefügt wird
	// (#564).
	Kartenpalette(ctx context.Context, karte *domain.CardRow, feld string) (
		[]service.Paletteneintrag, bool, []domain.Befund, error,
	)

	// **Und die Auswahllisten** (#1074) — [Wortlistenbauer], eingebettet und
	// nicht hier ausgeschrieben: `karteneditor.go` stand an der Zeilengrenze.
	Wortlistenbauer

	// **Und die Wahlbilder der Stiloptionen** (#1225) — [Wahlbildbauer], aus
	// demselben Grund eingebettet.
	Wahlbildbauer
}

// VorschauDTO ist ein gerastertes Kartenbild samt seinen Vorbehalten.
type VorschauDTO struct {
	// Bild ist die **Adresse** des PNG — unmittelbar in ein `src` einsetzbar.
	//
	// **Und seit #1270
	// keine Data-URI mehr.** Sie war eine: 182 KB PNG werden als Base64 zu
	// 242 KB, und die wandern als JSON-Zeichenkette durch die Wails-Naht.
	// Gemessen im laufenden Fenster am 2026-08-28 kostete das **554 ms**,
	// während der ganze Go-Weg — Planbau, Rasterung, Packen — **53 ms**
	// braucht. Über die Größe: 86 KB → 397 ms, 268 KB → 873 ms,
	// 694 KB → 1.648 ms.
	//
	// Jetzt trägt sie einen Weg unter [Bildwegpfad]; die Bytes holt der
	// Browser über den Assetserver, den Wails ohnehin betreibt. Warum das so
	// und nicht anders geht, steht in `bildweg.go`.
	Bild string `json:"bild"`

	Breite int `json:"breite"`
	Hoehe  int `json:"hoehe"`

	// DPI ist die Auflösung der Vorlage — mit ihr rechnet das Fenster die
	// Druckgröße (U5, #1416); 0 heißt: die Vorlage nennt keine.
	DPI int `json:"dpi"`

	// Befunde sind die Ebenen, die nicht gezeichnet werden konnten. Sie stehen
	// hier, weil eine Vorschau mit elf von zwölf Ebenen brauchbar ist und eine
	// leere ohne Meldung nicht (§8).
	Befunde []string `json:"befunde"`

	// Nennungen sind die Sätze, die neben diesem Bild stehen müssen.
	//
	// **Sie reisen mit der Vorschau**, weil die Vorschau die Oberfläche ist,
	// die den Bildausschnitt zeigt — auch dann noch, wenn der Zuschnittkasten
	// längst zu ist. Scryfall verlangt genau das, nachgelesen am 2026-08-23
	// unter `https://scryfall.com/docs/api`: „list the artist name and
	// copyright elsewhere in the same interface presenting the art crop".
	//
	// Bis #1025
	// stand die Nennung **nur** am offenen Zuschnitt. Eine Bedingung, die das
	// Schließen eines Kastens nicht überlebt, ist keine.
	Nennungen []NennungDTO `json:"nennungen"`

	// Flaechen sagen, welches Feld wo im Bild liegt — **in Bildpunkten**.
	//
	// AGENTS.md §6: die Karte ist das Hauptobjekt. Eine Vorschau, in der sich
	// nichts anklicken lässt, ist eine Abbildung des Hauptobjekts, nicht das
	// Hauptobjekt (#246).
	Flaechen []FeldflaecheDTO `json:"flaechen"`

	// Vorgaben sind die **gerechneten** Feldwerte — `default:`, `initial:`
	// oder die erste `choice:`, für Felder, die die Karte nicht selbst führt
	// (#563).
	//
	// **Sie stehen an der Vorschau und nicht an `Kartenfelder`**, weil sie zur
	// **Karte** gehören: ein `default:` ist ein Skript über `card`, und
	// dieselbe Vorlage gibt für zwei Karten zwei Werte.
	Vorgaben []VorgabeDTO `json:"vorgaben"`

	// Stil ist der Wert `stylesheet`, mit dem die Vorschau gezeichnet ist —
	// getrimmt, leer heißt „der des Sets". **Das Formular legt die Flächen
	// unter ihm ab** (Zweitblick über die Folgen vor 0.8.0, Befund 8): eine
	// Vorschau, die noch den vorigen Stil zeigt oder einen nur getippten, gehört
	// nicht dem, der gerade geschrieben steht.
	Stil string `json:"stil"`

	// Gezeichnet ist das Paket, mit dem der Plan gebaut ist — leer beim
	// Behelfsplan (#2314). **Nicht dasselbe wie [VorschauDTO.Stil]**: eine
	// Karte, die den Stil des Sets erbt, schreibt `''`, und welches Paket
	// das ist, ändert ein fremder Setstilwechsel.
	Gezeichnet string `json:"gezeichnet"`
}

// VorgabeDTO ist ein gerechneter Feldwert.
type VorgabeDTO struct {
	Feld string `json:"feld"`
	Wert string `json:"wert"`
}

// FeldSetzen schreibt einen Feldwert.
//
// **Der Feldname wird geprüft, der Wert nicht.** Ein leerer Feldname ergäbe
// eine Zeile, die niemand wiederfindet; ein leerer Wert ist ein gültiger Wert —
// wer ein Feld leert, meint das.
func (a *App) FeldSetzen(
	ctx context.Context, setID, karteID, feld, wert string,
) FeldAntwort {
	ctx = a.imProfil(ctx)

	// **Beide Kennungen, und das Set zuerst** (#808). Wer ein unbrauchbares
	// Set schickt, soll das erfahren, bevor er über die Karte nachdenkt.
	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return FeldAntwort{Fehler: befund}
	}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return FeldAntwort{Fehler: befund}
	}

	// **Die Regel steht in der Domäne** (#89, §2). Hier stand sie einmal
	// ausgeschrieben — und der MCP-Server ging daran vorbei, weil er diese
	// Fassade nicht nimmt. Sie bleibt hier gerufen, damit ein leerer Name gar
	// nicht erst zum Bestand geht; **entschieden** wird sie dort.
	if befund := domain.PruefeFeldname(feld); befund != nil {
		return FeldAntwort{Fehler: Uebersetze(ctx, befund)}
	}

	// **Wer sagt, was er gespeichert hat, sagt es dem Fenster** (#2535).
	if mit, sagt := a.schreiber.(Feldergebnisschreiber); sagt {
		gespeichert, bekannt, err := mit.FeldSetzenMitErgebnis(ctx, setkennung, kennung, feld, wert)
		if err != nil || !bekannt {
			return FeldAntwort{Fehler: Uebersetze(ctx, err)}
		}

		return FeldAntwort{Gespeichert: &gespeichert}
	}

	_, _, err := a.schreiber.FeldSetzenMitVorwert(ctx, setkennung, kennung, feld, wert)

	return FeldAntwort{Fehler: Uebersetze(ctx, err)}
}

// Feldergebnisschreiber sagt zusätzlich, **was gespeichert ist** — die
// Kartenaufsicht über den Feldweg (#2535). Wahlfrei wie dort.
type Feldergebnisschreiber interface {
	FeldSetzenMitErgebnis(ctx context.Context, setID, karteID int64, feld, wert string) (
		gespeichert string, bekannt bool, fehler error)
}

// FeldAntwort ist die Antwort auf ein geschriebenes Kartenfeld (#2535).
//
// **`Gespeichert` ist der Wert, der danach im Feld steht** — seit #2396 der
// gerechnete. Der Entwurf im Fenster bestätigt genau ihn, und ein fremder
// Wert, der vor dem Nachladen kommt, bleibt ein Konflikt. **`nil`, wo es
// niemand weiß**: ein Verbund steht nicht selbst im Bestand, und ein
// Schreiber ohne Rechnung behauptet keinen Wert. Neben [SchreibAntwort] und
// nicht in ihr, damit der Vertrag der übrigen Bindungen bleibt.
type FeldAntwort struct {
	Fehler      *Fehler `json:"fehler"`
	Gespeichert *string `json:"gespeichert"`
}

// SchreibAntwort trägt nur einen Fehler — oder keinen.
type SchreibAntwort struct {
	Fehler *Fehler `json:"fehler"`
}
