// **Die Idempotenzfälle des Handels** — abgetrennt am 2026-10-04, weil
// `idempotenzfaelle_test.go` mit `lesen.handelsbeleg` über die Zeilengrenze
// aus §2 ging (Hüllenprobe `nat04handel`,
// #1672).
//
// # Der Schnitt folgt der Frage
//
// Hier stehen die lesenden Werkzeuge über einen **Handelsvorgang** — sein
// Stand, seine Zusagen, sein Verlauf, sein Beleg. Zweimal nachgesehen ist
// derselbe Stand: Lesen bewegt keinen Vorgang, und über diese Fläche gibt es
// keinen Weg, der es täte. **Die Tabelle bleibt eine**: [mitHandelsfaellen]
// setzt diese Fälle in sie ein.

package mcp_test

import "maps"

// handelsfaelle sind die Argumente der lesenden Handelswerkzeuge.
var handelsfaelle = map[string]string{
	// **Der Stand eines Handels** (T1, #1466, angeschlossen mit #1637).
	"lesen.handel": `{"vorgang": "v-1"}`,

	// **Welche Besitzzeilen weggehen** (#1643).
	"lesen.handelszusagen":    `{"vorgang": "v-1"}`,
	"lesen.handelsvorgaenge":  `{"gegenueber": "bdu"}`, // UX-08.AK-04: dieselbe Liste
	"lesen.handelsuebersicht": `{"lagen": ["dran"]}`,   // TRADE-01: dieselbe Liste
	"lesen.handelsprotokoll":  `{"vorgang": "h-1"}`,    // TRADE-03: dasselbe Protokoll
	"lesen.handelsanhalt":     `{"vorgang": "h-1"}`,    // TRADE-02b: derselbe Anhalt
	"lesen.handelsbeleg":      `{"vorgang": "h-1"}`,    // nat04handel: derselbe Beleg
}

// mitHandelsfaellen setzt die Fälle des Handels in die Tabelle ein.
func mitHandelsfaellen(faelle map[string]string) map[string]string {
	maps.Copy(faelle, handelsfaelle)

	return faelle
}
