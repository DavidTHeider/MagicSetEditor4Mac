// **Die Idempotenzfälle der Paketpflege** — abgetrennt am 2026-09-27, weil
// `idempotenzfaelle_test.go` mit dem Plan vor einer Paketänderung über die
// Zeilengrenze aus §2 ging (DOM-01,
// #1696).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Fälle über **Bestand, Karte und Dienste**; hier die über
// das, was **auf der Platte** liegt — Pakete und, seit DOM-02, die Symbole
// der Bibliothek, die als `.mse-symbol` hinaus- und hereingehen — dieselbe Trennung wie zwischen
// `bestandswerkzeuge.go` und `paketpflege.go`. **Die Tabelle bleibt eine**:
// [mitPaketfaellen] setzt diese Fälle in sie ein.

package mcp_test

import "maps"

// paketfaelle sind die Argumente der Paketwerkzeuge.
var paketfaelle = map[string]string{
	"schreiben.paket.installieren": `{"quelle": "/tmp/magic-m15.mse-style"}`,
	"schreiben.paket.anlegen": `{"art": "mse-style", "name": "schlicht", ` +
		`"vollname": "Schlicht", "spiel": "probe"}`,
	"schreiben.paket.kopieren": `{"art": "mse-style", "name": "probe-schlicht"}`,

	// **Der Dateibaum** (FMT-02, #1691) — keines sagt Idempotenz zu: ein
	// zweites Anlegen wird abgewiesen, ein zweites Umbenennen oder Entfernen
	// findet die Datei nicht mehr.
	"schreiben.paketdatei.anlegen":    `{"paket": "probe.mse-style", "datei": "neu", "inhalt": ""}`,
	"schreiben.paketdatei.umbenennen": `{"paket": "probe.mse-style", "datei": "a.png", "neu": "b.png"}`,
	"schreiben.paketdatei.entfernen":  `{"paket": "probe.mse-style", "datei": "a.png"}`,
	"schreiben.paket.uebernehmen":     `{"ort": "/tmp/mse2/data"}`,
	"schreiben.paket.weitergeben":     `{"pakete": ["probe.mse-game"], "ziel": "/tmp/probe.mse-installer"}`,
	"schreiben.paket.einlesen":        `{"datei": "/tmp/probe.mse-installer"}`,
	"schreiben.paket.entfernen": `{"art": "mse-style", "name": "magic-m15", ` +
		`"pfad": "/pakete/magic-m15.mse-style"}`,

	// **Der Plan ist lesend und damit idempotent**; Ersetzen und Zurück sagen
	// keine Idempotenz zu — ein zweites Zurück tauscht wieder zurück.
	"lesen.paket.plan":         `{"vorgang": "entfernen", "art": "mse-style", "name": "magic-m15"}`,
	"schreiben.paket.ersetzen": `{"quelle": "/tmp/magic-m15.mse-style"}`,
	"schreiben.paket.zurueck":  `{"art": "mse-style", "name": "magic-m15"}`,
	"lesen.set.paketbedarf":    `{"setID": 1}`,

	// **Die Symbolbibliothek** (DOM-02, #1697) — die beiden lesenden sagen
	// Idempotenz zu, die schreibenden nicht: ein zweites Aufnehmen legt ein
	// zweites Symbol an, ein zweites Einsetzen eine zweite Datei.
	"lesen.symbolbibliothek":                 `{"suche": "kro"}`,
	"lesen.symbolbibliothek.symbol":          `{"id": 1}`,
	"schreiben.symbolbibliothek.aufnehmen":   `{"setID": 1, "name": "Krone"}`,
	"schreiben.symbolbibliothek.duplizieren": `{"id": 1, "name": "Zacke"}`,
	"schreiben.symbolbibliothek.umbenennen":  `{"id": 1, "name": "Reif"}`,
	"schreiben.symbolbibliothek.loeschen":    `{"id": 1, "bestaetigt": true}`,
	"schreiben.symbolbibliothek.einsetzen":   `{"id": 1, "setID": 1}`,
	"schreiben.symbolbibliothek.einlesen":    `{"inhalt": "mse_version: 0.3.5", "name": "Zinne"}`,

	// **Und die Bearbeitung** — nur das Lesen der Teile sagt Idempotenz zu.
	"lesen.symbolbibliothek.teile":             `{"id": 1}`,
	"schreiben.symbolbibliothek.sichern":       `{"id": 1, "teile": []}`,
	"schreiben.symbolbibliothek.zuruecknehmen": `{"id": 1}`,
	"schreiben.symbolbibliothek.nachziehen":    `{"id": 1}`,
}

// mitPaketfaellen setzt die Paketfälle in die Tabelle ein.
func mitPaketfaellen(faelle map[string]string) map[string]string {
	maps.Copy(faelle, paketfaelle)

	return faelle
}
