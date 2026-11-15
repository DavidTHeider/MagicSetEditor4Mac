// **Richtlinien: was gilt, ist die Schnittmenge** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Warum sie keine gewöhnliche Vorliebe sind
//
// **Weil eine Vorliebe die andere ersticht und eine Schranke sie addiert.**
// Bei „welche Schriftgröße" hat der Mensch eine Vorgabe und das Profil das
// letzte Wort — das ist [Verschmolzen], und es ist richtig: es geht um
// Geschmack. Bei „darf dieses Programm ins Netz" wäre dieselbe Regel ein
// Loch: wer als Mensch „nein" gesagt hat, sagt es für **jedes** seiner
// Profile, und ein Profil, das „doch" sagt, hebt eine Entscheidung auf, die
// nicht ihm gehört.
//
// **Deshalb wird hier nicht gestochen, sondern gefaltet**: der strengste Wert
// aller geltenden Ebenen gewinnt, und die Reihenfolge der Ebenen spielt keine
// Rolle. Das ist die Aussage aus dem Issue — „die wirksame Einschränkung ist
// die Schnittmenge aller geltenden Ebenen".
//
// # Und warum die Regel hier steht und nicht im Fenster
//
// **Weil vier Flächen dieselbe Antwort geben müssen** (§16). Stünde sie in
// React, hätte der MCP-Server keine — und dort sitzt gerade das, wovor ein
// „nur lokal" schützen soll.

package domain

import (
	"slices"
	"strconv"
	"strings"
)

// Die Schlüssel der übergreifenden Schranken.
const (
	// RichtlinieNurLokal schaltet jeden fremden Dienst ab.
	//
	// **Ein eigener Schlüssel und nicht vier Schalter auf einmal**: „nur
	// lokal" ist eine Aussage, die stehen bleibt, wenn ein fünfter Dienst
	// dazukommt. Wer stattdessen die vier Schalter umlegte, hätte den fünften
	// an dem Tag offen, an dem er entsteht.
	RichtlinieNurLokal = "richtlinie.nurlokal"

	// RichtlinieKostengrenze ist die Obergrenze in Cent — leer heißt: keine.
	//
	// **Die Form steht hier, die Wirkung kommt mit Phase 5.3.** Eine Grenze,
	// die erst entsteht, wenn der erste Auftrag schon läuft, ist keine
	// Grenze; dass sie heute niemand liest, steht in ihrem Kommentar und
	// nicht zwischen den Zeilen (§10).
	RichtlinieKostengrenze = "richtlinie.kostengrenze"
)

// Die zwei Stellungen von „nur lokal".
//
// **Aus ist die Vorgabe**, weil es der Zustand von gestern ist: eine
// Einstellung, die beim Einbau etwas abschaltet, schaltet es für jeden ab, der
// nie danach gefragt wurde.
const (
	NurLokalAus = "aus"
	NurLokalAn  = "an"
)

// KostengrenzeHoechstens ist die höchste Kostengrenze, die sich eintragen
// lässt — in Cent.
//
// **Zehn Millionen Cent sind hunderttausend Euro.** Die Zahl ist keine Aussage
// über Preise, sondern über Tippfehler: eine Grenze mit vierzehn Ziffern ist
// keine Grenze, und eine Zahl, die nicht mehr in eine `int64`-Rechnung passt,
// wäre eine Schranke, die überläuft. Wer wirklich mehr braucht, hat ein
// anderes Gespräch zu führen als dieses.
const KostengrenzeHoechstens = 10_000_000

// Woher ein wirksamer Wert kommt.
//
// **Drei und nicht vier**: `arbeitsplatz` ist eine Ebene der Schlüssel, keine
// der Auflösung — ein Layoutwert wird im Profil gewählt wie jeder andere.
//
// **`Geltung…` und nicht `Herkunft…`**, weil `Herkunftsart` in diesem Paket
// schon vergeben ist: sie sagt, aus wessen **Bestand** ein Deck kam
// ([0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md)). Zwei
// gleichnamige Begriffe für zwei Fragen wären die Verwechslung, die man
// später sucht.
const (
	GeltungProfil  = "profil"
	GeltungNutzer  = "nutzer"
	GeltungVorgabe = "vorgabe"
)

// richtlinienart sagt, wie zwei Werte zu vergleichen sind.
type richtlinienart int

const (
	// reihung: aufzählbare Werte, in [richtlinie.Strenge] von streng nach
	// locker.
	reihung richtlinienart = iota

	// obergrenze: eine Zahl, die kleinere ist strenger; leer heißt „keine".
	obergrenze
)

// richtlinie ist eine Zeile der Tabelle darunter.
type richtlinie struct {
	Art richtlinienart

	// Strenge zählt die Werte von streng nach locker — nur bei [reihung].
	Strenge []string
}

// richtlinien ordnet jedem festen Schlüssel seine Art zu.
//
// **Die Dienstschalter stehen nicht darin**, weil es sie nicht als feste Liste
// gibt: sie entstehen aus `dienstzeilen`, und eine zweite Aufzählung hier
// liefe beim fünften Dienst auseinander — dieselbe Falle, die in
// `einstellung.go` schon zweimal zugeschnappt ist. Für sie gilt die Regel
// unten in [findeRichtlinie].
var richtlinien = map[string]richtlinie{
	RichtlinieNurLokal: {Art: reihung, Strenge: []string{NurLokalAn, NurLokalAus}},

	// **Der zweite Faktor beim Profilwechsel ist eine Schranke, keine
	// Vorliebe** — und deshalb kann ein Import ihn nicht lockern. `verlangen`
	// ist strenger als `nurAnmelden`; die Reihenfolge hier sagt es, und
	// [Lockert] liest sie.
	NutzerFaktorbeimWechsel: {
		Art:     reihung,
		Strenge: []string{FaktorVerlangen, FaktorNurAnmelden},
	},

	RichtlinieKostengrenze: {Art: obergrenze},
}

// IstRichtlinie sagt, ob für diesen Schlüssel die Schnittmenge gilt.
func IstRichtlinie(schluessel string) bool {
	_, ja := findeRichtlinie(schluessel)

	return ja
}

// findeRichtlinie gibt die Zeile eines Schlüssels.
func findeRichtlinie(schluessel string) (richtlinie, bool) {
	if zeile, kennt := richtlinien[schluessel]; kennt {
		return zeile, true
	}

	// **Jeder Dienst*schalter* ist eine Richtlinie.** Am Präfix und nicht an
	// einer Liste: welche Dienste es gibt, steht in `dienste.go`, und wer
	// dort einen aufnimmt, bekommt die Schranke, ohne sie zu bestellen.
	//
	// # Und **nur** der Schalter
	//
	// **Ein Schlüssel mit einem dritten Teil ist keiner** — `dienst.openai`
	// ist der Schalter, `dienst.openai.adresse` die Adresse dahinter
	// (#1540).
	//
	// **Bis zum 2026-09-13 fiel beides unter dieselbe Zeile**, und eine
	// eingestellte Adresse galt deshalb **nicht**: `http://127.0.0.1:8099`
	// steht in keiner Reihung `{aus, an}`, `verschaerftMitStelle` gab
	// `stelle < 0`, und [richtliniengeltung] setzte die (leere) Vorgabe als
	// wirksam. Die Naht baute den Klienten mit der Vorgabeadresse des
	// Anbieters — die Zusage von A18 war an **keiner** Fläche eingelöst.
	//
	// **Gefunden in der Sichtprüfung am laufenden Fenster**, nicht in einem
	// Lauf: die Go-Läufe setzen den Stand als fertige Karte, die
	// Frontendläufe belegen die Naht, und die Faltung dazwischen prüfte
	// niemand an einem freien Wert (§3).
	if istDienstschalter(schluessel) {
		return richtlinie{Art: reihung, Strenge: []string{DienstAus, DienstAn}}, true
	}

	return richtlinie{}, false
}

// Verschaerft gibt den strengsten der Werte — leere zählen nicht mit.
//
// **Leer heißt „hat nichts gesagt" und nicht „hat das Lockerste gewählt".**
// Eine Ebene, auf der niemand etwas eingetragen hat, darf keine Schranke
// aufheben; sonst hätte die Vorgabe eines Profils die Wahl eines Menschen
// erstickt, und genau das soll die Faltung verhindern.
//
// **Für einen Schlüssel, der keine Richtlinie ist, gilt der erste Wert.** Es
// gibt dann keine Ordnung, nach der zu vergleichen wäre; wer diese Funktion
// dort ruft, hat sich in der Frage geirrt und bekommt keine erfundene Antwort.
func Verschaerft(schluessel string, werte ...string) string {
	wert, _ := verschaerftMitStelle(schluessel, werte...)

	return wert
}

// verschaerftMitStelle gibt dazu, aus welchem der Werte er stammt — `-1`,
// wenn keiner etwas gesagt hat.
func verschaerftMitStelle(schluessel string, werte ...string) (wert string, stelle int) {
	zeile, ist := findeRichtlinie(schluessel)
	if !ist {
		for stelle, wert := range werte {
			if wert != "" {
				return wert, stelle
			}
		}

		return "", -1
	}

	if zeile.Art == obergrenze {
		return kleinsteGrenze(werte...)
	}

	return strengsterAusReihung(zeile.Strenge, werte...)
}

// strengsterAusReihung gibt den Wert mit dem kleinsten Rang.
//
// **Ein unbekannter Wert wird übergangen.** Er kommt nicht aus dieser
// Anwendung — `PruefeEinstellung` weist ihn ab —, und ihn als „locker" oder
// „streng" zu deuten hieße, über etwas zu urteilen, das niemand kennt.
func strengsterAusReihung(strenge []string, werte ...string) (wert string, stelle int) {
	besterRang := len(strenge)
	besterWert := ""
	besteStelle := -1

	for stelle, wert := range werte {
		rang := slices.Index(strenge, wert)
		if rang < 0 || rang >= besterRang {
			continue
		}

		besterRang, besterWert, besteStelle = rang, wert, stelle
	}

	return besterWert, besteStelle
}

// kleinsteGrenze gibt die niedrigste gesetzte Obergrenze.
func kleinsteGrenze(werte ...string) (wert string, stelle int) {
	kleinste := int64(-1)
	besterWert := ""
	besteStelle := -1

	for stelle, wert := range werte {
		betrag, fehler := strconv.ParseInt(wert, 10, 64)
		if fehler != nil {
			continue
		}

		if kleinste >= 0 && betrag >= kleinste {
			continue
		}

		kleinste, besterWert, besteStelle = betrag, wert, stelle
	}

	return besterWert, besteStelle
}

// Lockert sagt, ob `neu` eine bestehende Schranke aufweichen würde.
//
// **Die Frage des Imports** (§7 und Vermerk 0079): eine Datei darf eine
// Vorliebe mitbringen, aber keine Schranke abbauen, die auf einer Ebene steht,
// die sie gar nicht betrifft. Wer prüft, ob etwas übernommen werden darf,
// fragt hier — und bekommt ein Ja nur, wenn die Faltung den neuen Wert
// stehen lässt.
func Lockert(schluessel, neu, schranke string) bool {
	if !IstRichtlinie(schluessel) || schranke == "" {
		return false
	}

	return Verschaerft(schluessel, neu, schranke) != neu
}

// pruefeKostengrenze urteilt über eine Obergrenze.
func pruefeKostengrenze(wert string) error {
	if wert == "" {
		return nil
	}

	cent, fehler := strconv.ParseInt(wert, 10, 64)
	if fehler != nil {
		return NewFieldError("wert", SchluesselValidierung,
			"%q ist keine Kostengrenze — erwartet wird eine ganze Zahl in "+
				"Cent oder gar nichts.", wert)
	}

	if cent < 0 || cent > KostengrenzeHoechstens {
		return NewFieldError("wert", SchluesselValidierung,
			"%d Cent liegen außerhalb von 0 bis %d.", cent, KostengrenzeHoechstens)
	}

	return nil
}

// richtlinienschluessel gibt die zwei neuen — für die Ebenentabelle.
func richtlinienschluessel() []string {
	// **Die Schranken des Handels kommen dazu** (`T1`, #1466) — und zwar
	// aus **einer** Aufzählung: eine zweite Liste hier liefe beim nächsten
	// Schlüssel auseinander, und das ist dieselbe Falle, die in
	// `einstellung.go` schon zweimal zugeschnappt ist.
	return append([]string{RichtlinieNurLokal, RichtlinieKostengrenze},
		handelsschluessel()...)
}

// haengeRichtlinienAn hängt die zwei Schranken an beide Reihen.
//
// **In beide, weil die Faltung zwei Ebenen braucht.** Eine Schranke, die es
// nur im Profil gibt, ist keine: sie schützt genau das Profil, das sie
// aufheben kann.
//
// **Gerufen aus dem `init` von `einstellung.go`** — die Begründung steht bei
// [haengeArbeitsplatzAn].
func haengeRichtlinienAn() {
	for _, zeile := range []struct {
		schluessel string
		werte      []string
		vorgabe    string
		pruefung   func(string) error
	}{
		{RichtlinieNurLokal, []string{NurLokalAus, NurLokalAn}, NurLokalAus, nil},
		{RichtlinieKostengrenze, nil, "", pruefeKostengrenze},
	} {
		einstellungen = append(einstellungen, einstellung{
			Schluessel: zeile.schluessel,
			Wirkung:    WirkungSofort,
			Bereich:    BereichRichtlinien,
			Werte:      zeile.werte,
			Vorgabe:    zeile.vorgabe,
			Pruefung:   zeile.pruefung,
		})

		nutzereinstellungen = append(nutzereinstellungen, nutzereinstellung{
			Schluessel: zeile.schluessel,
			Werte:      zeile.werte,
			Vorgabe:    zeile.vorgabe,
			Pruefung:   zeile.pruefung,
		})
	}
}

// istDienstschalter sagt, ob ein Schlüssel **der Schalter** eines Dienstes ist.
//
// **Gezählt und nicht geraten**: `dienst.<name>` hat zwei Teile, alles
// darunter mehr. Eine Aufzählung der Unterschlüssel stünde hier als zweite
// Wahrheit — und der nächste, den jemand anlegt, fehlte darin.
func istDienstschalter(schluessel string) bool {
	if !strings.HasPrefix(schluessel, dienstpraefix) {
		return false
	}

	return !strings.Contains(schluessel[len(dienstpraefix):], ".")
}

// dienstpraefix führt jeden Schlüssel, der zu einem Dienst gehört.
const dienstpraefix = "dienst."
