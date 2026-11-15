package app

// **Die vier weiteren Arten im Umschlag** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// # Was hier gemessen wird
//
// **Dass ein Deck hinaus und wieder herein geht** — mit seinen Plätzen und
// mit **neuen** Kennungen: die Zusage aus `N4` gilt für jede Art.
//
// **Dass eine Sammlungssicht angesehen und nicht angenommen wird.** Sie sagt,
// was ein Gegenüber hat; sie in die eigene Sammlung zu schreiben hieße,
// fremde Karten als eigene zu führen. Das ist die Zusage dieses Werkstücks:
// „Empfang/Import ist keine physische Abbuchung."
//
// **Und dass eine unbekannte Art abgewiesen wird**, statt still als Set
// durchzugehen.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// einUmschlagsziel gibt einen Ort für einen Umschlag.
func einUmschlagsziel(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(t.TempDir(), name+".msepackage")
}

// mitIdentitaet stellt eine Anwendung samt Peeridentität.
func mitIdentitaet(t *testing.T) *App {
	t.Helper()

	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(t.Context())); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	return anwendung
}

// **Ein Deck geht hinaus und wieder herein.**
func TestEinDeckGehtAlsUmschlagHinausUndWiederHerein(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	angelegt := anwendung.DeckAnlegen(ctx, "Wanderdeck", "magic", "eine Notiz")
	if angelegt.Fehler != nil {
		t.Fatalf("Deck anlegen: %+v", angelegt.Fehler)
	}

	// **Eine echte Druckkennung** — `Deckplatz` prüft sie, und eine erfundene
	// wäre ein Test gegen eine Welt, die es nicht gibt (§3).
	const derDruck = "a1b2c3d4-0000-4000-8000-00000000000a"

	gelegt := anwendung.DeckplatzLegen(ctx, angelegt.Deck.ID, "", derDruck,
		4, "hauptdeck")
	if gelegt.Fehler != nil {
		t.Fatalf("Platz legen: %+v", gelegt.Fehler)
	}

	ziel := einUmschlagsziel(t, "deck")

	geschnuert := anwendung.UmschlagSchnueren(ctx, "deck", angelegt.Deck.ID, ziel)
	if geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	if len(geschnuert.Objekte) != 1 || geschnuert.Objekte[0].Art != "deck" {
		t.Fatalf("der Umschlag führt %+v", geschnuert.Objekte)
	}

	// **Angenommen wird es als neues Deck** — ein Umschlag von außen
	// überschreibt nichts.
	angenommen := anwendung.UmschlagDeck(ctx, ziel)
	if angenommen.Fehler != nil {
		t.Fatalf("annehmen: %+v", angenommen.Fehler)
	}

	if angenommen.Deck == nil {
		t.Fatal("es kam kein Deck")
	}

	if angenommen.Deck.ID == angelegt.Deck.ID {
		t.Error("das angenommene Deck trägt dieselbe Kennung")
	}

	// **Der Name trägt einen sichtbaren Zusatz** — hier steht schon ein Deck
	// dieses Namens, und zwei gleich heißende nebeneinander wären
	// schlimmer.
	if angenommen.Deck.Name == "Wanderdeck" {
		t.Error("das angenommene Deck heißt wie das vorhandene")
	}

	// **Die Notiz bleibt hier** (FMT-05.AK-04, #1694): sie ist privat, und ein
	// Umschlag geht an andere. Bis dahin reiste sie ungefragt mit — samt der
	// Herkunft des Decks, die Profilnamen anderer Menschen nennt.
	if angenommen.Deck.Notiz != "" || angenommen.Deck.Herkunft.Quelle != "" {
		t.Errorf("Privates reiste mit: Notiz %q, Herkunft %+v", angenommen.Deck.Notiz, angenommen.Deck.Herkunft)
	}

	if len(angenommen.Deck.Plaetze) != 1 {
		t.Fatalf("es kamen %d Plätze", len(angenommen.Deck.Plaetze))
	}

	if angenommen.Deck.Plaetze[0].Menge != 4 ||
		angenommen.Deck.Plaetze[0].Printing != derDruck {
		t.Errorf("der Platz ist %+v", angenommen.Deck.Plaetze[0])
	}
}

// **Eine Sammlungssicht wird angesehen und nicht angenommen.**
func TestEineSammlungssichtWirdAngesehenUndNichtAngenommen(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := anwendung.imProfil(t.Context())

	fremd, erzeugt := peer.Erzeugen(nil)
	if erzeugt != nil {
		t.Fatalf("Erzeugen: %v", erzeugt)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(ctx, string(fremd.Kennung()),
		"Turfer"); antwort.Fehler != nil {
		t.Fatalf("Kontakt: %+v", antwort.Fehler)
	}

	// **Ohne Erlaubnis für den Zweck gibt es gar keine Sicht** (T3) — der
	// Kontakt muss bestätigt sein und den Zweck haben.
	if antwort := anwendung.PeerkontaktBestaetigen(ctx,
		string(fremd.Kennung()), fremd.Fingerabdruck()); antwort.Fehler != nil {
		t.Fatalf("Bestätigen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerzweckErlauben(ctx, string(fremd.Kennung()),
		"handel"); antwort.Fehler != nil {
		t.Fatalf("Zweck: %+v", antwort.Fehler)
	}

	// **Und eine Freigabe** — ohne sie gibt es keine Sicht, auch nicht eine
	// leere: „nichts freigegeben" ist eine andere Auskunft als „nichts da".
	if antwort := anwendung.SammlungFreigeben(ctx, SammlungsfreigabeDTO{
		Empfaenger: string(fremd.Kennung()), Umfang: "sammlung",
		Zweck: "handel", Auswahl: []string{},
	}); antwort.Fehler != nil {
		t.Fatalf("Freigabe: %+v", antwort.Fehler)
	}

	ziel := einUmschlagsziel(t, "sicht")

	// **Ohne Freigabe sieht niemand etwas** — die Sicht ist dann leer, und
	// das ist eine Auskunft und kein Fehler (T3).
	geschnuert := anwendung.UmschlagSchnueren(ctx, "sammlungssicht",
		string(fremd.Kennung())+"/handel", ziel)
	if geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	if len(geschnuert.Objekte) != 1 ||
		geschnuert.Objekte[0].Art != "sammlungssicht" {
		t.Fatalf("der Umschlag führt %+v", geschnuert.Objekte)
	}

	angesehen := anwendung.UmschlagSicht(ctx, ziel)
	if angesehen.Fehler != nil {
		t.Fatalf("ansehen: %+v", angesehen.Fehler)
	}

	if angesehen.Zeilen == nil {
		t.Error("die Zeilen sind nil statt leer")
	}

	// **Und die Sammlung hat sich nicht gerührt.**
	bestand := anwendung.Sammlung(ctx)
	if bestand.Fehler != nil {
		t.Fatalf("Sammlung: %+v", bestand.Fehler)
	}

	if len(bestand.Exemplare) != 0 {
		t.Errorf("die Sammlung führt %d Zeilen — ein Empfang hat gebucht",
			len(bestand.Exemplare))
	}
}

// **Eine Art, die es nicht gibt, kommt nicht durch.**
func TestEineUnbekannteUmschlagartWirdAbgewiesen(t *testing.T) {
	anwendung := mitIdentitaet(t)

	antwort := anwendung.UmschlagSchnueren(t.Context(), "tanzstunde", "1",
		einUmschlagsziel(t, "fremd"))
	if antwort.Fehler == nil {
		t.Fatal("die fremde Art kam durch")
	}

	if antwort.Objekte == nil {
		t.Error("die Liste ist nil")
	}
}
