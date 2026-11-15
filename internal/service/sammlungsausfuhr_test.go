package service_test

// **Die Sammlung ausgeben und zurücklesen** — Werkstück 4 von Phase 4.5
// (#1292).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// gefuellt gibt eine Ablage mit zwei Exemplaren.
func gefuellt(t *testing.T) *sammlungsdoppel {
	t.Helper()

	ablage := &sammlungsdoppel{}

	for _, exemplar := range []*domain.Exemplar{
		{
			Printing: probeprinting, Menge: 4,
			Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
			Name: "Lightning Bolt", Setcode: "m11", Nummer: "149",
			Erwerbspreis: 417, Erwerbswaehrung: "EUR", Notiz: "getauscht, 2026",
		},
		{
			Printing: "22222222-2222-4222-8222-222222222222", Menge: 1,
			Zustand: domain.ZustandPlayed, Sprache: "de", Finish: domain.FinishFoil,
			Name: "Counterspell", Setcode: "lea", Nummer: "54",
		},
	} {
		if _, fehler := ablage.Anlegen(t.Context(), exemplar); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	return ablage
}

// **Die Liste ist einfügbar — und sagt, was sie nicht trägt.**
func TestDieAusfuhrAlsListe(t *testing.T) {
	t.Parallel()

	text, fehler := service.NeueSammlungsausfuhr(gefuellt(t)).
		Ausgeben(t.Context(), service.FormListe)
	if fehler != nil {
		t.Fatalf("Ausgeben: %v", fehler)
	}

	if !strings.Contains(text, "4 Lightning Bolt (M11) 149") {
		t.Errorf("die erste Zeile fehlt:\n%s", text)
	}

	if !strings.Contains(text, "*F*") {
		t.Errorf("die Foilmarke fehlt:\n%s", text)
	}

	// **Und der Zustand steht nicht darin** — das ist die Zusage, nicht ein
	// Versehen. Ohne diese Zeile könnte jemand die Liste für vollständig
	// halten und die Sicherung darauf stützen.
	if strings.Contains(text, "played") {
		t.Errorf("die Liste trägt den Zustand — sie soll es nicht:\n%s", text)
	}
}

// **Das CSV trägt alles und liest sich vollständig zurück.**
//
// Der Rundlauf ist die Zusage: ausgeben, in einen leeren Bestand einlesen, und
// dieselben Zeilen bekommen.
func TestDerRundlaufUeberCSV(t *testing.T) {
	t.Parallel()

	hinaus := service.NeueSammlungsausfuhr(gefuellt(t))

	text, fehler := hinaus.Ausgeben(t.Context(), service.FormCSV)
	if fehler != nil {
		t.Fatalf("Ausgeben: %v", fehler)
	}

	leer := &sammlungsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(leer).Einfuegen(t.Context(), text, "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Angelegt) != 2 || bericht.Stueck != 5 {
		t.Fatalf("%d angelegt, %d Stück — erwartet 2 und 5: %+v",
			len(bericht.Angelegt), bericht.Stueck, bericht.Befunde)
	}

	zurueck, lesefehler := leer.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	// **Jede Angabe ist mitgekommen**, nicht nur Druck und Menge.
	for _, exemplar := range zurueck {
		if exemplar.Printing != probeprinting {
			continue
		}

		if exemplar.Erwerbspreis != 417 || exemplar.Erwerbswaehrung != "EUR" {
			t.Errorf("der Preis kam als %d %q zurück",
				exemplar.Erwerbspreis, exemplar.Erwerbswaehrung)
		}

		// **Das Komma in der Notiz hat überlebt.**
		if exemplar.Notiz != "getauscht, 2026" {
			t.Errorf("die Notiz kam als %q zurück", exemplar.Notiz)
		}
	}
}

// **Kein Preis bleibt kein Preis** — und wird nicht zu „0.00".
//
// Null heißt „kein Preis erfasst"; eine Null im CSV läse sich beim Zurücklesen
// als geschenkte Karte.
func TestOhnePreisBleibtDieSpalteLeer(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
		Printing: probeprinting, Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	text, fehler := service.NeueSammlungsausfuhr(ablage).
		Ausgeben(t.Context(), service.FormCSV)
	if fehler != nil {
		t.Fatalf("Ausgeben: %v", fehler)
	}

	if strings.Contains(text, "0.00") {
		t.Errorf("ein fehlender Preis wurde zu 0.00:\n%s", text)
	}
}

// **Eine Form, die es nicht gibt, ist ein fachliches Nein.**
func TestEineUnbekannteAusfuhrform(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).
		Ausgeben(t.Context(), "pergament")

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
	}
}

// **Ein Lesefehler des Bestands kommt durch.**
func TestEinLesefehlerBeiDerAusfuhr(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	_, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{fehler: gestolpert}).
		Ausgeben(t.Context(), service.FormCSV)

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}
}

// **Ein CSV ohne eine einzige brauchbare Zeile schreibt nichts — und sagt es.**
func TestEinCSVOhneBrauchbareZeile(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(ablage).
		Einfuegen(t.Context(), "menge,name\n3,Bolt\n", "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Angelegt) != 0 {
		t.Errorf("%d angelegt, erwartet 0", len(bericht.Angelegt))
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}
}

// **Die Befunde des Zerteilers gehen mit dem Bericht hinaus** (§8).
//
// Eine kaputte Zeile hält die übrigen nicht auf — und verschwindet auch nicht.
func TestDieBefundeDesZerteilersGehenMit(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(ablage).Einfuegen(t.Context(),
		"printing,menge,zustand,sprache\n"+probeprinting+",2,nearMint,en\n"+
			"kaputt,keineZahl,nearMint,en\n", "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Errorf("%d angelegt, erwartet 1", len(bericht.Angelegt))
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}

	if bericht.Befunde[0].Ort != "Zeile 3" {
		t.Errorf("der Befund zeigt auf %q", bericht.Befunde[0].Ort)
	}
}

// **Ein unlesbarer Preis ist ein fehlender, kein Befund.**
//
// Eine ganze Sammlung an einer Spalte scheitern zu lassen, die niemand
// braucht, wäre die falsche Strenge — die Karte geht mit, der Preis nicht.
func TestEinUnlesbarerPreis(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(ablage).Einfuegen(t.Context(),
		"printing,menge,zustand,sprache,erwerbspreis,erwerbswaehrung\n"+
			probeprinting+",1,nearMint,en,vierzehn,\n", "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Fatalf("%d angelegt, erwartet 1: %v", len(bericht.Angelegt), bericht.Befunde)
	}

	zeilen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if zeilen[0].Erwerbspreis != 0 {
		t.Errorf("Preis %d, erwartet 0", zeilen[0].Erwerbspreis)
	}
}

// **Eine Zeile, die der Fachkern abweist, wird zum Befund mit ihrer Stelle.**
func TestEineAbgewieseneCSVZeile(t *testing.T) {
	t.Parallel()

	bericht, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).
		Einfuegen(t.Context(), "printing,menge,zustand\n"+probeprinting+",1,erfunden\n", "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}

	if bericht.Befunde[0].Ort != "Zeile 2" {
		t.Errorf("der Befund zeigt auf %q", bericht.Befunde[0].Ort)
	}
}

// **Ein Systemfehler beim Einlesen bricht ab — und der Bericht kommt mit.**
//
// Derselbe Schnitt wie bei der Listeneinfuhr: ein Plattenfehler sagt über
// jede Zeile dasselbe, und dreihundert gleichlautende Befunde sind keine
// Auskunft (§8). Der Bericht geht trotzdem hinaus (§7, Regel 3).
func TestEinSystemfehlerBeimCSVEinlesen(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	bericht, fehler := service.NeueSammlungsausfuhr(&nurSchreibfehler{
		sammlungsdoppel: &sammlungsdoppel{},
		beimSchreiben:   gestolpert,
	}).Einfuegen(t.Context(), "printing,menge,zustand,sprache\n"+
		probeprinting+",1,nearMint,en\n"+probeprinting+"x,1,nearMint,en\n", "")

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}

	if len(bericht.Angelegt) != 0 {
		t.Errorf("%d angelegt, erwartet 0", len(bericht.Angelegt))
	}
}

// **Besitz und Wunsch desselben Drucks bleiben zwei Zeilen.**
//
// # Der Befund, der diesen Test verlangt
//
// `gleicheZeile` verglich **vier** Spalten, die Eindeutigkeit hat seit
// Migration 0026 **fünf**. Ein CSV mit demselben Druck einmal als Besitz und
// einmal als Wunsch legte beim Zurücklesen die zweite Zeile auf die erste —
// der Wunsch verschwand, der Besitz wuchs.
//
// **Der alte Rundlauftest hatte ihn nicht**: er benutzte zwei verschiedene
// Drucke und keine Wunschzeile. Angestrichen vom Zweitblick über Phase 4.5
// als Befund 1.
func TestBesitzUndWunschBleibenZweiZeilen(t *testing.T) {
	t.Parallel()

	quelle := &sammlungsdoppel{}

	for _, wunsch := range []bool{false, true} {
		if _, fehler := quelle.Anlegen(t.Context(), &domain.Exemplar{
			Printing: probeprinting, Menge: 2,
			Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
			Wunsch: wunsch,
		}); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	text, fehler := service.NeueSammlungsausfuhr(quelle).
		Ausgeben(t.Context(), service.FormCSV)
	if fehler != nil {
		t.Fatalf("Ausgeben: %v", fehler)
	}

	leer := &sammlungsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(leer).Einfuegen(t.Context(), text, "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Angelegt) != 2 || bericht.Erhoeht != 0 {
		t.Fatalf("%d angelegt, %d erhöht — erwartet 2 und 0",
			len(bericht.Angelegt), bericht.Erhoeht)
	}

	zurueck, lesefehler := leer.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(zurueck) != 2 {
		t.Fatalf("%d Zeilen, erwartet 2: %+v", len(zurueck), zurueck)
	}

	// **Beide Absichten sind zurückgekommen** — nicht zweimal dieselbe.
	besitz, wunsch := 0, 0

	for _, zeile := range zurueck {
		if zeile.Besitzt() {
			besitz += zeile.Menge
		} else {
			wunsch += zeile.Menge
		}
	}

	if besitz != 2 || wunsch != 2 {
		t.Errorf("%d besessen, %d gewünscht — erwartet 2 und 2", besitz, wunsch)
	}
}
