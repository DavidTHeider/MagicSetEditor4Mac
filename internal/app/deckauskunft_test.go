package app_test

// **Prüfung, Abgleich und Formate an der Naht** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `decknaht_test.go` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie ein Deck über die
// Naht geht, hier, was man darüber erfährt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Urteil kommt mit seinen fünf Wörtern — unübersetzt.**
func TestEinUrteilKommtMitSeinenWoertern(t *testing.T) {
	anwendung, _, _, pruefung, _ := deckanwendung(t)

	pruefung.urteil = domain.Deckurteil{
		Formatname: "Standard", Hauptdeck: 60, Sideboard: 15,
		Urteile: []domain.Kartenurteil{
			{
				PlatzID: 1, Bezeichnung: "Blitz", Menge: 4,
				Bereich: domain.BereichHauptdeck, Zulaessigkeit: domain.ZulaessigLegal,
				Quelle: "Scryfall",
			},
			{
				PlatzID: 2, Menge: 3, Bereich: domain.BereichHauptdeck,
				Zulaessigkeit: domain.ZulaessigUngeprueft, Grund: "eigener Entwurf",
			},
		},
		Groessen: []domain.Groessenbefund{{Feld: "hauptdeck", Meldung: "zu wenige"}},
	}
	pruefung.befunde = []domain.Befund{{Meldung: "Scryfall schwieg"}}

	antwort := anwendung.DeckPruefen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Urteil.Urteile[0].Zulaessigkeit != "legal" ||
		antwort.Urteil.Urteile[1].Zulaessigkeit != "ungeprüft" {
		t.Errorf("%+v", antwort.Urteil.Urteile)
	}

	// **`ungeprüft` steht dem Deck nicht entgegen** — hier tut es der
	// Größenbefund.
	if antwort.Urteil.Zulaessig {
		t.Error("ein Deck mit Größenbefund gilt als zulässig")
	}

	if antwort.Urteil.Ungeprueft != 3 {
		t.Errorf("Ungeprueft %d, erwartet 3", antwort.Urteil.Ungeprueft)
	}

	// **Ein Befund ist kein Fehler** — die Prüfung lief weiter.
	if len(antwort.Befunde) != 1 || antwort.Fehler != nil {
		t.Errorf("Befunde %v, Fehler %+v", antwort.Befunde, antwort.Fehler)
	}

	if antwort.Urteil.Groessen[0].Feld != "hauptdeck" {
		t.Errorf("%+v", antwort.Urteil.Groessen)
	}
}

// **Ein Abgleich trennt Fehlbestand von Unbekanntem.**
func TestEinAbgleichTrenntFehlendVonUnbekannt(t *testing.T) {
	anwendung, _, _, _, abgleich := deckanwendung(t)

	abgleich.stand = service.Habenstand{
		Zeilen: []service.Habenzeile{
			{Printing: "a", Bezeichnung: "Blitz", Gebraucht: 4, Vorhanden: 2, Fehlend: 2},
			{Gebraucht: 3, Grund: "eigener Entwurf"},
		},
		Fehlend: 2, Unbekannt: 3,
	}

	antwort := anwendung.DeckAbgleichen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Fehlend != 2 || antwort.Unbekannt != 3 {
		t.Errorf("Fehlend %d, Unbekannt %d", antwort.Fehlend, antwort.Unbekannt)
	}

	if len(antwort.Zeilen) != 2 || antwort.Zeilen[1].Grund == "" {
		t.Errorf("%+v", antwort.Zeilen)
	}
}

// **Die drei Bauarten kommen mit ihren Zahlen.**
func TestDieBauartenKommenMitZahlenAnDieNaht(t *testing.T) {
	anwendung, _, _, _, _ := deckanwendung(t)

	antwort := anwendung.Bauarten()

	if len(antwort.Bauarten) != 3 {
		t.Fatalf("%d Bauarten, erwartet 3", len(antwort.Bauarten))
	}

	for _, bauart := range antwort.Bauarten {
		if bauart.Bauart == "" || bauart.Format.Mindestgroesse == 0 {
			t.Errorf("%+v", bauart)
		}
	}
}

// **Ein Format entsteht aus einer Bauart und kommt als Text zurück.**
func TestEinFormatAusEinerBauartAnDerNaht(t *testing.T) {
	anwendung, _, _, _, _ := deckanwendung(t)

	antwort := anwendung.FormatAusBauart(
		probekontext(t), "commander", "Commander", "magic", "commander")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Format.ID != "5" || antwort.Format.Mindestgroesse != 100 ||
		antwort.Format.HoechstzahlGleich != 1 {
		t.Errorf("%+v", antwort.Format)
	}
}

// **`FormatSetzen` legt an, wenn keine Kennung dasteht — und ändert sonst.**
func TestFormatSetzenLegtAnUndAendert(t *testing.T) {
	anwendung, _, formate, _, _ := deckanwendung(t)

	angelegt := anwendung.FormatSetzen(t.Context(), app.FormatDTO{
		Name: "Eigen", Spiel: "magic", Mindestgroesse: 40,
	})
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if angelegt.Format.ID != "5" || angelegt.Format.Fassung != "1" {
		t.Errorf("%+v", angelegt.Format)
	}

	geaendert := anwendung.FormatSetzen(t.Context(), app.FormatDTO{
		ID: "5", Fassung: "1", Name: "Eigen", Spiel: "magic", Mindestgroesse: 50,
	})
	if geaendert.Fehler != nil {
		t.Fatalf("%+v", geaendert.Fehler)
	}

	if geaendert.Format.Fassung != "2" || formate.format.Mindestgroesse != 50 {
		t.Errorf("%+v", geaendert.Format)
	}
}

// **Und eine unbrauchbare Kennung im DTO nennt ihr Feld.**
func TestEineUnbrauchbareFormatkennungImDTO(t *testing.T) {
	anwendung, _, _, _, _ := deckanwendung(t)

	for name, dto := range map[string]app.FormatDTO{
		"Kennung": {ID: "x", Name: "F", Spiel: "magic"},
		"Fassung": {ID: "5", Fassung: "x", Name: "F", Spiel: "magic"},
	} {
		t.Run(name, func(t *testing.T) {
			if anwendung.FormatSetzen(t.Context(), dto).Fehler == nil {
				t.Error("die unbrauchbare Kennung kam durch")
			}
		})
	}
}

// **Ein Format, dessen Speicher schweigt, meldet sich auf beiden Wegen.**
func TestFormatSetzenMeldetSeinenNahtfehler(t *testing.T) {
	anwendung, _, formate, _, _ := deckanwendung(t)

	formate.fehler = errSchweigt

	if anwendung.FormatSetzen(t.Context(), app.FormatDTO{
		Name: "F", Spiel: "magic",
	}).Fehler == nil {
		t.Error("anlegen schwieg")
	}

	if anwendung.FormatSetzen(t.Context(), app.FormatDTO{
		ID: "5", Fassung: "1", Name: "F", Spiel: "magic",
	}).Fehler == nil {
		t.Error("ändern schwieg")
	}
}

// **Und eine Bauart, die es nicht gibt, ebenso.**
func TestEineUnbekannteBauartAnDerNaht(t *testing.T) {
	anwendung, _, _, _, _ := deckanwendung(t)

	if anwendung.FormatAusBauart(
		probekontext(t), "turnier", "T", "magic", "").Fehler == nil {
		t.Error("die erfundene Bauart kam durch")
	}
}

// errSchweigt ist der Fehler, den eine Naht in diesen Läufen gibt.
var errSchweigt = errors.New("die Naht schweigt")

// **Die Erfolgswege, die bisher nur im Fehlerfall betreten wurden.**
//
// **Ohne sie stünde in der Abdeckung ein Zweig, den niemand geht** — und ein
// Weg, der nur scheitert, ist keine Zusage über das Gelingen (§3).
func TestDieErfolgswegeDerDecknaht(t *testing.T) {
	anwendung, decks, formate, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{ID: 7, Fassung: 1, Name: "Aggro", Spiel: "magic"}
	formate.format = domain.Format{ID: 5, Fassung: 1, Name: "Standard", Spiel: "magic"}

	angelegt := anwendung.DeckAnlegen(t.Context(), "Aggro", "magic", "Notiz")
	if angelegt.Fehler != nil || angelegt.Deck.ID != "7" {
		t.Errorf("anlegen: %+v", angelegt)
	}

	liste := anwendung.Decks(t.Context(), "magic")
	if liste.Fehler != nil || len(liste.Decks) != 1 {
		t.Errorf("die Liste: %+v", liste)
	}

	geloescht := anwendung.DeckLoeschen(t.Context(), "7")
	if geloescht.Fehler != nil {
		t.Errorf("löschen: %+v", geloescht.Fehler)
	}

	genommen := anwendung.DeckplatzNehmen(t.Context(), "7", "3")
	if genommen.Fehler != nil {
		t.Errorf("nehmen: %+v", genommen.Fehler)
	}

	// **Eine eigene Karte legen** — der andere Zweig von `DeckplatzLegen`.
	gelegt := anwendung.DeckplatzLegen(t.Context(), "7", "9", "", 1, "sideboard")
	if gelegt.Fehler != nil || gelegt.Platz.KarteID != "9" {
		t.Errorf("legen: %+v", gelegt)
	}

	gewaehlt := anwendung.DeckFormatWaehlen(t.Context(), "7", "5")
	if gewaehlt.Fehler != nil || gewaehlt.Deck.FormatID != "5" {
		t.Errorf("wählen: %+v", gewaehlt)
	}

	formatliste := anwendung.Formate(t.Context(), "magic")
	if formatliste.Fehler != nil || len(formatliste.Formate) != 1 {
		t.Errorf("die Formatliste: %+v", formatliste)
	}

	if name := formatliste.Formate[0].Name; name != "Standard" {
		t.Errorf("Name %q", name)
	}
}

// **Ein Umbenennen, das erst liest und dann scheitert.**
//
// `DeckUmbenennen` liest den Stand, bevor es schreibt — sonst verlöre es die
// Formatwahl. Scheitert **beides**, wird der Schreibweg nie betreten, und sein
// Fehlerzweig bliebe unbetreten.
func TestEinUmbenennenScheitertNachDemLesen(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{ID: 7, Fassung: 1, Name: "Alt", Spiel: "magic"}
	decks.umbenennfehler = errSchweigt

	if anwendung.DeckUmbenennen(t.Context(), "7", "1", "Neu", "").Fehler == nil {
		t.Error("der Schreibfehler ging verloren")
	}
}
