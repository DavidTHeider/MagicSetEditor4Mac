package service_test

// **Der Weg der Datenpflege** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

var pflegezeit = time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

// **Die Vorschau ändert nichts** — und trägt ihre Frist.
func TestDieVorschauAendertNichtsUndTraegtIhreFrist(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{staende: [][]domain.Pflegebefund{
		{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
	}}

	plan, fehler := pflege(doppel).Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if !plan.Gueltigbis.Equal(pflegezeit.Add(domain.Pflegefrist)) {
		t.Errorf("die Frist endet %v", plan.Gueltigbis)
	}

	if doppel.gesichert != 0 || doppel.geschrieben != nil {
		t.Error("die Vorschau hat etwas angefasst")
	}
}

// **Ein Fehler beim Nachsehen kommt durch** — statt eines leeren Plans, der
// wie ein sauberer Bestand aussähe.
func TestEineGescheiterteVorschauMeldetSich(t *testing.T) {
	t.Parallel()

	if _, fehler := pflege(&pflegedoppel{
		lesefehler: errors.New("die Datenbank ist zu"),
	}).Vorschau(t.Context()); fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// **Gesichert wird vor dem ersten Schreiben** — nicht danach.
func TestGesichertWirdVorDemErstenSchreiben(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		// **Drei Stände, denn es wird dreimal gelesen**: die Vorschau, der
		// Abgleich vor dem Schreiben und das Nachsehen danach.
		staende: [][]domain.Pflegebefund{
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
			{},
		},
		gelaufen: []domain.Pflegezeile{
			{Art: "kartenwerteOhneKarte", Ausgang: domain.PflegeErledigt, Anzahl: 1},
		},
	}

	plan, fehler := pflege(doppel).Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	bilanz, pflegefehler := pflege(doppel).Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}})
	if pflegefehler != nil {
		t.Fatalf("pflegen: %v", pflegefehler)
	}

	if doppel.gesichert != 1 {
		t.Errorf("es wurde %d-mal gesichert", doppel.gesichert)
	}

	if bilanz.Gesichert == "" {
		t.Error("die Bilanz nennt den Ort der Sicherung nicht")
	}

	if bilanz.Zeilen[0].Ausgang != domain.PflegeErledigt {
		t.Errorf("die Zeile ist %+v", bilanz.Zeilen[0])
	}
}

// **Die Bilanz nennt beide Stände.** „Drei erledigt" ist keine Auskunft über
// den Bestand; „vorher zwei Befunde, nachher einer" ist eine.
func TestDieBilanzNenntBeideStaende(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		staende: [][]domain.Pflegebefund{
			{
				befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc"),
				befund("moeglicheDubletten", domain.PflegeNurInformieren, "def"),
			},
			{
				befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc"),
				befund("moeglicheDubletten", domain.PflegeNurInformieren, "def"),
			},
			{befund("moeglicheDubletten", domain.PflegeNurInformieren, "def")},
		},
		gelaufen: []domain.Pflegezeile{
			{Art: "kartenwerteOhneKarte", Ausgang: domain.PflegeErledigt, Anzahl: 1},
		},
	}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	bilanz, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}})
	if pflegefehler != nil {
		t.Fatalf("pflegen: %v", pflegefehler)
	}

	if len(bilanz.Vorher) != 2 || len(bilanz.Nachher) != 1 {
		t.Errorf("die Stände sind %d und %d", len(bilanz.Vorher), len(bilanz.Nachher))
	}
}

// **Nichts zu tun heißt: nicht sichern.** Eine Sicherung ohne Schreibvorgang
// wäre eine Datei, die niemand erklärt.
func TestOhneZuTuendesWirdNichtGesichert(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{staende: [][]domain.Pflegebefund{
		{befund("moeglicheDubletten", domain.PflegeNurInformieren, "def")},
	}}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	bilanz, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "moeglicheDubletten", Revision: "def"}})
	if pflegefehler != nil {
		t.Fatalf("pflegen: %v", pflegefehler)
	}

	if doppel.gesichert != 0 {
		t.Error("es wurde gesichert, obwohl nichts zu tun war")
	}

	if len(bilanz.Zeilen) != 1 || bilanz.Zeilen[0].Ausgang != domain.PflegeUebersprungen {
		t.Errorf("die Bilanz ist %+v", bilanz.Zeilen)
	}
}

// **Scheitert die Sicherung, wird nichts geschrieben.** Das ist der Sinn der
// Reihenfolge: wer ohne Rückweg löschte, hätte die Zusage nicht.
func TestOhneSicherungWirdNichtsGeschrieben(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		staende: [][]domain.Pflegebefund{
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
		},
		sicherungsfehler: errors.New("die Platte ist voll"),
	}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if _, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}}); pflegefehler == nil {
		t.Fatal("ohne Sicherung wurde gepflegt")
	}

	if doppel.geschrieben != nil {
		t.Error("es wurde geschrieben, obwohl die Sicherung scheiterte")
	}
}

// **Ein Konflikt aus dem Abgleich bleibt in der Bilanz stehen** — ihn
// herauszunehmen hieße, dem Aufrufer die Hälfte zu verschweigen.
func TestEinKonfliktBleibtInDerBilanzStehen(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		staende: [][]domain.Pflegebefund{
			{
				befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc"),
				befund("setangabenOhneSet", domain.PflegeSicher, "def"),
			},
			{
				befund("kartenwerteOhneKarte", domain.PflegeSicher, "GEAENDERT"),
				befund("setangabenOhneSet", domain.PflegeSicher, "def"),
			},
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "GEAENDERT")},
		},
		gelaufen: []domain.Pflegezeile{
			{Art: "setangabenOhneSet", Ausgang: domain.PflegeErledigt, Anzahl: 1},
		},
	}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	bilanz, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt, []domain.Pflegewahl{
		{Art: "kartenwerteOhneKarte", Revision: "abc"},
		{Art: "setangabenOhneSet", Revision: "def"},
	})
	if pflegefehler != nil {
		t.Fatalf("pflegen: %v", pflegefehler)
	}

	ausgaenge := map[string]domain.Pflegeausgang{}
	for _, zeile := range bilanz.Zeilen {
		ausgaenge[zeile.Art] = zeile.Ausgang
	}

	if ausgaenge["kartenwerteOhneKarte"] != domain.PflegeKonflikt {
		t.Errorf("die geänderte Maßnahme ist %q", ausgaenge["kartenwerteOhneKarte"])
	}

	if ausgaenge["setangabenOhneSet"] != domain.PflegeErledigt {
		t.Errorf("die unveränderte Maßnahme ist %q", ausgaenge["setangabenOhneSet"])
	}

	// **Und nur die bereite Maßnahme ging ans Schreiben.**
	if len(doppel.geschrieben) != 1 {
		t.Errorf("geschrieben wurden %v", doppel.geschrieben)
	}
}

// **Ein Fehler beim Schreiben kommt durch.**
func TestEinFehlerBeimSchreibenDerPflegeKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		staende: [][]domain.Pflegebefund{
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
		},
		schreibfehler: errors.New("die Datenbank ist zu"),
	}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if _, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}}); pflegefehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// **Ein Fehler beim Nachsehen hält die Bilanz nicht auf** (§7, Regel 3):
// gepflegt ist gepflegt, und die Zeilen stehen da. **Gemeldet wird er
// trotzdem** — sonst sähe der zweite Stand aus wie eine Messung.
func TestEinFehlerBeimNachsehenMeldetSichUndLaesstDieBilanzStehen(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		staende: [][]domain.Pflegebefund{
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
		},
		gelaufen: []domain.Pflegezeile{
			{Art: "kartenwerteOhneKarte", Ausgang: domain.PflegeErledigt, Anzahl: 1},
		},
	}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	// **Erst beim dritten Lesen scheitert es** — die Vorschau und der Abgleich
	// sind da schon durch.
	doppel.staende = [][]domain.Pflegebefund{
		{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
	}

	weg = pflegeMit(&nachsehfehler{pflegedoppel: doppel, abDem: 1})

	bilanz, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}})
	if pflegefehler == nil {
		t.Fatal("der Fehler beim Nachsehen wurde verschwiegen")
	}

	if len(bilanz.Zeilen) != 1 || bilanz.Zeilen[0].Ausgang != domain.PflegeErledigt {
		t.Errorf("die Bilanz ist %+v", bilanz.Zeilen)
	}
}

// nachsehfehler scheitert erst ab dem n-ten Lesen.
type nachsehfehler struct {
	*pflegedoppel
	abDem int
	rufe  int
}

func (n *nachsehfehler) lesen(ctx context.Context, jetzt time.Time) (
	[]domain.Pflegebefund, error,
) {
	n.rufe++

	if n.rufe > n.abDem {
		return nil, errors.New("die Datenbank ging zwischendurch zu")
	}

	return n.pflegedoppel.lesen(ctx, jetzt)
}

// pflege für den Sonderfall — sie nimmt die überschriebene Lesefunktion.
func pflegeMit(doppel *nachsehfehler) *service.Datenpflege {
	weg := service.NeueDatenpflege(doppel.lesen, doppel.schreiben, doppel.sichern)
	weg.MitUhr(func() time.Time { return pflegezeit })

	return weg
}

// **Ein Fehler beim Abgleichen kommt durch** — ein abgelaufener Plan ist ein
// Konflikt und keine leere Bilanz.
func TestEinAbgelaufenerPlanKommtAlsFehlerDurch(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{staende: [][]domain.Pflegebefund{
		{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
	}}

	weg := pflege(doppel)

	plan, fehler := weg.Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	// **Die Uhr geht weiter als die Frist** — das ist der Fall, für den es sie
	// gibt.
	weg.MitUhr(func() time.Time { return pflegezeit.Add(domain.Pflegefrist).Add(time.Second) })

	if _, pflegefehler := weg.Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}}); pflegefehler == nil {
		t.Fatal("der abgelaufene Plan kam durch")
	}

	if doppel.gesichert != 0 || doppel.geschrieben != nil {
		t.Error("es wurde gearbeitet, obwohl der Plan abgelaufen war")
	}
}

// **Und ein Fehler beim Lesen vor dem Abgleich ebenso.**
func TestEinFehlerVorDemAbgleichKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{
		staende: [][]domain.Pflegebefund{
			{befund("kartenwerteOhneKarte", domain.PflegeSicher, "abc")},
		},
	}

	plan, fehler := pflege(doppel).Vorschau(t.Context())
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	doppel.lesefehler = errors.New("die Datenbank ist zu")

	if _, pflegefehler := pflege(doppel).Pflegen(t.Context(), plan.Erstellt,
		[]domain.Pflegewahl{{Art: "kartenwerteOhneKarte", Revision: "abc"}}); pflegefehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}
