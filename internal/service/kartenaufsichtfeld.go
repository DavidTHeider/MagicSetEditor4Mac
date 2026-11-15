// Diese Datei bringt den Feldweg unter dieselbe Setgrenze wie den Lebenszyklus.
//
// **Eine eigene Datei, weil sie eine eigene Naht hat**: der Lebenszyklus geht
// an den Bestand, der Feldweg an `Feldweg` und den Massenschreiber. Getrennt
// gehalten heißt auch, dass ein Aufrufer, der nur den einen braucht, nicht den
// anderen mitnehmen muss.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Feldschreiber ist das Setzen eines einzelnen Feldes samt seinen
// Rücknahmewegen.
//
// **Alle drei, nicht nur das Setzen.** `FeldEntfernen` und
// `FeldSetzenAnStelle` schreiben genauso; eine Aufsicht, die nur den
// Hinweg hält, lässt den Rückweg offen.
type Feldschreiber interface {
	FeldSetzenMitVorwert(ctx context.Context, karteID int64, feld, wert string) (
		vorher string, hatte bool, fehler error)
	FeldEntfernen(ctx context.Context, karteID int64, feld string) (
		vorher string, hatte bool, stelle *int64, fehler error)
	FeldSetzenAnStelle(ctx context.Context, karteID int64, feld, wert string,
		stelle int64) (vorher string, hatte bool, fehler error)

	// FeldSetzenVonAussen setzt ein Feld und hält fest, woher der Wert kam.
	//
	// **Die Herkunft hängt am Aufruf, nicht am Schreiber** — dasselbe Fenster
	// schreibt getippte Werte und übernommene, im Wechsel. Leer heißt: von
	// hier.
	FeldSetzenVonAussen(ctx context.Context, karteID int64, feld, wert, herkunft string) (
		vorher string, hatte bool, fehler error)

	// FelderVonAussen setzt mehrere Felder in **einem** Zug.
	//
	// **Nicht dasselbe wie die Methode darüber in einer Schleife**
	// (#1026):
	// eine Schleife kann halb durch sein, dieser Weg nicht.
	FelderVonAussen(
		ctx context.Context, karteID int64, werte []domain.Value, herkunft string,
	) (vorher string, hatte bool, fehler error)

	// FelderVonAussenMitHerkunft setzt mehrere Felder **und schreibt, woher
	// sie kommen** — in derselben Transaktion (A11, #1450).
	// **`erwartet` ist der Stand, auf dem die Karte stehen muss** — leer
	// heißt: ohne Bedingung, und geprüft wird er in **derselben**
	// Transaktion (Befund 2 des Zweitblicks über Phase 5.3).
	//
	// **Dieselbe Gestalt wie [Massenfeldschreiber.FelderSetzenMitErwartung]**
	// (U6, #1417) — dort für den Altwert je Karte, hier für den Stand der
	// einen. Eine Prüfung vor dem Schreiben ist eine Auskunft; dieselbe
	// Prüfung in der Transaktion ist eine Bedingung.
	FelderVonAussenMitHerkunft(
		ctx context.Context, karteID int64, werte []domain.Value,
		herkunft string, herkuenfte []domain.KIFeldherkunft,
		angaben map[string][]domain.KIBildparameter, erwartet string,
	) (vorher string, hatte bool, fehler error)

	// GruppeSetzenMitHerkunft schreibt **mehrere Karten in einem Zug**
	// (A13.1b, #1452).
	//
	// **Nicht die Methode darüber in einer Schleife**: die wären N
	// Transaktionen, und dazwischen stünde ein Bestand, in dem die halbe
	// Gruppe den Vorschlag trägt.
	GruppeSetzenMitHerkunft(ctx context.Context, schritte []domain.Kartenschritt) error
}

// Massenfeldschreiber ist dasselbe für viele Karten auf einmal.
type Massenfeldschreiber interface {
	FelderSetzen(ctx context.Context, karteIDs []int64, feld, wert string) (
		vorgang string, fehler error)

	// FelderSetzenMitErwartung schreibt nur, wenn jede Karte noch auf dem
	// Altwert steht, den die Vorschau gezeigt hat (U6, #1417) — sonst ein
	// [domain.ErrConflict], der die abweichenden Karten nennt.
	FelderSetzenMitErwartung(ctx context.Context, erwartet []domain.Feldvorwert, feld, wert string) (
		vorgang string, fehler error)
}

// MitFeldweg hängt der Aufsicht die beiden Feldwege an.
//
// **Nachträglich und nicht im Bauwerk**: der MCP-Server bekam den Lebenszyklus
// zuerst, und ein zweiter Erzeuger mit vier Parametern hätte jeden Aufrufer zu
// einer Angabe gezwungen, die er nicht braucht. Wer nur den Lebenszyklus
// beaufsichtigen lässt, ruft `MitFeldweg` nicht — und bekommt beim Feldweg
// einen Zeigerfehler statt einer stillen Umgehung.
func (a *Kartenaufsicht) MitFeldweg(
	feld Feldschreiber, massen Massenfeldschreiber,
) *Kartenaufsicht {
	a.feld = feld
	a.massen = massen

	return a
}

// MitFeldformen hängt der Aufsicht die Vorlage an — für [Kartenaufsicht.FelderSetzen].
//
// **Nachträglich und wahlfrei**, aus demselben Grund wie [Kartenaufsicht.MitFeldweg]:
// wer die Massenänderung nicht anbietet, braucht sie nicht.
func (a *Kartenaufsicht) MitFeldformen(formen Kartenfeldquelle) *Kartenaufsicht {
	a.formen = formen

	return a
}

// FeldSetzenMitVorwert setzt ein Feld, wenn die Karte in dieses Set gehört.
func (a *Kartenaufsicht) FeldSetzenMitVorwert(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return "", false, fehler
	}

	return a.feld.FeldSetzenMitVorwert(ctx, karteID, feld, wert)
}

// Feldergebnisschreiber sagt zusätzlich, **was gespeichert ist** — der
// [Feldweg] tut es seit #2535. Wahlfrei, damit ein Schreiber ohne Rechnung
// nicht einen Wert behaupten muss, den er nicht kennt.
type Feldergebnisschreiber interface {
	FeldSetzenMitErgebnis(ctx context.Context, karteID int64, feld, wert string) (
		gespeichert string, bekannt bool, fehler error)
}

// FeldSetzenMitErgebnis setzt ein Feld wie [Kartenaufsicht.FeldSetzenMitVorwert]
// und nennt den gespeicherten Wert, wo der Schreiber ihn kennt (#2535).
func (a *Kartenaufsicht) FeldSetzenMitErgebnis(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (gespeichert string, bekannt bool, fehler error) {
	if fremd := a.gehoertZu(ctx, setID, karteID); fremd != nil {
		return "", false, fremd
	}

	if mit, sagt := a.feld.(Feldergebnisschreiber); sagt {
		return mit.FeldSetzenMitErgebnis(ctx, karteID, feld, wert)
	}

	_, _, fehler = a.feld.FeldSetzenMitVorwert(ctx, karteID, feld, wert)

	return "", false, fehler
}

// FeldSetzenVonAussen setzt ein Feld und hält fest, **woher der Wert kam**.
//
// # Warum die Herkunft am Aufruf hängt und nicht am Schreiber
//
// Sie gehört zur **Änderung**, nicht zur Identität dessen, der schreibt: das
// Fenster schreibt getippte Werte und übernommene, im Wechsel. Ein Schreiber
// „von Scryfall" müsste zwischen beiden umgeschaltet werden, und wer das
// vergisst, hängt eine falsche Herkunft an eine Tipparbeit.
//
// **§16, Regel 6**: „Was von außen in ein Kartenfeld will, zeigt vorher, was
// sich ändert, und trägt **danach seine Herkunft**."
func (a *Kartenaufsicht) FeldSetzenVonAussen(
	ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return "", false, fehler
	}

	return a.feld.FeldSetzenVonAussen(ctx, karteID, feld, wert, herkunft)
}

// FelderVonAussen setzt mehrere Felder, wenn die Karte in dieses Set gehört.
//
// **Die Aufsicht fragt einmal**, nicht je Feld: es ist eine Karte, und ihre
// Zugehörigkeit ändert sich zwischen zwei Feldern desselben Zuges nicht.
func (a *Kartenaufsicht) FelderVonAussen(
	ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return "", false, fehler
	}

	return a.feld.FelderVonAussen(ctx, karteID, werte, herkunft)
}

// FelderVonAussenMitHerkunft setzt mehrere Felder samt ihrer Herkunft, wenn
// die Karte in dieses Set gehört (A11, #1450).
//
// **Dieselbe Zugehörigkeitsprüfung wie oben**: eine Übernahme aus einem
// KI-Vorschlag darf so wenig in ein fremdes Set schreiben wie jede andere.
// **Die Zugehörigkeit zuerst.** Eine Karte aus einem fremden Set wird
// abgewiesen, bevor irgendetwas gelesen oder geschrieben wird; welcher Stand
// erwartet wird, ist dann ohne Belang.
func (a *Kartenaufsicht) FelderVonAussenMitHerkunft(
	ctx context.Context, setID, karteID int64, werte []domain.Value,
	herkunft string, herkuenfte []domain.KIFeldherkunft,
	angaben map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return "", false, fehler
	}

	return a.feld.FelderVonAussenMitHerkunft(
		ctx, karteID, werte, herkunft, herkuenfte, angaben, erwartet)
}

// GruppeVonAussenMitHerkunft schreibt eine **atomare Gruppe**.
//
// # Die Aufsicht gilt je Karte
//
// **Und vor dem ersten Schreiben.** `gehoertZu` ist die Prüfung, die einen
// Schreibweg von außen von einem beliebigen Griff in den Bestand
// unterscheidet; in einer Gruppe muss sie für **jede** Karte gelten, und wenn
// eine durchfällt, wird keine geschrieben. Ein Prüfen mitten im Schreiben
// hätte genau den Teilbestand, den das Issue ausschließt.
func (a *Kartenaufsicht) GruppeVonAussenMitHerkunft(
	ctx context.Context, schritte []domain.Kartenschritt,
) error {
	for _, schritt := range schritte {
		if fehler := a.gehoertZu(ctx, schritt.SetID, schritt.KarteID); fehler != nil {
			return fehler
		}
	}

	return a.feld.GruppeSetzenMitHerkunft(ctx, schritte)
}

// FeldEntfernen nimmt ein Feld weg, wenn die Karte in dieses Set gehört.
func (a *Kartenaufsicht) FeldEntfernen(
	ctx context.Context, setID, karteID int64, feld string,
) (vorher string, hatte bool, stelle *int64, fehler error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return "", false, nil, fehler
	}

	return a.feld.FeldEntfernen(ctx, karteID, feld)
}

// FeldSetzenAnStelle setzt ein Feld an seine alte Stelle zurück.
func (a *Kartenaufsicht) FeldSetzenAnStelle(
	ctx context.Context, setID, karteID int64, feld, wert string, stelle int64,
) (vorher string, hatte bool, fehler error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return "", false, fehler
	}

	return a.feld.FeldSetzenAnStelle(ctx, karteID, feld, wert, stelle)
}

// FelderSetzen setzt ein Feld auf vielen Karten desselben Sets.
//
// **Erst alle prüfen, dann schreiben.** Eine fremde Kennung mitten in der
// Auswahl darf nicht dazu führen, dass die davorstehenden schon geschrieben
// sind — dieselbe Überlegung, die in `app.FelderSetzen` für **unlesbare**
// Kennungen steht. Sie gilt für fremde genauso.
//
// **Der Vorgang kommt mit heraus** (#1393): mit ihm nimmt eine Fläche genau
// diese Änderung zurück und nicht den jüngsten Schritt des Sets.
func (a *Kartenaufsicht) FelderSetzen(
	ctx context.Context, setID int64, karteIDs []int64, feld, wert string,
) (string, error) {
	if fehler := a.darfInDieMasse(ctx, setID, feld); fehler != nil {
		return "", fehler
	}

	for _, karteID := range karteIDs {
		if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
			return "", fehler
		}
	}

	if rechner, rechnet := a.feld.(Massenrechner); rechnet {
		return rechner.MassenSetzen(ctx, karteIDs, feld, wert, nil)
	}

	return a.massen.FelderSetzen(ctx, karteIDs, feld, wert)
}

// Massenrechner rechnet eine Massenänderung je Karte (#2516) — der Feldweg
// über die Feldnaht. **Wahlfrei**: ein Feldschreiber ohne Rechnung schreibt
// den Rohwert wie bisher.
type Massenrechner interface {
	MassenSetzen(ctx context.Context, karteIDs []int64, feld, wert string,
		erwartet []domain.Feldvorwert) (vorgang string, fehler error)
	Gerechnet(ctx context.Context, karte *domain.CardRow, feld, wert string) string
}

// darfInDieMasse sagt, ob ein Feld über viele Karten gesetzt werden darf.
//
// # Zwei Arten dürfen nicht, und beide standen offen
//
// **`editable: false`** — MSE lässt ein solches Feld nicht einmal auswählen:
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–157, Fassung f9d9356 vom
// 2021-07-06. 24 der 249 Kartenfelder von `magic.mse-game` tragen es.
//
// **`save value: false`** — ein Feld, das aus anderen entsteht. Wer
// hineinschreibt, schreibt einen Wert, den die Vorlage nicht liest; 46 der 249
// tragen es. **Und anders als beim Einzelweg wird nicht aufgelöst, sondern
// abgewiesen**: der Einzelweg zerlegt einen zusammengesetzten Text in seine
// Teile, weil dort **eine** Karte ihn bekommt. Über viele Karten hinweg gibt
// es keinen solchen Text — derselbe Wert müsste für jede anders zerfallen.
//
// Angestrichen vom Zweitblick über Phase 3.5 als Befund 1: die Oberfläche bot
// beide an, und dieser Weg reichte sie ungeprüft bis in den Speicher durch.
//
// # Eine unlesbare Vorlage hält nichts auf
//
// **Dieselbe Duldung wie beim Einzelweg, und derselbe Grund** (#720): ohne die
// Vorlage weiß niemand, ob ein Feld berechnet ist, und wer deshalb gar nicht
// schriebe, machte jedes Set unbearbeitbar, dessen Spielpaket fehlt. Ein Set
// aus MTG JSON ist genau das. Der Preis steht ausführlich bei [Feldweg.setzen].
//
// **Ohne Feldquelle ebenso.** Wer `MitFeldformen` nicht ruft, bekommt das
// vorige Verhalten — der MCP-Server etwa, der diesen Weg gar nicht anbietet.
func (a *Kartenaufsicht) darfInDieMasse(
	ctx context.Context, setID int64, feld string,
) error {
	if a.formen == nil {
		return nil
	}

	felder, lesefehler := a.formen.Kartenfelder(ctx, setID)
	if lesefehler != nil {
		// **Die Duldung steht oben im Absatz „Eine unlesbare Vorlage hält
		// nichts auf".** `nilerr` schlägt hier zu Recht an: ein verschluckter
		// Fehler sieht wie ein übersehener aus. Er ist keiner — er ist der
		// Preis dafür, dass ein Set ohne Spielpaket bearbeitbar bleibt.
		//nolint:nilerr // die Duldung ist beabsichtigt und oben begründet (#720)
		return nil
	}

	form, da := form(felder, feld)
	if !da {
		// **Ein Feld, das die Vorlage nicht kennt, geht durch.** Ein Set aus
		// MTG JSON trägt solche; sie zu verweigern hieße, sie unbearbeitbar zu
		// machen.
		return nil
	}

	if !form.Editierbar {
		return domain.NewFieldError("feld", domain.SchluesselNichtEditierbar,
			"Das Feld %q lässt sich nicht bearbeiten.", feld)
	}

	if !form.Gespeichert {
		return domain.NewFieldError("feld", domain.SchluesselBerechnetesFeld,
			"Das Feld %q entsteht aus anderen Feldern und lässt sich "+
				"nicht über viele Karten setzen.", feld)
	}

	return nil
}
