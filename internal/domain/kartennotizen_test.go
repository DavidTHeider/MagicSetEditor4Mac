// **Die Notizen einer Karte** — Werkstück 6 von Phase 3.7
// (#1168,
// Befund Z1-01).
//
// # Warum eine Ableitung und kein zweites Feld
//
// `card.notes` hatte hier **zwei** Darstellungen: `CardRow.Notes` als eigenes
// Feld — das **kein Leser dieses Projekts füllte** — und daneben den
// gewöhnlichen Feldwert `notes`, den der Einleser aus dem `card:`-Block macht.
// Welche galt, entschied eine Reihenfolge in `umgebungAus`, und die stand
// nirgends als Zusage.
//
// **MSE hat nur eine**, und zwar die Eigenschaft: `REFLECT(notes)` steht in
// `IMPLEMENT_REFLECTION(Card)` **vor** `REFLECT_NAMELESS(data)`
// (`src/data/card.cpp`, Zeilen 91 und 95, Fassung `f9d9356` vom 2021-07-06),
// und `GetMember::handle` nimmt den **ersten** Treffer
// (`src/util/io/get_member.hpp`, Zeile 94: `if (!gdm.result() && …)`). Ein
// Kartenfeld namens `notes` wäre dort verdeckt.
//
// **Hier bleibt der Feldwert die Darstellung**, und `Notizen()` ist der Blick
// darauf. Der Grund ist die Rundlaufzusage: `Values` ist die Liste, die
// zurückgeschrieben wird — wer `notes` herausnähme, schriebe eine Datei ohne
// Notizen zurück (`service/ausfuhr.go` baut die `Card` aus genau dieser
// Liste).
package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Notiz kommt aus dem Feldwert.**
func TestDieNotizEinerKarteKommtAusIhremWert(t *testing.T) {
	t.Parallel()

	karte := &domain.CardRow{Values: []domain.Value{
		{Field: "name", Text: "Akroma"},
		{Field: "notes", Text: "gemerkt"},
	}}

	if karte.Notizen() != "gemerkt" {
		t.Fatalf("Notizen = %q, erwartet \"gemerkt\"", karte.Notizen())
	}
}

// **Eine Karte ohne Notiz hat eine leere.**
//
// Nicht „kein `notes`": MSE gibt es an **jeder** Karte, leer, wenn die Datei
// keines trägt — und fünf der neunzehn Exportvorlagen des Packs lesen es.
func TestEineKarteOhneNotizHatEineLeere(t *testing.T) {
	t.Parallel()

	karte := &domain.CardRow{Values: []domain.Value{{Field: "name", Text: "Akroma"}}}

	if karte.Notizen() != "" {
		t.Fatalf("Notizen = %q, erwartet leer", karte.Notizen())
	}
}
