package storage_test

// **Welche MTGJSON-Kennung zu welchem Druck gehört** — Werkstück A3 von Phase
// 5.3 (#1442).
//
// # Wofür es das braucht
//
// Die Preisdatei von MTGJSON ist nach **seiner** Kennung geordnet und führt
// jede Karte, die es gibt — über hunderttausend. Der Bestand führt die, die
// jemand besitzt. Was dazwischen liegt, ist diese Karte: **wer im Bestand ist,
// wird gelesen; der Rest wird übersprungen, ohne ihn zu lesen.**

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitKennungen legt einen Druck mit den genannten Kennungen an.
func mitKennungen(t *testing.T, drucke *storage.Drucke, name string,
	kennungen ...domain.Druckkennung,
) {
	t.Helper()

	druck := &domain.Druck{
		Name: name, Setcode: "m11", Nummer: "1", Sprache: "en",
		Kennungen: kennungen,
	}

	if fehler := drucke.Sichern(probekontext(t), druck); fehler != nil {
		t.Fatalf("%s sichern: %v", name, fehler)
	}
}

// mtgjsonkennung baut eine Kennung von MTGJSON.
func mtgjsonkennung(wert string) domain.Druckkennung {
	return domain.Druckkennung{
		System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: wert,
	}
}

// **Nur Drucke mit beiden Kennungen stehen in der Karte.**
func TestDerKrosswalkFuehrtNurDruckeMitBeidenKennungen(t *testing.T) {
	db := frischeDatenbank(t)
	drucke := storage.NeueDrucke(db)

	mitKennungen(t, drucke, "Mit beidem",
		scryfallkennung("11111111-1111-4111-8111-111111111111"),
		mtgjsonkennung("22222222-2222-4222-8222-222222222222"))

	mitKennungen(t, drucke, "Nur Scryfall",
		scryfallkennung("33333333-3333-4333-8333-333333333333"))

	mitKennungen(t, drucke, "Nur MTGJSON",
		mtgjsonkennung("44444444-4444-4444-8444-444444444444"))

	karte, fehler := storage.Druckkrosswalk(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("Druckkrosswalk: %v", fehler)
	}

	if len(karte) != 1 {
		t.Fatalf("die Karte führt %d Einträge: %v", len(karte), karte)
	}

	if karte["22222222-2222-4222-8222-222222222222"] !=
		"11111111-1111-4111-8111-111111111111" {
		t.Errorf("die Karte ist %v", karte)
	}
}

// **Ein Konzept ist kein Druck.**
//
// Eine Oracle-ID nennt das Konzept, das dreißig Drucke teilen (0086). Wer sie
// als Druckkennung führte, ordnete denselben Preis dreißig Karten zu.
func TestDerKrosswalkNimmtNurDieSemantikDruck(t *testing.T) {
	db := frischeDatenbank(t)
	drucke := storage.NeueDrucke(db)

	mitKennungen(t, drucke, "Mit Konzept",
		scryfallkennung("11111111-1111-4111-8111-111111111111"),
		domain.Druckkennung{
			System: domain.SystemMTGJSON, Semantik: domain.SemantikKonzept,
			Wert: "55555555-5555-4555-8555-555555555555",
		})

	karte, fehler := storage.Druckkrosswalk(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("Druckkrosswalk: %v", fehler)
	}

	if len(karte) != 0 {
		t.Errorf("die Karte führt ein Konzept: %v", karte)
	}
}

// **Ein leerer Bestand gibt eine leere Karte — und keinen Fehler.**
func TestDerKrosswalkEinesLeerenBestandsIstLeer(t *testing.T) {
	karte, fehler := storage.Druckkrosswalk(probekontext(t), frischeDatenbank(t))
	if fehler != nil {
		t.Fatalf("Druckkrosswalk: %v", fehler)
	}

	if len(karte) != 0 {
		t.Errorf("die Karte ist %v", karte)
	}
}

// **Und ein geschlossener Bestand meldet sich** (§8).
func TestDerKrosswalkMeldetEinenGeschlossenenBestand(t *testing.T) {
	if _, fehler := storage.Druckkrosswalk(
		probekontext(t), kaputterBestand(t)); fehler == nil {
		t.Error("es kam durch")
	}
}

// **Eine Zeile, die sich nicht lesen lässt, meldet sich.**
//
// **Über den Stolpertreiber**, weil eine echte Datenbank sie nicht hergibt:
// beide Spalten sind `TEXT NOT NULL`, und ein `Scan` auf zwei Zeichenketten
// scheitert dort nicht. Der Weg gehört trotzdem geprüft — im Betrieb steht am
// Ende einer Leitung, was niemand vorhersagt (§3).
func TestEineKrummeKrosswalkzeileMeldetSich(t *testing.T) {
	db, fehler := sql.Open("stolper-krosswalk-scan", ":memory:")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, err := storage.Druckkrosswalk(probekontext(t), db); err == nil {
		t.Error("es kam durch")
	}
}
