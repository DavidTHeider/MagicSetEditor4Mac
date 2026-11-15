package app

// **Die offenen Peersitzungen der Anwendung** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum es hier schon eine Verwaltung gibt, obwohl noch nichts verbindet
//
// **Weil der Profilwechsel sie jetzt braucht.** Das Issue verlangt: „Bei
// Nutzer-/Profilwechsel Sessions pausieren und Schlüsselkontext schließen,
// verspätete Antworten verwerfen." Diese Zusage entsteht nicht mit dem
// Transport, sie entsteht mit dem **Wechsel** — und der ist gebaut.
//
// **Den Wähler gibt es seit N2**: [adresswaehler] wählt die von Hand
// eingetragene Adresse eines Kontakts. libp2p kommt nicht
// ([0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md)), und N3 läuft
// über tailcat
// ([0112](../../docs/entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md)).

import (
	"context"
	"crypto/tls"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Sitzungstiefe ist, wie viele Nachrichten je Sitzung warten dürfen.
//
// **Acht, und die Zahl ist eine Entscheidung**: genug, damit ein Schub
// Präsenznachrichten nicht sofort staut, und wenig genug, dass ein Gegenüber,
// das nicht liest, keinen Speicher bindet. Wer sie ändert, ändert, wann der
// Rückstau greift.
const Sitzungstiefe = 8

// peerfelder ist, was eine Anwendung für Peerverbindungen hält.
//
// **Eine eigene Struktur**, weil `anwendung.go` mit den Feldern über die
// Zeilengrenze aus §2 ginge — und weil sie zusammengehören.
type peerfelder struct {
	// waehler baut Peerverbindungen — nil heißt: der eingebaute
	// [adresswaehler]. Ein eingesetzter ist die Vorrichtung eines Laufs.
	waehler service.Waehler

	// serverzeugnis baut das Zeugnis eines Dienstes auf der Peeridentität —
	// nil heißt: das echte aus `internal/leitung` (NET-03, #1707).
	serverzeugnis func(peer.Identitaet) (*tls.Config, error)

	// sitzungen hält die offenen Sitzungen dieses Profils.
	//
	// **Eine je Anwendung und nicht je Ruf**: sie **hält** die Sitzungen,
	// und eine, die bei jedem Ruf neu entstünde, hätte keine.
	sitzungen *service.Peersitzungen

	// einloesegrenze zählt die Einlöseversuche (NET-01b) — **eine je
	// Anwendung** aus demselben Grund wie die Sitzungen: sie zählt über die
	// Rufe hinweg.
	einloesegrenze *service.Einloesegrenze

	// horchen ist der offene Port — was daran hängt, steht in
	// `peerhorchen.go`.
	//
	// **Hier und nicht an [App]**, weil es dasselbe beantwortet wie die
	// zwei darüber: was diese Anwendung für Peerverbindungen hält.
	horchen horchfelder
}

// waehlernaht fragt die Anwendung bei jedem Wählen nach ihrem Transport.
//
// **Deshalb darf [App.MitWaehler] noch nach dem Bauen der Verwaltung
// kommen** — sie hält keinen Wähler fest, sondern die Anwendung.
type waehlernaht struct{ app *App }

func (w waehlernaht) Waehle(
	ctx context.Context, kennung peer.Kennung, protokoll string,
) (service.Verbindung, error) {
	return w.app.waehlerOderVorgabe().Waehle(ctx, kennung, protokoll)
}

// ausweisnaht reicht die Ausweisfrage an das Peerwerk.
//
// **Je Ruf ein Werk**, damit Bestandsverbindung und Schlüsselbund die des
// Augenblicks sind — derselbe Grund wie bei [kontaktrechte] darunter.
type ausweisnaht struct{ app *App }

func (a ausweisnaht) Ausweis(ctx context.Context) (peer.Identitaet, error) {
	return a.app.peerwerk().Ausweis(ctx)
}

// kontaktrechte reicht die Rechtefrage an das Kontaktwerk.
//
// **Je Ruf ein Werk**, damit die Bestandsverbindung die des Augenblicks ist —
// derselbe Grund wie bei [App.peerwerk].
type kontaktrechte struct{ app *App }

func (k kontaktrechte) Darf(
	ctx context.Context, kennung peer.Kennung, zweck domain.Zweck,
) (bool, error) {
	return k.app.kontaktwerk().Darf(ctx, kennung, zweck)
}

// sitzungsverwaltung gibt die eine Verwaltung dieser Anwendung.
//
// **Sie steht seit [NeueApp]**, damit der Wächter über die Nähte sie sieht —
// eine Naht, die erst beim ersten Ruf entsteht, ist bis dahin `nil`, und
// „gesetzt" heißt dann „noch nicht abgestürzt".
func (a *App) sitzungsverwaltung() *service.Peersitzungen {
	return a.peerlage.sitzungen
}

// waehlerOderVorgabe gibt den eingesetzten Wähler — oder den echten.
//
// [adresswaehler] schlägt die von Hand eingetragene Adresse des Kontakts nach
// und baut eine gesicherte Leitung dorthin.
//
// **Wer keine eingetragen hat, bekommt das gesagt** — und zwar von
// [service.Kontaktwerk.Adresse], also mit dem Namen der Sache statt mit einem
// Satz über den Bauzustand dieses Programms.
func (a *App) waehlerOderVorgabe() service.Waehler {
	if a.peerlage.waehler != nil {
		return a.peerlage.waehler
	}

	return adresswaehler{app: a}
}

// MitWaehler setzt den Transport — für Läufe und für N2 und N3.
func (a *App) MitWaehler(waehler service.Waehler) *App {
	a.peerlage.waehler = waehler

	return a
}

// sitzungenPausieren schließt jede offene Sitzung dieses Profils.
//
// **Beim Wechsel und nicht beim Schließen des Fensters.** Eine Sitzung gehört
// dem Profil, das sie geöffnet hat; wer sie über den Wechsel hinweg offen
// ließe, trüge Daten zwischen zwei Menschen, die an einem Rechner arbeiten.
func (a *App) sitzungenPausieren() {
	if fehler := a.sitzungsverwaltung().Pausiere(); fehler != nil {
		// **Gemeldet und nicht gestoppt.** Der Wechsel findet statt; eine
		// Leitung, die sich beim Niederlegen beschwert, ist kein Grund,
		// das Profil offen zu lassen.
		slog.Default().Error("eine Peersitzung ließ sich nicht schließen",
			"grund", fehler)
	}
}
