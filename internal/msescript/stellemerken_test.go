package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die Ordnung wird einmal gebaut, nicht je Frage** — `Set::order_cache`
// (`src/data/set.cpp`, Zeilen 277–299 und 315–318, Fassung f9d9356 vom
// 2021-07-06):
//
//	OrderCacheP& order = order_cache[make_pair(order_by,filter)];
//	if (!order) { … order = make_intrusive<OrderCache<CardP>>(…); }
//	return order->find(card);
//
// Ohne den Speicher rechnete jede der 3.788 Karten von `Black.mse-set` den
// Schlüssel aller 3.788 — gut vierzehn Millionen Mal statt 3.788.

// schluesselzaehler gibt, wie oft ein Schlüsselausdruck gelaufen ist: er meldet
// sich mit `trace` bei jedem Lauf.
func schluesselzaehler(umgebung *msescript.Umgebung) int {
	return len(umgebung.Meldungen())
}

func TestPositionBautDieOrdnungEinmalWennDerSpeicherDaIst(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton", "Berta", "Cäsar")
	umgebung.OrdnungenMerken()

	// **Die Funktion steht einmal da, und ihr Block ist derselbe Knoten bei
	// jedem Ruf** — wie `card_number` im Initskript des Spiels. Mit einer Quelle,
	// die bei jeder Frage neu zerteilt würde, hinge der Test am Zwischenspeicher
	// der Bäume, den ein anderer Test leeren darf.
	if err := msescript.Vorbelegen(t.Context(),
		`nummer := { position(of: card, in: set, order_by: {trace(card.name)}) }`,
		umgebung); err != nil {
		t.Fatalf("erklären: %v", err)
	}

	for _, karte := range karten {
		eigene := umgebung.Darueber()
		eigene.Setze("card", karte)

		if _, err := msescript.AuswertenQuelle(t.Context(), `nummer()`, eigene); err != nil {
			t.Fatalf("auswerten: %v", err)
		}
	}

	if gelaufen := schluesselzaehler(umgebung); gelaufen != len(karten) {
		t.Errorf("der Schlüssel lief %d-mal, erwartet %d", gelaufen, len(karten))
	}
}

func TestPositionRechnetOhneSpeicherJedesMalNeu(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton", "Berta")

	for _, karte := range karten {
		stelleimSet(t, umgebung, karte, `, order_by: {trace(card.name)}`)
	}

	// Zwei Fragen, je zwei Karten.
	if gelaufen := schluesselzaehler(umgebung); gelaufen != 4 {
		t.Errorf("der Schlüssel lief %d-mal, erwartet 4", gelaufen)
	}
}

func TestPositionUnterscheidetFilterBeimMerken(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Anton", "Berta")
	umgebung.OrdnungenMerken()

	alle := stelleimSet(t, umgebung, karten[1], `, order_by: {card.name}`)
	nurBerta := stelleimSet(t, umgebung, karten[1],
		`, order_by: {card.name}, filter: {card.name == "Berta"}`)

	if alle != "1" || nurBerta != "0" {
		t.Errorf("ohne Filter %s, mit Filter %s — erwartet 1 und 0", alle, nurBerta)
	}
}

func TestPositionVergisstDieOrdnungAufWunsch(t *testing.T) {
	t.Parallel()

	umgebung, karten := umgebungMitSet("Berta", "Anton")
	umgebung.OrdnungenMerken()

	if gesehen := stelleimSet(t, umgebung, karten[0], `, order_by: {card.name}`); gesehen != "1" {
		t.Fatalf("Stelle %s, erwartet 1", gesehen)
	}

	// Ein anderes Set unter demselben Ausdruck — wie die zweite Runde der
	// Ausfuhr, deren Karten inzwischen ihre gerechneten Werte tragen.
	neu, neueKarten := satzmitKarten("Anton", "Berta")
	umgebung.Setze("set", neu)

	umgebung.OrdnungenVergessen()

	if gesehen := stelleimSet(t, umgebung, neueKarten[0], `, order_by: {card.name}`); gesehen != "0" {
		t.Errorf("nach dem Vergessen: Stelle %s, erwartet 0", gesehen)
	}
}

// Vergessen ohne Speicher ist kein Fehler.
func TestOrdnungenVergessenOhneSpeicherTutNichts(t *testing.T) {
	t.Parallel()

	msescript.Bibliothek().Darueber().OrdnungenVergessen()
}
