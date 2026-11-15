package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Paket sagt selbst, wohin es in einer Liste gehört.**
//
//	installer group: magic/m15 style/normal cards
//	position hint: 095
//
// `magic-m15-punk-show.mse-style/style`, Zeilen 6–7 — und damit sind es zwei
// Angaben, die dieses Projekt bislang übersprungen hat. MSE2 baut daraus
// seinen Paketbaum:
//
//	String group = ip->description->installer_group;
//	if (group.empty()) group = _("custom");
//	root.add(ip, group);
//
// `src/gui/package_update_list.cpp`, Zeilen 204–207, Fassung f9d9356 vom
// 2021-07-06. Der Pfad wird an `/` zerlegt, jede Stufe wird ein Knoten
// (Zeilen 18–59), und `position hint` ordnet innerhalb einer Stufe
// (`compare_pos_hint`, Zeilen 61–67).
//
// **Über `full-magic-pack` tragen 416 von 526 Paketen eine `installer group`
// und 495 einen `position hint`** — gezählt am 2026-08-19 über die
// Hauptdateien aller sieben Arten. Die Reichweite dieser Zahl ist genau dieses
// eine Pack (§10); sie sagt, dass die Angabe der Normalfall ist, nicht dass
// sie immer dasteht.
func TestEinPaketNenntSeineEinordnung(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
short name: Probe
installer group: magic/m15 style/normal cards
position hint: 095
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer schliesse(t, paket)

	if paket.Gruppe != "magic/m15 style/normal cards" {
		t.Errorf("Gruppe = %q", paket.Gruppe)
	}

	// **`095` ist keine Oktalzahl, sondern die Zahl 95.** Die Packs schreiben
	// ihre Ränge dreistellig mit führender Null, damit sie in einem Texteditor
	// untereinander stehen; wer sie mit Basis 0 läse, bekäme für `095` einen
	// Fehler und für `071` die 57.
	if paket.Rang != 95 {
		t.Errorf("Rang = %d", paket.Rang)
	}
}

// **Fehlt die Angabe, ist der Rang der größte.**
//
//	TreeItem() : position_type(TYPE_OTHER), position_hint(1000000) {}
//
// `src/gui/package_update_list.hpp`, Zeile 52, Fassung f9d9356 vom
// 2021-07-06 — ein Paket ohne Hinweis steht
// hinter jedem mit Hinweis, nicht vor ihm. Eine 0 als Vorgabewert kehrte die
// Reihenfolge um, und das fiele erst an einer Liste auf, die schon falsch ist.
func TestOhneEinordnungStehtDasPaketHinten(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, "mse version: 2.0.0\nshort name: Probe\n")

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer schliesse(t, paket)

	if paket.Gruppe != "" {
		t.Errorf("Gruppe = %q", paket.Gruppe)
	}

	if paket.Rang != msepaket.RangOhneHinweis {
		t.Errorf("Rang = %d, erwartet %d", paket.Rang, msepaket.RangOhneHinweis)
	}
}

// **Eine unlesbare Zahl ist ein Befund, kein stiller Vorgabewert.**
//
// Dieselbe Regel wie bei `version:` (§8): fremde Daten sind der Normalfall,
// und wer eine kaputte Angabe stillschweigend durch 1.000.000 ersetzt, gibt
// dem Paket eine Einordnung, die niemand geschrieben hat.
func TestEinUnlesbarerRangIstEinBefund(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
short name: Probe
position hint: bald
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer schliesse(t, paket)

	if paket.Rang != msepaket.RangOhneHinweis {
		t.Errorf("Rang = %d", paket.Rang)
	}

	if len(paket.Befunde) != 1 {
		t.Fatalf("%d Befunde: %+v", len(paket.Befunde), paket.Befunde)
	}
}

// **Der Pfad wird gesäubert, bevor er zerlegt wird.**
//
// Vier Stile im `full-magic-pack` schreiben ihre Gruppe mit einem führenden
// Tabulator — `\tmagic/m15 style/normal cards`. Wer das roh zerlegt, bekommt
// einen Knoten namens „⇥magic" neben „magic", und beide sehen in der Liste
// gleich aus.
func TestEineGruppeWirdGesaeubert(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, "mse version: 2.0.0\nshort name: Probe\ninstaller group: \tmagic/m15 style \n")

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer schliesse(t, paket)

	if paket.Gruppe != "magic/m15 style" {
		t.Errorf("Gruppe = %q", paket.Gruppe)
	}
}

// schliesse gibt das Paket frei und meldet, wenn das misslingt.
//
// **Und wirft den Fehler nicht weg.** Ein `_ =` wäre hier kürzer und verlöre
// genau den Fall, den `Paket.Schliessen` überhaupt hat: ein Archiv, das sich
// nicht schließen lässt.
func schliesse(t *testing.T, paket *msepaket.Paket) {
	t.Helper()

	if fehler := paket.Schliessen(); fehler != nil {
		t.Errorf("schließen: %v", fehler)
	}
}
