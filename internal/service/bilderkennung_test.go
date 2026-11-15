package service_test

// **Aus einem Kartenbild werden Kandidaten** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass die Kandidaten aus der Suche kommen und nicht aus dem Modell.** Das
// ist die Zusage des Issues: „explizite Zuordnung über A2". Eine
// Bilderkennung kann damit keinen Druck nennen, den die Suche nicht auch
// gefunden hätte.
//
// **Dass die Unsicherheit gezählt ist** — und dass „mindestens" nicht als
// „genau" durchgeht.
//
// **Und dass nichts geschrieben wird.**

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildsuche ist die Suche, die der Erkenner benutzt — dieselbe wie A2.
type bildsuche struct {
	treffer scryfall.Treffer
	gefragt string
	fehler  error

	// rufe zählt mit — **neben** der Frage und nicht statt ihrer.
	//
	// **Sonst sieht „gar nicht gerufen" aus wie „mit leerer Frage gerufen".**
	// Eine Gegenprobe, die den leeren Suchbegriff doch hinausließ, blieb
	// genau daran grün (§10).
	rufe int
}

func (b *bildsuche) Druckesuche(_ context.Context, frage string) (
	scryfall.Treffer, time.Duration, []domain.Befund, error,
) {
	b.rufe++
	b.gefragt = frage

	return b.treffer, time.Hour, nil, b.fehler
}

// einDruck gibt einen Druck, wie Scryfall ihn führt.
//
// **Nicht `eineKarte`**: den Namen trägt in diesem Paket schon die Vorrichtung
// der Ausfuhr, und sie gibt eine `domain.CardRow`. Zwei Bedeutungen unter
// einem Namen sind eine Falle für den nächsten Leser.
func einDruck(name, set, nummer string) *scryfall.Karte {
	return &scryfall.Karte{
		Kennung: set + "-" + nummer, Name: name,
		Set: set, Setname: set, Nummer: nummer,
	}
}

// **Der Suchbegriff kommt aus den Beobachtungen, die Kandidaten aus der
// Suche.**
func TestDieKandidatenKommenAusDerSuche(t *testing.T) {
	t.Parallel()

	suche := &bildsuche{treffer: scryfall.Treffer{
		Karten: []*scryfall.Karte{einDruck("Lightning Bolt", "lea", "161")},
	}}

	deutung, fehler := service.NeuerBilderkenner(suche).Deuten(
		context.Background(), domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Wert: "Lightning Bolt", Lesbar: true},
				{Art: domain.EvidenzSetcode, Wert: "LEA", Lesbar: true},
			},
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if suche.gefragt != "Lightning Bolt LEA" {
		t.Errorf("gefragt wurde %q", suche.gefragt)
	}

	if len(deutung.Kandidaten) != 1 || deutung.Drucke != 1 {
		t.Fatalf("die Deutung ist %+v", deutung)
	}

	// **Beide Beobachtungen stützen ihn** — und der Setcode, obwohl er in
	// Kapitälchen gelesen wurde.
	if len(deutung.Kandidaten[0].Gestuetzt) != 2 {
		t.Errorf("gestützt von %v", deutung.Kandidaten[0].Gestuetzt)
	}

	if !deutung.Eindeutig() {
		t.Error("ein einziger Druck galt nicht als eindeutig")
	}
}

// **„Mindestens" geht nicht als „genau" durch.**
//
// **Das ist der Unterschied, an dem die Auswahl hängt**: Scryfall gibt eine
// Seite, und `Weitere` sagt, dass es dahinter weitergeht. Ein Kandidat, hinter
// dem noch etwas liegt, ist nicht eindeutig.
func TestEinTrefferMitWeiterenIstNichtEindeutig(t *testing.T) {
	t.Parallel()

	suche := &bildsuche{treffer: scryfall.Treffer{
		Karten:  []*scryfall.Karte{einDruck("Counterspell", "ice", "64")},
		Weitere: true,
	}}

	deutung, fehler := service.NeuerBilderkenner(suche).Deuten(
		context.Background(), domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Wert: "Counterspell", Lesbar: true},
			},
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if !deutung.Weitere {
		t.Error("die Deutung verschweigt, dass es mehr gibt")
	}

	if deutung.Eindeutig() {
		t.Error("ein Treffer mit Weiteren galt als eindeutig")
	}
}

// **Ein Bild, auf dem nichts zu lesen war, fragt die Suche gar nicht.**
//
// **Weil eine Suche über die leere Zeichenkette alles fände** — und aus
// „nichts gelesen" würde „alle Karten passen".
func TestEinUnlesbaresBildFragtDieSucheNicht(t *testing.T) {
	t.Parallel()

	suche := &bildsuche{}

	deutung, fehler := service.NeuerBilderkenner(suche).Deuten(
		context.Background(), domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Lesbar: false},
			},
			NichtGelesen: []domain.Evidenzart{domain.EvidenzSetcode},
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if suche.rufe != 0 {
		t.Errorf("die Suche wurde %dmal gerufen — zuletzt mit %q",
			suche.rufe, suche.gefragt)
	}

	if len(deutung.Kandidaten) != 0 || deutung.Drucke != 0 {
		t.Errorf("die Deutung ist %+v", deutung)
	}
}

// **Ein Kunstwerk stützt keinen Kandidaten — und steht trotzdem da.**
//
// **Beides zusammen ist die Zusage**: „OCR/Layout/Face/Set/Kunstwerk als
// getrennte Evidenz behandeln." Getrennt heißt auch: nicht stillschweigend
// weggelassen.
func TestEinKunstwerkStuetztKeinenKandidaten(t *testing.T) {
	t.Parallel()

	suche := &bildsuche{treffer: scryfall.Treffer{
		Karten: []*scryfall.Karte{einDruck("Shivan Dragon", "lea", "174")},
	}}

	deutung, fehler := service.NeuerBilderkenner(suche).Deuten(
		context.Background(), domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Wert: "Shivan Dragon", Lesbar: true},
				{Art: domain.EvidenzKunstwerk, Wert: "ein roter Drache", Lesbar: true},
				{Art: domain.EvidenzLayout, Wert: "normal", Lesbar: true},
			},
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(deutung.Kandidaten[0].Gestuetzt) != 1 ||
		deutung.Kandidaten[0].Gestuetzt[0] != domain.EvidenzName {
		t.Errorf("gestützt von %v", deutung.Kandidaten[0].Gestuetzt)
	}

	// **Und die drei Beobachtungen stehen alle in der Antwort.**
	if len(deutung.Erkennung.Beobachtungen) != 3 {
		t.Errorf("die Beobachtungen sind %+v", deutung.Erkennung.Beobachtungen)
	}
}

// **Eine klemmende Suche reicht ihren Fehler durch.**
func TestEineKlemmendeSucheReichtDurch(t *testing.T) {
	t.Parallel()

	suche := &bildsuche{fehler: errors.New("die Suche klemmt")}

	if _, fehler := service.NeuerBilderkenner(suche).Deuten(
		context.Background(), domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Wert: "Counterspell", Lesbar: true},
			},
		}); fehler == nil {
		t.Fatal("der Fehler der Suche kam nicht durch")
	}
}

// **Mehr Kandidaten als die Grenze werden abgeschnitten — die Zahl nicht.**
func TestDieKandidatenlisteEndetAnDerGrenze(t *testing.T) {
	t.Parallel()

	karten := make([]*scryfall.Karte, 0, service.Erkennungsgrenze+5)
	for i := range service.Erkennungsgrenze + 5 {
		karten = append(karten, einDruck("Counterspell", "set", string(rune('a'+i))))
	}

	deutung, fehler := service.NeuerBilderkenner(
		&bildsuche{treffer: scryfall.Treffer{Karten: karten}},
	).Deuten(context.Background(), domain.Bilderkennung{
		Beobachtungen: []domain.Beobachtung{
			{Art: domain.EvidenzName, Wert: "Counterspell", Lesbar: true},
		},
	})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(deutung.Kandidaten) != service.Erkennungsgrenze {
		t.Errorf("es sind %d Kandidaten", len(deutung.Kandidaten))
	}

	// **Die Zahl bleibt die volle** — sie ist die Unsicherheit, und eine
	// abgeschnittene Zahl verschwiege sie.
	if deutung.Drucke != service.Erkennungsgrenze+5 {
		t.Errorf("die Drucke sind %d", deutung.Drucke)
	}
}

// **Jede Evidenzart wird gegen ihr eigenes Feld gehalten.**
//
// # Warum tabellengetrieben und nicht in einem Lauf
//
// **Weil jede Art ein eigener Vergleich ist** und einer davon ohne
// Rücksicht auf Groß- und Kleinschreibung ausgerechnet nicht: eine
// Sammlernummer ist eine Zeichenkette, in der „07" und „7" zwei verschiedene
// Karten sind.
//
// **Der Lauf über die unbekannte Art steht mit darin.** Das Schema nennt die
// sechs, aber ein Modell muss sich nicht daran halten
// ([domain.BefundErfundeneEvidenz] gibt es genau deswegen) — und was dann
// geschieht, ist eine Aussage über diesen Vergleich und keine über das
// Schema.
func TestJedeEvidenzartWirdGegenIhrFeldGehalten(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		art       domain.Evidenzart
		wert      string
		gestuetzt bool
	}{
		{"der Setcode passt", domain.EvidenzSetcode, "LEA", true},
		{"der Setcode passt nicht", domain.EvidenzSetcode, "leb", false},
		{"die Nummer passt", domain.EvidenzNummer, "174", true},
		{"eine Nummer mit Null ist eine andere", domain.EvidenzNummer, "0174", false},
		{"die Typzeile passt", domain.EvidenzTypzeile, "creature — dragon", true},
		{"die Typzeile passt nicht", domain.EvidenzTypzeile, "Instant", false},
		{"eine erfundene Art stützt nichts", domain.Evidenzart("wasserzeichen"), "lea", false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			druck := einDruck("Shivan Dragon", "lea", "174")
			druck.Typzeile = "Creature — Dragon"

			suche := &bildsuche{treffer: scryfall.Treffer{
				Karten: []*scryfall.Karte{druck},
			}}

			// **Der Name steht immer dabei**, damit der Suchbegriff nicht
			// leer ist: eine Suche über nichts wird gar nicht erst gefragt,
			// und dann gäbe es keinen Kandidaten zu stützen.
			deutung, fehler := service.NeuerBilderkenner(suche).Deuten(
				context.Background(), domain.Bilderkennung{
					Beobachtungen: []domain.Beobachtung{
						{Art: domain.EvidenzName, Wert: "Shivan Dragon", Lesbar: true},
						{Art: fall.art, Wert: fall.wert, Lesbar: true},
					},
				})
			if fehler != nil {
				t.Fatalf("deuten: %v", fehler)
			}

			gestuetzt := deutung.Kandidaten[0].Gestuetzt

			erwartet := 1
			if fall.gestuetzt {
				erwartet = 2
			}

			if len(gestuetzt) != erwartet {
				t.Errorf("gestützt von %v, erwartet waren %d Arten",
					gestuetzt, erwartet)
			}
		})
	}
}

// **Eine unlesbare Stelle stützt keinen Kandidaten — auch mit Wert nicht.**
//
// **Das ist der Zweig, der sie überspringt**, und er ist nicht derselbe wie
// der Vergleich darunter: dort würde „lea" gegen „lea" passen. Wer sagt, er
// sehe das Set nicht, und trotzdem eines nennt, hat geraten — und ein
// geratener Wert darf keinen Druck stützen.
func TestEineUnlesbareStelleStuetztNichts(t *testing.T) {
	t.Parallel()

	suche := &bildsuche{treffer: scryfall.Treffer{
		Karten: []*scryfall.Karte{einDruck("Shivan Dragon", "lea", "174")},
	}}

	deutung, fehler := service.NeuerBilderkenner(suche).Deuten(
		context.Background(), domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Wert: "Shivan Dragon", Lesbar: true},

				// **Ein Wert an einer unlesbaren Stelle** — genau der Fall
				// aus [domain.BefundWertOhneLesbarkeit], und er passte.
				{Art: domain.EvidenzSetcode, Wert: "lea", Lesbar: false},

				// **Und eine lesbare Stelle ohne Wert.**
				{Art: domain.EvidenzNummer, Wert: "", Lesbar: true},
			},
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if gestuetzt := deutung.Kandidaten[0].Gestuetzt; len(gestuetzt) != 1 ||
		gestuetzt[0] != domain.EvidenzName {
		t.Errorf("gestützt von %v — erwartet war allein der Name", gestuetzt)
	}

	// **Der Befund steht trotzdem da** — er ist die Auskunft über das Raten.
	if len(deutung.Befunde) != 1 ||
		deutung.Befunde[0].Art != domain.BefundWertOhneLesbarkeit {
		t.Errorf("die Befunde sind %+v", deutung.Befunde)
	}
}
