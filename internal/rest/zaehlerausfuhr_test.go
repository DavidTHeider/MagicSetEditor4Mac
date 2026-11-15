// **Die zählenden Antworten der Ausfuhr** — einfahren, sichern, zeichnen, den
// Bogen legen, MTG JSON, die Paketpflege.
//
// **Eine eigene Datei**, seit `zaehlerdoppel_test.go` mit der Restparität über
// die Zeilengrenze aus §2 lief
// (#1291).
// Der Schnitt folgt der Frage: dort zählt das Double, **was am Bestand**
// geschieht, hier, **was die Platte verlässt oder auf sie kommt**.
package rest_test

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (z *zaehlender) Einfahren(
	ctx context.Context, pfad string,
) (setID int64, karten int, befunde []string, fehler error) {
	z.zaehle()

	return z.williger.Einfahren(ctx, pfad)
}

func (z *zaehlender) Sichern(ctx context.Context, setID int64, ziel string) ([]string, error) {
	z.zaehle()

	return z.williger.Sichern(ctx, setID, ziel)
}

func (z *zaehlender) Zeichnen(
	ctx context.Context, karteID int64, ziel, guete string,
) (breite, hoehe int, befunde []string, fehler error) {
	z.zaehle()

	return z.williger.Zeichnen(ctx, karteID, ziel, guete)
}

func (z *zaehlender) Bogen(
	ctx context.Context, setID int64, ziel string,
) (seiten, karten int, befunde []string, fehler error) {
	z.zaehle()

	return z.williger.Bogen(ctx, setID, ziel)
}

func (z *zaehlender) MTGJSONLesen(ctx context.Context, pfad, spiel string) (
	angelegt []int64, befunde []string, fehler error,
) {
	z.zaehle()

	return z.williger.MTGJSONLesen(ctx, pfad, spiel)
}

func (z *zaehlender) MTGJSONSchreiben(ctx context.Context, setID int64, spiel, ziel string) (
	karten int, befunde []string, fehler error,
) {
	z.zaehle()

	return z.williger.MTGJSONSchreiben(ctx, setID, spiel, ziel)
}

func (z *zaehlender) MTGJSONSpiele() []string {
	z.zaehle()

	return z.williger.MTGJSONSpiele()
}

func (z *zaehlender) PaketInstallieren(ctx context.Context, quelle string) (rest.Paket, error) {
	z.zaehle()

	return z.williger.PaketInstallieren(ctx, quelle)
}

func (z *zaehlender) PaketAnlegen(ctx context.Context, art, name, voll, spiel string) (
	rest.Paket, error,
) {
	z.zaehle()

	return z.williger.PaketAnlegen(ctx, art, name, voll, spiel)
}

func (z *zaehlender) PaketKopieren(ctx context.Context, art, name, neu string) (rest.Paket, error) {
	z.zaehle()

	return z.williger.PaketKopieren(ctx, art, name, neu)
}

func (z *zaehlender) Nutzer(ctx context.Context) ([]rest.Nutzer, error) {
	z.zaehle()

	return z.williger.Nutzer(ctx)
}

func (z *zaehlender) NutzerAnlegen(ctx context.Context, name, bild string) (rest.Nutzer, error) {
	z.zaehle()

	return z.williger.NutzerAnlegen(ctx, name, bild)
}

func (z *zaehlender) ProfilAnlegen(ctx context.Context, nutzerID, name string) (rest.Profil, error) {
	z.zaehle()

	return z.williger.ProfilAnlegen(ctx, nutzerID, name)
}

func (z *zaehlender) GezeigtesProfil(ctx context.Context) (rest.Profil, error) {
	z.zaehle()

	return z.williger.GezeigtesProfil(ctx)
}

func (z *zaehlender) ProfilWechseln(ctx context.Context, profilID, nachweis string) (
	rest.Profil, error,
) {
	z.zaehle()

	return z.williger.ProfilWechseln(ctx, profilID, nachweis)
}

func (z *zaehlender) WerArbeitet(ctx context.Context) (string, error) {
	z.zaehle()

	return z.williger.WerArbeitet(ctx)
}

func (z *zaehlender) Arbeitskontext(ctx context.Context) (rest.Arbeitskontext, error) {
	z.zaehle()

	return z.williger.Arbeitskontext(ctx)
}

func (z *zaehlender) NutzerWechseln(ctx context.Context, nutzerID, nachweis string) (
	rest.Profil, error,
) {
	z.zaehle()

	return z.williger.NutzerWechseln(ctx, nutzerID, nachweis)
}

func (z *zaehlender) Anbieterkonten(ctx context.Context, nutzerID string) (
	[]rest.Konto, error,
) {
	z.zaehle()

	return z.williger.Anbieterkonten(ctx, nutzerID)
}

func (z *zaehlender) KontoLoesen(ctx context.Context, nutzerID, anbieter string) error {
	z.zaehle()

	return z.williger.KontoLoesen(ctx, nutzerID, anbieter)
}

func (z *zaehlender) NutzerAendern(ctx context.Context, angaben rest.Nutzerangaben) (
	rest.Nutzer, error,
) {
	z.zaehle()

	return z.williger.NutzerAendern(ctx, angaben)
}

func (z *zaehlender) ProfilAendern(ctx context.Context, angaben rest.Profilangaben) (
	rest.Profil, error,
) {
	z.zaehle()

	return z.williger.ProfilAendern(ctx, angaben)
}

func (z *zaehlender) NutzerLoeschen(ctx context.Context, nutzerID, name string) error {
	z.zaehle()

	return z.williger.NutzerLoeschen(ctx, nutzerID, name)
}

func (z *zaehlender) ProfilLoeschen(ctx context.Context, profilID, name string) error {
	z.zaehle()

	return z.williger.ProfilLoeschen(ctx, profilID, name)
}

func (z *zaehlender) KontoEintragen(ctx context.Context, kennung, kennwort, name string) (
	rest.Nutzer, error,
) {
	z.zaehle()

	return z.williger.KontoEintragen(ctx, kennung, kennwort, name)
}

func (z *zaehlender) KontoAnmelden(ctx context.Context, kennung, kennwort, nachweis string) (
	rest.Nutzer, error,
) {
	z.zaehle()

	return z.williger.KontoAnmelden(ctx, kennung, kennwort, nachweis)
}

func (z *zaehlender) Faktoren(ctx context.Context, nutzerID string) (
	rest.Faktorstand, error,
) {
	z.zaehle()

	return z.williger.Faktoren(ctx, nutzerID)
}

func (z *zaehlender) FaktorEinrichten(ctx context.Context, nutzerID string) (
	rest.Faktoreinrichtung, error,
) {
	z.zaehle()

	return z.williger.FaktorEinrichten(ctx, nutzerID)
}

func (z *zaehlender) FaktorBestaetigen(ctx context.Context, nutzerID, code string) error {
	z.zaehle()

	return z.williger.FaktorBestaetigen(ctx, nutzerID, code)
}

func (z *zaehlender) BiometrieEinrichten(ctx context.Context, nutzerID string) error {
	z.zaehle()

	return z.williger.BiometrieEinrichten(ctx, nutzerID)
}

func (z *zaehlender) FaktorEntfernen(ctx context.Context, nutzerID, art string) error {
	z.zaehle()

	return z.williger.FaktorEntfernen(ctx, nutzerID, art)
}

func (z *zaehlender) KontoAnonym(ctx context.Context) (rest.Nutzer, error) {
	z.zaehle()

	return z.williger.KontoAnonym(ctx)
}

func (z *zaehlender) PaketEntfernen(ctx context.Context, art, name, pfad string) error {
	z.zaehle()

	return z.williger.PaketEntfernen(ctx, art, name, pfad)
}

func (z *zaehlender) PaketUebernehmen(ctx context.Context, ort string) (rest.Uebernahme, error) {
	z.zaehle()

	return z.williger.PaketUebernehmen(ctx, ort)
}

func (z *zaehlender) MSE2Orte() []string {
	z.zaehle()

	return z.williger.MSE2Orte()
}

func (z *zaehlender) AusDatei(
	ctx context.Context, setID, karteID int64, feld, pfad string,
) (rest.Bildmasse, error) {
	z.zaehle()

	return z.williger.AusDatei(ctx, setID, karteID, feld, pfad)
}

func (z *zaehlender) Vorlagen(ctx context.Context, setID int64) (
	[]rest.Vorlagenzeile, []string, error,
) {
	z.zaehle()

	return z.williger.Vorlagen(ctx, setID)
}

func (z *zaehlender) Ausfuehren(
	ctx context.Context, setID int64, paket, ziel string, wahl map[string]string,
) ([]string, error) {
	z.zaehle()

	return z.williger.Ausfuehren(ctx, setID, paket, ziel, wahl)
}

func (z *zaehlender) VorlagenoptionSetzen(ctx context.Context, vorlage, feld, wert string) error {
	z.zaehle()

	return z.williger.VorlagenoptionSetzen(ctx, vorlage, feld, wert)
}

// **Das Profilbild** — Werkstück 8 von Phase 4.12 (#1351).
func (z *zaehlender) NutzerbildSetzen(ctx context.Context, nutzerID, fassung, pfad string) (
	rest.Nutzer, error,
) {
	z.zaehle()

	return z.williger.NutzerbildSetzen(ctx, nutzerID, fassung, pfad)
}

func (z *zaehlender) NutzerbildEntfernen(ctx context.Context, nutzerID, fassung string) (
	rest.Nutzer, error,
) {
	z.zaehle()

	return z.williger.NutzerbildEntfernen(ctx, nutzerID, fassung)
}

// **Die Einstellungen und die Auskunft** — Werkstück 8 desselben Issues.
func (z *zaehlender) Nutzereinstellungen(ctx context.Context, nutzerID string) (
	[]rest.Einstellungswahl, error,
) {
	z.zaehle()

	return z.williger.Nutzereinstellungen(ctx, nutzerID)
}

func (z *zaehlender) NutzereinstellungSetzen(
	ctx context.Context, nutzerID, schluessel, wert string,
) ([]rest.Einstellungswahl, error) {
	z.zaehle()

	return z.williger.NutzereinstellungSetzen(ctx, nutzerID, schluessel, wert)
}

func (z *zaehlender) Personendaten(ctx context.Context, nutzerID string) (
	json.RawMessage, error,
) {
	z.zaehle()

	return z.williger.Personendaten(ctx, nutzerID)
}

// **Das Abzeichen eines Profils** — Werkstück 9 desselben Issues.
func (z *zaehlender) ProfilbadgeSetzen(ctx context.Context, profilID, fassung, pfad string) (
	rest.Profil, error,
) {
	z.zaehle()

	return z.williger.ProfilbadgeSetzen(ctx, profilID, fassung, pfad)
}

func (z *zaehlender) ProfilbadgeEntfernen(ctx context.Context, profilID, fassung string) (
	rest.Profil, error,
) {
	z.zaehle()

	return z.williger.ProfilbadgeEntfernen(ctx, profilID, fassung)
}
