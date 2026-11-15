package service_test

// **Die portable Referenz kommt an** — Werkstück A2 von Phase 5.3
// (#1441, 0086).
//
// Ein CSV aus Bestand A trägt je Zeile Fassung, System und Kennung des Drucks
// samt Rückfall — und **nie** eine lokale Zeilennummer. In Bestand B trifft
// sie denselben fachlichen Druck: über die Kennung, sonst über den
// vollständigen Rückfall; trifft sie mehrere, ist das ein Konflikt und kein
// Automerge; trifft sie keinen, entsteht der Druck **mit** der Kennung, damit
// die nächste Einfuhr ihn findet.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const csvkopf = "printing,menge,name,setcode,nummer,sprache,verweisfassung,kennungssystem,kennung\n"

func TestDieReferenzTrifftDenselbenDruckInEinemAndersNummeriertenBestand(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	// **Bestand B zählt anders**: zwei fremde Drucke zuerst, dann der gesuchte.
	scryfallDruck(t, register, "11111111-1111-4111-8111-111111111111", "Giant Growth", "lea", "196")
	scryfallDruck(t, register, "22222222-2222-4222-8222-222222222222", "Counterspell", "lea", "54")
	erwartet := scryfallDruck(t, register, probeprinting, "Lightning Bolt", "m11", "149")

	ablage := &sammlungsdoppel{}
	bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Einfuegen(t.Context(),
		csvkopf+probeprinting+",4,Lightning Bolt,m11,149,en,1,scryfall,"+probeprinting+"\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 1 || len(bericht.Befunde) != 0 {
		t.Fatalf("Bericht %+v", bericht)
	}

	if ablage.zeilen[0].DruckID != erwartet || ablage.zeilen[0].Printing != probeprinting {
		t.Errorf("die Zeile zeigt auf %d mit Alias %q, erwartet %d", ablage.zeilen[0].DruckID, ablage.zeilen[0].Printing, erwartet)
	}
}

func TestEineUnbekannteReferenzLegtDenDruckMitSeinerKennungAn(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	ablage := &sammlungsdoppel{}
	einfuhr := service.NeueSammlungsausfuhr(ablage).MitDrucken(register)

	// **Eine MTGJSON-Kennung, die dieser Bestand nicht kennt** — ohne Alias.
	zeile := ",2,Lightning Bolt,m11,149,en,1,mtgjson,aaaaaaaa-0000-4000-8000-000000000001\n"

	bericht, err := einfuhr.Einfuegen(t.Context(), csvkopf+zeile, "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 1 || len(register.drucke) != 1 {
		t.Fatalf("Bericht %+v, %d Drucke", bericht, len(register.drucke))
	}

	druck := register.drucke[0]
	if druck.Name != "Lightning Bolt" || druck.Setcode != "m11" || druck.Sprache != "en" ||
		len(druck.Kennungen) != 1 || druck.Kennungen[0].System != domain.SystemMTGJSON {
		t.Errorf("der neue Druck: %+v", druck)
	}

	if ablage.zeilen[0].DruckID != druck.ID || !ablage.zeilen[0].Vorlaeufig() {
		t.Errorf("die Zeile ist nicht vorläufig auf den neuen Druck: %+v", ablage.zeilen[0])
	}

	// **Dieselbe Zeile noch einmal ist ein höherer Stapel**, kein zweiter
	// Druck — die Kennung ist jetzt bekannt, und der Vergleich geht über die
	// Zeile des Drucks, nicht über den leeren Alias.
	zweiter, err := einfuhr.Einfuegen(t.Context(), csvkopf+zeile, "")
	if err != nil {
		t.Fatalf("zweites Einfuegen: %v", err)
	}

	if zweiter.Erhoeht != 1 || len(zweiter.Angelegt) != 0 || len(register.drucke) != 1 || ablage.zeilen[0].Menge != 4 {
		t.Errorf("zweiter Bericht %+v, %d Drucke, Menge %d", zweiter, len(register.drucke), ablage.zeilen[0].Menge)
	}
}

func TestDerRueckfallTrifftOhneKennung(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	erwartet := scryfallDruck(t, register, probeprinting, "Lightning Bolt", "m11", "149")
	ablage := &sammlungsdoppel{}

	// **Kein Alias, keine Kennung — nur Name, Setcode, Nummer, Sprache.**
	bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Einfuegen(t.Context(),
		csvkopf+",1,lightning bolt,M11,149,en,,,\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 1 {
		t.Fatalf("Bericht %+v", bericht)
	}

	// **Und der Alias kommt vom getroffenen Druck.**
	if ablage.zeilen[0].DruckID != erwartet || ablage.zeilen[0].Printing != probeprinting {
		t.Errorf("Zeile %+v", ablage.zeilen[0])
	}
}

func TestEineMehrdeutigeReferenzIstEinKonfliktUndKeinAutomerge(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	// **Zweimal derselbe Rückfall**: zwei Drucke, die sich nur in der Kennung unterscheiden.
	scryfallDruck(t, register, "11111111-1111-4111-8111-111111111111", "Lightning Bolt", "m11", "149")
	scryfallDruck(t, register, "22222222-2222-4222-8222-222222222222", "Lightning Bolt", "m11", "149")
	ablage := &sammlungsdoppel{}

	bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Einfuegen(t.Context(),
		csvkopf+",1,Lightning Bolt,m11,149,en,,,\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 0 || len(bericht.Befunde) != 1 || len(register.drucke) != 2 {
		t.Fatalf("Bericht %+v, %d Drucke", bericht, len(register.drucke))
	}

	if bericht.Befunde[0].Ort != "Zeile 2" || !strings.Contains(bericht.Befunde[0].Meldung, "2 Drucke") {
		t.Errorf("Befund %+v", bericht.Befunde[0])
	}
}

func TestEineNeuereReferenzfassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(&druckregisterdoppel{}).Einfuegen(t.Context(),
		csvkopf+probeprinting+",1,Lightning Bolt,m11,149,en,2,scryfall,"+probeprinting+"\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 0 || len(bericht.Befunde) != 1 ||
		!strings.Contains(bericht.Befunde[0].Meldung, "Fassung 2") ||
		!strings.Contains(bericht.Befunde[0].Meldung, "1") {
		t.Errorf("Bericht %+v", bericht)
	}
}

func TestOhneReferenzUndOhneAliasBleibtEsEinEingabebefund(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(&druckregisterdoppel{}).Einfuegen(t.Context(),
		csvkopf+",1,Lightning Bolt,,,en,,,\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 0 || len(bericht.Befunde) != 1 {
		t.Errorf("Bericht %+v", bericht)
	}
}

// **Die Ausfuhr trägt die Referenz** — und ohne Register bleibt alles wie vor 0.5.0.
func TestDieAusfuhrTraegtDieReferenzUndOhneRegisterNurDenAlias(t *testing.T) {
	t.Parallel()

	text, err := service.NeueSammlungsausfuhr(gefuellt(t)).Ausgeben(t.Context(), service.FormCSV)
	if err != nil {
		t.Fatalf("Ausgeben: %v", err)
	}

	if !strings.Contains(text, ",1,scryfall,"+probeprinting) {
		t.Errorf("die Referenz fehlt: %q", text)
	}

	leer := &sammlungsdoppel{}
	bericht, err := service.NeueSammlungsausfuhr(leer).Einfuegen(t.Context(), text, "")
	if err != nil {
		t.Fatalf("Einfuegen ohne Register: %v", err)
	}

	if len(bericht.Angelegt) != 2 || leer.zeilen[0].DruckID != 0 {
		t.Errorf("ohne Register: %+v, Zeile %+v", bericht, leer.zeilen[0])
	}
}

// **Ein Register, das klemmt, ist ein Systemfehler** — kein Befund je Zeile.
func TestEinKlemmendesRegisterBrichtDieEinfuhrAb(t *testing.T) {
	t.Parallel()

	for _, weg := range []string{"NachKennung", "Kandidaten", "Sichern", "Fuer"} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			register := &druckregisterdoppel{fehlerBei: weg}
			if weg == "Fuer" {
				register.fehlerBei = ""
				scryfallDruck(t, register, probeprinting, "Lightning Bolt", "m11", "149")
				register.fehlerBei = weg
			}

			_, err := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).MitDrucken(register).Einfuegen(t.Context(),
				csvkopf+",1,Lightning Bolt,m11,149,en,1,scryfall,"+probeprinting+"\n", "")
			if !errors.Is(err, errRegister) {
				t.Errorf("Fehler %v, erwartet das Register", err)
			}
		})
	}
}

// **Ein Druck ohne Alias nennt in der Ausfuhr seine Kennung aus dem Register.**
func TestDieAusfuhrHoltDieKennungEinesVorlaeufigenDrucks(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	druck := &domain.Druck{
		Name: "Lightning Bolt", Setcode: "m11", Nummer: "149", Sprache: "en",
		Kennungen: []domain.Druckkennung{{System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: "aaaaaaaa-0000-4000-8000-000000000001"}},
	}
	if err := register.Sichern(t.Context(), druck); err != nil {
		t.Fatalf("Sichern: %v", err)
	}

	ablage := &sammlungsdoppel{}
	if _, err := ablage.Anlegen(t.Context(), &domain.Exemplar{DruckID: druck.ID, Menge: 1, Name: "Lightning Bolt"}); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	text, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Ausgeben(t.Context(), service.FormCSV)
	if err != nil {
		t.Fatalf("Ausgeben: %v", err)
	}

	if !strings.Contains(text, ",1,mtgjson,aaaaaaaa-0000-4000-8000-000000000001") {
		t.Errorf("die Kennung fehlt: %q", text)
	}

	register.fehlerBei = "Fuer"

	if _, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Ausgeben(t.Context(), service.FormCSV); !errors.Is(err, errRegister) {
		t.Errorf("Fehler %v, erwartet das Register", err)
	}
}

// **Ein CSV von vor 0.5.0 — nur der Alias — löst über die Scryfall-Kennung auf.**
func TestEinAltesCSVLoestUeberDenAliasAuf(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	erwartet := scryfallDruck(t, register, probeprinting, "Lightning Bolt", "m11", "149")
	ablage := &sammlungsdoppel{}

	bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Einfuegen(t.Context(),
		"printing,menge\n"+probeprinting+",3\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 1 || ablage.zeilen[0].DruckID != erwartet {
		t.Errorf("Bericht %+v, Zeile %+v", bericht, ablage.zeilen[0])
	}
}

// **Eine Zeile ohne Kennung und ohne Namen ist ein Befund des Exemplars**, nicht der Referenz.
func TestEineLeereReferenzIstEinExemplarbefund(t *testing.T) {
	t.Parallel()

	bericht, err := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).MitDrucken(&druckregisterdoppel{}).
		Einfuegen(t.Context(), csvkopf+",1,,,,en,,,\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 0 || len(bericht.Befunde) != 1 ||
		!strings.Contains(bericht.Befunde[0].Meldung, "Druck") {
		t.Errorf("Bericht %+v", bericht)
	}
}

// **Eine ausdrückliche Referenz, die keine ist, wird zum Befund der Zeile** —
// nicht zum Abbruch der Einfuhr.
func TestEineUnbrauchbareReferenzIstEinBefundDerZeile(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	bericht, err := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).MitDrucken(register).Einfuegen(t.Context(),
		csvkopf+",1,Lightning Bolt,m11,149,en,1,scryfall,abc\n", "")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(bericht.Angelegt) != 0 || len(bericht.Befunde) != 1 || len(register.drucke) != 0 ||
		bericht.Befunde[0].Ort != "Zeile 2" {
		t.Errorf("Bericht %+v", bericht)
	}
}

// **Ein bloßer Alias bleibt nachsichtig** — wie vor 0.5.0: die Zeile geht
// durch, und die Ablage hängt beim Schreiben den Druck an. Das gilt auch für
// die eigene Ausfuhr, deren Referenz den Alias nur wiederholt.
func TestEinUnbekannterAliasGehtOhneRegisterzeileDurch(t *testing.T) {
	t.Parallel()

	for name, text := range map[string]string{
		"nur der Alias":                "printing,menge\nabc,2\n",
		"die eigene Ausfuhr des Alias": csvkopf + "abc,2,,,,en,1,scryfall,abc\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			register := &druckregisterdoppel{}
			ablage := &sammlungsdoppel{}
			bericht, err := service.NeueSammlungsausfuhr(ablage).MitDrucken(register).Einfuegen(t.Context(), text, "")
			if err != nil {
				t.Fatalf("Einfuegen: %v", err)
			}

			if len(bericht.Angelegt) != 1 || len(register.drucke) != 0 || ablage.zeilen[0].Printing != "abc" ||
				ablage.zeilen[0].DruckID != 0 {
				t.Errorf("Bericht %+v, Zeile %+v, %d Drucke", bericht, ablage.zeilen[0], len(register.drucke))
			}
		})
	}
}
