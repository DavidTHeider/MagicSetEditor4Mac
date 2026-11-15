package domain_test

// **Eine Vertragsart ohne Angebot** — #2435. Teilen und Deckleihe tragen
// keine Positionen; was die Schranke für sie sagt, ist nur, ob überhaupt
// gehandelt wird und ob am Markt.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieHandelsrichtlinieSagtJeVertragsart(t *testing.T) {
	t.Parallel()

	ohneHandel := domain.Standardhandelsrichtlinie()
	ohneHandel.Handeln = false

	mitMarkt := domain.Standardhandelsrichtlinie()
	mitMarkt.Markt = true

	for _, fall := range []struct {
		name       string
		richtlinie domain.Handelsrichtlinie
		art        domain.Handelsart
		erlaubt    bool
	}{
		{"teilt ohne Einstellung", domain.Standardhandelsrichtlinie(), domain.HandelTeilen, true},
		{"verleiht ohne Einstellung", domain.Standardhandelsrichtlinie(), domain.HandelDeckleihe, true},
		{"handelt ohne Einstellung nicht am Markt", domain.Standardhandelsrichtlinie(), domain.HandelMarkt, false},
		{"handelt am Markt, wenn er erlaubt ist", mitMarkt, domain.HandelMarkt, true},
		{"teilt nicht, wenn das Profil nicht handelt", ohneHandel, domain.HandelTeilen, false},
		{"verleiht nicht, wenn das Profil nicht handelt", ohneHandel, domain.HandelDeckleihe, false},
		{"tauscht nicht, wenn das Profil nicht handelt", ohneHandel, domain.HandelTausch, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.richtlinie.ErlaubtArt(fall.art)
			if fall.erlaubt != (fehler == nil) {
				t.Fatalf("ErlaubtArt(%s) = %v, erlaubt erwartet: %v", fall.art, fehler, fall.erlaubt)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehlerklasse %v, erwartet ErrValidation", fehler)
			}
		})
	}
}

// **Die Frist einer Zusage, gemessen an der Grenze** — #2435. Kein Weg nennt
// heute eine Frist; eine Zusage ohne hielte unbegrenzt fest, und das ist
// länger als jede Grenze.
func TestDieZusagefristFolgtDerGrenze(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	eng := domain.Handelsrichtlinie{Handeln: true, Fristgrenze: time.Hour}

	for _, fall := range []struct {
		name       string
		richtlinie domain.Handelsrichtlinie
		gewuenscht time.Time
		erwartet   time.Time
		abgewiesen bool
	}{
		{"ohne Wunsch die Grenze", eng, time.Time{}, jetzt.Add(time.Hour), false},
		{"ein kürzerer Wunsch bleibt", eng, jetzt.Add(time.Minute), jetzt.Add(time.Minute), false},
		{"genau die Grenze bleibt", eng, jetzt.Add(time.Hour), jetzt.Add(time.Hour), false},
		{"ein längerer Wunsch wird abgewiesen", eng, jetzt.Add(2 * time.Hour), time.Time{}, true},
		{"ohne Grenze bleibt kein Wunsch keiner", domain.Standardhandelsrichtlinie(), time.Time{}, time.Time{}, false},
		{"ohne Grenze bleibt jeder Wunsch", domain.Standardhandelsrichtlinie(), jetzt.Add(1000 * time.Hour), jetzt.Add(1000 * time.Hour), false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			frist, fehler := fall.richtlinie.Zusagefrist(jetzt, fall.gewuenscht)
			if fall.abgewiesen != (fehler != nil) || (fehler != nil && !errors.Is(fehler, domain.ErrValidation)) {
				t.Fatalf("Fehler %v, abgewiesen erwartet: %v", fehler, fall.abgewiesen)
			}

			if !frist.Equal(fall.erwartet) {
				t.Errorf("die Frist ist %v, erwartet %v", frist, fall.erwartet)
			}
		})
	}
}
