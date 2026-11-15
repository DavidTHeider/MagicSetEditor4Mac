package app_test

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// **An der Fassade und nicht am Dienst**: was eine Rückkehr hinterlässt,
// entscheidet `internal/service`; hier steht, **wie es beim Fenster ankommt**.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ruecksprungdoppel gibt feste Antworten.
type ruecksprungdoppel struct {
	lage      domain.Ruecksprunglage
	epochen   []domain.Wiederherstellungsepoche
	quittiert int64
	fehler    error
}

func (r *ruecksprungdoppel) Ruecksprunglage(context.Context, string) (
	domain.Ruecksprunglage, error,
) {
	return r.lage, r.fehler
}

func (r *ruecksprungdoppel) OffeneRueckkehren(context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return r.epochen, r.fehler
}

func (r *ruecksprungdoppel) Rueckkehren(context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return r.epochen, r.fehler
}

func (r *ruecksprungdoppel) RueckkehrQuittieren(_ context.Context, kennung int64) error {
	r.quittiert = kennung

	return r.fehler
}

func mitRuecksprung(doppel *ruecksprungdoppel) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitRuecksprung(doppel)

	return anwendung
}

// **Vier Arten, und die zwei Neins stehen als Marke da.**
func TestDieVierArtenKommenMitIhrenHindernissen(t *testing.T) {
	t.Parallel()

	antwort := mitRuecksprung(&ruecksprungdoppel{}).Ruecksprungarten(t.Context())

	if antwort.Fehler != nil {
		t.Fatalf("die Arten: %v", antwort.Fehler)
	}

	hindernisse := map[string]string{}
	for _, art := range antwort.Arten {
		hindernisse[art.Art] = art.Hindernis
	}

	for art, erwartet := range map[string]string{
		"appfassung": domain.RuecksprungNachProzessende,
		"schema":     domain.RuecksprungNurVorwaerts,
		"daten":      "",
		"pakete":     "",
	} {
		if hindernisse[art] != erwartet {
			t.Errorf("%s gibt %q, erwartet war %q", art, hindernisse[art], erwartet)
		}
	}
}

// **Der Verlust steht als eigene Zahl im DTO** — vier Flächen rechneten ihn
// sonst viermal, und die vierte anders (§10).
func TestDieLageTraegtDenVerlustAlsZahl(t *testing.T) {
	t.Parallel()

	anwendung := mitRuecksprung(&ruecksprungdoppel{
		lage: domain.Ruecksprungvergleich(
			domain.Bestandsstand{
				AngelegtAm: time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC),
				Sets:       3, Karten: 107, Decks: -1, Exemplare: 4,
				SammlungGelesen: true,
				Sammlung: []domain.Sammlungsmenge{
					{Profil: 1, Schluessel: "scryfall:abc", Menge: 4, Name: "Bolt"},
				},
			},
			domain.Bestandsstand{
				Sets: 4, Karten: 119, Decks: 2, Exemplare: 1,
				SammlungGelesen: true,
				Sammlung: []domain.Sammlungsmenge{
					{Profil: 1, Schluessel: "scryfall:abc", Menge: 1},
				},
			},
		),
	})

	antwort := anwendung.Ruecksprunglage(t.Context(), "/probe.msesicherung", "")
	if antwort.Fehler != nil {
		t.Fatalf("die Lage: %v", antwort.Fehler)
	}

	if antwort.Datenstand != "2026-09-01T08:30:00Z" {
		t.Errorf("der Datenstand ist %q", antwort.Datenstand)
	}

	verluste := map[string]int{}
	unbekannt := map[string]bool{}

	for _, zeile := range antwort.Zeilen {
		verluste[zeile.Gegenstand] = zeile.Verlust
		unbekannt[zeile.Gegenstand] = zeile.Unbekannt
	}

	if verluste["karten"] != 12 {
		t.Errorf("Karten verlieren %d, erwartet waren 12", verluste["karten"])
	}

	if !unbekannt["decks"] {
		t.Error("die Deckzeile gilt nicht als unbekannt")
	}

	// **Und die unklare Menge kommt mit ihrem lesbaren Bezug.**
	if len(antwort.Unklar) != 1 || antwort.Unklar[0].Menge != 3 ||
		antwort.Unklar[0].Name != "Bolt" {
		t.Errorf("unklar ist %+v", antwort.Unklar)
	}
}

// **Ein Fehler kommt als Fehler zurück und nicht als leere Lage.**
func TestEineGescheiterteLageMeldetSich(t *testing.T) {
	t.Parallel()

	antwort := mitRuecksprung(&ruecksprungdoppel{
		fehler: errors.New("die Datei ist zu"),
	}).Ruecksprunglage(t.Context(), "/probe.msesicherung", "")

	if antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}

	// **Und die Listen sind leer und nicht `null`** (§3): „nichts unklar" ist
	// eine Aussage, `null` ist keine.
	if antwort.Zeilen == nil || antwort.Unklar == nil {
		t.Errorf("eine Liste ist null: %+v", antwort)
	}
}

func eineEpoche() domain.Wiederherstellungsepoche {
	return domain.Wiederherstellungsepoche{
		ID:              7,
		Zurueckgespielt: time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC),
		Datenstand:      time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC),
		Quelle:          "bestand.msesicherung",
		Klaerung:        domain.KlaerungOffen,
		Mengen: []domain.Sammlungsmenge{
			{Profil: 1, Schluessel: "scryfall:abc", Menge: 3, Name: "Bolt"},
		},
	}
}

// **Die offenen und alle kommen in derselben Gestalt.**
func TestDieRueckkehrenKommenMitZeitpunktUndMengen(t *testing.T) {
	t.Parallel()

	anwendung := mitRuecksprung(&ruecksprungdoppel{
		epochen: []domain.Wiederherstellungsepoche{eineEpoche()},
	})

	for name, antwort := range map[string]app.RueckkehrenAntwort{
		"alle":   anwendung.Rueckkehren(t.Context()),
		"offene": anwendung.OffeneRueckkehren(t.Context()),
	} {
		t.Run(name, func(t *testing.T) {
			if antwort.Fehler != nil {
				t.Fatalf("die Rückkehren: %v", antwort.Fehler)
			}

			if len(antwort.Rueckkehren) != 1 {
				t.Fatalf("es kamen %d zurück", len(antwort.Rueckkehren))
			}

			eine := antwort.Rueckkehren[0]

			if eine.Zurueckgespielt != "2026-09-10T12:00:00Z" {
				t.Errorf("die Rückkehr steht auf %q", eine.Zurueckgespielt)
			}

			// **Eine offene Epoche hat keinen Klärungszeitpunkt** — und der
			// bleibt **leer**, nicht `0001-01-01T00:00:00Z`.
			if eine.GeklaertAm != "" {
				t.Errorf("die offene Epoche nennt %q als Klärungszeitpunkt", eine.GeklaertAm)
			}

			if len(eine.Mengen) != 1 || eine.Mengen[0].Name != "Bolt" {
				t.Errorf("die Mengen sind %+v", eine.Mengen)
			}
		})
	}
}

// **Ein Fehler beim Lesen kommt durch** — statt einer leeren Liste, die wie
// „keine Rückkehr" aussähe.
func TestEinFehlerBeimLesenDerRueckkehrenKommtDurch(t *testing.T) {
	t.Parallel()

	antwort := mitRuecksprung(&ruecksprungdoppel{
		fehler: errors.New("das Buch ist zu"),
	}).Rueckkehren(t.Context())

	if antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}

	if antwort.Rueckkehren == nil {
		t.Error("die Liste ist null statt leer")
	}
}

func TestQuittierenReichtDieKennungDurch(t *testing.T) {
	t.Parallel()

	doppel := &ruecksprungdoppel{}

	if antwort := mitRuecksprung(doppel).RueckkehrQuittieren(t.Context(), 7); antwort.Fehler != nil {
		t.Fatalf("quittieren: %v", antwort.Fehler)
	}

	if doppel.quittiert != 7 {
		t.Errorf("quittiert wurde %d", doppel.quittiert)
	}
}

// **Und die Wurzel für die Flächen ohne Fenster geht wirklich durch den
// Bestand.**
//
// **Über eine echte Datenbank und nicht über ein Doppel**: die Frage ist, ob
// die Naht verdrahtet ist — ein Doppel dazwischen beantwortete sie nicht (§10).
func TestDieWurzelOhneFensterLiestAusDemBestand(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.RuecksprungUeber(db, ort)

	arten := anwendung.Ruecksprungarten(probekontext(t))
	if arten.Fehler != nil || len(arten.Arten) != 4 {
		t.Fatalf("die Arten: %+v", arten)
	}

	// **Ein frischer Bestand hat keine Rückkehr** — und das ist eine leere
	// Liste, kein Fehler.
	offene := anwendung.OffeneRueckkehren(probekontext(t))
	if offene.Fehler != nil {
		t.Fatalf("die offenen Rückkehren: %v", offene.Fehler)
	}

	if len(offene.Rueckkehren) != 0 {
		t.Errorf("ein frischer Bestand nennt Rückkehren: %+v", offene.Rueckkehren)
	}

	if alle := anwendung.Rueckkehren(probekontext(t)); alle.Fehler != nil {
		t.Fatalf("alle Rückkehren: %v", alle.Fehler)
	}

	// **Und eine Kennung, die es nicht gibt, wird abgewiesen.**
	if antwort := anwendung.RueckkehrQuittieren(probekontext(t), 4711); antwort.Fehler == nil {
		t.Error("eine unbekannte Rückkehr ließ sich quittieren")
	}

	// **Eine Datei, die es nicht gibt, meldet sich** — der Weg packt aus,
	// bevor er vergleicht.
	if lage := anwendung.Ruecksprunglage(probekontext(t),
		filepath.Join(t.TempDir(), "gibtesnicht.msesicherung"), ""); lage.Fehler == nil {
		t.Error("eine fehlende Datei kam durch")
	}
}
