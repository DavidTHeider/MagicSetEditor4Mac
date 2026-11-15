package main

// **Der Dienst hält eine KI-Anwendung, nicht eine je Ruf** — angestrichen
// beim Bau von A13.3
// (#1452).
//
// # Der Befund
//
// **Ein Batch über den Dienst gab Kennungen aus, die niemand einlösen
// konnte.** [service.KIAuftragswerk] hält seine offenen Aufträge **im
// Prozess** und nicht im Bestand — das steht dort so, und es ist richtig: ein
// Vorschlag gilt eine Frist lang und ist kein Besitz. Wer aber je Ruf eine
// neue Anwendung baut, baut je Ruf ein neues Werk, und das kennt keinen
// Auftrag von vorhin.
//
// **Dieselbe Gestalt wie an der MCP-Naht**, wo sie beim Bau von A13.1b
// gefunden und behoben wurde. Dort stand daneben der Satz, `mse-rest` baue
// „je Aufruf neu" — **das ist falsch**: `mse-rest` ist ein Dienst, seine
// Nahttafel entsteht **einmal** beim Start (`dienst.go`), und sie lebt so
// lange wie der Prozess. Der Satz hat die Korrektur hier verhindert.
//
// # Was dieser Lauf hält und was nicht
//
// **Dass es dieselbe Anwendung ist.** Dass ein Auftrag darin wirklich
// einlösbar bleibt, hält er nicht — dafür braucht es ein Modell, und das
// steht hinter `netzlauf`.

import (
	"database/sql"
	"testing"
)

// **Zweimal gefragt, einmal gebaut.**
func TestDerDienstHaeltEineKIAnwendung(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var ohneBestand *sql.DB

	naht := &preise{db: ohneBestand}

	erste := naht.kianwendung()
	if erste == nil {
		t.Fatal("die erste Anwendung fehlt")
	}

	if zweite := naht.kianwendung(); zweite != erste {
		t.Error("der zweite Ruf baute eine neue Anwendung — ein Auftrag aus " +
			"dem ersten wäre damit nicht einlösbar")
	}
}
