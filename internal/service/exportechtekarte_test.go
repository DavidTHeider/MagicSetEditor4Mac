//go:build fremdpakete

// **Derselbe Lauf mit einer echten Karte** —
// #609.
//
// # Wofür
//
// `TestExportvorlagenAusFremdpaketen` fährt eine Probekarte mit fünf gefüllten
// Feldern. Das Werkstück verlangt die Gegenprobe: **kommt dieselbe Zahl
// heraus, wenn die Karte aus einem kanonischen Set stammt?** Ist sie kleiner,
// misst der Lauf seine Vorrichtung; ist sie gleich, misst er die Ausfuhr.
//
//	MSE_FREMDPAKETE="…/M15-Magic-Pack/data" \
//	  MSE_SETPAKET="…/kanonische-sets/Vanguard.mse-set" \
//	  go test -tags fremdpakete ./internal/service/ -run EchterKarte -v
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestExportvorlagenMitEchterKarte(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	satz := os.Getenv("MSE_SETPAKET")

	if wurzel == "" || satz == "" {
		t.Skip("MSE_FREMDPAKETE und MSE_SETPAKET setzen")
	}

	karten, setwerte := ausEchtemSet(t, satz)

	t.Logf("die Karte trägt %d gefüllte Felder, das Set %d Angaben",
		len(karten[0].Values), len(setwerte))

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("Wurzel lesen: %v", err)
	}

	gelesen, gelaufen, leer, ausser := 0, 0, 0, 0
	gruende := map[string]int{}

	for _, eintrag := range eintraege {
		if msepaket.ArtAusName(eintrag.Name()) != msepaket.ArtExportvorlage {
			continue
		}

		vorlage, da := vorlageAusPaket(t, wurzel, eintrag.Name())
		if !da || vorlage.Skript == "" {
			continue
		}

		gelesen++

		ausgabe, _, lauffehler := ausfuhrprobe{
			ausfuehrer: service.NeuerVorlagenausfuehrer().
				MitSymbolen(echteSymbolquelle(wurzel),
					render.NeuerRasterer(msepaket.NeueBildquelle(wurzel))),
			mittel: service.Ausfuhrmittel{
				Bilder:            leerschreiber{},
				Bildziel:          t.TempDir(),
				Beigaben:          leerbeigaben{},
				Symbolzeichnungen: msepaket.NeueBildquelle(wurzel),
				Setpaket:          satz,
				Verzeichnis:       "ausfuhr-files",
			},
		}.Fuehre(t.Context(), vorlage, karten, setwerte, nil)
		if lauffehler != nil {
			meldung := lauffehler.Error()
			if stelle := strings.LastIndex(meldung, ": "); stelle > 0 {
				meldung = meldung[stelle+2:]
			}

			gruende[meldung]++

			// **Und hier wird er rot** — bis zum 2026-08-22 nicht.
			//
			// Der Lauf zählte, meldete und endete grün, **auch wenn jede
			// einzelne Vorlage scheiterte**. Als Messung war das zulässig; als
			// Beleg dafür, dass die Ausfuhr mit einer echten Karte trägt, war
			// es keiner — die Zahl stand im Protokoll, und niemand hielt sie
			// gegen etwas. Gefunden im Zweitblick über Phase L.
			//
			// Die Ausnahmen sind dieselben wie beim Schwesterlauf und aus
			// demselben Grund: `card_regions` gibt es an dieser MSE2-Fassung
			// nicht ([0036](../../docs/entscheidungen/0036-card-regions-ist-ausser-reichweite.md)).
			if grund, jenseits := ausserReichweite[eintrag.Name()]; jenseits {
				ausser++

				t.Logf("  ⊘ %s: außer Reichweite (%s, Vermerk 0036)",
					eintrag.Name(), grund)
			} else {
				t.Errorf("%s: %s", eintrag.Name(), meldung)
			}

			continue
		}

		if ausgabe != "" {
			gelaufen++

			continue
		}

		// **Eine leere Ausgabe ohne Fehler ist nicht dasselbe wie ein
		// Fehlschlag — und sie verschwand hier lautlos.**
		//
		// Sie zählte in `gelesen` und nicht in `gelaufen`, und niemand nannte
		// sie. „17 von 19" las sich damit als „zwei kaputt", während eines
		// außer Reichweite war und eines **richtig** nichts ausgab:
		// `magic-namecheck` gibt „a text list of cards that share names with
		// real Magic cards" — für eine unauffällige Karte ist leer die
		// richtige Antwort.
		//
		// **Unterscheiden kann der Lauf die beiden nicht**: ob leer richtig
		// ist, weiß nur die Vorlage. Er nennt sie deshalb, statt sie
		// wegzurechnen — die Zahl bleibt, der Name kommt dazu.
		leer++

		if eintrag.Name() != stilleVorlage {
			t.Errorf("%s: leere Ausgabe ohne Fehler", eintrag.Name())
		}

		t.Logf("  ⌀ %s: ohne Fehler, aber ohne Ausgabe", eintrag.Name())
	}

	t.Logf("%d Exportvorlagen gelesen, %d liefen durch, %d ohne Ausgabe, "+
		"%d außer Reichweite — mit einer Karte aus %s",
		gelesen, gelaufen, leer, ausser, filepath.Base(satz))

	for grund, zahl := range gruende {
		t.Logf("    %d × %s", zahl, grund)
	}

	// # Wie weit diese Zahl reicht
	//
	// **Über eine Paketwurzel, eine Karte und ein Set — mehr nicht** (§10).
	// Gemessen am 2026-08-22 mit einer Karte aus `Vanguard.mse-set`:
	//
	//	M15-Magic-Pack   19 gelesen, 18 durch, 1 außer Reichweite
	//	full-magic-pack  22 gelesen, 21 durch, 1 außer Reichweite
	//
	// **Beide Male dieselbe eine**, `magic-lackeyccg` an `card_regions`
	// ([0036](../../docs/entscheidungen/0036-card-regions-ist-ausser-reichweite.md)).
	// Dass die Zahl mit der Wurzel wächst und der Ausfall nicht, ist die
	// eigentliche Auskunft — sie sagt aber **nichts** darüber, ob eine
	// Vorlage das Richtige ausgibt: geprüft ist, dass sie **etwas** ausgibt
	// und nicht scheitert.
	//
	// **Und nichts über andere Karten.** Eine Vorlage, die auf ein Feld
	// zugreift, das diese eine Karte nicht führt, liefe hier durch und
	// anderswo nicht. Der Schwesterlauf mit der Probekarte deckt genau das
	// von der anderen Seite ab: er trägt fünf gefüllte Felder und den Rest
	// aus dem Spiel.
	//
	// **Und eine Zusicherung, dass überhaupt gelesen wurde.**
	//
	// Ohne sie ist eine Wurzel ohne eine einzige Exportvorlage der grünste
	// Lauf von allen: die Schleife läuft nullmal, keine der Zusicherungen
	// darin wird betreten, und am Ende steht „0 gelesen" im Protokoll — das
	// niemand liest, wenn der Test grün ist.
	if gelesen == 0 {
		t.Fatalf("keine einzige Exportvorlage unter %s", wurzel)
	}
}

// ausEchtemSet liest die erste Karte und die Setangaben eines `.mse-set`.
func ausEchtemSet(t *testing.T, ort string) ([]*domain.CardRow, []domain.Value) {
	t.Helper()

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("%s öffnen: %v", ort, err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // gelesen ist gelesen

	karten := paket.Kopf.Alle("card")
	if len(karten) == 0 {
		t.Fatalf("%s trägt keine Karte", ort)
	}

	return []*domain.CardRow{{
		ID: 1, Number: 1, Values: werteAusBlock(karten[0]),
	}}, werteAusBlock(paket.Kopf.Ersten("set_info"))
}

// werteAusBlock macht aus den Kindern eines Blocks Feldwerte.
func werteAusBlock(knoten *msepaket.Knoten) []domain.Value {
	if knoten == nil {
		return nil
	}

	var werte []domain.Value

	for _, kind := range knoten.Kinder {
		if kind.Zeilenart != msepaket.ArtEintrag {
			continue
		}

		text := kind.Wert
		if text == "" {
			text = kind.Skripttext()
		}

		werte = append(werte, domain.Value{
			Field: msepaket.Entkanonisch(kind.Schluessel), Text: text,
		})
	}

	return werte
}

// vorlageAusPaket liest eine Exportvorlage samt Spielfeldern.
func vorlageAusPaket(t *testing.T, wurzel, name string) (msepaket.Exportvorlage, bool) {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, name))
	if err != nil {
		return msepaket.Exportvorlage{}, false
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel))

	vorlage := paket.Exportmodell()

	spielpaket, fehler := msepaket.Oeffne(filepath.Join(wurzel, vorlage.Spiel+".mse-game"))
	if fehler != nil {
		return vorlage, true
	}

	defer func() { _ = spielpaket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	spielpaket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel))

	modell, _ := spielpaket.Spielmodell()
	vorlage.Kartenfelder = modell.Fields
	vorlage.Setfelder = modell.SetFields
	vorlage.Spielskript = spielpaket.Initskript()

	// **Und die Fassungen des Setsymbols** (#660) — dieselbe Auskunft, die
	// `Vorlagenliste.spielangaben` im Betrieb holt. Ohne sie gäbe es
	// `symbol_variation` nicht, und der Lauf spräche über eine Vorrichtung,
	// die ärmer ist als der Betrieb (§3).
	vorlage.Symbolfassungen = spielpaket.Spielsymbolfassungen()

	return vorlage, true
}
