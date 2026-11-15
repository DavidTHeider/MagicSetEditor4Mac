// Package vcard liest vCards, wie ein Adressbuch sie ausgibt — ID-04
// (#1703).
//
// # Was es liest
//
// **Name, Kennung, E-Mail, Telefon und ein Bild** — das, was ID-04.AK-02
// übernimmt, und nichts sonst. Nach RFC 6350 (vCard 4.0) und RFC 2426 (3.0,
// die Fassung, die Apple Kontakte schreibt), nachgelesen am 2026-09-27:
// Zeilen enden mit CRLF, eine Zeile, die mit **einem** Leerraum beginnt,
// setzt die vorige fort (6350 §3.2), und Komma, Semikolon, Backslash und
// Zeilenwechsel sind maskiert (§3.4). Ein Bild steht als Base64 mit
// `ENCODING=b` (2426 §3.1.4) oder als `data:`-Adresse (6350 §6.2.4).
//
// # Was es nicht tut
//
// **Nichts holen.** Ein Bild als Adresse bleibt weg und wird genannt — ein
// Adressbuch darf nicht bestimmen, wohin diese Anwendung Anfragen schickt.
// **Kein Quoted-Printable** aus vCard 2.1: es wird genannt statt geraten.
// Und es kennt kein Projekt — es gibt Karten, was daraus wird, entscheidet
// `service`.
//
// **Neu** und nicht portiert: MSE2 hat kein Adressbuch.
package vcard

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Karte ist eine gelesene vCard.
type Karte struct {
	// Zeile ist die Zeile, in der `BEGIN:VCARD` stand.
	Zeile int

	// Kennung ist die UID der Karte — oder Apples `X-ABUID`. Leer, wenn
	// keine dasteht.
	Kennung string

	Name     string
	Emails   []Weg
	Telefone []Weg

	// Bild ist das Bild als Bytes, Bildart seine Art (`image/jpeg`) —
	// beides leer ohne Bild.
	Bild    []byte
	Bildart string
}

// Weg ist eine Adresse oder Nummer mit ihren Arten (`home`, `cell`, …).
type Weg struct {
	Wert  string
	Arten []string
}

// Befund ist, was nicht gelesen wurde — mit Zeile.
type Befund struct {
	Zeile int
	Satz  string
}

// zeile ist eine logische Zeile samt der Zeile, in der sie beginnt.
type zeile struct {
	nummer int
	text   string
}

// Lies liest alle Karten einer Datei.
//
// **Eine kaputte Karte hält die anderen nicht auf** (§8): sie wird genannt,
// und gelesen wird weiter.
func Lies(daten []byte) ([]Karte, []Befund) {
	zeilen := entfalten(string(daten))
	if len(zeilen) == 0 {
		return nil, []Befund{{Satz: "Die Datei enthält keine vCard."}}
	}

	leser := &kartenleser{}

	for _, eine := range zeilen {
		leser.zeile(eine)
	}

	if leser.offen != nil {
		leser.nennen(leser.offen.Zeile, "Die Karte endet nicht mit END:VCARD — sie bleibt weg.")
	}

	return leser.karten, leser.befunde
}

// kartenleser geht Zeile für Zeile.
type kartenleser struct {
	karten  []Karte
	befunde []Befund
	offen   *Karte

	// draussen sagt, dass die vorige Zeile außerhalb einer Karte stand —
	// eine Strecke solcher Zeilen ist **ein** Befund.
	draussen bool
}

func (l *kartenleser) nennen(nummer int, satz string, werte ...any) {
	l.befunde = append(l.befunde, Befund{Zeile: nummer, Satz: fmt.Sprintf(satz, werte...)})
}

func (l *kartenleser) zeile(eine zeile) {
	kopf, wert, gefunden := teilen(eine.text)
	name, parameter := eigenschaft(kopf)

	switch {
	case gefunden && name == "BEGIN" && strings.EqualFold(wert, "VCARD"):
		if l.offen != nil {
			l.nennen(l.offen.Zeile, "Die Karte ab Zeile %d ist noch offen, als Zeile %d eine neue beginnt — sie bleibt weg.",
				l.offen.Zeile, eine.nummer)
		}

		l.offen, l.draussen = &Karte{Zeile: eine.nummer}, false

	case l.offen == nil:
		if !l.draussen {
			l.nennen(eine.nummer, "Zeile %d steht außerhalb einer Karte — sie wird übergangen.", eine.nummer)
		}

		l.draussen = true

	case gefunden && name == "END" && strings.EqualFold(wert, "VCARD"):
		l.abschliessen()

	case !gefunden:
		l.nennen(eine.nummer, "Zeile %d hat keinen Doppelpunkt — sie wird übergangen.", eine.nummer)

	default:
		l.eintragen(eine.nummer, name, parameter, wert)
	}
}

// abschliessen nimmt die offene Karte auf — wenn sie einen Namen hat.
func (l *kartenleser) abschliessen() {
	karte := l.offen
	l.offen = nil

	if strings.TrimSpace(karte.Name) == "" {
		l.nennen(karte.Zeile, "Die Karte ab Zeile %d hat keinen Namen — sie wird nicht übernommen.", karte.Zeile)

		return
	}

	l.karten = append(l.karten, *karte)
}

// eintragen setzt eine Eigenschaft in die offene Karte.
func (l *kartenleser) eintragen(nummer int, name string, parameter map[string][]string, wert string) {
	if kodierung := parameter["ENCODING"]; len(kodierung) > 0 && strings.EqualFold(kodierung[0], "QUOTED-PRINTABLE") {
		l.nennen(nummer, "Zeile %d ist Quoted-Printable (vCard 2.1) — das liest diese Anwendung nicht.", nummer)

		return
	}

	karte := l.offen

	switch name {
	case "FN":
		karte.Name = demaskieren(wert)
	case "N":
		if karte.Name == "" {
			karte.Name = ausN(wert)
		}
	case "UID":
		karte.Kennung = demaskieren(wert)
	case "X-ABUID":
		if karte.Kennung == "" {
			karte.Kennung = demaskieren(wert)
		}
	case "EMAIL":
		karte.Emails = wegDazu(karte.Emails, wert, parameter)
	case "TEL":
		karte.Telefone = wegDazu(karte.Telefone, wert, parameter)
	case "PHOTO":
		l.bild(nummer, parameter, wert)
	}
}

// bild liest ein Bild — als Base64 oder `data:`-Adresse, nie über das Netz.
func (l *kartenleser) bild(nummer int, parameter map[string][]string, wert string) {
	art, roh := "", wert

	switch {
	case strings.HasPrefix(strings.ToLower(wert), "data:"):
		kopf, rest, _ := strings.Cut(wert[len("data:"):], ",")
		art, _, _ = strings.Cut(kopf, ";")
		roh = rest
	case len(parameter["ENCODING"]) > 0 && strings.EqualFold(parameter["ENCODING"][0], "b"):
		if typ := parameter["TYPE"]; len(typ) > 0 {
			art = "image/" + strings.ToLower(typ[0])
		}
	default:
		l.nennen(nummer, "Zeile %d nennt ein Bild als Adresse — es wird nicht geholt.", nummer)

		return
	}

	bild, fehler := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(roh), ""))
	if fehler != nil || len(bild) == 0 {
		l.nennen(nummer, "Das Bild in Zeile %d lässt sich nicht lesen — es bleibt weg.", nummer)

		return
	}

	l.offen.Bild, l.offen.Bildart = bild, strings.ToLower(strings.Replace(art, "jpg", "jpeg", 1))
}

// wegDazu nimmt eine Adresse oder Nummer auf — leere zählen nicht.
func wegDazu(wege []Weg, wert string, parameter map[string][]string) []Weg {
	wert = strings.TrimSpace(demaskieren(wert))
	if wert == "" {
		return wege
	}

	arten := []string{}
	for _, art := range parameter["TYPE"] {
		arten = append(arten, strings.ToLower(art))
	}

	return append(wege, Weg{Wert: wert, Arten: arten})
}

// ausN setzt einen Namen aus `N` zusammen: Vorsatz, Vorname, weitere,
// Familienname, Nachsatz (RFC 6350 §6.2.2).
func ausN(wert string) string {
	teile := maskiertTeilen(wert, ';')
	for len(teile) < 5 {
		teile = append(teile, "")
	}

	reihe := []string{}

	for _, stelle := range []int{3, 1, 2, 0, 4} {
		if teil := strings.TrimSpace(demaskieren(teile[stelle])); teil != "" {
			reihe = append(reihe, teil)
		}
	}

	return strings.Join(reihe, " ")
}
