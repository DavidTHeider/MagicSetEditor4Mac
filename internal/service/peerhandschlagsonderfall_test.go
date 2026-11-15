package service_test

// **Die Sonderfälle des Handschlags** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine zweite Datei**, weil `peerhandschlagfehler_test.go` mit ihnen über
// die Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, wo eine
// **Leitung reißt** — hier, wo sie trägt und trotzdem nichts zustande kommt.

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

// **Eine Antwort, die keine Annahme ist, schließt nichts.**
//
// **Die Leitung trägt Bytes, nicht Bedeutung**: wer an ihrem Ende etwas
// anderes schickt, kommt an der Prüfung in `internal/peer` nicht vorbei — und
// die anbietende Seite muss das melden, statt eine Sitzung zu eröffnen.
func TestEineAntwortDieKeineAnnahmeIstSchliesstNichts(t *testing.T) {
	t.Parallel()

	// **Wer was empfängt, entscheidet die Zuordnung**: die annehmende Seite
	// liest Angebot **und** Abschluss, die anbietende nur die Annahme. Der
	// erste Anlauf hatte es vertauscht — und der Lauf war rot, zu Recht.
	faelle := map[string]struct{ hier, dort *stolperleitung }{
		"statt des Angebots kommt Unsinn": {
			hier: &stolperleitung{}, dort: &stolperleitung{unsinnAb: 1},
		},
		"statt der Annahme kommt Unsinn": {
			hier: &stolperleitung{unsinnAb: 1}, dort: &stolperleitung{},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hier, dort, kennung := mitStolper(t, fall.hier, fall.dort)

			var (
				warten     sync.WaitGroup
				dortFehler error
			)

			warten.Add(1)

			go func() {
				defer warten.Done()

				_, dortFehler = dort.Annehmen(t.Context(), fall.dort)
			}()

			_, hierFehler := hier.Oeffne(
				t.Context(), kennung, domain.ZweckTeilen)

			warten.Wait()

			if hierFehler == nil {
				t.Error("die anbietende Seite kam durch")
			}

			if dortFehler == nil {
				t.Error("die annehmende Seite kam durch")
			}
		})
	}
}

// **Ohne Zufall entsteht kein Angebot — und keine halbe Sitzung.**
func TestOhneZufallEntstehtKeinAngebot(t *testing.T) {
	t.Parallel()

	hier, _, kennung := mitStolper(t, &stolperleitung{}, &stolperleitung{})

	hier.MitZufall(versiegteQuelle{})

	if _, fehler := hier.Oeffne(
		t.Context(), kennung, domain.ZweckTeilen); fehler == nil {
		t.Fatal("das Angebot entstand ohne Zufall")
	}

	if len(hier.Alle()) != 0 {
		t.Fatal("es steht etwas in der Übersicht")
	}
}

// versiegteQuelle gibt nie etwas her.
type versiegteQuelle struct{}

func (versiegteQuelle) Read([]byte) (int, error) {
	return 0, errors.New("die Quelle ist zu")
}

// **Die annehmende Seite wählt nicht.**
//
// # Wofür dieser Lauf an die Stelle eines anderen getreten ist
//
// **Hier stand „ein Wähler, der beim Annehmen nicht wählt, gibt keine
// Sitzung"** — ein Lauf über einen Fehlerweg, den es nicht mehr gibt. Bis zu
// diesem Schnitt rief `Annehmen` selbst `Waehle(ctx, "", "")`: eine Kennung,
// die es nicht kennt, und ein Protokoll, das es nicht kennt. Wer horcht, ist
// der Transport; er nimmt an und reicht herein.
//
// **Und die Zusage wird am Schaden geprüft.** Ein Wähler, der bei jedem Ruf
// stürbe, belegte nur, dass dieser eine Lauf nicht wählt. Der Zähler sagt es
// für jeden Ruf — und die Gegenprobe steht daneben: mit dem alten
// `p.waehler.Waehle(ctx, "", "")` in `Annehmen` wird dieser Lauf rot
// (`gerufen` ist dann 1), nachgestellt am 2026-09-14.
func TestDieAnnehmendeSeiteWaehltNicht(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	// **Ein Wähler, der stürbe, wenn ihn jemand riefe** — so bleibt die
	// Gegenprobe eine Zusicherung und kein Hänger: mit dem alten Ruf
	// scheitert dieser Lauf an der Zahl darunter, nicht an der Frist.
	waehler := &leitungswaehler{fehler: errors.New("hier wählt niemand")}

	verwaltung := service.NeuePeersitzungen(&rechteprobe{darf: true},
		waehler, ausweisprobe{ich: paar.dortige}, einePeeruhr, 2)

	// **Die hereingereichte Leitung trägt nichts**, und das genügt:
	// gemessen wird, dass niemand nach einer zweiten fragt.
	stumm := &zaehlleitung{}

	if _, fehler := verwaltung.Annehmen(t.Context(), stumm); fehler == nil {
		t.Fatal("das Annehmen kam ohne Gegenüber durch")
	}

	if waehler.gerufen != 0 {
		t.Fatalf("die annehmende Seite wählte %dmal", waehler.gerufen)
	}

	// **Und sie hat die hereingereichte auch wirklich benutzt** — ohne
	// diese Zeile wäre der Zähler oben auch dann null, wenn `Annehmen`
	// gar nichts täte.
	if stumm.empfangen != 1 {
		t.Fatalf("die hereingereichte Leitung wurde %dmal gelesen",
			stumm.empfangen)
	}
}

// **Der dritte Schritt hat keine Quittung — und das ist eine Eigenschaft.**
//
// # Was ein Dreischritt-Handschlag nicht kann
//
// **Die anbietende Seite erfährt nie, ob ihr Abschluss ankam.** Sie schickt
// ihn und ist fertig; eine Bestätigung dafür bräuchte einen vierten Schritt,
// dessen Ankunft wieder niemand bestätigt. Das ist nicht zu lösen, sondern zu
// **benennen** — und der Preis ist ein Fenster, in dem eine Seite die Sitzung
// für offen hält und die andere nicht.
//
// **Was ihn schließt, ist die erste Fachnachricht.** Die Gegenseite hat
// aufgelegt; der nächste Versuch scheitert, und zwar dort, wo jemand hinsieht.
// Genau das misst dieser Lauf.
func TestDerDritteSchrittHatKeineQuittung(t *testing.T) {
	t.Parallel()

	dortigeLeitung := &stolperleitung{unsinnAb: 2}

	hier, dort, kennung := mitStolper(t, &stolperleitung{}, dortigeLeitung)

	var (
		warten     sync.WaitGroup
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		_, dortFehler = dort.Annehmen(t.Context(), dortigeLeitung)
	}()

	griff, hierFehler := hier.Oeffne(t.Context(), kennung, domain.ZweckTeilen)

	warten.Wait()

	// **Die annehmende Seite weist ab** — sie hat den Abschluss gelesen und
	// erkannt, dass er keiner ist.
	if dortFehler == nil {
		t.Fatal("die annehmende Seite kam durch")
	}

	// **Und die anbietende hält die Sitzung für offen.** Das ist kein
	// Fehler, sondern die Grenze des Ablaufs.
	if hierFehler != nil {
		t.Fatalf("die anbietende Seite scheiterte: %v", hierFehler)
	}

	// **Bis sie etwas schicken will.**
	if fehler := griff.Sende(
		t.Context(), []byte("hallo"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	if fehler := griff.Leere(t.Context()); fehler == nil {
		t.Fatal("die Nachricht ging an eine Leitung, die niemand mehr führt")
	}
}

// **Eine Leitung, die beim Niederlegen klagt, hält den Wechsel nicht auf.**
func TestEineKlagendeLeitungHaeltDenWechselNichtAuf(t *testing.T) {
	t.Parallel()

	hierLeitung := &stolperleitung{schliessfehler: errors.New("sie klagt")}
	dortLeitung := &stolperleitung{}

	hier, dort, kennung := mitStolper(t, hierLeitung, dortLeitung)

	var warten sync.WaitGroup

	warten.Add(1)

	go func() {
		defer warten.Done()

		_, _ = dort.Annehmen(t.Context(), dortLeitung) //nolint:errcheck // hier zählt die andere Seite
	}()

	if _, fehler := hier.Oeffne(
		t.Context(), kennung, domain.ZweckTeilen); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	warten.Wait()

	if fehler := hier.Pausiere(); fehler == nil {
		t.Fatal("die klagende Leitung schwieg")
	}

	// **Und der Wechsel hat trotzdem stattgefunden.**
	if len(hier.Alle()) != 0 {
		t.Fatal("die Sitzung überlebte den Wechsel")
	}
}

// reihenwaehler gibt seine Leitungen der Reihe nach heraus.
type reihenwaehler struct {
	schloss   sync.Mutex
	leitungen []service.Verbindung
	naechste  int
}

func (r *reihenwaehler) Waehle(
	context.Context, peer.Kennung, string,
) (service.Verbindung, error) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	if r.naechste >= len(r.leitungen) {
		return nil, errors.New("keine Leitung mehr")
	}

	gewaehlt := r.leitungen[r.naechste]
	r.naechste++

	return gewaehlt, nil
}

// **Dasselbe Gegenüber zweimal für denselben Zweck wird abgewiesen — auch
// von der annehmenden Seite.**
//
// # Warum es diesen Schutz zweimal gibt
//
// **Der eine steht vor dem Wählen, der andere vor dem Eintragen.** Der erste
// spart die Leitung; der zweite ist der, der wirklich trägt: zwischen beiden
// liegt der Handschlag, und in dieser Zeit kann von **außen** eine zweite
// Sitzung für dieselbe Sache entstehen. Die annehmende Seite hat den ersten
// gar nicht — sie weiß erst nach dem Angebot, wer da ist.
//
// **Ein Vorgang je Datensatz** (§7, Regel 4): zwei Sitzungen auf dieselbe
// Sache gingen gegen einen Stand, den die andere Seite noch nicht kennt.
func TestDasselbeGegenueberZweimalWirdAuchBeimAnnehmenAbgewiesen(t *testing.T) {
	t.Parallel()

	rufende, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	annehmende, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	ersteHier, ersteDort := leitung.Schleife("kanal-eins")
	zweiteHier, zweiteDort := leitung.Schleife("kanal-zwei")

	t.Cleanup(func() {
		for _, zu := range []*leitung.Leitung{
			ersteHier, ersteDort, zweiteHier, zweiteDort,
		} {
			_ = zu.Schliesse() //nolint:errcheck // nur ein Lauf
		}
	})

	// **Zwei Verwaltungen mit *derselben* Identität** — so, wie es aussähe,
	// wenn dasselbe Gegenüber ein zweites Mal anklopft.
	waehlen := func(leitungen ...service.Verbindung) *service.Peersitzungen {
		return service.NeuePeersitzungen(&rechteprobe{darf: true},
			&reihenwaehler{leitungen: leitungen},
			ausweisprobe{ich: rufende}, einePeeruhr, 2)
	}

	dort := service.NeuePeersitzungen(&rechteprobe{darf: true}, nil,
		ausweisprobe{ich: annehmende}, einePeeruhr, 2)

	// **Zwei Läufe, zwei Leitungen** — die annehmende Seite bekommt sie
	// hereingereicht, so wie ein Horcher sie hereinreichte.
	laeufe := []struct {
		hier    *service.Peersitzungen
		dortige service.Verbindung
	}{
		{waehlen(ersteHier), ersteDort},
		{waehlen(zweiteHier), zweiteDort},
	}

	for lauf, fall := range laeufe {
		hier := fall.hier

		var (
			warten     sync.WaitGroup
			dortFehler error
		)

		warten.Add(1)

		go func() {
			defer warten.Done()

			_, dortFehler = dort.Annehmen(t.Context(), fall.dortige)
		}()

		_, _ = hier.Oeffne( //nolint:errcheck // der zweite Lauf darf scheitern
			t.Context(), annehmende.Kennung(), domain.ZweckTeilen)

		warten.Wait()

		if lauf == 0 {
			if dortFehler != nil {
				t.Fatalf("der erste Lauf scheiterte: %v", dortFehler)
			}

			continue
		}

		if !errors.Is(dortFehler, domain.ErrConflict) {
			t.Fatalf("der zweite Lauf sagte: %v", dortFehler)
		}
	}

	if len(dort.Alle()) != 1 {
		t.Fatalf("es stehen %d Sitzungen darin", len(dort.Alle()))
	}
}
