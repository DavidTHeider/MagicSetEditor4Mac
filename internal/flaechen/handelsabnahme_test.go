package flaechen_test

// **Ein Handel über einen echten Draht, so weit er kommt** — Werkstück `T12`
// von Phase 5.6
// (#1477).
//
// # Was hier gemessen wird
//
// **Ob zwei Prozesse mit getrennten Beständen sich auf denselben Stand
// bringen.** Der Lauf daneben (`handelsdraht_test.go`) trägt **eine**
// Nachricht; hier geht es hin **und zurück** — und erst daran zeigt sich, ob
// der Vertrag aus `T1` über eine Leitung überhaupt vorankommt.

import (
	"maps"
	"strings"
	"testing"
	"time"
)

// derAbnahmevorgang ist der Handel dieser Abnahme.
const derAbnahmevorgang = "v-abnahme"

// zweiRechnerAmDraht stellt beide auf, lässt beide horchen und gibt ihre
// Adressen bekannt.
func zweiRechnerAmDraht(t *testing.T) (anna, bert einRechner) {
	t.Helper()

	anna, bert = stelleAuf(t), stelleAuf(t)

	kenneUndErlaube(t, anna, bert)
	kenneUndErlaube(t, bert, anna)

	// **Beide horchen** — eine Zustimmung muss in beide Richtungen reisen,
	// und ein Rechner, der nur wählt, hört die Antwort nicht.
	annasAdresse, _ := horcheImHintergrund(t, anna)
	bertsAdresse, _ := horcheImHintergrund(t, bert)

	mussKommandozeile(t, bert.umgebung,
		"kontakt-adresse", anna.kennung, annasAdresse)
	mussKommandozeile(t, anna.umgebung,
		"kontakt-adresse", bert.kennung, bertsAdresse)

	return anna, bert
}

// trageHinueber legt den Stand in den Ausgang und trägt ihn hinaus.
func trageHinueber(t *testing.T, von einRechner, zu string) {
	t.Helper()

	mussKommandozeile(t, von.umgebung,
		"handelspost-schicken", derAbnahmevorgang, zu)
	mussKommandozeile(t, von.umgebung, "handelszustellung", zu)
}

// warteAufZustand liest, bis der Vorgang drüben dort steht.
func warteAufZustand(
	t *testing.T, bei einRechner, zustand string,
) {
	t.Helper()

	for versuch := range 50 {
		aus, klage, code := kommandozeile(t, bei.umgebung,
			"handel", "stand", derAbnahmevorgang)
		if code == 0 && strings.Contains(aus, zustand) {
			return
		}

		if versuch == 49 {
			t.Fatalf("der Vorgang steht nicht auf %q: %q / %q",
				zustand, aus, klage)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// **Zwei Prozesse einigen sich über einen echten Draht.**
//
// **Das ist die Abnahme, für die es `T12` gibt**: bis hierher war jede
// Zustimmung eine, die im selben Prozess eingetragen wurde.
func TestZweiProzesseEinigenSichUeberEinenEchtenDraht(t *testing.T) {
	anna, bert := zweiRechnerAmDraht(t)

	positionen := `[{"von":"` + anna.kennung +
		`","printing":"bolt-m10","menge":1,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"}]`

	mitEingabe(t, anna.umgebung, positionen,
		"handel", "eroeffnen", derAbnahmevorgang, "tausch", bert.kennung)
	mussKommandozeile(t, anna.umgebung,
		"handel", "anbieten", derAbnahmevorgang)

	trageHinueber(t, anna, bert.kennung)
	warteAufZustand(t, bert, "angeboten")

	// **Bert stimmt auf seinem Rechner zu** — und trägt es hinüber.
	mussKommandozeile(t, bert.umgebung,
		"handel", "zustimmen", derAbnahmevorgang)

	trageHinueber(t, bert, anna.kennung)

	// **Annas Stimme fehlt noch**, also steht es drüben wie hier.
	mussKommandozeile(t, anna.umgebung,
		"handel", "zustimmen", derAbnahmevorgang)

	warteAufZustand(t, anna, "angenommen")

	trageHinueber(t, anna, bert.kennung)
	warteAufZustand(t, bert, "angenommen")
}

// **Und sie kommen bis zur offenen Übergabe** — über denselben Draht —,
// **und danach bis zum Abschluss** (#2285).
//
// **Weiter als bis zur offenen Übergabe trägt das Protokoll nicht**, und
// das ist keine Lücke dieses Laufs: `verbuchen`, `abgleichen` und
// `bestaetigen` sind **Abschlussschritte**, jede Seite geht sie mit ihrem
// eigenen Nachweis, und die Gegenzeichnung von drüben reist als Beleg — sie
// anzuwenden ist gebaut nicht vorhanden und steht in `docs/ABNAHME-T12.md`.
// Der Lauf fährt sie deshalb **auf jeder Maschine**, wie ein Mensch es täte
// (`schliesseAb`), und vergleicht danach die Bestände.
func TestZweiProzesseKommenBisZumAbschluss(t *testing.T) {
	anna, bert := zweiRechnerAmDraht(t)

	// **Beide geben etwas weg** — ein echter Tausch, und genau der war bis
	// zum Zweitblick über Phase 5.6 nicht abschließbar: nur der Rechner, der
	// die zweite Übergabestimme **selbst** abgab, vermerkte seine Zusagen.
	//
	// **In der vereinbarten Variante** (#2492): ohne Angabe ist ein Exemplar
	// englisch, und der Tausch nennt Deutsch — festgehalten wurde es
	// trotzdem, und die Belege passten danach nicht zusammen.
	annasStueck := eineKennungAus(t, mussKommandozeile(t, anna.umgebung,
		"--json", "exemplar", "bolt-m10", "3", "nearMint", "de", "nonfoil"))
	bertsStueck := eineKennungAus(t, mussKommandozeile(t, bert.umgebung,
		"--json", "exemplar", "swamp-m10", "2", "nearMint", "de", "nonfoil"))

	positionen := `[{"von":"` + anna.kennung +
		`","printing":"bolt-m10","menge":1,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"},{"von":"` + bert.kennung +
		`","printing":"swamp-m10","menge":1,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"}]`

	mitEingabe(t, anna.umgebung, positionen,
		"handel", "eroeffnen", derAbnahmevorgang, "tausch", bert.kennung)
	mussKommandozeile(t, anna.umgebung,
		"handel", "anbieten", derAbnahmevorgang)

	trageHinueber(t, anna, bert.kennung)
	warteAufZustand(t, bert, "angeboten")

	mussKommandozeile(t, bert.umgebung,
		"handel", "zustimmen", derAbnahmevorgang)
	trageHinueber(t, bert, anna.kennung)

	mussKommandozeile(t, anna.umgebung,
		"handel", "zustimmen", derAbnahmevorgang)
	warteAufZustand(t, anna, "angenommen")

	trageHinueber(t, anna, bert.kennung)
	warteAufZustand(t, bert, "angenommen")

	// **Jeder hält sein eigenes Stück fest** — `vorbereiten` verlangt
	// Deckung für das, was von **hier** weggeht (#1643).
	mussKommandozeile(t, anna.umgebung, "handel-festhalten",
		derAbnahmevorgang, annasStueck, "1", "1")
	mussKommandozeile(t, bert.umgebung, "handel-festhalten",
		derAbnahmevorgang, bertsStueck, "1", "1")

	mussKommandozeile(t, anna.umgebung,
		"handel", "vorbereiten", derAbnahmevorgang)

	trageHinueber(t, anna, bert.kennung)
	warteAufZustand(t, bert, "vorbereitet")

	// **Übergeben braucht beide Stimmen, und sie reisen.**
	mussKommandozeile(t, bert.umgebung,
		"handel", "uebergeben", derAbnahmevorgang)
	trageHinueber(t, bert, anna.kennung)

	mussKommandozeile(t, anna.umgebung,
		"handel", "uebergeben", derAbnahmevorgang)
	warteAufZustand(t, anna, "uebergabeOffen")

	trageHinueber(t, anna, bert.kennung)
	warteAufZustand(t, bert, "uebergabeOffen")

	// **Und die Zusage gilt auf beiden Rechnern als übergeben** — daran
	// findet jede Verbuchung ihre Abgangszeile.
	//
	// **Bei Bert entstand der Wechsel über eine Nachricht**, nicht über
	// einen Schritt von Hand; bis zum Zweitblick über Phase 5.6 geschah dort
	// an den Exemplaren nichts, und `verbuchen` scheiterte mit Unterdeckung.
	for _, wer := range []einRechner{anna, bert} {
		zusagen := mussKommandozeile(t, wer.umgebung,
			"--json", "handelszusagen", derAbnahmevorgang)

		if !strings.Contains(zusagen, `"uebergeben": true`) {
			t.Errorf("die Zusage gilt nicht als übergeben: %s", zusagen)
		}
	}

	for _, wer := range []einRechner{anna, bert} {
		schliesseAb(t, wer, anna.kennung, bert.kennung)
	}

	// **Und die Karten haben den Besitzer gewechselt** — je ein Stück, in
	// beide Richtungen.
	for _, fall := range []struct {
		wer    einRechner
		mengen map[string]int
	}{
		{anna, map[string]int{"bolt-m10": 2, "swamp-m10": 1}},
		{bert, map[string]int{"bolt-m10": 1, "swamp-m10": 1}},
	} {
		if hat := mengenJeDruck(t, fall.wer); !maps.Equal(hat, fall.mengen) {
			t.Errorf("der Bestand ist %v, erwartet %v", hat, fall.mengen)
		}
	}
}
