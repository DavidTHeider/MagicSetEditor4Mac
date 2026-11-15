package service_test

// **Die Neins der Sitzungsverwaltung** — Werkstück N1 von Phase 5.4
// (#1454),
// über eine echte Leitung seit N2.
//
// **Eine zweite Datei**, weil `peersitzungen_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, was eine
// Sitzung **tut** — hier, woran sie endet.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Fehler der Rechteauskunft ist kein Nein, sondern ein Fehler.**
//
// **Der Unterschied zählt**: „darf nicht" ist eine Antwort, „ich weiß es
// nicht" ist keine — und eine Sitzung auf einer unbeantworteten Frage wäre
// eine ohne Freigabe.
func TestEinFehlerDerRechteauskunftOeffnetNichts(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	paar.hiesigeRechte.fehler = errors.New("kein Bestand")

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); fehler == nil {
		t.Fatal("die Sitzung kam durch")
	}

	if paar.hiesigerWaehler.gerufen != 0 {
		t.Fatalf("gewählt wurde %dmal", paar.hiesigerWaehler.gerufen)
	}
}

// **Ein Zweck, den es nicht gibt, hat kein Protokoll.**
func TestEinZweckOhneProtokollOeffnetNichts(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), "leihgabe",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Zweck kam durch: %v", fehler)
	}
}

// **Ein Wähler, der nicht wählt, gibt keine Sitzung.**
func TestEinWaehlerDerNichtWaehltGibtKeineSitzung(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	paar.hiesigerWaehler.fehler = errors.New("keine Leitung")

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); fehler == nil {
		t.Fatal("die Sitzung kam durch")
	}

	if len(paar.hier.Alle()) != 0 {
		t.Fatal("eine Sitzung ohne Leitung steht in der Übersicht")
	}
}

// **Ein Gegenüber, das schweigt, lässt keine Sitzung entstehen.**
//
// **Der Handschlag ist ein Wechsel**: wer wählt und nichts hört, hat eine
// Leitung und keine Sitzung. Ohne Frist hinge er dort für immer — §16 verlangt
// für jede Anfrage eine.
func TestEinSchweigendesGegenueberLaesstKeineSitzungEntstehen(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	// **Die Gegenseite hört nicht zu** — sie wird geschlossen, statt zu
	// antworten.
	if fehler := paar.dortigeLeitung.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); fehler == nil {
		t.Fatal("die Sitzung kam ohne Antwort zustande")
	}

	if len(paar.hier.Alle()) != 0 {
		t.Fatal("sie steht in der Übersicht")
	}
}

// **Ein Wechsel mitten im Handschlag lässt keine Leitung offen.**
//
// # Der Fund
//
// **Zwischen dem Prüfen und dem Eintragen liegt der Handschlag**, und der
// dauert: er geht dreimal über das Netz. Wer in dieser Lücke das Profil
// wechselt, bekäme sonst eine Sitzung, die niemand mehr führt — und die
// Leitung bliebe offen.
func TestEinWechselMittenImHandschlagLaesstKeineLeitungOffen(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	// **Der Wechsel geschieht, während gewählt wird** — also mitten im
	// Ablauf und nicht davor oder danach.
	paar.hiesigerWaehler.dazwischen = func() {
		if fehler := paar.hier.Pausiere(); fehler != nil {
			t.Errorf("Pausiere: %v", fehler)
		}
	}

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		_, _ = paar.dort.Annehmen( //nolint:errcheck // die Gegenseite darf hier scheitern
			t.Context(), paar.dortigeLeitung)
	}()

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Sitzung kam durch: %v", fehler)
	}

	if len(paar.hier.Alle()) != 0 {
		t.Fatal("sie steht in der Übersicht")
	}

	<-fertig
}

// **Eine zu große Nachricht kommt nicht in die Schlange.**
//
// **Dieselbe Grenze wie in der Hülle**, und hier, weil sie hier zuerst greift:
// wer sie erst beim Verpacken prüfte, hätte den Speicher schon gestellt.
func TestEineZuGrosseNachrichtKommtNichtInDieSchlange(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckTeilen)

	if fehler := griff.Sende(
		t.Context(), make([]byte, peer.Hoechstlaenge+1), service.Zuverlaessig,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die zu große Nachricht kam durch: %v", fehler)
	}

	if griff.Stand().Wartend != 0 {
		t.Fatalf("es warten %d", griff.Stand().Wartend)
	}
}

// **Was nicht hinausgeht, bleibt stehen.**
//
// **Eine zuverlässige Zusage bricht man nicht still.** Wer die Nachricht
// wegwürfe, hätte sie zugesagt und nicht geliefert — und der Absender erführe
// es nicht.
func TestWasNichtHinausgehtBleibtStehen(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckTeilen)

	if fehler := griff.Sende(
		t.Context(), []byte("hallo"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	// **Die Gegenseite legt auf**, bevor geleert wird.
	if fehler := paar.dortigeLeitung.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if fehler := griff.Leere(t.Context()); fehler == nil {
		t.Fatal("das Leeren schwieg")
	}

	if griff.Stand().Wartend != 1 {
		t.Fatalf("es warten %d", griff.Stand().Wartend)
	}

	// **Und nach dem Wechsel leert niemand mehr.**
	if fehler := paar.hier.Pausiere(); fehler != nil {
		t.Fatalf("Pausiere: %v", fehler)
	}

	if fehler := griff.Leere(t.Context()); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("das Leeren kam durch: %v", fehler)
	}
}

// **Eine Antwort auf eine offene Sitzung gilt — und zweimal schließen ist
// einmal.**
func TestEineAntwortAufEineOffeneSitzungGilt(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckPartie)

	if !griff.Empfangen([]byte("antwort")) {
		t.Fatal("die Antwort wurde verworfen")
	}

	if fehler := griff.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	// **Ein zweites Schließen ist kein Fehler** — das Ziel ist erreicht.
	if fehler := griff.Schliesse(); fehler != nil {
		t.Fatalf("das zweite Schließen klagte: %v", fehler)
	}
}

// **Und ein Fehler der Rechteauskunft hält auch eine offene Sitzung an.**
func TestEinFehlerDerRechteauskunftHaeltDieSitzungAn(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	griff, _ := paar.eineSitzung(t, domain.ZweckHandel)

	paar.hiesigeRechte.fehler = errors.New("kein Bestand")

	if fehler := griff.Sende(
		t.Context(), []byte("eins"), service.Zuverlaessig); fehler == nil {
		t.Fatal("die Nachricht kam durch")
	}
}
