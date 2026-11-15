package app

// **Was an einem Schalter hängt und läuft, endet mit ihm** — OPS-10b
// (#1978).
//
// Ein Lauf, dessen Beenden der Schalter sperrt, wäre danach nicht mehr
// anzuhalten: ein offener Port, den das Fenster nicht mehr schließen kann.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// netzlaeufe ist, was an `netz` und `spielen` läuft.
type netzlaeufe struct {
	anwendung    *App
	ankuendigung *probeankuendigung
	gast         *angenommenendoppel
}

// laeuftAlles gibt eine Anwendung, an der jeder Lauf von `netz` und
// `spielen` läuft: Horchposten, Ankündigung, Fernhorchen, Freigabe und der
// Sitz an einem fremden Tisch.
func laeuftAlles(t *testing.T) netzlaeufe {
	t.Helper()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	ankuendigung := &probeankuendigung{}

	anwendung.MitEntdeckung(func(entdeckung.Angebot) (Ankuendigung, error) {
		return ankuendigung, nil
	}, nil)
	anwendung.MitFernhorcher(func(peer.Identitaet) (Fernhorcher, error) {
		return &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil
	})
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

	t.Cleanup(anwendung.horchenBeenden)
	t.Cleanup(anwendung.fernhorchenBeenden)

	if antwort := anwendung.HorchStarten(ctx, ":0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SichtbarWerden(ctx, "Davids Rechner"); antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	anwendung.partie.freigabe.schloss.Lock()
	anwendung.partie.freigabe.wirt = "p1"
	anwendung.partie.freigabe.schloss.Unlock()

	gast := &angenommenendoppel{}
	anwendung.beitrittMerken(gast, "")

	return netzlaeufe{anwendung: anwendung, ankuendigung: ankuendigung, gast: gast}
}

func (n netzlaeufe) freigegeben() bool {
	n.anwendung.partie.freigabe.schloss.Lock()
	defer n.anwendung.partie.freigabe.schloss.Unlock()

	return n.anwendung.partie.freigabe.wirt != ""
}

func (n netzlaeufe) sitzt() bool {
	griff, _, _ := n.anwendung.beitritt()

	return griff != nil
}

func (n netzlaeufe) geschlossen() bool {
	n.gast.schloss.Lock()
	defer n.gast.schloss.Unlock()

	return n.gast.geschlossen
}

// **`spielen` beendet die Partie im Netz und lässt das Netz stehen** — und
// `netz` beendet danach den Rest.
func TestSpielenUndNetzBeendenIhreLaeufe(t *testing.T) {
	t.Parallel()

	laeufe := laeuftAlles(t)
	anwendung := laeufe.anwendung
	ctx := context.Background()

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("sammeln"), domain.SchalterAus)

	if !laeufe.freigegeben() || !laeufe.sitzt() || !anwendung.horchstand().Laeuft {
		t.Fatal("ein fremder Schalter beendete einen Lauf")
	}

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("spielen"), domain.SchalterAus)

	if laeufe.freigegeben() || laeufe.sitzt() || !laeufe.geschlossen() {
		t.Error("die Partie im Netz läuft nach `spielen` aus weiter")
	}

	stand := anwendung.horchstand()
	if !stand.Laeuft || !stand.Sichtbar || !anwendung.fernstand().Laeuft {
		t.Fatalf("`spielen` beendete das Netz: %+v", stand)
	}

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("netz"), domain.SchalterAus)

	if stand := anwendung.horchstand(); stand.Laeuft || stand.Sichtbar ||
		!laeufe.ankuendigung.beendet || anwendung.fernstand().Laeuft {
		t.Errorf("das Netz läuft nach `netz` aus weiter: %+v", stand)
	}
}

// **`netz` allein beendet alle fünf** — auch die Partie, deren Beenden an
// beiden hängt.
func TestNetzBeendetAuchDiePartieImNetz(t *testing.T) {
	t.Parallel()

	laeufe := laeuftAlles(t)

	laeufe.anwendung.EinstellungSetzen(context.Background(),
		domain.SchalterEinstellung("netz"), domain.SchalterAus)

	if laeufe.freigegeben() || laeufe.sitzt() || laeufe.anwendung.horchstand().Laeuft ||
		laeufe.anwendung.fernstand().Laeuft {
		t.Error("nach `netz` aus läuft noch etwas")
	}
}

// abbrechenderIndex zählt, wie oft der Aufbau abgebrochen wurde.
type abbrechenderIndex struct {
	Retrievalfuehrer

	abgebrochen int
}

func (i *abbrechenderIndex) Abbrechen() bool {
	i.abgebrochen++

	return true
}

// **`ki` bricht den Indexaufbau und jeden Bildauftrag ab** — `netz` keinen.
func TestKIBrichtIndexUndBilderAb(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

	index := &abbrechenderIndex{}
	anwendung.MitRetrieval(index)

	karten := []int64{7, 8}
	auftraege := make([]context.Context, 0, len(karten))

	for _, karte := range karten {
		auftrag, fertig, fehler := anwendung.bildlaeufe.anmelden(context.Background(), karte)
		if fehler != nil {
			t.Fatal(fehler)
		}

		t.Cleanup(fertig)

		auftraege = append(auftraege, auftrag)
	}

	ctx := context.Background()

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("netz"), domain.SchalterAus)

	if index.abgebrochen != 0 || auftraege[0].Err() != nil {
		t.Fatal("`netz` brach die KI ab")
	}

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("ki"), domain.SchalterAus)

	if index.abgebrochen != 1 {
		t.Errorf("der Index wurde %dmal abgebrochen", index.abgebrochen)
	}

	for i, auftrag := range auftraege {
		if !errors.Is(auftrag.Err(), context.Canceled) {
			t.Errorf("Bildauftrag %d läuft weiter", i)
		}
	}
}

// **`werkzeugserver` beendet den Werkzeugserver und lässt den Dienst.**
func TestDerWerkzeugserverEndetMitSeinemSchalter(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := einePostenhalterin(t, "http://127.0.0.1:8123", "geheim", "", nil)
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

	gebunden := NeueBindungen(anwendung)
	for _, starten := range []func(context.Context, string, int) DienstAntwort{gebunden.RESTStarten, gebunden.MCPStarten} {
		if antwort := starten(t.Context(), "127.0.0.1", 0); antwort.Fehler != nil {
			t.Fatalf("starten: %+v", antwort.Fehler)
		}
	}

	anwendung.EinstellungSetzen(context.Background(),
		domain.SchalterEinstellung("werkzeugserver"), domain.SchalterAus)

	if anwendung.dienststand(domain.DienstMCP).Laeuft || !anwendung.dienststand(domain.DienstREST).Laeuft {
		t.Error("`werkzeugserver` beendete den falschen")
	}
}

// vergesslicheEinstellungen schreibt und kann danach nichts lesen.
type vergesslicheEinstellungen struct{ merkendeEinstellungen }

func (vergesslicheEinstellungen) Alle(context.Context) (map[string]string, error) {
	return nil, errors.New("der Bestand ist zu")
}

// **Lässt sich der Stand nicht lesen, zählt der eben geschriebene Schalter**
// — der Lauf endet trotzdem. Ein Lesefehler darf einen Port nicht offen
// halten, den danach niemand mehr schließen kann.
func TestOhneLesbarenStandEndetDerLaufTrotzdem(t *testing.T) {
	t.Parallel()

	laeufe := laeuftAlles(t)
	laeufe.anwendung.MitEinstellungen(&vergesslicheEinstellungen{
		merkendeEinstellungen{werte: map[string]string{}},
	})

	laeufe.anwendung.EinstellungSetzen(context.Background(),
		domain.SchalterEinstellung("netz"), domain.SchalterAus)

	if laeufe.anwendung.horchstand().Laeuft {
		t.Error("der Horchposten läuft weiter")
	}
}

// **Legt ein anderer Prozess den Schalter um, endet der Lauf beim nächsten
// Weg, der fragt** — `mse`, MCP und REST setzen dieselben Einstellungen, und
// das Fenster erfährt es erst am Tor.
func TestEinAndererProzessBeendetDenLaufAmTor(t *testing.T) {
	t.Parallel()

	laeufe := laeuftAlles(t)
	einstellungen := &merkendeEinstellungen{werte: map[string]string{}}
	laeufe.anwendung.MitEinstellungen(einstellungen)

	einstellungen.werte[domain.SchalterEinstellung("netz")] = domain.SchalterAus

	if !laeufe.anwendung.horchstand().Laeuft {
		t.Fatal("der Horchposten endete ohne einen Weg, der fragt")
	}

	abgewiesen := func() (panik any) {
		defer func() { panik = recover() }()

		NeueBindungen(laeufe.anwendung).HorchStand(t.Context())

		return nil
	}()

	if _, istTor := abgewiesen.(torfehler); !istTor {
		t.Fatalf("das Tor wies nicht ab: %v", abgewiesen)
	}

	if laeufe.anwendung.horchstand().Laeuft || laeufe.sitzt() {
		t.Error("nach dem Tor läuft das Netz weiter")
	}
}
