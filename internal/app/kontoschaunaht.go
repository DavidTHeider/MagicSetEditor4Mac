package app

// **Die Naht der Kontoschau** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// **Weil nur hier `internal/storage` stehen darf** (§2) — dieselbe Gestalt
// wie bei [cardtradermarktUeber] nebenan, und aus demselben Grund: derselbe
// Klient, derselbe Schalter, dieselbe Drossel. Ein zweiter Klient daneben
// hätte eine zweite Drossel unter derselben Grenze (§16, Regel 4).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kontostand ist, was die Schau zurückgibt — der Name der Naht, damit
// `kontoschau.go` ohne `internal/service` auskommt.
type kontostand = service.Kontostand

// kontoschau baut die Schau für diesen Ruf.
//
// **Der Marktplatz wird zuerst geprüft**: ein Name, den es nicht gibt, soll
// nicht erst den Schlüsselbund öffnen.
func (a *App) kontoschau(ctx context.Context, quelle string) (
	*service.Kontoschau, error,
) {
	if quelle != string(cardtrader.Familie) {
		return nil, keinMarktplatz(quelle, string(cardtrader.Familie))
	}

	verbindung := a.laufendeVerbindung()
	if verbindung == nil {
		return nil, fehlenderBestand()
	}

	einstellungen := storage.NeueEinstellungen(verbindung)
	riegel := AutorisierungAus(verbindung)

	markt, fehler := cardtraderkontoUeber(ctx,
		NeueEinstellungswahl(einstellungen).MitRiegel(riegel),
		a.bundFuerDenMarkt())
	if fehler != nil {
		return nil, fehler
	}

	return service.NeueKontoschau(markt,
		NeueEinstellungswahl(einstellungen), riegel), nil
}

// cardtraderkontoUeber baut den Klienten für die drei Kontowege.
//
// **Dieselbe Marke wie am Marktplatz**: ein zweiter Schlüsselbundeintrag für
// dasselbe Geheimnis wäre eine zweite Stelle, an der es veralten kann.
func cardtraderkontoUeber(ctx context.Context, wahl Dienstwahl, bund Geheimnisleser) (
	*cardtrader.Klient, error,
) {
	marke, fehler := geheimnis(ctx, bund,
		domain.ZugangKonto(domain.DienstCardTrader, domain.ZugangMarke))
	if fehler != nil {
		return nil, fehler
	}

	if marke == "" {
		return nil, domain.ZugangFehlt(
			domain.DienstCardTrader, domain.ZugangMarke)
	}

	adresse := cardtraderadresse()

	// **Mit langer Frist** — der volle Bestandsexport dauert beim Anbieter
	// Sekunden bis Minuten, und er empfiehlt selbst 120 bis 180 Sekunden
	// (Befund 6 des Zweitblicks über Phase 5.6).
	return cardtrader.NeuerKlient(adresse,
		dienstklientMit(netz.LangeFristen(), wahl, domain.DienstCardTrader),
		cardtrader.Zugang{Marke: marke}).
		MitDrossel(drosselwerk.Fuer(adresse, cardtrader.Abstand)), nil
}
