package service_test

// **Die Formatprüfung fragt gebündelt** — Befund 6 des Zweitblicks über Phase
// 4.6, #1373.
//
// # Was der Befund sagte
//
// Die Prüfung rief `KarteNachKennung` je Platz. Bei kaltem Zwischenspeicher
// sind das für ein Commander-Deck hundert Anfragen nacheinander, und die
// Kartenwege sind auf 600 ms gedrosselt (§16, Regel 4) — rund eine Minute,
// bevor das erste Urteil steht.
//
// **Was hier geprüft wird, ist die Zahl der Rufe und nicht die Zeit.** Eine
// Sekundenzahl in einem Test misst diesen Rechner unter dieser Last; die
// Messung am echten Dienst steht im Abnahmelauf des Issues.

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// druckNr gibt den n-ten Druck einer Reihe unterscheidbarer UUIDs.
func druckNr(n int) string {
	return fmt.Sprintf("a1b2c3d4-0000-0000-0000-%012d", n)
}

// **Achtzig verschiedene Drucke sind zwei Anfragen, nicht achtzig.**
func TestVieleDruckeWerdenGebuendeltGefragt(t *testing.T) {
	t.Parallel()

	werk := neuesPruefwerk(t)

	const drucke = 80

	for nummer := range drucke {
		werk.drucke.legalitaeten[druckNr(nummer)] = map[string]string{"standard": "legal"}
		werk.drucke.namen[druckNr(nummer)] = fmt.Sprintf("Karte %d", nummer)
		werk.legeDruck(t, druckNr(nummer), 1)
	}

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if werk.drucke.gefragt != 0 {
		t.Errorf("%d Einzelanfragen — erwartet keine", werk.drucke.gefragt)
	}

	// **Zwei Schübe**: 80 Kennungen bei einer Grenze von 75.
	if werk.drucke.gesammelt != 2 {
		t.Errorf("%d Sammelanfragen (%v) — erwartet 2",
			werk.drucke.gesammelt, werk.drucke.schuebe)
	}

	for _, groesse := range werk.drucke.schuebe {
		if groesse > scryfall.Sammelgrenze {
			t.Errorf("ein Schub trug %d Kennungen — Scryfall nimmt %d",
				groesse, scryfall.Sammelgrenze)
		}
	}

	// **Und das Ergebnis bleibt dasselbe.** Eine Bündelung, die schneller ist
	// und anders antwortet, hat nichts gelöst.
	if len(urteil.Urteile) != drucke {
		t.Fatalf("%d Urteile, erwartet %d", len(urteil.Urteile), drucke)
	}

	for _, eines := range urteil.Urteile {
		if eines.Zulaessigkeit != domain.ZulaessigLegal {
			t.Errorf("Platz %d ist %q (%s)", eines.PlatzID, eines.Zulaessigkeit, eines.Grund)
		}
	}
}

// **Vier Ausfertigungen desselben Drucks sind eine Kennung.**
//
// Der alte Weg fragte je Platz; vier Wälder waren vier Rufe, von denen drei
// aus dem Zwischenspeicher kamen. Gebündelt wäre es eine Kennung viermal im
// selben Rumpf — und Scryfall zählt sie gegen die Grenze von 75.
func TestDerselbeDruckWirdNurEinmalGefragt(t *testing.T) {
	t.Parallel()

	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.drucke.typzeilen[druckA] = "Basic Land — Forest"
	werk.drucke.namen[druckA] = "Forest"

	for range 4 {
		werk.legeDruck(t, druckA, 1)
	}

	if _, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID); err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if werk.drucke.gesammelt != 1 || werk.drucke.schuebe[0] != 1 {
		t.Errorf("%d Sammelanfragen mit den Größen %v — erwartet eine mit einer Kennung",
			werk.drucke.gesammelt, werk.drucke.schuebe)
	}
}

// **Ein Druck, den Scryfall nicht kennt, bleibt ungeprüft** — und nimmt die
// übrigen nicht mit.
//
// Beim Einzelweg war das ein 404 je Karte. Im Sammelweg steht er unter
// `not_found`, und wer die Antwort über den Index zuordnete, verschöbe alles
// Folgende (nachgelesen, §16 Regel 1 — der Satz steht in `sammlung.go`).
func TestEinUnbekannterDruckBleibtUngeprueft(t *testing.T) {
	t.Parallel()

	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.drucke.namen[druckA] = "Lightning Bolt"
	werk.drucke.unbekannt[druckB] = true

	werk.legeDruck(t, druckB, 1) // der unbekannte zuerst
	werk.legeDruck(t, druckA, 1)

	urteil, befunde, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigUngeprueft {
		t.Errorf("der unbekannte Druck ist %q", urteil.Urteile[0].Zulaessigkeit)
	}

	if urteil.Urteile[0].Grund == "" {
		t.Error("er nennt keinen Grund")
	}

	// **Der zweite bekommt trotzdem sein Wort** — nicht das des ersten.
	if urteil.Urteile[1].Zulaessigkeit != domain.ZulaessigLegal {
		t.Errorf("der bekannte Druck ist %q (%s)",
			urteil.Urteile[1].Zulaessigkeit, urteil.Urteile[1].Grund)
	}

	if urteil.Urteile[1].Bezeichnung != "Lightning Bolt" {
		t.Errorf("er heißt %q — die Zuordnung ist verschoben",
			urteil.Urteile[1].Bezeichnung)
	}

	if len(befunde) == 0 {
		t.Error("kein Befund zum unbekannten Druck")
	}
}

// **Scheitert der Sammelweg, springt der Einzelweg ein.**
//
// §16, Regel 5: „Ohne Netz arbeitet die Anwendung weiter. Der letzte gute
// Stand bleibt." Der Einzelweg holt ihn aus dem Zwischenspeicher je Druck; der
// Sammelweg kennt nur den Schub als Ganzes. Ohne diesen Rückfall wäre ein
// Deck, das gestern noch geprüft werden konnte, heute ohne Netz **ganz**
// ungeprüft — eine stille Verschlechterung gegenüber dem Einzelweg.
func TestOhneSammelwegSpringtDerEinzelwegEin(t *testing.T) {
	t.Parallel()

	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.drucke.namen[druckA] = "Lightning Bolt"
	werk.drucke.sammelfehler = errors.New("die Gegenstelle schweigt")

	werk.legeDruck(t, druckA, 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if werk.drucke.gefragt != 1 {
		t.Errorf("%d Einzelanfragen — erwartet eine als Rückfall", werk.drucke.gefragt)
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigLegal {
		t.Errorf("die Karte ist %q (%s) — der Rückfall trug nicht",
			urteil.Urteile[0].Zulaessigkeit, urteil.Urteile[0].Grund)
	}
}

// **Eine groß geschriebene Druckkennung wird trotzdem beurteilt.**
//
// # Warum das nicht von selbst geht
//
// Scryfall **nimmt** eine UUID mit großen Hexadezimalbuchstaben an und
// antwortet kanonisch klein — nachgemessen im Nachlauf über Phase 3.4, der
// Satz steht in `scryfall/kartenachkennung.go`. Ein Deck, das die Kennung groß
// hält, stellte damit eine Frage, deren Antwort anders heißt; ohne Glättung
// fände die Zuordnung sie nicht wieder, und **jede** Karte eines solchen Decks
// bliebe ungeprüft.
//
// **Beim Einzelweg fiel das nicht auf**, weil `druckadresse` dort glättete,
// bevor gefragt wurde, und die Antwort niemand mehr zuordnen musste. Die
// Bündelung braucht die Zuordnung — also braucht sie die Glättung.
func TestEineGrossGeschriebeneKennungWirdBeurteilt(t *testing.T) {
	t.Parallel()

	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.drucke.namen[druckA] = "Lightning Bolt"

	werk.legeDruck(t, strings.ToUpper(druckA), 1)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	if urteil.Urteile[0].Zulaessigkeit != domain.ZulaessigLegal {
		t.Errorf("die Karte ist %q (%s) — die Antwort fand ihre Frage nicht wieder",
			urteil.Urteile[0].Zulaessigkeit, urteil.Urteile[0].Grund)
	}
}
