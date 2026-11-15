package app

// **Was der Nachlauf über Phase 6.12 fand** — N1 und N2
// (#1905).
//
// Die Korrekturrunde fragte nach dem Start den **Stand** der Schalter. Der
// sagt nicht, ob dazwischen abgeschaltet wurde: aus und wieder an während
// eines Starts ließ den Lauf weiterlaufen (N2), und ein Gast, der sich
// eintrug, nachdem das Abschalten die Runde entlassen hatte, blieb (N1).
// Jetzt zählt das Abschalten je Lauf mit, und ein Start vergleicht mit dem
// Zählerstand von vorher.

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **N2 — aus und wieder an, während ein Dienst startet: er läuft danach
// nicht.** Der Start wird **in** der Naht festgehalten, bis beide Schalter
// geschrieben sind; ein `time.Sleep` wäre ein Wettlauf mit der Maschine.
func TestAusUndAnWaehrendDesStartsBeendetDenDienst(t *testing.T) {
	t.Parallel()

	drin := make(chan struct{})
	weiter := make(chan struct{})

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})
	anwendung.MitDienststarter(func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {
		close(drin)
		<-weiter

		return probelauf("http://127.0.0.1:8123", "geheim", ""), nil
	})
	t.Cleanup(anwendung.diensteBeenden)

	gestartet := make(chan error, 1)

	go func() {
		_, fehler := anwendung.dienstStarten(context.Background(), domain.DienstREST, domain.Bindung{})
		gestartet <- fehler
	}()

	<-drin

	ctx := context.Background()
	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("dienst"), domain.SchalterAus)
	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("dienst"), domain.SchalterAn)

	close(weiter)

	if fehler := <-gestartet; fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	if anwendung.dienststand(domain.DienstREST).Laeuft {
		t.Error("der Dienst läuft, obwohl er während des Starts abgeschaltet wurde")
	}
}

// **N2 — eine verspätete Nachfrage beendet nur ihren eigenen Lauf.** Ein
// Bildauftrag fragte nach dem Start noch nach `ki`; dazwischen war `ki` aus
// und wieder an, und ein neuer Auftrag lief an. Der darf bleiben.
func TestEineVerspaeteteNachfrageBeendetNurIhrenLauf(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	ctx := context.Background()

	marke := anwendung.laufmarke(domain.LaufBildauftraege)
	anwendung.laeufeBeenden(ctx, domain.SchalterstandAus(map[string]string{
		domain.SchalterEinstellung("ki"): domain.SchalterAus,
	}))

	neuer, fertig, fehler := anwendung.bildlaeufe.anmelden(ctx, 8)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(fertig)

	if beendet := anwendung.nachDemStart(ctx, domain.LaufBildauftraege, marke, func(context.Context) {
		anwendung.bildlaeufe.abbrechen(7)
	}); !beendet {
		t.Error("die Nachfrage sah das Abschalten nicht")
	}

	if neuer.Err() != nil {
		t.Error("die verspätete Nachfrage brach einen neuen Auftrag ab")
	}
}

// **N1 — ein Gast, der sich nach dem Abschalten einträgt, geht.**
func TestEinVerspaeteterGastGeht(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

	gast := &angenommenendoppel{}
	marke := anwendung.laufmarke(domain.LaufPartiefreigabe)

	anwendung.EinstellungSetzen(context.Background(), domain.SchalterEinstellung("spielen"), domain.SchalterAus)
	anwendung.EinstellungSetzen(context.Background(), domain.SchalterEinstellung("spielen"), domain.SchalterAn)

	if bleibt := anwendung.gastEintragen(context.Background(), gast, einGastplatz(gast), marke); bleibt {
		t.Error("der Gast blieb nach dem Abschalten")
	}

	gast.schloss.Lock()
	defer gast.schloss.Unlock()

	if !gast.geschlossen || len(anwendung.partie.runde.Sitze()) != 0 {
		t.Errorf("geschlossen %v, Sitze %v", gast.geschlossen, anwendung.partie.runde.Sitze())
	}
}

// **N1 an der Aufnahme selbst — ein anderer Prozess hat `spielen`
// abgeschaltet.** Die Wache dieses Fensters hat den Stand noch nicht
// gelesen, die Freigabe steht noch; ein Gast, der jetzt ankommt, wird nicht
// bedient, sondern abgewiesen.
func TestEinGastNachFremdemAbschaltenWirdNichtBedient(t *testing.T) {
	t.Parallel()

	einstellungen := &merkendeEinstellungen{werte: map[string]string{}}
	anwendung := freigabeanwendung(t)
	anwendung.MitEinstellungen(einstellungen)
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	// **Am Fenster vorbei**, so wie ein zweiter Prozess in den Bestand schreibt.
	einstellungen.werte[domain.SchalterEinstellung("spielen")] = domain.SchalterAus

	gast := eineAngenommene(t, domain.ZweckPartie)
	anwendung.partiesitzungAufnehmen(t.Context(), gast)

	if !gast.istZu() || len(anwendung.partie.runde.Sitze()) != 0 {
		t.Errorf("zu %v, Sitze %v", gast.istZu(), anwendung.partie.runde.Sitze())
	}
}

// wartenderAufbau hält den Aufbau fest, bis sein Kontext endet.
type wartenderAufbau struct {
	Retrievalfuehrer

	drin chan struct{}
}

func (w *wartenderAufbau) Aufbauen(ctx context.Context, _ service.Aufbauauftrag) (service.Aufbaubericht, error) {
	close(w.drin)
	<-ctx.Done()

	return service.Aufbaubericht{}, ctx.Err()
}

func (w *wartenderAufbau) Abbrechen() bool { return false }

// **`ki` aus erreicht einen Aufbau, der schon läuft** — über den Abbruch,
// den der Start abgelegt hat, und nicht erst über den Führer. Seit N2 endet
// ein Start über seinen eigenen Abbruch; dieser Weg hatte danach keinen Lauf
// mehr.
func TestAbschaltenBrichtDenLaufendenAufbauAb(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

	aufbau := &wartenderAufbau{drin: make(chan struct{})}
	anwendung.MitRetrieval(aufbau)

	antwort := make(chan IndexaufbauAntwort, 1)

	go func() {
		antwort <- anwendung.IndexAufbauen(context.Background(), "m", "q", "r", "", false)
	}()

	<-aufbau.drin
	anwendung.EinstellungSetzen(context.Background(), domain.SchalterEinstellung("ki"), domain.SchalterAus)

	select {
	case geantwortet := <-antwort:
		if geantwortet.Fehler == nil {
			t.Error("der Aufbau endete ohne Fehler, obwohl `ki` ihn abbrach")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("`ki` aus hat den laufenden Aufbau nicht erreicht")
	}
}
