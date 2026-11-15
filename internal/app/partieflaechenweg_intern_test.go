package app

// **Der Bauweg der kopflosen Flächen** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Und der Weg danach**: dieselbe Fläche stellt, liest, hält an und
// beendet ([TestDieKopfloseFlaecheGehtDenselbenWeg]) — der Lauf stand bis
// #1851 in `partie_intern_test.go`, und der Beginn davor hätte die Datei
// über die Zeilengrenze aus §2 gebracht.
//
// # Warum das hier geprüft wird und nicht nur in `internal/flaechen`
//
// **Weil dort die Abdeckung dieses Pakets nicht zählt.** Die Flächenläufe
// fahren gebaute Programme; was sie betreten, steht in ihrem eigenen
// Profil. `PartieAusWuenschen` ist der Weg, den `mse`, MCP und der Dienst
// **alle drei** nehmen — und er steht hier, damit er nicht dreimal dastünde
// (§2). Ein Weg, den nur ein anderes Paket betritt, ist hier ungemessen.

import (
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Aus zwei Wünschen wird ein Tisch — und der erste Sitz ist der Wirt.**
func TestDieKopfloseFlaecheBautAusWuenschen(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	deckID := einDeckImBestand(t, db)

	antwort := PartieflaecheUeber(db).PartieAusWuenschen(ctx, "magic", "",
		[]Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: deckID},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
		})
	if antwort.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", antwort.Fehler)
	}

	switch {
	case antwort.Partie.Wirt != "sitz-1":
		t.Errorf("Wirt %q — ohne Angabe ist es der erste Sitz",
			antwort.Partie.Wirt)
	case antwort.Partie.Regelwerk.Kennung != "sandkasten":
		t.Errorf("Regelwerk %+v — die Vorgabe ist der Sandkasten",
			antwort.Partie.Regelwerk)
	case len(antwort.Partie.Sitze) != 2:
		t.Errorf("%d Sitze", len(antwort.Partie.Sitze))
	}

	// **Die Kennung ist frisch und nicht leer.** Zwei Partien aus denselben
	// Decks sind zwei Partien; eine feste Kennung ließe die zweite die erste
	// überschreiben.
	if len(antwort.Partie.ID) != 32 {
		t.Errorf("die Kennung %q ist keine frische", antwort.Partie.ID)
	}
}

// **Und ein Deck, das keine Zahl ist, wird abgewiesen, bevor etwas entsteht.**
//
// **Der Fehlerweg von [App.partieAusWuenschenMit]** — er steht hier, weil er
// sonst nur über die drei Flächen ohne Fenster betreten würde.
func TestDieKopfloseFlaecheWeistEinUnleserlichesDeckAb(t *testing.T) {
	t.Parallel()

	antwort := PartieflaecheUeber(bestandFuerDenRiegel(t)).PartieAusWuenschen(
		riegelkontext(t), "magic", "", []Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: "kein Deck"},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: "auch nicht"},
		})

	if antwort.Fehler == nil {
		t.Fatal("das unleserliche Deck kam durch")
	}

	// **Der Satz nennt den Sitz** — sonst suchte jemand in zwei Feldern.
	if !strings.Contains(antwort.Fehler.Meldung, "sitz-1") {
		t.Errorf("die Meldung nennt den Sitz nicht: %q", antwort.Fehler.Meldung)
	}

	// **Und die Listen sind leer und nicht `nil`** — dieselbe Zusage wie auf
	// dem Erfolgsweg.
	if antwort.Partie.Sitze == nil || antwort.Erlaubt == nil {
		t.Errorf("der Fehlerweg gibt nil: %+v", antwort.Partie)
	}
}

// **Das Sandkastenwerk ist die Vorgabe, und sie steht an einer Stelle** —
// `regelwerkAus` über einem leeren Wunsch. Hier stand bis #2436 ein eigenes
// `Sandkastenwerk()`, „damit die drei Flächen ohne Fenster dasselbe nehmen";
// keine nahm es, denn alle gehen über `regelwerkAus`.
func TestDasSandkastenwerkIstDieVorgabe(t *testing.T) {
	t.Parallel()

	werk := regelwerkAus(RegelwerkDTO{})

	if werk.Kennung != "sandkasten" || werk.Fassung != 1 ||
		string(werk.Strenge) != "sandkasten" {
		t.Errorf("%+v", werk)
	}
}

// **Und zwei frische Kennungen sind zwei.**
func TestNeuePartiekennungGibtJedesMalEineAndere(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for i := range 32 {
		kennung := NeuePartiekennung()

		if gesehen[kennung] {
			t.Fatalf("die Kennung %q kam beim %d. Mal wieder",
				kennung, i+1)
		}

		gesehen[kennung] = true

		if _, fehler := strconv.ParseUint(kennung[:8], 16, 64); fehler != nil {
			t.Errorf("die Kennung %q ist nicht hexadezimal", kennung)
		}
	}
}

// **Die kopflose Fläche geht denselben Weg.**
func TestDieKopfloseFlaecheGehtDenselbenWeg(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	deckID := einDeckImBestand(t, db)
	flaeche := PartieflaecheUeber(db)

	kennung, fehler := strconv.ParseInt(deckID, 10, 64)
	if fehler != nil {
		t.Fatalf("die Deckkennung: %v", fehler)
	}

	aufbau := NeueApp("")
	MitPartienleitung(aufbau, db)

	partie, fehler := aufbau.partieaufbau().Aufbauen(ctx, "p1", "magic",
		regelwerkAus(RegelwerkDTO{}), "sitz-1", []service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: kennung},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: kennung},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	angelegt := flaeche.PartieAnlegen(ctx, partie)
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	// **Erst der Beginn, über dieselbe Fläche** — CR 103.8 (#1851).
	beginnAbschliessen(t, ueberDieFlaeche(ctx, flaeche), angelegt.Partie)

	objekt := inDerBibliothek(t, flaeche.Partie(ctx, "p1", "sitz-1").
		Partie.Objekte, "sitz-1")

	if gestellt := flaeche.PartieBefehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: "objekt.tappen", Nutzlast: map[string]any{"objekt": objekt},
	}); gestellt.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", gestellt.Fehler)
	}

	if gelesen := flaeche.Partie(ctx, "p1", "sitz-1"); gelesen.Fehler != nil {
		t.Fatalf("Partie: %+v", gelesen.Fehler)
	}

	if pausiert := flaeche.PartiePausieren(
		ctx, "p1", "sitz-1"); pausiert.Partie.Stand != "pausiert" {
		t.Errorf("PartiePausieren: %+v", pausiert)
	}

	if fort := flaeche.PartieFortsetzen(
		ctx, "p1", "sitz-1"); fort.Partie.Stand != "laeuft" {
		t.Errorf("PartieFortsetzen: %+v", fort)
	}

	if beendet := flaeche.PartieBeenden(
		ctx, "p1", "sitz-1"); beendet.Partie.Stand != "beendet" {
		t.Errorf("PartieBeenden: %+v", beendet)
	}
}
