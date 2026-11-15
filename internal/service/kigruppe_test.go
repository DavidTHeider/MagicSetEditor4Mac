package service_test

// **Die atomare Gruppe** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Der Satz, für den es diesen Lauf gibt
//
// > Gewählte Gruppe entweder vollständig mit Herkunft übernommen oder klarer
// > Konflikt **ohne Teilbestand**; kein Rückrollen bereits separat
// > bestätigter Gruppen.
//
// **Zwei Zusagen, und die zweite wird beim Bauen der ersten gebrochen.** Wer
// „alles oder nichts" als Schleife mit Rücknahme baut, nimmt beim Scheitern
// auch zurück, was eine **frühere** Gruppe geschrieben hat — und die war
// bestätigt. Deshalb wird hier **vorher** geprüft und **danach** in einem Zug
// geschrieben; zurückzunehmen gibt es dann nichts.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine vollständige Gruppe geht in einem Zug hinaus.**
func TestEineGruppeGehtInEinemZugHinaus(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	uebernahme, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"name"}},
		}, 7)
	if fehler != nil {
		t.Fatalf("die Gruppe scheiterte: %v", fehler)
	}

	if uebernahme.Karten != 2 || uebernahme.Felder != 2 {
		t.Errorf("übernommen wurden %+v", uebernahme)
	}

	// **Ein Ruf an die Ablage und nicht zwei.** Zwei wären zwei
	// Transaktionen, und dazwischen läge der Teilbestand, den das Issue
	// ausschließt.
	if hand.karten.gruppenrufe != 1 {
		t.Errorf("die Ablage bekam %d Rufe", hand.karten.gruppenrufe)
	}

	if len(hand.karten.gruppe) != 2 {
		t.Fatalf("die Gruppe trug %d Schritte", len(hand.karten.gruppe))
	}

	// **Und jeder Schritt trägt seine Herkunft.** Eine Übernahme ohne sie
	// wäre eine, die man später nicht mehr von getippten Werten
	// unterscheidet (A11, #1450).
	for _, schritt := range hand.karten.gruppe {
		if len(schritt.Herkunft) == 0 {
			t.Errorf("Schritt zu Karte %d ohne Herkunft", schritt.Karte)
		}
	}
}

// **Ein Fehler in einem Schritt lässt keinen halben Bestand zurück.**
//
// Der zweite Vorschlag trägt kein Feld `farbe`; die Auswahl nennt es
// trotzdem. **Geschrieben werden darf dann gar nichts** — auch nicht die
// erste Karte, deren Auswahl in Ordnung ist.
func TestEinFehlerInEinemSchrittSchreibtGarNichts(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	_, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"farbe"}},
		}, 7)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if hand.karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", hand.karten.gruppenrufe)
	}

	// **Und beide Aufträge stehen noch offen.** Ein verbrauchter Schlüssel
	// nach einem gescheiterten Schreiben nähme den zweiten Versuch weg.
	for _, kennung := range []string{hand.eins, hand.zwei} {
		if _, fehler := hand.werk.GruppeUebernehmen(context.Background(),
			[]service.KIGruppenwahl{{Kennung: kennung, Felder: []string{"name"}}}, 7,
		); fehler != nil {
			t.Errorf("%s ist nicht mehr offen: %v", kennung, fehler)
		}
	}
}

// **Ein veränderter Stand bricht die ganze Gruppe — als Konflikt.**
func TestEinVeraenderterStandBrichtDieGanzeGruppe(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)
	hand.karten.revision = "eine andere"

	_, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"name"}},
		}, 7)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if hand.karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", hand.karten.gruppenrufe)
	}
}

// **Eine leere Gruppe ist ein Fehler und kein Nichts.**
//
// **Sonst wäre sie die stillste Art, nichts zu tun** — dieselbe Überlegung
// wie bei der leeren Auswahl in [service.KIAuftragswerk.Uebernehmen].
func TestEineLeereGruppeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	if _, fehler := hand.werk.GruppeUebernehmen(
		context.Background(), nil, 7,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Dieselbe Karte zweimal in einer Gruppe ist ein Fehler.**
//
// **Weil die zweite die erste überschriebe** — in derselben Transaktion, ohne
// dass jemand es sähe. Welcher der beiden Werte am Ende stünde, hinge an der
// Reihenfolge der Auswahl.
func TestDieselbeKarteZweimalWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	_, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.eins, Felder: []string{"name"}},
		}, 7)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if hand.karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", hand.karten.gruppenrufe)
	}
}

// **Scheitert die Ablage, bleibt die Gruppe offen.**
//
// **Das ist der Unterschied zwischen „nichts geschrieben" und „Schlüssel
// weg"**: die Transaktion ist zurückgerollt, und ein zweiter Versuch muss
// möglich sein.
func TestEineGescheiterteAblageLaesstDieGruppeOffen(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)
	hand.karten.gruppenfehler = errors.New("die Platte ist voll")

	if _, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"name"}},
		}, 7); fehler == nil {
		t.Fatal("der Fehler der Ablage kam nicht durch")
	}

	hand.karten.gruppenfehler = nil

	if _, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"name"}},
		}, 7); fehler != nil {
		t.Errorf("der zweite Versuch scheiterte: %v", fehler)
	}
}

// gruppenkarten ist die Ablage, die eine **Gruppe** entgegennimmt.
//
// **Mit einem eigenen Zähler**: der Satz „ein Ruf und nicht zwei" ist die
// halbe Zusage der Atomarität, und an `schreibvorgaenge` der Einzelprobe wäre
// er nicht abzulesen — der zählt Rufe an `FelderSetzen`, und die gibt es hier
// gerade **nicht** geben.
type gruppenkarten struct {
	kiKartenprobe

	gruppenrufe   int
	gruppe        []service.KIGruppenschritt
	gruppenfehler error
}

func (g *gruppenkarten) FelderGruppeSetzen(
	_ context.Context, schritte []service.KIGruppenschritt,
) error {
	if g.gruppenfehler != nil {
		return g.gruppenfehler
	}

	g.gruppenrufe++
	g.gruppe = schritte

	return nil
}

// zweiVorschlaege ist die Vorrichtung: ein Werk, seine Ablage und die beiden
// **wirklich vergebenen** Kennungen.
//
// **Die Kennungen kommen von [service.KIAuftragswerk.Anlegen]** und werden
// nicht erfunden — ein Lauf gegen erdachte prüfte den Weg „den Auftrag gibt
// es nicht" und hielte sich für einen über Gruppen.
type zweiVorschlaege struct {
	werk   *service.KIAuftragswerk
	karten *gruppenkarten
	eins   string
	zwei   string
}

// einWerkMitZweiVorschlaegen legt zwei Aufträge an und beantwortet sie.
//
// **Zwei und nicht drei**: die Zusagen hier sind über „alles oder nichts",
// und dafür genügt ein Paar. Ein drittes prüfte die Schleife.
func einWerkMitZweiVorschlaegen(t *testing.T) zweiVorschlaege {
	t.Helper()

	karten := &gruppenkarten{kiKartenprobe: kiKartenprobe{revision: "stand-abc"}}
	werk := einWerkMit(t, karten)

	kennungen := make([]string, 0, 2)

	for _, karte := range []int64{42, 43} {
		wunsch := einWunsch()
		wunsch.Zielkarte = karte

		auftrag, fehler := werk.Anlegen(t.Context(), wunsch)
		if fehler != nil {
			t.Fatalf("den Auftrag anlegen: %v", fehler)
		}

		if _, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
			Auftrag: auftrag.Kennung,
			Felder:  map[string][]string{"name": {"Blitzschlag"}},
		}); fehler != nil {
			t.Fatalf("den Vorschlag ablegen: %v", fehler)
		}

		kennungen = append(kennungen, auftrag.Kennung)
	}

	return zweiVorschlaege{
		werk: werk, karten: karten,
		eins: kennungen[0], zwei: kennungen[1],
	}
}

// **Einen Auftrag, den es nicht gibt, weist die Gruppe ab.**
func TestEinenUnbekanntenAuftragWeistDieGruppeAb(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	if _, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: "gibtesnicht", Felder: []string{"name"}},
		}, 7); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if hand.karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", hand.karten.gruppenrufe)
	}
}

// **Ein Auftrag ohne Vorschlag ebenso.**
//
// **Und das ist ein anderer Fall als der darüber**: den Auftrag gibt es, nur
// hat noch kein Modell geantwortet. Wer beide zusammenwürfe, sagte einem
// Menschen „gibt es nicht" über etwas, das er gerade angelegt hat.
func TestEinAuftragOhneVorschlagWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	karten := &gruppenkarten{kiKartenprobe: kiKartenprobe{revision: "stand-abc"}}
	werk := einWerkMit(t, karten)

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("den Auftrag anlegen: %v", fehler)
	}

	if _, fehler := werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: auftrag.Kennung, Felder: []string{"name"}},
		}, 7); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", karten.gruppenrufe)
	}
}

// **Eine abgelaufene Frist bricht die Gruppe — als Konflikt.**
//
// **Gemessen an der Uhr des Werks**: [einWerk] gibt ihr einen festen
// Zeitpunkt, und die Frist hängt daran. Ein Lauf, der wartete, prüfte die
// Geduld des Rechners.
func TestEineAbgelaufeneFristBrichtDieGruppe(t *testing.T) {
	t.Parallel()

	karten := &gruppenkarten{kiKartenprobe: kiKartenprobe{revision: "stand-abc"}}

	// **Eine Uhr, die springt**: beim Anlegen steht sie auf dem Lauf, beim
	// Übernehmen weit dahinter.
	jetzt := derLauf
	werk := service.NeuesKIAuftragswerk(
		derKatalog(t), karten, &kostenprobe{}, &schemaprobe{},
		service.KIOrtsregel{}, func() time.Time { return jetzt })

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("den Auftrag anlegen: %v", fehler)
	}

	if _, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung,
		Felder:  map[string][]string{"name": {"Blitzschlag"}},
	}); fehler != nil {
		t.Fatalf("den Vorschlag ablegen: %v", fehler)
	}

	jetzt = derLauf.Add(service.KIFrist + time.Minute)

	if _, fehler := werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: auftrag.Kennung, Felder: []string{"name"}},
		}, 7); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", karten.gruppenrufe)
	}
}

// **Und die Gruppe schickt ihn für jede Karte mit** — Befund 2 des
// Zweitblicks über Phase 5.3.
//
// **Für jede und nicht für die erste**: eine Gruppe, die nur den Stand ihres
// ersten Schritts hielte, ließe jede weitere Karte ungeschützt — und die
// Zusage der Gruppe lautet „ganz oder gar nicht".
func TestDieGruppeSchicktJedenStandMit(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	if _, fehler := hand.werk.GruppeUebernehmen(context.Background(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"name"}},
		}, 7); fehler != nil {
		t.Fatalf("GruppeUebernehmen: %v", fehler)
	}

	if len(hand.karten.gruppe) != 2 {
		t.Fatalf("an der Ablage kamen %d Schritte an", len(hand.karten.gruppe))
	}

	for _, schritt := range hand.karten.gruppe {
		if schritt.Erwartet != "stand-abc" {
			t.Errorf("der Schritt für Karte %d trug %q",
				schritt.Karte, schritt.Erwartet)
		}
	}
}
