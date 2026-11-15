package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Rücknahme führt die Herkunft mit** —
// #1036.
//
// # Was fehlte
//
// Migration 0014 sagt zu: „Beide werden in derselben Transaktion geführt […]
// Auseinanderlaufen können sie nicht." Der **direkte** Schreibweg hielt das —
// die Rücknahme ging nie durch ihn: `kartenwegGehen` schrieb unmittelbar nach
// `card_values`.
//
// **Was das anrichtete**: nach der Rücknahme eines Scryfall-Bildes blieb die
// Nennung an einem wiederhergestellten eigenen Bild hängen — genau die falsche
// Nennung, die #1025 ausschließen sollte.
//
// Angestrichen vom Nachlauf über Phase V.
func TestEineRuecknahmeStelltDieHerkunftWiederHer(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	schreiber := storage.NeuerKartenschreiber(db).Als(domain.UrheberFenster)
	leser := storage.NeuerSetLeser(db)

	// **Erst von außen** — die Zeile bekommt ihre Herkunft.
	if _, _, fehler := schreiber.FeldSetzenVonAussen(probekontext(t), karteID,
		"kunst", "lotus.png", "https://scryfall.com/card/lea/232"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	// **Dann von hier** — sie fällt.
	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "kunst", "eigenes-foto.png"); fehler != nil {
		t.Fatalf("FeldSetzen: %v", fehler)
	}

	// **Und jetzt zurück.**
	if _, gab, fehler := storage.NeuerVerlaufsleser(db).Zuruecknehmen(
		probekontext(t), 1); fehler != nil || !gab {
		t.Fatalf("Zuruecknehmen: %v (gab: %v)", fehler, gab)
	}

	herkuenfte, lesefehler := leser.Herkuenfte(probekontext(t), karteID)
	if lesefehler != nil {
		t.Fatalf("Herkuenfte: %v", lesefehler)
	}

	if herkuenfte["kunst"] != "https://scryfall.com/card/lea/232" {
		t.Errorf("Herkunft nach der Rücknahme = %q — der Wert von Scryfall steht "+
			"wieder da, seine Nennung muss es auch", herkuenfte["kunst"])
	}
}

// **Und die Gegenrichtung: ein zurückgenommener Wert von außen verliert sie.**
//
// Sonst behauptete die Anwendung über das wiederhergestellte eigene Bild, es
// sei ein Ausschnitt von Scryfall.
func TestEineRuecknahmeNimmtEineFremdeHerkunftWeg(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	schreiber := storage.NeuerKartenschreiber(db).Als(domain.UrheberFenster)

	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "kunst", "eigenes-foto.png"); fehler != nil {
		t.Fatalf("FeldSetzen: %v", fehler)
	}

	if _, _, fehler := schreiber.FeldSetzenVonAussen(probekontext(t), karteID,
		"kunst", "lotus.png", "https://scryfall.com/card/lea/232"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	if _, gab, fehler := storage.NeuerVerlaufsleser(db).Zuruecknehmen(
		probekontext(t), 1); fehler != nil || !gab {
		t.Fatalf("Zuruecknehmen: %v (gab: %v)", fehler, gab)
	}

	herkuenfte, lesefehler := storage.NeuerSetLeser(db).Herkuenfte(probekontext(t), karteID)
	if lesefehler != nil {
		t.Fatalf("Herkuenfte: %v", lesefehler)
	}

	if _, da := herkuenfte["kunst"]; da {
		t.Errorf("»kunst« trägt weiter %q — das eigene Bild steht wieder da, "+
			"und eine falsche Nennung ist schlimmer als keine", herkuenfte["kunst"])
	}
}

// **Eine zurückgeholte Karte trägt ihre Herkünfte** —
// Rest von #1036.
//
// # Warum das ein eigener Weg ist
//
// `wert_herkunft` hängt über `ON DELETE CASCADE` an `cards`. Eine gelöschte
// Karte nimmt ihre Herkunftszeilen mit — und das Wiederherstellen baut nur die
// **Werte** wieder auf. Ein zurückgeholtes Scryfall-Bild stünde danach ohne
// seine Nennung da: genau die fehlende Nennung, die #1025 verhindern sollte.
func TestEineZurueckgeholteKarteTraegtIhreHerkuenfte(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	schreiber := storage.NeuerKartenschreiber(db).Als(domain.UrheberFenster)

	if _, _, fehler := schreiber.FeldSetzenVonAussen(probekontext(t), karteID,
		"kunst", "lotus.png", "https://scryfall.com/card/lea/232"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	if fehler := schreiber.KarteLoeschen(probekontext(t), karteID); fehler != nil {
		t.Fatalf("KarteLoeschen: %v", fehler)
	}

	if _, gab, fehler := storage.NeuerVerlaufsleser(db).Zuruecknehmen(
		probekontext(t), 1); fehler != nil || !gab {
		t.Fatalf("Zuruecknehmen: %v (gab: %v)", fehler, gab)
	}

	herkuenfte, lesefehler := storage.NeuerSetLeser(db).Herkuenfte(probekontext(t), karteID)
	if lesefehler != nil {
		t.Fatalf("Herkuenfte: %v", lesefehler)
	}

	if herkuenfte["kunst"] != "https://scryfall.com/card/lea/232" {
		t.Errorf("Herkunft nach dem Zurückholen = %q — das Bild ist wieder da, "+
			"seine Nennung muss es auch sein", herkuenfte["kunst"])
	}
}
