package service_test

// **Preise für viele Drucke** — Werkstück 6 von Phase 4.5
// (#1292).

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vieleDrucke legt so viele Exemplare an, wie verlangt — jedes mit eigenem Druck.
func vieleDrucke(t *testing.T, anzahl int) *sammlungsdoppel {
	t.Helper()

	ablage := &sammlungsdoppel{}

	for zaehler := range anzahl {
		if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
			Printing: druckkennung(zaehler), Menge: 1,
			Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
		}); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	return ablage
}

// **Ein Schub holt höchstens die Schubgrenze — und sagt, was bleibt.**
//
// Der Grund ist die Drossel (§16, Regel 4): dieses Programm wartet 600 ms je
// Anfrage, also sind fünfhundert Drucke rund **fünf Minuten** — und ein
// Aufruf, der das am Stück täte, wäre fünf Minuten ohne Lebenszeichen.
//
// Hier stand „knapp eine Minute", nachdem der Produkttext längst berichtigt
// war; der Nachlauf über Phase 4.5 hat den Rest gefunden.
func TestEinSchubHoltHoechstensDieSchubgrenze(t *testing.T) {
	t.Parallel()

	pflege := service.NeuePreispflege(&preisdoppel{},
		vieleDrucke(t, service.Schubgrenze+5), quellendoppel{
			preise: []scryfall.Preis{
				{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 100},
			},
		}, probeuhr())

	schub, fehler := pflege.PreiseInMasse(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("PreiseInMasse: %v", fehler)
	}

	if schub.Geholt != service.Schubgrenze {
		t.Errorf("%d geholt, erwartet %d", schub.Geholt, service.Schubgrenze)
	}

	if schub.Offen != 5 {
		t.Errorf("%d offen, erwartet 5", schub.Offen)
	}

	if schub.Fertig() {
		t.Error("der Schub gilt als fertig, obwohl fünf bleiben")
	}
}

// **Ein zweiter Schub holt den Rest — und ist dann fertig.**
func TestZweiSchuebeHolenAlles(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}

	pflege := service.NeuePreispflege(ablage,
		vieleDrucke(t, service.Schubgrenze+5), quellendoppel{
			preise: []scryfall.Preis{
				{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 100},
			},
		}, probeuhr())

	if _, fehler := pflege.PreiseInMasse(t.Context(), time.Hour); fehler != nil {
		t.Fatalf("erster Schub: %v", fehler)
	}

	zweiter, fehler := pflege.PreiseInMasse(t.Context(), time.Hour)
	if fehler != nil {
		t.Fatalf("zweiter Schub: %v", fehler)
	}

	if zweiter.Geholt != 5 {
		t.Errorf("%d geholt, erwartet 5", zweiter.Geholt)
	}

	if !zweiter.Fertig() {
		t.Errorf("der zweite Schub gilt als nicht fertig: %+v", zweiter)
	}
}

// **Ein frischer Preis wird nicht zweimal geholt.**
//
// Ohne diese Auswahl fragte jeder Lauf alles neu — und die Drossel machte
// daraus jedes Mal dieselbe Minute.
func TestEinFrischerPreisWirdNichtZweimalGeholt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}

	pflege := service.NeuePreispflege(ablage, vieleDrucke(t, 3), quellendoppel{
		preise: []scryfall.Preis{
			{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 100},
		},
	}, probeuhr())

	if _, fehler := pflege.PreiseInMasse(t.Context(), time.Hour); fehler != nil {
		t.Fatalf("erster Schub: %v", fehler)
	}

	zweiter, fehler := pflege.PreiseInMasse(t.Context(), time.Hour)
	if fehler != nil {
		t.Fatalf("zweiter Schub: %v", fehler)
	}

	if zweiter.Geholt != 0 || !zweiter.Fertig() {
		t.Errorf("der zweite Schub holte %d, offen %d", zweiter.Geholt, zweiter.Offen)
	}
}

// **Ein Alter von null heißt: alles neu holen.**
func TestEinAlterVonNullHoltAllesNeu(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}

	pflege := service.NeuePreispflege(ablage, vieleDrucke(t, 3), quellendoppel{
		preise: []scryfall.Preis{
			{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 100},
		},
	}, probeuhr())

	if _, fehler := pflege.PreiseInMasse(t.Context(), 0); fehler != nil {
		t.Fatalf("erster Schub: %v", fehler)
	}

	zweiter, fehler := pflege.PreiseInMasse(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("zweiter Schub: %v", fehler)
	}

	if zweiter.Geholt != 3 {
		t.Errorf("%d geholt, erwartet 3", zweiter.Geholt)
	}
}

// **Derselbe Druck zweimal wird einmal geholt.**
//
// Wer eine Karte in zwei Zuständen besitzt, hat zwei Zeilen und **einen**
// Druck — und die Drossel soll ihn nicht doppelt zählen.
func TestDerselbeDruckWirdEinmalGeholt(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	for _, zustand := range []domain.Zustand{
		domain.ZustandNearMint, domain.ZustandPlayed,
	} {
		if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
			Printing: probeprinting, Menge: 1, Zustand: zustand,
			Sprache: domain.Vorgabesprache,
		}); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	schub, fehler := service.NeuePreispflege(&preisdoppel{}, ablage, quellendoppel{
		preise: []scryfall.Preis{
			{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 100},
		},
	}, probeuhr()).PreiseInMasse(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("PreiseInMasse: %v", fehler)
	}

	if schub.Geholt != 1 {
		t.Errorf("%d geholt, erwartet 1", schub.Geholt)
	}
}

// **Ein Ausfall hält den Schub nicht auf** (§8).
func TestEinAusfallHaeltDenSchubNichtAuf(t *testing.T) {
	t.Parallel()

	schub, fehler := service.NeuePreispflege(&preisdoppel{}, vieleDrucke(t, 3),
		quellendoppel{fehler: errors.New("kein Netz")}, probeuhr()).
		PreiseInMasse(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("PreiseInMasse: %v", fehler)
	}

	if schub.Geholt != 0 {
		t.Errorf("%d geholt, erwartet 0", schub.Geholt)
	}

	// **Drei Befunde, einer je Druck** — und die Stelle nennt den Druck.
	if len(schub.Befunde) != 3 {
		t.Fatalf("%d Befunde, erwartet 3: %v", len(schub.Befunde), schub.Befunde)
	}

	if schub.Befunde[0].Ort == "" {
		t.Error("der Befund nennt den Druck nicht")
	}
}

// **Ein Lesefehler des Bestands kommt durch.**
func TestEinLesefehlerBeimSchub(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	_, fehler := service.NeuePreispflege(&preisdoppel{},
		&sammlungsdoppel{fehler: gestolpert}, quellendoppel{}, probeuhr()).
		PreiseInMasse(t.Context(), 0)

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}
}

// **Und einer der Preisablage ebenso.**
func TestEinLesefehlerDerPreisablageBeimSchub(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	_, fehler := service.NeuePreispflege(&preisdoppel{fehler: gestolpert},
		vieleDrucke(t, 1), quellendoppel{}, probeuhr()).
		PreiseInMasse(t.Context(), 0)

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}
}

// **Ein Schub zählt Versuche, nicht Erfolge.**
//
// # Der Befund, der diesen Test verlangt
//
// Die Abbruchbedingung lautete `schub.Geholt >= Schubgrenze`, und ein
// Fehlschlag zählte nicht mit. **Scheiterten fünfhundert Drucke, rief ein
// einziger „Schub" alle fünfhundert** — genau die Wartezeit ohne
// Lebenszeichen, gegen die es die Schübe gibt.
//
// Befund 7 des Zweitblicks über Phase 4.5
// (#1292).
func TestEinSchubZaehltVersucheUndNichtErfolge(t *testing.T) {
	t.Parallel()

	quelle := &zaehlendeQuelle{fehler: errors.New("kein Netz")}

	schub, fehler := service.NeuePreispflege(&preisdoppel{},
		vieleDrucke(t, service.Schubgrenze+30), quelle, probeuhr()).
		PreiseInMasse(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("PreiseInMasse: %v", fehler)
	}

	if quelle.rufe.Load() != int64(service.Schubgrenze) {
		t.Errorf("%d Anfragen, erwartet %d — ein Fehlschlag kostet dieselbe "+
			"Wartezeit wie ein Treffer", quelle.rufe.Load(), service.Schubgrenze)
	}

	// **Und die gescheiterten bleiben offen**, denn sie sind es: kein Preis
	// wurde abgelegt, und der nächste Ruf findet sie wieder.
	//
	// Hier stand `30` — die Zahl, die herauskam, wenn ein Fehlschlag `Offen`
	// verminderte. Sie schrieb die falsche Annahme fest, und der Nachlauf über
	// Phase 4.5 hat es angestrichen: sie sagt, zwanzig von fünfzig seien
	// erledigt, während nichts geschrieben wurde.
	if schub.Offen != service.Schubgrenze+30 {
		t.Errorf("%d offen, erwartet %d", schub.Offen, service.Schubgrenze+30)
	}

	if schub.Geholt != 0 {
		t.Errorf("%d geholt, erwartet 0", schub.Geholt)
	}

	if schub.Fehlgeschlagen != service.Schubgrenze {
		t.Errorf("%d fehlgeschlagen, erwartet %d",
			schub.Fehlgeschlagen, service.Schubgrenze)
	}

	// **Und der Schub sagt, dass er steht.** Ohne diesen Satz riefe ein
	// Aufrufer in der Schleife endlos weiter — dieselbe Anfrage, dasselbe
	// Ergebnis.
	if !schub.Steht() {
		t.Error("der Schub meldet nicht, dass er steht")
	}
}

// zaehlendeQuelle zählt, wie oft sie gefragt wurde.
type zaehlendeQuelle struct {
	rufe   atomic.Int64
	fehler error
}

func (z *zaehlendeQuelle) PreiseFuerDruck(_ context.Context, _ string) (
	[]domain.Preisstand, time.Duration, []domain.Befund, error,
) {
	z.rufe.Add(1)

	return nil, 0, nil, z.fehler
}

// Name gibt die Quelle, in deren Namen dieses Doppel antwortet.
func (*zaehlendeQuelle) Name() domain.Preisquelle { return domain.PreisquelleScryfall }

// **Ein Schub, in dem nichts geglückt ist, meldet nicht »fertig«.**
//
// # Der Befund, der diesen Test verlangt
//
// Ein Fehlschlag verminderte `Offen`. Der nächste Ruf rechnet die offenen
// Drucke aber aus der **unveränderten** Preisablage neu — die gescheiterten
// stehen dann wieder da. Bei genau `Schubgrenze` Fehlschlägen meldete der
// Schub `Fertig`, obwohl kein einziger Preis abgelegt wurde.
//
// **Der alte Grund für das Vermindern war richtig gesehen und falsch
// verbucht**: „ein gescheiterter Druck bleibt sonst für immer offen, und der
// Aufrufer riefe endlos weiter." Das stimmt — nur ist die Abhilfe nicht, eine
// falsche Zahl zu melden, sondern die Fehlschläge zu **zählen**: wer `Geholt:
// 0` neben `Fehlgeschlagen: 20` sieht, weiß, dass ein weiterer Ruf dasselbe
// gibt.
//
// Nachlauf über Phase 4.5, Befund 7
// (#1292).
func TestEinSchubOhneErfolgIstNichtFertig(t *testing.T) {
	t.Parallel()

	quelle := &zaehlendeQuelle{fehler: errors.New("kein Netz")}

	schub, fehler := service.NeuePreispflege(&preisdoppel{},
		vieleDrucke(t, service.Schubgrenze), quelle, probeuhr()).
		PreiseInMasse(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("PreiseInMasse: %v", fehler)
	}

	if schub.Fertig() {
		t.Errorf("der Schub meldet »fertig«, obwohl kein Preis abgelegt wurde: %+v", schub)
	}

	if schub.Offen != service.Schubgrenze {
		t.Errorf("%d offen, erwartet %d — kein Preis wurde abgelegt",
			schub.Offen, service.Schubgrenze)
	}

	if schub.Fehlgeschlagen != service.Schubgrenze {
		t.Errorf("%d fehlgeschlagen, erwartet %d",
			schub.Fehlgeschlagen, service.Schubgrenze)
	}
}
