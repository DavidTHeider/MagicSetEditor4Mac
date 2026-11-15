// **Nutzer und Profile für den Dienst** — Werkstück 1 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// nutzer erfüllt `rest.Fremdnutzer` über den Nähten von `internal/app`.
type nutzer struct{ db *sql.DB }

// Nutzer gibt alle Nutzer samt ihren Profilen.
func (n nutzer) Nutzer(ctx context.Context) ([]rest.Nutzer, error) {
	antwort := app.NutzerflaecheKopflos(n.db).Nutzer(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Nutzer, 0, len(antwort.Nutzer))
	for _, einer := range antwort.Nutzer {
		zeilen = append(zeilen, alsRestnutzer(einer))
	}

	return zeilen, nil
}

// NutzerAnlegen legt einen Nutzer samt seinem ersten Profil an.
func (n nutzer) NutzerAnlegen(ctx context.Context, name, bild string) (rest.Nutzer, error) {
	antwort := app.NutzerflaecheKopflos(n.db).NutzerAnlegen(ctx, name, bild)
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// ProfilAnlegen legt ein weiteres Profil an.
func (n nutzer) ProfilAnlegen(ctx context.Context, nutzerID, name string) (rest.Profil, error) {
	antwort := app.NutzerflaecheKopflos(n.db).ProfilAnlegen(ctx, nutzerID, name)
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}

// GezeigtesProfil sagt, in welchem Profil gearbeitet wird.
func (n nutzer) GezeigtesProfil(ctx context.Context) (rest.Profil, error) {
	antwort := app.NutzerflaecheKopflos(n.db).GezeigtesProfil(ctx)
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}

// ProfilWechseln stellt das Profil um, in dem gearbeitet wird.
func (n nutzer) ProfilWechseln(ctx context.Context, profilID, nachweis string) (
	rest.Profil, error,
) {
	// **Ohne Wechsel in ein Profil, in dem nur gelesen werden dürfte** —
	// Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// **Dieser Prozess hält eine Verbindung über seine ganze Laufzeit**, und
	// jede Naht hat sie sich beim Aufbau abgeholt; sie mitten im Betrieb auf
	// eine nur lesende umzuhängen hieße, sie alle neu zu bauen. Derselbe
	// Grund, aus dem hier auch das Zurückspielen fehlt.
	//
	// **Beim Start greift es nicht**: steht der Zeiger schon dort, öffnet
	// dieser Prozess seine Verbindung gleich nur lesend.
	antwort := app.NutzerflaecheKopflos(n.db).OhneLeserwechsel().
		ProfilWechseln(ctx, profilID, nachweis)
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}

// Anbieterkonten gibt die Konten eines Nutzers.
func (n nutzer) Anbieterkonten(ctx context.Context, nutzerID string) ([]rest.Konto, error) {
	antwort := app.NutzerflaecheKopflos(n.db).Anbieterkonten(ctx, nutzerID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	konten := make([]rest.Konto, 0, len(antwort.Konten))
	for _, eines := range antwort.Konten {
		konten = append(konten, rest.Konto{
			Anbieter: eines.Anbieter, Anzeigename: eines.Anzeigename,
		})
	}

	return konten, nil
}

// KontoLoesen nimmt eine fremde Kennung von einem Nutzer.
func (n nutzer) KontoLoesen(ctx context.Context, nutzerID, anbieter string) error {
	antwort := app.NutzerflaecheKopflos(n.db).KontoLoesen(ctx, nutzerID, anbieter)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// NutzerAendern schreibt Name und Bild eines Nutzers fort.
func (n nutzer) NutzerAendern(ctx context.Context, angaben rest.Nutzerangaben) (
	rest.Nutzer, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).NutzerAendern(ctx, app.NutzerangabenDTO{
		ID:           angaben.NutzerID,
		Name:         angaben.Name,
		Bild:         angaben.Bild,
		Anrede:       angaben.Anrede,
		Namenszusatz: angaben.Namenszusatz,
		Vorname:      angaben.Vorname,
		Weiterenamen: angaben.Weiterenamen,
		Nachname:     angaben.Nachname,
		Slogan:       angaben.Slogan,
		Kurzebio:     angaben.Kurzebio,
		Langebio:     angaben.Langebio,
		Fassung:      angaben.Fassung,
	})
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// NutzerbildSetzen legt eine Bilddatei als Profilbild ab.
func (n nutzer) NutzerbildSetzen(ctx context.Context, nutzerID, fassung, pfad string) (
	rest.Nutzer, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).
		NutzerbildSetzen(ctx, nutzerID, fassung, pfad)
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// NutzerbildEntfernen nimmt das Profilbild eines Nutzers wieder weg.
func (n nutzer) NutzerbildEntfernen(ctx context.Context, nutzerID, fassung string) (
	rest.Nutzer, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).NutzerbildEntfernen(ctx, nutzerID, fassung)
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// ProfilAendern schreibt den Namen eines Profils fort.
func (n nutzer) ProfilAendern(ctx context.Context, angaben rest.Profilangaben) (
	rest.Profil, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).ProfilAendern(ctx, app.ProfilangabenDTO{
		ID:                angaben.ProfilID,
		Name:              angaben.Name,
		Subslogan:         angaben.Subslogan,
		Kurzebeschreibung: angaben.Kurzebeschreibung,
		Langebeschreibung: angaben.Langebeschreibung,
		Badge:             angaben.Badge,
		Fassung:           angaben.Fassung,
	})
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}

// NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
func (n nutzer) NutzerLoeschen(ctx context.Context, nutzerID, name string) error {
	antwort := app.NutzerflaecheKopflos(n.db).NutzerLoeschen(ctx, nutzerID, name)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// ProfilLoeschen nimmt ein Profil heraus.
func (n nutzer) ProfilLoeschen(ctx context.Context, profilID, name string) error {
	antwort := app.NutzerflaecheKopflos(n.db).ProfilLoeschen(ctx, profilID, name)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// alsRestnutzer übersetzt ein DTO der Fassade.
func alsRestnutzer(einer app.NutzerDTO) rest.Nutzer {
	profile := make([]rest.Profil, 0, len(einer.Profile))
	for _, eines := range einer.Profile {
		profile = append(profile, alsRestprofil(eines))
	}

	return rest.Nutzer{
		ID: einer.ID, Name: einer.Name, Bild: einer.Bild,
		Email: einer.Email, Telefon: einer.Telefon, Anonym: einer.Anonym,

		Anrede: einer.Anrede, Namenszusatz: einer.Namenszusatz,
		Vorname: einer.Vorname, Weiterenamen: einer.Weiterenamen,
		Nachname: einer.Nachname, Slogan: einer.Slogan,
		Kurzebio: einer.Kurzebio, Langebio: einer.Langebio,
		Fassung: einer.Fassung,

		Profile: profile,
	}
}

// alsRestprofil übersetzt ein Profil-DTO.
func alsRestprofil(eines app.ProfilDTO) rest.Profil {
	return rest.Profil{
		ID: eines.ID, NutzerID: eines.NutzerID, Name: eines.Name,

		Subslogan:         eines.Subslogan,
		Kurzebeschreibung: eines.Kurzebeschreibung,
		Langebeschreibung: eines.Langebeschreibung,
		Badge:             eines.Badge,

		Fassung: eines.Fassung,
	}
}

// KontoEintragen legt ein Konto unter einer Adresse oder einer Rufnummer an.
func (n nutzer) KontoEintragen(ctx context.Context, kennung, kennwort, name string) (
	rest.Nutzer, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).KontoEintragen(ctx, kennung, kennwort, name)
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// KontoAnmelden sagt, wer unter dieser Kennung mit diesem Kennwort arbeitet.
func (n nutzer) KontoAnmelden(ctx context.Context, kennung, kennwort, nachweis string) (
	rest.Nutzer, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).
		KontoAnmelden(ctx, kennung, kennwort, nachweis)
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// KontoAnonym legt ein Konto ohne jede Angabe an.
func (n nutzer) KontoAnonym(ctx context.Context) (rest.Nutzer, error) {
	antwort := app.NutzerflaecheKopflos(n.db).KontoAnonym(ctx)
	if antwort.Fehler != nil {
		return rest.Nutzer{}, antwort.Fehler.AlsFehler()
	}

	return alsRestnutzer(*antwort.Nutzer), nil
}

// Faktoren gibt die zweiten Faktoren eines Nutzers samt der Rechnerauskunft.
func (n nutzer) Faktoren(ctx context.Context, nutzerID string) (rest.Faktorstand, error) {
	antwort := app.NutzerflaecheKopflos(n.db).Faktoren(ctx, nutzerID)
	if antwort.Fehler != nil {
		return rest.Faktorstand{}, antwort.Fehler.AlsFehler()
	}

	faktoren := make([]rest.Faktor, 0, len(antwort.Faktoren))
	for _, einer := range antwort.Faktoren {
		faktoren = append(faktoren, rest.Faktor{
			Art: einer.Art, Bezeichnung: einer.Bezeichnung,
		})
	}

	return rest.Faktorstand{
		Faktoren:       faktoren,
		Biometrie:      antwort.Biometrie,
		Biometrieart:   antwort.Biometrieart,
		Biometriegrund: antwort.Biometriegrund,
	}, nil
}

// FaktorEinrichten zieht ein TOTP-Geheimnis.
func (n nutzer) FaktorEinrichten(ctx context.Context, nutzerID string) (
	rest.Faktoreinrichtung, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).FaktorEinrichten(ctx, nutzerID)
	if antwort.Fehler != nil {
		return rest.Faktoreinrichtung{}, antwort.Fehler.AlsFehler()
	}

	return rest.Faktoreinrichtung{
		Adresse: antwort.Adresse, Geheimnis: antwort.Geheimnis, Codes: antwort.Codes,
	}, nil
}

// FaktorBestaetigen trägt den TOTP-Faktor ein.
func (n nutzer) FaktorBestaetigen(ctx context.Context, nutzerID, code string) error {
	antwort := app.NutzerflaecheKopflos(n.db).FaktorBestaetigen(ctx, nutzerID, code)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// BiometrieEinrichten trägt Touch ID oder Face ID als zweiten Faktor ein.
func (n nutzer) BiometrieEinrichten(ctx context.Context, nutzerID string) error {
	antwort := app.NutzerflaecheKopflos(n.db).BiometrieEinrichten(ctx, nutzerID)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// FaktorEntfernen nimmt einen zweiten Faktor heraus.
func (n nutzer) FaktorEntfernen(ctx context.Context, nutzerID, art string) error {
	antwort := app.NutzerflaecheKopflos(n.db).FaktorEntfernen(ctx, nutzerID, art)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// ProfilbadgeSetzen legt eine Bilddatei als Abzeichen eines Profils ab.
func (n nutzer) ProfilbadgeSetzen(ctx context.Context, profilID, fassung, pfad string) (
	rest.Profil, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).
		ProfilbadgeSetzen(ctx, profilID, fassung, pfad)
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}

// ProfilbadgeEntfernen nimmt das Abzeichen eines Profils wieder weg.
func (n nutzer) ProfilbadgeEntfernen(ctx context.Context, profilID, fassung string) (
	rest.Profil, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).ProfilbadgeEntfernen(ctx, profilID, fassung)
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}

// WerArbeitet nennt den Menschen, der gerade arbeitet — Werkstück 17 von Phase
// 4.12 (#1400).
func (n nutzer) WerArbeitet(ctx context.Context) (string, error) {
	antwort := app.NutzerflaecheKopflos(n.db).WerArbeitet(ctx)
	if antwort.Fehler != nil {
		return "", antwort.Fehler.AlsFehler()
	}

	return antwort.NutzerID, nil
}

// Arbeitskontext sagt, wer wo als was arbeitet — Werkstück O1 von Phase 5.2
// (#1425).
func (n nutzer) Arbeitskontext(ctx context.Context) (rest.Arbeitskontext, error) {
	antwort := app.NutzerflaecheKopflos(n.db).Arbeitskontext(ctx)
	if antwort.Fehler != nil {
		return rest.Arbeitskontext{}, antwort.Fehler.AlsFehler()
	}

	return rest.Arbeitskontext{
		NutzerID: antwort.NutzerID, ProfilID: antwort.ProfilID, EignerID: antwort.EignerID,
		Rolle: antwort.Rolle, NurLesend: antwort.NurLesend, Modus: antwort.Modus,
		Revision: antwort.Revision,
	}, nil
}

// NutzerWechseln stellt um, wer arbeitet — gegen den Ausweis des Ziels.
func (n nutzer) NutzerWechseln(ctx context.Context, nutzerID, nachweis string) (
	rest.Profil, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).NutzerWechseln(ctx, nutzerID, nachweis)
	if antwort.Fehler != nil {
		return rest.Profil{}, antwort.Fehler.AlsFehler()
	}

	return alsRestprofil(*antwort.Profil), nil
}
