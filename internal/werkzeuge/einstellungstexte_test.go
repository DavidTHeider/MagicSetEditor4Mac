package werkzeuge_test

// **Jeder Wert, den Go ausgeben kann, hat im Fenster einen Namen** —
// Werkstück 8 von Phase 4.8
// (#1321).
//
// # Warum es diesen Wächter gibt
//
// `Wahl` schlägt die Beschriftung eines Wertes im Katalog unter
// `<schluessel>.<wert>` nach. Fehlt der Eintrag, kommt `undefined` heraus, und
// React zeichnet eine Option **ohne Text** — ohne zu klagen.
//
// Genau das ist mit den vier Dienstschaltern geschehen: `dienst.scryfall.aus`
// steht in keinem Katalog, `Dienstschalter` reichte die vorhandene
// Eigenschaft `werteNamen` nicht weiter, und im Fenster standen **vier leere
// Auswahllisten**. Der Wert war richtig gesetzt; man konnte nur nicht sehen,
// ob ein Dienst an oder aus ist.
//
// **Kein Test hat es gesehen**, weil jeder das Feld über seine Beschriftung
// sucht und den **Wert** prüft. Der Text der Optionen kam in keiner
// Zusicherung vor. Gefunden hat es die Sichtprüfung am laufenden Fenster.
//
// # Und der Kommentar im Katalog behauptete ihn bereits
//
// Über `namen` stand „ein Wächter hält beide Seiten zusammen". Den gab es
// nicht. Das ist die Gestalt aus §10: ein Kopf, der mehr verspricht als sein
// Rumpf hält, ist schlimmer als gar keiner — niemand prüft nach.
//
// # Was er hält
//
// Für **jeden** Schlüssel aus `domain.Einstellungsreihe()` und **jeden**
// seiner erlaubten Werte: es gibt einen Namen, in beiden Sprachen. Wie ein
// Schlüssel zu seinen Namen kommt, sagt [namensweg]
// (`einstellungsnamenswege_test.go`) — eine Aufzählung, damit
// eine neue Einstellung die Entscheidung erzwingt statt still in einer Gruppe
// zu verschwinden.
//
// # Was er nicht hält
//
// **Ob der Name der richtige ist.** Er sieht, dass „Aus" dasteht, nicht dass
// es den Zustand trifft. Und er sieht den Katalog, nicht das Fenster: dass
// `Dienstschalter` die Namen auch **weiterreicht**, prüft
// `frontend/src/seiten/einstellungen.test.tsx`. Beide zusammen decken den
// Weg; einer allein deckt eine Hälfte.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJederEinstellungswertHatEinenNamen(t *testing.T) {
	t.Parallel()

	eigennamen := eigennamenAusDemFenster(t)

	for _, sprache := range []string{"de", "en"} {
		t.Run(sprache, func(t *testing.T) {
			t.Parallel()

			katalog := katalogtexte(t, sprache)

			for _, schluessel := range domain.Einstellungsreihe() {
				pfad, bekannt := namensweg(schluessel)
				if !bekannt {
					t.Errorf("die Einstellung %q ist in namensweg nicht "+
						"eingetragen — wie kommen ihre Werte zu einem Namen?",
						schluessel)

					continue
				}

				pruefeWerte(t, schluessel, pfad, katalog, eigennamen)
			}
		})
	}
}

// pruefeWerte sieht jeden erlaubten Wert eines Schlüssels an.
func pruefeWerte(
	t *testing.T, schluessel string, pfad weg,
	katalog map[string]bool, eigennamen map[string]bool,
) {
	t.Helper()

	if pfad == ausEigenerSeite {
		return
	}

	// **Der Schlüssel selbst braucht einen Namen, nicht nur seine Werte** —
	// #1540.
	//
	// **Hier stand er im Kommentar und nicht im Rumpf.** Über dem
	// Hochladepfad steht seit Phase 4.12: „Sein **Name** steht trotzdem im
	// Katalog; die Schleife unten läuft über null Werte und prüft damit nur
	// ihn." Sie prüfte ihn **nicht** — eine Schleife über null Werte läuft
	// nullmal, und ihr Rumpf ist das Einzige, was prüfte.
	//
	// Das ist die Gestalt aus §10, und zwar in genau dem Wächter, der gegen
	// sie geschrieben wurde: ein Kopf, der mehr verspricht als sein Rumpf
	// hält. Aufgefallen ist es an den drei Anbieteradressen, die im Fenster
	// ihren rohen Schlüssel trugen.
	if pfad == ausKatalog && !katalog[schluessel] {
		t.Errorf("im Katalog fehlt die Beschriftung %q — im Fenster stünde "+
			"der rohe Schlüssel", schluessel)
	}

	for _, wert := range domain.ErlaubteEinstellungswerte(schluessel) {
		switch pfad {
		case ausKatalog:
			if !katalog[schluessel+"."+wert] {
				t.Errorf("im Katalog fehlt der Name für %q — das Fenster "+
					"zeichnet dann eine Auswahl ohne Text",
					schluessel+"."+wert)
			}

		case ausSchalterpaar:
			// **Je Wert sein Name** — Zusatzbefund 3 des Nachlaufs über
			// Phase 4.8. Vorher prüfte dieser Zweig für **jeden** Wert
			// dasselbe: dass `dienstAus` und `dienstAn` im Katalog stehen.
			// Käme in `domain` ein dritter Zustand dazu, bliebe er grün —
			// während `Dienstschalter` nur für zwei einen Namen kennt.
			//
			// Der Kopf verspricht einen Namen für **jeden erlaubten Wert**;
			// jetzt hält der Rumpf das auch.
			name, kennt := schalternamen[wert]
			if !kennt {
				t.Errorf("%q erlaubt den Wert %q — dafür kennt "+
					"`Dienstschalter` keinen Namen; er reicht `werteNamen` "+
					"nur für »aus« und »an« weiter", schluessel, wert)

				continue
			}

			if !katalog[name] {
				t.Errorf("im Katalog fehlt %q — %q braucht es für seinen "+
					"Wert %q", name, schluessel, wert)
			}

			// **Und der Schalter selbst.** Seine Beschriftung kommt aus
			// `dienstnamen.ts`; `Dienstschalter` weicht sonst auf den rohen
			// Schlüssel aus, und am Schalter stünde `dienst.erfunden`.
			//
			// **Das ist der Fall, den die Gegenprobe grün ließ**: ein neuer
			// Dienst braucht keinen neuen Katalogeintrag für „aus" und „an" —
			// aber einen Namen braucht er.
			if kennung := strings.TrimPrefix(schluessel, "dienst."); !eigennamen[kennung] {
				t.Errorf("in dienstnamen.ts fehlt %q — am Schalter stünde "+
					"sonst der rohe Schlüssel %q", kennung, schluessel)
			}

		case ausEigennamen:
			if !eigennamen[wert] {
				t.Errorf("in dienstnamen.ts fehlt %q — %q listet ihn als "+
					"Wert, und das Fenster zeigte dann nichts",
					wert, schluessel)
			}

		case ausEigenerSeite:
		}
	}
}

func eigennamenAusDemFenster(t *testing.T) map[string]bool {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "bausteine", "dienstnamen.ts"))
	if err != nil {
		t.Fatalf("dienstnamen.ts lesen: %v", err)
	}

	block := regexp.MustCompile(
		`dienstnamen: Record<string, string> = \{([^}]*)\}`).
		FindSubmatch(roh)
	if block == nil {
		t.Fatal("in dienstnamen.ts steht kein Block `dienstnamen: Record<…> = {…}`")
	}

	gefunden := map[string]bool{}

	for _, zeile := range regexp.MustCompile(`(?m)^\s*([a-z][a-z0-9]*):`).
		FindAllStringSubmatch(string(block[1]), -1) {
		gefunden[zeile[1]] = true
	}

	return gefunden
}

// unterschluesselEinesDienstes sagt, ob ein Schlüssel **unter** einem Dienst
// hängt statt sein Schalter zu sein.
//
// **Gezählt und nicht aufgezählt**: `dienst.<name>` ist der Schalter, alles
// mit einem weiteren Punkt darunter etwas anderes — heute die Adresse, morgen
// vielleicht mehr. Eine Liste hier ließe den nächsten wieder durch, und genau
// so ist #1540
// entstanden.
func unterschluesselEinesDienstes(schluessel string) bool {
	rest, istDienst := strings.CutPrefix(schluessel, "dienst.")

	return istDienst && strings.Contains(rest, ".")
}
