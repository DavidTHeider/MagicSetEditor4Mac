package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Verweis auf einen echten Druck** — Werkstück 2 von Phase 3.6
// (#1167).

// druckdoppel gibt eine feste Antwort auf `KarteNachKennung`.
type druckdoppel struct {
	karte   *scryfall.Karte
	alter   time.Duration
	befunde []domain.Befund
	fehler  error
	gefragt string
}

func (d *druckdoppel) KarteNachKennung(_ context.Context, kennung string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	d.gefragt = kennung

	return d.karte, d.alter, d.befunde, d.fehler
}

// verweisdoppel hält Verweise im Arbeitsspeicher.
type verweisdoppel struct {
	verweise map[int64]domain.Kartenverweis
	fehler   error
}

func neuesVerweisdoppel() *verweisdoppel {
	return &verweisdoppel{verweise: map[int64]domain.Kartenverweis{}}
}

func (v *verweisdoppel) Setzen(
	_ context.Context, karteID int64, verweis domain.Kartenverweis,
) error {
	if v.fehler != nil {
		return v.fehler
	}

	v.verweise[karteID] = verweis

	return nil
}

func (v *verweisdoppel) Fuer(_ context.Context, karteID int64) (*domain.Kartenverweis, error) {
	if v.fehler != nil {
		return nil, v.fehler
	}

	verweis, da := v.verweise[karteID]
	if !da {
		return nil, nil //nolint:nilnil // „keiner" ist kein Fehler
	}

	return &verweis, nil
}

func (v *verweisdoppel) Loesen(_ context.Context, karteID int64) (bool, error) {
	if v.fehler != nil {
		return false, v.fehler
	}

	_, da := v.verweise[karteID]
	delete(v.verweise, karteID)

	return da, nil
}

func (v *verweisdoppel) JeSet(_ context.Context, _ int64) (
	map[int64]domain.Kartenverweis, error,
) {
	return v.verweise, v.fehler
}

func probedruck() *scryfall.Karte {
	return &scryfall.Karte{
		Kennung: "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1",
		Name:    "Lightning Bolt", Set: "lea", Setname: "Limited Edition Alpha",
		Nummer: "161", Adresse: "https://scryfall.com/card/lea/161",
	}
}

func verweisdienst(t *testing.T, druck *druckdoppel, ablage *verweisdoppel) *service.Kartenverweise {
	t.Helper()

	jetzt := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	return service.NeueKartenverweise(druck, ablage, func() time.Time { return jetzt })
}

// **Ein Verweis wird beim Setzen aufgelöst, nicht geglaubt.**
//
// Wer eine UUID ablegte, ohne sie nachzuschlagen, hätte einen Verweis, den
// niemand lesen kann — und der sich beim ersten Abgleich als Tippfehler
// herausstellte, lange nachdem ihn jemand gesetzt hat.
func TestEinVerweisWirdBeimSetzenAufgeloest(t *testing.T) {
	t.Parallel()

	druck := &druckdoppel{karte: probedruck()}
	ablage := neuesVerweisdoppel()

	verweis, _, err := verweisdienst(t, druck, ablage).
		Setzen(t.Context(), 7, "516E137A-D9D9-4BA7-89FB-E680DDB2C2D1")
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if verweis.Name != "Lightning Bolt" || verweis.Setcode != "lea" ||
		verweis.Nummer != "161" || verweis.Adresse != probedruck().Adresse {
		t.Errorf("aufgelöst wurde %+v", *verweis)
	}

	if abgelegt := ablage.verweise[7]; abgelegt.Printing != probedruck().Kennung {
		t.Errorf("abgelegt wurde %+v", abgelegt)
	}

	if !verweis.Gesetzt.Equal(time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("gesetzt am %v", verweis.Gesetzt)
	}
}

// **Ein Druck, den es nicht gibt, wird nicht abgelegt.**
func TestEinUnbekannterDruckWirdNichtAbgelegt(t *testing.T) {
	t.Parallel()

	druck := &druckdoppel{fehler: domain.ErrNotFound}
	ablage := neuesVerweisdoppel()

	_, _, err := verweisdienst(t, druck, ablage).Setzen(t.Context(), 7, "gibtsnicht")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Fehler war %v — will ErrNotFound", err)
	}

	if len(ablage.verweise) != 0 {
		t.Errorf("es wurde trotzdem abgelegt: %v", ablage.verweise)
	}
}

// **Die Befunde der Antwort reisen mit.** Ein unbekannter Aufzählungswert soll
// den sehen, der den Verweis setzt — nicht nur den, der die Antwort gelesen
// hat (§8, §16).
func TestDieBefundeDesDrucksReisenMit(t *testing.T) {
	t.Parallel()

	druck := &druckdoppel{
		karte:   probedruck(),
		befunde: []domain.Befund{{Meldung: "layout »zerplatzt« ist unbekannt"}},
	}

	_, befunde, err := verweisdienst(t, druck, neuesVerweisdoppel()).
		Setzen(t.Context(), 7, probedruck().Kennung)
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if len(befunde) != 1 || befunde[0].Meldung == "" {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Ein Druck ohne Namen wird abgewiesen.** Die Domäne verlangt ihn (§16,
// Regel 5); dass die Prüfung auch auf diesem Weg greift, ist die Zusage.
func TestEinDruckOhneNamenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	druck := &druckdoppel{karte: &scryfall.Karte{Kennung: "516e137a"}}
	ablage := neuesVerweisdoppel()

	_, _, err := verweisdienst(t, druck, ablage).Setzen(t.Context(), 7, "516e137a")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", err)
	}

	if len(ablage.verweise) != 0 {
		t.Errorf("es wurde trotzdem abgelegt: %v", ablage.verweise)
	}
}

// **Lösen, wo keiner war, ist ein Nein und kein Erfolg.**
func TestVerweisLoesenOhneVerweisIstNichtGefunden(t *testing.T) {
	t.Parallel()

	err := verweisdienst(t, &druckdoppel{}, neuesVerweisdoppel()).Loesen(t.Context(), 7)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Fehler war %v — will ErrNotFound", err)
	}
}

// **Und lösen, wo einer war, geht durch.**
func TestVerweisLoesenEntferntIhn(t *testing.T) {
	t.Parallel()

	ablage := neuesVerweisdoppel()
	dienst := verweisdienst(t, &druckdoppel{karte: probedruck()}, ablage)

	if _, _, err := dienst.Setzen(t.Context(), 7, probedruck().Kennung); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if err := dienst.Loesen(t.Context(), 7); err != nil {
		t.Fatalf("lösen: %v", err)
	}

	if len(ablage.verweise) != 0 {
		t.Errorf("es blieb %v zurück", ablage.verweise)
	}
}

// **Ohne Kennung wird nichts nachgeschlagen.** Der Aufruf endet, bevor er
// hinausgeht — eine leere Kennung ist eine Eingabe und kein Netzfall.
func TestOhneKennungWirdNichtsNachgeschlagen(t *testing.T) {
	t.Parallel()

	druck := &druckdoppel{karte: probedruck()}

	_, _, err := verweisdienst(t, druck, neuesVerweisdoppel()).Setzen(t.Context(), 7, "  ")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", err)
	}

	if druck.gefragt != "" {
		t.Errorf("es wurde nach %q gefragt", druck.gefragt)
	}
}

// **Lesen gibt den Verweis oder nichts.**
func TestVerweisLesenGibtIhnOderNichts(t *testing.T) {
	t.Parallel()

	ablage := neuesVerweisdoppel()
	dienst := verweisdienst(t, &druckdoppel{karte: probedruck()}, ablage)

	gelesen, err := dienst.Fuer(t.Context(), 7)
	if err != nil || gelesen != nil {
		t.Fatalf("ohne Verweis kam %v, %v", gelesen, err)
	}

	if _, _, setzfehler := dienst.Setzen(
		t.Context(), 7, probedruck().Kennung); setzfehler != nil {
		t.Fatalf("setzen: %v", setzfehler)
	}

	gelesen, err = dienst.Fuer(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen == nil || gelesen.Name != "Lightning Bolt" {
		t.Errorf("gelesen wurde %v", gelesen)
	}
}

// **Und die Verweise eines Sets kommen in einem Zug.**
func TestDieVerweiseEinesSetsKommenInEinemZug(t *testing.T) {
	t.Parallel()

	ablage := neuesVerweisdoppel()
	dienst := verweisdienst(t, &druckdoppel{karte: probedruck()}, ablage)

	if _, _, err := dienst.Setzen(t.Context(), 7, probedruck().Kennung); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	verweise, err := dienst.JeSet(t.Context(), 3)
	if err != nil {
		t.Fatalf("je Set: %v", err)
	}

	if len(verweise) != 1 || verweise[7].Name != "Lightning Bolt" {
		t.Errorf("es kamen %v", verweise)
	}
}

// **Ein Bestand, der nicht schreiben kann, meldet es** — und der Verweis
// entsteht nicht.
func TestEinUnschreibbarerBestandMeldetEs(t *testing.T) {
	t.Parallel()

	ablage := neuesVerweisdoppel()
	ablage.fehler = domain.ErrBelegt

	_, _, err := verweisdienst(t, &druckdoppel{karte: probedruck()}, ablage).
		Setzen(t.Context(), 7, probedruck().Kennung)
	if !errors.Is(err, domain.ErrBelegt) {
		t.Fatalf("Fehler war %v — will ErrBelegt", err)
	}
}

// **Und ein Bestand, der beim Lösen scheitert, ebenso.**
//
// **Kein „nicht gefunden" daraus machen**: „ich konnte nicht nachsehen" ist
// etwas anderes als „es war keiner da", und die zweite Auskunft ließe jemanden
// glauben, er habe die falsche Karte erwischt.
func TestEinBestandDerBeimLoesenScheitertMeldetEs(t *testing.T) {
	t.Parallel()

	ablage := neuesVerweisdoppel()
	ablage.fehler = domain.ErrBelegt

	err := verweisdienst(t, &druckdoppel{}, ablage).Loesen(t.Context(), 7)
	if !errors.Is(err, domain.ErrBelegt) {
		t.Fatalf("Fehler war %v — will ErrBelegt", err)
	}
}

// **Der Zeitpunkt ist der des Standes, nicht der des Klicks** — Befund 9 des
// Zweitblicks über Phase 3.6.
//
// # Warum das zählt
//
// `KarteNachKennung` gibt ohne Netz den letzten guten Stand samt seinem Alter
// (§16, Regel 5). Hier stand die Uhr des Aufrufs, und damit trug ein drei Tage
// alter Name den heutigen Zeitpunkt — die Oberfläche zeigt ihn als „bezogen
// am", und das ist eine Aussage über das **Alter der Anzeige**.
func TestDerZeitpunktIstDerDesStandes(t *testing.T) {
	t.Parallel()

	druck := &druckdoppel{karte: probedruck(), alter: 72 * time.Hour}

	verweis, _, err := verweisdienst(t, druck, neuesVerweisdoppel()).
		Setzen(t.Context(), 7, probedruck().Kennung)
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	jetzt := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	if !verweis.Gesetzt.Equal(jetzt.Add(-72 * time.Hour)) {
		t.Errorf("gesetzt am %v — erwartet drei Tage vor %v", verweis.Gesetzt, jetzt)
	}
}
