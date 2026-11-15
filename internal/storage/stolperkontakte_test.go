// **Was der Stolpertreiber der Kontaktablage antwortet** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Warum es diese vier Stellen gibt
//
// **Eine echte Datenbank gibt sie nicht her.** Ein Fehler **mitten** im Lesen
// und eine Zeile, die nicht in ihre Felder passt, entstehen im Treiber und
// nirgends sonst: die Tabellen sind `STRICT` und `NOT NULL`, und was sich mit
// einer Hand an der Datei noch herbeiführen ließe — eine fehlende Tabelle, ein
// krummer Zeitpunkt, ein Auslöser, der abbricht — läuft gegen SQLite selbst
// und steht deshalb in `peerkontaktfehler_test.go`.
package storage

import (
	"database/sql/driver"
	"strings"
)

// kontaktabfrage antwortet auf die zwei Abfragen der Kontaktablage.
func kontaktabfrage(stelle stolperstelle, befehl string) (driver.Rows, bool) {
	if zeilen, erkannt := kontaktzeilen(stelle, befehl); erkannt {
		return zeilen, true
	}

	return rechtezeilen(stelle, befehl)
}

// kontaktzeilen antwortet auf die Abfrage über die Kontakte selbst.
func kontaktzeilen(stelle stolperstelle, befehl string) (driver.Rows, bool) {
	if !strings.HasPrefix(befehl, "\nSELECT id, kennung, fingerabdruck") {
		return nil, false
	}

	// **Die Spalten sind die der Abfrage**, und eine zu wenig verschiebt
	// still, was dieser Treiber misst: der Scan scheitert dann an der
	// Spaltenzahl, die Rechte werden nie gefragt, und die beiden Stellen
	// darin bleiben unbetreten — bei **grünem** Lauf. Genau das ist am
	// 2026-09-14 geschehen, als `adresse` zur Abfrage kam (N2, #1455);
	// gefunden hat es der Riegel, der unbetretene Anweisungen zählt.
	// **Seit `T3` (#1468) sind es zehn**: `erstmals` und `notiz` kamen dazu.
	// **Seit NET-01 (#1705) elf**: `vorgaenger` — und der Riegel fand es
	// wieder, diesmal an `erstfehler`.
	spalten := []string{
		"id", "kennung", "fingerabdruck", "name", "eigener_name", "stand",
		"gesehen", "adresse", "erstmals", "notiz", "vorgaenger",
	}

	// **Für die Rechtestellen kommt hier eine *heile* Zeile.** Ohne sie
	// endete die Schleife bei null Kontakten, die Rechte würden nie gefragt,
	// und der Lauf wäre grün, ohne die Stelle betreten zu haben (§10).
	if stelle == beimRechteende || stelle == beimRechtescan {
		return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{
			{
				int64(1), "bwer", "ab", "Wer", "", "unbestaetigt",
				"2026-09-14T03:00:00Z", "", "", "", "",
			},
		}}, true
	}

	// **Eine erste Begegnung, die keine ist** (T3, #1468). Sie kommt aus der
	// Datei; wer sie für den Nullwert hielte, machte aus einer kaputten Zeile
	// ein „nicht bekannt" — und das ist eine Aussage (§8).
	if stelle == beimKontaktfrist {
		return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{
			{
				int64(1), "bwer", "ab", "Wer", "", "unbestaetigt",
				"2026-09-14T03:00:00Z", "", "vorgestern", "", "",
			},
		}}, true
	}

	if stelle != beimKontaktende && stelle != beimKontaktscan {
		return nil, false
	}

	if stelle == beimKontaktende {
		return &druckzeilen{spalten: spalten, stolpert: true}, true
	}

	// **Ein `NULL`, wo eine Zeichenkette hingehört** — daran scheitert der
	// Scan, und nur daran. Eine Zahl täte es nicht: `database/sql` wandelt
	// sie stillschweigend um, und der Lauf bliebe grün.
	return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{
		{
			int64(1), nil, "a", "b", "c", "unbestaetigt",
			"2026-09-14T03:00:00Z", "", "", "", "",
		},
	}}, true
}

// rechtezeilen antwortet auf die Abfrage über die Freigaben.
func rechtezeilen(stelle stolperstelle, befehl string) (driver.Rows, bool) {
	if stelle != beimRechteende && stelle != beimRechtescan {
		return nil, false
	}

	if !strings.HasPrefix(befehl, "SELECT zweck FROM peerrechte") {
		return nil, false
	}

	spalten := []string{"zweck"}

	if stelle == beimRechteende {
		return &druckzeilen{spalten: spalten, stolpert: true}, true
	}

	return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{
		{nil},
	}}, true
}
