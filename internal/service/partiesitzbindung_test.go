package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein an einen Sitz gebundenes Token sieht nur diesen Sitz** —
// #1819,
// Befund 2 des Zweitblicks über Phase 6.1.
//
// Jede Stelle der Partieleitung, die eine Partie liest, verändert oder
// ausgibt, fragt die Bindung aus dem Kontext. Der Lauf geht sie einzeln ab —
// gebunden an `sitz-1` von `p1` — und prüft die Gegenprobe: ungebunden geht
// jeder Weg wie bisher.

// anSitzEins ist ein Kontext, gebunden an `sitz-1` von `p1`.
func anSitzEins() context.Context {
	return service.MitSitzbindung(context.Background(),
		domain.Sitzbindung{Partie: "p1", Sitz: "sitz-1"})
}

// sitzwege sind die Wege, die an einem Sitz hängen.
func sitzwege(leitung *service.Partieleitung) map[string]func(ctx context.Context, sitz string) error {
	return map[string]func(ctx context.Context, sitz string) error{
		"die Sicht": func(ctx context.Context, sitz string) error {
			_, fehler := leitung.Partie(ctx, "p1", sitz)
			return fehler
		},
		"das Tischbild": func(ctx context.Context, sitz string) error {
			_, fehler := leitung.Tischbild(ctx, "p1", sitz)
			return fehler
		},
		"der Verlauf": func(ctx context.Context, sitz string) error {
			_, fehler := leitung.Verlauf(ctx, "p1", sitz)
			return fehler
		},
		"der Wirtsstand": func(ctx context.Context, sitz string) error {
			_, fehler := leitung.Wirtsstand(ctx, "p1", sitz)
			return fehler
		},
		"das Nachreichen": func(ctx context.Context, sitz string) error {
			_, fehler := leitung.Nachreichen(ctx, "p1", sitz, 0)
			return fehler
		},
		"ein Befehl": func(ctx context.Context, sitz string) error {
			_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
				Kennung: "b-" + sitz, PartieID: "p1", Absender: sitz,
				Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
			})
			return fehler
		},
		"das Zurücknehmen einer Freigabe": func(ctx context.Context, sitz string) error {
			return leitung.Zuruecknehmen(ctx, "p1", sitz)
		},
	}
}

func TestEinGebundenesTokenSiehtNurSeinenSitz(t *testing.T) {
	t.Parallel()

	for name := range sitzwege(nil) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			leitung, speicher := eineLeitung(t)
			speicher.partien["p1"] = einePartie()
			wege := sitzwege(leitung)

			// **Der fremde Sitz**: erlaubnis, gleich was der Weg sonst täte.
			if fehler := wege[name](anSitzEins(), "sitz-2"); !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Errorf("fremder Sitz: %v, erwartet erlaubnis", fehler)
			}

			// **Der eigene Sitz**: nicht an der Bindung abgewiesen.
			if fehler := wege[name](anSitzEins(), "sitz-1"); errors.Is(fehler, domain.ErrErlaubnis) {
				t.Errorf("eigener Sitz: %v", fehler)
			}

			// **Die Gegenprobe**: ungebunden geht der fremde Sitz wie bisher.
			if fehler := wege[name](context.Background(), "sitz-2"); errors.Is(fehler, domain.ErrErlaubnis) {
				t.Errorf("ungebunden: %v", fehler)
			}
		})
	}
}

// **Eine andere Partie** ist für ein gebundenes Token zu — auch auf den
// Wegen, die keinen Sitz nennen.
func TestEinGebundenesTokenBleibtInSeinerPartie(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	speicher.partien["p1"] = einePartie()

	ctx := anSitzEins()

	for name, fehler := range map[string]error{
		"eine andere Partie ansehen":   ohneErgebnis(leitung.Partie(ctx, "p2", "sitz-1")),
		"eine andere Partie anhalten":  leitung.Pausieren(ctx, "p2"),
		"eine andere Partie beenden":   leitung.Beenden(ctx, "p2"),
		"eine andere fortsetzen":       leitung.Fortsetzen(ctx, "p2"),
		"die Freigaben einer anderen":  ohneErgebnis(leitung.Freigaben(ctx, "p2")),
		"für einen fremden Sitz geben": leitung.Offenlegen(ctx, "p1", domain.Offenlegung{Von: "sitz-2", An: []string{"sitz-1"}}),
	} {
		if !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Errorf("%s: %v, erwartet erlaubnis", name, fehler)
		}
	}

	// **Die eigene Partie ohne Sitz geht** — anhalten darf, wer mitspielt.
	if fehler := leitung.Pausieren(ctx, "p1"); fehler != nil {
		t.Errorf("die eigene anhalten: %v", fehler)
	}
}

// **Der ganze Zustand geht nie an ein gebundenes Token** — Ausfuhr, Kopie,
// Einfuhr, Löschen und Vergessen tragen jede Hand oder räumen sie weg.
func TestEinGebundenesTokenBekommtNieDenGanzenZustand(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	speicher.partien["p1"] = einePartie()

	ctx := anSitzEins()

	for name, fehler := range map[string]error{
		"die Ausfuhr":   ohneErgebnis(leitung.Ausfuhr(ctx, "p1")),
		"die Kopie":     ohneErgebnis(leitung.Kopieren(ctx, "p1", "p9")),
		"die Einfuhr":   ohneErgebnisMitBefunden(leitung.Einfuhr(ctx, []byte("{}"), "p9")),
		"das Löschen":   leitung.Loeschen(ctx, "p1"),
		"das Vergessen": leitung.Vergessen(ctx, "p1"),
	} {
		if !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Errorf("%s: %v, erwartet erlaubnis", name, fehler)
		}
	}

	// **Die Gegenprobe**: ungebunden geht die Ausfuhr.
	if _, fehler := leitung.Ausfuhr(context.Background(), "p1"); fehler != nil {
		t.Errorf("ungebunden: %v", fehler)
	}
}

func ohneErgebnis[T any](_ T, fehler error) error { return fehler }

func ohneErgebnisMitBefunden[T, B any](_ T, _ B, fehler error) error { return fehler }
