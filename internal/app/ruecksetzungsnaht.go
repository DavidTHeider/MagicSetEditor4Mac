package app

// **Die Naht zur Rücksetzung** — Werkstück O4 von Phase 5.2
// (#1428).
//
// # Warum sie die Verbindung holt und nicht hält
//
// **Weil ein Neuanfang sie tauscht** — derselbe Grund wie bei `pflegenaht`,
// nur schärfer: hier wird der Bestand ausgewechselt, während die Naht steht.
// Wer die Verbindung festhielte, läse danach eine Datei, die es nicht mehr
// gibt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ruecksetzungsablage bringt die Zählungen aus `storage` an die Schnittstelle.
type ruecksetzungsablage struct{ holen func() *sql.DB }

func (r ruecksetzungsablage) Profilumfang(ctx context.Context, profilID int64) (
	[]domain.Betroffenes, error,
) {
	return storage.Profilumfang(ctx, r.holen(), profilID)
}

func (r ruecksetzungsablage) OffeneRueckkehrmengen(ctx context.Context, profilID int64) (
	int, error,
) {
	return storage.OffeneRueckkehrmengen(ctx, r.holen(), profilID)
}

func (r ruecksetzungsablage) Nutzerzahl(ctx context.Context) (int, error) {
	return storage.Nutzerzahl(ctx, r.holen())
}

func (r ruecksetzungsablage) Profile(ctx context.Context, nutzerID int64) (
	[]domain.Profil, error,
) {
	return storage.NeueNutzerablage(r.holen()).Profile(ctx, nutzerID)
}

func (r ruecksetzungsablage) AlleProfile(ctx context.Context) ([]domain.Profil, error) {
	return storage.AlleProfile(ctx, r.holen())
}

func (r ruecksetzungsablage) FremdeZugaenge(ctx context.Context, nutzerID int64) (
	domain.Betroffenes, error,
) {
	return storage.FremdeZugaenge(ctx, r.holen(), nutzerID)
}

func (r ruecksetzungsablage) EinstellungenZuruecksetzen(
	ctx context.Context, schluessel []string,
) (int, error) {
	return storage.EinstellungenZuruecksetzen(ctx, r.holen(), schluessel)
}

func (r ruecksetzungsablage) NutzereinstellungenZuruecksetzen(
	ctx context.Context, nutzerID int64,
) (int, error) {
	return storage.NutzereinstellungenZuruecksetzen(ctx, r.holen(), nutzerID)
}

// RuecksetzungUeber baut die Rücksetzung.
//
// **Der Weg kommt als Funktion**, weil er bei jedem Ruf neu entsteht: ein
// festgehaltener trüge nach einem Zurückspielen die alte Verbindung — dieselbe
// Entscheidung wie bei `ruecksprungnaht`.
func RuecksetzungUeber(
	holen func() *sql.DB, weg func() *service.Bestandsweg,
	loescher service.Ruecksetzungsloescher, wache service.Ruecksetzungswache,
) *service.Ruecksetzung {
	return ruecksetzungMit(holen, loescher, wache,
		func(ctx context.Context) (string, error) { return weg().Neuanfang(ctx) })
}

// ruecksetzungMit baut die Rücksetzung mit einem eigenen Neuanfang — das
// Fenster legt ihn in den Rahmen seines Tors ([App.imTausch]).
func ruecksetzungMit(
	holen func() *sql.DB, loescher service.Ruecksetzungsloescher,
	wache service.Ruecksetzungswache, neuanfang func(context.Context) (string, error),
) *service.Ruecksetzung {
	return service.NeueRuecksetzung(ruecksetzungsablage{holen: holen}, loescher, wache, neuanfang)
}

// ruecksetzungVerdrahten hängt die Rücksetzung ins Fenster.
//
// **Über die laufende Verbindung und den Weg als Funktion**: der Neuanfang
// tauscht den Bestand aus, während die Naht steht — wer beides festhielte,
// läse danach eine Datei, die es nicht mehr gibt.
//
// **Mit dem Auftragsbuch**, das dieses Fenster ohnehin führt (U11): was gerade
// läuft, wird nicht weggelöscht. Ein Befehl hat keines — er tut, was er tut,
// und daneben läuft nichts.
//
// **Und der Neuanfang läuft im Rahmen des Tors** (A02): er tauscht den
// Bestand aus wie das Zurückspielen.
func (a *App) ruecksetzungVerdrahten() {
	a.MitRuecksetzung(ruecksetzungMit(
		a.laufendeVerbindung, a.nutzer, a.autorisierung, a.neuanfangImTausch,
	).MitAuftraegen(a.laufendeAuftraege))
}

// neuanfangImTausch legt den Bestand beiseite und öffnet einen frischen — im
// Rahmen des Tors.
func (a *App) neuanfangImTausch(ctx context.Context) (string, error) {
	var beiseite string

	fehler := a.imTausch(func() error {
		var tauschfehler error
		beiseite, tauschfehler = a.bestandsweg().Neuanfang(ctx)

		return tauschfehler
	})

	return beiseite, fehler
}

// laufendeAuftraege gibt, was dieses Fenster gerade tut.
//
// **Ein Ruf und kein festgehaltenes Buch**: `verdrahten` läuft ein zweites
// Mal, wenn jemand als Leser in ein fremdes Profil wechselt, und legt dabei
// ein neues Buch an. Wer das alte hielte, sähe eine Anwendung, die nichts tut.
func (a *App) laufendeAuftraege() []domain.Auftrag { return a.auftragsbuch().Auftraege() }
