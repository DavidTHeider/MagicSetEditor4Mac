// **Die Notizen einer Karte** — Werkstück 3 von Phase 3.5
// (#1166).
//
// # Woher die Notizen kommen
//
// Aus der Datei. Ein `card:`-Block trägt seine Feldwerte **und** ein `notes`
// (`String notes;`, `src/data/card.hpp`, Zeile 39, Fassung `641acff` vom
// 2020-05-19, `REFLECT(notes)` in `card.cpp`, Zeile 91). Der Einleser dieses
// Projekts macht aus **jedem** Kind eines `card:` einen [domain.Value] — auch
// aus `notes`, so wie aus `has styling` und `styling data`.
//
// # Zwei Einträge unter einem Namen, und der erste gewinnt
//
// Daneben hängt `umgebungAus` hinter die Feldwerte einen Eintrag `notes` aus
// `CardRow.Notes` — einem Feld, das **kein Leser dieses Projekts füllt**. Es
// steht dort, damit `notes` an **jeder** Karte steht, auch an einer ohne
// (#609); fünf der neunzehn Exportvorlagen des Packs lesen es.
//
// **Damit stehen bei einer Karte mit Notizen zwei Einträge `notes` da**, und
// welcher gilt, entscheidet [msescript.Wert.Eintrag]: der **erste**. Das ist
// der Wert aus der Datei, und das ist richtig.
//
// **Diese beiden Läufe halten genau das fest.** Es ist eine Zusage über eine
// Reihenfolge, die niemand sieht — wer die Zeilen in `umgebungAus` tauscht,
// bekommt an jeder Karte des kanonischen `Vanguard` ein leeres `notes`, und
// `magic.mse-game` liest daraus die Kartennummer.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein `notes` aus der Datei kommt beim Skript an.**
//
// # Warum das nicht selbstverständlich ist
//
// Es stehen **zwei** Einträge `notes` in der Abbildung: der aus der Datei und
// der leere Rückfall. Diese Zusage ist die Reihenfolge zwischen ihnen.
//
// **Und sie zählt.** `magic.mse-game` liest `card.notes` an drei Stellen,
// darunter die Kartennummer:
// `exporter_num_grabber(exporter_num_filter(card.notes))`. Über die 107 Karten
// des kanonischen `Vanguard` trägt **jede** ein `notes` — am 2026-08-27 im
// laufenden Fenster nachgezählt.
func TestEinNotesAusDerDateiErreichtDasSkript(t *testing.T) {
	t.Parallel()

	vorlage := probevorlage(t)
	vorlage.Skript = `cards[0].notes`

	karten := []*domain.CardRow{{
		ID: 1,
		Values: []domain.Value{
			{Field: "name", Text: "Feuerball"},
			{Field: "notes", Text: "Sets: [pmoa];"},
		},
	}}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "Sets: [pmoa];" {
		t.Errorf("Ausgabe %q — das `notes` der Datei kam nicht an", ausgabe)
	}
}

// **Und ohne eines bleibt der leere Wert.**
//
// Die Gegenseite: `notes` steht an **jeder** Karte, auch wenn die Datei keines
// nennt — sonst brächen die fünf Exportvorlagen des Packs, die es lesen, mit
// „die Abbildung führt kein Glied `notes`" ab
// (#609).
//
// **Ohne diese Hälfte wäre auch ein Weg grün, der den Eintrag ganz weglässt.**
func TestOhneNotesGibtEsTrotzdemEines(t *testing.T) {
	t.Parallel()

	vorlage := probevorlage(t)
	vorlage.Skript = `to_string(cards[0].notes) + "|"`

	karten := []*domain.CardRow{{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "|" {
		t.Errorf("Ausgabe %q, erwartet `|`", ausgabe)
	}
}
