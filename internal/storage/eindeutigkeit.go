package storage

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine verletzte Eindeutigkeit sagt, woran es liegt** — ENG-05.AK-03
// (#1730),
// nativer Befund F08.
//
// # Was vorher geschah
//
// **Jede hieß „ein Eintrag mit diesen Angaben liegt schon da"**, und das
// Fenster erklärte sie als fremde Änderung. Ein zweites Deck gleichen Namens
// bekam „Der Eintrag wurde zwischenzeitlich geändert. Bitte neu laden." — und
// wer neu lud, fand dasselbe Formular und denselben Namen darin.
//
// # Die Art entscheidet der Code, das Feld die Tafel
//
// **Dass es eine verletzte Eindeutigkeit ist**, sagt der erweiterte Fehlercode
// ([einordnen], §8). **Wo sie steht**, sagt nur der Text des Treibers:
// `UNIQUE constraint failed: decks.profil_id, decks.spiel, decks.name` — SQLite
// kennt für eine Eindeutigkeit keinen Namen, den es zurückgäbe. Gelesen wird
// der Text deshalb **nur für den Feldbezug**: fällt die Zerlegung bei einer
// neuen Fassung des Treibers aus, bleibt es ein doppelter Eintrag ohne Feld,
// aber nie eine falsche Art. `TestDieEindeutigkeitNenntTafelUndSpalten` wird
// dann rot.
//
// # Und die Zuordnung steht an einer Stelle
//
// **Hier**, und ein Wächter hält jede Zeile gegen das Schema
// (`TestJedeZuordnungZeigtAufEineEindeutigkeit`). Was nicht dasteht, ist
// `doppelt` ohne Feld — richtig, nur weniger genau.

// eindeutigkeitsziel ist, wo eine Eindeutigkeit im Fenster erscheint.
type eindeutigkeitsziel struct {
	feld       string
	schluessel domain.Schluessel
}

// eindeutigkeiten ordnet die Eindeutigkeiten, an die eine Eingabe stoßen kann,
// ihrem Feld zu — geschlüsselt wie der Treiber sie nennt: Tafel und Spalten.
//
// **Nur, was aus einer Eingabe erreichbar ist.** Eine Eindeutigkeit, die der
// Bestand selbst vergibt — eine Reihenfolge, eine Folgenummer —, trifft kein
// Mensch; käme sie trotzdem, ist `doppelt` ohne Feld die richtige Auskunft.
var eindeutigkeiten = map[string]eindeutigkeitsziel{
	"decks(profil_id, spiel, name)":                   {"name", domain.SchluesselDoppelterName},
	"turniere(profil_id, spiel, name)":                {"name", domain.SchluesselDoppelterName},
	"formate(profil_id, spiel, name)":                 {"name", domain.SchluesselDoppelterName},
	"symbolbibliothek(profil_id, name)":               {"name", domain.SchluesselDoppelterName},
	"profile(nutzer_id, name)":                        {"name", domain.SchluesselDoppelterName},
	"nutzer(name)":                                    {"name", domain.SchluesselDoppelterName},
	"zugriffstoken(profil_id, bezeichnung)":           {"bezeichnung", domain.SchluesselDoppelterName},
	"nutzer(email)":                                   {"email", domain.SchluesselDoppelt},
	"nutzer(telefon)":                                 {"telefon", domain.SchluesselDoppelt},
	"spiel_autoersetzungen(profil_id, spiel, muster)": {"muster", domain.SchluesselDoppelt},
	"sprachpakete(kennung)":                           {"kennung", domain.SchluesselDoppelt},

	// **Die Sammlung hat kein einzelnes Feld** — doppelt ist die Verbindung
	// aus Druck, Zustand, Sprache und Ausführung, und ihr Satz sagt das. Der
	// Druck zählt über Alias **und** Zeile (Migration 0086): sonst waren zwei
	// verschiedene vorläufige Drucke dieselbe Zeile.
	"sammlung(profil_id, printing, druck_id, zustand, sprache, finish, wunsch)": {
		"", domain.SchluesselDoppeltesExemplar,
	},

	// **Und eine Zeile ohne Druckzeile** (Migration 0087) — nur ein
	// Altbestand trägt sie, und für sie gilt die Regel von vorher.
	"sammlung(profil_id, printing, zustand, sprache, finish, wunsch)": {
		"", domain.SchluesselDoppeltesExemplar,
	},
}

// eindeutigkeitsmarke steht im Treibertext vor Tafel und Spalten.
const eindeutigkeitsmarke = "UNIQUE constraint failed: "

// doppelt ordnet eine verletzte Eindeutigkeit ein.
func doppelt(fehler error) error {
	ziel := eindeutigkeitsziel{schluessel: domain.SchluesselDoppelt}

	if tafel, spalten := verletzteEindeutigkeit(fehler); tafel != "" {
		if bekannt, ist := eindeutigkeiten[tafel+"("+strings.Join(spalten, ", ")+")"]; ist {
			ziel = bekannt
		}
	}

	return domain.NewFieldError(ziel.feld, ziel.schluessel, "%s", satzZu(ziel.schluessel)).MitGrund(fehler)
}

// satzZu ist der Satz für die Flächen ohne Katalog — Kommandozeile, MCP, REST.
func satzZu(schluessel domain.Schluessel) string {
	switch schluessel {
	case domain.SchluesselDoppelterName:
		return "Diesen Namen gibt es hier schon — wähle einen anderen."

	case domain.SchluesselDoppeltesExemplar:
		return "Diesen Druck führt die Sammlung schon — im selben Zustand, in " +
			"derselben Sprache und Ausführung. Ändere dort die Menge oder " +
			"wähle eine andere Angabe."
	}

	return "Einen Eintrag mit diesen Angaben gibt es schon."
}

// verletzteEindeutigkeit liest Tafel und Spalten aus dem Text des Treibers.
//
// **Leer, wenn der Text nicht die erwartete Gestalt hat** — dann bleibt der
// Befund ohne Feld.
func verletzteEindeutigkeit(fehler error) (tafel string, spalten []string) {
	if fehler == nil {
		return "", nil
	}

	_, rest, gefunden := strings.Cut(fehler.Error(), eindeutigkeitsmarke)
	if !gefunden {
		return "", nil
	}

	for teil := range strings.SplitSeq(rest, ", ") {
		vorn, spalte, mitPunkt := strings.Cut(strings.TrimSpace(teil), ".")
		if !mitPunkt || (tafel != "" && vorn != tafel) {
			return "", nil
		}

		tafel = vorn
		spalten = append(spalten, spalte)
	}

	return tafel, spalten
}
