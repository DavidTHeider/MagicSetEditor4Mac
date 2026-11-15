package app_test

// **Ein KI-Auftrag lässt seine Reservierung nicht offen** —
// #1542.
//
// # Der Befund
//
// **Nach einer Messung am laufenden Fenster stand im Journal:**
//
//	('probe/openai/gpt-image-langsam', 'openai', 'reserviert', 0)
//
// Für immer. `service.KIBudget` hatte **genau eine** Methode — `Reservieren` —,
// und keine der drei Abschlussarten der Kostenwache wurde je gerufen. Die
// Reservierung band Budget, obwohl der Auftrag längst zurück war: eine Grenze,
// die nur noch steigt, bis kein Auftrag mehr hinausgeht.
//
// # Warum `unklar` und nicht `abgeschlossen`
//
// **Weil `Abschliessen(…, 0)` „war kostenlos" hieße, und das ist falsch.** Der
// Anbieter hat geantwortet und rechnet ab; was der Auftrag gekostet hat, sagt
// er an dieser Stelle nicht, und dieses Programm führt keine Preisliste. O12
// hat den Stand genau dafür: „zählt wie ausgegeben, bis jemand nachgesehen
// hat" (#1439).
//
// # Warum die Läufe daneben es nicht sahen
//
// **Weil sie ein Doppel am Journal einsetzen** (§3): `stummeskijournal` bucht
// nichts, und „nichts steht offen" ist über ein leeres Buch keine Aussage.
// Hier läuft das **echte** Journal aus [einekiwelt].

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Nach einem Vorschlag steht keine Reservierung mehr offen.**
func TestNachEinemVorschlagBleibtKeineReservierungOffen(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	freigabe := welt.freigabe(t)

	if antwort := welt.vorschlagenMitFreigabe(t,
		freigabe); antwort.Fehler != nil {
		t.Fatalf("der Vorschlag: %+v", antwort.Fehler)
	}

	welt.pruefeOhneReservierung(t)
}

// **Und nach einer Messung ebenso wenig.**
//
// **Derselbe Weg, ein anderer Aufrufer** (§10: die Klasse, nicht der Fall):
// die Messung bucht neben dem Auftragswerk, und ein Abschluss nur dort ließe
// sie offen.
func TestNachEinerMessungBleibtKeineReservierungOffen(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	// **Die Messung erwartet ihr eigenes Schema** — `{"probe": …}` und nicht
	// die Felder einer Karte.
	welt.fremd.antwort = app.KIRohantwort{
		Inhalt: `{"probe": "bereit"}`, Modell: derCloudalias,
	}

	if antwort := welt.anwendung.KIModellPruefen(probekontext(t),
		derCloudanbieter, derCloudalias,
		welt.freigabe(t)); antwort.Fehler != nil {
		t.Fatalf("KIModellPruefen: %+v", antwort.Fehler)
	}

	welt.pruefeOhneReservierung(t)
}

// **Ein lokaler Auftrag bucht weiterhin gar nichts.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): ein Abschluss, der auch ohne
// Buchung liefe, riefe die Wache für einen Posten, den es nicht gibt — und
// die Meldung daraus stünde an einem Auftrag, der nichts gekostet hat.
func TestEinLokalerAuftragBuchtWeiterhinNichts(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe einen Regeltext.", "")
	if antwort.Fehler != nil {
		t.Fatalf("der lokale Auftrag: %+v", antwort.Fehler)
	}

	posten := welt.anwendung.Kostenposten(probekontext(t), "ollama",
		string(domain.ZeitraumTag))
	if posten.Fehler != nil {
		t.Fatalf("Kostenposten: %+v", posten.Fehler)
	}

	if len(posten.Posten) != 0 {
		t.Fatalf("der lokale Auftrag hat gebucht: %+v", posten.Posten)
	}
}

// freigabe erteilt eine und gibt ihre Kennung.
func (c cloudwelt) freigabe(t *testing.T) string {
	t.Helper()

	erteilt := c.anwendung.KIFreigabeErteilen(probekontext(t),
		derCloudanbieter, 500, dieWaehrung)
	if erteilt.Fehler != nil {
		t.Fatalf("KIFreigabeErteilen: %+v", erteilt.Fehler)
	}

	return erteilt.Kennung
}

// pruefeOhneReservierung hält, dass kein Posten mehr reserviert ist.
func (c cloudwelt) pruefeOhneReservierung(t *testing.T) {
	t.Helper()

	posten := c.anwendung.Kostenposten(probekontext(t), derCloudanbieter,
		string(domain.ZeitraumTag))
	if posten.Fehler != nil {
		t.Fatalf("Kostenposten: %+v", posten.Fehler)
	}

	// **Erst der Beleg, dass überhaupt gebucht wurde** — ohne ihn träfe
	// „nichts steht offen" auch auf ein leeres Buch zu (§3).
	if len(posten.Posten) == 0 {
		t.Fatal("es steht kein Posten da — dann bucht der Auftrag gar nicht")
	}

	for _, einer := range posten.Posten {
		if einer.Stand == string(domain.KostenReserviert) {
			t.Fatalf("%q steht auf %q — die Reservierung bindet weiter Budget",
				einer.Kennung, einer.Stand)
		}
	}
}

// sperrigeskijournal bucht, lässt sich aber nicht schließen.
type sperrigeskijournal struct{ stummeskijournal }

func (sperrigeskijournal) AlsUnklarMerken(context.Context, string, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, errors.New("die Platte ist voll")
}

func (sperrigeskijournal) Freigeben(context.Context, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, errors.New("die Platte ist voll")
}

// **Ein Buchungsfehler beim Schließen nimmt den Vorschlag nicht zurück.**
//
// # Warum das so und nicht andersherum
//
// **Weil der Auftrag geschehen ist** (§7, Regel 3). Das Modell hat geantwortet,
// der Vorschlag liegt vor — wer die Meldung über eine misslungene Buchung an
// die Antwort hängte, machte daraus einen gescheiterten Vorschlag, und zwar
// wegen etwas, das niemand mehr rückgängig machen kann. Der Fehler geht ins
// Protokoll und nicht an das Fenster.
func TestEinBuchungsfehlerNimmtDenVorschlagNichtZurueck(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	// **Die Freigabe zuerst, dann das sperrige Journal** — jene geht durch
	// das echte, und ein Doppel davor hätte keinen Nachweis auszustellen.
	freigabe := welt.freigabe(t)

	welt.anwendung.MitKosten(sperrigeskijournal{})

	antwort := welt.vorschlagenMitFreigabe(t, freigabe)
	if antwort.Fehler != nil {
		t.Fatalf("der Vorschlag: %+v", antwort.Fehler)
	}

	if len(antwort.Felder) == 0 {
		t.Fatal("der Vorschlag kam ohne Felder zurück")
	}
}
