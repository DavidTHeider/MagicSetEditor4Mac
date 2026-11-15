// **Eine Massenänderung lässt sich gezielt zurücknehmen** —
// #1393.
//
// # Der Befund
//
// `Zuruecknehmen` nimmt den **jüngsten noch geltenden Schritt des Sets**. Am
// Fenster ist das genau richtig: ⌘Z heißt „das da eben". Über MCP,
// Kommandozeile und REST ist es das nicht — dort schreibt womöglich ein
// zweiter Prozess in dasselbe Set, und dann trifft die Rücknahme **dessen**
// Änderung.
//
// Angestrichen vom Zweitblick über die Phasen 4.0 bis 4.9 (Befund 3).
//
// # Warum die Gruppe schon da ist
//
// Die Massenänderung schreibt ihre Einträge längst unter **einer**
// Gruppenkennung — ohne sie nähme ⌘Z danach die vorige Änderung zurück,
// während die geänderten Karten stehen blieben (#52). Sie wurde nur nie
// zurückgegeben.
//
// # Warum die Klasse und nicht der Fall
//
// Der Vermerk nannte drei Wege, und die ersten beiden hätten die
// Massenänderung geheilt: die Vorwerte mitgeben, oder vorher lesen. Der dritte
// heilt **jede** Änderung — auch die einzelne, denn `schrittkennung` gibt einem
// Eintrag ohne Gruppe die Kennung `einzeln:<id>`.
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Rücknahme trifft den genannten Vorgang, nicht den jüngsten.**
//
// Der Fall ist **hergestellt und nicht beschrieben**: zwischen Massenänderung
// und Rücknahme schreibt ein zweiter Weg in dasselbe Set. Ohne Gruppenkennung
// ginge dessen Änderung zurück, und die dreihundert blieben stehen.
func TestDieRuecknahmeTrifftDenGenanntenVorgang(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)
	schreiber := storage.NeuerKartenschreiber(db)

	vorgang, err := schreiber.FelderSetzen(probekontext(t), karten, "typ", "Zauber")
	if err != nil {
		t.Fatalf("die Massenänderung: %v", err)
	}

	if vorgang == "" {
		t.Fatal("die Massenänderung nennt ihren Vorgang nicht")
	}

	// **Ein anderer schreibt dazwischen** — das ist der ganze Befund.
	//
	// **Über `FeldSetzenMitVorwert` und nicht über `FeldSetzen`**: nur der
	// erste schreibt einen Verlaufseintrag. Der erste Anlauf dieses Tests nahm
	// den zweiten, und dann gab es gar keinen jüngeren Schritt — er wäre auch
	// ohne Gruppenkennung grün gewesen und hätte nichts belegt. Aufgefallen
	// ist es erst an der Gegenprobe: sie blieb grün, wo sie rot sein musste.
	if _, _, zwischenfehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karten[0], "name", "Dazwischen"); zwischenfehler != nil {
		t.Fatalf("die Zwischenänderung: %v", zwischenfehler)
	}

	schritt, bewegt, err := storage.NeuerVerlaufsleser(db).VorgangZuruecknehmen(
		probekontext(t), setID, vorgang)
	if err != nil || !bewegt {
		t.Fatalf("die Rücknahme: bewegt=%v, %v", bewegt, err)
	}

	if schritt.Umfang != len(karten) {
		t.Errorf("der Umfang ist %d, erwartet waren %d", schritt.Umfang, len(karten))
	}

	// **Die Massenänderung ist fort** …
	for stelle, wert := range werte(t, db, setID, "typ") {
		if wert != "«fehlt»" {
			t.Errorf("Karte %d trägt noch den Typ %q", stelle, wert)
		}
	}

	// … **und die Zwischenänderung steht**. Das ist die Hälfte, die eine
	// Rücknahme des jüngsten Schritts falsch macht.
	if namen := werte(t, db, setID, "name"); namen[0] != "Dazwischen" {
		t.Errorf("die Zwischenänderung heißt %q, erwartet war „Dazwischen\"", namen[0])
	}
}

// **Ein Vorgang, den es nicht gibt, bewegt nichts — und ist kein Fehler.**
//
// Dieselbe Zusage wie bei ⌘Z am Anfang des Verlaufs: „nichts mehr da" ist eine
// gewöhnliche Lage. Wer zweimal dieselbe Rücknahme schickt, bekommt beim
// zweiten Mal `bewegt=false` statt einer Meldung.
func TestEinUnbekannterVorgangBewegtNichts(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	if _, err := storage.NeuerKartenschreiber(db).FelderSetzen(
		probekontext(t), karten, "typ", "Zauber"); err != nil {
		t.Fatalf("die Massenänderung: %v", err)
	}

	_, bewegt, err := storage.NeuerVerlaufsleser(db).VorgangZuruecknehmen(
		probekontext(t), setID, "denvorganggibtesnicht")
	if err != nil {
		t.Fatalf("die Rücknahme: %v", err)
	}

	if bewegt {
		t.Error("ein unbekannter Vorgang hat etwas bewegt")
	}

	// **Und die Massenänderung steht noch**: eine Rücknahme ins Leere darf
	// nicht ersatzweise irgendetwas anderes treffen.
	for stelle, wert := range werte(t, db, setID, "typ") {
		if wert != "Zauber" {
			t.Errorf("Karte %d trägt %q statt „Zauber\"", stelle, wert)
		}
	}
}

// **Ein zurückgenommener Vorgang lässt sich wiederherstellen.**
//
// Sonst wäre die gezielte Rücknahme eine Einbahnstraße, und ⌘⇧Z gäbe an den
// Flächen etwas anderes zurück als das, was eben wegging.
func TestEinZurueckgenommenerVorgangLaesstSichWiederherstellen(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)
	verlauf := storage.NeuerVerlaufsleser(db)

	vorgang, err := storage.NeuerKartenschreiber(db).FelderSetzen(
		probekontext(t), karten, "typ", "Zauber")
	if err != nil {
		t.Fatalf("die Massenänderung: %v", err)
	}

	if _, _, err := verlauf.VorgangZuruecknehmen(probekontext(t), setID, vorgang); err != nil {
		t.Fatalf("die Rücknahme: %v", err)
	}

	if _, bewegt, err := verlauf.VorgangWiederherstellen(
		probekontext(t), setID, vorgang); err != nil || !bewegt {
		t.Fatalf("die Wiederherstellung: bewegt=%v, %v", bewegt, err)
	}

	for stelle, wert := range werte(t, db, setID, "typ") {
		if wert != "Zauber" {
			t.Errorf("Karte %d trägt %q statt „Zauber\"", stelle, wert)
		}
	}
}
