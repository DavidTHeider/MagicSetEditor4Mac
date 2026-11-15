// **Die unwilligen Antworten der Ausfuhr** — einfahren, sichern, zeichnen, den
// Bogen legen, MTG JSON, die Paketpflege, das Kartenbild.
//
// **Eine eigene Datei**, seit `unwilliger_test.go` mit der Restparität über die
// Zeilengrenze aus §2 lief
// (#1291).
// Der Schnitt folgt der Frage: dort scheitert, **was am Bestand** geschieht,
// hier, **was die Platte verlässt oder auf sie kommt**.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (u unwilliger) Einfahren(
	context.Context, string,
) (setID int64, karten int, befunde []string, fehler error) {
	return 0, 0, nil, u.fehler
}

func (u unwilliger) Sichern(context.Context, int64, string) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Zeichnen(
	context.Context, int64, string, string,
) (breite, hoehe int, befunde []string, fehler error) {
	return 0, 0, nil, u.fehler
}

func (u unwilliger) Bogen(
	context.Context, int64, string,
) (seiten, karten int, befunde []string, fehler error) {
	return 0, 0, nil, u.fehler
}

func (u unwilliger) MTGJSONLesen(context.Context, string, string) (
	angelegt []int64, befunde []string, fehler error,
) {
	return nil, nil, u.fehler
}

func (u unwilliger) MTGJSONSchreiben(context.Context, int64, string, string) (
	karten int, befunde []string, fehler error,
) {
	return 0, nil, u.fehler
}

func (u unwilliger) MTGJSONSpiele() []string { return []string{"magic"} }

func (u unwilliger) Vorschlag(
	context.Context, int64, int64, string,
) (rest.Vorschlag, error) {
	return rest.Vorschlag{}, u.fehler
}

func (u unwilliger) Uebernehmen(
	context.Context, int64, int64, string, []string,
) (vorher []rest.Aenderung, fehler error) {
	return nil, u.fehler
}

// **Und Stile, Pakete, Kennzahlen, Verlauf und Einstellungen scheitern
// ebenso** (#1107).
func (u unwilliger) Stile(context.Context, int64) ([]rest.Stil, error) {
	return nil, u.fehler
}

func (u unwilliger) Folgen(context.Context, int64, string) (rest.Wechselfolgen, error) {
	return rest.Wechselfolgen{}, u.fehler
}

func (u unwilliger) Wechseln(context.Context, int64, string) error { return u.fehler }

func (u unwilliger) Optionen(context.Context, int64) ([]rest.Stiloption, error) {
	return nil, u.fehler
}

func (u unwilliger) OptionSetzen(context.Context, int64, string, string) error {
	return u.fehler
}

func (u unwilliger) OptionZuruecksetzen(context.Context, int64, string) error {
	return u.fehler
}

func (u unwilliger) Liste(context.Context) ([]rest.Paket, []string, error) {
	return nil, nil, u.fehler
}

func (u unwilliger) Beziehen(context.Context, string, string, string, string) (
	rest.Bezug, error,
) {
	return rest.Bezug{}, u.fehler
}

func (u unwilliger) Statistik(context.Context, int64) (rest.Setzahlen, error) {
	return rest.Setzahlen{}, u.fehler
}

func (u unwilliger) Verteilung(context.Context, int64, string) ([]rest.Wertzahl, error) {
	return nil, u.fehler
}

func (u unwilliger) Zuruecknehmen(context.Context, int64, string) (rest.Bewegung, error) {
	return rest.Bewegung{}, u.fehler
}

func (u unwilliger) Wiederherstellen(context.Context, int64, string) (rest.Bewegung, error) {
	return rest.Bewegung{}, u.fehler
}

func (u unwilliger) Alle(context.Context) ([]rest.Einstellung, error) {
	return nil, u.fehler
}

func (u unwilliger) EinstellungSetzen(context.Context, string, string) error {
	return u.fehler
}

// **Die Sammlung** (#1165).
func (u unwilliger) Sammlungsfenster(
	context.Context, string, string, string, int, int,
) (fenster []rest.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	return nil, domain.Fensterzahlen{}, u.fehler
}

func (u unwilliger) Sammlung(context.Context) ([]rest.Exemplarsicht, error) {
	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}}, u.fehler
}

func (u unwilliger) SammlungFuer(context.Context, string) ([]rest.Exemplarsicht, error) {
	return []rest.Exemplarsicht{{ID: 1, Printing: "probe-druck", Menge: 1}}, u.fehler
}

func (u unwilliger) ExemplarAnlegen(context.Context, rest.Exemplarsicht) (int64, error) {
	return 1, u.fehler
}

func (u unwilliger) ExemplarAendern(context.Context, rest.Exemplarsicht) error {
	return u.fehler
}

func (u unwilliger) ExemplarLoeschen(context.Context, int64, int64) error {
	return u.fehler
}

// **Und die Preise** (#1165).
func (u unwilliger) PreiseHolen(context.Context, string) ([]rest.Preissicht, []string, error) {
	return []rest.Preissicht{probepreis()}, []string{}, u.fehler
}

// **Und das Löschen eines Sets** (#1218).
func (u unwilliger) SetLoeschen(context.Context, int64, string) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Preise(context.Context, string) ([]rest.Preissicht, error) {
	return []rest.Preissicht{probepreis()}, u.fehler
}

func (u unwilliger) Preisverlauf(context.Context, rest.Preisreihe) (
	[]rest.Preissicht, error,
) {
	return []rest.Preissicht{probepreis()}, u.fehler
}

func (u unwilliger) Preisbulk(context.Context, string, string) (
	rest.Bulkbilanzsicht, error,
) {
	return rest.Bulkbilanzsicht{}, u.fehler
}

func (u unwilliger) Marktangebote(context.Context, rest.Marktwunsch) (
	rest.Marktsicht, error,
) {
	return rest.Marktsicht{}, u.fehler
}

func (u unwilliger) Portfolio(context.Context) ([]rest.Portfoliozeilensicht, int, error) {
	return []rest.Portfoliozeilensicht{{
		Waehrung: "EUR", Wert: "4.17", Exemplare: 1,
		AeltesterPreis: "2026-08-27T12:00:00Z", Quelle: "scryfall",
	}}, 0, u.fehler
}

// Die Zwischenablage scheitert wie alles andere (#1168).
func (u unwilliger) AblageKopieren(context.Context, int64, []int64) (string, error) {
	return "", u.fehler
}

func (u unwilliger) AblageAusschneiden(context.Context, int64, []int64) (string, error) {
	return "", u.fehler
}

func (u unwilliger) AblageEinfuegen(context.Context, int64, string) (
	karten, befunde []string, fehler error,
) {
	return nil, nil, u.fehler
}

// Die Verweise scheitern wie alles andere (#1167).
func (u unwilliger) VerweisSetzen(context.Context, int64, string) (rest.Verweissicht, error) {
	return rest.Verweissicht{}, u.fehler
}

func (u unwilliger) Verweis(context.Context, int64) (*rest.Verweissicht, error) {
	return nil, u.fehler
}

func (u unwilliger) VerweisLoesen(context.Context, int64) error { return u.fehler }

// **Und die Marken** (#1208) — vier Wege, ein Fehler.
func (u unwilliger) MarkeSetzen(context.Context, int64, int64, string) error { return u.fehler }

func (u unwilliger) MarkeEntfernen(context.Context, int64, int64, string) error {
	return u.fehler
}

func (u unwilliger) KartenMarken(context.Context, int64, int64) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Setmarken(context.Context, int64) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Verweise(context.Context, int64) ([]rest.Verweissicht, error) {
	return nil, u.fehler
}

// Kartensuche scheitert wie alles andere (#1167).
func (u unwilliger) Kartensuche(context.Context, string) (rest.Suchergebnis, error) {
	return rest.Suchergebnis{}, u.fehler
}

// Dienstkatalog scheitert wie alles andere (#1440).
func (u unwilliger) Grenzen(context.Context) (rest.Grenzen, error) {
	return rest.Grenzen{}, u.fehler
}

func (u unwilliger) Gesundheit(context.Context) (
	zustand any, bereit bool, fehler error,
) {
	return nil, false, u.fehler
}

func (u unwilliger) Dienstkatalog(context.Context) ([]rest.Dienstfaehigkeit, error) {
	return nil, u.fehler
}

// **Und die drei Peerwege scheitern wie alles andere** (N1, #1454).
func (u unwilliger) Peeridentitaet(context.Context) (rest.Peeridentitaet, error) {
	return rest.Peeridentitaet{}, u.fehler
}

func (u unwilliger) PeeridentitaetSicherstellen(context.Context) (rest.Peeridentitaet, error) {
	return rest.Peeridentitaet{}, u.fehler
}

func (u unwilliger) PeeridentitaetRotieren(context.Context) (rest.Peeridentitaet, error) {
	return rest.Peeridentitaet{}, u.fehler
}

// Das Ersetzen scheitert wie alles andere (#1168).
func (u unwilliger) ErsetzungVorschau(
	context.Context, int64, domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	return domain.Ersetzungsvorschau{}, u.fehler
}

func (u unwilliger) ErsetzungAusfuehren(
	context.Context, int64, domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	return domain.Ersetzungsvorschau{}, u.fehler
}

// Die Kartenliste scheitert wie alles andere (#1168).
func (u unwilliger) KartenFiltern(context.Context, int64, string) ([]int64, error) {
	return nil, u.fehler
}

func (u unwilliger) Listenaufbau(context.Context, int64) (
	domain.Kartenlistenaufbau, error,
) {
	return domain.Kartenlistenaufbau{}, u.fehler
}

func (u unwilliger) ListenspaltenSetzen(
	context.Context, int64, []domain.Spaltenwahl,
) error {
	return u.fehler
}

func (u unwilliger) ListensortierungSetzen(
	context.Context, int64, domain.Kartenordnung,
) error {
	return u.fehler
}

// **Und die Sicherung des ganzen Bestands** (#1169).
func (u unwilliger) BestandSichern(context.Context, string, string, bool) error { return u.fehler }

func (u unwilliger) SicherungSchauen(context.Context, string, string) (rest.Sicherungsangaben, error) {
	return rest.Sicherungsangaben{}, u.fehler
}

// **Und der Netzweg** (N5, #1571).
func (u unwilliger) Regelfreigabe(
	context.Context, string, string, int, string, []string,
) (
	rest.Regelfreigabeangaben, error,
) {
	return rest.Regelfreigabeangaben{}, u.fehler
}

func (u unwilliger) Netzstand(context.Context) (rest.Netzstandangaben, error) {
	return rest.Netzstandangaben{}, u.fehler
}

// **Und der portable Umschlag** (N4, #1457).
func (u unwilliger) UmschlagSchnueren(context.Context, string, string, string) (
	rest.Umschlagangaben, error,
) {
	return rest.Umschlagangaben{}, u.fehler
}

func (u unwilliger) UmschlagAnsehen(context.Context, string) (rest.Umschlagangaben, error) {
	return rest.Umschlagangaben{}, u.fehler
}

func (u unwilliger) UmschlagUebernehmen(context.Context, string) (rest.Umschlagannahme, error) {
	return rest.Umschlagannahme{}, u.fehler
}

func (u unwilliger) AusDatei(context.Context, int64, int64, string, string) (
	rest.Bildmasse, error,
) {
	return rest.Bildmasse{}, u.fehler
}

func (u unwilliger) Vorlagen(context.Context, int64) ([]rest.Vorlagenzeile, []string, error) {
	return nil, nil, u.fehler
}

func (u unwilliger) Ausfuehren(
	context.Context, int64, string, string, map[string]string,
) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) VorlagenoptionSetzen(context.Context, string, string, string) error {
	return u.fehler
}

// **Die Listeneinfuhr** (#1292) — beide scheitern.
func (u unwilliger) ListeDeuten(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) ListeEinlesen(context.Context, []rest.Listeneinfuhrzeile, string) (any, error) {
	return nil, u.fehler
}

// **Die Decklisteneinfuhr** (T8, #1473) — beide scheitern.
func (u unwilliger) DecklisteVorschau(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) DecklisteUebernehmen(
	context.Context, int64, string, []rest.Deckeinfuhrzeile,
) (any, error) {
	return nil, u.fehler
}

// **Die Sammlungsausfuhr** (#1292) — beide scheitern.
