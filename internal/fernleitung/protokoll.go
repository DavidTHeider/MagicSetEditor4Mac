package fernleitung

// **Kein Kennwort ins Protokoll** — Befund Z1 des Zweitblicks über Phase 6.7
// (NET-Teil).
//
// # Warum hier und nicht im Protokoll der Anwendung
//
// **Weil nur hier bekannt ist, dass eine fremde Bibliothek Adressen samt
// Zugangsdaten in ihre Sätze setzt.** tailcat schreibt ohne eigenen Logger
// mit `log.Printf` (`tailcat.go`; zwischen v0.6.0 und v0.7.0 kam genau eine
// Zeile hinzu, `exec %v: %v`, nachgelesen am 2026-10-07), und die gepinnte
// Tailscale-Fassung setzt bei einer Proxyabsage die ganze Proxyadresse in
// ihren Fehler (`derp/derphttp/derphttp_client.go`, Zeile 918 in
// v1.103.0-pre.0.20260916030321-a2263542f260); `url.URL.String()` trägt
// Benutzer und Kennwort mit. Das Protokoll der Anwendung schwärzt nur das
// Zuhause.
//
// **Was bereinigt wird**: `schema://benutzer:kennwort@` wird zu
// `schema://…@`, in jeder Zeile und in jedem Fehler, der von tailcat
// heraufkommt. Was keine Adresse ist, bleibt, wie es war.

import (
	"fmt"
	"log"
	"regexp"
)

// zugangsdaten trifft den Teil einer Adresse vor dem `@`.
var zugangsdaten = regexp.MustCompile(`([A-Za-z][A-Za-z0-9+.-]*://)[^/@\s]*@`)

// einladungsschluessel trifft die Schlüssel, aus denen eine Einladung besteht
// — SEC-21 (#2162). Eine Einladung trägt den öffentlichen Knotenschlüssel,
// den Disco-Schlüssel („the unguessable part of the server's tailcat
// address") und den Pre-Shared Key (`ConnInfo`, `tailcat.go`, Zeilen
// 155–169, in v0.6.0 und v0.7.0 gleich); tailcat schreibt die ersten beiden
// mit jeder `NetworkMap` hinaus. Ohne den Pre-Shared Key kommt damit niemand
// in den Tunnel — aber ein Protokoll, das man herumreicht, soll keinen Teil
// einer Einladung tragen.
var einladungsschluessel = regexp.MustCompile(`\b(nodekey|discokey):[0-9a-f]+`)

func ohneZugangsdaten(text string) string {
	text = einladungsschluessel.ReplaceAllString(text, "${1}:…")

	return zugangsdaten.ReplaceAllString(text, "${1}…@")
}

// protokollNach gibt einen Logger in der Form, die tailcat erwartet — mit
// bereinigten Zeilen.
func protokollNach(schreibe func(string)) func(format string, werte ...any) {
	return func(format string, werte ...any) {
		schreibe(ohneZugangsdaten(fmt.Sprintf(format, werte...)))
	}
}

// tailcatprotokoll schreibt, wohin `log` schreibt — nur bereinigt.
var tailcatprotokoll = protokollNach(func(zeile string) { log.Print(zeile) })

// bereinigterFehler zeigt seinen Satz ohne Zugangsdaten und behält die Kette
// darunter — `errors.Is` findet, was es vorher fand.
type bereinigterFehler struct{ unten error }

func (b bereinigterFehler) Error() string { return ohneZugangsdaten(b.unten.Error()) }

func (b bereinigterFehler) Unwrap() error { return b.unten }

// ohneZugangsdatenFehler bereinigt einen Fehler von tailcat — `nil` bleibt.
func ohneZugangsdatenFehler(fehler error) error {
	if fehler == nil {
		return nil
	}

	return bereinigterFehler{unten: fehler}
}
