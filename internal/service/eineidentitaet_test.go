package service_test

// **Eine Verbindung hat eine Identität** — SEC-16
// (#2145,
// Vermerk 0176).
//
// Die horchende Seite verlangte ein Zeugnis und ließ danach jede gültige
// Unterschrift durch: TLS mit dem Schlüssel A und ein Handschlag von B
// ergaben eine Sitzung mit B. Jetzt wird beides gegeneinander gehalten —
// auf beiden Seiten. Ohne Zeugnis (eine Röhre im selben Prozess) sagt allein
// der Handschlag, wer da ist.

import (
	"context"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zeugnisleitung ist eine Röhre, die behauptet, ein Zeugnis gesehen zu
// haben — so, wie es eine TLS-Leitung tut.
type zeugnisleitung struct {
	*leitung.Leitung
	zeugnis peer.Kennung
}

func (z zeugnisleitung) Gegenstelle() peer.Kennung { return z.zeugnis }

type zeugniswaehler struct{ leitung zeugnisleitung }

func (w zeugniswaehler) Waehle(context.Context, peer.Kennung, string) (service.Verbindung, error) {
	return w.leitung, nil
}

func erzeugt(t *testing.T) peer.Identitaet {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return ich
}

func TestEineVerbindungHatEineIdentitaet(t *testing.T) {
	t.Parallel()

	hiesige, dortige, fremde := erzeugt(t), erzeugt(t), erzeugt(t)

	faelle := map[string]struct {
		hierSieht, dortSieht peer.Kennung
		gelingt              bool
	}{
		"beide Zeugnisse passen":                  {hierSieht: dortige.Kennung(), dortSieht: hiesige.Kennung(), gelingt: true},
		"ohne Zeugnis entscheidet der Handschlag": {gelingt: true},
		"die horchende Seite sieht ein fremdes":   {hierSieht: dortige.Kennung(), dortSieht: fremde.Kennung()},
		"die wählende Seite sieht ein fremdes":    {hierSieht: fremde.Kennung(), dortSieht: hiesige.Kennung()},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hiesigeLeitung, dortigeLeitung := leitung.Schleife("kanal-" + name)
			t.Cleanup(func() {
				_ = hiesigeLeitung.Schliesse() //nolint:errcheck // nur ein Lauf
				_ = dortigeLeitung.Schliesse() //nolint:errcheck // nur ein Lauf
			})

			hier := service.NeuePeersitzungen(&rechteprobe{darf: true},
				zeugniswaehler{leitung: zeugnisleitung{Leitung: hiesigeLeitung, zeugnis: fall.hierSieht}},
				ausweisprobe{ich: hiesige}, einePeeruhr, 2).MitZufall(nil)
			dort := service.NeuePeersitzungen(&rechteprobe{darf: true}, nil,
				ausweisprobe{ich: dortige}, einePeeruhr, 2)

			var (
				warten     sync.WaitGroup
				dortFehler error
			)

			warten.Add(1)

			go func() {
				defer warten.Done()

				_, dortFehler = dort.Annehmen(t.Context(),
					zeugnisleitung{Leitung: dortigeLeitung, zeugnis: fall.dortSieht})
				if dortFehler != nil {
					_ = dortigeLeitung.Schliesse() //nolint:errcheck // die andere Seite soll es merken
				}
			}()

			_, hierFehler := hier.Oeffne(t.Context(), dortige.Kennung(), domain.ZweckTeilen)
			if hierFehler != nil {
				_ = hiesigeLeitung.Schliesse() //nolint:errcheck // die andere Seite soll es merken
			}

			warten.Wait()

			if gelang := hierFehler == nil && dortFehler == nil; gelang != fall.gelingt {
				t.Errorf("gelungen %v, erwartet %v — hier %v, dort %v", gelang, fall.gelingt, hierFehler, dortFehler)
			}
		})
	}
}
