package service_test

// **Der Tisch für die Phasenabnahme** — die Vorrichtung zu
// `partieabnahme_test.go`
// (#1593).
//
// # Warum sie eine eigene Datei hat
//
// **Weil sie eine andere Frage beantwortet** (§2): nebenan steht, *was* eine
// ganze Partie beweist; hier steht, *womit* sie gespielt wird — zwei Decks,
// ein gestellter Tisch und der Weg, auf dem ein Befehl hineingeht.

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// handgriffe sind die Befehle, mit denen ein Mensch eine Regel **von Hand**
// durchsetzt.
//
// **Sie sind der Sandkasten** (`G1`): Karten schieben, Leben setzen, Schaden
// markieren, tappen, zerstören. Sie gehören dazu — ohne sie ließe sich
// nichts spielen, was das Programm noch nicht kann. **In der Abnahme kommt
// keiner vor**, und das ist ihre Aussage.
//
// # Was ausdrücklich keiner ist
//
// **Eine Wahl, die das Regelwerk dem Spieler lässt.** Abwerfen auf die
// Handgrenze (CR 514.1), die Legendenwahl (CR 704.5j), die Reihung
// gleichzeitiger Auslöser (CR 603.3b), der Weg zurück in die Kommandozone
// (CR 903.9a), der Blick in die eigene Bibliothek samt Mischpflicht und das
// Aufgeben sind **Spielhandlungen** und keine Durchsetzung von Hand. Sie
// stehen deshalb nicht in dieser Liste — auch wenn dieser Lauf keine davon
// braucht.
var handgriffe = []string{
	// **Der Sandkasten aus `G1`** — Buchführung am Objekt und am Sitz.
	service.BefehlSchieben, service.BefehlLeben, service.BefehlSchaden,
	service.BefehlTappen, service.BefehlEnttappen, service.BefehlUmdrehen,
	service.BefehlZaehler, service.BefehlAnhaengen, service.BefehlLoesen,
	service.BefehlMana, service.BefehlVergiftung,
	service.BefehlZugende, service.BefehlAbkuerzen,

	// **Die Handlungen aus CR 701** (`M7`) — sie gehören auf eine Karte,
	// und bis es sie dort gibt, stellt sie ein Mensch.
	service.BefehlZerstoeren, service.BefehlOpfern, service.BefehlExilieren,
	service.BefehlNeutralisieren, service.BefehlKaempfen,
	service.BefehlMahlen, service.BefehlHellsehen, service.BefehlUeberwachen,

	// **Die Gestalten und Bezeichnungen** (`M12`, `M20`) — dasselbe: was
	// sie erzeugt oder vergibt, steht auf einer Karte.
	service.BefehlSpielstein, service.BefehlEmblem,
	service.BefehlBeherrschen, service.BefehlUmwandeln,
	service.BefehlKopieren, service.BefehlMonarch,
	service.BefehlInitiative, service.BefehlStrahlung,
}

// abnahmedeck ist ein Deck aus zwei Karten — ein Gebirge und ein Geschöpf.
func abnahmedeck(name, staerke, widerstand string) domain.Partiedeck {
	return domain.Partiedeck{
		Abdruck: "abnahme",
		Karten: []domain.Abzugskarte{{
			Kennung: "gebirge", Name: "Gebirge", Menge: 21,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name":       "Gebirge",
				"super type": "Basic Land",
				"sub type":   "Mountain",
			}}},
		}, {
			Kennung: "wesen", Name: name, Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name":         name,
				"super type":   "Creature",
				"sub type":     "Beast",
				"casting cost": "{R}",
				"power":        staerke,
				"toughness":    widerstand,
			}}},
		}},
	}
}

// eineAbnahmepartie stellt den Tisch: je zwei Karten auf der Hand, der Rest
// in der Bibliothek.
//
// **Gestellt und nicht gemischt** — der Grund steht im Kopf von
// `partieabnahme_test.go`.
func eineAbnahmepartie() domain.Partie {
	// **Vierundvierzig**: je Sitz zwei auf der Hand und zwanzig in der
	// Bibliothek.
	objekte := make([]domain.Spielobjekt, 0, 44)

	for _, sitz := range []string{"sitz-1", "sitz-2"} {
		objekte = append(objekte,
			domain.Spielobjekt{
				Kennung: sitz + "/gebirge/0", Art: domain.ObjektKarte,
				Beherrscher: sitz, Eigner: sitz,
				Zone: domain.ZoneHand, Karte: "gebirge",
			},
			domain.Spielobjekt{
				Kennung: sitz + "/wesen/0", Art: domain.ObjektKarte,
				Beherrscher: sitz, Eigner: sitz,
				Zone: domain.ZoneHand, Karte: "wesen",
			})

		// **Genug zum Ziehen** — eine leere Bibliothek verliert die Partie
		// (CR 704.5b), und das wäre ein anderes Ende als das gemeinte.
		for lauf := range 20 {
			objekte = append(objekte, domain.Spielobjekt{
				Kennung: fmt.Sprintf("%s/gebirge/%d", sitz, lauf+1),
				Art:     domain.ObjektKarte, Beherrscher: sitz, Eigner: sitz,
				Zone: domain.ZoneBibliothek, Karte: "gebirge",
				Ordnung: lauf + 1,
			})
		}
	}

	return domain.Partie{
		ID:    "abnahme",
		Spiel: "magic",
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Profil: domain.Grundprofil("magic"),
		Wirt:   "sitz-1",
		Stand:  domain.StandLaeuft,
		Sitze: []domain.Sitz{
			{
				Kennung: "sitz-1", Name: "Anna",
				Deck: abnahmedeck("Koloss", "20", "20"),
			},
			{
				Kennung: "sitz-2", Name: "Bert",
				Deck: abnahmedeck("Bär", "2", "2"),
			},
		},
		Objekte: objekte,
	}.MitStartleben().ZugBeginnen("sitz-1")
}

// abnahmelauf führt die Partie und schreibt mit, was gestellt wurde.
type abnahmelauf struct {
	t        *testing.T
	leitung  *service.Partieleitung
	speicher *partieprobe
	benutzt  []string
	zaehler  int

	// nachDemBefehl ist, was die Anwendung nach einem geglückten Befehl tut
	// — den Tisch benachrichtigen (#1627). **Nil heißt: dieser Lauf hat
	// keinen Tisch**, und das ist der Normalfall der Läufe hier.
	nachDemBefehl func(domain.Spielbefehl, service.Befehlsausgang)
}

// einAbnahmelauf legt die Partie an und gibt den Lauf darüber.
func einAbnahmelauf(t *testing.T) *abnahmelauf {
	t.Helper()

	leitung, speicher := eineLeitung(t)

	if _, fehler := leitung.Anlegen(
		context.Background(), eineAbnahmepartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	return &abnahmelauf{t: t, leitung: leitung, speicher: speicher}
}

// partie gibt den **vollen** Stand.
//
// # Warum aus dem Speicher und nicht aus der Sicht
//
// **Weil eine Sicht die Bibliothek verbirgt** — das ist ihre Aufgabe (`G1`),
// und ein Lauf, der nachzählt, ob jemand gezogen hat, sähe sie nicht. Die
// **Befehle** gehen trotzdem durch die Naht; gelesen wird hier die Wahrheit,
// gegen die sie gerechnet haben.
func (a *abnahmelauf) partie() domain.Partie {
	a.t.Helper()

	partie, da := a.speicher.partien["abnahme"]
	if !da {
		a.t.Fatal("die Partie steht nicht im Speicher")
	}

	return partie
}

// stelle gibt einen Befehl — und prüft vorher, dass er angeboten war.
//
// **Die Prüfung davor ist die halbe Aussage des Laufs** (§16): ein Befehl,
// den [service.Partieleitung.VerfuegbareAktionen] nicht nennt, hat im
// Fenster keinen Knopf. Wer ihn hier stellte, prüfte einen Weg, den niemand
// gehen kann (#1623).
func (a *abnahmelauf) stelle(wer, art string, nutzlast map[string]any) {
	a.t.Helper()

	ctx := context.Background()

	erlaubt, fehler := erlaubtVon(a.leitung, ctx, "abnahme", wer)
	if fehler != nil {
		a.t.Fatalf("Erlaubnis: %v", fehler)
	}

	if !slices.Contains(erlaubt, art) {
		a.t.Fatalf("%q ist für %q nicht angeboten — angeboten ist %v",
			art, wer, erlaubt)
	}

	a.zaehler++

	if !slices.Contains(a.benutzt, art) {
		a.benutzt = append(a.benutzt, art)
	}

	befehl := domain.Spielbefehl{
		Kennung:  fmt.Sprintf("b%d", a.zaehler),
		PartieID: "abnahme",
		Absender: wer,
		Art:      art,
		Nutzlast: nutzlast,
	}

	ausgang, fehler := a.leitung.Befehl(ctx, befehl)
	if fehler != nil {
		a.t.Fatalf("%q von %q: %v", art, wer, fehler)
	}

	// **Was die Anwendung danach tut, tut dieser Lauf auch.**
	// `App.PartieBefehl` trägt das Ergebnis an den Tisch (#1627); ein Lauf,
	// der es ausließe, prüfte einen Weg, den es nicht gibt.
	if a.nachDemBefehl != nil {
		a.nachDemBefehl(befehl, ausgang)
	}
}

// passeBis lässt passen, bis die Bedingung hält.
//
// **Wer passt, ist der mit der Priorität** — und niemand sonst. Eine
// Partie ohne Priorität steht auf einer zugbasierten Handlung oder auf einer
// offenen Frage; wer dort passte, prüfte einen Weg, den das Regelwerk nicht
// kennt.
func (a *abnahmelauf) passeBis(wohin string, fertig func(domain.Partie) bool) {
	a.t.Helper()

	for range 200 {
		partie := a.partie()

		if fertig(partie) {
			return
		}

		if partie.Zug.Prioritaet == "" {
			a.t.Fatalf("auf dem Weg zu %s hat niemand Priorität: "+
				"Zug %d, %s/%s", wohin, partie.Zug.Nummer,
				partie.Zug.Abschnitt.Phase, partie.Zug.Abschnitt.Schritt)
		}

		a.stelle(partie.Zug.Prioritaet, service.BefehlPassen, nil)
	}

	a.t.Fatalf("nach 200 Schritten steht %s immer noch aus", wohin)
}
