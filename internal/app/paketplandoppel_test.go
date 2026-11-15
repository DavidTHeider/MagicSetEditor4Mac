package app_test

// **Plan, Ersetzen und Zurück am Doppel** — DOM-01.AK-02 (#1696). Eine
// eigene Datei, damit `paketkatalog_test.go` unter der Zeilengrenze bleibt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func (p *paketfuehrer) PlanEntfernen(_ context.Context, art, name string) (service.Paketplan, error) {
	p.geplant = "entfernen " + art + " " + name

	return p.plan, p.fehler
}

func (p *paketfuehrer) PlanEinbinden(_ context.Context, quelle string) (service.Paketplan, error) {
	p.geplant = "einbinden " + quelle

	return p.plan, p.fehler
}

func (p *paketfuehrer) PlanZurueck(_ context.Context, art, name string) (service.Paketplan, error) {
	p.geplant = "zurueck " + art + " " + name

	return p.plan, p.fehler
}

func (p *paketfuehrer) Ersetzen(_ context.Context, quelle string, bestaetigt bool) (domain.Paket, error) {
	p.installiert = quelle
	if bestaetigt {
		p.installiert += " (bestätigt)"
	}

	return domain.Paket{Art: "mse-style", Name: "ersetzt"}, p.fehler
}

func (p *paketfuehrer) Zurueck(_ context.Context, art, name string) (domain.Paket, error) {
	p.entfernt = "zurück " + art + " " + name

	return domain.Paket{Art: art, Name: name}, p.fehler
}

func (p *paketfuehrer) Setbedarf(_ context.Context, setID int64) ([]service.Abhaengigkeitsbefund, error) {
	p.gefragtesSet = setID

	return p.bedarf, p.fehler
}
