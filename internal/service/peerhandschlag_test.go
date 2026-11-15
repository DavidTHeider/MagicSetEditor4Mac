package service_test

// **Der Handschlag in der Sitzungsverwaltung** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass eine Sitzung erst offen ist, wenn beide Seiten sich belegt haben.**
// N1.3 hat den Handschlag gebaut, N1.5 den Lebenszyklus — und **nichts
// verband die beiden**: `Oeffne` prüfte das Recht und wählte, mehr nicht. Eine
// `peer.Sitzung` entstand oberhalb von `internal/peer` in keiner Schicht.
//
// **Und dass ein Gegenüber ohne Freigabe keine Antwort bekommt.** Die
// Reihenfolge ist die Zusage: erst prüfen, dann antworten. Wer zuerst
// antwortet, hat sich einem Fremden gegenüber ausgewiesen, dem er nichts
// erlaubt.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ausweisprobe gibt eine feste Identität — oder keine.
type ausweisprobe struct {
	ich    peer.Identitaet
	fehler error
}

func (a ausweisprobe) Ausweis(context.Context) (peer.Identitaet, error) {
	return a.ich, a.fehler
}

// leitungswaehler gibt genau eine vorbereitete Leitung — und zählt mit.
type leitungswaehler struct {
	leitung *leitung.Leitung
	fehler  error
	gerufen int

	// dazwischen läuft mitten im Wählen — für den Wechsel, der genau dann
	// stattfindet.
	dazwischen func()

	// umhuellen legt eine Hülle um die Leitung — für einen Lauf, der ein
	// Senden anhalten muss (`sitzungtiefe_test.go`).
	umhuellen func(service.Verbindung) service.Verbindung
}

func (l *leitungswaehler) Waehle(
	context.Context, peer.Kennung, string,
) (service.Verbindung, error) {
	l.gerufen++

	if l.dazwischen != nil {
		l.dazwischen()
	}

	if l.fehler != nil {
		return nil, l.fehler
	}

	if l.umhuellen != nil {
		return l.umhuellen(l.leitung), nil
	}

	return l.leitung, nil
}

// paarung ist alles, was ein Lauf über zwei Seiten braucht.
type paarung struct {
	hier, dort       *service.Peersitzungen
	hiesige, dortige peer.Identitaet
	hiesigeLeitung   *leitung.Leitung
	dortigeLeitung   *leitung.Leitung
	hiesigerWaehler  *leitungswaehler
	hiesigeRechte    *rechteprobe
	dortigeRechte    *rechteprobe
}

// eineSitzung führt den Handschlag und gibt beide Griffe.
//
// **Über eine echte Leitung und zwei echte Identitäten** — was hier nicht echt
// ist, sind die Rechte und die Uhr. Ein Doppel, das den Handschlag nur
// nachspielte, prüfte eine Sitzung, die es so nie gibt.
func (p *paarung) eineSitzung(
	t *testing.T, zweck domain.Zweck,
) (hier, dort *service.Sitzungsgriff) {
	t.Helper()

	var (
		warten     sync.WaitGroup
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		dort, dortFehler = p.dort.Annehmen(t.Context(), p.dortigeLeitung)
	}()

	hier, fehler := p.hier.Oeffne(t.Context(), p.dortige.Kennung(), zweck)

	warten.Wait()

	if fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	if dortFehler != nil {
		t.Fatalf("Annehmen: %v", dortFehler)
	}

	return hier, dort
}

// zweiSeiten baut zwei Verwaltungen, die über eine Schleife verbunden sind.
//
// **Zwei echte Identitäten und eine echte Leitung** — was hier nicht echt ist,
// sind die Rechte und die Uhr.
func zweiSeiten(t *testing.T, darf bool) *paarung {
	t.Helper()

	hiesige, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	dortige, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	hiesigeLeitung, dortigeLeitung := leitung.Schleife("kanal-schleife")

	t.Cleanup(func() {
		_ = hiesigeLeitung.Schliesse() //nolint:errcheck // nur ein Lauf
		_ = dortigeLeitung.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	hiesigeRechte := &rechteprobe{darf: true}
	dortigeRechte := &rechteprobe{darf: darf}
	hiesigerWaehler := &leitungswaehler{leitung: hiesigeLeitung}

	return &paarung{
		hier: service.NeuePeersitzungen(hiesigeRechte, hiesigerWaehler,
			ausweisprobe{ich: hiesige}, einePeeruhr, 2),
		// **Die annehmende Seite bekommt `nil` als Wähler**, und das ist
		// die Aussage: sie wählt nicht. Griffe einer von ihnen doch
		// danach, stürzte dieser Lauf — lauter, als ein Zähler es sagen
		// könnte.
		dort: service.NeuePeersitzungen(dortigeRechte, nil,
			ausweisprobe{ich: dortige}, einePeeruhr, 2),
		hiesige:         hiesige,
		dortige:         dortige,
		hiesigeLeitung:  hiesigeLeitung,
		dortigeLeitung:  dortigeLeitung,
		hiesigerWaehler: hiesigerWaehler,
		hiesigeRechte:   hiesigeRechte,
		dortigeRechte:   dortigeRechte,
	}
}

// **Beide Seiten kommen auf dieselbe Sitzung — über eine echte Leitung.**
//
// # Der Lauf, für den es diesen ganzen Schnitt gibt
//
// **Er fährt jede Schicht außer den Bytes auf dem Draht**: zwei Identitäten,
// zwei Rechteprüfungen, der Handschlag über drei Nachrichten, die Rahmung,
// die Sitzungsverwaltung. Ohne Netz, ohne Berechtigung, ohne zweiten Rechner
// — und deshalb ohne Flackern.
func TestBeideSeitenKommenUeberEineLeitungAufDieselbeSitzung(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	hierGriff, dortGriff := paar.eineSitzung(t, domain.ZweckTeilen)

	hiesigeSitzung := hierGriff.Sitzung()
	dortigeSitzung := dortGriff.Sitzung()

	// **Dieselbe Sitzungskennung**, und keine Seite hat sie gewählt.
	if hiesigeSitzung.Kennung == "" ||
		hiesigeSitzung.Kennung != dortigeSitzung.Kennung {
		t.Fatalf("zwei Sitzungen: %q und %q",
			hiesigeSitzung.Kennung, dortigeSitzung.Kennung)
	}

	// **Jede Seite sieht die andere als Kennung** — belegt, nicht behauptet.
	if hiesigeSitzung.Gegenueber != paar.dortige.Kennung() ||
		dortigeSitzung.Gegenueber != paar.hiesige.Kennung() {
		t.Fatalf("die Gegenüber sind %s und %s",
			hiesigeSitzung.Gegenueber, dortigeSitzung.Gegenueber)
	}

	// **Und die Rollen stehen über Kreuz.**
	if hiesigeSitzung.MeineRolle != dortigeSitzung.SeineRolle ||
		dortigeSitzung.MeineRolle != hiesigeSitzung.SeineRolle {
		t.Fatalf("die Rollen sind %+v und %+v",
			hiesigeSitzung, dortigeSitzung)
	}
}

// **Ein Gegenüber ohne Freigabe bekommt keine Antwort.**
//
// # Die Reihenfolge ist die Zusage
//
// **Erst prüfen, dann antworten.** Die annehmende Seite erfährt erst aus dem
// Angebot, **wer** da ist und **was** er will; die Antwort geht danach hinaus.
// Wer zuerst antwortete, hätte sich einem Fremden gegenüber ausgewiesen, dem
// er nichts erlaubt — und der wüsste damit, dass es dieses Profil gibt.
func TestEinGegenueberOhneFreigabeBekommtKeineAntwort(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, false)

	var (
		warten     sync.WaitGroup
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		_, dortFehler = paar.dort.Annehmen(t.Context(), paar.dortigeLeitung)
	}()

	// **Die anbietende Seite scheitert am Warten auf eine Antwort**, die
	// nicht kommt — die Leitung wird von der Gegenseite geschlossen.
	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckHandel,
	); fehler == nil {
		t.Error("die Sitzung kam ohne Freigabe zustande")
	}

	warten.Wait()

	if !errors.Is(dortFehler, domain.ErrValidation) {
		t.Fatalf("die annehmende Seite sagte: %v", dortFehler)
	}

	if len(paar.dort.Alle()) != 0 {
		t.Fatal("die abgelehnte Sitzung steht in der Übersicht")
	}
}

// **Ohne Ausweis wird gar nicht erst gewählt.**
//
// **Ein Profil ohne Identität kann nicht unterschreiben** — und eine Leitung
// aufzubauen, um dann festzustellen, dass man nichts unterschreiben kann,
// hätte dem Gegenüber schon gesagt, dass es dieses Programm gibt.
func TestOhneAusweisWirdGarNichtErstGewaehlt(t *testing.T) {
	t.Parallel()

	waehler := &leitungswaehler{}

	verwaltung := service.NeuePeersitzungen(
		&rechteprobe{darf: true}, waehler,
		ausweisprobe{fehler: domain.MitGrund(
			domain.ErrNotFound, nil, "dieses Profil hat keine Identität")},
		einePeeruhr, 2)

	kennung, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if _, fehler := verwaltung.Oeffne(
		t.Context(), kennung.Kennung(), domain.ZweckTeilen,
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam etwas anderes: %v", fehler)
	}

	// **Und beim Annehmen ebenso** — dort steht schon eine Leitung, aber
	// ohne Ausweis wird nicht einmal gelesen.
	//
	// **Und „nicht einmal gelesen" ist gemessen** (§10): der Satz stand
	// hier als Behauptung neben einem Lauf, der nur den Fehler prüfte. Eine
	// Leitung, die das Lesen zählt, sagt es.
	stumm := &zaehlleitung{}

	if _, fehler := verwaltung.Annehmen(t.Context(), stumm); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("das Annehmen kam durch: %v", fehler)
	}

	if stumm.empfangen != 0 {
		t.Fatalf("es wurde %dmal gelesen", stumm.empfangen)
	}
}

// zaehlleitung zählt, was an ihr geschieht, und trägt nichts.
type zaehlleitung struct {
	empfangen   int
	gesendet    int
	geschlossen bool
}

func (z *zaehlleitung) Kanal() string           { return "" }
func (*zaehlleitung) Gegenstelle() peer.Kennung { return "" }

func (z *zaehlleitung) Sende(context.Context, []byte) error {
	z.gesendet++

	return nil
}

func (z *zaehlleitung) Empfange(context.Context) ([]byte, error) {
	z.empfangen++

	return nil, errors.New("hier kommt nichts")
}

func (z *zaehlleitung) Schliesse() error {
	z.geschlossen = true

	return nil
}
