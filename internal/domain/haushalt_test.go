package domain_test

import (
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ressourcenbudgets ohne gegenseitige Aussperrung** — Werkstück O5 von
// Phase 5.2 (#1429).

func TestEinGuthabenFuelltSichMitDerZeitNach(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	kontingent := domain.Kontingent{Burst: 10, Nachfuellen: 60, Fenster: time.Minute}
	guthaben := domain.NeuesGuthaben(kontingent, start)

	// Zehn Rufe zu je einer Einheit gehen durch, der elfte nicht — und die
	// Wartezeit ist die bis zur nächsten Einheit: eine Sekunde.
	for i := range 10 {
		if ok, _ := guthaben.Abbuchen(1, start); !ok {
			t.Fatalf("Ruf %d abgewiesen", i+1)
		}
	}

	ok, warten := guthaben.Abbuchen(1, start)
	if ok || warten != time.Second {
		t.Errorf("elfter Ruf: ok=%v warten=%v", ok, warten)
	}

	// Nach einer halben Sekunde reicht es noch nicht, nach einer schon.
	if ok, _ := guthaben.Abbuchen(1, start.Add(500*time.Millisecond)); ok {
		t.Errorf("nach einer halben Sekunde durchgelassen")
	}

	if ok, _ := guthaben.Abbuchen(1, start.Add(time.Second)); !ok {
		t.Errorf("nach einer Sekunde abgewiesen")
	}

	// **Der Burst ist die Decke**: nach einer Stunde stehen nicht 3.600 bereit.
	if guthaben.Stand(start.Add(time.Hour)) != 10 {
		t.Errorf("Stand nach einer Stunde: %v", guthaben.Stand(start.Add(time.Hour)))
	}
}

func TestEinGuthabenVertraegtUhrspruengeUndGrosseKosten(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	guthaben := domain.NeuesGuthaben(domain.Kontingent{Burst: 5, Nachfuellen: 5, Fenster: time.Minute}, start)

	// **Eine Uhr, die zurückspringt, füllt nichts ab** — und lässt nichts negativ.
	if ok, _ := guthaben.Abbuchen(5, start); !ok {
		t.Fatal("Burst nicht abbuchbar")
	}

	if stand := guthaben.Stand(start.Add(-time.Hour)); stand != 0 {
		t.Errorf("nach Rücksprung: %v", stand)
	}

	// Kosten über dem Burst gehen nie durch — und die Wartezeit sagt es nicht
	// als „unendlich", sondern als die Zeit bis zur vollen Decke.
	ok, warten := guthaben.Abbuchen(7, start.Add(time.Hour))
	if ok || warten != time.Minute {
		t.Errorf("über dem Burst: ok=%v warten=%v", ok, warten)
	}

	// Null und Negatives kosten nichts.
	if ok, _ := guthaben.Abbuchen(0, start.Add(-time.Hour)); !ok {
		t.Errorf("null Kosten abgewiesen")
	}
}

func TestDieSperrlageVerdoppeltSichBisZumDeckel(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	var lage domain.Sperrlage

	if gesperrt, _ := lage.Gesperrt(start); gesperrt {
		t.Fatal("frisch und schon gesperrt")
	}

	// Die ersten drei Fehlversuche sperren nicht — ein Tippfehler ist kein Angriff.
	for range domain.SperreFreieVersuche {
		lage = lage.Fehlgeschlagen(start)
	}

	if gesperrt, _ := lage.Gesperrt(start); gesperrt {
		t.Errorf("nach %d Versuchen gesperrt", domain.SperreFreieVersuche)
	}

	erwartet := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	for i, dauer := range erwartet {
		lage = lage.Fehlgeschlagen(start)

		gesperrt, rest := lage.Gesperrt(start)
		if !gesperrt || rest != dauer {
			t.Errorf("Versuch %d: gesperrt=%v rest=%v, erwartet %v", i+1, gesperrt, rest, dauer)
		}
	}

	// **Der Deckel**: nach vielen Versuchen bleibt es bei der Höchstdauer.
	for range 20 {
		lage = lage.Fehlgeschlagen(start)
	}

	if _, rest := lage.Gesperrt(start); rest != domain.SperreHoechstens {
		t.Errorf("Deckel: %v", rest)
	}

	// Nach Ablauf ist die Sperre vorbei; die Restzeit ist nie negativ.
	if gesperrt, rest := lage.Gesperrt(start.Add(domain.SperreHoechstens + time.Second)); gesperrt || rest != 0 {
		t.Errorf("nach Ablauf: gesperrt=%v rest=%v", gesperrt, rest)
	}

	// Ein Erfolg löscht alles.
	if frisch := lage.Gelungen(); frisch.Fehlversuche != 0 || !frisch.GesperrtBis.IsZero() {
		t.Errorf("nach Erfolg: %+v", frisch)
	}
}

func TestDieVorgabegrenzenSindBenanntUndPositiv(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()

	for _, klasse := range domain.Ressourcenklassen() {
		if grenzen.Parallel[klasse] <= 0 {
			t.Errorf("%s ohne Parallelitätsgrenze", klasse)
		}
	}

	// **Und die Gegenrichtung.** Der Haushalt legt seine Kanäle nach
	// `Parallel` an; eine Grenze für eine Klasse, die niemand aufzählt, wäre
	// eine Grenze, die greift, ohne dass sie irgendwo steht — und ihr
	// Gegenstück, eine aufgezählte Klasse ohne Grenze, ist **unbegrenzt**
	// (A15, #1514).
	// Bis zum 2026-09-12 hielt dieser Lauf nur die eine Hälfte.
	for klasse := range grenzen.Parallel {
		if !slices.Contains(domain.Ressourcenklassen(), klasse) {
			t.Errorf("%s hat eine Grenze, wird aber nicht aufgezählt", klasse)
		}
	}

	if grenzen.Antwort <= 0 || grenzen.Principal.Burst <= 0 || grenzen.Global.Burst <= grenzen.Reserve {
		t.Errorf("Grenzen: %+v", grenzen)
	}

	// Jede Kostenklasse kostet etwas, und Lesen ist die billigste — **bis
	// auf den Abbruch**, der Last nimmt statt macht (#2003).
	for _, klasse := range domain.Kostenklassen() {
		if klasse != domain.KostenAbbruch && klasse.Kosten() < domain.KostenLesen.Kosten() {
			t.Errorf("%s billiger als Lesen", klasse)
		}
	}
}

func TestEinKontingentOhneZulaufGibtNurDenBurst(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	guthaben := domain.NeuesGuthaben(domain.Kontingent{Burst: 2}, start)

	for range 2 {
		if ok, _ := guthaben.Abbuchen(1, start); !ok {
			t.Fatal("Burst nicht abbuchbar")
		}
	}

	// **Ohne Zulauf kommt nichts nach** — auch nach einer Stunde nicht, und die
	// Wartezeit ist das Fenster, also null: die Auskunft „nie" als Zahl.
	if ok, warten := guthaben.Abbuchen(1, start.Add(time.Hour)); ok || warten != 0 {
		t.Errorf("ohne Zulauf: ok=%v warten=%v", ok, warten)
	}

	// Nichts dauert nicht.
	if dauer := (domain.Kontingent{Burst: 2, Nachfuellen: 2, Fenster: time.Minute}).Dauer(0); dauer != 0 {
		t.Errorf("Dauer für nichts: %v", dauer)
	}

	// Eine unbekannte Kostenklasse kostet, was Lesen kostet.
	if domain.Kostenklasse("fremd").Kosten() != domain.KostenLesen.Kosten() {
		t.Errorf("fremde Klasse: %d", domain.Kostenklasse("fremd").Kosten())
	}
}

// **Auftrags- und Verbindungsgrenze stehen bei den übrigen** — #2003. Bis zum
// 2026-10-02 standen sie als Konstanten in `internal/rest` und `cmd/mse-rest`,
// und keine Auskunft nannte sie: wer die vier Aufträge kennenlernen wollte,
// las sie aus dem Satz der Ablehnung, die sechzehn Verbindungen gar nicht.
func TestAuftragsUndVerbindungsgrenzeSindVorgaben(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()

	if grenzen.Verbindungen != 16 {
		t.Errorf("Verbindungen: %d", grenzen.Verbindungen)
	}

	if grenzen.Auftraege != (domain.Auftragsgrenzen{Laufend: 4, Behalten: 16, Aufbewahrung: time.Hour}) {
		t.Errorf("Aufträge: %+v", grenzen.Auftraege)
	}

	// **Und der Schalter der Kontingente lässt sie stehen** — er nimmt nur
	// die Guthaben (0161).
	ohne := grenzen.NachEinstellungen(map[string]string{domain.GrenzenKontingente: domain.KontingenteAus})
	if ohne.Verbindungen != grenzen.Verbindungen || ohne.Auftraege != grenzen.Auftraege {
		t.Errorf("ohne Kontingente: %+v", ohne)
	}
}
