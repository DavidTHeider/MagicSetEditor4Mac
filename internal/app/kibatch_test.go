package app_test

// **Ein Batch von Vorschlägen** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Was hier geprüft wird
//
// **Dass ein Batch eine Menge ist und keine Schleife.** Er bucht **einmal**
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)),
// er schreibt **nichts**, und er nennt je Ziel, was daraus geworden ist —
// auch dort, wo es scheiterte.

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dieBatchziele sind drei Karten.
func dieBatchziele() []int64 { return []int64{4, 5, 6} }

// **Ein Batch schlägt für jedes Ziel vor — und schreibt nichts.**
func TestEinBatchSchlaegtFuerJedesZielVor(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	antwort := welt.batch(t, dieBatchziele())
	if antwort.Fehler != nil {
		t.Fatalf("der Batch: %+v", antwort.Fehler)
	}

	if len(antwort.Ziele) != len(dieBatchziele()) {
		t.Fatalf("es kamen %d Lagen zurück: %+v", len(antwort.Ziele), antwort.Ziele)
	}

	for _, lage := range antwort.Ziele {
		if lage.Kennung == "" || len(lage.Felder) == 0 {
			t.Errorf("die Lage von %d ist %+v", lage.Zielkarte, lage)
		}
	}

	// **Und die Reihenfolge ist die der Anfrage.** Wer zwölf Karten schickt,
	// liest die Antwort neben seiner Auswahl; eine Liste, die sich nach der
	// Antwortzeit ordnet, wäre bei jedem Lauf eine andere.
	for i, lage := range antwort.Ziele {
		if lage.Zielkarte != dieBatchziele()[i] {
			t.Fatalf("an Stelle %d steht %d", i, lage.Zielkarte)
		}
	}

	if welt.geschrieben() != 0 {
		t.Errorf("der Batch hat %d Felder geschrieben", welt.geschrieben())
	}
}

// **Er bucht einmal und nicht je Ziel.**
//
// **Das ist 0105**: ein Mensch hat **einen** Betrag für **eine** abgegrenzte
// Menge Arbeit freigegeben. Drei Buchungen wären drei Freigaben, von denen er
// zwei nie gesehen hat.
func TestEinBatchBuchtEinmalUndNichtJeZiel(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	if antwort := welt.batch(t, dieBatchziele()); antwort.Fehler != nil {
		t.Fatalf("der Batch: %+v", antwort.Fehler)
	}

	posten := welt.anwendung.Kostenposten(probekontext(t), derCloudanbieter,
		string(domain.ZeitraumTag))
	if posten.Fehler != nil {
		t.Fatalf("Kostenposten: %+v", posten.Fehler)
	}

	if len(posten.Posten) != 1 {
		t.Fatalf("es stehen %d Posten da: %+v", len(posten.Posten), posten.Posten)
	}
}

// **Ein Ziel, das scheitert, hält den Batch nicht auf.**
//
// **Sonst wäre ein Batch eine Kette**: eine Karte, die es nicht mehr gibt,
// nähme elf anderen ihren Vorschlag weg — und die Freigabe wäre verbraucht.
func TestEinScheiterndesZielHaeltDenBatchNichtAuf(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)
	welt.karten.fehler = errors.New("die Karte 5 gibt es nicht")

	antwort := welt.batch(t, dieBatchziele())
	if antwort.Fehler != nil {
		t.Fatalf("der Batch: %+v", antwort.Fehler)
	}

	for _, lage := range antwort.Ziele {
		if lage.Fehler == nil {
			t.Errorf("%d kam durch, obwohl der Bestand klagt", lage.Zielkarte)
		}
	}
}

// **Ein Batch über der Grenze geht gar nicht erst hinaus.**
//
// **Vor dem ersten Ruf und nicht nach dem vierundzwanzigsten** (§16, Regel 4):
// eine Grenze, die erst beim Buchen greift, hat schon bezahlt.
func TestEinBatchUeberDerGrenzeGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	ziele := make([]int64, domain.KIBatchHoechstzahl+1)
	for i := range ziele {
		ziele[i] = int64(i + 1)
	}

	antwort := welt.batch(t, ziele)
	if antwort.Fehler == nil {
		t.Fatal("der zu große Batch ging hinaus")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "24") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}

	if welt.gefragt() != 0 {
		t.Errorf("es gingen %d Anfragen hinaus", welt.gefragt())
	}
}

// **Und ohne Freigabe geht auch ein Batch nicht hinaus.**
func TestOhneFreigabeGehtAuchEinBatchNichtHinaus(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), derCloudanbieter, derCloudalias,
		"Schreibe einen Regeltext.", "")
	if antwort.Fehler == nil {
		t.Fatal("der Batch ging ohne Freigabe hinaus")
	}

	if welt.gefragt() != 0 {
		t.Errorf("es gingen %d Anfragen hinaus", welt.gefragt())
	}
}

// batch fährt einen — mit einer frisch erteilten Freigabe.
func (c cloudwelt) batch(t *testing.T, ziele []int64) app.KIBatchAntwort {
	t.Helper()

	return c.anwendung.KIBatchVorschlagen(probekontext(t), ziele,
		derCloudanbieter, derCloudalias, "Schreibe einen Regeltext.",
		c.freigabe(t))
}

// gefragt zählt die Rufe an die Gegenstelle.
func (c cloudwelt) gefragt() int { return len(c.fremd.adressen) }

// geschrieben zählt die Felder, die an einer Karte gelandet sind.
func (c cloudwelt) geschrieben() int { return c.karten.geschrieben }

// **Die Ziele laufen nicht alle auf einmal hinaus.**
//
// # Warum das gemessen wird und nicht behauptet
//
// **Weil eine Schranke, die niemand zählt, keine ist** (§10). Gemessen wird
// der **gleichzeitige** Höchststand an der Gegenstelle, nicht die Summe —
// die sagt über Parallelität nichts.
func TestEinBatchHaeltDieParallelitaet(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	zaehler := &gleichzeitig{}
	welt.fremd.vorZaehler = zaehler

	ziele := make([]int64, domain.KIBatchHoechstzahl)
	for i := range ziele {
		ziele[i] = int64(i + 1)
	}

	if antwort := welt.batch(t, ziele); antwort.Fehler != nil {
		t.Fatalf("der Batch: %+v", antwort.Fehler)
	}

	if hoechst := zaehler.hoechststand(); hoechst > domain.KIBatchParallel {
		t.Errorf("gleichzeitig liefen %d, erlaubt sind %d",
			hoechst, domain.KIBatchParallel)
	}
}

// **Und die Ziele stehen in der Reihenfolge der Anfrage, nicht der Antwort.**
func TestDieLagenStehenInDerReihenfolgeDerAnfrage(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	ziele := []int64{9, 3, 7, 1}

	antwort := welt.batch(t, ziele)
	if antwort.Fehler != nil {
		t.Fatalf("der Batch: %+v", antwort.Fehler)
	}

	gesehen := make([]int64, 0, len(antwort.Ziele))
	for _, lage := range antwort.Ziele {
		gesehen = append(gesehen, lage.Zielkarte)
	}

	if !slices.Equal(gesehen, ziele) {
		t.Fatalf("die Reihenfolge ist %v, erwartet %v", gesehen, ziele)
	}
}

// **Ein Batch, dessen Anbieter zwischendurch ausgeht, gibt seinen Posten
// frei.**
//
// # Warum freigeben und nicht als unklar merken
//
// **Weil belegt ist, dass nichts hinausging** (#1542): das Auftragswerk kam
// gar nicht zustande, also wurde nie gefragt. „Unklar" wäre eine Zahl, die
// nie entstanden ist — und sie bände Budget, bis ein Mensch nachsieht.
func TestEinBatchOhneWerkGibtSeinenPostenFrei(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	// **Die Freigabe zuerst** — sie lautet auf diesen Anbieter, und ohne sie
	// scheiterte der Batch schon beim Buchen.
	freigabe := welt.freigabe(t)

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAus); fehler != nil {
		t.Fatalf("den Anbieter abschalten: %v", fehler)
	}

	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), derCloudanbieter, derCloudalias,
		"Schreibe einen Regeltext.", freigabe)
	if antwort.Fehler == nil {
		t.Fatal("der Batch lief ohne eingeschalteten Anbieter")
	}

	posten := welt.anwendung.Kostenposten(probekontext(t), derCloudanbieter,
		string(domain.ZeitraumTag))
	if posten.Fehler != nil {
		t.Fatalf("Kostenposten: %+v", posten.Fehler)
	}

	// **Erst der Beleg, dass gebucht wurde** — sonst träfe „nichts ist
	// reserviert" auch auf ein leeres Buch zu (§3).
	if len(posten.Posten) != 1 {
		t.Fatalf("es stehen %d Posten da: %+v", len(posten.Posten), posten.Posten)
	}

	if stand := posten.Posten[0].Stand; stand != string(domain.KostenFreigegeben) {
		t.Errorf("der Posten steht auf %q, erwartet %q",
			stand, domain.KostenFreigegeben)
	}

	if welt.gefragt() != 0 {
		t.Errorf("es gingen %d Anfragen hinaus", welt.gefragt())
	}
}

// **Und eine erfundene Freigabe trägt auch keinen Batch.**
func TestEineErfundeneFreigabeTraegtKeinenBatch(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	antwort := welt.anwendung.KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), derCloudanbieter, derCloudalias,
		"Schreibe einen Regeltext.", "gibtesnicht")
	if antwort.Fehler == nil {
		t.Fatal("die erfundene Freigabe trug einen Batch")
	}

	if welt.gefragt() != 0 {
		t.Errorf("es gingen %d Anfragen hinaus", welt.gefragt())
	}
}

// **Ohne Bestand gibt es keine Adresse — und keinen Batch.**
func TestOhneBestandGibtEsKeinenBatch(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KIBatchVorschlagen(probekontext(t),
		dieBatchziele(), derCloudanbieter, derCloudalias,
		"Schreibe einen Regeltext.", "gleichwelche")
	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Batch zustande")
	}
}

// **Ein Batch ohne Profil wird abgewiesen — vor dem Buchen.**
//
// **Ohne Profil lässt sich nicht sagen, wessen Geld gemeint ist** — und eine
// Buchung, die keinem gehört, ließe sich später keinem zuordnen.
func TestEinBatchOhneProfilWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIBatchVorschlagen(t.Context(),
		dieBatchziele(), "ollama", "gemma4:8b", "x", "")
	if antwort.Fehler == nil {
		t.Fatal("der Batch ohne Profil kam durch")
	}

	if antwort.Kennung != "" {
		t.Errorf("er trägt schon eine Kennung: %q", antwort.Kennung)
	}
}
