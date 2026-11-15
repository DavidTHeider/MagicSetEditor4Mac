// **Was der Zähler über Dienste sagt** — herausgezogen für Werkstück A13.3
// von Phase 5.3
// (#1452),
// als `zaehlerdoppel_test.go` an der Zeilengrenze aus §2 stand.
//
// **Der Schnitt folgt der Frage**: dort steht, was der Zähler über den
// **Bestand** sagt — Sets, Karten, Felder, Exemplare. Hier steht, was er über
// **fremde Dienste** sagt: Preise, Angebote, Modelle, Sicherungen.
//
// **Er zählt und antwortet dabei wie [williger]** — er ist der Beleg, dass
// ein gültiger Aufruf genau **einmal** durchkommt. Ein Doppel, das laut
// scheiterte, prüfte die Absage und nicht den Weg (§3).
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Und die Preise** (#1165).
func (z *zaehlender) PreiseHolen(context.Context, string) ([]rest.Preissicht, []string, error) {
	z.rufe++

	return []rest.Preissicht{probepreis()}, []string{}, nil
}

// **Und das Löschen eines Sets** (#1218).
func (z *zaehlender) SetLoeschen(context.Context, int64, string) ([]string, error) {
	z.rufe++

	return nil, nil
}

func (z *zaehlender) Preise(context.Context, string) ([]rest.Preissicht, error) {
	z.rufe++

	return []rest.Preissicht{probepreis()}, nil
}

func (z *zaehlender) Preisbulk(context.Context, string, string) (
	rest.Bulkbilanzsicht, error,
) {
	z.rufe++

	return rest.Bulkbilanzsicht{Quelle: "mtgjson", Umfang: "heute"}, nil
}

func (z *zaehlender) Marktangebote(_ context.Context, wunsch rest.Marktwunsch) (
	rest.Marktsicht, error,
) {
	z.rufe++

	return rest.Marktsicht{Quelle: wunsch.Quelle, Fassung: "89"}, nil
}

func (z *zaehlender) Anbieterstand(_ context.Context, dienst string) (
	rest.Anbietersicht, error,
) {
	z.rufe++

	return rest.Anbietersicht{Dienst: dienst, Rest: 84213}, nil
}

func (z *zaehlender) Fremdpreisverlauf(_ context.Context, wunsch rest.Verlaufsbezug) (
	rest.Verlaufssicht, error,
) {
	z.rufe++

	return rest.Verlaufssicht{Quelle: wunsch.Quelle, Fassung: "mtg-m11-149"}, nil
}

func (z *zaehlender) KIModelle(context.Context) (rest.KIKatalogsicht, error) {
	z.rufe++

	return rest.KIKatalogsicht{NurLokal: true}, nil
}

func (z *zaehlender) Preisverlauf(context.Context, rest.Preisreihe) (
	[]rest.Preissicht, error,
) {
	z.rufe++

	return []rest.Preissicht{probepreis()}, nil
}

func (z *zaehlender) Portfolio(context.Context) ([]rest.Portfoliozeilensicht, int, error) {
	z.rufe++

	return []rest.Portfoliozeilensicht{{
		Waehrung: "EUR", Wert: "4.17", Exemplare: 1,
		AeltesterPreis: "2026-08-27T12:00:00Z", Quelle: "scryfall",
	}}, 0, nil
}

// dienstMit baut einen Dienst über einem gegebenen Double.
func (z *zaehlender) BestandSichern(context.Context, string, string, bool) error {
	z.rufe++

	return nil
}

func (z *zaehlender) SicherungSchauen(context.Context, string, string) (rest.Sicherungsangaben, error) {
	z.rufe++

	return rest.Sicherungsangaben{AngelegtAm: "2026-08-28T00:00:00.000Z", Sets: 1}, nil
}

// **Und die Regelfreigabe** (G4, #1461).
func (z *zaehlender) Regelfreigabe(
	context.Context, string, string, int, string, []string,
) (
	rest.Regelfreigabeangaben, error,
) {
	z.rufe++

	return rest.Regelfreigabeangaben{
		Matrix:      []rest.Regelbereichsangaben{},
		Luecken:     []string{},
		Hindernisse: []string{},
	}, nil
}

// **Und der Netzweg** (N5, #1571).
func (z *zaehlender) Netzstand(context.Context) (rest.Netzstandangaben, error) {
	z.rufe++

	return rest.Netzstandangaben{
		ProxyHTTPS:      rest.Netzfeld{Wert: "http://proxy:3128", Herkunft: "einstellung"},
		FristVerbindung: 15, FristKopf: 30, FristStillstand: 60,
		ProxyBenutzt: true,
	}, nil
}

// **Und der portable Umschlag** (N4, #1457).
func (z *zaehlender) UmschlagSchnueren(context.Context, string, string, string) (
	rest.Umschlagangaben, error,
) {
	z.rufe++

	return rest.Umschlagangaben{Objekte: []rest.Umschlagobjekt{}}, nil
}

func (z *zaehlender) UmschlagAnsehen(context.Context, string) (
	rest.Umschlagangaben, error,
) {
	z.rufe++

	return rest.Umschlagangaben{
		Urheber: "b1", Fingerabdruck: "AB CD",
		Objekte: []rest.Umschlagobjekt{{Art: "set", Bezeichnung: "set.mse-set", Groesse: 7}},
	}, nil
}

func (z *zaehlender) UmschlagUebernehmen(context.Context, string) (
	rest.Umschlagannahme, error,
) {
	z.rufe++

	return rest.Umschlagannahme{SetID: "9", Karten: 3, Befunde: []string{}}, nil
}

// Pruefen meldet einen Befund und **zählt mit**.
//
// Ohne das Mitzählen wäre der Endpunkt grün, sobald er 200 gibt — auch wenn
// er den Anwendungsfall gar nicht ruft. Genau die Lücke hat der Nachlauf zum
// Zweitblick über Phase U gefunden.
func (z *zaehlender) Pruefen(string) (zeile, spalte int, meldung string) {
	z.rufe++

	return 0, 0, ""
}

// FelderSetzen setzt ein Feld auf mehreren Karten und **zählt mit** (#1390).
func (z *zaehlender) FelderSetzen(context.Context, int64, []int64, string, string) (
	string, error,
) {
	z.rufe++

	return "vorgang7", nil
}

// FelderVorschau gibt eine Zeile je Karte; die zweite steht schon so (U6).
func (z *zaehlender) FelderVorschau(_ context.Context, _ int64, karteIDs []int64, feld, wert string) (
	domain.Massenvorschau, error,
) {
	z.rufe++

	zeilen := make([]domain.Massenvorschauzeile, 0, len(karteIDs))
	for i, karteID := range karteIDs {
		zeile := domain.Massenvorschauzeile{KarteID: karteID, Name: "Karte", Alt: "Alt", GabAlt: true, Neu: wert}
		if i == 1 {
			zeile.Alt, zeile.Ausschluss = wert, domain.AusschlussGleich
		}

		zeilen = append(zeilen, zeile)
	}

	return domain.Massenvorschau{Feld: feld, Wert: wert, Zeilen: zeilen}, nil
}

func (z *zaehlender) FelderSetzenMitErwartung(
	context.Context, int64, []domain.Feldvorwert, string, string,
) (string, error) {
	z.rufe++

	return "vorgang8", nil
}

// Verschluesselungsstand meldet einen Klartextbestand und zählt mit (#1351,
// Werkstück 14).
//
// **Das Mitzählen ist der Punkt**: der Wächter über die vier Klassen belegt
// mit ihm, dass „gültig" wirklich bis zum Anwendungsfall durchgeht und nicht
// nur einen Status zurückgibt. Ohne die Zeile meldete er „0mal gerufen".
func (z *zaehlender) Verschluesselungsstand() (bool, error) {
	z.rufe++

	return false, nil
}

// Sicherungsstand zählt mit wie [zaehlender.Verschluesselungsstand] (OPS01).
func (z *zaehlender) Sicherungsstand(context.Context) (any, error) {
	z.rufe++

	return map[string]any{"stufe": "unbekannt"}, nil
}

// KIErkennung liest ein Kartenfoto und zählt mit — A13.3 (#1452).
func (z *zaehlender) KIErkennung(
	ctx context.Context, bezug rest.KIErkennungsbezug,
) (any, error) {
	z.rufe++

	return z.williger.KIErkennung(ctx, bezug)
}

// **Und die Partien** (G1, #1458).
func (z *zaehlender) PartieAnlegen(
	_ context.Context, _, _ string, sitze []rest.Sitzwunsch,
	abgewaehlt []string,
) (rest.Partieangaben, error) {
	z.rufe++
	z.sitze = sitze
	z.abgewaehlt = abgewaehlt

	return rest.Partieangaben{
		Sitze: []rest.Partiesitz{}, Objekte: []rest.Partieobjekt{},
		Erlaubt: []string{},
	}, nil
}

func (z *zaehlender) Partie(context.Context, string, string) (
	rest.Partieangaben, error,
) {
	z.rufe++

	return rest.Partieangaben{
		Sitze: []rest.Partiesitz{}, Objekte: []rest.Partieobjekt{},
		Erlaubt: []string{},
	}, nil
}

// Partien gibt eine Zeile — die Liste ist ein eigener Weg (#1600).
func (z *zaehlender) Partien(context.Context) (
	[]rest.Partiezeilenangaben, error,
) {
	z.rufe++

	return []rest.Partiezeilenangaben{{
		ID: "p1", Spiel: "magic", Stand: "laeuft", Objekte: 12,
		Sitze: []rest.Zeilensitzangaben{{Kennung: "anna", Name: "Anna"}},
	}}, nil
}

func (z *zaehlender) PartieBefehl(context.Context, rest.Spielbefehl) (
	rest.Befehlsangaben, error,
) {
	z.rufe++

	return rest.Befehlsangaben{Ereignisse: []rest.Partieereignis{}}, nil
}

func (z *zaehlender) PartieVerlauf(context.Context, string, string) (
	[]rest.Partieereignis, error,
) {
	z.rufe++

	return []rest.Partieereignis{}, nil
}

func (z *zaehlender) PartieStand(context.Context, string, string, string) (
	rest.Partieangaben, error,
) {
	z.rufe++

	return rest.Partieangaben{
		Sitze: []rest.Partiesitz{}, Objekte: []rest.Partieobjekt{},
		Erlaubt: []string{},
	}, nil
}
