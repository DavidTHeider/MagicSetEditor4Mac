// Der Abgleich mit Scryfall für den Dienst auf der Rückschleife
// (#1005).
//
// **Hier wird nur übersetzt.** Was ein Vorschlag ist, weiß `internal/app`;
// wie er über das Protokoll geht, `internal/rest`. Dieses Verzeichnis kennt
// `service` nicht (§2) — sonst stünde die Fachlichkeit in drei Composition
// Roots.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// fremdabgleich erfüllt `rest.Fremdabgleich` über der Naht von `internal/app`.
type fremdabgleich struct{ db *sql.DB }

// Vorschlag sagt, was sich ändern würde — und ändert nichts.
func (f fremdabgleich) Vorschlag(
	ctx context.Context, setID, karteID int64, name string,
) (rest.Vorschlag, error) {
	naht, err := app.Scryfallabgleichnaht(f.db, domain.UrheberDienst)
	if err != nil {
		return rest.Vorschlag{}, err
	}

	vorschlag, holfehler := naht.Vorschlag(ctx, setID, karteID, name)
	if holfehler != nil {
		return rest.Vorschlag{}, holfehler
	}

	return rest.Vorschlag{
		Aenderungen:  alsRestAenderungen(vorschlag.Aenderungen),
		Befunde:      vorschlag.Befunde,
		Herkunft:     vorschlag.Herkunft,
		AlterStunden: vorschlag.AlterStunden,
	}, nil
}

// Uebernehmen schreibt die genannten Felder und gibt ihre vorigen Werte.
func (f fremdabgleich) Uebernehmen(
	ctx context.Context, setID, karteID int64, name string, felder []string,
) ([]rest.Aenderung, error) {
	naht, err := app.Scryfallabgleichnaht(f.db, domain.UrheberDienst)
	if err != nil {
		return nil, err
	}

	uebernommen, schreibfehler := naht.Uebernehmen(ctx, setID, karteID, name, felder)
	if schreibfehler != nil {
		return nil, schreibfehler
	}

	return alsRestAenderungen(uebernommen), nil
}

// alsRestAenderungen übersetzt die Feldvorschläge.
func alsRestAenderungen(aenderungen []app.FeldvorschlagDTO) []rest.Aenderung {
	gestalt := make([]rest.Aenderung, 0, len(aenderungen))
	for _, eine := range aenderungen {
		gestalt = append(gestalt, rest.Aenderung{Feld: eine.Feld, Alt: eine.Alt, Neu: eine.Neu})
	}

	return gestalt
}

// fremdbilder führt den Bildweg an die Naht von `internal/rest`.
type fremdbilder struct{ db *sql.DB }

// Einsetzen holt das Bild und legt es an der Karte ab.
func (f fremdbilder) Einsetzen(
	ctx context.Context, setID, karteID int64, feld, name, art string,
) (rest.Bildeinsatz, error) {
	naht, err := app.Fremdbildnaht(f.db, domain.UrheberDienst)
	if err != nil {
		return rest.Bildeinsatz{}, err
	}

	eingesetzt, bildfehler := naht.Einsetzen(ctx, setID, karteID, feld, name, art)
	if bildfehler != nil {
		return rest.Bildeinsatz{}, bildfehler
	}

	return rest.Bildeinsatz{
		Feld: eingesetzt.Feld, Kuenstler: eingesetzt.Kuenstler,
		Nennung: eingesetzt.Nennung, Adresse: eingesetzt.Adresse,
		IllustratorGesetzt: eingesetzt.IllustratorGesetzt,
		AlterStunden:       int(eingesetzt.AlterStunden),
		Befunde:            eingesetzt.Befunde,
	}, nil
}

// fremdauskunft führt die lesenden Auskünfte an die Naht von `internal/rest`.
// **Sie hält den Bestand, seit es Dienstschalter gibt** (Phase 4.8): dort
// steht, ob Scryfall gefragt werden darf.
type fremdauskunft struct{ db *sql.DB }

// Regelhinweise gibt die Hinweise zu einer Karte.
func (f fremdauskunft) Regelhinweise(ctx context.Context, name string) (rest.Regelauskunft, error) {
	naht, err := app.Fremdauskunftsnaht(f.db)
	if err != nil {
		return rest.Regelauskunft{}, err
	}

	geholt, holfehler := naht.Regelhinweise(ctx, name)
	if holfehler != nil {
		return rest.Regelauskunft{}, holfehler
	}

	hinweise := make([]rest.Regelhinweis, 0, len(geholt.Hinweise))
	for _, einer := range geholt.Hinweise {
		hinweise = append(hinweise, rest.Regelhinweis{
			Offiziell: einer.Offiziell, Quelle: einer.Quelle,
			Datum: einer.Datum, Text: einer.Text,
		})
	}

	return rest.Regelauskunft{
		Kartenname: geholt.Kartenname, Hinweise: hinweise, Adresse: geholt.Adresse,
		Befunde: geholt.Befunde, AlterStunden: geholt.AlterStunden,
	}, nil
}

// Manakosten prüft einen Kosten gegen die Symbologie.
func (f fremdauskunft) Manakosten(ctx context.Context, kosten string) ([]string, error) {
	naht, err := app.Fremdauskunftsnaht(f.db)
	if err != nil {
		return nil, err
	}

	return naht.Manakosten(ctx, kosten)
}

// fremdsammeldaten führt die Sammeldateien an die Naht von `internal/rest`.
// **Sie hält den Bestand, seit es Dienstschalter gibt** (Phase 4.8): dort
// steht, ob Scryfall gefragt werden darf.
type fremdsammeldaten struct{ db *sql.DB }

// Stand sagt, was auf der Platte liegt.
func (f fremdsammeldaten) Stand(ctx context.Context) (rest.Sammelstand, error) {
	return restsammelstand(ctx, f.db, false)
}

// Holen lädt die Sammeldatei.
func (f fremdsammeldaten) Holen(ctx context.Context) (rest.Sammelstand, error) {
	return restsammelstand(ctx, f.db, true)
}

// restsammelstand führt beide Wege.
func restsammelstand(ctx context.Context, db *sql.DB, holen bool) (
	rest.Sammelstand, error,
) {
	naht, err := app.Sammeldatennaht(app.NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(app.AutorisierungAus(db)))
	if err != nil {
		return rest.Sammelstand{}, err
	}

	stand, standfehler := naht.Stand(ctx)
	if holen {
		stand, standfehler = naht.Holen(ctx)
	}

	if standfehler != nil {
		return rest.Sammelstand{}, standfehler
	}

	return rest.Sammelstand{
		Art: stand.Art, Pfad: stand.Pfad, Vorhanden: stand.Vorhanden,
		Bytes: stand.Bytes, Angeboten: stand.Angeboten, Stand: stand.Stand,
	}, nil
}

// fremdanlage erfüllt `rest.Fremdanlage` über der Naht von `internal/app`
// (#1032).
type fremdanlage struct{ db *sql.DB }

// Anlegen legt eine Karte an und füllt sie aus Scryfall.
func (f fremdanlage) Anlegen(ctx context.Context, setID int64, name string) (
	rest.Anlage, error,
) {
	naht, err := app.Scryfallabgleichnaht(f.db, domain.UrheberDienst)
	if err != nil {
		return rest.Anlage{}, err
	}

	// **Die angelegte Karte reist auch im Fehlerfall mit** (#155): sie ist da,
	// und ein Aufrufer, der ihre Kennung nicht bekommt, kann sie nicht
	// zurücknehmen.
	ergebnis, anlegefehler := naht.Anlegen(ctx, setID, name)

	return rest.Anlage{
		KarteID: ergebnis.KarteID, Felder: ergebnis.Felder,
		Befunde: ergebnis.Befunde, Herkunft: ergebnis.Herkunft,
		AlterStunden: ergebnis.AlterStunden,
	}, anlegefehler
}

// fremdsaetze erfüllt `rest.Fremdsaetze` über der Naht von `internal/app`
// (#1033).
type fremdsaetze struct{ db *sql.DB }

// Einlesen liest einen ganzen Satz als Set ein.
func (f fremdsaetze) Einlesen(ctx context.Context, code, spiel string) (rest.Satz, error) {
	naht, err := app.ScryfallsatznahtUeber(f.db)
	if err != nil {
		return rest.Satz{}, err
	}

	ergebnis, einlesefehler := naht.Einlesen(ctx, code, spiel)

	return rest.Satz{
		SetID: ergebnis.SetID, Karten: ergebnis.Karten,
		Befunde: ergebnis.Befunde, SchonOffen: ergebnis.SchonOffen,
	}, einlesefehler
}
