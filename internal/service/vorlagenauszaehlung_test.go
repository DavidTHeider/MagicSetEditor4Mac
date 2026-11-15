package service_test

// **Die Auskunft, die Dimensionen und Karten zusammenbringt** — Werkstück C
// von Phase 4.11
// (#1338).
//
// Gegen die **Saat** und nicht gegen ein bezogenes Paket (§11):
// `probe.mse-game` bringt seit #1338 einen `statistics dimension:`-Block mit,
// dieselbe Gestalt wie `magic.mse-game/statistics_en`, eigener Inhalt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kartenablageprobe gibt die Karten eines Sets.
type kartenablageprobe struct {
	karten []*domain.CardRow
	fehler error
}

func (k *kartenablageprobe) Karten(_ context.Context, _ int64) (
	[]*domain.CardRow, error,
) {
	return k.karten, k.fehler
}

// **Die Auskunft findet die Dimension des Spielpakets — und zählt damit.**
func TestDieVorlagenauszaehlungFindetDieDimension(t *testing.T) {
	t.Parallel()

	stile := service.NeuerStilspeicher(&koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}, saatwurzel())

	karten := &kartenablageprobe{karten: []*domain.CardRow{
		{ID: 1, SetID: 7, Values: []domain.Value{{Field: "farbe", Text: "rot"}}},
		{ID: 2, SetID: 7, Values: []domain.Value{{Field: "farbe", Text: "blau"}}},
	}}

	auskunft := service.NeueVorlagenauszaehlung(stile, karten)

	angaben, vorlage, fehler := auskunft.Auszaehldimensionen(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Auszaehldimensionen: %v", fehler)
	}

	// **Die erklärte `farbe` am Ende**, davor die automatischen der Felder
	// `name` und `text` (#1985); das Feld `farbe` gibt keine zweite, weil die
	// erklärte gleichen Namens vorgeht.
	if len(angaben) != 3 || angaben[2].Name != "farbe" || angaben[2].Beschreibung == "" {
		t.Fatalf("die Angaben lauten %+v", angaben)
	}

	// **Die ganze Vorlage kommt mit**, nicht nur ihre Feldlisten — ohne den
	// Stil scheitert jedes `default:` (siehe [service.Auszaehlwerte]).
	if len(vorlage.Kartenfelder) == 0 || vorlage.Stil.Paket == "" {
		t.Errorf("die Vorlage lautet %+v", vorlage.Stil.Paket)
	}

	// **Und die Karten kommen durch.**
	gelesen, kartenfehler := auskunft.Karten(t.Context(), 7)
	if kartenfehler != nil || len(gelesen) != 2 {
		t.Errorf("die Karten lauten %+v / %v", gelesen, kartenfehler)
	}

	// **Und je Karte steht ihre Vorlage bereit** — ohne eigenen Stil die des
	// Sets.
	if eigene := auskunft.VorlageFuer(t.Context(), gelesen[0]); eigene.Stil.Paket == "" {
		t.Error("die Karte bekam keine Vorlage")
	}

	// **Der ganze Weg, nicht nur seine Hälften**: der Dienst über dieser
	// Auskunft zählt zwei Karten in die erklärte Reihenfolge — „blau" vor
	// „rot", obwohl „rot" alphabetisch früher käme.
	auszaehlung, laufFehler := service.NeuerAuszaehldienst(auskunft).
		Auszaehlen(t.Context(), 7, "farbe")
	if laufFehler != nil {
		t.Fatalf("Auszaehlen: %v", laufFehler)
	}

	if len(auszaehlung.Gruppen) != 2 ||
		auszaehlung.Gruppen[0].Name != "blau" ||
		auszaehlung.Gruppen[1].Name != "rot" {
		t.Errorf("die Gruppen lauten %+v", auszaehlung.Gruppen)
	}
}

// **Ein Set ohne Vorlage hat keine Dimensionen — und keinen Fehler.**
func TestEinSetOhneVorlageHatKeineDimensionen(t *testing.T) {
	t.Parallel()

	auskunft := service.NeueVorlagenauszaehlung(
		service.NeuerStilspeicher(&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Style: "gibtsnicht", Game: "gibtsnicht"},
		}}, saatwurzel()),
		&kartenablageprobe{})

	angaben, _, fehler := auskunft.Auszaehldimensionen(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Auszaehldimensionen: %v", fehler)
	}

	if len(angaben) != 0 {
		t.Errorf("die Angaben lauten %+v", angaben)
	}

	// **Und ohne Vorlage bekommt eine Karte die leere** — kein Absturz an
	// einem `nil`.
	if eigene := auskunft.VorlageFuer(
		t.Context(), &domain.CardRow{ID: 1, SetID: 7}); eigene.Stil.Paket != "" {
		t.Errorf("die Vorlage lautet %+v", eigene.Stil.Paket)
	}
}

// **Ein Lesefehler an den Karten kommt durch die Auskunft.**
func TestEinKartenfehlerKommtDurchDieVorlagenauszaehlung(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Bestand antwortet nicht")

	auskunft := service.NeueVorlagenauszaehlung(
		service.NeuerStilspeicher(&koepfequelle{}, saatwurzel()),
		&kartenablageprobe{fehler: eigener})

	if _, fehler := auskunft.Karten(t.Context(), 7); !errors.Is(fehler, eigener) {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}
