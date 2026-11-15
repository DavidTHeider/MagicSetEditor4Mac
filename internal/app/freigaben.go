// **Freigaben am Fenster und an den Flächen** — Werkstück 11 von Phase 4.12
// (#1351).
//
// # Warum die Antwort die eigene Rolle mitträgt
//
// **Weil die Seite sonst raten müsste, was sie zeigen darf.** Ein Knopf zum
// Vergeben einer Freigabe, den nur ein Verwalter drücken kann, gehört einem
// Leser gar nicht erst hingestellt — und die Antwort auf „darf ich das" steht
// ohnehin schon in derselben Abfrage.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FreigabeDTO ist eine Freigabe, wie eine Fläche sie sieht.
type FreigabeDTO struct {
	ProfilID string `json:"profilID"`
	NutzerID string `json:"nutzerID"`

	// Name ist der Anzeigename des Nutzers — eine Kennung allein sagt nichts.
	Name string `json:"name"`

	// Rolle ist `leser`, `mitarbeiter` oder `verwalter`.
	Rolle string `json:"rolle"`

	Angelegt string `json:"angelegt"`
}

// FreigabenAntwort trägt die Freigaben eines Profils samt der eigenen Rolle.
type FreigabenAntwort struct {
	Freigaben []FreigabeDTO `json:"freigaben"`

	// Sichtbarkeit ist `privat` oder `bestand`.
	Sichtbarkeit string `json:"sichtbarkeit"`

	// EigeneRolle sagt, was der arbeitende Mensch an diesem Profil darf —
	// leer heißt: gar nichts.
	EigeneRolle string `json:"eigeneRolle"`

	// Nachladefehler sagt: die Änderung **gilt**, nur die Liste danach war
	// nicht zu lesen (§7, Regel 3; #1970) — etwa, weil sich jemand selbst
	// herabgestuft hat. Dann ist `freigaben` leer und `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// FreigabeangabenDTO ist, was eine Fläche an einer Freigabe setzen darf.
type FreigabeangabenDTO struct {
	ProfilID string `json:"profilID"`
	NutzerID string `json:"nutzerID"`
	Rolle    string `json:"rolle"`
}

// Freigabenfuehrer ist die Naht zur Freigabenverwaltung.
type Freigabenfuehrer interface {
	Freigaben(ctx context.Context, profilID int64) (
		[]domain.Profilfreigabe, error)
	Rolle(ctx context.Context, nutzerID int64, profil domain.Profil) (
		domain.Rolle, error)
	Verlangt(ctx context.Context, nutzerID int64, profil domain.Profil,
		verlangt domain.Rolle) error
	FreigabeSetzen(ctx context.Context, freigabe domain.Profilfreigabe) error
	FreigabeLoeschen(ctx context.Context, profilID, nutzerID int64) error
}

// MitFreigaben setzt die Naht.
func (a *App) MitFreigaben(fuehrer Freigabenfuehrer) { a.freigaben = fuehrer }

// Freigaben gibt die Freigaben eines Profils samt der eigenen Rolle.
func (a *App) Freigaben(ctx context.Context, profilID string) FreigabenAntwort {
	ctx = a.imProfil(ctx)

	leer := []FreigabeDTO{}

	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return FreigabenAntwort{Freigaben: leer, Fehler: befund}
	}

	profil, fehler := a.profilMitKennung(ctx, kennung)
	if fehler != nil {
		return FreigabenAntwort{Freigaben: leer, Fehler: Uebersetze(ctx, fehler)}
	}

	// **Die eigene Rolle kommt in derselben Antwort** — die Seite entscheidet
	// daran, was sie überhaupt anbietet (§7, Regel 2).
	eigene, rollenfehler := a.eigeneRolle(ctx, profil)
	if rollenfehler != nil {
		return FreigabenAntwort{Freigaben: leer, Fehler: Uebersetze(ctx, rollenfehler)}
	}

	// **Wer darin steht, liest, wer außer dem Eigner an einem Profil arbeitet
	// — an einem anderen nur, wer dort mitarbeitet** (#1817). Die Liste nennt
	// Menschen und ihre Rollen; aus dem eigenen Profil heraus las sie jeder
	// für jedes. **Mit der Rolle von eben** — ein zweites Nachschlagen fragte
	// dasselbe noch einmal.
	if fehler := lesbarMit(ctx, kennung, eigene); fehler != nil {
		return FreigabenAntwort{Freigaben: leer, Fehler: Uebersetze(ctx, fehler)}
	}

	gefunden, lesefehler := a.freigaben.Freigaben(ctx, kennung)
	if lesefehler != nil {
		return FreigabenAntwort{Freigaben: leer, Fehler: Uebersetze(ctx, lesefehler)}
	}

	zeilen, namensfehler := a.freigabenzeilen(ctx, gefunden)
	if namensfehler != nil {
		return FreigabenAntwort{Freigaben: leer, Fehler: Uebersetze(ctx, namensfehler)}
	}

	return FreigabenAntwort{
		Freigaben:    zeilen,
		Sichtbarkeit: profil.Sichtbarkeit,
		EigeneRolle:  string(eigene),
	}
}

// FreigabeSetzen legt eine Freigabe an oder schreibt ihre Rolle fort.
//
// **Nur ein Verwalter darf das** — und der Eigner ist einer, ohne Zeile im
// Bestand.
func (a *App) FreigabeSetzen(
	ctx context.Context, angaben FreigabeangabenDTO,
) FreigabenAntwort {
	ctx = a.imProfil(ctx)

	profilID, nutzerID, befund := a.beideKennungen(angaben)
	if befund != nil {
		return FreigabenAntwort{Freigaben: []FreigabeDTO{}, Fehler: befund}
	}

	if fehler := a.alsVerwalter(ctx, profilID); fehler != nil {
		return FreigabenAntwort{
			Freigaben: []FreigabeDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	if fehler := a.freigaben.FreigabeSetzen(ctx, domain.Profilfreigabe{
		ProfilID: profilID, NutzerID: nutzerID,
		Rolle: domain.Rolle(angaben.Rolle),
	}); fehler != nil {
		return FreigabenAntwort{
			Freigaben: []FreigabeDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return a.freigabenNachDemSchreiben(ctx, angaben.ProfilID)
}

// FreigabeLoeschen zieht eine Freigabe zurück.
func (a *App) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID string,
) FreigabenAntwort {
	ctx = a.imProfil(ctx)

	angaben := FreigabeangabenDTO{ProfilID: profilID, NutzerID: nutzerID}

	profil, nutzer, befund := a.beideKennungen(angaben)
	if befund != nil {
		return FreigabenAntwort{Freigaben: []FreigabeDTO{}, Fehler: befund}
	}

	if fehler := a.alsVerwalter(ctx, profil); fehler != nil {
		return FreigabenAntwort{
			Freigaben: []FreigabeDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	if fehler := a.freigaben.FreigabeLoeschen(ctx, profil, nutzer); fehler != nil {
		return FreigabenAntwort{
			Freigaben: []FreigabeDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return a.freigabenNachDemSchreiben(ctx, profilID)
}

// freigabenNachDemSchreiben gibt die Liste nach einer **geglückten** Änderung.
//
// **Geschrieben ist geschrieben** (§7, Regel 3;
// #1970):
// wer sich selbst herabstuft oder seine Freigabe löscht, darf die Liste
// danach nicht mehr lesen — und das Nein dazu ist kein Nein zur Änderung.
// Es steht in [FreigabenAntwort.Nachladefehler].
func (a *App) freigabenNachDemSchreiben(
	ctx context.Context, profilID string,
) FreigabenAntwort {
	antwort := a.Freigaben(ctx, profilID)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// beideKennungen liest Profil- und Nutzerkennung.
func (a *App) beideKennungen(angaben FreigabeangabenDTO) (
	profilID, nutzerID int64, befund *Fehler,
) {
	profilID, befund = benannteKennungAus("profilID", angaben.ProfilID)
	if befund != nil {
		return 0, 0, befund
	}

	nutzerID, befund = benannteKennungAus("nutzerID", angaben.NutzerID)
	if befund != nil {
		return 0, 0, befund
	}

	return profilID, nutzerID, nil
}

// alsVerwalter weist ab, wer keiner ist — am genannten Profil
// ([App.rolleAm]).
func (a *App) alsVerwalter(ctx context.Context, profilID int64) error {
	return a.rolleAm(ctx, profilID, domain.RolleVerwalter)
}

// eigeneRolle gibt die Rolle des arbeitenden Menschen an einem Profil.
func (a *App) eigeneRolle(ctx context.Context, profil domain.Profil) (
	domain.Rolle, error,
) {
	arbeitend, fehler := a.arbeitenderNutzer(ctx)
	if fehler != nil {
		return "", fehler
	}

	return a.freigaben.Rolle(ctx, arbeitend, profil)
}

// arbeitenderNutzer gibt den Menschen, der gerade arbeitet.
//
// **Seit Werkstück 17 von Phase 4.12 nicht mehr der Eigner des gezeigten
// Profils** (#1400).
// Hier stand die Ableitung, und [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)
// nannte sie als Tatsache: „diese Anwendung hat keinen Begriff einer Sitzung
// neben dem Profilzeiger". Sie hat jetzt einen — und ohne ihn zeigte diese
// Seite dem, der in ein fremdes Profil gewechselt ist, die Rolle **des
// Eigners**: Verwalter.
func (a *App) arbeitenderNutzer(ctx context.Context) (int64, error) {
	return a.profilwechsel.WerArbeitet(ctx)
}

// profilMitKennung sucht ein Profil über die Nutzerübersicht.
func (a *App) profilMitKennung(ctx context.Context, profilID int64) (
	domain.Profil, error,
) {
	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	for _, zeile := range zeilen {
		for _, profil := range zeile.Profile {
			if profil.ID == profilID {
				return profil, nil
			}
		}
	}

	return domain.Profil{}, fmt.Errorf(
		"%w: das Profil %d führt der Bestand nicht", domain.ErrNotFound, profilID)
}

// freigabenzeilen bildet die Freigaben ab — **samt dem Namen des Nutzers**.
//
// **Eine Kennung allein sagt nichts**: `2` ist kein Mensch, und eine Liste
// daraus ließe sich nicht bedienen.
func (a *App) freigabenzeilen(
	ctx context.Context, gefunden []domain.Profilfreigabe,
) ([]FreigabeDTO, error) {
	zeilen, fehler := a.nutzer.Uebersicht(ctx)
	if fehler != nil {
		return nil, fehler
	}

	namen := map[int64]string{}
	for _, zeile := range zeilen {
		namen[zeile.Nutzer.ID] = zeile.Nutzer.Name
	}

	abgebildet := make([]FreigabeDTO, 0, len(gefunden))

	for _, eine := range gefunden {
		abgebildet = append(abgebildet, FreigabeDTO{
			ProfilID: zahltext(eine.ProfilID),
			NutzerID: zahltext(eine.NutzerID),
			Name:     namen[eine.NutzerID],
			Rolle:    string(eine.Rolle),
			Angelegt: eine.Angelegt,
		})
	}

	return abgebildet, nil
}
