package app

// **Eine Partie über eine Leitung, von einem Ende zum anderen** — Werkstück
// G5 von Phase 5.5
// (#1462).
//
// # Wofür es diesen Lauf gibt
//
// Das Issue verlangt ihn wörtlich: „Dieselbe festgelegte Command-Reihenfolge
// … über Loopback … liefert äquivalenten Fachstrom und **gleiche
// sichtbezogene Zustände**."
//
// **Und er ist der einzige, der die Kette ganz durchläuft.** Freigabe,
// Horchposten, Aufnahme, Sitzvergabe, Wirt, Netz, Gast, Meldeweg — jedes Stück
// hat seinen eigenen Lauf, und keiner davon sagt, dass sie zusammenpassen.
// Genau dafür steht §16s Satz über Wächter: er sieht, dass etwas da ist, nicht
// dass es dasselbe tut.
//
// # Was er nicht hält
//
// **Keinen zweiten Rechner.** Beide Enden laufen in diesem Prozess über eine
// Schleife; was ein echtes Netz dazwischen anders macht — Fristen, Abrisse,
// MTU —, steht hier nicht und bleibt eine Handprüfung (Vermerk 0007).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Gast setzt sich, zieht, und beide sehen dasselbe.**
func TestEinGastSetztSichUndSeinZugKommtAn(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)

	// **Die Partieleitung muss dran** — `PeerflaecheUeber` hängt nur die
	// Peerseite ein. Im Betrieb tut das `app.go` für beide zugleich; hier
	// steht der Ruf, den es tut.
	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, objekt := einTischBeimWirt(t, wirt)

	// **Der Wirt öffnet den Tisch** — ohne das läuft jede Sitzung ins Leere.
	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	// **Der Wirt nimmt an, was hereinkommt** — im Betrieb tut das der
	// Horchposten; hier steht der Ruf, den er tut.
	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{})

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	// **Der Gast bekommt einen Sitz, den der Wirt vergeben hat.**
	stand := wirt.anwendung.PartieFreigabeStand(wirt.ctx)
	if len(stand.Vergeben) != 1 || stand.Vergeben[0].Sitz != "bert" {
		t.Fatalf("vergeben ist %+v", stand.Vergeben)
	}

	gast.altesVerwerfen()

	// **Und sein Wunsch wirkt beim Wirt.**
	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	}); antwort.Fehler != nil {
		t.Fatalf("PartieWuenschen: %+v", antwort.Fehler)
	}

	getappt := warteAufGetapptesObjekt(t, wirt, partieID, objekt)
	if !getappt {
		t.Fatal("der Zug des Gastes hat beim Wirt nichts bewirkt")
	}

	// **Und der Gast sieht es** — über das Tischbild, nicht über einen
	// eigenen Kern.
	warteAufMeldung(t, gast.meldungen, EreignisPartietischbild)
}

// einTischBeimWirt legt eine Partie an, bringt ihren Beginn zu Ende und gibt
// eine Objektkennung von Bert.
//
// **Berts und nicht Annas**: der Gast sitzt auf Berts Sitz, und ein Befehl
// auf ein fremdes Objekt prüfte eine andere Frage.
//
// **Der Beginn geht über dieselbe Fläche wie das Anlegen** (#1851): vorher
// nähme der Wirt vom Gast nur ihn an (CR 103.8), und der Zug, den dieser
// Tisch tragen soll, käme nie an.
func einTischBeimWirt(t *testing.T, wirt *seite) (partieID, objekt string) {
	t.Helper()

	deckID := einDeckImBestand(t, wirt.anwendung.db)
	flaeche := PartieflaecheUeber(wirt.anwendung.db)

	antwort := flaeche.PartieAusWuenschen(
		wirt.ctx, "magic", "anna", []Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})
	if antwort.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", antwort.Fehler)
	}

	beginnAbschliessen(t, ueberDieFlaeche(wirt.ctx, flaeche), antwort.Partie)

	begonnen := flaeche.Partie(wirt.ctx, antwort.Partie.ID, "anna")
	if begonnen.Fehler != nil {
		t.Fatalf("Partie: %+v", begonnen.Fehler)
	}

	return antwort.Partie.ID,
		inDerBibliothek(t, begonnen.Partie.Objekte, "bert")
}

// warteAufMeldung wartet, bis ein bestimmtes Ereignis gemeldet wurde.
//
// **Auf die Wirkung und nicht auf den Ruf** (§3): der Wunsch ist hinaus,
// sobald `PartieWuenschen` zurückkehrt — was er bewirkt, kommt über die
// Empfangsschleife und damit später.
func warteAufMeldung(t *testing.T, gemeldet chan string, gesucht string) {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		select {
		case ereignis := <-gemeldet:
			if ereignis == gesucht {
				return
			}

		case <-frist:
			t.Fatalf("%q wurde nicht gemeldet", gesucht)
		}
	}
}

// warteAufGetapptesObjekt fragt den Wirt, bis sein Stand den Zug zeigt.
//
// **Gefragt und nicht gewartet**: der Wirt schreibt in seinen Bestand, und
// wann er damit fertig ist, sagt niemand — der Gast bekommt es später zu
// sehen als der Bestand es weiß.
func warteAufGetapptesObjekt(
	t *testing.T, wirt *seite, partieID, objekt string,
) bool {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		antwort := wirt.anwendung.Partie(wirt.ctx, partieID, "anna")
		if antwort.Fehler != nil {
			t.Fatalf("Partie: %+v", antwort.Fehler)
		}

		for _, eines := range antwort.Partie.Objekte {
			if eines.Kennung == objekt && eines.Getappt {
				return true
			}
		}

		select {
		case <-frist:
			return false
		case <-time.After(10 * time.Millisecond):
		}
	}
}
