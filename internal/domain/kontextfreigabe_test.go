package domain_test

// **Privates geht nur hinaus, wenn jemand es ausgewählt hat** — Werkstück
// A13.4 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass eine Recherche ohne Freigabe in keinem privaten Korpus sucht.** Das
// Issue sagt es wörtlich: „Referenzkarten/Decks/private Setnotizen **nur
// ausdrücklich ausgewählt**". Bis A13.4 suchte `aehnlich` von sich aus in den
// eigenen Karten — und wer eine Karte mit einem Modell verglich, schickte
// dabei seine eigenen mit, ohne es zu sagen.
//
// **Und dass die Weglassung sichtbar ist.** Ein Korpus, in dem nicht gesucht
// wurde, ist eine Einschränkung der Antwort und kein Formfehler (§8).

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zwei der fünf Quellen sind privat.**
//
// **Und eine unbekannte gilt als privat.** Dieselbe Regel wie bei
// [service.kategorieFuer]: eine neue Quelle stillschweigend als öffentlich zu
// führen wäre die Gestalt, in der sie unbemerkt hinausginge.
func TestZweiQuellenSindPrivat(t *testing.T) {
	t.Parallel()

	privat := map[domain.Retrievalquelle]bool{
		domain.QuelleEigeneKarte: true, domain.QuelleEigenerText: true,
		domain.QuelleOracle: false, domain.QuelleRuling: false,
		domain.QuelleRegelwerk: false,
		"gibtesnochnicht":      true,
	}

	for quelle, erwartet := range privat {
		if quelle.IstPrivat() != erwartet {
			t.Errorf("%q ist privat: %v, erwartet war %v",
				quelle, quelle.IstPrivat(), erwartet)
		}
	}
}

// **Ohne Freigabe sucht eine Ähnlichkeitsrecherche nur öffentlich.**
func TestOhneFreigabeBleibtDerEigeneBestandAussen(t *testing.T) {
	t.Parallel()

	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 4, Grenze: 5,
	}

	quellen := anfrage.Suchquellen()

	if slices.Contains(quellen, domain.QuelleEigeneKarte) {
		t.Errorf("ohne Freigabe wurde in %v gesucht", quellen)
	}

	if !slices.Contains(quellen, domain.QuelleOracle) {
		t.Errorf("die öffentliche Quelle fehlt: %v", quellen)
	}
}

// **Mit Freigabe kommt sie dazu.**
//
// **Das ist die Gegenprobe** — ohne sie sagte der Lauf darüber nur, dass die
// Liste kürzer ist, und nicht, dass die Freigabe etwas bewirkt.
func TestMitFreigabeKommtDerEigeneBestandDazu(t *testing.T) {
	t.Parallel()

	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 4, Grenze: 5,
		Freigegeben: []domain.Retrievalquelle{domain.QuelleEigeneKarte},
	}

	if !slices.Contains(anfrage.Suchquellen(), domain.QuelleEigeneKarte) {
		t.Errorf("die Freigabe wirkte nicht: %v", anfrage.Suchquellen())
	}
}

// **Eine Freigabe für eine Quelle, die dieser Fall gar nicht sucht, wird
// abgewiesen.**
//
// # Warum das ein Fehler ist und keine stille Weglassung
//
// **Weil eine Freigabe eine Entscheidung ist.** Wer „mein Regelwerk ist
// freigegeben" ankreuzt und eine Kartensuche fährt, hat etwas anderes
// entschieden, als er glaubt — und bekäme eine Antwort, die aussieht, als sei
// sein Kreuz berücksichtigt worden.
func TestEineFreigabeInsLeereWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallRegeln, Frage: "Wie blockt man?", Grenze: 5,
		Freigegeben: []domain.Retrievalquelle{domain.QuelleEigeneKarte},
	}

	if fehler := anfrage.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und eine Freigabe für eine öffentliche Quelle ebenso.**
//
// **Sie bedeutet nichts**: öffentliche Korpora werden ohnehin gesucht. Ein
// Kreuz ohne Wirkung ist eine Zusage, die niemand einlöst (§10).
func TestEineFreigabeFuerOeffentlichesWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallRegeln, Frage: "Wie blockt man?", Grenze: 5,
		Freigegeben: []domain.Retrievalquelle{domain.QuelleRegelwerk},
	}

	if fehler := anfrage.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Die Anfrage sagt, welche privaten Quellen sie auslässt.**
//
// **Damit es die Antwort sagen kann** — ein Korpus, in dem nicht gesucht
// wurde, ist eine Einschränkung des Ergebnisses (§8, §16 Regel 5).
func TestDieAnfrageNenntWasSieAuslaesst(t *testing.T) {
	t.Parallel()

	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 4, Grenze: 5,
	}

	ausgelassen := anfrage.AusgelasseneQuellen()

	if len(ausgelassen) != 1 || ausgelassen[0] != domain.QuelleEigeneKarte {
		t.Errorf("ausgelassen wurde %v", ausgelassen)
	}

	anfrage.Freigegeben = []domain.Retrievalquelle{domain.QuelleEigeneKarte}

	if len(anfrage.AusgelasseneQuellen()) != 0 {
		t.Errorf("mit Freigabe bleibt %v ausgelassen", anfrage.AusgelasseneQuellen())
	}
}
