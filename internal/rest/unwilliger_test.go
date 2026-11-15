// **Der unwillige Double** — er scheitert an allem
// (#1107).
//
// **Eine eigene Datei**, weil `durchreichen_test.go` mit den Setangaben über
// die Zeilengrenze aus §2 wuchs. Der Schnitt folgt derselben Frage wie bei
// `vorrichtungbestand_test.go`: dort die **Prüfung**, hier die **Antwort**,
// die der Double gibt.
package rest_test

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// unwilliger scheitert an allem.
type unwilliger struct{ fehler error }

// Die fünf Wege der Autoersetzung scheitern ebenso (#1341).
func (u unwilliger) Autoersetzungen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) AutoersetzungSetzen(context.Context, int64, string, string, bool, bool) error {
	return u.fehler
}

func (u unwilliger) AutoersetzungLoeschen(context.Context, int64, string) error {
	return u.fehler
}

func (u unwilliger) AutoersetzungenZuruecksetzen(context.Context, int64) error {
	return u.fehler
}

func (u unwilliger) AutoersetzungSchalten(context.Context, int64, bool) error { return u.fehler }

// Die Parameterhilfen scheitern ebenso (#1342).
func (u unwilliger) Parameterhilfen(context.Context, int64, int64) (any, error) {
	return nil, u.fehler
}

// Die beiden Wege der Auszählung scheitern ebenso (#1338).
func (u unwilliger) Auszaehldimensionen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Auszaehlen(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

// NeuesteFassung und FassungLaden scheitern wie alles andere (#1100).
func (u unwilliger) NeuesteFassung(context.Context) (rest.Fassungsstand, error) {
	return rest.Fassungsstand{}, u.fehler
}

func (u unwilliger) FassungLaden(context.Context) (string, error) {
	return "", u.fehler
}

// Auswerten scheitert wie alles andere (#1070).
func (u unwilliger) Auswerten(context.Context, int64, string) (rest.Skriptergebnis, error) {
	return rest.Skriptergebnis{}, u.fehler
}

// Die drei Züge der Paketwerkstatt scheitern ebenso (#1164); der Dateibaum
// steht in `dateibaumdoppel_test.go`.
func (u unwilliger) DateienUndBeiwerk(context.Context, string) (dateien, beiwerk []string, fehler error) {
	return nil, nil, u.fehler
}

func (u unwilliger) Lesen(context.Context, string, string) (inhalt, stand string, fehler error) {
	return "", "", u.fehler
}

func (u unwilliger) Schreiben(context.Context, string, string, string, string) (string, error) {
	return "", u.fehler
}

// Namen scheitert ebenso (#1164).
func (u unwilliger) Namen(context.Context, int64) ([]string, error) {
	return nil, u.fehler
}

// Wortlisten und Wortwahl scheitern ebenso (#1074).
func (u unwilliger) Wortlisten(context.Context, int64, string) ([]rest.Wortbereich, error) {
	return nil, u.fehler
}

func (u unwilliger) Wortwahl(context.Context, int64, int64, string, int, []int) error {
	return u.fehler
}

// Die Packwege scheitern ebenso (#889).
// SetAnlegen scheitert wie alles andere (#891).
func (u unwilliger) SetAnlegen(context.Context, string, string, string) (int64, error) {
	return 0, u.fehler
}

func (u unwilliger) Packarten(context.Context, int64) (rest.Packangebot, error) {
	return rest.Packangebot{}, u.fehler
}

func (u unwilliger) PackZiehen(context.Context, int64, string, string) (
	rest.Packziehung, error,
) {
	return rest.Packziehung{}, u.fehler
}

// Die Schlüsselwortwege scheitern ebenso (#1062).
func (u unwilliger) Schluesselwoerter(context.Context, int64) (
	rest.Schluesselwortliste, error,
) {
	return rest.Schluesselwortliste{}, u.fehler
}

func (u unwilliger) SchluesselwortAnlegen(context.Context, int64) (int64, error) {
	return 0, u.fehler
}

func (u unwilliger) SchluesselwortSetzen(
	context.Context, int64, int64, string, string,
) error {
	return u.fehler
}

func (u unwilliger) SchluesselwortLoeschen(context.Context, int64, int64) error {
	return u.fehler
}

// Anlegen scheitert wie alles andere (#1032).
func (u unwilliger) Anlegen(context.Context, int64, string) (rest.Anlage, error) {
	return rest.Anlage{}, u.fehler
}

// Einlesen scheitert ebenso (#1033).
func (u unwilliger) Einlesen(context.Context, string, string) (rest.Satz, error) {
	return rest.Satz{}, u.fehler
}

func (u unwilliger) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, u.fehler
}

func (u unwilliger) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, u.fehler
}

func (u unwilliger) Stand(context.Context) (rest.Sammelstand, error) {
	return rest.Sammelstand{}, u.fehler
}

func (u unwilliger) Holen(context.Context) (rest.Sammelstand, error) {
	return rest.Sammelstand{}, u.fehler
}

func (u unwilliger) Regelhinweise(context.Context, string) (rest.Regelauskunft, error) {
	return rest.Regelauskunft{}, u.fehler
}

func (u unwilliger) Manakosten(context.Context, string) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Einsetzen(context.Context, int64, int64, string, string, string) (
	rest.Bildeinsatz, error,
) {
	return rest.Bildeinsatz{}, u.fehler
}

func (u unwilliger) FeldSetzenMitVorwert(
	context.Context, int64, int64, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, u.fehler
}

// **Und das Nachzeichnen** (#1340).
func (u unwilliger) SymbolNachzeichnen(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

// Pruefen meldet keinen Befund — die Prüfung hat keinen Fehlerweg.
//
// **Sie kann nicht scheitern**: der Text kommt mit, es wird nichts gelesen
// und nichts geschrieben. Ein `fehler` wie bei den Nachbarn wäre ein Zweig,
// den kein Lauf betreten kann (§3).
func (u unwilliger) Pruefen(string) (zeile, spalte int, meldung string) {
	return 0, 0, ""
}

// FelderSetzen scheitert wie alles hier (#1390).
func (u unwilliger) FelderSetzen(context.Context, int64, []int64, string, string) (
	string, error,
) {
	return "", u.fehler
}

func (u unwilliger) FelderVorschau(context.Context, int64, []int64, string, string) (
	domain.Massenvorschau, error,
) {
	return domain.Massenvorschau{}, u.fehler
}

func (u unwilliger) FelderSetzenMitErwartung(
	context.Context, int64, []domain.Feldvorwert, string, string,
) (string, error) {
	return "", u.fehler
}

// **Der Symboleditor** (#1391) — alles scheitert.
func (u unwilliger) SymbolLesen(context.Context, int64, string) (any, error) { return nil, u.fehler }

func (u unwilliger) SymbolAblegen(context.Context, int64, json.RawMessage, string) (string, error) {
	return "", u.fehler
}

func (u unwilliger) SymbolRechnen(string, json.RawMessage) (any, error) { return nil, u.fehler }

func (u unwilliger) Palette(context.Context, int64, string) (any, error) { return nil, u.fehler }

// Verschluesselungsstand scheitert wie alles an diesem Doppel.
func (u unwilliger) Verschluesselungsstand() (bool, error) { return false, u.fehler }

func (u unwilliger) Sicherungsstand(context.Context) (any, error) { return nil, u.fehler }

// KIGruppe stolpert — A13.1b (#1452).
//
// **Mit dem eingestellten Fehler und nicht mit einem eigenen**: der Lauf
// darüber prüft, dass **die Klasse** durchkommt, und ein blanker Fehler würde
// dort zu „unbekannt".
func (u unwilliger) KIGruppe(context.Context, rest.KIGruppenbezug) (
	rest.KIGruppensicht, error,
) {
	return rest.KIGruppensicht{}, u.fehler
}

// KIErkennung stolpert — A13.3 (#1452), mit dem **eingestellten** Fehler:
// der Lauf darüber prüft, dass die Klasse durchkommt.
func (u unwilliger) KIErkennung(context.Context, rest.KIErkennungsbezug) (
	any, error,
) {
	return nil, u.fehler
}
