package service

// **Der Plan vor einer Paketänderung** — DOM-01.AK-02
// (#1696):
// „Installation, Aktualisierung, Deinstallation und Rückkehr zur vorigen
// kompatiblen Version besitzen eine Vorschau betroffener Sets/Stile und
// erkennen Abhängigkeitskonflikte."
//
// # Wie gerechnet wird
//
// **Gegen den Katalog, wie er danach aussähe.** Ein Entfernen nimmt das Paket
// heraus, ein Einbinden legt es hinein, ein Ersetzen und eine Rückkehr
// tauschen die Fassung — und dann wird geprüft, was die bestehende Prüfung
// ohnehin prüft ([pruefeBedarf]): was das neue Paket braucht, und was die
// Pakete brauchen, die es brauchen. Ein Konflikt ist deshalb derselbe Befund
// wie in [Paketverwalter.Abhaengigkeiten], mit denselben Wörtern im Fenster.
//
// # Wo das von MSE2 abweicht
//
// Portierungsmarke **adapted**. MSE2 setzt eine Aktion und zieht die Kaskade
// **selbst** nach: ein Einbinden markiert fehlende Abhängigkeiten zum
// Einbinden, ein Entfernen markiert die Abhängigen zum Entfernen
// (`set_package_action`, `add_package_dependency`,
// `remove_package_dependency`, `src/data/installer.cpp`, Fassung f9d9356 vom
// 2021-07-06). **Hier nennt der Plan sie, und ausgeführt wird nur, was
// gewählt ist**: es gibt keinen Installer, aus dem eine fehlende Abhängigkeit
// käme, und ein Entfernen, das still weitere Pakete mitnimmt, nähme mehr, als
// jemand gewählt hat.

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Paketvorgang ist, was geplant wird.
type Paketvorgang string

const (
	PaketvorgangEntfernen Paketvorgang = "entfernen"
	PaketvorgangEinbinden Paketvorgang = "einbinden"
	PaketvorgangErsetzen  Paketvorgang = "ersetzen"
	PaketvorgangZurueck   Paketvorgang = "zurueck"
)

// Paketplan ist, was ein Vorgang täte — bevor er es tut.
type Paketplan struct {
	Vorgang Paketvorgang

	// Paket ist der Dateiname — `magic.mse-game`.
	Paket string

	// Von und Zu sind die Fassungen vorher und nachher; leer, wo keine ist.
	Von, Zu string

	// Geaendert sagt, dass das vorhandene Paket vom Stand beim Einbinden
	// abweicht — ein Ersetzen hebt die Änderung als vorige Fassung auf.
	Geaendert bool

	// Moeglich ist falsch, wenn der Vorgang gar nicht geht: ein Paket
	// außerhalb des verwalteten Verzeichnisses, ein Zurück ohne vorige
	// Fassung.
	Moeglich bool

	// Sets sind die Sets dieses Profils, die das Paket nennen.
	Sets []string

	// Konflikte sind die Abhängigkeiten, die danach nicht mehr stimmten.
	Konflikte []Abhaengigkeitsbefund
}

// PlanEntfernen plant das Entfernen eines Pakets.
func (v *Paketverwalter) PlanEntfernen(ctx context.Context, art, name string) (Paketplan, error) {
	pakete, alt, err := v.vorhandenes(ctx, art, name)
	if err != nil {
		return Paketplan{}, err
	}

	plan, err := v.planen(ctx, PaketvorgangEntfernen, pakete, &alt, nil)
	plan.Moeglich = imVerzeichnis(alt.Pfad, v.verwaltet)

	return plan, err
}

// PlanEinbinden plant das Einbinden eines Pakets von der Platte — ein
// **Ersetzen**, wenn es ein Paket dieses Namens schon gibt.
func (v *Paketverwalter) PlanEinbinden(ctx context.Context, quelle string) (Paketplan, error) {
	neu, err := v.paketLesen(quelle)
	if err != nil {
		return Paketplan{}, err
	}

	pakete, err := v.katalog.Pakete(ctx)
	if err != nil {
		return Paketplan{}, err
	}

	for _, paket := range mitAenderung(pakete) {
		if paket.Art == neu.Art && paket.Name == neu.Name {
			plan, fehler := v.planen(ctx, PaketvorgangErsetzen, pakete, &paket, neu)
			plan.Moeglich = imVerzeichnis(paket.Pfad, v.verwaltet)

			return plan, fehler
		}
	}

	plan, err := v.planen(ctx, PaketvorgangEinbinden, pakete, nil, neu)
	plan.Moeglich = true

	return plan, err
}

// PlanZurueck plant die Rückkehr zur vorigen Fassung.
func (v *Paketverwalter) PlanZurueck(ctx context.Context, art, name string) (Paketplan, error) {
	pakete, alt, err := v.vorhandenes(ctx, art, name)
	if err != nil {
		return Paketplan{}, err
	}

	vorige, befund := lies(v.vorigeOrt(alt), alt.Dateiname(), nil)
	if befund != nil {
		plan, fehler := v.planen(ctx, PaketvorgangZurueck, pakete, &alt, &alt)
		plan.Zu, plan.Konflikte, plan.Moeglich = "", []Abhaengigkeitsbefund{}, false

		return plan, fehler
	}

	plan, err := v.planen(ctx, PaketvorgangZurueck, pakete, &alt, vorige)
	plan.Moeglich = imVerzeichnis(alt.Pfad, v.verwaltet)

	return plan, err
}

// vorhandenes liest den Katalog samt lokaler Änderung und sucht ein Paket
// darin.
func (v *Paketverwalter) vorhandenes(ctx context.Context, art, name string) (
	pakete []domain.Paket, gefunden domain.Paket, err error,
) {
	pakete, err = v.Liste(ctx)
	if err != nil {
		return nil, domain.Paket{}, err
	}

	for _, paket := range pakete {
		if paket.Art == art && paket.Name == name {
			return pakete, paket, nil
		}
	}

	return nil, domain.Paket{}, fmt.Errorf("%w: das Paket %q der Art %q ist nicht installiert",
		domain.ErrNotFound, name, art)
}

// planen rechnet den Plan gegen den Katalog danach: `alt` geht, `neu` kommt.
func (v *Paketverwalter) planen(
	ctx context.Context, vorgang Paketvorgang, pakete []domain.Paket, alt, neu *domain.Paket,
) (Paketplan, error) {
	ziel := neu
	if ziel == nil {
		ziel = alt
	}

	plan := Paketplan{Vorgang: vorgang, Paket: ziel.Dateiname(), Sets: []string{}, Konflikte: []Abhaengigkeitsbefund{}}
	if alt != nil {
		plan.Von, plan.Geaendert = alt.Version, alt.Geaendert
	}

	danach := make(map[string]domain.Paket, len(pakete))
	for _, paket := range pakete {
		if paket.Dateiname() != plan.Paket {
			danach[paket.Dateiname()] = paket
		}
	}

	if neu != nil {
		plan.Zu = neu.Version
		danach[plan.Paket] = *neu

		for _, bedarf := range neu.Braucht {
			if befund := pruefeBedarf(plan.Paket, bedarf, danach); befund != nil {
				plan.Konflikte = append(plan.Konflikte, *befund)
			}
		}
	}

	for _, paket := range pakete {
		for _, bedarf := range paket.Braucht {
			if paket.Dateiname() == plan.Paket || bedarf.Paket != plan.Paket {
				continue
			}

			if befund := pruefeBedarf(paket.Dateiname(), bedarf, danach); befund != nil {
				plan.Konflikte = append(plan.Konflikte, *befund)
			}
		}
	}

	nutzung, err := v.nutzung(ctx)
	if err != nil {
		return Paketplan{}, err
	}

	plan.Sets = append(plan.Sets, nutzung[plan.Paket]...)
	slices.Sort(plan.Sets)

	return plan, nil
}

// vorigeOrt ist, wo die vorige Fassung eines Pakets liegt — **neben** dem
// verwalteten Verzeichnis und nicht darin: dort fände der Katalog sie als
// zweites Paket desselben Namens.
func (v *Paketverwalter) vorigeOrt(paket domain.Paket) string {
	return filepath.Join(v.verwaltet+"-vorige", paket.Dateiname())
}
