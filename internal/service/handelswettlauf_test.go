package service_test

// **Was geschieht, wenn zwei gleiche Zustellungen sich überholen** — Befund 4
// des Zweitblicks über Phase 5.6
// (#1477).
//
// # Warum das eine eigene Datei ist
//
// **Weil `handelsanschluss_test.go` an der Zeilengrenze stand** (§2) — und
// weil die Frage eine andere ist: dort geht es darum, **woran** eine
// Nachricht anschließt, hier darum, was bei **gleichzeitig** ankommenden
// geschieht.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Wettlauf zweier gleicher Zustellungen endet mit derselben Antwort** —
// Befund 4 des Zweitblicks über Phase 5.6.
//
// # Warum das die zugesagte Idempotenz verletzte
//
// **Die Ablage fängt den zweiten Zug am Primärschlüssel**, und ihr eigener
// Kopf sagt: „der Aufrufer liest danach nach, was dasteht — statt zu raten."
// Der Aufrufer las nicht nach; er reichte den Schreibfehler durch. Zwei
// gleiche Nachrichten, die sich im Fenster zwischen Nachschlagen und
// Eintragen überholen, bekamen deshalb **verschiedene** Antworten — gerade
// im Abriss- und Wiederholungsfall, für den es die Idempotenz gibt.
func TestZweiGleicheZustellungenGebenDieselbeAntwort(t *testing.T) {
	t.Parallel()

	// **Ein Briefkasten, der beim Eintragen den Wettlauf meldet** und beim
	// Nachschlagen die Antwort von drüben kennt.
	briefkasten := &wettlaufprobe{
		ergebnis: derPostvorgang + "/2", abdruck: einNachfolger().Abdruck(),
	}

	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	antwort, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil))
	if fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	switch {
	case antwort.Art != domain.ZustellungBekannt:
		t.Errorf("die Zustellung gilt als %q", antwort.Art)
	case antwort.Ergebnis != derPostvorgang+"/2":
		t.Errorf("das Ergebnis ist %q", antwort.Ergebnis)
	}
}

// **Und wenn der Blick danach scheitert, kommt sein Fehler.**
// **Ein Rennen mit anderem Inhalt ist ein Konflikt** — Befund Z2 des
// Zweitblicks über Phase 6.8 (#1669). Hier stand „bekannt", gleich was der
// Vermerk von drüben trug: zwei verschiedene Nachrichten unter demselben
// Schlüssel bekamen beide eine Erfolgsantwort.
func TestEinWettlaufMitAnderemInhaltIstEinKonflikt(t *testing.T) {
	t.Parallel()

	post := service.NeueHandelspost(
		&wettlaufprobe{ergebnis: derPostvorgang + "/2", abdruck: "ein-anderer-abdruck"},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	antwort, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil))
	if fehler != nil || antwort.Art != domain.ZustellungKonflikt {
		t.Errorf("Antwort %+v, Fehler %v", antwort, fehler)
	}
}

func TestEinGescheiterterBlickNachDemWettlaufKommtDurch(t *testing.T) {
	t.Parallel()

	post := service.NeueHandelspost(
		&wettlaufprobe{leseklage: errors.New("der Bestand klemmt")},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); fehler == nil {
		t.Error("der Lesefehler kam als Erfolg zurück")
	}
}

// **Und ein Wettlauf ohne Eintrag dahinter bleibt ein Fehler.**
//
// **Die Gegenprobe**: der Schlüsselkonflikt heißt „jemand war schneller".
// Ist danach trotzdem nichts da, hat niemand etwas eingetragen — und ein
// erfundenes „schon behandelt" ließe die Nachricht nie wiederkommen.
func TestEinWettlaufOhneEintragBleibtEinFehler(t *testing.T) {
	t.Parallel()

	post := service.NeueHandelspost(&wettlaufprobe{},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); fehler == nil {
		t.Error("der leere Wettlauf kam als Erfolg zurück")
	}
}

// wettlaufprobe meldet beim Eintragen den Schlüsselkonflikt und kennt beim
// zweiten Nachschlagen das Ergebnis des Gewinners — oder nichts.
type wettlaufprobe struct {
	ergebnis string

	// abdruck ist der Nutzlastabdruck, den der Vermerk von drüben trägt.
	abdruck string

	gefragt int

	// leseklage lässt das **zweite** Nachschlagen scheitern — den Blick
	// nach dem verlorenen Wettlauf.
	leseklage error
}

func (w *wettlaufprobe) Nachschlagen(
	_ context.Context, schluessel string,
) (*domain.Inboxeintrag, error) {
	w.gefragt++

	// **Der erste Blick sieht nichts** — sonst führte die Post gar nicht
	// erst aus, und der Wettlauf käme nicht zustande.
	if w.gefragt == 1 {
		return nil, nil
	}

	if w.leseklage != nil {
		return nil, w.leseklage
	}

	if w.ergebnis == "" {
		return nil, nil
	}

	return &domain.Inboxeintrag{
		Schluessel: schluessel, Ergebnis: w.ergebnis, Nutzlastabdruck: w.abdruck,
	}, nil
}

func (w *wettlaufprobe) Zustellen(
	context.Context, domain.Inboxeintrag, domain.Handelsangebot, domain.Handelsangebot,
	domain.Zusagenwirkungsart,
) error {
	return domain.MitGrund(domain.ErrConflict, nil,
		"diesen Idempotenzschlüssel gibt es schon")
}
