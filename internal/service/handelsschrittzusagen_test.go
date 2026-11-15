package service_test

// **Was ein Schritt an den festgehaltenen Exemplaren bewirkt** —
// #1643.
//
// # Warum das in der Leitung steht und nicht daneben
//
// **Weil es die Bedeutung der Schritte ist und keine Nebenwirkung.**
// `vorbereiten` heißt in der Zustandsmatrix „die Exemplare sind reserviert";
// bis zu diesem Werkstück bewegte es den **Zustand** und hielt **nichts**
// fest. `uebergeben` heißt „sie sind über den Tisch gegangen", und genau das
// liest die Verbuchung aus der Spalte `uebergeben`.
//
// # Und warum die Wirkung vor dem Schreiben steht
//
// **Damit eine Absage nichts hinterlässt** — dieselbe Zusage wie im Kopf von
// `handelsleitung.go`: jede Ablehnung geschieht, bevor irgendetwas
// geschrieben wurde.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einHandelMitAbgang legt einen Vorgang an, bei dem **diese** Seite etwas
// weggibt — der Fall, für den es die Zusagen gibt.
func einHandelMitAbgang(
	zustand domain.Handelszustand,
) domain.Handelsangebot {
	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-abgang", Revision: 1,
		Art: domain.HandelTausch, Initiator: "bich",
		Parteien: []string{"bich", "bdort"},
		Zustand:  zustand,
		Positionen: []domain.Handelsposition{{
			Von: "bich", Printing: "a1b2", Menge: 2,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	}

	// **Ohne vorgemerkte Stimmen.** Ein Beschluss ist mit seinem Wechsel
	// verbraucht ([domain.Handelsangebot.mitZustand]); wer sie vorher
	// hineinlegte, ließe die erste Stimme wie eine Wiederholung aussehen.
	return angebot
}

// eineLeitungMitZusagen gibt sie über einer Ablage und einem Zusagenhalter.
func eineLeitungMitZusagen(
	angebot domain.Handelsangebot,
) (*service.Handelsleitung, *handelsprobe, *zusagenhaltprobe) {
	ablage := &handelsprobe{staende: map[string]domain.Handelsangebot{
		angebot.Vorgang: angebot,
	}}

	zusagen := &zusagenhaltprobe{}

	return service.NeueHandelsleitung(ablage, zusagen), ablage, zusagen
}

// **`vorbereiten` verlangt, dass festgehalten ist, was weggeht.**
//
// **Sonst hieße der Schritt etwas anderes als er tut.** Die Zustandsmatrix
// sagt „die Exemplare sind reserviert"; ohne diese Prüfung bewegte er den
// Zustand und sonst nichts — und die Verbuchung fände später nichts vor.
func TestVorbereitenVerlangtWasWeggeht(t *testing.T) {
	t.Parallel()

	leitung, ablage, _ := eineLeitungMitZusagen(
		einHandelMitAbgang(domain.HandelAngenommen))

	_, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittVorbereiten, "bich")
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **Ist genug festgehalten, geht er durch.**
func TestVorbereitenGehtMitGenugFestgehaltenem(t *testing.T) {
	t.Parallel()

	leitung, ablage, zusagen := eineLeitungMitZusagen(
		einHandelMitAbgang(domain.HandelAngenommen))

	zusagen.zusagen = []domain.Reservierung{
		{Vorgang: "v-abgang", Revision: 1, ExemplarID: 7, Menge: 2},
	}

	if _, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittVorbereiten, "bich"); fehler != nil {
		t.Fatalf("Schreiten: %v", fehler)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **Wer nichts weggibt, braucht nichts festzuhalten.**
//
// **Die Gegenprobe zur Zeile darüber**: ein Handel, bei dem alles von drüben
// kommt, hat null wegzugebende Stücke — und eine Bedingung, die auch dann
// zuschlüge, machte jeden einseitigen Handel unmöglich.
func TestWerNichtsWeggibtBrauchtNichtsFestzuhalten(t *testing.T) {
	t.Parallel()

	angebot := einHandelMitAbgang(domain.HandelAngenommen)
	angebot.Positionen[0].Von = "bdort"

	leitung, ablage, _ := eineLeitungMitZusagen(angebot)

	if _, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittVorbereiten, "bich"); fehler != nil {
		t.Fatalf("Schreiten: %v", fehler)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **`uebergeben` vermerkt die Zusagen als übergeben — wenn beide gehen.**
//
// **Das ist die Zeile, die der Verbuchung fehlte**: `abgaengeFuer` überspringt
// jede Zusage ohne diesen Vermerk, und es gab keine Anweisung, die ihn setzte.
//
// **Und erst der zweite Schritt vermerkt**, denn `uebergeben` ist ein
// **Beschluss** (`T1`): nach der ersten Stimme steht der Vorgang noch, wo er
// stand — und was nicht geschehen ist, wird nicht vermerkt.
func TestUebergebenVermerktDieZusagen(t *testing.T) {
	t.Parallel()

	leitung, _, zusagen := eineLeitungMitZusagen(
		einHandelMitAbgang(domain.HandelVorbereitet))

	zusagen.zusagen = []domain.Reservierung{
		{Vorgang: "v-abgang", Revision: 1, ExemplarID: 7, Menge: 2},
	}

	ctx := context.Background()

	if _, fehler := leitung.Schreiten(
		ctx, "v-abgang", domain.SchrittUebergeben, "bich"); fehler != nil {
		t.Fatalf("die erste Stimme: %v", fehler)
	}

	if zusagen.zusagen[0].Uebergeben {
		t.Fatal("eine Stimme hat schon vermerkt")
	}

	if _, fehler := leitung.Schreiten(
		ctx, "v-abgang", domain.SchrittUebergeben, "bdort"); fehler != nil {
		t.Fatalf("die zweite Stimme: %v", fehler)
	}

	if !zusagen.zusagen[0].Uebergeben {
		t.Error("die Zusage gilt nicht als übergeben")
	}
}

// **Ein Abbruch gibt sie wieder frei.**
func TestEinAbbruchGibtDieZusagenFrei(t *testing.T) {
	t.Parallel()

	leitung, _, zusagen := eineLeitungMitZusagen(
		einHandelMitAbgang(domain.HandelAngeboten))

	zusagen.zusagen = []domain.Reservierung{
		{Vorgang: "v-abgang", Revision: 1, ExemplarID: 7, Menge: 2},
	}

	if _, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittAbbrechen, "bich"); fehler != nil {
		t.Fatalf("Schreiten: %v", fehler)
	}

	if len(zusagen.zusagen) != 0 {
		t.Errorf("festgehalten sind noch %+v", zusagen.zusagen)
	}
}

// **Und was die Zusagen ablehnen, hält den Schritt auf — ohne Schreiben.**
func TestEineAbgelehnteZusagenwirkungSchreibtNicht(t *testing.T) {
	t.Parallel()

	leitung, ablage, zusagen := eineLeitungMitZusagen(
		einHandelMitAbgang(domain.HandelAngeboten))

	zusagen.klagt = errors.New("der Bestand klemmt")

	if _, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittAbbrechen, "bich"); fehler == nil {
		t.Fatal("die Absage kam nicht an")
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **Ein Bestand, der die Zusagen nicht hergibt, hält `vorbereiten` auf.**
//
// **Und nicht „dann eben ohne Deckung"**: „nichts festgehalten" und „ich
// weiß nicht, was festgehalten ist" sind zwei Dinge (§8), und die zweite
// Antwort darf keinen Schritt durchlassen, der Bestand bewegt.
func TestEinUnlesbarerZusagenbestandHaeltVorbereitenAuf(t *testing.T) {
	t.Parallel()

	leitung, ablage, zusagen := eineLeitungMitZusagen(
		einHandelMitAbgang(domain.HandelAngenommen))

	zusagen.klagt = errors.New("der Bestand klemmt")

	if _, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittVorbereiten, "bich"); fehler == nil {
		t.Fatal("der Schritt ging trotz unlesbarer Zusagen durch")
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **Auch wer nicht eröffnet hat, bestätigt die Übergabe erst mit Deckung** —
// gefunden mit zwei Fenstern (Hüllenprobe `nat04handel`, S12.NAT-04,
// #1672).
//
// **`vorbereiten` geht nur der Initiator** (Handelsmatrix), und nur seine
// Deckung wurde geprüft. Gab die Gegenseite etwas weg, bestätigte sie die
// Übergabe ohne ein festgehaltenes Exemplar — und das Verbuchen sagte
// danach „es fehlen 1 Stück an bestätigter Übergabe", aus einem Zustand,
// aus dem nur noch die Klärung führt.
func TestUebergebenVerlangtWasWeggeht(t *testing.T) {
	t.Parallel()

	angebot := einHandelMitAbgang(domain.HandelVorbereitet)
	angebot.Initiator = "bdort"

	leitung, ablage, _ := eineLeitungMitZusagen(angebot)

	_, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittUebergeben, "bich")
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **Ist genug festgehalten, stimmt sie zu.**
func TestUebergebenGehtMitGenugFestgehaltenem(t *testing.T) {
	t.Parallel()

	angebot := einHandelMitAbgang(domain.HandelVorbereitet)
	angebot.Initiator = "bdort"

	leitung, ablage, zusagen := eineLeitungMitZusagen(angebot)

	zusagen.zusagen = []domain.Reservierung{
		{Vorgang: "v-abgang", Revision: 1, ExemplarID: 7, Menge: 2},
	}

	if _, fehler := leitung.Schreiten(context.Background(), "v-abgang",
		domain.SchrittUebergeben, "bich"); fehler != nil {
		t.Fatalf("Schreiten: %v", fehler)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
	}
}

// **Die Stimme des Gegenübers prüft nicht diesen Bestand.** Wer sie hier
// einträgt (`HandelSchritt` mit `wer`), trägt eine Zusage ein, deren
// Exemplare drüben liegen — gezählt gegen die Zusagen **hier**, stünde sie
// immer bei null, und keine fremde Stimme käme je an.
func TestDieStimmeDesGegenuebersPrueftNichtDiesenBestand(t *testing.T) {
	t.Parallel()

	for _, schritt := range []domain.Handelsschritt{
		domain.SchrittVorbereiten, domain.SchrittUebergeben,
	} {
		t.Run(string(schritt), func(t *testing.T) {
			t.Parallel()

			zustand := domain.HandelVorbereitet
			if schritt == domain.SchrittVorbereiten {
				zustand = domain.HandelAngenommen
			}

			leitung, ablage, _ := eineLeitungMitZusagen(
				einHandelMitAbgang(zustand))

			if _, fehler := leitung.SchreitenAls(context.Background(),
				"v-abgang", "", schritt, "bich", "bdort"); fehler != nil {
				t.Fatalf("Schreiten: %v", fehler)
			}

			if ablage.geschrieben != 1 {
				t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
			}
		})
	}
}
