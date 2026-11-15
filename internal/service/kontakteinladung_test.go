package service_test

// **Einladen, annehmen, vergleichen** — NET-01
// (#1705).
//
// # Was hier gemessen wird
//
// **Zwei echte Identitäten**, eine lädt ein und eine nimmt an. Geprüft wird
// am Ergebnis im Bestandsdouble: was steht nach der Annahme da — und was
// steht **nicht** da, wenn sie scheitert (AK-04).

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

var einladungsuhr = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

// einladungsablage ist ein Bestand für Kontakte und angenommene Einladungen.
type einladungsablage struct {
	kontakte map[string]domain.Peerkontakt
	nonces   map[string]bool
	fehler   error

	// Je Weg ein eigener Fehler — für die Wege hinter dem ersten Lesen.
	vorgaengerfehler, noncefehler, aufnahmefehler error

	// Was dieses Profil ausgestellt hat und was davon eingelöst ist
	// (NET-01b) — die Wege stehen in `einloesung_test.go`.
	ausgestellt                    map[string]time.Time
	eingeloest                     map[string]bool
	ausstellfehler, einloesefehler error
}

func neueEinladungsablage() *einladungsablage {
	return &einladungsablage{
		kontakte: map[string]domain.Peerkontakt{}, nonces: map[string]bool{},
		ausgestellt: map[string]time.Time{}, eingeloest: map[string]bool{},
	}
}

func (a *einladungsablage) Lesen(
	_ context.Context, kennung string,
) (domain.Peerkontakt, error) {
	if a.fehler != nil {
		return domain.Peerkontakt{}, a.fehler
	}

	if a.vorgaengerfehler != nil && kennung == "12D3KooWVorgaenger" {
		return domain.Peerkontakt{}, a.vorgaengerfehler
	}

	kontakt, da := a.kontakte[kennung]
	if !da {
		return domain.Peerkontakt{}, domain.ErrNotFound
	}

	return kontakt, nil
}

func (a *einladungsablage) Angenommen(_ context.Context, nonce string) (bool, error) {
	return a.nonces[nonce], a.noncefehler
}

func (a *einladungsablage) Aufnehmen(
	_ context.Context, aufnahme domain.Kontaktaufnahme, _ time.Time,
) error {
	if a.aufnahmefehler != nil {
		return a.aufnahmefehler
	}

	a.nonces[aufnahme.Nonce] = true
	a.kontakte[aufnahme.Kontakt.Kennung] = aufnahme.Kontakt

	return nil
}

// festeRotation sagt, welche Kennung vorher galt.
type festeRotation struct{ vorher peer.Kennung }

func (r festeRotation) Lage(context.Context) (service.Peerlage, error) {
	return service.Peerlage{Vorher: r.vorher}, nil
}

// einladungsseiten baut Anna, die einlädt, und Bert, der annimmt.
func einladungsseiten(t *testing.T) (
	anna, bert peer.Identitaet, annaWerk, bertWerk *service.Einladungswerk,
	bertAblage *einladungsablage,
) {
	t.Helper()

	anna, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	bert, fehler = peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	uhr := func() time.Time { return einladungsuhr }
	bertAblage = neueEinladungsablage()

	annaWerk = service.NeuesEinladungswerk(festerAusweis{ich: anna},
		festeRotation{}, neueEinladungsablage(), uhr)
	bertWerk = service.NeuesEinladungswerk(festerAusweis{ich: bert},
		festeRotation{}, bertAblage, uhr)

	return anna, bert, annaWerk, bertWerk, bertAblage
}

func annasCode(t *testing.T, werk *service.Einladungswerk) string {
	t.Helper()

	ausgestellt, fehler := werk.Erstellen(t.Context(), service.Einladungsbitte{
		Name: "Anna", Zweck: domain.ZweckHandel, Adresse: "192.168.1.42:41234",
	})
	if fehler != nil {
		t.Fatalf("Erstellen: %v", fehler)
	}

	if !ausgestellt.Ablauf.Equal(einladungsuhr.Add(domain.KontakteinladungGueltig)) {
		t.Fatalf("Ablauf %v", ausgestellt.Ablauf)
	}

	return ausgestellt.Code
}

func TestEinladungVorschau(t *testing.T) {
	t.Parallel()

	anna, bert, annaWerk, bertWerk, _ := einladungsseiten(t)

	vorschau, fehler := bertWerk.Pruefen(t.Context(), annasCode(t, annaWerk))
	if fehler != nil {
		t.Fatalf("Pruefen: %v", fehler)
	}

	if vorschau.Einladung.Kennung != string(anna.Kennung()) ||
		vorschau.Einladung.Name != "Anna" ||
		vorschau.Einladung.Zweck != domain.ZweckHandel ||
		vorschau.Lage != domain.LageUnbekannt || vorschau.Angenommen ||
		vorschau.Fingerabdruck != anna.Fingerabdruck() {
		t.Fatalf("Vorschau: %+v", vorschau)
	}

	// **Beide Seiten rechnen denselben Code** — Anna aus ihrer Sicht.
	if vorschau.Vergleichscode != domain.Vergleichscode(
		string(anna.Kennung()), string(bert.Kennung())) {
		t.Fatalf("Vergleichscode %q", vorschau.Vergleichscode)
	}
}

func TestEinladungAnnehmen(t *testing.T) {
	t.Parallel()

	anna, bert, annaWerk, bertWerk, ablage := einladungsseiten(t)
	code := annasCode(t, annaWerk)
	vergleich := domain.Vergleichscode(string(anna.Kennung()), string(bert.Kennung()))

	kontakt, fehler := bertWerk.Annehmen(t.Context(), service.Einladungsannahme{
		Code: code, EigenerName: "Anna vom Stammtisch", Verglichen: vergleich,
		Freigeben: true,
	})
	if fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if kontakt.Stand != domain.KontaktBestaetigt ||
		kontakt.EigenerName != "Anna vom Stammtisch" ||
		kontakt.Adresse != "192.168.1.42:41234" ||
		!slices.Equal(kontakt.Rechte, []domain.Zweck{domain.ZweckHandel}) {
		t.Fatalf("Kontakt: %+v", kontakt)
	}

	// **Dieselbe Einladung ein zweites Mal** — AK-04.
	_, fehler = bertWerk.Annehmen(t.Context(), service.Einladungsannahme{Code: code})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("zweite Annahme: %v", fehler)
	}

	if len(ablage.kontakte) != 1 {
		t.Fatalf("Kontakte: %d", len(ablage.kontakte))
	}
}

func TestEinladungOhneVergleich(t *testing.T) {
	t.Parallel()

	_, _, annaWerk, bertWerk, _ := einladungsseiten(t)

	kontakt, fehler := bertWerk.Annehmen(t.Context(),
		service.Einladungsannahme{Code: annasCode(t, annaWerk)})
	if fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	// **Aufgenommen, nicht bestätigt, und nichts freigegeben** — das ist
	// kein halber Kontakt, sondern der Stand „unbestätigt".
	if kontakt.Stand != domain.KontaktUnbestaetigt || len(kontakt.Rechte) != 0 {
		t.Fatalf("Kontakt: %+v", kontakt)
	}
}

func TestEinladungScheitertOhneSpur(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name  string
		bauen func(t *testing.T, code string, bert peer.Identitaet) service.Einladungsannahme
		ziel  error
	}{
		{"falscher Vergleichscode", func(_ *testing.T, code string, _ peer.Identitaet) service.Einladungsannahme {
			return service.Einladungsannahme{Code: code, Verglichen: "12345 67890 12345 67890 12345 67890"}
		}, domain.ErrValidation},
		{"Freigabe ohne Vergleich", func(_ *testing.T, code string, _ peer.Identitaet) service.Einladungsannahme {
			return service.Einladungsannahme{Code: code, Freigeben: true}
		}, domain.ErrValidation},
		{"verändert", func(_ *testing.T, code string, _ peer.Identitaet) service.Einladungsannahme {
			return service.Einladungsannahme{Code: fremdeDaten(t, code)}
		}, domain.ErrValidation},
		{"Name zu lang", func(_ *testing.T, code string, _ peer.Identitaet) service.Einladungsannahme {
			return service.Einladungsannahme{Code: code, EigenerName: strings.Repeat("x", 65)}
		}, domain.ErrValidation},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, bert, annaWerk, bertWerk, ablage := einladungsseiten(t)

			_, fehler := bertWerk.Annehmen(t.Context(), fall.bauen(t, annasCode(t, annaWerk), bert))
			if !errors.Is(fehler, fall.ziel) {
				t.Fatalf("Annehmen = %v", fehler)
			}

			if len(ablage.kontakte) != 0 || len(ablage.nonces) != 0 {
				t.Fatalf("es blieb etwas: %v %v", ablage.kontakte, ablage.nonces)
			}
		})
	}
}

// fremdeDaten tauscht den Namen in den Daten, ohne neu zu unterschreiben.
func fremdeDaten(t *testing.T, code string) string {
	t.Helper()

	einladung, _, fehler := domain.ZerlegeKontakteinladung(code)
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, unterschrift, _ := strings.Cut(code, ".")
	einladung.Name = "Mallory"

	neu, fehler := einladung.Kodiere(func([]byte) []byte { return nil })
	if fehler != nil {
		t.Fatal(fehler)
	}

	daten, _, _ := strings.Cut(neu, ".")

	return daten + "." + unterschrift
}

func TestEinladungAbgelaufen(t *testing.T) {
	t.Parallel()

	anna, _, annaWerk, _, _ := einladungsseiten(t)
	code := annasCode(t, annaWerk)

	spaeter := service.NeuesEinladungswerk(festerAusweis{ich: anna},
		festeRotation{}, neueEinladungsablage(), func() time.Time {
			return einladungsuhr.Add(domain.KontakteinladungGueltig)
		})

	// **Anna prüft ihre eigene** — der Ablauf greift vor der Frage, wessen
	// Einladung es ist; beides endet ohne Kontakt.
	if _, fehler := spaeter.Pruefen(t.Context(), code); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("abgelaufen: %v", fehler)
	}
}

func TestEinladungDieEigene(t *testing.T) {
	t.Parallel()

	_, _, annaWerk, _, _ := einladungsseiten(t)

	_, fehler := annaWerk.Pruefen(t.Context(), annasCode(t, annaWerk))
	if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), "eigene") {
		t.Fatalf("eigene Einladung: %v", fehler)
	}
}
