package app

// **Wo der Bilderkenner herkommt** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum die Naht erst beim ersten Ruf baut
//
// **Weil [Scryfallklient] scheitern kann und [App.verdrahten] keine Rückgabe
// hat.** Er braucht das Benutzerverzeichnis für seinen Zwischenspeicher; wer
// ihn beim Start bauen wollte, hätte die Wahl zwischen einem verschluckten
// Fehler und einem Start, den eine Bilderkennung aufhält, die niemand
// verlangt hat.
//
// **Und weil ein Schalter sich ändert.** Derselbe Grund steht bei
// [Scryfallklient]: die Sperre gegen einen abgeschalteten Dienst wird **vor
// jeder Anfrage** gefragt, nicht beim Bauen. Eine Naht, die den Klienten
// festhielte, hielte auch seine Erlaubnisfrage fest.
//
// # Und warum sie trotzdem beim Start gesetzt wird
//
// **Weil `nil` hier »es ist kein Bestand offen« hieße** — an jeder Fläche,
// bei offenem Bestand. Der Wächter über die Nähte
// (`naehte_intern_test.go`) hat genau das angestrichen: gebaut und nicht
// angeschlossen ist dieselbe Gestalt wie
// #1539.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bilderkennungsnaht führt die gelesenen Stellen an dieselbe Drucksuche, die
// eine getippte Kartenliste benutzt.
type bilderkennungsnaht struct{ holen func() *sql.DB }

// Deuten baut den Klienten und gibt die Frage weiter.
func (b bilderkennungsnaht) Deuten(
	ctx context.Context, erkennung domain.Bilderkennung,
) (service.Bilddeutung, error) {
	db := b.holen()
	if db == nil {
		return service.Bilddeutung{}, fehlenderBestand()
	}

	// **Derselbe Klient wie überall, über die eine Stelle**: er bringt die
	// Drosseln mit, und zwei unter derselben Grenze halten sie nicht ein
	// (§16, Regel 4).
	klient, fehler := Scryfallklient(dienstwahlUeber(db))
	if fehler != nil {
		return service.Bilddeutung{}, fehler
	}

	return service.NeuerBilderkenner(klient).Deuten(ctx, erkennung)
}

// bilderkennungVerdrahten hängt sie an die Anwendung.
func (a *App) bilderkennungVerdrahten() {
	a.MitBilderkenner(bilderkennungsnaht{holen: a.laufendeVerbindung})
}
