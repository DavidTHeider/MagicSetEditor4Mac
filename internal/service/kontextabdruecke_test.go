package service_test

// **Die zwei Abdrücke eines Kontextpakets** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Warum eine eigene Datei
//
// **Weil `kontextdienst_test.go` über die Zeilengrenze aus §2 ging**, und der
// Schnitt folgt der Frage: dort steht, **was** in ein Paket kommt — Belege,
// Budget, Abweisungen; hier, **wogegen** es entstanden ist.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Manifest nennt beide Abdrücke — und keiner ist die Modellrevision.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil hier bis zum 2026-09-12 ein Feld leer stand und eines das Falsche
// trug**: `Indexfingerabdruck` wurde nie gefüllt, und `Raumabdruck` bekam die
// **Modellrevision** — eine Angabe, die vier andere verschweigt (Zerteiler,
// Pooling, Normalisierung, Abstandsmaß). Zwei Pakete aus zwei verschiedenen
// Räumen sahen damit gleich aus.
//
// **Sie beantworten zwei Fragen.** Der Raum sagt, ob zwei Vektoren
// vergleichbar sind; der Index sagt, ob zwei Antworten aus derselben Liste
// stammen.
func TestDasManifestNenntBeideAbdruecke(t *testing.T) {
	t.Parallel()

	stand := einRegelstand()

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand:   stand,
		treffer: []domain.Retrievaltreffer{einRegeltreffer("702.9b", "Trampelschaden ändert die Zuweisung.", 1)},
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	manifest := paket.Manifest

	if manifest.Indexfingerabdruck == "" || manifest.Raumabdruck == "" {
		t.Fatalf("ein Abdruck fehlt: %q / %q",
			manifest.Indexfingerabdruck, manifest.Raumabdruck)
	}

	// **Der Raumabdruck ist der des Raums** und nicht irgendeine Zeichenkette
	// daraus.
	if manifest.Raumabdruck != stand.Manifest.Raum().Abdruck() {
		t.Errorf("der Raumabdruck ist %q", manifest.Raumabdruck)
	}

	// **Und er ist nicht die Modellrevision** — das war der Fehler.
	if manifest.Raumabdruck == stand.Manifest.Modellrevision {
		t.Errorf("der Raumabdruck ist die Modellrevision")
	}

	// **Die zwei sind nicht dasselbe**: der Index zählt Verfahren und
	// Korpusstand mit, der Raum nicht.
	if manifest.Indexfingerabdruck == manifest.Raumabdruck {
		t.Errorf("beide Abdrücke sind gleich: %q", manifest.Raumabdruck)
	}
}

// **Und ein anderer Zerteiler gibt einen anderen Raumabdruck.**
//
// **Die Gegenprobe** (§10): ohne sie sagte der Lauf darüber nur, dass zwei
// Zeichenketten nicht leer sind.
func TestEinAndererZerteilerGibtEinenAnderenRaumabdruck(t *testing.T) {
	t.Parallel()

	abdruecke := make([]string, 0, 2)

	for _, zerteiler := range []string{"bert", "sentencepiece"} {
		stand := einRegelstand()
		stand.Manifest.Vertrag.Zerteiler = zerteiler

		paket, fehler := einKontextdienst(&recherchequellendoppel{
			stand:   stand,
			treffer: []domain.Retrievaltreffer{einRegeltreffer("702.9b", "Trampelschaden ändert die Zuweisung.", 1)},
		}).Zusammenstellen(t.Context(), einKontextwunsch())
		if fehler != nil {
			t.Fatalf("Zusammenstellen mit %q: %v", zerteiler, fehler)
		}

		abdruecke = append(abdruecke, paket.Manifest.Raumabdruck)
	}

	if abdruecke[0] == abdruecke[1] {
		t.Fatalf("zwei Zerteiler geben denselben Raumabdruck: %q",
			abdruecke[0])
	}
}
