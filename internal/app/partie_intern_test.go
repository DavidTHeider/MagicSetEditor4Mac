package app

// **Die Partie über die Naht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die fünfte Abnahme des Werkstücks
//
// „Fähigkeiten Session erstellen/lesen/Command/pausieren/beenden sind über
// die vier bestehenden Flächen **fachlich gleich**, mit derselben Profil-
// und Rollenprüfung."
//
// **Hier zwei davon gegen denselben Bestand** — das Fenster über [Bindungen]
// und die kopflose Fläche über [Partieflaeche]; deren Lauf steht bei ihrem
// Bauweg in `partieflaechenweg_intern_test.go`. Dass auch `mse`, MCP und
// REST denselben Weg gehen, misst `internal/flaechen` an den **gebauten**
// Programmen; ein Lauf hier könnte das nicht belegen.

import (
	"database/sql"
	"slices"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einDeckImBestand legt ein spielbares Deck an und gibt seine Kennung.
//
// **Zehn Karten**: sieben für die Starthand (CR 103.5) und drei, damit auch
// der nächste Zug noch ziehen kann. Bis #1851 waren es sechs — genug für
// eine Partie, die ohne Beginn spielte, und zu wenig für eine Starthand.
func einDeckImBestand(t *testing.T, db *sql.DB) string {
	t.Helper()

	ctx := riegelkontext(t)
	ablage := storage.NeueDeckablage(db)

	deck, fehler := ablage.Anlegen(ctx, domain.Deck{Name: "Probe", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("Deck anlegen: %v", fehler)
	}

	for _, platz := range []domain.Deckplatz{
		{Printing: "wald", Menge: 8, Bereich: domain.BereichHauptdeck},
		{Printing: "baer", Menge: 2, Bereich: domain.BereichHauptdeck},
	} {
		if _, fehler := ablage.PlatzSetzen(ctx, deck.ID, platz); fehler != nil {
			t.Fatalf("Platz setzen: %v", fehler)
		}
	}

	return strconv.FormatInt(deck.ID, 10)
}

// eineSpielende baut eine Anwendung mit Partien über einem echten Bestand.
func eineSpielende(t *testing.T) (gebunden *Bindungen, deckID string) {
	t.Helper()

	db := bestandFuerDenRiegel(t)

	anwendung := MitPartienleitung(NeueApp(""), db)
	anwendung.db = db

	// **Das gezeigte Profil steht sonst nirgends.** Im Fenster liest
	// [App.Starten] den Zeiger, bevor die erste Bindung ruft; ein Lauf, der
	// keine Anwendung startet, setzt ihn hier — der Grund steht bei
	// [App.imProfil].
	anwendung.profil.Store(1)

	return NeueBindungen(anwendung), einDeckImBestand(t, db)
}

// einEntwurf ist der Wunsch nach einer Partie über zwei Sitze.
func einEntwurf(deckID string) PartieentwurfDTO {
	return PartieentwurfDTO{
		Spiel: "magic",
		Sitze: []SitzwunschDTO{
			{Kennung: "sitz-1", Name: "David", DeckID: deckID},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
		},
	}
}

// **Eine Partie entsteht aus zwei Decks und lässt sich spielen.**
func TestEinePartieEntstehtUndLaesstSichSpielen(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	// **Zwanzig Objekte** — zehn je Sitz, alle in der Bibliothek.
	if len(angelegt.Partie.Objekte) != 20 {
		t.Fatalf("es liegen %d Objekte", len(angelegt.Partie.Objekte))
	}

	// **Der Wirt ist der erste Sitz**, weil der Entwurf keinen nennt.
	if angelegt.Partie.Wirt != "sitz-1" || angelegt.Partie.Fuer != "sitz-1" {
		t.Errorf("Wirt %q, Sicht für %q",
			angelegt.Partie.Wirt, angelegt.Partie.Fuer)
	}

	// **Und der Sandkasten gilt**, weil der Entwurf kein Regelwerk nennt.
	if angelegt.Partie.Regelwerk.Kennung != "sandkasten" ||
		angelegt.Partie.Regelwerk.Fassung != domain.Grundprofilfassung {
		t.Errorf("das Regelwerk ist %+v", angelegt.Partie.Regelwerk)
	}

	// **Dem Wirt ist erst der Beginn erlaubt** — CR 103.8 (Fassung
	// 2026-09-25), auch an dieser Naht: die Starthand und, was jederzeit
	// geht (#1851).
	if !slices.Equal(angelegt.Erlaubt, []string{
		"spieler.starthand", "zug.aufgeben", "spieler.muenzwurf",
		"spieler.wuerfeln",
	}) {
		t.Errorf("dem Wirt ist erlaubt: %v", angelegt.Erlaubt)
	}

	fassung := beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)

	// **Ein Befehl wirkt und kommt mit seinem Ereignis zurück.**
	erstes := inDerBibliothek(t, gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1").
		Partie.Objekte, "sitz-1")

	gestellt := gebunden.PartieBefehl(t.Context(), angelegt.Partie.ID, "b1", "sitz-1",
		"objekt.schieben", map[string]any{"objekt": erstes, "zone": "hand"}, "")
	if gestellt.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", gestellt.Fehler)
	}

	if len(gestellt.Ereignisse) != 1 || gestellt.Ereignisse[0].Kette == "" {
		t.Fatalf("die Ereignisse: %+v", gestellt.Ereignisse)
	}

	if gestellt.Fassung != naechsteFassung(t, fassung) {
		t.Errorf("die Fassung ist %q nach %q", gestellt.Fassung, fassung)
	}

	// **Und die Sicht danach zeigt die Karte in der Hand ihres Sitzes.**
	danach := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1")
	if danach.Fehler != nil {
		t.Fatalf("Partie: %+v", danach.Fehler)
	}

	for _, objekt := range danach.Partie.Objekte {
		if objekt.Kennung != erstes {
			continue
		}

		if objekt.Zone != "hand" || objekt.Karte == "" {
			t.Errorf("das eigene Objekt: %+v", objekt)
		}
	}
}

// **Das Gegenüber sieht die Hand nicht — über dieselbe Naht.**
//
// **Das ist die Zusage aus `domain`, hier an der Fläche gemessen**: wer nur
// die Projektion prüft, hat nicht geprüft, dass die Naht sie auch benutzt.
func TestUeberDieNahtSiehtDasGegenueberKeineHand(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)

	erstes := inDerBibliothek(t, gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1").
		Partie.Objekte, "sitz-1")

	if gestellt := gebunden.PartieBefehl(t.Context(), angelegt.Partie.ID, "b1", "sitz-1",
		"objekt.schieben", map[string]any{"objekt": erstes, "zone": "hand"},
		""); gestellt.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", gestellt.Fehler)
	}

	fremd := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-2")
	if fremd.Fehler != nil {
		t.Fatalf("Partie: %+v", fremd.Fehler)
	}

	// **Die ganze Hand und nicht nur die geschobene Karte** — seit #1851
	// bringt die Starthand sieben mit, und die sind genauso verborgen.
	verborgen := 0

	for _, objekt := range fremd.Partie.Objekte {
		if objekt.Beherrscher == "sitz-1" && objekt.Zone == "hand" {
			verborgen++

			if objekt.Karte != "" || !objekt.Geheim {
				t.Errorf("das Gegenüber liest die Handkarte: %+v", objekt)
			}
		}

		// **Und keine Bibliotheksordnung kommt mit** — für keinen von beiden.
		if objekt.Zone == "bibliothek" && objekt.Ordnung != 0 {
			t.Errorf("die Ordnung steht da: %+v", objekt)
		}
	}

	if verborgen != domain.Starthand+1 {
		t.Errorf("das Gegenüber sieht %d Handkarten, erwartet %d",
			verborgen, domain.Starthand+1)
	}
}

// **Pausieren, fortsetzen, beenden — und was danach nicht mehr geht.**
func TestDerLebenszyklusUeberDieNaht(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	kennung := angelegt.Partie.ID

	pausiert := gebunden.PartiePausieren(t.Context(), kennung, "sitz-1")
	if pausiert.Fehler != nil || pausiert.Partie.Stand != "pausiert" {
		t.Fatalf("PartiePausieren: %+v", pausiert)
	}

	// **In der Pause ist nichts erlaubt** — eine leere Liste, kein `nil`.
	if pausiert.Erlaubt == nil || len(pausiert.Erlaubt) != 0 {
		t.Errorf("in der Pause sind %v erlaubt", pausiert.Erlaubt)
	}

	fortgesetzt := gebunden.PartieFortsetzen(t.Context(), kennung, "sitz-1")
	if fortgesetzt.Fehler != nil || fortgesetzt.Partie.Stand != "laeuft" {
		t.Fatalf("PartieFortsetzen: %+v", fortgesetzt)
	}

	beendet := gebunden.PartieBeenden(t.Context(), kennung, "sitz-1")
	if beendet.Fehler != nil || beendet.Partie.Stand != "beendet" {
		t.Fatalf("PartieBeenden: %+v", beendet)
	}

	// **Eine beendete Partie kommt nicht zurück.**
	if nochmal := gebunden.PartieFortsetzen(t.Context(),
		kennung, "sitz-1"); nochmal.Fehler == nil {
		t.Error("die beendete Partie ließ sich fortsetzen")
	}
}

// **Was über die Naht nicht durchkommt.**
func TestWasUeberDieNahtNichtDurchkommt(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	faelle := map[string]PartieentwurfDTO{
		"mit einem Deck, das keine Kennung ist": {
			Spiel: "magic",
			Sitze: []SitzwunschDTO{
				{Kennung: "sitz-1", Name: "David", DeckID: "kein deck"},
				{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
			},
		},
		"mit einem Tisch ohne jeden Sitz": {
			Spiel: "magic",
			Sitze: []SitzwunschDTO{},
		},
		"ohne Spiel": {
			Sitze: []SitzwunschDTO{
				{Kennung: "sitz-1", Name: "David", DeckID: deckID},
				{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
			},
		},
	}

	for name, entwurf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := gebunden.PartieAnlegen(t.Context(), entwurf)
			if antwort.Fehler == nil {
				t.Fatal("der Entwurf kam durch")
			}

			// **Und der Fehlerweg trägt leere Listen** — dieselbe Zusage wie
			// der Erfolgsweg. Genau hier ist sie in Phase 5.4 zweimal
			// gebrochen (N1, N4).
			if antwort.Erlaubt == nil || antwort.Partie.Objekte == nil ||
				antwort.Partie.Sitze == nil || antwort.Partie.Zuschauer == nil {
				t.Errorf("der Fehlerweg gibt nil: %+v", antwort)
			}
		})
	}

	// **Und eine Partie, die es nicht gibt.**
	if antwort := gebunden.Partie(t.Context(),
		"gibtesnicht", "sitz-1"); antwort.Fehler == nil {
		t.Error("eine erfundene Partie kam durch")
	}
}

// **Eine erwartete Fassung, die nicht mehr gilt, ist ein Konflikt.**
//
// **Über die Naht gemessen**, denn dort kommt sie als **Text** herein — und
// was keine Zahl ist, heißt null und damit „gegen die, die gerade gilt".
func TestDieErwarteteFassungUeberDieNaht(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	kennung := angelegt.Partie.ID
	fassung := beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)
	objekt := inDerBibliothek(t, gebunden.Partie(t.Context(), kennung, "sitz-1").
		Partie.Objekte, "sitz-1")

	// **Mit der richtigen Fassung: durch.**
	if erste := gebunden.PartieBefehl(t.Context(), kennung, "b1", "sitz-1",
		"objekt.tappen", map[string]any{"objekt": objekt},
		fassung); erste.Fehler != nil {
		t.Fatalf("mit Fassung %s: %+v", fassung, erste.Fehler)
	}

	// **Mit einer überholten: Konflikt.**
	zweite := gebunden.PartieBefehl(t.Context(), kennung, "b2", "sitz-1",
		"objekt.enttappen", map[string]any{"objekt": objekt}, fassung)
	if zweite.Fehler == nil {
		t.Fatal("die überholte Fassung kam durch")
	}

	// **Und was keine Zahl ist, heißt null** — und kommt damit durch.
	if dritte := gebunden.PartieBefehl(t.Context(), kennung, "b3", "sitz-1",
		"objekt.enttappen", map[string]any{"objekt": objekt},
		"keine Zahl"); dritte.Fehler != nil {
		t.Fatalf("ohne Fassung: %+v", dritte.Fehler)
	}
}
