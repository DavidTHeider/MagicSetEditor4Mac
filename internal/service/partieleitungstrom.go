package service

// **Wie ein Befehl in den Strom kommt** — die Kette und das Schreiben.
// Herausgelöst aus `partieleitung.go`, als die Rücknahme (P8,
// #2376)
// einen zweiten Weg zum Schreiben brachte und die Datei über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort, **ob** ein
// Befehl gilt; hier, **wie** er festgehalten wird.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schreiben kettet die Ereignisse und schreibt sie mit dem neuen Stand.
//
// **Und hier werden die Sichtkennungen nachgeführt** (#2480, Vermerk 0201):
// an einer Stelle nach jedem Befehl, auch nach einer Rücknahme — ein
// Schreiber einer Zone kann es nicht vergessen.
func (l *Partieleitung) schreiben(
	ctx context.Context, partie domain.Partie, kennung string,
	danach domain.Partie, ereignisse []domain.Ereignis,
) (Befehlsausgang, error) {
	// **Eine Partie aus der Zeit davor bekommt ihr Geheimnis jetzt** — mit
	// demselben Schreiben, das diesen Befehl festhält.
	if danach.Geheimnis == "" {
		danach.Geheimnis = l.geheimnis()
	}

	partie.Geheimnis = danach.Geheimnis
	danach = domain.Sichtwechsel(partie, danach)
	ereignisse = domain.Benennen(partie, danach, ereignisse)

	if fehler := l.ketten(ctx, partie.ID, ereignisse); fehler != nil {
		return Befehlsausgang{}, fehler
	}

	danach.Fassung = partie.Fassung + 1

	if fehler := l.partien.Anwenden(
		ctx, partie.Fassung, kennung, danach, ereignisse); fehler != nil {
		return Befehlsausgang{}, fehler
	}

	return Befehlsausgang{Fassung: danach.Fassung, Ereignisse: ereignisse}, nil
}

// ketten hängt die Ereignisse an den Strom.
//
// **Hier und nicht im Ausleger**: die Folge ist eine Eigenschaft des Stroms
// und nicht der Regel. Ein Ausleger, der sie setzte, müsste den Strom kennen
// — und wäre damit keine reine Auslegung mehr.
func (l *Partieleitung) ketten(
	ctx context.Context, partieID string, ereignisse []domain.Ereignis,
) error {
	vorher, folge, fehler := l.partien.LetzteKette(ctx, partieID)
	if fehler != nil {
		return fehler
	}

	for i := range ereignisse {
		folge++

		ereignisse[i].PartieID = partieID
		ereignisse[i].Folge = folge

		abdruck, fehler := ereignisse[i].Abdruck(vorher)
		if fehler != nil {
			return fehler
		}

		ereignisse[i].Kette = abdruck
		vorher = abdruck
	}

	return nil
}
