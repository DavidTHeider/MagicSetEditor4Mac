package app

// **Was der Bestand als Datei ist** — sichern, verschlüsseln, nur lesen, und
// die Frage, ob ein zweiter Prozess ihn schon hält.
//
// Eigene Datei, weil `anwendung.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort stehen die Nähte, die an einem **Set** oder
// einer **Karte** arbeiten — hier die, die an der **Datei darunter** arbeiten
// und von keinem Set etwas wissen.
//
// **Zwei davon hängen an keiner Verbindung** (`sicherungspruefer`,
// `alleinpruefung`): sie fragen die Datei selbst. Das ist kein Widerspruch zum
// Absatz darüber, sondern seine Zuspitzung — sie sind die einzigen Nähte, die
// auch ohne offene Datenbank etwas sagen können.
import "database/sql"

// bestandsnaehte ist in [App] eingebettet.
type bestandsnaehte struct {
	// bestandssicherer schreibt eine Sicherung der ganzen Datenbank (#799).
	bestandssicherer Bestandssicherer

	// bestandschiffre schaltet die Verschlüsselung des Bestands um
	// (Werkstück 14 von Phase 4.12).
	bestandschiffre Bestandschiffre

	// rueckkehrnaht führt das Buch über die Rückkehren (O7, #1431).
	rueckkehrnaht rueckkehrnaht

	// nurLesen ist die **zweite** Verbindung — Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// **Sie steht nur, solange der Zeiger in einem fremden Profil steht, in
	// das jemand als [domain.RolleLeser] gekommen ist.** Über ihr sind dann
	// alle Repositories gebaut; sie nimmt keinen Schreibvorgang an, und zwar
	// im Treiber.
	//
	// **`nil`, wenn gewöhnlich gearbeitet wird** — dann ist [App.db] die
	// einzige.
	nurLesen *sql.DB

	// ohneLeserwechsel weist einen Wechsel ab, nach dem nur noch gelesen
	// werden dürfte — Werkstück 16 von Phase 4.12.
	//
	// **Für Dienst und MCP-Server**, die ihre Verbindung nicht wechseln
	// können. Der Grund steht bei [Nutzerflaeche.OhneLeserwechsel].
	ohneLeserwechsel bool

	// sicherungspruefer liest eine Sicherung, bevor sie den Bestand ersetzt
	// (#940).
	//
	// **Nil heißt „die echte Prüfung"**, nicht „keine": sie hängt an keiner
	// Verbindung und braucht deshalb keinen Composition Root — anders als
	// [Bestandssicherer], der die offene Datenbank kopiert.
	sicherungspruefer Sicherungspruefer

	// alleinpruefung sagt, ob ein zweiter Prozess die Bestandsdatei hält
	// (#961).
	//
	// **Dieselbe Gestalt wie [Sicherungspruefer]**: sie hängt an keiner
	// Verbindung — sie fragt die Datei selbst — und wird im Composition Root
	// gesetzt, damit es nur eine Stelle gibt, an der steht, was benutzt wird.
	alleinpruefung Alleinpruefung
}
