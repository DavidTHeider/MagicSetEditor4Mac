package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Fehlerwege der Stilauswahl.
//
// Jeder von ihnen ist im Betrieb erreichbar: ein Bestand, der gerade nicht
// lesbar ist, ein Paketverzeichnis, das jemand geräumt hat, ein Set, dessen
// Spiel nicht installiert ist.

// **Ein Fehler des Bestands kommt durch — auf jedem Weg.**
func TestDerBestandsfehlerKommtDurchDieStilwahl(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")
	kopf := domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"}

	// **Der Katalog scheitert**, die Köpfe stehen.
	mitKatalogfehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{kopf}},
		&ablage{fehler: kaputt}, &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)
	if _, err := mitKatalogfehler.Stile(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("Stile: %v", err)
	}

	// **Die Köpfe scheitern** — dann kommt gar nichts.
	mitKopffehler := service.NeuerStilwaehler(
		&koepfequelle{fehler: kaputt},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := mitKopffehler.Stile(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("Stile: %v", err)
	}

	if _, err := mitKopffehler.Optionen(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("Optionen: %v", err)
	}

	if _, err := mitKopffehler.Folgen(t.Context(), 1, "probe-schlicht"); !errors.Is(err, kaputt) {
		t.Errorf("Folgen: %v", err)
	}

	if err := mitKopffehler.OptionSetzen(t.Context(), 1, "hintergrund", "hell"); !errors.Is(err, kaputt) {
		t.Errorf("OptionSetzen: %v", err)
	}

	// **Die Stilwerte scheitern** — auf allen drei Wegen, die sie lesen.
	mitWertefehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{kopf}},
		saatkatalog(), &stilablage{fehler: kaputt}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := mitWertefehler.Optionen(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("Optionen: %v", err)
	}

	if _, err := mitWertefehler.Folgen(t.Context(), 1, "probe-schlicht"); !errors.Is(err, kaputt) {
		t.Errorf("Folgen: %v", err)
	}

	if err := mitWertefehler.Wechseln(t.Context(), 1, "probe-schlicht"); !errors.Is(err, kaputt) {
		t.Errorf("Wechseln: %v", err)
	}

	// **`OptionSetzen` steht hier nicht mehr**, und das ist kein Versehen:
	// seit dem Zweitblick über Phase L liest es die Werte gar nicht, sondern
	// schreibt die eine Option in **einer** Anweisung. Ein Lesefehler kann ihm
	// deshalb nicht mehr begegnen — sein Fehlerweg ist der Schreibfehler, und
	// den prüft `stilvorgabezustand_test.go`.
}

// **Ein Stil, den es nicht gibt, wird auf jedem Weg benannt.**
func TestEinFehlenderStilWirdAufJedemWegBenannt(t *testing.T) {
	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "niegesehen"},
		&stilablage{}, &feldliste{})

	if _, err := waehle.Optionen(t.Context(), 1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Optionen: %v", err)
	}

	if _, err := waehle.Folgen(t.Context(), 1, "auchnicht"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Folgen: %v", err)
	}

	if err := waehle.OptionSetzen(t.Context(), 1, "hintergrund", "hell"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("OptionSetzen: %v", err)
	}
}

// **Ein Spielpaket, das nicht installiert ist, wird benannt.**
//
// Der Fall entsteht bei einem Stil, dessen Spiel fehlt — häufiger, als man
// denkt: ein Stilpaket ohne sein Spiel ist eine halbe Installation.
func TestEinFehlendesSpielpaketHaeltDieFolgenAuf(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{{
		Art: "mse-style", Name: "ohnespiel", Pfad: "/p/ohnespiel.mse-style",
		Braucht: []domain.Bedarf{{Paket: "gibtsnicht.mse-game"}},
	}}}

	// Der Stil selbst liegt in der Saat — sein Spiel nicht.
	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "anderes", Style: "probe-schlicht"}}},
		katalog, &stilablage{}, &feldliste{felder: []string{"name"}}, t.TempDir(),
	)

	if _, err := waehle.Folgen(t.Context(), 1, "probe-schlicht"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Folgen: %v", err)
	}
}

// **Ein Fehler beim Lesen der Feldnamen kommt durch.**
func TestEinFehlerBeiDenFeldnamenKommtDurch(t *testing.T) {
	kaputt := errors.New("die Feldnamen sind nicht lesbar")
	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "anderes", Style: "fremd"}}},
		saatkatalog(), &stilablage{}, &feldliste{fehler: kaputt}, festeSaatwurzel(t),
	)

	if _, err := waehle.Folgen(t.Context(), 1, "probe-schlicht"); !errors.Is(err, kaputt) {
		t.Errorf("Folgen: %v", err)
	}
}

// **Ein Stil ohne `depends on:` nennt kein Spiel — und gilt damit als fremd.**
//
// Der Preis steht bei `spielAus`: die Angabe kommt aus dem Katalog, und ein
// Paket ohne Abhängigkeit hat sie nicht. Sichtbar bleibt der Stil trotzdem.
func TestEinStilOhneAbhaengigkeitGiltAlsFremd(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-style", Name: "einsam", Pfad: "/p/einsam.mse-style"},
		{
			Art: "mse-style", Name: "mitanderem", Pfad: "/p/mitanderem.mse-style",
			Braucht: []domain.Bedarf{{Paket: "irgendein-modul.mse-include"}},
		},
	}}
	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe"}}},
		katalog, &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	angebote, err := waehle.Stile(t.Context(), 1)
	if err != nil {
		t.Fatalf("Stile: %v", err)
	}

	if len(angebote) != 2 {
		t.Fatalf("Angebote: %+v", angebote)
	}

	for _, angebot := range angebote {
		if angebot.EigenesSpiel || angebot.Spiel != "" {
			t.Errorf("Angebot: %+v", angebot)
		}
	}
}

// **Die Ordnung gilt in beide Richtungen.**
//
// `slices.SortStableFunc` ruft den Vergleich in der Reihenfolge, die ihm passt;
// eine Liste, in der der fremde Stil zuerst steht, geht durch den anderen Zweig
// — und ohne diesen Fall bliebe er ungeprüft.
func TestDerEigeneStilKommtAuchVonHintenNachVorn(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{
		{
			Art: "mse-style", Name: "aaa-fremd", Pfad: "/p/aaa-fremd.mse-style",
			Braucht: []domain.Bedarf{{Paket: "anderes.mse-game"}},
		},
		{
			Art: "mse-style", Name: "zzz-eigen", Pfad: "/p/zzz-eigen.mse-style",
			Braucht: []domain.Bedarf{{Paket: "probe.mse-game"}},
		},
	}}
	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe"}}},
		katalog, &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	angebote, err := waehle.Stile(t.Context(), 1)
	if err != nil {
		t.Fatalf("Stile: %v", err)
	}

	if angebote[0].Paket != "zzz-eigen" {
		t.Errorf("Reihenfolge: %+v", angebote)
	}
}

// **Ein Stil, dessen Spiel nirgends liegt, hält die Folgen auf.**
//
// Der Weg geht dabei durch **zwei** Wurzeln: die erste hat den Stil, aber nicht
// das Spiel, die zweite keines von beidem. Ohne diesen Fall bliebe das
// Weitersuchen ungeprüft.
func TestEinStilOhneInstalliertesSpielHaeltDieFolgenAuf(t *testing.T) {
	wurzel := t.TempDir()
	paket(t, wurzel, "verwaist.mse-style", "mse version: 2.0.0\ngame: gibtsnicht\n")

	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "anderes", Style: "verwaist"}}},
		saatkatalog(), &stilablage{}, &feldliste{felder: []string{"name"}},
		wurzel, festeSaatwurzel(t),
	)

	_, err := waehle.Folgen(t.Context(), 1, "verwaist")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Folgen: %v", err)
	}
}

// **Ein Spiel, dessen Felder aus einem Modul kommen, wird trotzdem gelesen.**
//
// # Woher dieser Fall kommt
//
// Aus dem laufenden Fenster. `magic.mse-game` führt **kein** `card field:` im
// Kopf — seine Felder stehen in `include file: card_fields`. Ohne
// `ModuleEinbinden` hatte das Spiel null Felder, und ein Wechsel meldete
// `name` und `text` als ungedeutet, obwohl Magic beide führt. Kein Test hat es
// gesehen: die Saat schreibt ihre Felder unmittelbar hin.
func TestEinSpielMitFeldernAusEinemModulWirdGelesen(t *testing.T) {
	wurzel := t.TempDir()
	paket(t, wurzel, "modulspiel.mse-game",
		"mse version: 2.0.0\nshort name: Modulspiel\ninclude file: /felder.mse-include/felder\n")
	paket(t, wurzel, "felder.mse-include", "")

	if err := os.WriteFile(filepath.Join(wurzel, "felder.mse-include", "felder"),
		[]byte("card field:\n\ttype: text\n\tname: name\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket(t, wurzel, "modulstil.mse-style", "mse version: 2.0.0\ngame: modulspiel\n")

	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "anderes", Style: "modulstil"}}},
		&ablage{}, &stilablage{}, &feldliste{felder: []string{"name", "fremd"}}, wurzel,
	)

	folgen, err := waehle.Folgen(t.Context(), 1, "modulstil")
	if err != nil {
		t.Fatalf("Folgen: %v", err)
	}

	// **Nur `fremd`.** Ohne das Einbinden stünde hier auch `name`.
	if len(folgen.UngedeuteteFelder) != 1 || folgen.UngedeuteteFelder[0] != "fremd" {
		t.Errorf("ungedeutete Felder: %v", folgen.UngedeuteteFelder)
	}
}
