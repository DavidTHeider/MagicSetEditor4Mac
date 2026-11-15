package app

// **Auch der Zug eines Gastes bewegt den Datenstand** — Befund 1 des
// Nachlaufs zum Gesamtreview über 0.6.0 (#1660).
//
// Der Horchposten nimmt drei Zwecke an: Handel, Paarung und Partie. Die
// ersten beiden sagten seit #2060 `nebenherGeschrieben`, die Partie nicht —
// und sie schreibt am häufigsten: jeder Zug eines Gastes geht im Prozess des
// Wirts in den Bestand. Das Tischbild kam an, die Partienliste des Wirts hing
// am alten Stand.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinGastzugBewegtDenDatenstandDesWirts(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)
	wirt.anwendung.MitStand(festerStand(7))

	partieID, objekt := einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	if antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{},
	); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	wirt.altesVerwerfen()

	// **Erst nach dem Beitritt gemessen**: ob schon das Platznehmen den Stand
	// bewegt, ist nicht die Frage.
	vorher := wirt.anwendung.Datenstand(wirt.ctx).Stand

	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	}); antwort.Fehler != nil {
		t.Fatalf("PartieWuenschen: %+v", antwort.Fehler)
	}

	// **Auf die Wirkung gewartet** (§3): die Meldung an das Fenster des
	// Wirts kommt erst, nachdem der Zug geschrieben ist.
	sammleBis(t, wirt.meldungen, EreignisEigenerZug)

	if nachher := wirt.anwendung.Datenstand(wirt.ctx).Stand; nachher == vorher {
		t.Errorf("der Gast hat gezogen, und der Datenstand des Wirts blieb %d", nachher)
	}
}

// ablehnenderNehmer weist jeden Zug ab.
type ablehnenderNehmer struct{ service.Befehlsnehmer }

func (ablehnenderNehmer) BefehlVerkuendet(
	context.Context, domain.Spielbefehl, service.Verkuendung,
) (service.Befehlsausgang, error) {
	return service.Befehlsausgang{}, errors.New("der Zug ist nicht erlaubt")
}

// **Ein abgewiesener Zug schrieb nichts** — und das Fenster des Wirts lüde
// sonst nach jedem Versuch eines Gastes nach. Dass der Nehmer zuhört, belegt
// der Lauf darüber.
func TestEinAbgewiesenerGastzugLaesstDenDatenstand(t *testing.T) {
	t.Parallel()

	gesagt := 0
	nehmer := nebenherNehmer{Befehlsnehmer: ablehnenderNehmer{}, geschrieben: func() { gesagt++ }}

	if _, fehler := nehmer.BefehlVerkuendet(t.Context(), domain.Spielbefehl{}, nil); fehler == nil {
		t.Fatal("der Zug kam durch, obwohl der Dienst ihn abwies")
	}

	if gesagt != 0 {
		t.Errorf("ein abgewiesener Zug sagte %d-mal, es sei geschrieben worden", gesagt)
	}
}
