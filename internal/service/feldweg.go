package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Feldweg entscheidet, **wohin ein Feldwert wirklich gehört** —
// #696.
//
// # Das Feld, das es gar nicht gibt
//
// `magic.mse-game` erklärt `text` mit `save value: false` und rechnet es aus
// zwei anderen aus:
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>",
//	                field2: card.flavor_text)
//
// Wer in dieses Feld tippte, schrieb eine Zeile `text:` in den Bestand — und
// sah beim nächsten Zeichnen wieder den alten Regeltext, weil die Vorlage `text`
// nie liest. Die Eingabe war weg, ohne dass jemand einen Fehler bekommen hätte.
//
// # Was dieser Weg tut
//
// Er sieht vor jedem Schreiben nach, ob das Spiel dieses Feld überhaupt
// speichert. Tut es das nicht, gibt es zwei Ausgänge und keinen dritten:
//
//   - Das Feld ist ein umkehrbarer Verbund — dann werden **seine Teile**
//     geschrieben, am Trenner gespalten ([msescript.Verbund]).
//   - Sonst wird nichts geschrieben und der Aufrufer bekommt gesagt, warum.
//
// # Warum hier und nicht in der Fassade
//
// Weil es eine fachliche Entscheidung über mehrere Dinge ist — Karte, Spiel,
// Bestand (§2). Die Fassade übersetzt sie nur.
type Feldweg struct {
	karten Kartensetleser
	felder Kartenfeldquelle
	ablage Feldablage

	// nachzug rechnet, was gespeichert wird — wahlfrei (#2396,
	// `feldwegnachzug.go`).
	nachzug *feldnachzug
}

// Kartensetleser sagt, zu welchem Set eine Karte gehört.
type Kartensetleser interface {
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)
}

// Kartenfeldquelle gibt die Kartenfelder des Spiels hinter einem Set.
type Kartenfeldquelle interface {
	Kartenfelder(ctx context.Context, setID int64) ([]domain.Field, error)
}

// Feldablage schreibt Feldwerte und gibt den vorigen zurück.
type Feldablage interface {
	FeldSetzenMitVorwert(ctx context.Context, karteID int64, feld, wert string) (
		vorher string, hatte bool, fehler error,
	)

	// FelderSetzenMitVorwert schreibt **mehrere** Felder einer Karte in einer
	// Transaktion und gibt den vorigen Wert des ersten.
	//
	// **Alles oder nichts** (#714):
	// der Rückweg eines Verbundes schrieb Teil für Teil, und jeder Aufruf
	// eröffnete seine eigene Transaktion — scheiterte Teil zwei, war Teil eins
	// schon bestätigt.
	FelderSetzenMitVorwert(ctx context.Context, karteID int64, werte []domain.Value) (
		vorher string, hatte bool, fehler error,
	)

	// Die beiden Fassungen mit Herkunft — **je eine für jeden Ausgang dieses
	// Wegs** (#888).
	//
	// **Zwei Methoden statt eines Bauers**, und der Grund ist die
	// Schichtenregel: ein Bauer müsste den Schnittstellentyp zurückgeben, und
	// `internal/storage` darf `internal/service` nicht kennen (§2). Ein
	// Adapter im Composition Root nur dafür wäre Umstand — die Herkunft
	// gehört ohnehin zur **Änderung** und nicht zum Schreiber.
	//
	// **Leer heißt: von hier.**
	FeldSetzenVonAussen(ctx context.Context, karteID int64, feld, wert, herkunft string) (
		vorher string, hatte bool, fehler error,
	)

	// FelderSetzenMitHerkunft schreibt Werte **und ihre Herkunft in einem
	// Zug** — Werkstück A11 (#1450).
	//
	// **Neben `FelderSetzenVonAussen` und nicht statt seiner**: die meisten
	// Schreibwege haben keine Herkunft, die über „von dort" hinausginge.
	// Eine Übernahme aus einem KI-Vorschlag hat sie, und sie gehört in
	// dieselbe Transaktion wie die Werte.
	// **`erwartet` ist der Stand, auf dem die Karte stehen muss** — leer
	// heißt: ohne Bedingung, und geprüft wird er in **derselben**
	// Transaktion (Befund 2 des Zweitblicks über Phase 5.3).
	FelderSetzenMitHerkunft(
		ctx context.Context, karteID int64, werte []domain.Value,
		herkunft string, herkuenfte []domain.KIFeldherkunft,
		angaben map[string][]domain.KIBildparameter, erwartet string,
	) (vorher string, hatte bool, fehler error)

	// FelderSetzenVonAussen schreibt eine Reihe — **auf dem Stand `erwartet`**,
	// aus dem sie gerechnet ist; leer heißt ohne Bedingung (Gesamtreview über
	// 0.9.0, Befund 2).
	FelderSetzenVonAussen(
		ctx context.Context, karteID int64, werte []domain.Value, herkunft, erwartet string,
	) (vorher string, hatte bool, fehler error)

	// GruppeSetzenMitHerkunft schreibt mehrere Karten in **einer**
	// Transaktion (A13.1b, #1452).
	GruppeSetzenMitHerkunft(ctx context.Context, schritte []domain.Kartenschritt) error

	// FelderGerechnetSetzen schreibt je Karte ihre gerechneten Werte in
	// **einem** Schritt, mit der Erwartung aus der Vorschau und dem Stand je
	// Karte (#2516).
	FelderGerechnetSetzen(ctx context.Context, schritte []domain.Kartenschritt,
		erwartet []domain.Feldvorwert, feld string) (vorgang string, fehler error)
}

// NeuerFeldweg baut ihn.
func NeuerFeldweg(
	karten Kartensetleser, felder Kartenfeldquelle, ablage Feldablage,
) *Feldweg {
	return &Feldweg{karten: karten, felder: felder, ablage: ablage}
}

// FeldSetzenMitVorwert schreibt — dorthin, wo der Wert hingehört.
//
// **Die Signatur ist die der Ablage**, damit dieser Weg an ihre Stelle treten
// kann, ohne dass jeder Aufrufer davon weiß.
//
// **Ein Fehler beim Nachsehen hält nichts auf.** Wer das Spiel nicht lesen
// kann, weiß nicht, ob ein Feld gespeichert wird — und ein Bestand, der beim
// Tippen stumm bleibt, weil ein Paket fehlt, wäre schlechter als einer, der den
// Wert wie bisher ablegt (§8).
func (f *Feldweg) FeldSetzenMitVorwert(
	ctx context.Context, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	return f.setzen(ctx, karteID, feld, wert, "")
}

// FeldSetzenVonAussen schreibt und hält fest, **woher der Wert kam**
// (#888).
//
// **Derselbe Rumpf wie die Methode darüber**, und das ist der Punkt: ein
// übernommener Wert geht durch **denselben** Weg wie ein getippter — er wird
// genauso in die Teile eines Verbunds zerlegt, und ein nicht speicherbares
// Feld wird genauso abgewiesen. Eine zweite Fassung liefe auseinander.
func (f *Feldweg) FeldSetzenVonAussen(
	ctx context.Context, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	return f.setzen(ctx, karteID, feld, wert, herkunft)
}

// FeldSetzenMitErgebnis schreibt wie [Feldweg.FeldSetzenMitVorwert] und
// sagt, **was im Feld gespeichert ist** — seit #2396 der gerechnete Wert. Das
// Fenster bestätigt seinen Entwurf mit genau ihm und nicht mit der nächsten
// Änderung des Bestands (#2535). **`bekannt` ist falsch für ein Verbundfeld**:
// es steht nicht selbst im Bestand, seine Teile tun es.
func (f *Feldweg) FeldSetzenMitErgebnis(
	ctx context.Context, karteID int64, feld, wert string,
) (gespeichert string, bekannt bool, fehler error) {
	_, _, gespeichert, bekannt, fehler = f.setzenMitErgebnis(ctx, karteID, feld, wert, "")

	return gespeichert, bekannt, fehler
}

// setzen ist der gemeinsame Rumpf.
func (f *Feldweg) setzen(
	ctx context.Context, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	vorher, hatte, _, _, fehler = f.setzenMitErgebnis(ctx, karteID, feld, wert, herkunft)

	return vorher, hatte, fehler
}

// setzenMitErgebnis ist der Rumpf, der auch den gespeicherten Wert nennt.
func (f *Feldweg) setzenMitErgebnis(
	ctx context.Context, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, gespeichert string, bekannt bool, fehler error) {
	// **Einmal gelesen und dreimal gebraucht.** Die Felder sagen, ob dieses
	// Feld gespeichert wird, und danach, wie die Teile eines Verbunds wirklich
	// heißen; ein zweiter Aufruf ginge bei jedem Tastendruck erneut ans Paket.
	karte, felder, lesefehler := f.karteUndFelder(ctx, karteID)

	// **„Die Vorlage ließ sich nicht lesen" ist nicht „das Feld gibt es
	// nicht"** — #720,
	// angestrichen vom Zweitblick durch Codex am 2026-08-21.
	//
	// Beide Fälle enden hier gleich: der Wert geht roh in den Bestand. Der
	// Unterschied steht trotzdem im Code, weil er im Kopf sonst verschwindet —
	// und der Preis gehört ausgesprochen.
	//
	// **Warum nicht abgewiesen wird.** Ohne die Vorlage weiß niemand, ob dieses
	// Feld berechnet ist. Wer deshalb **gar nicht** schriebe, machte jedes Set
	// unbearbeitbar, dessen Spielpaket nicht installiert ist — ein Set aus MTG
	// JSON ist genau das, und es ist ein Weg, den diese Anwendung ausdrücklich
	// anbietet. Der schlechtere Tausch.
	//
	// **Warum es trotzdem tragbar ist.** Über das Fenster kann dieser Fall ein
	// berechnetes Feld nicht treffen: es bietet an, was `Kartenfelder`
	// hergibt, und ohne Vorlage gibt es nichts her (#696, #687). Bleibt der
	// MCP-Server, und dort ist der rohe Wert derselbe, den es vor #696 überall
	// gab — die Vorlage liest ihn nicht, verliert aber auch nichts.
	//
	// **Was fehlt, ist die Sichtbarkeit**: `SchreibAntwort` trägt einen Fehler
	// und keine Vorbehalte. Wer das ändern will, braucht einen zweiten Kanal
	// auf dem Schreibweg — und der gehört nicht in ein Werkstück über
	// berechnete Felder.
	if lesefehler != nil {
		vorher, hatte, fehler = f.ablage.FeldSetzenVonAussen(ctx, karteID, feld, wert, herkunft)

		return vorher, hatte, wert, true, fehler
	}

	teile, fehler := aufloesen(felder, feld, wert)
	verbund := teile != nil

	switch {
	case fehler != nil:
		return "", false, "", false, fehler

	case teile == nil && f.nachzug == nil:
		vorher, hatte, fehler = f.ablage.FeldSetzenVonAussen(ctx, karteID, feld, wert, herkunft)

		return vorher, hatte, wert, true, fehler

	case teile == nil:
		teile = []domain.Value{{Field: feld, Text: wert}}
	}

	// **Gespeichert wird, was gerechnet ist** — samt dem, was es liest (#2396),
	// und auf dem Stand, aus dem gerechnet wurde.
	werte, stand := f.nachzug.nachgezogen(ctx, karte, teile)

	vorher, hatte, fehler = f.ablage.FelderSetzenVonAussen(ctx, karteID, werte, herkunft, stand)
	if verbund {
		return vorher, hatte, "", false, fehler
	}

	// **Das gesetzte Feld steht vorn** (`nachgezogen`).
	return vorher, hatte, werte[0].Text, true, fehler
}

// aufloesen sagt, in **welche** Felder ein Wert wirklich gehört — oder nichts,
// wenn er in sein eigenes gehört.
//
// # Warum das eine eigene Frage ist
//
// Sie wird an zwei Stellen gebraucht: beim Schreiben eines Feldes und beim
// Schreiben mehrerer in einem Zug
// (#1026).
// Zwei Fassungen liefen auseinander — und die zweite fiele niemandem auf, weil
// beide grün wären.
//
// **`nil` heißt „kein Verbund"**, nicht „nichts zu tun": der Wert geht dann
// roh in sein eigenes Feld.
//
// **Die unlesbare Vorlage bleibt beim Aufrufer.** Sie ist keine Frage der
// Auflösung, sondern eine Duldung — ausführlich begründet bei
// [Feldweg.setzen] —, und ein hier verschluckter Fehler sähe genauso aus wie
// ein übersehener.
func aufloesen(felder []domain.Field, feld, wert string) ([]domain.Value, error) {
	form, da := form(felder, feld)
	if !da || form.Gespeichert {
		return nil, nil
	}

	verbund, umkehrbar := msescript.VerbundAus(form.Skript)
	if !umkehrbar {
		return nil, nichtSpeicherbar(felder, form, feld)
	}

	return verbundwerte(felder, verbund, wert), nil
}

// nichtSpeicherbar sagt, warum dieses Feld keine Eingabe annimmt — und wohin
// sie stattdessen gehört.
//
// **Der Satz nennt die Felder, wenn er sie kennt.** `type` und `pt` rechnen
// ihren Trenner aus (`language().type_separator`); spalten lässt sich ihr Wert
// deshalb nicht, aber die Felder dahinter stehen im Aufruf. „Das geht nicht"
// ohne das „sondern" ist eine Sackgasse — und über den MCP-Server ist dieser
// Weg die einzige Auskunft, die ein Gegenüber bekommt
// (#687).
//
// **Das Fenster erreicht ihn nicht mehr**: es bietet an der Stelle eines
// solchen Feldes seine Teile an ([Teilfelder]).
func nichtSpeicherbar(felder []domain.Field, form domain.Field, feld string) error {
	teile := Teilfelder(felder, form)
	if len(teile) == 0 {
		return domain.NewFieldError(feld, domain.SchluesselValidierung,
			"das Feld %q wird von der Vorlage berechnet und nicht gespeichert",
			feld)
	}

	return domain.NewFieldError(feld, domain.SchluesselValidierung,
		"das Feld %q entsteht aus %s — geändert werden diese",
		feld, strings.Join(teile, " und "))
}

// verbundwerte spaltet den Wert und ordnet jedem Teil sein Feld zu.
//
// **Geschrieben wird danach in einer Transaktion und nicht Teil für Teil**
// (#714).
//
// Hier stand eine Schleife über einzelne Schreibaufrufe, und jeder eröffnete
// seine eigene Transaktion: scheiterte Teil zwei, war Teil eins schon
// bestätigt. Bei `text` einer M15-Karte heißt das, der Regeltext steht neu da
// und der Flavortext alt — eine halb geänderte Karte, und niemand kann sagen,
// welche Hälfte. Angestrichen vom Zweitblick durch Codex am 2026-08-21.
//
// **Der Vorwert ist der des ersten Teils.** Er geht an den Verlauf des
// Aufrufers; der Bestand trägt je Feld einen eigenen Eintrag, und seit #2396
// nimmt ein ⌘Z sie zusammen zurück.
func verbundwerte(
	felder []domain.Field, verbund msescript.Verbund, wert string,
) []domain.Value {
	teile := verbund.Zerlege(wert)

	werte := make([]domain.Value, 0, len(verbund.Felder))
	for stelle, name := range verbund.Felder {
		werte = append(werte, domain.Value{
			Field: feldname(felder, name), Text: teile[stelle],
		})
	}

	return werte
}

// karteUndFelder liest die Karte und die Kartenfelder des Spiels dahinter —
// die Karte braucht der Nachzug (#2396), damit sie nicht zweimal gelesen wird.
//
// **Der Lesefehler kommt mit** (#720). Er sagt etwas anderes als eine leere
// Liste: „ich konnte nicht nachsehen" gegen „ich habe nachgesehen, und das Feld
// steht nicht darin". Der Aufrufer entscheidet, was daraus folgt.
func (f *Feldweg) karteUndFelder(ctx context.Context, karteID int64) (
	*domain.CardRow, []domain.Field, error,
) {
	karte, err := f.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, nil, err
	}

	// **Eine Karte, die es nicht gibt, ist kein Lesefehler.** Sie ist eine
	// Antwort: das Feld gehört zu nichts, und der Bestand sagt es gleich
	// darauf mit seiner eigenen Meldung.
	if karte == nil {
		return nil, nil, nil
	}

	felder, err := f.felder.Kartenfelder(ctx, karte.SetID)
	if err != nil {
		return nil, nil, err
	}

	return karte, felder, nil
}
