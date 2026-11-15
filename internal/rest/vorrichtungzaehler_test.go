// **Die gezählten Antworten für Volltext und Lebenszyklus**
// (#1107).
//
// **Eine eigene Datei**, weil `vorrichtung_test.go` an der Zeilengrenze aus §2
// stand. Der Schnitt folgt derselben Frage wie bei `vorrichtungbestand_test.go`:
// dort die Vorrichtung, hier die Antworten.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (z *zaehlender) Felder(ctx context.Context, setID int64) ([]string, error) {
	z.zaehle()

	return z.williger.Felder(ctx, setID)
}

func (z *zaehlender) Suche(
	ctx context.Context, frage domain.SearchQuery,
) (domain.SearchResult, error) {
	z.zaehle()

	return z.williger.Suche(ctx, frage)
}

func (z *zaehlender) KarteAnlegen(ctx context.Context, setID int64) (int64, error) {
	z.zaehle()

	return z.williger.KarteAnlegen(ctx, setID)
}

func (z *zaehlender) KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error) {
	z.zaehle()

	return z.williger.KarteKopieren(ctx, setID, karteID)
}

func (z *zaehlender) KarteLoeschen(ctx context.Context, setID, karteID int64, _ bool) error {
	z.zaehle()

	return z.williger.KarteLoeschen(ctx, setID, karteID, false)
}

func (z *zaehlender) KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error {
	z.zaehle()

	return z.williger.KartenTauschen(ctx, setID, eineID, andereID)
}

func (z *zaehlender) Angaben(ctx context.Context, setID int64) (
	[]rest.Feldbeschreibung, []rest.Setwert, error,
) {
	z.zaehle()

	return z.williger.Angaben(ctx, setID)
}

func (z *zaehlender) Setzen(ctx context.Context, setID int64, feld, wert string) (
	vorher string, hatte bool, fehler error,
) {
	z.zaehle()

	return z.williger.Setzen(ctx, setID, feld, wert)
}

func (z *zaehlender) Vorbehalte(ctx context.Context, setID int64) ([]rest.Vorbehalt, error) {
	z.zaehle()

	return z.williger.Vorbehalte(ctx, setID)
}

func (z *zaehlender) Beispielsets() ([]rest.Beispielset, error) {
	z.zaehle()

	return z.williger.Beispielsets()
}

// **Und die gezählten Antworten für die sechs neuen Nähte** (#1107).
func (z *zaehlender) Stile(ctx context.Context, setID int64) ([]rest.Stil, error) {
	z.zaehle()

	return z.williger.Stile(ctx, setID)
}

func (z *zaehlender) Folgen(ctx context.Context, setID int64, paket string) (
	rest.Wechselfolgen, error,
) {
	z.zaehle()

	return z.williger.Folgen(ctx, setID, paket)
}

func (z *zaehlender) Wechseln(ctx context.Context, setID int64, paket string) error {
	z.zaehle()

	return z.williger.Wechseln(ctx, setID, paket)
}

func (z *zaehlender) Optionen(ctx context.Context, setID int64) ([]rest.Stiloption, error) {
	z.zaehle()

	return z.williger.Optionen(ctx, setID)
}

func (z *zaehlender) OptionSetzen(ctx context.Context, setID int64, feld, wert string) error {
	z.zaehle()

	return z.williger.OptionSetzen(ctx, setID, feld, wert)
}

func (z *zaehlender) OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error {
	z.zaehle()

	return z.williger.OptionZuruecksetzen(ctx, setID, feld)
}

func (z *zaehlender) Liste(ctx context.Context) ([]rest.Paket, []string, error) {
	z.zaehle()

	return z.williger.Liste(ctx)
}

func (z *zaehlender) Beziehen(
	ctx context.Context, adresse, stand, unterverzeichnis, name string,
) (rest.Bezug, error) {
	z.zaehle()

	return z.williger.Beziehen(ctx, adresse, stand, unterverzeichnis, name)
}

func (z *zaehlender) Statistik(ctx context.Context, setID int64) (rest.Setzahlen, error) {
	z.zaehle()

	return z.williger.Statistik(ctx, setID)
}

func (z *zaehlender) Verteilung(ctx context.Context, setID int64, feld string) (
	[]rest.Wertzahl, error,
) {
	z.zaehle()

	return z.williger.Verteilung(ctx, setID, feld)
}

func (z *zaehlender) Zuruecknehmen(
	ctx context.Context, setID int64, vorgang string,
) (rest.Bewegung, error) {
	z.zaehle()

	return z.williger.Zuruecknehmen(ctx, setID, vorgang)
}

func (z *zaehlender) Wiederherstellen(
	ctx context.Context, setID int64, vorgang string,
) (rest.Bewegung, error) {
	z.zaehle()

	return z.williger.Wiederherstellen(ctx, setID, vorgang)
}

func (z *zaehlender) Alle(ctx context.Context) ([]rest.Einstellung, error) {
	z.zaehle()

	return z.williger.Alle(ctx)
}

func (z *zaehlender) EinstellungSetzen(ctx context.Context, schluessel, wert string) error {
	z.zaehle()

	return z.williger.EinstellungSetzen(ctx, schluessel, wert)
}

// Die Zwischenablage zählt mit (#1168).
func (z *zaehlender) AblageKopieren(context.Context, int64, []int64) (string, error) {
	z.rufe++

	return "game_name: probe\ncard:\n", nil
}

func (z *zaehlender) AblageAusschneiden(context.Context, int64, []int64) (string, error) {
	z.rufe++

	return "game_name: probe\ncard:\n", nil
}

func (z *zaehlender) AblageEinfuegen(context.Context, int64, string) (
	karten, befunde []string, fehler error,
) {
	z.rufe++

	return nil, nil, nil
}

// Die Verweise zählen mit (#1167).
func (z *zaehlender) VerweisSetzen(_ context.Context, _ int64, printing string) (
	rest.Verweissicht, error,
) {
	z.rufe++

	return rest.Verweissicht{KarteID: "2", Printing: printing}, nil
}

func (z *zaehlender) Verweis(context.Context, int64) (*rest.Verweissicht, error) {
	z.rufe++

	return nil, nil //nolint:nilnil // „keiner" ist kein Fehler
}

func (z *zaehlender) VerweisLoesen(context.Context, int64) error {
	z.rufe++

	return nil
}

// **Und die Marken** (#1208) — vier Wege, vier Zähler.
func (z *zaehlender) MarkeSetzen(context.Context, int64, int64, string) error {
	z.rufe++

	return nil
}

func (z *zaehlender) MarkeEntfernen(context.Context, int64, int64, string) error {
	z.rufe++

	return nil
}

func (z *zaehlender) KartenMarken(context.Context, int64, int64) ([]string, error) {
	z.rufe++

	return []string{"fertig"}, nil
}

func (z *zaehlender) Setmarken(context.Context, int64) ([]string, error) {
	z.rufe++

	return []string{"fertig"}, nil
}

func (z *zaehlender) Verweise(context.Context, int64) ([]rest.Verweissicht, error) {
	z.rufe++

	return nil, nil
}

// Kartensuche zählt mit (#1167).
func (z *zaehlender) Kartensuche(context.Context, string) (rest.Suchergebnis, error) {
	z.zaehle()

	return rest.Suchergebnis{}, nil
}

// Dienstkatalog zählt mit (#1440).
func (z *zaehlender) Grenzen(context.Context) (rest.Grenzen, error) {
	z.zaehle()

	return rest.GrenzenVon(domain.Vorgabegrenzen()), nil
}

// **Der Zähler antwortet „bereit"** — sonst gäbe `/bereit` eine 503, und der
// Lauf über die vier Klassen erwartet für den gültigen Fall eine 200.
func (z *zaehlender) Gesundheit(context.Context) (
	zustand any, bereit bool, fehler error,
) {
	z.zaehle()

	return map[string]any{"lebt": true, "bereit": true}, true, nil
}

func (z *zaehlender) Dienstkatalog(context.Context) ([]rest.Dienstfaehigkeit, error) {
	z.zaehle()

	return nil, nil
}

// **Die drei Peerwege zählen mit** (N1, #1454).
func (z *zaehlender) Peeridentitaet(context.Context) (rest.Peeridentitaet, error) {
	z.zaehle()

	return rest.Peeridentitaet{}, nil
}

func (z *zaehlender) PeeridentitaetSicherstellen(context.Context) (rest.Peeridentitaet, error) {
	z.zaehle()

	return rest.Peeridentitaet{}, nil
}

func (z *zaehlender) PeeridentitaetRotieren(context.Context) (rest.Peeridentitaet, error) {
	z.zaehle()

	return rest.Peeridentitaet{}, nil
}

// Das Ersetzen zählt mit (#1168).
func (z *zaehlender) ErsetzungVorschau(
	context.Context, int64, domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	z.rufe++

	return domain.Ersetzungsvorschau{}, nil
}

func (z *zaehlender) ErsetzungAusfuehren(
	context.Context, int64, domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	z.rufe++

	return domain.Ersetzungsvorschau{}, nil
}

// Die Kartenliste zählt mit (#1168).
func (z *zaehlender) KartenFiltern(context.Context, int64, string) ([]int64, error) {
	z.rufe++

	return []int64{}, nil
}

func (z *zaehlender) Listenaufbau(context.Context, int64) (
	domain.Kartenlistenaufbau, error,
) {
	z.rufe++

	return domain.Kartenlistenaufbau{}, nil
}

func (z *zaehlender) ListenspaltenSetzen(
	context.Context, int64, []domain.Spaltenwahl,
) error {
	z.rufe++

	return nil
}

func (z *zaehlender) ListensortierungSetzen(
	context.Context, int64, domain.Kartenordnung,
) error {
	z.rufe++

	return nil
}

func (z *zaehlender) Metriken(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"reihen": []any{}, "grenze": 40}, nil
}

func (z *zaehlender) Meta(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"fassung": "0.4.0", "flaeche": "dienst"}, nil
}

func (z *zaehlender) Vertraeglichkeit(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"urteil": "vertraeglich", "hindernisse": []any{}}, nil
}

func (z *zaehlender) Ruecksprungarten(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"arten": []any{}}, nil
}

func (z *zaehlender) Ruecksprunglage(context.Context, string, string) (any, error) {
	z.zaehle()

	return map[string]any{"zeilen": []any{}, "unklar": []any{}}, nil
}

func (z *zaehlender) Rueckkehren(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"rueckkehren": []any{}}, nil
}

func (z *zaehlender) RueckkehrQuittieren(context.Context, int64) error {
	z.zaehle()

	return nil
}

func (z *zaehlender) Pflegevorschau(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"befunde": []any{}}, nil
}

func (z *zaehlender) Pflegen(context.Context, string, []rest.Pflegewahl) (any, error) {
	z.zaehle()

	return map[string]any{"zeilen": []any{}}, nil
}
