package service_test

// **Der Bestand, den die Rücksetzung sieht** — Werkstück O4 von Phase 5.2
// (#1428).
//
// Aus `ruecksetzung_test.go` herausgelöst, als es über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort stehen die **Läufe**, hier das Doppel,
// gegen das sie laufen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ruecksetzungsdoppel spielt Bestand.
type ruecksetzungsdoppel struct {
	umfang  []domain.Betroffenes
	offen   int
	nutzer  int
	profile []domain.Profil
	fremde  []domain.Profil

	// fremdeZugaenge sind die Zugänge des Menschen an fremden Profilen.
	fremdeZugaenge int
	zurueck        int
	geloescht      string
	neuanfaenge    int
	fehler         error

	// beimLoeschen ist der Umfang, den das Schreibgeschäft **dann** sieht.
	//
	// **Nil heißt: derselbe wie in der Vorschau** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2. Wer ihn setzt, spielt den Fall, für den es die
	// Bindung gibt: zwischen Hinsehen und Löschen kam etwas dazu.
	beimLoeschen []domain.Betroffenes
}

// gebunden ruft die Bindung so, wie die Ablage es im Geschäft täte.
func (r *ruecksetzungsdoppel) gebunden(
	profile []domain.Profil,
	bindung []func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	if len(bindung) == 0 || bindung[0] == nil {
		return nil
	}

	umfaenge := make([][]domain.Betroffenes, 0, len(profile))
	for range profile {
		umfang := r.beimLoeschen
		if umfang == nil {
			umfang = r.umfang
		}

		umfaenge = append(umfaenge, umfang)
	}

	return bindung[0](profile, umfaenge)
}

// zielprofile sind die Profile, die das Schreibgeschäft darin sähe.
func (r *ruecksetzungsdoppel) zielprofile(profilID int64) []domain.Profil {
	if profilID == 0 {
		return r.profile
	}

	for _, profil := range r.profile {
		if profil.ID == profilID {
			return []domain.Profil{profil}
		}
	}

	return nil
}

func (r *ruecksetzungsdoppel) Profilumfang(context.Context, int64) (
	[]domain.Betroffenes, error,
) {
	return r.umfang, r.fehler
}

func (r *ruecksetzungsdoppel) OffeneRueckkehrmengen(context.Context, int64) (int, error) {
	return r.offen, r.fehler
}

func (r *ruecksetzungsdoppel) Nutzerzahl(context.Context) (int, error) {
	return r.nutzer, r.fehler
}

func (r *ruecksetzungsdoppel) Profile(context.Context, int64) ([]domain.Profil, error) {
	return r.profile, r.fehler
}

func (r *ruecksetzungsdoppel) AlleProfile(context.Context) ([]domain.Profil, error) {
	return append(r.profile, r.fremde...), r.fehler
}

func (r *ruecksetzungsdoppel) FremdeZugaenge(context.Context, int64) (domain.Betroffenes, error) {
	return domain.Betroffenes{
		Was: domain.BetroffenZugaenge, Anzahl: r.fremdeZugaenge, Namen: []string{},
	}, r.fehler
}

func (r *ruecksetzungsdoppel) EinstellungenZuruecksetzen(
	_ context.Context, schluessel []string,
) (int, error) {
	r.zurueck = len(schluessel)

	if len(schluessel) == 0 {
		r.zurueck = -1
	}

	return r.zurueck, r.fehler
}

func (r *ruecksetzungsdoppel) NutzereinstellungenZuruecksetzen(
	context.Context, int64,
) (int, error) {
	r.zurueck = 7

	return r.zurueck, r.fehler
}

func (r *ruecksetzungsdoppel) ProfilLoeschen(_ context.Context, profilID int64, name string,
	bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	if gebunden := r.gebunden(r.zielprofile(profilID), bindung); gebunden != nil {
		return gebunden
	}

	r.geloescht = "profil:" + name

	return r.fehler
}

func (r *ruecksetzungsdoppel) NutzerLoeschen(_ context.Context, _ int64, name string,
	bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
) error {
	if gebunden := r.gebunden(r.zielprofile(0), bindung); gebunden != nil {
		return gebunden
	}

	r.geloescht = "nutzer:" + name

	return r.fehler
}

func (r *ruecksetzungsdoppel) neuanfang(context.Context) (string, error) {
	r.neuanfaenge++

	return "/wo/der/alte/liegt.db", r.fehler
}
