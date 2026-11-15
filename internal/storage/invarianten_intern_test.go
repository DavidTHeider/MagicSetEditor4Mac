package storage

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

// **Die Invarianten des Bestands, aus dem Schema gelesen** — ENG-05.AK-01
// (#1730).
//
// # Was hier geprüft wird
//
// **Jede Klasse des Katalogs in `docs/INVARIANTEN.md` hat hier einen
// Wächter, der das Schema selbst befragt** — nicht eine Liste daneben. Eine
// neue Tafel, eine neue Menge, ein neuer Fremdschlüssel erzwingt eine
// Entscheidung: eine Bedingung im Schema, oder eine Zeile in den Listen
// unten, **mit Grund**.
//
// # Was hier nicht geprüft wird
//
//   - **Ob eine Bedingung das Richtige verlangt.** Der Wächter sieht, dass
//     an `menge` ein `CHECK` hängt, nicht, dass er `>= 1` sagt. Wo es darauf
//     ankommt, prüft ein Lauf das Verhalten (`sammlungsregeln_intern_test.go`).
//   - **Ob ein Grund stimmt.** Er steht da; nachgelesen wird er im Review.
//   - **Die Übergänge eines Zustands.** Sie stehen in der Domäne, und dort
//     gehören sie hin (AK-02: kein Auslöser als zweite Geschäftslogik).

// kennungenOhneVerweis sind Kennungsspalten, die mit Absicht auf nichts
// verweisen — **jede mit ihrem Grund**. Die Wurzeln eines Profils stehen
// nicht hier, sondern in [wurzeln].
var kennungenOhneVerweis = map[string]string{
	"freigabenachweise.nutzer_id": "0047: ein gelöschter Nutzer nimmt seine Profile mit (0037), und " +
		"über `profil_id … ON DELETE CASCADE` gehen die Nachweise ohnehin",
	"handelsausgang.nachricht_id": "keine Zeile dieses Bestands, sondern die Kennung einer Nachricht an der " +
		"Gegenstelle (0073)",
	"handelsjournal.exemplar_id": "zeigt auf die Besitzzeile, als gebucht wurde — ein Abgang löscht sie " +
		"(`besitzzeileLoeschen`), das Journal bleibt; 0 heißt: ohne Zeile (0074)",
	"history.karte_id": "ein Verlaufsschritt überdauert seine Karte, sonst ließe sich ihr Löschen " +
		"nicht zurücknehmen (`HandlungKarte`); 0 heißt: das Set selbst (0006)",
	"matchbindungen.partie_id": "0067 nennt keinen Grund; die Datenpflege löscht beendete Partien nach " +
		"der Frist, und die Bindung bleibt — der Verweis kann ins Leere zeigen (INVARIANTEN.md, offen)",
	"matchmeldungen.partie_id": "wie `matchbindungen.partie_id`",
}

// wurzeln sind die Tafeln, deren `profil_id` auf kein Profil verweist.
//
// **Absicht, und seit 0035 benannt**: ein `ON DELETE CASCADE` hieße, dass ein
// Klick Sets, Decks, Sammlung und Turniere eines Profils löscht. Was bleibt,
// ist unsichtbar und nicht weg (`ProfilLoeschen`). Dass es keinem neuen
// Profil zufällt, hält [TestJedeWurzelGehoertEinemProfil] über
// `AUTOINCREMENT`.
var wurzeln = map[string]string{
	"sets": "0035", "decks": "0035", "turniere": "0035", "formate": "0035",
	"sammlung": "0035", "settings": "0035", "export_options": "0035",
	"spiel_listenspalten": "0035", "spiel_sortierung": "0035",
	"spiel_autoersetzungen": "0035", "spiel_autoersetzung_an": "0035",
	"angeheftet":     "0050: dieselbe Gestalt wie die Wurzeln aus 0035; was es nicht mehr gibt, räumt der Dienst",
	"arbeitsverlauf": "0050: wie `angeheftet`",
	"retrieval_dokumente": "0060: 0 heißt öffentlich, und ein Profil geht über den Dienst — ein " +
		"CASCADE räumte hinter dem Rücken der Sperren weg",
	"retrieval_gesperrt":       "0060: 0 heißt eine Sperre für alle",
	"wiederherstellungsmengen": "0053: hängt an `wiederherstellungen`; das Profil grenzt nur die Mengen ab",
}

// wurzelnMitNull sind die Wurzeln, deren `profil_id` eine Vorgabe von 0 trägt
// — sie kam mit `ALTER TABLE … ADD COLUMN` (0035) — und darum einen Auslöser
// brauchen, der eine fehlende Angabe abweist.
var wurzelnMitNull = []string{"sets", "decks", "turniere", "formate"}

// ohneBedingung sind Mengen und Beträge, die nur die Domäne hält — **jede mit
// ihrem Grund**.
var ohneBedingung = map[string]string{
	"set_pack_items.anzahl": "fremde Daten aus einem Paket: MSE2 kennt für `amount` keine Untergrenze " +
		"(src/data/pack.hpp, Zeile 68, Fassung f9d9356 vom 2021-07-06)",
	"sammlung.erwerbspreis": "Text: leer heißt kein Preis, sonst Ziffern in der kleinsten Einheit (0078); " +
		"die Domäne prüft ihn (`Exemplar.Preisbefunde`), und kein Geldfluss hängt daran",
	"sammlung.erwerbswaehrung": "leer heißt: kein Preis — die Domäne prüft die Währung zusammen mit dem " +
		"Preis (`Exemplar.Preisbefunde`)",
	"preisstaende.waehrung": "fremde Daten eines Preisdienstes, gelesen, wie sie kommen (§16)",
}

// fremdschluesselOhneIndex sind Verweise, deren Spalten kein Index anführt.
//
// **Leer, und so soll es bleiben**: ohne Index sucht SQLite beim Löschen der
// Elternzeile jede Kindzeile ab (sqlite.org/foreignkeys.html, Abschnitt 3).
// Gemessen und behoben in Wanderung 0080 (`docs/ABFRAGEPLAENE.md`).
var fremdschluesselOhneIndex = map[string]string{}

func TestJedeKennungsspalteVerweistOderStehtBegruendet(t *testing.T) {
	t.Parallel()

	bild := schemaAus(t)

	verwiesen := map[string]bool{}
	for _, verweis := range bild.verweise {
		for _, spalte := range verweis.spalten {
			verwiesen[verweis.tafel+"."+spalte] = true
		}
	}

	gesehen := map[string]bool{}

	for _, spalte := range bild.spalten {
		name := spalte.tafel + "." + spalte.name
		if !strings.HasSuffix(spalte.name, "_id") || verwiesen[name] {
			continue
		}

		if spalte.name == "profil_id" && wurzeln[spalte.tafel] != "" {
			continue
		}

		if kennungenOhneVerweis[name] == "" {
			t.Errorf("%s verweist auf nichts und steht nicht in kennungenOhneVerweis", name)
		}

		gesehen[name] = true
	}

	for name := range kennungenOhneVerweis {
		if !gesehen[name] {
			t.Errorf("%s steht in kennungenOhneVerweis, verweist aber oder ist fort", name)
		}
	}
}

func TestJederFremdschluesselHatSeinenIndex(t *testing.T) {
	t.Parallel()

	bild := schemaAus(t)

	for _, verweis := range bild.verweise {
		name := verweis.tafel + "(" + strings.Join(verweis.spalten, ", ") + ")"

		gedeckt := slices.ContainsFunc(bild.indizes[verweis.tafel], func(folge []string) bool {
			return beginntMit(folge, verweis.spalten)
		})

		if !gedeckt && fremdschluesselOhneIndex[name] == "" {
			t.Errorf("%s → %s: kein Index beginnt mit diesen Spalten", name, verweis.ziel)
		}
	}
}

// **Jede Zeile gehört einem Profil** — über einen Fremdschlüssel auf
// `profile` oder als Wurzel, und eine Profilkennung wird nie wieder vergeben.
func TestJedeWurzelGehoertEinemProfil(t *testing.T) {
	t.Parallel()

	bild := schemaAus(t)

	verweisAufProfil := map[string]string{}
	for _, verweis := range bild.verweise {
		if slices.Contains(verweis.spalten, "profil_id") {
			verweisAufProfil[verweis.tafel] = verweis.ziel + " " + verweis.beimLoeschen
		}
	}

	gesehen := map[string]bool{}

	for _, spalte := range bild.spalten {
		if spalte.name != "profil_id" {
			continue
		}

		if !spalte.pflicht {
			t.Errorf("%s.profil_id darf leer sein", spalte.tafel)
		}

		verweis, hatVerweis := verweisAufProfil[spalte.tafel]

		switch {
		case hatVerweis && !strings.HasSuffix(verweis, "CASCADE"):
			t.Errorf("%s.profil_id verweist ohne CASCADE: %s", spalte.tafel, verweis)

		case !hatVerweis && wurzeln[spalte.tafel] == "":
			t.Errorf("%s.profil_id verweist auf nichts und steht nicht in wurzeln", spalte.tafel)

		case !hatVerweis:
			gesehen[spalte.tafel] = true
		}
	}

	for tafel := range wurzeln {
		if !gesehen[tafel] {
			t.Errorf("%s steht in wurzeln, verweist aber auf ein Profil oder ist fort", tafel)
		}
	}

	// **Eine Wurzel mit Vorgabe 0 braucht ihren Auslöser**, sonst ginge eine
	// Zeile ohne Profil still hinein.
	for _, tafel := range wurzelnMitNull {
		if bild.ausloeser[tafel+"_profil_pflicht"] == "" || bild.ausloeser[tafel+"_profil_bleibt"] == "" {
			t.Errorf("%s hat nicht beide Auslöser für die fehlende profil_id", tafel)
		}
	}

	// **Eine gelöschte Kennung kommt nie wieder** — sonst fielen die Wurzeln
	// eines gelöschten Profils dem nächsten neuen zu.
	for _, tafel := range []string{"profile", "nutzer"} {
		if !strings.Contains(bild.definition[tafel], "AUTOINCREMENT") {
			t.Errorf("%s vergibt Kennungen wieder — ohne AUTOINCREMENT", tafel)
		}
	}
}

// mengenspalte erkennt Spalten, die eine Stückzahl oder einen Betrag halten.
var mengenspalte = regexp.MustCompile(`^(menge|anzahl)$|betrag|preis|waehrung$`)

func TestJedeMengeUndJederBetragHatIhreBedingung(t *testing.T) {
	t.Parallel()

	bild := schemaAus(t)
	gesehen := map[string]bool{}

	for _, spalte := range bild.spalten {
		if !mengenspalte.MatchString(spalte.name) {
			continue
		}

		name := spalte.tafel + "." + spalte.name
		if bedingtIm(bild, spalte.tafel, spalte.name) {
			continue
		}

		gesehen[name] = true

		if ohneBedingung[name] == "" {
			t.Errorf("%s hat weder CHECK noch Auslöser und steht nicht in ohneBedingung", name)
		}
	}

	for name := range ohneBedingung {
		if !gesehen[name] {
			t.Errorf("%s steht in ohneBedingung, hat aber eine oder ist fort", name)
		}
	}
}

// bedingtIm sagt, ob eine CHECK-Bedingung der Tafel oder ein Auslöser an ihr
// die Spalte nennt.
func bedingtIm(bild schemabild, tafel, spalte string) bool {
	wort := regexp.MustCompile(`\b` + regexp.QuoteMeta(spalte) + `\b`)

	for _, bedingung := range pruefungenIn(bild.definition[tafel]) {
		if wort.MatchString(bedingung) {
			return true
		}
	}

	for name, definition := range bild.ausloeser {
		if bild.tafelVon[name] == tafel && strings.Contains(definition, "NEW."+spalte) {
			return true
		}
	}

	return false
}

// pruefungenIn gibt den Ausdruck jeder CHECK-Bedingung einer Tafel.
func pruefungenIn(definition string) []string {
	var ausdruecke []string

	rest := definition
	for {
		_, danach, gefunden := strings.Cut(rest, "CHECK")
		if !gefunden {
			return ausdruecke
		}

		start := strings.Index(danach, "(")
		if start < 0 {
			return ausdruecke
		}

		tiefe, ende := 0, start
		for ende < len(danach) {
			switch danach[ende] {
			case '(':
				tiefe++
			case ')':
				tiefe--
			}

			if tiefe == 0 {
				break
			}

			ende++
		}

		ausdruecke = append(ausdruecke, danach[start:min(ende+1, len(danach))])
		rest = danach[min(ende+1, len(danach)):]
	}
}
