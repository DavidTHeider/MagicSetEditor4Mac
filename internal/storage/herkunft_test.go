package storage_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was von außen kam, ist danach erkennbar** —
// Abnahmekriterium aus #888:
// „Herkunft je Feld: was von dort kam, ist erkennbar."
//
// # Warum der Urheber dafür nicht reicht
//
// `urheber` sagt, **welche Fläche** geschrieben hat — Fenster, Server,
// Kommandozeile, Dienst. Er sagt nichts darüber, **woher der Wert stammt**:
// ein von Hand getippter Regeltext und einer von Scryfall kommen beide über
// dasselbe Fenster. Ohne eine eigene Angabe ist die Übernahme nach dem
// Schreiben nicht mehr von einer Eingabe zu unterscheiden — und §16, Regel 6
// verlangt genau das: „trägt danach seine Herkunft".
func TestEineUebernahmeTraegtIhreHerkunft(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	setID := int64(1)

	schreiber := storage.NeuerKartenschreiber(db).Als(domain.UrheberFenster)

	// **Ohne Herkunft**: eine Eingabe von Hand.
	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "name", "Selbst getippt"); fehler != nil {
		t.Fatalf("FeldSetzen: %v", fehler)
	}

	// **Mit Herkunft**: eine Übernahme von außen.
	vonScryfall := schreiber.Von("https://scryfall.com/card/lea/161")
	if _, _, fehler := vonScryfall.FeldSetzenMitVorwert(
		probekontext(t), karteID, "rule text", "Fliegend."); fehler != nil {
		t.Fatalf("FeldSetzen mit Herkunft: %v", fehler)
	}

	eintraege, lesefehler := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Verlauf: %v", lesefehler)
	}

	nach := map[string]domain.Verlaufseintrag{}
	for _, eintrag := range eintraege {
		nach[eintrag.Feld] = eintrag
	}

	if nach["name"].Herkunft != "" {
		t.Errorf("die getippte Eingabe trägt eine Herkunft: %q", nach["name"].Herkunft)
	}

	if !strings.Contains(nach["rule text"].Herkunft, "scryfall.com") {
		t.Errorf("Herkunft = %q, erwartet die Kartenseite bei Scryfall",
			nach["rule text"].Herkunft)
	}

	// **Und der Urheber bleibt, was er war** — die beiden Angaben beantworten
	// verschiedene Fragen und dürfen sich nicht ersetzen.
	if nach["rule text"].Urheber != domain.UrheberFenster {
		t.Errorf("Urheber = %q, erwartet das Fenster", nach["rule text"].Urheber)
	}
}

// **Auch eine Reihe von Feldern trägt ihre Herkunft.**
//
// # Warum das ein eigener Fall ist
//
// Ein Feld mit `save value: false` wird in **seine Teile** zerlegt und in
// einer Transaktion geschrieben ([service.Feldweg], #714). Ein übernommener
// Wert geht denselben Weg — und dann muss jede der entstehenden Zeilen ihre
// Herkunft tragen, nicht nur die erste.
func TestEineReiheTraegtIhreHerkunft(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, fehler := schreiber.FelderSetzenVonAussen(probekontext(t), karteID,
		[]domain.Value{
			{Field: "rule text", Text: "Fliegend."},
			{Field: "flavor text", Text: "Und leise."},
		}, "https://scryfall.com/card/lea/161", ""); fehler != nil {
		t.Fatalf("FelderSetzen: %v", fehler)
	}

	eintraege, lesefehler := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), 1)
	if lesefehler != nil {
		t.Fatalf("Verlauf: %v", lesefehler)
	}

	gezaehlt := 0

	for _, eintrag := range eintraege {
		if eintrag.Feld == "rule text" || eintrag.Feld == "flavor text" {
			gezaehlt++

			if eintrag.Herkunft != "https://scryfall.com/card/lea/161" {
				t.Errorf("Herkunft von %q = %q", eintrag.Feld, eintrag.Herkunft)
			}
		}
	}

	if gezaehlt != 2 {
		t.Errorf("%d Zeilen geprüft, erwartet 2", gezaehlt)
	}
}

// **Und die Einzelfassung ebenso** — sie ist der Weg, den der Feldabgleich
// nimmt.
func TestFeldSetzenVonAussenTraegtDieHerkunft(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	if _, _, fehler := storage.NeuerKartenschreiber(db).FeldSetzenVonAussen(
		probekontext(t), karteID, "rule text", "Fliegend.",
		"https://scryfall.com/card/lea/161"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	eintraege, lesefehler := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), 1)
	if lesefehler != nil {
		t.Fatalf("Verlauf: %v", lesefehler)
	}

	if len(eintraege) == 0 ||
		!strings.Contains(eintraege[0].Herkunft, "scryfall.com") {
		t.Errorf("die Herkunft fehlt: %+v", eintraege)
	}
}
