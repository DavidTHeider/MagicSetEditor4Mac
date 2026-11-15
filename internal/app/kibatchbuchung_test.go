package app_test

// **Was ein Batch bucht, wenn er scheitert** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Warum eine eigene Datei
//
// **Weil `kibatch_test.go` an der Zeilengrenze aus §2 stand.** Der Schnitt
// folgt der Frage: dort steht, **was ein Batch tut**, hier, **was er
// hinterlässt**, wenn er es nicht tut.
//
// # Und die Unterscheidung, um die es geht
//
// **Freigeben heißt „es ging nichts hinaus", unklar heißt „der Preis ist
// unbekannt"** (#1542). Ein Batch, dessen Werk gar nicht zustande kommt, hat
// nie gefragt — „unklar" wäre dort eine Zahl, die nie entstanden ist.
import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Batch an einen Anbieter, den es nicht gibt, geht nicht hinaus.**
//
// **Und er scheitert beim Buchen und nicht beim Fragen**: die Freigabe
// lautet auf einen anderen Anbieter, und ein Nachweis gilt seinem Ziel.
func TestEinBatchAnEinenUnbekanntenAnbieterGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), "gibtesnicht", derCloudalias,
		"Schreibe einen Regeltext.", welt.freigabe(t))
	if antwort.Fehler == nil {
		t.Fatal("der Batch an einen unbekannten Anbieter ging hinaus")
	}

	if welt.gefragt() != 0 {
		t.Errorf("es gingen %d Anfragen hinaus", welt.gefragt())
	}
}

// **Ein lokaler Batch bucht nichts — auch nicht, wenn er scheitert.**
//
// **Die Gegenprobe zum Freigeben** (§10): wo es keinen Posten gibt, darf
// niemand einen freigeben — die Wache bekäme eine Kennung, die sie nicht
// kennt, und die Meldung stünde an einem Auftrag, der nichts gekostet hat.
func TestEinLokalerBatchBuchtNichtsUndGibtNichtsFrei(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	// **Ein Alias, den der Katalog nicht führt** — damit das Werk zwar
	// steht, der Auftrag darin aber scheitert.
	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), "ollama", "gibtesnicht",
		"Schreibe einen Regeltext.", "")
	if antwort.Fehler != nil {
		t.Fatalf("der lokale Batch: %+v", antwort.Fehler)
	}

	for _, lage := range antwort.Ziele {
		if lage.Fehler == nil {
			t.Errorf("%d kam mit einem unbekannten Modell durch", lage.Zielkarte)
		}
	}

	posten := welt.anwendung.Kostenposten(probekontext(t), "ollama",
		string(domain.ZeitraumTag))
	if posten.Fehler != nil {
		t.Fatalf("Kostenposten: %+v", posten.Fehler)
	}

	if len(posten.Posten) != 0 {
		t.Fatalf("der lokale Batch hat gebucht: %+v", posten.Posten)
	}
}

// **Und ein Buchungsfehler beim Freigeben hält den Batch nicht auf.**
//
// **Dieselbe Zusage wie beim Schließen** (§7, Regel 3): der Batch ist
// gescheitert, und das ist die Meldung, die jemand lesen soll — nicht die
// über eine Buchung, die niemand mehr richten kann.
func TestEinFehlerBeimFreigebenHaeltDenBatchNichtAuf(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	freigabe := welt.freigabe(t)

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAus); fehler != nil {
		t.Fatalf("den Anbieter abschalten: %v", fehler)
	}

	welt.anwendung.MitKosten(sperrigeskijournal{})

	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), derCloudanbieter, derCloudalias,
		"Schreibe einen Regeltext.", freigabe)
	if antwort.Fehler == nil {
		t.Fatal("der Batch lief ohne eingeschalteten Anbieter")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "eingeschaltet") {
		t.Errorf("die Meldung ist %q — sie nennt den Grund nicht",
			antwort.Fehler.Meldung)
	}
}

// **Ein lokaler Batch ohne Laufzeit gibt nichts frei — es gibt nichts.**
//
// **Der Weg, auf dem `KostenpostenOhneKosten` ankommt**: lokal wird nicht
// gebucht, das Werk kommt nicht zustande, und die Wache darf trotzdem keine
// Kennung sehen, die sie nicht kennt.
func TestEinLokalerBatchOhneLaufzeitGibtNichtsFrei(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.aus(t)

	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), "ollama", "gemma4:8b",
		"Schreibe einen Regeltext.", "")
	if antwort.Fehler == nil {
		t.Fatal("der Batch lief ohne eingeschaltete Laufzeit")
	}

	posten := welt.anwendung.Kostenposten(probekontext(t), "ollama",
		string(domain.ZeitraumTag))
	if posten.Fehler != nil {
		t.Fatalf("Kostenposten: %+v", posten.Fehler)
	}

	if len(posten.Posten) != 0 {
		t.Fatalf("ohne Laufzeit wurde gebucht: %+v", posten.Posten)
	}
}
