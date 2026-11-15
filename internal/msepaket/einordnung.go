package msepaket

import (
	"strconv"
	"strings"
)

// RangOhneHinweis ist der Rang eines Pakets, das keinen `position hint:` nennt.
//
//	TreeItem() : position_type(TYPE_OTHER), position_hint(1000000) {}
//
// `src/gui/package_update_list.hpp`, Zeile 52, Fassung f9d9356 vom 2021-07-06.
// **Eine Million und nicht null**: ein Paket ohne Hinweis steht *hinter* jedem
// mit Hinweis. Die Zahl ist genau die von MSE, damit ein Paket, das 999.999
// schreibt, hier dieselbe Nachbarschaft hat wie dort.
const RangOhneHinweis = 1000000

// einordnungLesen füllt `Gruppe` und `Rang` aus dem Kopf.
//
// **Zwei Angaben, die dieses Projekt bis #244 übersprungen hat.** Sie sind der
// Grund, warum MSE2 eine Paketverwaltung zeigt und dieses Programm eine Liste:
// die Gruppierung steht in den Daten und muss nicht erfunden werden.
func (p *Paket) einordnungLesen(baum *Baum) {
	// **Gesäubert, bevor er zerlegt wird.** Vier Stile im `full-magic-pack`
	// schreiben ihre Gruppe mit führendem Tabulator; roh zerlegt entstünde ein
	// Knoten „⇥magic" neben „magic", und beide sähen in der Liste gleich aus.
	p.Gruppe = strings.TrimSpace(wert(baum.Ersten("installer group")))
	p.Rang = p.rangOderBefund(baum.Ersten("position hint"))
}

// rangOderBefund liest den Rang; eine unlesbare Angabe wird gemeldet.
//
// **Basis 10 ausdrücklich.** Die Packs schreiben dreistellig mit führender
// Null (`095`, `071`), damit die Zahlen im Texteditor untereinander stehen.
// Mit Basis 0 wäre `095` ein Fehler und `071` die 57 — beides still falsch.
func (p *Paket) rangOderBefund(knoten *Knoten) int {
	if knoten == nil || strings.TrimSpace(knoten.Wert) == "" {
		return RangOhneHinweis
	}

	rang, err := strconv.ParseInt(strings.TrimSpace(knoten.Wert), 10, 32)
	if err != nil {
		p.melde(knoten.Stelle, "`position hint` ist keine Zahl: "+knoten.Wert)

		return RangOhneHinweis
	}

	return int(rang)
}
