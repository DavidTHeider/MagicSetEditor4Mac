package msepaket_test

import "testing"

func TestSprachmodellLiestFeldbeschriftungen(t *testing.T) {
	sprache := oeffneTestpaket(t, "de.mse-locale").Sprachmodell()

	if sprache.Name != "Deutsch" {
		t.Errorf("Name %q", sprache.Name)
	}
	if beschriftung := sprache.Feldbeschriftung("probe", "farbe"); beschriftung != "Farbe" {
		t.Errorf("Beschriftung %q, erwartet war %q", beschriftung, "Farbe")
	}
	if beschriftung := sprache.Paketbeschriftung("probe.mse-game"); beschriftung != "Probespiel" {
		t.Errorf("Paketbeschriftung %q", beschriftung)
	}
}

// Ohne Übersetzung kommt der Originalname zurück, nicht die leere Zeichenkette.
//
// Dasselbe Verhältnis wie beim Rückfallsatz einer Fehlermeldung: eine fehlende
// Übersetzung kostet eine unübersetzte Beschriftung, nicht eine leere. Ein
// leeres Feldetikett wäre eine Oberfläche, die nichts mehr sagt.
func TestOhneUebersetzungKommtDerOriginalnameZurueck(t *testing.T) {
	sprache := oeffneTestpaket(t, "de.mse-locale").Sprachmodell()

	faelle := map[string]string{
		"ein unbekanntes Feld":  sprache.Feldbeschriftung("probe", "gibtsnicht"),
		"ein unbekanntes Spiel": sprache.Feldbeschriftung("anderes", "farbe"),
		"ein unbekanntes Paket": sprache.Paketbeschriftung("fremd.mse-game"),
	}
	erwartet := map[string]string{
		"ein unbekanntes Feld":  "gibtsnicht",
		"ein unbekanntes Spiel": "farbe",
		"ein unbekanntes Paket": "fremd.mse-game",
	}

	for name, gefunden := range faelle {
		t.Run(name, func(t *testing.T) {
			if gefunden != erwartet[name] {
				t.Errorf("Beschriftung %q, erwartet war %q", gefunden, erwartet[name])
			}
		})
	}
}

// Kommentare im Sprachpaket werden keine Beschriftungen.
//
// Sie tragen keinen Schlüssel; sie als Eintrag zu lesen ergäbe ein Feld namens
// `# ein Kommentar…`, das in keiner Oberfläche vorkommt und trotzdem gepflegt
// werden müsste.
func TestKommentareImSprachpaketWerdenKeineBeschriftungen(t *testing.T) {
	sprache := oeffneTestpaket(t, "de.mse-locale").Sprachmodell()

	if beschriftung := sprache.Feldbeschriftung("probe", "name"); beschriftung != "Name" {
		t.Errorf("Beschriftung %q — ein Kommentar hat die Zuordnung verschoben", beschriftung)
	}
	if beschriftung := sprache.Paketbeschriftung("probe.mse-game"); beschriftung != "Probespiel" {
		t.Errorf("Paketbeschriftung %q", beschriftung)
	}
}
