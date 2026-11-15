package service

// **Die Wege nach draußen zu einer Karte** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Warum es dafür einen Dienst gibt und nicht nur die Domäne
//
// **Weil die Kennungen im Bestand liegen.** Welche Anbieterkennungen ein
// Druck führt, weiß `A2` (#1441);
// die Domäne baut daraus die Adressen, und dieser Dienst holt sie.
//
// # Und warum ein Druck, den es nicht gibt, trotzdem Links bekommt
//
// **Weil eine Suche nach dem Namen keine Kennung braucht.** Wer eine eigene
// Karte ansieht, hat keinen kanonischen Druck — und bekommt deshalb keine
// **geratenen** Produktlinks, sondern eine Suche, wenn er einen Namen hat,
// und sonst die Startseite. Das Issue sagt es: „eigene Karten ohne
// kanonischen Druck bekommen keine geratenen Marktlinks."

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Druckkennungen gibt die Kennungen eines Drucks.
type Druckkennungen interface {
	NachScryfall(ctx context.Context, printing string) (*domain.Druck, error)
}

// Linkwerk baut die Wege nach draußen.
type Linkwerk struct{ drucke Druckkennungen }

// NeuesLinkwerk baut es über einer Druckquelle.
func NeuesLinkwerk(drucke Druckkennungen) *Linkwerk {
	return &Linkwerk{drucke: drucke}
}

// Linksfuer gibt je Anbieter den genauesten Weg, der zu haben ist.
//
// **`printing` darf leer sein** — dann gibt es keine Kennung und keinen
// kanonischen Druck, und es bleibt beim Namen. Wer beides nicht hat, bekommt
// Startseiten; das ist eine Auskunft und kein Fehler.
func (l *Linkwerk) Linksfuer(
	ctx context.Context, printing, name string,
) ([]domain.Marktlink, error) {
	var kennungen []domain.Druckkennung

	if printing != "" {
		druck, fehler := l.drucke.NachScryfall(ctx, printing)
		if fehler != nil {
			return nil, fehler
		}

		if druck != nil {
			kennungen = druck.Kennungen

			if name == "" {
				name = druck.Name
			}
		}
	}

	// **Was hier herauskommt, hält die Prüfung** — und der Beleg dafür steht
	// nicht hier.
	//
	// **Ein Lauf über die gebauten Adressen wäre ein Zweig, den kein Fall
	// betritt** (§10): `domain.Marktlinks` baut sie, und dass jede von ihnen
	// `PruefeMarktlink` hält, misst `TestJederAnbieterBekommtGenauEinenWeg` —
	// für **jeden** Anbieter und bei jedem Lauf. Ein zweiter Wächter an
	// dieser Stelle prüfte dasselbe und wäre nie rot.
	//
	// **Geprüft wird dafür beim Öffnen** ([Linkwerk.Oeffne]), denn dorthin
	// kommt eine Adresse aus fremder Hand.
	return domain.Marktlinks(kennungen, name), nil
}

// Oeffner macht eine Adresse im Standardbrowser auf.
//
// **Eingesetzt und nicht erzeugt** (§16): ein Dienst, der `exec` selbst ruft,
// lässt sich nicht prüfen, ohne einen Browser zu öffnen.
type Oeffner func(ctx context.Context, adresse string) error

// Oeffne prüft eine Adresse und übergibt sie dem Plattformöffner.
//
// # Warum hier ein zweites Mal geprüft wird
//
// **Weil die Adresse dazwischen durch fremde Hände geht.** Sie stand in einer
// Antwort, lag im Fenster, kam über REST oder von einem Gegenüber zurück —
// und was von dort kommt, ist nicht das, was hier gebaut wurde, sondern das,
// was jemand schickt. Das Issue sagt es: „Peerpayload darf niemals selbst
// öffnen."
//
// **Und der Anbieter kommt mit.** Ohne ihn wüsste diese Stelle nicht, gegen
// welchen Wirt zu prüfen ist — und „irgendeiner der sechs" wäre eine
// Positivliste, die keine ist.
func (l *Linkwerk) Oeffne(
	ctx context.Context, oeffnen Oeffner, anbieter, adresse string,
) error {
	link := domain.Marktlink{
		Anbieter: domain.Linkanbieter(anbieter), Adresse: adresse,
	}

	if fehler := domain.PruefeMarktlink(link); fehler != nil {
		return fehler
	}

	if oeffnen == nil {
		return domain.MitGrund(domain.ErrValidation, nil,
			"diese Fläche kann keinen Browser öffnen")
	}

	return oeffnen(ctx, adresse)
}
