package sammelliste_test

// **Was hineingeht, muss auch wieder heraus** — Werkstück 4 von Phase 4.5
// (#1292).
//
// # Zwei Formate, weil es zwei Fragen sind
//
//   - **Die Liste** ist die Gestalt, die [Lies] versteht: einfügbar, lesbar,
//     und **verlustbehaftet** — Zustand, Preis und Notiz stehen darin nicht,
//     weil das Format sie nicht kennt.
//   - **Das CSV** ist vollständig. Es ist die Gestalt, in der eine Sammlung
//     dieses Haus verlässt und wieder hereinkommt, ohne dass etwas fehlt.
//
// **Beides zu haben ist kein Luxus.** Wer seine Sammlung in einen
// Deckbaukasten bringen will, braucht die erste; wer sie sichern will, die
// zweite. Eine Ausfuhr, die nur eines kann, zwingt zur Wahl zwischen lesbar
// und vollständig.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// zeilen gibt zwei Exemplare, die sich in allem unterscheiden.
func zeilen() []sammelliste.Ausfuhrzeile {
	return []sammelliste.Ausfuhrzeile{
		{
			Printing: "00000000-0000-4000-8000-000000000001",
			Menge:    4, Zustand: "nearMint", Sprache: "en", Finish: domain.FinishNonfoil,
			Name: "Lightning Bolt", Setcode: "m11", Nummer: "149",
			Erwerbsdatum: "2026-08-30", Erwerbspreis: "4.17",
			Erwerbswaehrung: "EUR", Notiz: "getauscht",
			Lagerort: "Kiste 4",

			// **Die portable Referenz** (A2, #1441, 0086): Fassung, System, Wert.
			Verweisfassung: 1, Kennungssystem: "scryfall",
			Kennung: "00000000-0000-4000-8000-000000000001",
		},
		{
			Printing: "00000000-0000-4000-8000-000000000002",
			Menge:    1, Zustand: "played", Sprache: "de", Finish: domain.FinishFoil,
			Name: "Counterspell", Setcode: "lea", Nummer: "54",
			Wunsch: true,
		},
	}
}

// **Die Liste liest sich zurück — genau das ist ihre Zusage.**
//
// Der Rundlauf ist die Prüfung, die zählt: was [sammelliste.SchreibeListe]
// ausgibt, muss [sammelliste.Lies] wieder als dieselben Karten lesen.
func TestDieListeLiestSichZurueck(t *testing.T) {
	t.Parallel()

	geschrieben := sammelliste.SchreibeListe(zeilen())

	gelesen, befunde := sammelliste.Lies(geschrieben)
	if len(befunde) != 0 {
		t.Fatalf("die eigene Ausgabe gibt Befunde: %v\n%s", befunde, geschrieben)
	}

	if len(gelesen) != 2 {
		t.Fatalf("%d Zeilen gelesen, erwartet 2:\n%s", len(gelesen), geschrieben)
	}

	if gelesen[0].Menge != 4 || gelesen[0].Name != "Lightning Bolt" ||
		gelesen[0].Setcode != "m11" || gelesen[0].Nummer != "149" {
		t.Errorf("die erste Zeile kam als %+v zurück", gelesen[0])
	}

	// **Die Foilmarke überlebt.** Sie ist die einzige Angabe außer Menge,
	// Name und Druck, die das Listenformat trägt.
	if gelesen[1].Finish != domain.FinishFoil {
		t.Errorf("die Foilmarke ist verloren: %q", geschrieben)
	}
}

// **Ohne lesbaren Namen steht die Kennung da.**
//
// Eine Zeile, die nur eine UUID hat, darf nicht als leere Zeile herausfallen —
// sie wäre sonst beim Rundlauf weg.
func TestOhneNamenStehtDieKennungInDerListe(t *testing.T) {
	t.Parallel()

	geschrieben := sammelliste.SchreibeListe([]sammelliste.Ausfuhrzeile{{
		Printing: "00000000-0000-4000-8000-000000000003", Menge: 2,
	}})

	if !strings.Contains(geschrieben, "00000000-0000-4000-8000-000000000003") {
		t.Errorf("die Kennung fehlt: %q", geschrieben)
	}
}

// **Das CSV trägt jede Spalte — und eine Kopfzeile, die sie benennt.**
func TestDasCSVTraegtJedeSpalte(t *testing.T) {
	t.Parallel()

	geschrieben := sammelliste.SchreibeCSV(zeilen())

	kopf := strings.SplitN(geschrieben, "\n", 2)[0]

	for _, spalte := range []string{
		"printing", "menge", "zustand", "sprache", "finish", "foil",
		"name", "setcode",
		"nummer", "erwerbsdatum", "erwerbspreis", "erwerbswaehrung", "notiz",
		"lagerort", "wunsch", "verweisfassung", "kennungssystem", "kennung",
	} {
		if !strings.Contains(kopf, spalte) {
			t.Errorf("die Kopfzeile nennt %q nicht: %q", spalte, kopf)
		}
	}

	if !strings.Contains(geschrieben, "getauscht") {
		t.Errorf("die Notiz fehlt: %q", geschrieben)
	}
}

// **Und es liest sich zurück — vollständig.**
//
// Das ist die Zusage, die die Liste **nicht** gibt: hier kommt jede Angabe
// wieder heraus, auch Zustand, Preis und Notiz.
func TestDasCSVLiestSichVollstaendigZurueck(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV(sammelliste.SchreibeCSV(zeilen()))
	if len(befunde) != 0 {
		t.Fatalf("die eigene Ausgabe gibt Befunde: %v", befunde)
	}

	erwartet := zeilen()

	if len(gelesen) != len(erwartet) {
		t.Fatalf("%d Zeilen, erwartet %d", len(gelesen), len(erwartet))
	}

	for stelle, zeile := range gelesen {
		if zeile != erwartet[stelle] {
			t.Errorf("Zeile %d:\ngelesen  %+v\nerwartet %+v",
				stelle+1, zeile, erwartet[stelle])
		}
	}
}

// **Ein Komma in der Notiz bricht das CSV nicht.**
//
// Der Fall, den ein selbst gebautes `strings.Join(",")` verlöre — und er ist
// nicht erfunden: „getauscht, 2026" ist eine Notiz, die jemand schreibt.
func TestEinKommaInDerNotiz(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV(sammelliste.SchreibeCSV(
		[]sammelliste.Ausfuhrzeile{{
			Printing: "abc", Menge: 1, Zustand: "nearMint", Sprache: "en",
			Notiz: `getauscht, 2026 — "billig"`,
		}}))
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(gelesen) != 1 || gelesen[0].Notiz != `getauscht, 2026 — "billig"` {
		t.Errorf("die Notiz kam als %+v zurück", gelesen)
	}
}

// **Ein CSV mit einer unbekannten Spalte wird gelesen, nicht abgewiesen.**
//
// §16: Pflichtfelder streng, Zusatzfelder tolerant. Wer eine Sammlung aus
// einem Marktplatz einliest, bringt Spalten mit, die hier niemand kennt.
func TestEineUnbekannteSpalte(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV(
		"printing,menge,lagerort\nabc,3,Kiste 4\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(gelesen) != 1 || gelesen[0].Menge != 3 {
		t.Errorf("gelesen: %+v", gelesen)
	}
}

// **Ein CSV von vor 0.5.0 kennt die Referenz nicht — und liest sich trotzdem.**
//
// Die Fassung bleibt dann 0 und sagt genau das: keine Referenz, nur der
// Alias. Und eine Fassung, die keine Zahl ist, ist ebenso keine.
func TestEinAltesCSVOhneReferenzLiestSich(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV(
		"printing,menge,verweisfassung\nabc,3,\ndef,1,zwei\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(gelesen) != 2 || gelesen[0].Verweisfassung != 0 || gelesen[1].Verweisfassung != 0 ||
		gelesen[0].Kennungssystem != "" || gelesen[0].Kennung != "" {
		t.Errorf("gelesen: %+v", gelesen)
	}
}

// **Ohne Pflichtspalte ist es ein Befund** — und die Stelle steht dabei.
func TestEinCSVOhnePflichtspalte(t *testing.T) {
	t.Parallel()

	_, befunde := sammelliste.LiesCSV("menge,name\n3,Bolt\n")

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "printing") {
		t.Errorf("der Befund nennt die fehlende Spalte nicht: %q", befunde[0].Meldung)
	}
}

// **Eine kaputte Zeile hält den Rest nicht auf** (§8).
func TestEineKaputteCSVZeile(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV(
		"printing,menge\nabc,1\ndef,keineZahl\nghi,2\n")

	if len(gelesen) != 2 {
		t.Errorf("%d Zeilen, erwartet 2: %+v", len(gelesen), gelesen)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if befunde[0].Ort != "Zeile 3" {
		t.Errorf("der Befund zeigt auf %q, erwartet \"Zeile 3\"", befunde[0].Ort)
	}
}

// **Ein leeres CSV ist keine Sammlung — und kein Absturz.**
func TestEinLeeresCSV(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV("")

	if len(gelesen) != 0 || len(befunde) != 0 {
		t.Errorf("%d Zeilen und %d Befunde, erwartet 0 und 0", len(gelesen), len(befunde))
	}
}

// **Eine Datei, die kein CSV ist, wird als solche gemeldet.**
//
// Ein unbalanciertes Anführungszeichen bricht die Zerlegung — und dann ist
// nicht eine Zeile kaputt, sondern die Datei. Der Befund sagt das, statt eine
// halbe Sammlung herauszugeben (§8).
func TestEineDateiDieKeinCSVIst(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV("printing,menge\n\"abc,1\n")

	if len(gelesen) != 0 {
		t.Errorf("%d Zeilen, erwartet 0", len(gelesen))
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "kein lesbares CSV") {
		t.Errorf("der Befund lautet %q", befunde[0].Meldung)
	}
}

// **Eine geätzte Zeile geht als `etched` hinaus und kommt als `etched` wieder.**
//
// **Und die alte Spalte sagt `false`** — Werkstück `T2` (#1467). Sie kann
// „geätzt" nicht sagen; wer sie auf `true` setzte, machte aus der Zeile in
// jedem älteren Leser ein gewöhnliches Foil statt einer gewöhnlichen Karte,
// und beides wäre falsch. Was sie nicht ausdrücken kann, drückt sie nicht aus.
func TestEineGeaetzteZeileUeberlebtDasCSV(t *testing.T) {
	t.Parallel()

	geschrieben := sammelliste.SchreibeCSV([]sammelliste.Ausfuhrzeile{{
		Printing: "scryfall:abc", Menge: 1, Zustand: "nearMint",
		Sprache: "en", Finish: domain.FinishEtched,
	}})

	if !strings.Contains(geschrieben, "etched,false") {
		t.Errorf("die Ausfuhr ist %q", geschrieben)
	}

	gelesen, befunde := sammelliste.LiesCSV(geschrieben)
	if len(befunde) != 0 {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	if len(gelesen) != 1 || gelesen[0].Finish != domain.FinishEtched {
		t.Errorf("zurück kam %+v", gelesen)
	}
}

// **Eine Datei von gestern kennt nur `foil` — und wird trotzdem gelesen.**
//
// **Das ist die andere Hälfte derselben Zusage.** Ohne diesen Lauf hielte auch
// eine Fassung, die `foil` beim Lesen gar nicht mehr ansieht: jede aufgehobene
// Ausfuhr würde damit still gewöhnlich.
func TestEineAusfuhrVonGesternWirdWeiterGelesen(t *testing.T) {
	t.Parallel()

	gelesen, befunde := sammelliste.LiesCSV(
		"printing,menge,zustand,sprache,foil\n" +
			"scryfall:abc,1,nearMint,en,true\n")
	if len(befunde) != 0 {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	if len(gelesen) != 1 || gelesen[0].Finish != domain.FinishFoil {
		t.Errorf("zurück kam %+v", gelesen)
	}
}

// **Die Liste trägt die Ausführung als Marke — und liest sie zurück.**
//
// **`*etched*` und nicht `*F*`**: die Liste ist verlustbehaftet, aber was sie
// sagt, soll stimmen. Ein geätztes Stück als Foil zu schreiben wäre keine
// Kürzung, sondern eine andere Karte.
func TestDieListeTraegtDieAusfuehrungAlsMarke(t *testing.T) {
	t.Parallel()

	// **`gelesen` ist nicht immer `geschrieben`**: eine Zeile ohne Marke
	// kommt **ohne Angabe** zurück, und leer heißt `nonfoil` — dieselbe
	// Vorgabe wie im Bestand. Wer hier `nonfoil` erwartete, prüfte, ob der
	// Leser eine Angabe erfindet, die niemand geschrieben hat.
	for name, fall := range map[string]struct {
		finish  domain.Finish
		marke   string
		zurueck domain.Finish
	}{
		"gewöhnlich": {domain.FinishNonfoil, "", ""},
		"foil":       {domain.FinishFoil, " *F*", domain.FinishFoil},
		"geätzt":     {domain.FinishEtched, " *etched*", domain.FinishEtched},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			geschrieben := sammelliste.SchreibeListe([]sammelliste.Ausfuhrzeile{{
				Printing: "scryfall:abc", Menge: 1, Name: "Shivan Dragon",
				Finish: fall.finish,
			}})

			if geschrieben != "1 Shivan Dragon"+fall.marke+"\n" {
				t.Fatalf("geschrieben wurde %q", geschrieben)
			}

			gelesen, befunde := sammelliste.Lies(geschrieben)
			if len(befunde) != 0 {
				t.Fatalf("die Befunde sind %+v", befunde)
			}

			if len(gelesen) != 1 || gelesen[0].Finish != fall.zurueck {
				t.Errorf("zurück kam %+v", gelesen)
			}
		})
	}
}

// **Rang und Art überstehen den Rundlauf** — Werkstück T4 von Phase 5.6
// (#1469).
//
// **Ein CSV, das sie verlöre, wäre keine Sicherung.** Wer eine Dringlichkeit
// von Hand gesetzt hat, bekäme sie nach dem Zurücklesen nicht wieder — und
// die Reihenfolge stünde danach anders da, ohne dass jemand etwas geändert
// hätte.
func TestRangUndArtUeberstehenDenRundlauf(t *testing.T) {
	t.Parallel()

	hin := []sammelliste.Ausfuhrzeile{{
		Printing: "a1b2", Menge: 2, Zustand: "played", Sprache: "de",
		Finish: domain.FinishEtched, Wunsch: true,
		Wunschrang: string(domain.RangSehrHoch),
		Wunschart:  string(domain.WunschJederDruck),
	}}

	zurueck, befunde := sammelliste.LiesCSV(sammelliste.SchreibeCSV(hin))
	if len(befunde) != 0 {
		t.Fatalf("es kamen Befunde: %v", befunde)
	}

	if len(zurueck) != 1 {
		t.Fatalf("es kamen %d Zeilen", len(zurueck))
	}

	if zurueck[0].Wunschrang != string(domain.RangSehrHoch) ||
		zurueck[0].Wunschart != string(domain.WunschJederDruck) {
		t.Errorf("zurück kam %q und %q", zurueck[0].Wunschrang, zurueck[0].Wunschart)
	}
}
