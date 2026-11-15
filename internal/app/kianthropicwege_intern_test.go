package app

// **Die Fehlerwege der Anthropic-Naht** — Werkstück A8 von Phase 5.3
// (#1447),
// abgetrennt mit A9
// (#1448),
// weil `kianthropicquelle_intern_test.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, was die Naht **tut**, wenn alles steht; hier, was sie tut, wenn
// etwas fehlt — ein unlesbarer Stand, ein stummer Schlüsselbund, eine
// Verdrahtung, die nichts baut.

import (
	"context"
	"errors"
	"testing"
)

// **Ein unlesbarer Stand hält den Ruf auf — an beiden Lesezugriffen.**
//
// # Warum beide gefahren werden
//
// **Weil `pruefeAnthropic` zweimal liest**: einmal für die Ortsregel, einmal
// für den Schalter. Ein Doppel, das beim ersten scheitert, erreicht den
// zweiten nie — und wer nur eines prüfte, ließe den anderen Zweig
// unbetreten (§10).
//
// **Und „im Zweifel durch" gibt es hier nicht**: was über »nur lokal« nicht
// feststeht, ist kein Grund, hinauszugehen.
func TestEinUnlesbarerStandHaeltDenAnthropicrufAuf(t *testing.T) {
	t.Parallel()

	server := eineAnthropicgegenstelle(t, "end_turn")

	for name, gutBis := range map[string]int{
		"die Ortsregel ist unlesbar": 0,
		"der Schalter ist unlesbar":  1,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := NeueApp("").MitMarktbund(erfundenerBund{})
			anwendung.MitEinstellungen(&halbstolpernde{gutBis: gutBis})

			quelle := anwendung.anthropicnaht(server.URL)

			if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe",
				"x", nil); fehler == nil {
				t.Fatal("der Ruf ging trotz unlesbarer Einstellungen hinaus")
			}
		})
	}
}

// **Ein Schlüsselbund, der nicht antwortet, gibt keinen Schlüssel.**
//
// **Und keinen Fehler**: das Blatt weist die Anfrage von selbst ab, bevor
// sie eine wird — mit dem Satz, der sagt, was fehlt. Ein zweiter Fehlerweg
// daneben sagte dasselbe zweimal.
func TestEinStummerSchluesselbundGibtKeinenSchluessel(t *testing.T) {
	t.Parallel()

	anwendung := KIAnwendungUeber(anthropicAn(t)).MitMarktbund(stummerBund{})

	if schluessel := anwendung.anthropicschluessel(""); schluessel != "" {
		t.Errorf("der Schlüssel ist %q", schluessel)
	}
}

// stummerBund antwortet auf jede Frage mit einem Fehler.
type stummerBund struct{}

func (stummerBund) Lies(context.Context, string, string) (string, error) {
	return "", errors.New("der Schlüsselbund antwortet nicht")
}

// **Die Verdrahtung baut jeden Anbieter.**
//
// **Ein Bauplan, der nie gerufen wird, ist keiner** (§10): die Verdrahtung
// setzt Hüllen, und ohne diesen Lauf stünde nur da, dass sie gesetzt wurden —
// nicht, dass sie etwas bauen.
func TestDieVerdrahtungBautJedenAnbieter(t *testing.T) {
	t.Parallel()

	anwendung := KIAnwendungUeber(anthropicAn(t)).MitMarktbund(erfundenerBund{})

	anbieter, fehler := anwendung.kianbieter(riegelkontext(t))
	if fehler != nil {
		t.Fatalf("kianbieter: %v", fehler)
	}

	// **Die lokale Laufzeit, Apples Gerätemodell und jeder Cloudanbieter**
	// — gezählt gegen [cloudanbieter] und nicht gegen eine Zahl von Hand
	// (A10, #1449; AI-05, #2042).
	if len(anbieter) != len(cloudanbieter())+2 {
		t.Fatalf("es sind %d Anbieter, die Liste führt %d Cloudanbieter",
			len(anbieter), len(cloudanbieter()))
	}

	for _, einer := range anbieter {
		if einer.bauen == nil {
			t.Fatalf("%s hat keinen Bauplan", einer.name)
		}

		if einer.bauen() == nil {
			t.Errorf("der Bauplan von %s baut nichts", einer.name)
		}
	}
}

// **Auch `kiVerdrahten` baut jeden Anbieter.**
//
// # Warum das nicht dasselbe ist wie der Lauf darüber
//
// **Zwei Composition Roots, zwei Verdrahtungen** (§2): `KIAnwendungUeber`
// baut die für das Fenster, `kiVerdrahten` die für den Bestand, der ohne
// Fenster aufgeht. Beide setzen eine Hülle um [App.anthropicnaht], und eine
// von beiden zu prüfen sagte über die andere nichts.
//
// **Und gezählt wird gegen [cloudanbieter]** (A10, #1449). Eine Aufzählung
// von Hand daneben wäre grün, sobald jemand einen Anbieter in die Liste
// schreibt und hier vergisst — dann stünde er am Fenster und würde nie
// gebaut.
func TestAuchKiVerdrahtenBautJedenAnbieter(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("").MitMarktbund(erfundenerBund{})
	anwendung.kiVerdrahten()

	if anwendung.kilaufzeit == nil {
		t.Fatal("die Verdrahtung setzt keine lokale Laufzeit")
	}

	if len(anwendung.kicloud) != len(cloudanbieter()) {
		t.Fatalf("verdrahtet sind %d Cloudanbieter, die Liste führt %d",
			len(anwendung.kicloud), len(cloudanbieter()))
	}

	for _, cloud := range cloudanbieter() {
		bauen := anwendung.kicloud[cloud.name]
		if bauen == nil {
			t.Errorf("für %s ist kein Bauplan verdrahtet", cloud.name)

			continue
		}

		// **Mit der Vorgabeadresse** (A18, #1533): leer heißt „die des
		// Anbieters", und der Bauplan reicht sie durch.
		if bauen("") == nil {
			t.Errorf("der Bauplan für %s baut nichts", cloud.name)
		}
	}
}

// **Ein unlesbarer Stand hält auch `kieinstellungen` auf.**
//
// **Und „im Zweifel an" gibt es dort nicht**: was über den Schalter nicht
// feststeht, ist kein Grund, ihn für eingeschaltet zu halten.
func TestEinUnlesbarerStandHaeltDieKieinstellungenAuf(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&halbstolpernde{gutBis: 0})

	an, adresse, fehler := anwendung.kieinstellungen(riegelkontext(t))
	if fehler == nil {
		t.Fatal("der Lesefehler kam nicht durch")
	}

	if an || adresse != "" {
		t.Errorf("der Stand ist an=%v adresse=%q", an, adresse)
	}
}

// **Ein Nahtfehler kommt auch beim Übernehmen durch.**
//
// **Und nicht erst beim Schreiben**: `kiwerkOhneQuelle` braucht keine
// Quelle, wohl aber ein Werk — und wenn dessen Bau schon gescheitert ist,
// ist die Auskunft darüber der Fehler und nicht ein leeres Werk.
func TestEinNahtfehlerKommtBeimUebernehmenDurch(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.kinahtfehler = errors.New("die Naht ist nicht gebaut")

	if _, fehler := anwendung.kiwerkOhneQuelle(riegelkontext(t)); fehler == nil {
		t.Fatal("der Nahtfehler kam nicht durch")
	}
}
