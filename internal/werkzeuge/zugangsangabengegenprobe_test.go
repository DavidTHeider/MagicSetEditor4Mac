package werkzeuge_test

// **Die Gegenproben zu den Zugangsbeschriftungen** —
// #1534.
//
// # Warum sie hier stehen und nicht daneben
//
// **Weil ein Wächter, der nie rot war, nichts belegt.** Die zwei Läufe in
// `zugangsangabenamen_test.go` sind grün — und wären es auch, wenn ihre
// Bedingungen nie zuträfen. Dieselbe Überlegung wie bei
// `internal/schichten/gegenproben_test.go`: dort steht die **Regel**, hier
// steht, dass sie **wirkt**.
//
// # Und sie ändern keine Datei im Baum
//
// **Der Katalog wird erfunden, nicht angefasst.** Ein Eintrag aus
// `seiteneinstellungen.ts` zu löschen, den Lauf rot zu sehen und ihn
// zurückzuschreiben, ist eine Gegenprobe, die niemand wiederholt — und die
// beim zweiten Mal vergessen wird. Deshalb liest [dienstangabenAus] einen
// Text und keine Datei.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einKatalog baut einen Katalogtext mit genau diesen Einträgen.
func einKatalog(eintraege ...string) string {
	zeilen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		zeilen = append(zeilen, "    '"+eintrag+"': 'Name',")
	}

	return "  dienstangaben: {\n" + strings.Join(zeilen, "\n") + "\n  },\n"
}

// **Ein fehlender Dienstname bei einer geteilten Angabeart wird abgewiesen.**
//
// **Das ist der Befund selbst**: `marke` bedient MTGJSON und CardTrader, und
// stünde nur einer der beiden im Katalog, bekäme der andere „Patreon-Marke".
func TestDieGeteilteAngabeartWeistEinenFehlendenDienstAb(t *testing.T) {
	t.Parallel()

	dienstejeAngabe := map[string][]domain.Dienst{}
	for _, dienst := range domain.Dienste() {
		for _, angabe := range dienst.Zugangsangaben() {
			dienstejeAngabe[angabe] = append(dienstejeAngabe[angabe], dienst)
		}
	}

	// **Nur MTGJSON steht da, CardTrader fehlt** — genau der Stand vor der
	// Korrektur, nur mit dem neuen Block.
	jeDienst := dienstangabenAus(t,
		einKatalog(string(domain.DienstMTGJSON)+"."+domain.ZugangMarke),
		"probe")

	klagen := &sammler{}

	pruefeGeteilteMit(klagen, "probe", domain.ZugangMarke,
		dienstejeAngabe[domain.ZugangMarke], jeDienst)

	if len(klagen.saetze) == 0 {
		t.Fatal("der fehlende Dienstname kam durch — dann sagt der Lauf " +
			"nichts über den Befund, für den es ihn gibt")
	}

	// **Und er nennt den Dienst, der fehlt** — eine Klage ohne ihn zwänge
	// zum Suchen.
	if !strings.Contains(strings.Join(klagen.saetze, " "),
		string(domain.DienstCardTrader)) {
		t.Errorf("die Klage nennt CardTrader nicht: %v", klagen.saetze)
	}
}

// **Und eine Angabeart mit nur einem Dienst wird nicht abgewiesen.**
//
// **Die andere Hälfte** (§10): ein Lauf, der alles rot macht, prüft nichts.
// `team` benutzt allein Scrydex — ein eigener Name dafür wäre eine Zeile
// ohne Gewinn.
func TestEineAngabeartMitEinemDienstBrauchtKeinenEigenenNamen(t *testing.T) {
	t.Parallel()

	klagen := &sammler{}

	pruefeGeteilteMit(klagen, "probe", domain.ZugangTeam,
		[]domain.Dienst{domain.DienstScrydex}, map[string]bool{})

	if len(klagen.saetze) != 0 {
		t.Fatalf("die Angabeart eines einzelnen Dienstes wurde abgewiesen: %v",
			klagen.saetze)
	}
}

// **Und eine ausdrücklich geteilte ebenso wenig.**
func TestEineBegruendetGeteilteAngabeartWirdNichtAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, geteilt := geteilteAngaben[domain.ZugangSchluessel]; !geteilt {
		t.Fatal("`schluessel` steht nicht in `geteilteAngaben` — dann prüft " +
			"dieser Lauf eine Ausnahme, die es nicht gibt")
	}

	klagen := &sammler{}

	pruefeGeteilteMit(klagen, "probe", domain.ZugangSchluessel,
		[]domain.Dienst{domain.DienstScrydex, domain.DienstAnthropic},
		map[string]bool{})

	if len(klagen.saetze) != 0 {
		t.Fatalf("die begründete Ausnahme wurde abgewiesen: %v", klagen.saetze)
	}
}

// **Ein Schlüssel, der keinen wirklichen Dienst nennt, wird abgewiesen.**
//
// **Der stille Rückfall**: `cardtradr.marke` sähe im Katalog aus wie ein
// Eintrag und wirkte im Fenster nicht — der Befund wäre zurück, mit einem
// grünen Wächter davor.
func TestEinErfundenerDienstangabenschluesselWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, schluessel := range map[string]string{
		"vertippter Dienst": "cardtradr." + domain.ZugangMarke,
		"vertippte Angabe":  string(domain.DienstCardTrader) + ".mark",
		"Dienst ohne diese Angabe": string(domain.DienstScryfall) + "." +
			domain.ZugangMarke,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			klagen := &sammler{}

			pruefeSchluesselMit(klagen, "probe",
				dienstangabenAus(t, einKatalog(schluessel), "probe"),
				echteDienstangaben())

			if len(klagen.saetze) == 0 {
				t.Fatalf("%q kam durch — im Fenster fiele er still auf den "+
					"Rückfall zurück", schluessel)
			}

			if !strings.Contains(klagen.saetze[0], schluessel) {
				t.Errorf("die Klage nennt den Schlüssel nicht: %q",
					klagen.saetze[0])
			}
		})
	}
}

// **Und ein wirklicher Schlüssel kommt durch.**
//
// **Die andere Hälfte** (§10): ein Lauf, der jeden Schlüssel abweist, hielte
// die Zusage darüber, ohne etwas zu prüfen.
func TestEinWirklicherDienstangabenschluesselKommtDurch(t *testing.T) {
	t.Parallel()

	klagen := &sammler{}

	pruefeSchluesselMit(klagen, "probe", dienstangabenAus(t, einKatalog(
		string(domain.DienstCardTrader)+"."+domain.ZugangMarke), "probe"),
		echteDienstangaben())

	if len(klagen.saetze) != 0 {
		t.Fatalf("ein wirklicher Schlüssel wurde abgewiesen: %v", klagen.saetze)
	}
}

// sammler nimmt Klagen entgegen, statt einen Lauf rot zu machen.
type sammler struct{ saetze []string }

func (s *sammler) klage(satz string) { s.saetze = append(s.saetze, satz) }
