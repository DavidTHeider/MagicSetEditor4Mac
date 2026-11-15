package domain_test

// **Wer was sieht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die zweite Abnahme des Werkstücks
//
// „Fremde Hand und Bibliotheksreihenfolge fehlen in serialisierten DTOs,
// Events, Fehlern, Chat-/Diagnosekontext und Assetmanifest eines
// Gegners/Zuschauers; Hostprivileg wird als dokumentierte Ausnahme geprüft."
//
// **Deshalb steht die Projektion in `domain` und nicht im Dienst**: sie ist
// die Regel selbst. Ein Dienst, der sie rufen **müsste**, wäre ein Dienst,
// der sie vergessen kann.

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineTischlage baut eine Partie mit Objekten in jeder Zone.
func eineTischlage() domain.Partie {
	partie := einePartie()
	partie.ID = "p1"
	partie.Stand = domain.StandLaeuft
	partie.Zuschauer = []domain.Zuschauer{{Kennung: "z1", Name: "Nachbar"}}

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "h1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneHand, Karte: "wald",
		},
		{
			Kennung: "h2", Art: domain.ObjektKarte, Beherrscher: "sitz-2",
			Eigner: "sitz-2", Zone: domain.ZoneHand, Karte: "gebirge",
		},
		{
			Kennung: "b1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneBibliothek, Karte: "insel",
			Ordnung: 1,
		},
		{
			Kennung: "b2", Art: domain.ObjektKarte, Beherrscher: "sitz-2",
			Eigner: "sitz-2", Zone: domain.ZoneBibliothek, Karte: "sumpf",
			Ordnung: 2,
		},
		{
			Kennung: "t1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: "baer",
		},
	}

	return partie
}

// sichtobjekte gibt die Objekte einer Sicht nach Kennung — zum Vergleichen.
//
// **Zurückgerechnet über die Partie** (#2480): die Sicht nennt Sichtkennungen,
// und gesucht wird hier nach der Kennung, die der Lauf angelegt hat.
func sichtobjekte(
	partie domain.Partie, sicht domain.Partiesicht,
) map[string]domain.Sichtobjekt {
	kennungen := make(map[string]string, len(partie.Objekte))
	for _, objekt := range partie.Objekte {
		kennungen[partie.Sichtkennung(objekt)] = objekt.Kennung
	}

	gefunden := make(map[string]domain.Sichtobjekt, len(sicht.Objekte))

	for _, objekt := range sicht.Objekte {
		gefunden[kennungen[objekt.Kennung]] = objekt
	}

	return gefunden
}

// **Ein Spieler sieht seine Hand und nicht die fremde.**
func TestEinSpielerSiehtSeineHandUndNichtDieFremde(t *testing.T) {
	t.Parallel()

	sicht := eineTischlage().Sicht("sitz-1")

	gesehen := sichtobjekte(eineTischlage(), sicht)

	if eigene, da := gesehen["h1"]; !da || eigene.Karte != "wald" {
		t.Errorf("die eigene Hand fehlt: %+v", eigene)
	}

	fremde, da := gesehen["h2"]
	if !da {
		t.Fatal("die fremde Handkarte fehlt ganz — dann wüsste niemand, " +
			"dass das Gegenüber eine Hand hat")
	}

	if fremde.Karte != "" || fremde.Name != "" {
		t.Errorf("die fremde Handkarte ist lesbar: %+v", fremde)
	}

	if !fremde.Geheim {
		t.Error("die fremde Handkarte ist nicht als geheim ausgewiesen")
	}
}

// **Die Bibliothek sieht niemand — auch ihr eigener Beherrscher nicht.**
//
// **Das ist keine Härte gegen den Spieler**, sondern die Bedingung dafür,
// dass eine Bibliothek eine ist: wer seine eigene läse, hätte sie sortiert.
func TestDieBibliothekSiehtNiemand(t *testing.T) {
	t.Parallel()

	partie := eineTischlage()

	for _, wer := range []string{"sitz-1", "sitz-2", "z1"} {
		t.Run(wer, func(t *testing.T) {
			t.Parallel()

			gesehen := sichtobjekte(partie, partie.Sicht(wer))

			for _, kennung := range []string{"b1", "b2"} {
				karte, da := gesehen[kennung]
				if !da {
					t.Fatalf("%s fehlt ganz — dann wäre die Bibliothek leer",
						kennung)
				}

				if karte.Karte != "" {
					t.Errorf("%s ist für %s lesbar", kennung, wer)
				}

				// **Und die Reihenfolge geht nicht mit** — sie ist das
				// eigentliche Geheimnis einer Bibliothek.
				if karte.Ordnung != 0 {
					t.Errorf("%s trägt die Ordnung %d", kennung, karte.Ordnung)
				}
			}
		})
	}
}

// **Ein Zuschauer sieht keine einzige Hand.**
func TestEinZuschauerSiehtKeineHand(t *testing.T) {
	t.Parallel()

	gesehen := sichtobjekte(eineTischlage(), eineTischlage().Sicht("z1"))

	for _, kennung := range []string{"h1", "h2"} {
		if karte := gesehen[kennung]; karte.Karte != "" {
			t.Errorf("der Zuschauer liest %s: %+v", kennung, karte)
		}
	}

	// **Den Tisch sieht er** — sonst wäre Zusehen keines.
	if tisch := gesehen["t1"]; tisch.Karte != "baer" {
		t.Errorf("der Tisch fehlt: %+v", tisch)
	}
}

// **Wer gar nicht dabei ist, sieht so viel wie ein Zuschauer.**
//
// **Der sichere Ausgang**: eine unbekannte Kennung bekommt keine Sonderrechte
// — sie bekommt die engste Sicht, die es gibt.
func TestEineUnbekannteKennungSiehtWieEinZuschauer(t *testing.T) {
	t.Parallel()

	partie := eineTischlage()

	fremd := sichtobjekte(partie, partie.Sicht("wer-auch-immer"))
	zuschauer := sichtobjekte(partie, partie.Sicht("z1"))

	for kennung, objekt := range fremd {
		if !reflect.DeepEqual(objekt, zuschauer[kennung]) {
			t.Errorf("%s sieht für einen Fremden anders aus: %+v gegen %+v",
				kennung, objekt, zuschauer[kennung])
		}
	}

	// **Und die leere Kennung ebenso.**
	if len(sichtobjekte(partie, partie.Sicht(""))) != len(zuschauer) {
		t.Error("die leere Kennung sieht etwas anderes")
	}
}

// **Der Wirt sieht nicht mehr als jeder andere Spieler.**
//
// **Das ist das Hostprivileg als dokumentierte Ausnahme** (#1458): technisch
// **kennt** er den Vollzustand — er hält ihn —, aber seine **Projektion** ist
// dieselbe wie die jedes anderen Sitzes. Wer das umdrehte, hätte den
// Vollzustand zum Normalfall gemacht.
func TestDerWirtSiehtNichtMehrAlsJederAndereSitz(t *testing.T) {
	t.Parallel()

	partie := eineTischlage()

	if !partie.IstWirt("sitz-1") {
		t.Fatal("die Vorrichtung meint einen anderen Wirt")
	}

	alsWirt := sichtobjekte(partie, partie.Sicht("sitz-1"))

	if fremde := alsWirt["h2"]; fremde.Karte != "" {
		t.Errorf("der Wirt liest die fremde Hand: %+v", fremde)
	}

	// **Der Vollzustand ist ein eigener Weg und heißt auch so.**
	voll := partie.Vollzustand()

	if len(voll.Objekte) != len(partie.Objekte) {
		t.Errorf("der Vollzustand hat %d von %d Objekten",
			len(voll.Objekte), len(partie.Objekte))
	}

	for _, objekt := range voll.Objekte {
		if objekt.Kennung == "h2" && objekt.Karte != "gebirge" {
			t.Error("der Vollzustand verdeckt etwas — dann hielte der Wirt " +
				"nicht, was er halten soll")
		}
	}
}

// **Ein Ereignis trägt seine private Ergänzung nur zu ihrem Empfänger.**
func TestEinEreignisTraegtSeineErgaenzungNurZuIhremEmpfaenger(t *testing.T) {
	t.Parallel()

	ereignis := domain.Ereignis{
		PartieID: "p1", Folge: 7, Art: "karte.gezogen",
		Oeffentlich: map[string]any{"sitz": "sitz-1"},
		Privat: []domain.Ergaenzung{
			{Empfaenger: "sitz-1", Inhalt: map[string]any{"karte": "wald"}},
			{Empfaenger: "sitz-2", Inhalt: map[string]any{"karte": "insel"}},
		},
	}

	fuerEinen := ereignis.Fuer("sitz-1")

	if fuerEinen.Oeffentlich["sitz"] != "sitz-1" {
		t.Error("der öffentliche Teil fehlt")
	}

	if len(fuerEinen.Privat) != 1 ||
		fuerEinen.Privat[0].Inhalt["karte"] != "wald" {
		t.Errorf("die eigene Ergänzung stimmt nicht: %+v", fuerEinen.Privat)
	}

	// **Ein Zuschauer bekommt gar keine** — und der öffentliche Teil bleibt.
	fuerZuschauer := ereignis.Fuer("z1")

	if len(fuerZuschauer.Privat) != 0 {
		t.Errorf("der Zuschauer bekam %d Ergänzungen", len(fuerZuschauer.Privat))
	}

	if fuerZuschauer.Oeffentlich["sitz"] != "sitz-1" {
		t.Error("dem Zuschauer fehlt der öffentliche Teil")
	}

	// **Und die Kette bleibt dieselbe** — sonst könnte niemand prüfen, was
	// er bekommen hat.
	if fuerZuschauer.Kette != ereignis.Kette || fuerEinen.Folge != 7 {
		t.Error("die Projektion verändert den Strom")
	}
}

// **Eine verdeckt liegende Karte ist auch auf dem Tisch geheim.**
//
// # Der Fall, den die Zonentabelle allein nicht fängt
//
// **Der Tisch ist eine offene Zone** — und eine verdeckt gespielte Kreatur
// liegt trotzdem verdeckt. Wer nur die Zone fragte, schriebe ihren Namen in
// die Antwort jedes Gegenübers, und niemand sähe, warum das Spiel nicht mehr
// funktioniert.
//
// **Zwei verschiedene Tatsachen, zwei Felder**: `Verdeckt` heißt „sie liegt
// verdeckt" und gilt für alle; `Geheim` heißt „du darfst nicht hineinsehen"
// und gilt je Empfänger.
func TestEineVerdeckteKarteIstAuchAufDemTischGeheim(t *testing.T) {
	t.Parallel()

	partie := eineTischlage()
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "v1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
		Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: "morph",
		Verdeckt: true,
	})

	// **Sein Beherrscher weiß, was es ist.**
	seines := sichtobjekte(partie, partie.Sicht("sitz-1"))["v1"]

	if seines.Karte != "morph" {
		t.Errorf("der eigene Morph ist dem Beherrscher geheim: %+v", seines)
	}

	if !seines.Verdeckt || seines.Geheim {
		t.Errorf("die zwei Tatsachen stimmen nicht: %+v", seines)
	}

	// **Jeder andere sieht nur, dass dort etwas liegt.**
	for _, wer := range []string{"sitz-2", "z1", "", "fremd"} {
		t.Run(wer, func(t *testing.T) {
			t.Parallel()

			fremdes := sichtobjekte(partie, partie.Sicht(wer))["v1"]

			if fremdes.Karte != "" || fremdes.Name != "" {
				t.Errorf("%q liest den Morph: %+v", wer, fremdes)
			}

			if !fremdes.Verdeckt || !fremdes.Geheim {
				t.Errorf("%q bekommt die falschen Marken: %+v", wer, fremdes)
			}
		})
	}

	// **Und im Vollzustand steht er** — der Wirt hält ihn.
	for _, objekt := range partie.Vollzustand().Objekte {
		if objekt.Kennung == "v1" && objekt.Karte != "morph" {
			t.Error("der Vollzustand verdeckt den Morph")
		}
	}
}

// **Die Sitze einer Sicht nennen den Abdruck und die Zahl, nicht das Deck.**
func TestEineSichtNenntDenDeckabdruckUndNichtDasDeck(t *testing.T) {
	t.Parallel()

	sicht := eineTischlage().Sicht("sitz-1")

	if len(sicht.Sitze) != 2 {
		t.Fatalf("die Sicht führt %d Sitze", len(sicht.Sitze))
	}

	for _, sitz := range sicht.Sitze {
		if sitz.Deckabdruck == "" || sitz.Karten != 40 {
			t.Errorf("der Sitz %q: Abdruck %q, %d Karten",
				sitz.Kennung, sitz.Deckabdruck, sitz.Karten)
		}
	}

	// **Und die Sicht sagt, für wen sie gerechnet wurde** — damit ein
	// Empfänger merkt, wenn er die eines anderen bekommen hat.
	if sicht.Fuer != "sitz-1" || sicht.Wirt != "sitz-1" {
		t.Errorf("die Sicht ist für %q, Wirt ist %q", sicht.Fuer, sicht.Wirt)
	}

	// **Eine Partie ohne Zuschauer gibt eine leere Liste und kein `nil`** —
	// sonst steht im Fenster `zuschauer: null`.
	ohne := einePartie()

	if gerechnet := ohne.Sicht("sitz-1"); gerechnet.Zuschauer == nil {
		t.Error("die Zuschauerliste ist nil")
	}
}
