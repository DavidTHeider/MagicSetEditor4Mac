package app

// **Die Mischpflicht über die Naht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum sie einen eigenen Lauf bekommt
//
// **Weil sie die öffentliche Hälfte eines Geheimnisses ist.** Was ein Blick
// zerstört, macht erst das Mischen wieder gut; dass jemand gesehen und noch
// nicht gemischt hat, muss **jeden** erreichen — auch den, der gar nicht
// hingesehen hat. Fiele sie an dieser Naht weg, sähe die Antwort
// vollständiger aus, nicht falscher (§3, die Naht zwischen den Prüfebenen).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mischpflichtige Leitung gibt eine Sicht mit einer offenen Mischpflicht.
// mischpflichtigeLeitung antwortet wie die schweigsame — bis auf die Sicht.
//
// **Eingebettet und nicht abgeschrieben**: was diese Naht kann, wächst mit
// jedem Werkstück, und ein Doppel, das jede Erweiterung nachträgt, ist ein
// zweiter Ort für dieselbe Aufzählung (§10).
type mischpflichtigeLeitung struct{ schweigsameLeitung }

func (mischpflichtigeLeitung) Anlegen(
	_ context.Context, partie domain.Partie,
) (domain.Partie, error) {
	return partie, nil
}

func (mischpflichtigeLeitung) Partie(
	_ context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	return domain.Partiesicht{
		PartieID: kennung, Fuer: fuer,
		Mischpflichten: []domain.Mischpflicht{
			{Sitz: "sitz-2", Zone: domain.ZoneBibliothek},
		},

		// **Und alles andere, was für jeden gilt** — die offene
		// Legendenwahl (CR 704.5j), die wartenden Fähigkeiten samt der
		// Frage nach ihrer Reihenfolge (CR 603.3b) und der Kampfstand
		// (CR 508.1a, CR 509.1a). **Zusammen in einem Doppel**, weil sie
		// dieselbe Zusage tragen: was am Tisch öffentlich ist, kommt
		// ungekürzt durch die Naht. Der Lauf dazu steht in
		// `partieoffenefragen_intern_test.go`.
		Legendenwahlen: []domain.Legendenwahl{
			{Sitz: "sitz-2", Name: "Held"},
		},
		Ausgeloest: []domain.Ausloeser{{
			Art: domain.AusloeserMonarchzug, Beherrscher: "sitz-2",
			Wen: "sitz-1", Kennung: "sitz-2/monarch.zug/0",
		}},
		Ausloeserreihung: "sitz-2",
		Zug: domain.Zugstand{
			Nummer: 3, Aktiv: "sitz-1",
			Kampf: domain.Kampfstand{
				Angriffe: []domain.Angriff{
					{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
				},
				Bloecke: []domain.Block{
					{Kreatur: "sitz-2/wand/0", Angreifer: "sitz-1/baer/0"},
				},
			},
		},
	}, nil
}

func (m mischpflichtigeLeitung) Tischbild(
	ctx context.Context, kennung, fuer string,
) (domain.Tischbild, error) {
	sicht, fehler := m.Partie(ctx, kennung, fuer)

	return domain.Tischbild{
		Sicht: sicht, Erlaubt: []string{}, Handlungen: []domain.Objekthandlungen{},
	}, fehler
}

func (mischpflichtigeLeitung) BefehlVerkuendet(
	context.Context, domain.Spielbefehl, service.Verkuendung,
) (service.Befehlsausgang, error) {
	return service.Befehlsausgang{}, nil
}

func (mischpflichtigeLeitung) InReihe(tun func() error) error { return tun() }

func (mischpflichtigeLeitung) Verlauf(
	context.Context, string, string,
) ([]domain.Ereignis, error) {
	return nil, nil
}

func (mischpflichtigeLeitung) Wirtsstand(
	context.Context, string, string,
) (domain.Wirtsstand, error) {
	return domain.Wirtsstand{Epoche: domain.Partieepoche, AelteteFolge: 1}, nil
}

func (mischpflichtigeLeitung) Nachreichen(
	context.Context, string, string, int64,
) ([]domain.Ereignis, error) {
	return nil, nil
}

func (mischpflichtigeLeitung) Platznehmen(
	context.Context, string, string, string, domain.Partiedeck, service.Verkuendung,
) (service.Befehlsausgang, error) {
	return service.Befehlsausgang{}, nil
}

func (mischpflichtigeLeitung) Pausieren(context.Context, string) error {
	return nil
}

func (mischpflichtigeLeitung) Fortsetzen(context.Context, string) error {
	return nil
}

func (mischpflichtigeLeitung) Beenden(context.Context, string) error {
	return nil
}

// **Die Mischpflicht kommt als DTO an — bei jedem.**
func TestDieMischpflichtKommtAlsDTOAn(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartienaht(&mischpflichtigeLeitung{})

	// **Gefragt wird aus der Sicht eines *anderen* Sitzes** — gerade der
	// soll sie sehen.
	antwort := NeueBindungen(anwendung).Partie(t.Context(), "p1", "sitz-1")
	if antwort.Fehler != nil {
		t.Fatalf("Partie: %+v", antwort.Fehler)
	}

	if len(antwort.Partie.Mischpflichten) != 1 {
		t.Fatalf("%d Mischpflichten: %+v",
			len(antwort.Partie.Mischpflichten), antwort.Partie.Mischpflichten)
	}

	pflicht := antwort.Partie.Mischpflichten[0]
	if pflicht.Sitz != "sitz-2" || pflicht.Zone != "bibliothek" {
		t.Errorf("%+v", pflicht)
	}
}

// **Und ohne eine steht eine leere Liste da, kein `nil`.**
//
// **Derselbe Satz wie bei den Sitzen und den Objekten** — ein `null` im
// Fenster lässt jede Schleife darüber umfallen (§3).
func TestOhneMischpflichtStehtEineLeereListeDa(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartienaht(&klagendePartieleitung{})

	antwort := NeueBindungen(anwendung).Partie(t.Context(), "p1", "sitz-1")
	if antwort.Fehler != nil {
		t.Fatalf("Partie: %+v", antwort.Fehler)
	}

	if antwort.Partie.Mischpflichten == nil {
		t.Error("die Mischpflichten stehen als nil da")
	}

	// **Und auf dem Fehlerweg ebenso** — dort ist die Zusage in Phase 5.4
	// zweimal gebrochen worden, und beide Male galt sie nur für den Erfolg.
	leer := leerePartie(t.Context(), domain.ErrValidation)
	if leer.Partie.Mischpflichten == nil {
		t.Error("der Fehlerweg gibt nil")
	}
}
