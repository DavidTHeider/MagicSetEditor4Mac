package service_test

// **Der Lebenszyklus einer Peersitzung** — Werkstück N1 von Phase 5.4
// (#1454),
// über eine echte Leitung seit N2
// (#1455).
//
// # Was hier gemessen wird
//
// Das Issue zählt es auf: „Verbindungs-Lifecycle, Fristen, Abbruch, begrenzte
// Warteschlangen, Backpressure und Berechtigungsentzug festhalten. Bei
// Nutzer-/Profilwechsel Sessions pausieren und Schlüsselkontext schließen,
// verspätete Antworten verwerfen."
//
// # Und warum diese Läufe seit N2 anders aussehen
//
// **Sie liefen gegen ein Doppel, das nur senden konnte.** Seit `Oeffne` den
// Handschlag führt, kommt eine Sitzung nur zustande, wenn am anderen Ende
// wirklich jemand antwortet — und ein Doppel, das das nachspielte, prüfte eine
// Sitzung, die es so nie gibt. Sie laufen jetzt über zwei echte Identitäten
// und eine echte Rahmung; nicht echt sind nur Rechte und Uhr.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// rechteprobe sagt ja oder nein.
type rechteprobe struct {
	darf   bool
	fehler error

	// dazwischen läuft **während** der Rechtefrage — für den Wechsel, der
	// genau dann stattfindet.
	dazwischen func()
}

func (r *rechteprobe) Darf(
	context.Context, peer.Kennung, domain.Zweck,
) (bool, error) {
	if r.dazwischen != nil {
		r.dazwischen()
	}

	return r.darf, r.fehler
}

// **Ohne Freigabe wird gar nicht erst gewählt.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Die Rechteprüfung steht vor dem Wählen und nicht danach.** Wer erst
// verbindet und dann fragt, hat dem Gegenüber schon gesagt, dass es dieses
// Programm gibt — und bei einem kostenpflichtigen Transport auch bezahlt.
func TestOhneFreigabeWirdGarNichtErstGewaehlt(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	paar.hiesigeRechte.darf = false

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckHandel,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Sitzung kam durch: %v", fehler)
	}

	if paar.hiesigerWaehler.gerufen != 0 {
		t.Fatalf("gewählt wurde %dmal", paar.hiesigerWaehler.gerufen)
	}
}

// **Eine offene Sitzung steht in der Übersicht — auf beiden Seiten.**
func TestEineOffeneSitzungStehtInDerUebersicht(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckTeilen)

	staende := paar.hier.Alle()
	if len(staende) != 1 {
		t.Fatalf("es stehen %d darin", len(staende))
	}

	if staende[0].Gegenueber != paar.dortige.Kennung() ||
		staende[0].Zweck != domain.ZweckTeilen {
		t.Errorf("der Stand ist %+v", staende[0])
	}

	// **Der Abdruck steht dabei** — eine Fläche, die ihn selbst rechnen
	// müsste, kann ihn vergessen.
	if staende[0].Fingerabdruck != peer.Fingerabdruck(paar.dortige.Kennung()) {
		t.Errorf("der Abdruck ist %q", staende[0].Fingerabdruck)
	}

	// **Und der Kanal kommt aus der Leitung**, nicht aus einer Vermutung.
	if staende[0].Kanal != paar.hiesigeLeitung.Kanal() {
		t.Errorf("der Kanal ist %q", staende[0].Kanal)
	}

	if fehler := griff.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if len(paar.hier.Alle()) != 0 {
		t.Fatal("die geschlossene Sitzung steht noch da")
	}
}

// **Ein zweiter Ruf für dasselbe Gegenüber und denselben Zweck ist ein
// Konflikt.**
//
// **Ein Vorgang je Datensatz** (§7, Regel 4): zwei Sitzungen auf dieselbe
// Sache gingen gegen einen Stand, den die andere Seite noch nicht kennt.
//
// **Und er endet vor der Leitung** — sonst stünde ein zweiter Handschlag auf
// einem Draht, auf dem schon einer liegt.
func TestEineZweiteSitzungAufDieselbeSacheIstEinKonflikt(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	paar.eineSitzung(t, domain.ZweckTeilen)

	vorher := paar.hiesigerWaehler.gerufen

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die zweite Sitzung kam durch: %v", fehler)
	}

	if paar.hiesigerWaehler.gerufen != vorher {
		t.Fatal("für die zweite Sitzung wurde gewählt")
	}
}

// **Eine zuverlässige Nachricht geht hinaus.**
func TestEineZuverlaessigeNachrichtGehtHinaus(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckTeilen)

	if fehler := griff.Sende(
		t.Context(), []byte("hallo"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	angekommen := make(chan []byte, 1)

	go func() {
		gelesen, fehler := paar.dortigeLeitung.Empfange(t.Context())
		if fehler != nil {
			t.Errorf("Empfange: %v", fehler)
		}

		angekommen <- gelesen
	}()

	if fehler := griff.Leere(t.Context()); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}

	if gelesen := <-angekommen; string(gelesen) != "hallo" {
		t.Fatalf("angekommen ist %q", gelesen)
	}
}

// **Eine volle Warteschlange staut zurück — und wirft Verwerfbares weg.**
//
// # Die zwei Zustellungen sind der Punkt
//
// **Zuverlässig heißt: der Absender erfährt es.** Bei voller Schlange wird
// abgewiesen, statt zu wachsen — das ist der Rückstau. **Verwerfbar heißt:
// der neuere Stand ersetzt den älteren** — derselben Art; eine
// Präsenznachricht, die eine Leitung aufhielte, wäre schlimmer als keine.
//
// **Hier verdrängte sie bis zum Nachlauf über Phase 6.2 die älteste
// Nachricht gleich welcher Art** — und dieser Lauf schrieb es fest: „eins"
// war zuverlässig angenommen und fiel still weg. Jetzt weicht die neue
// verwerfbare, weil nichts Verwerfbares ansteht (`sitzungtiefe_test.go`).
func TestEineVolleWarteschlangeStautZurueckUndWirftWeg(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckPartie)

	// **Die Tiefe ist zwei** — zwei passen hinein.
	for _, was := range []string{"eins", "zwei"} {
		if fehler := griff.Sende(
			t.Context(), []byte(was), service.Zuverlaessig); fehler != nil {
			t.Fatalf("Sende %q: %v", was, fehler)
		}
	}

	if fehler := griff.Sende(
		t.Context(), []byte("drei"), service.Zuverlaessig,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die dritte kam durch: %v", fehler)
	}

	// **Eine verwerfbare wird angenommen — und weicht selbst**: die zwei
	// davor sind zuverlässig zugesagt.
	if fehler := griff.Sende(
		t.Context(), []byte("praesenz"), service.Verwerfbar); fehler != nil {
		t.Fatalf("Sende verwerfbar: %v", fehler)
	}

	angekommen := make(chan string, 2)

	go func() {
		for range 2 {
			gelesen, fehler := paar.dortigeLeitung.Empfange(t.Context())
			if fehler != nil {
				t.Errorf("Empfange: %v", fehler)

				return
			}

			angekommen <- string(gelesen)
		}
	}()

	if fehler := griff.Leere(t.Context()); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}

	if erste := <-angekommen; erste != "eins" {
		t.Fatalf("die erste ist %q — eine zuverlässige wurde verdrängt", erste)
	}

	if zweite := <-angekommen; zweite != "zwei" {
		t.Fatalf("die zweite ist %q", zweite)
	}

	// **Und das Verwerfen steht in der Übersicht** — richtig ist nicht
	// unbemerkt.
	if griff.Stand().Verworfen != 1 {
		t.Fatalf("der Stand ist %+v", griff.Stand())
	}
}

// **Ein Entzug zwischen zwei Nachrichten wirkt bei der zweiten.**
//
// **Das ist der „Berechtigungsentzug" aus dem Issue.** Wer eine Freigabe
// zurücknimmt, während eine Sitzung offen ist, will nicht, dass sie bis zum
// Schließen weiterläuft.
func TestEinEntzugWirktBeiDerNaechstenNachricht(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckHandel)

	if fehler := griff.Sende(
		t.Context(), []byte("eins"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	paar.hiesigeRechte.darf = false

	if fehler := griff.Sende(
		t.Context(), []byte("zwei"), service.Zuverlaessig,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die zweite kam durch: %v", fehler)
	}
}

// **Ein Profilwechsel schließt jede Sitzung — und verwirft, was danach
// kommt.**
//
// # Warum das Verwerfen dazugehört
//
// **Weil eine Antwort zu dem Stand gehört, aus dem sie stammt** (§7, Regel 2).
// Eine Nachricht, die nach dem Wechsel eintrifft, gehört zum **alten** Profil;
// sie im neuen einzutragen hieße, Daten zwischen zwei Menschen zu tragen, die
// an einem Rechner arbeiten.
func TestEinProfilwechselSchliesstJedeSitzung(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckTeilen)

	if fehler := paar.hier.Pausiere(); fehler != nil {
		t.Fatalf("Pausiere: %v", fehler)
	}

	if len(paar.hier.Alle()) != 0 {
		t.Error("die Sitzung steht noch in der Übersicht")
	}

	// **Der alte Griff schreibt nicht mehr.**
	if fehler := griff.Sende(
		t.Context(), []byte("zu spät"), service.Zuverlaessig,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die verspätete Nachricht kam durch: %v", fehler)
	}

	// **Und eine verspätete Antwort wird verworfen, nicht eingetragen.**
	if angenommen := griff.Empfangen([]byte("antwort")); angenommen {
		t.Error("die verspätete Antwort wurde angenommen")
	}

	// **Die Leitung ist zu** — sie gehört einem Profil, das niemand mehr
	// benutzt.
	if fehler := paar.hiesigeLeitung.Sende(
		t.Context(), []byte("hallo")); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("die Leitung blieb offen: %v", fehler)
	}
}
