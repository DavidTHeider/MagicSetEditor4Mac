package app_test

// **Was ein Auftrag kosten darf** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Eigene Datei
//
// `kianthropic_test.go` stand mit diesen Läufen an der Zeilengrenze aus §2 —
// und der Schnitt folgt der Frage: dort, **wer** antwortet, hier, **was es
// kosten darf**.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein erschöpftes Budget hält den Auftrag auf.**
//
// # Warum das neben der fehlenden Freigabe steht
//
// **Weil es ein anderer Weg ist.** Das Issue nennt drei: „Unter Local-only,
// fehlendem Scope oder **Budget** entstehen null Anbieteranfragen." Die
// ersten beiden weist `kibudget` selbst ab; das Budget weist die
// **Kostenwache** ab, und ihr Nein muss durchkommen, statt unterwegs zu
// einer leeren Antwort zu werden.
func TestEinErschoepftesBudgetHaeltDenAuftragAuf(t *testing.T) {
	t.Parallel()

	fremd := &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
		antwort: app.KIRohantwort{Inhalt: `{"name": "x"}`, Modell: "claude-opus-5"},
	}

	welt := einekiwelt(t)
	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"anthropic": func(string) app.KILaufzeitquelle { return fremd },
		})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	welt.anwendung.MitKosten(volleskijournal{})
	welt.anwendung.MitKIFreigabe(
		"der Mensch hat zugestimmt", domain.Betrag{Wert: 100, Waehrung: "EUR"})

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"anthropic", "claude-opus-5", "Schreibe.", "")
	if antwort.Fehler == nil {
		t.Fatal("der Auftrag ging trotz erschöpftem Budget hinaus")
	}

	// **Die Gegenprobe zur Zusage**: „null Anbieteranfragen" heißt null.
	if len(fremd.schemen) != 0 {
		t.Errorf("Anthropic bekam %d Fragen", len(fremd.schemen))
	}
}

// volleskijournal weist jede Reservierung ab — das Budget ist erschöpft.
type volleskijournal struct{ stummeskijournal }

func (volleskijournal) Reservieren(
	_ context.Context, _ service.Kostenauftrag, _ string,
) (domain.Kostenposten, error) {
	return domain.Kostenposten{}, fmt.Errorf(
		"%w: der Haushalt für diesen Zeitraum ist ausgeschöpft", domain.ErrConflict)
}

// **Das Journal des Fensters ist wirklich verdrahtet.**
//
// # Warum dieser Lauf eigens steht
//
// **Weil jeder andere es ersetzt.** Die Läufe darüber setzen ein Doppel ein,
// damit sie über Reservieren und Nachweis sprechen können — und dabei bleibt
// die **echte** Verdrahtung ungeprüft: [app.KIAnwendungUeber] baut eine
// [service.Kostenwache] über den offenen Bestand, und ob der Bestand dort
// wirklich ankommt, sagt keiner von ihnen.
//
// **Geprüft wird an der Wirkung**: der Auftrag geht durch, und danach steht
// ein Posten im Journal dieses Bestands. Ein Lauf, der nur „kein Fehler"
// prüfte, hinge an derselben Stelle wie ein Doppel.
func TestDasJournalDesFenstersIstVerdrahtet(t *testing.T) {
	t.Parallel()

	fremd := &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
		antwort: app.KIRohantwort{
			Inhalt: `{"name": "Blitzschlag"}`, Modell: "claude-opus-5",
		},
	}

	welt := einekiwelt(t)
	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"anthropic": func(string) app.KILaufzeitquelle { return fremd },
		})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	if antwort := welt.anwendung.KostenbudgetSetzen(probekontext(t),
		"anthropic", "tag", 50_000, "EUR"); antwort.Fehler != nil {
		t.Fatalf("das Budget setzen: %+v", antwort.Fehler)
	}

	// **Ein echter Nachweis und kein erfundener.** Die Wache schlägt ihn im
	// Bestand nach; ein Wort, das dort nicht steht, ist keine Freigabe — und
	// genau das hat der erste Anlauf dieses Laufs angestrichen.
	jetzt := time.Now()

	if fehler := storage.NeueNachweisablage(welt.db).NachweisAnlegen(
		probekontext(t), domain.Nachweis{
			Kennung: "freigabe", NutzerID: 1, ProfilID: 1,
			Zweck: service.KostenjournalZweck, Ziel: "anthropic",
			Grenze:      domain.Kostengrenze{Betrag: 30_000, Waehrung: "EUR"},
			Ausgestellt: jetzt.Add(-time.Minute), Ablauf: jetzt.Add(time.Hour),
		}); fehler != nil {
		t.Fatalf("den Nachweis legen: %v", fehler)
	}

	// **Kein `MitKosten` hier** — das ist der Punkt: es bleibt bei dem, was
	// `KIAnwendungUeber` verdrahtet hat.
	welt.anwendung.MitKIFreigabe(
		"freigabe", domain.Betrag{Wert: 20_000, Waehrung: "EUR"})

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"anthropic", "claude-opus-5", "Schreibe.", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	posten := welt.anwendung.Kostenposten(probekontext(t), "anthropic", "")
	if posten.Fehler != nil {
		t.Fatalf("%+v", posten.Fehler)
	}

	if len(posten.Posten) != 1 {
		t.Fatalf("im Journal stehen %d Posten", len(posten.Posten))
	}

	if posten.Posten[0].Aufgabe != app.KIAufgabe {
		t.Errorf("der Posten ist %+v", posten.Posten[0])
	}
}
