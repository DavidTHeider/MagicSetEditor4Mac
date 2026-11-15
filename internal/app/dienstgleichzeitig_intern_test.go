package app

// **Zwei Rufe zugleich öffnen einen Posten und nicht zwei** — Befund 3 des
// Zweitblicks über Phase 5.4.
//
// # Der Anlass
//
// **„Läuft schon?" wurde unter dem Schloss gelesen und danach losgelassen.**
// Zwischen der Antwort und dem Eintrag liegt das Aufmachen selbst — ein
// Prozess, ein Lauscher, ein Tunnel —, und solange steht der Abbruchgriff
// auf `nil`. Zwei gleichzeitige Rufe sahen beide „frei", öffneten beide, und
// der zweite überschrieb den Griff des ersten. Der erste lief weiter: über
// das Beenden und über den Profilwechsel hinaus (§7, Regel 4).
//
// # Warum dieser Lauf nicht flackert
//
// **Weil nichts geraten wird.** Der erste Ruf wird **in** der Naht
// festgehalten, bis der zweite geantwortet hat; erst dann geht er weiter.
// Ein `time.Sleep` wäre ein Wettlauf mit der Maschine, und ein schneller
// Rechner machte ihn grün, ohne etwas zu messen.

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Dienstposten.**
func TestZweiRufeOeffnenEinenDienstUndNichtZwei(t *testing.T) {
	t.Parallel()

	drin := make(chan struct{})
	weiter := make(chan struct{})

	var versuche atomic.Int32

	anwendung := NeueApp("")

	anwendung.MitDienststarter(func(
		context.Context, domain.Dienstart, []domain.Bindung,
	) (dienstlauf, error) {
		if versuche.Add(1) == 1 {
			close(drin)
		}

		<-weiter

		return probelauf("http://127.0.0.1:8123", "geheim", ""), nil
	})

	t.Cleanup(anwendung.diensteBeenden)

	erster := make(chan error, 1)

	go func() {
		_, fehler := anwendung.dienstStarten(context.Background(),
			domain.DienstREST, domain.Bindung{})

		erster <- fehler
	}()

	<-drin

	// **Der zweite Ruf trifft einen Posten, der noch entsteht.**
	_, fehler := anwendung.dienstStarten(context.Background(),
		domain.DienstREST, domain.Bindung{})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der zweite Ruf gab %v", fehler)
	}

	close(weiter)

	if fehler := <-erster; fehler != nil {
		t.Fatalf("der erste Ruf gab %v", fehler)
	}

	// **Und das Programm lief genau einmal.** Das ist die Zusage: nicht
	// „der zweite bekommt einen Fehler", sondern „der zweite Port geht gar
	// nicht erst auf".
	if gezaehlt := versuche.Load(); gezaehlt != 1 {
		t.Errorf("das Programm wurde %dmal gestartet", gezaehlt)
	}
}

// **Der Horchposten — dieselbe Gestalt, derselbe Grund.**
func TestZweiRufeOeffnenEinenHorchpostenUndNichtZwei(t *testing.T) {
	t.Parallel()

	drin := make(chan struct{})
	weiter := make(chan struct{})

	var versuche atomic.Int32

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	anwendung.MitHorcher(func(
		context.Context, string, peer.Identitaet,
	) (service.Hereinkommend, error) {
		if versuche.Add(1) == 1 {
			close(drin)
		}

		<-weiter

		return &probelauscher{}, nil
	})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	t.Cleanup(anwendung.horchenBeenden)

	erster := make(chan *Fehler, 1)

	go func() { erster <- anwendung.HorchStarten(ctx, ":0").Fehler }()

	<-drin

	if antwort := anwendung.HorchStarten(ctx, ":0"); antwort.Fehler == nil {
		t.Error("der zweite Ruf kam durch")
	}

	close(weiter)

	if fehler := <-erster; fehler != nil {
		t.Fatalf("der erste Ruf gab %+v", fehler)
	}

	if gezaehlt := versuche.Load(); gezaehlt != 1 {
		t.Errorf("es wurden %d Lauscher aufgemacht", gezaehlt)
	}
}

// **Der Fernhorchposten — die dritte Stelle mit derselben Gestalt.**
//
// **Sie stand in der Korrekturrunde ohne Barrierelauf da** (Nachlauf zum
// Zweitblick, nachrangiger Befund 1): die Belegung war eingebaut, und
// gemessen war sie nur an zwei von drei Geschwistern. Ein Schutz, der nur
// teilweise gegen Rückschritte gehalten wird, ist teilweise keiner.
func TestZweiRufeOeffnenEineFerneinladungUndNichtZwei(t *testing.T) {
	t.Parallel()

	drin := make(chan struct{})
	weiter := make(chan struct{})

	var versuche atomic.Int32

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	anwendung.MitFernhorcher(func(peer.Identitaet) (Fernhorcher, error) {
		if versuche.Add(1) == 1 {
			close(drin)
		}

		<-weiter

		return &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil
	})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	t.Cleanup(anwendung.fernhorchenBeenden)

	erster := make(chan *Fehler, 1)

	go func() { erster <- anwendung.FernHorchStarten(ctx).Fehler }()

	<-drin

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler == nil {
		t.Error("der zweite Ruf kam durch")
	}

	close(weiter)

	if fehler := <-erster; fehler != nil {
		t.Fatalf("der erste Ruf gab %+v", fehler)
	}

	if gezaehlt := versuche.Load(); gezaehlt != 1 {
		t.Errorf("es wurden %d Tunnel aufgemacht", gezaehlt)
	}
}

// **Und die Ankündigung im Netz — die vierte.**
//
// **Sie stand in der Korrekturrunde gar nicht** (Nachlauf, Befund 1): die
// Klasse war benannt und an vier Stellen behoben, und die fünfte hatte
// dieselbe Gestalt. Genau davor warnt §10 — „welche anderen Beispiele hätten
// denselben Weg genommen?".
func TestZweiRufeKuendigenEinmalAnUndNichtZweimal(t *testing.T) {
	t.Parallel()

	drin := make(chan struct{})
	weiter := make(chan struct{})

	var versuche atomic.Int32

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	anwendung.MitEntdeckung(func(
		entdeckung.Angebot,
	) (Ankuendigung, error) {
		if versuche.Add(1) == 1 {
			close(drin)
		}

		<-weiter

		return &probeankuendigung{}, nil
	}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	if antwort := anwendung.HorchStarten(ctx, ":0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	erster := make(chan *Fehler, 1)

	go func() { erster <- anwendung.SichtbarWerden(ctx, "Davids Rechner").Fehler }()

	<-drin

	if antwort := anwendung.SichtbarWerden(ctx, "Davids Rechner"); antwort.Fehler == nil {
		t.Error("der zweite Ruf kam durch")
	}

	close(weiter)

	if fehler := <-erster; fehler != nil {
		t.Fatalf("der erste Ruf gab %+v", fehler)
	}

	if gezaehlt := versuche.Load(); gezaehlt != 1 {
		t.Errorf("es wurde %dmal angekündigt", gezaehlt)
	}
}

// **Wer währenddessen aufhört, hat aufgehört.**
//
// **Die zweite Hälfte von Befund 1**, und sie ist die unauffälligere: eine
// Anmeldung, die **nach** dem Aufhören einträfe, strahlte Peerkennung und
// Namen für einen Port aus, den niemand mehr hält — und niemand sähe, warum.
func TestEineZurueckgenommeneAnkuendigungWirdNichtEingetragen(t *testing.T) {
	t.Parallel()

	drin := make(chan struct{})
	weiter := make(chan struct{})

	// **Und sie klagt beim Einziehen.** Das ist der zweite Fall in einem:
	// wer sie zurücknimmt, kann nicht auch noch verlangen, dass das
	// gelingt — gemeldet wird er, aufgehalten wird nichts.
	angemeldet := &probeankuendigung{klagt: errors.New("mdns mault")}

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	anwendung.MitEntdeckung(func(
		entdeckung.Angebot,
	) (Ankuendigung, error) {
		close(drin)
		<-weiter

		return angemeldet, nil
	}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	if antwort := anwendung.HorchStarten(ctx, ":0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	erster := make(chan *Fehler, 1)

	go func() { erster <- anwendung.SichtbarWerden(ctx, "Davids Rechner").Fehler }()

	<-drin

	// **Mitten im Anmelden wird zurückgenommen.**
	anwendung.sichtbarkeitBeenden()

	close(weiter)

	if fehler := <-erster; fehler == nil {
		t.Error("der Ruf tat, als sei er durchgekommen")
	}

	// **Und die Anmeldung wird wirklich wieder eingezogen** — nicht nur
	// nicht eingetragen. Das ist der Unterschied zwischen „das Fenster sagt
	// nichts mehr" und „im Netz steht nichts mehr".
	if !angemeldet.beendet {
		t.Error("die Anmeldung blieb im Netz stehen")
	}

	if stand := anwendung.horchstand(); stand.Sichtbar {
		t.Errorf("der Stand sagt sichtbar: %+v", stand)
	}
}
