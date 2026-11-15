package storage

// **Wo nachgesehen wird** — die Tafel der Datenpflege.
//
// # Warum sie in einer eigenen Datei steht
//
// **Weil sie eine andere Frage beantwortet als der Rest.** `datenpflege.go`
// sagt, **wie** nachgesehen und ausgeführt wird — lesen, Revision rechnen,
// im Geschäft löschen. Hier steht, **wo**: dreizehn Stellen, jede mit ihrer
// Klasse. Wer eine vierzehnte aufnimmt, fasst diese Datei an und keine
// andere.
//
// **Der Anlass war die Zeilengrenze** (§2): mit der Stelle für alte
// Protokolle (G5,
// #1462)
// stand die gemeinsame Datei bei 403 Zeilen. Der Schnitt folgt trotzdem der
// Frage und nicht der Zahl — er war vorher schon der richtige.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// pflegestellen sind die dreizehn Stellen, an denen nachgesehen wird.
//
// **Jede mit ihrer Klasse, und die Klasse ist die Entscheidung.** Was
// `sicher` heißt, zeigt auf nichts und hält nichts fest. Was `entscheidung`
// heißt, sieht nach einem Rest aus und **kann** ein Nachweis sein. Was
// `nurInformieren` heißt, wird **nicht** angefasst — dort steht deshalb keine
// Löschabfrage, und das ist keine Vergesslichkeit, sondern die Zusage.
func pflegestellen() []pflegestelle {
	return []pflegestelle{
		{
			art: "kartenwerteOhneKarte", klasse: domain.PflegeSicher,
			kennungen: `SELECT 'card_values:' || card_id || ':' || feld FROM card_values
			              WHERE card_id NOT IN (SELECT id FROM cards)
			              ORDER BY card_id, feld`,
			loeschen: `DELETE FROM card_values WHERE card_id NOT IN (SELECT id FROM cards)`,
		},
		{
			art: "setangabenOhneSet", klasse: domain.PflegeSicher,
			kennungen: `SELECT 'set_values:' || set_id || ':' || feld FROM set_values
			              WHERE set_id NOT IN (SELECT id FROM sets)
			              ORDER BY set_id, feld`,
			loeschen: `DELETE FROM set_values WHERE set_id NOT IN (SELECT id FROM sets)`,
		},
		{
			art: "suchzeilenOhneKarte", klasse: domain.PflegeSicher,
			// **`card_suche` ist eine FTS5-Tabelle ohne eigene Kennung** — was
			// eine Zeile ausmacht, ist ihre `rowid`.
			kennungen: `SELECT 'card_suche:' || rowid FROM card_suche
			              WHERE card_id NOT IN (SELECT id FROM cards) ORDER BY rowid`,
			loeschen: `DELETE FROM card_suche WHERE card_id NOT IN (SELECT id FROM cards)`,
		},
		{
			art: "deckplaetzeOhneDeck", klasse: domain.PflegeSicher,
			kennungen: `SELECT 'deckplaetze:' || id FROM deckplaetze
			              WHERE deck_id NOT IN (SELECT id FROM decks) ORDER BY id`,
			loeschen: `DELETE FROM deckplaetze WHERE deck_id NOT IN (SELECT id FROM decks)`,
		},
		{
			art: "zuordnungenOhneExemplar", klasse: domain.PflegeSicher,
			kennungen: `SELECT 'deckzuordnungen:' || id FROM deckzuordnungen
			              WHERE exemplar_id NOT IN (SELECT id FROM sammlung) ORDER BY id`,
			loeschen: `DELETE FROM deckzuordnungen
			             WHERE exemplar_id NOT IN (SELECT id FROM sammlung)`,
		},
		{
			art: "verlaufOhneSet", klasse: domain.PflegeSicher,
			kennungen: `SELECT 'arbeitsverlauf:' || id FROM arbeitsverlauf
			              WHERE art = 'set' AND kennung NOT IN (SELECT CAST(id AS TEXT) FROM sets)
			              ORDER BY id`,
			loeschen: `DELETE FROM arbeitsverlauf
			             WHERE art = 'set' AND kennung NOT IN (SELECT CAST(id AS TEXT) FROM sets)`,
		},
		{
			art: "abgelaufeneNachweise", klasse: domain.PflegeEntscheidung,
			kennungen: `SELECT 'freigabenachweise:' || kennung FROM freigabenachweise
			              WHERE eingeloest = 0 AND ablauf < ? ORDER BY kennung`,
			loeschen:     `DELETE FROM freigabenachweise WHERE eingeloest = 0 AND ablauf < ?`,
			zeitgebunden: true,
		},
		{
			art: "abgelaufeneToken", klasse: domain.PflegeEntscheidung,
			kennungen: `SELECT 'zugriffstoken:' || id FROM zugriffstoken
			              WHERE ablauf <> '' AND ablauf < ? ORDER BY id`,
			loeschen:     `DELETE FROM zugriffstoken WHERE ablauf <> '' AND ablauf < ?`,
			zeitgebunden: true,
		},
		{
			// **Gleiche Printingdaten sind nicht automatisch Duplikate** — das
			// Issue sagt es, und die Sammlung führt genau die Felder, die sie
			// unterscheiden: Zustand, Sprache, Ausführung, Erwerb, Lagerort.
			// Was hier steht, ist deshalb eine **Auskunft** und kein Auftrag.
			art: "moeglicheDubletten", klasse: domain.PflegeNurInformieren,
			kennungen: `SELECT 'sammlung:' || profil_id || ':' || printing FROM sammlung
			              WHERE printing <> '' AND wunsch = 0
			              GROUP BY profil_id, printing HAVING count(*) > 1
			              ORDER BY profil_id, printing`,
		},
		{
			// **Ein beendetes Protokoll steht nach der Aufbewahrungsfrist
			// zur Entscheidung** (G5,
			// #1462).
			//
			// **`stand = 'beendet'` und nicht „alt genug"**: eine Partie, an
			// der noch jemand sitzt, ist kein Protokoll. Und die Frist zählt
			// ab `geaendert` — das ist der letzte Zug, nicht der erste.
			//
			// **Die Tage kommen aus [domain.AufbewahrungPartien]** und
			// stehen nicht in diesem Text; der Grund steht dort.
			art: domain.PflegeProtokolle, klasse: domain.PflegeEntscheidung,
			//
			// **Und sie klemmt auf das Profil.** Eine Partie **hat** eines,
			// anders als die verwaisten Zeilen daneben; ohne die Klausel
			// böte dieser Lauf das Protokoll eines anderen Profils zum
			// Löschen an. `internal/werkzeuge/profilklausel_test.go` hält
			// das.
			kennungen: fmt.Sprintf(`SELECT 'partien:' || id FROM partien
			              WHERE profil_id = ? AND stand = 'beendet'
			                AND geaendert < datetime(?, '-%d days')
			              ORDER BY id`, domain.AufbewahrungPartien),
			loeschen: fmt.Sprintf(`DELETE FROM partien
			             WHERE profil_id = ? AND stand = 'beendet'
			               AND geaendert < datetime(?, '-%d days')`,
				domain.AufbewahrungPartien),
			profilgebunden: true,
			zeitgebunden:   true,
		},
		{
			// **Die Integrität selbst** — Befund 2 des Zweitblicks über Phase
			// 5.2. Das Issue nennt „Integrität/Fremdschlüssel" zuerst, und
			// die zehn Stellen darüber sahen nur nach **bekannten** Resten:
			// jede prüft eine Tafel, die jemand aufgezählt hat. Was SQLite
			// selbst über seine Datei sagt, stand nicht darin.
			//
			// **Und es wird nichts gelöscht**: „echte Korruption nicht durch
			// kosmetisches Löschen verdecken" steht so im Issue. Ein Befund
			// hier ist die Aufforderung, eine Sicherung einzuspielen — keine
			// Maßnahme, die ein Knopf erledigt.
			art: "integritaet", klasse: domain.PflegeNurInformieren,
			klageIstBefund: true,
			kennungen: `SELECT 'integritaet:' || integrity_check
			              FROM pragma_integrity_check
			             WHERE integrity_check <> 'ok' ORDER BY 1`,
		},
		{
			// **Und die Fremdschlüssel** — dieselbe Überlegung. `PRAGMA
			// foreign_keys` steht in diesem Bestand an; eine verwaiste Zeile
			// entsteht damit nicht im Betrieb, wohl aber durch eine Sicherung
			// von Hand, eine Wanderung oder ein fremdes Werkzeug.
			art: "fremdschluessel", klasse: domain.PflegeNurInformieren,
			klageIstBefund: true,
			kennungen: `SELECT 'fremdschluessel:' || "table" || ':' ||
			                   coalesce(CAST(rowid AS TEXT), '?')
			              FROM pragma_foreign_key_check ORDER BY 1`,
		},
		{
			// **Eine offene Rückkehr wird nie weggeräumt** (O7,
			// [0088](../../docs/entscheidungen/0088-die-epoche-einer-rueckkehr.md)):
			// sie ist der Nachweis darüber, dass etwas unklar ist, und ein
			// Nachweis, der mit dem Alter verschwindet, ist keiner.
			art: "offeneRueckkehren", klasse: domain.PflegeNurInformieren,
			kennungen: `SELECT 'wiederherstellungen:' || id FROM wiederherstellungen
			              WHERE klaerung = 'offen' ORDER BY id`,
		},
	}
}
