package service_test

// **Was hinausgeht** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass eine Nachricht vollständig und unterschrieben im Ausgang landet.**
// `T5` hat `Handelspost.Senden` gebaut und niemand hat es gerufen; der
// Umschlag entstand in keinem Weg. Gemessen wird deshalb der Umschlag selbst:
// seine Art folgt dem Zustand, seine Folge zählt hoch, und seine Unterschrift
// prüft gegen die eigene Kennung.
//
// **Und dass ein Entwurf nicht hinausgeht** — die Gegenprobe. Aus drei
// Zuständen geht keine Nachricht hinaus (`domain.NachrichtenartFuer`); ohne
// diesen Fall sagte das Grün oben nur, dass irgendetwas eingereiht wird.

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// folgenprobe zählt die Folge im Arbeitsspeicher.
type folgenprobe struct {
	stand map[string]int64
	klagt error
}

func (f *folgenprobe) NaechsteFolge(
	_ context.Context, vorgang string,
) (int64, error) {
	if f.klagt != nil {
		return 0, f.klagt
	}

	if f.stand == nil {
		f.stand = map[string]int64{}
	}

	f.stand[vorgang]++

	return f.stand[vorgang], nil
}

// dieVersandzeit ist die Uhr dieses Laufs.
var dieVersandzeit = time.Date(2026, 9, 19, 8, 0, 0, 0, time.UTC)

// einVersand stellt ihn über einem Stand auf.
func einVersand(t *testing.T, stand domain.Handelsangebot) (
	*service.Handelsversand, *ausgangsprobe, peer.Identitaet,
) {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	ausgang := &ausgangsprobe{}
	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: stand},
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusgang(ausgang)

	versand := service.NeuerHandelsversand(post, standprobe{stand: stand},
		&folgenprobe{}, ausweisprobe{ich: ich}).
		MitUhr(func() time.Time { return dieVersandzeit }).
		MitKennungen(func() string { return "n-fest" })

	return versand, ausgang, ich
}

// **Ein angebotener Stand geht als Vorschlag hinaus — unterschrieben.**
func TestEinAngebotenerStandGehtAlsVorschlagHinaus(t *testing.T) {
	t.Parallel()

	stand := derHeutigeStand()
	versand, ausgang, ich := einVersand(t, stand)

	nachrichtID, fehler := versand.Schicken(
		context.Background(), stand.Vorgang, dasGegenueber)
	if fehler != nil {
		t.Fatalf("Schicken: %v", fehler)
	}

	if len(ausgang.zeilen) != 1 {
		t.Fatalf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
	}

	zeile := ausgang.zeilen[0]
	umschlag := zeile.Nachricht.Umschlag

	if nachrichtID != umschlag.NachrichtID {
		t.Errorf("zurück kam %q, im Umschlag steht %q",
			nachrichtID, umschlag.NachrichtID)
	}

	if umschlag.Art != domain.NachrichtVorschlag {
		t.Errorf("die Art ist %q", umschlag.Art)
	}

	if umschlag.Folge != 1 || zeile.Folge != 1 {
		t.Errorf("die Folge ist %d im Umschlag und %d an der Zeile",
			umschlag.Folge, zeile.Folge)
	}

	if umschlag.Absender != string(ich.Kennung()) {
		t.Errorf("der Absender ist %q", umschlag.Absender)
	}

	if umschlag.Abdruck != stand.Abdruck() {
		t.Errorf("der Abdruck gehört nicht zum Stand")
	}

	// **Und die Befunde des Umschlags sind leer** — er muss drüben durch
	// [domain.Handelsumschlag.Validate], und eine Nachricht, die dort
	// scheitert, ist hier schon falsch gebaut.
	if befunde := umschlag.Validate(); len(befunde) > 0 {
		t.Errorf("der Umschlag prüft nicht: %v", befunde[0])
	}

	unterschrift, fehler := hex.DecodeString(umschlag.Unterschrift)
	if fehler != nil {
		t.Fatalf("die Unterschrift ist kein Hex: %v", fehler)
	}

	if fehler := peer.Pruefe(ich.Kennung(),
		umschlag.Unterschriftsform(), unterschrift); fehler != nil {
		t.Errorf("die Unterschrift prüft nicht: %v", fehler)
	}
}

// **Ein Entwurf geht nicht hinaus.**
func TestEinEntwurfGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	stand := derHeutigeStand()
	stand.Zustand = domain.HandelEntwurf

	versand, ausgang, _ := einVersand(t, stand)

	if _, fehler := versand.Schicken(context.Background(),
		stand.Vorgang, dasGegenueber); fehler == nil {
		t.Fatal("der Entwurf ging hinaus")
	}

	if len(ausgang.zeilen) != 0 {
		t.Errorf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
	}
}

// **Die Wiederaufnahme nennt, was nachzuliefern ist.**
//
// **Der eigene Fortschritt wird gelesen und nicht vom Aufrufer genannt** —
// wer ihn mitschickte, nennte einen Stand, den er von irgendwoher hat.
func TestDieWiederaufnahmeNenntWasNachzuliefernIst(t *testing.T) {
	t.Parallel()

	stand := derHeutigeStand()
	versand, ausgang, _ := einVersand(t, stand)
	ctx := context.Background()

	if _, fehler := versand.Schicken(
		ctx, stand.Vorgang, dasGegenueber); fehler != nil {
		t.Fatalf("Schicken: %v", fehler)
	}

	// **Die Gegenseite hat noch nichts bestätigt** — dann geht alles ab der
	// ersten Folge noch einmal hinaus.
	offen, klaerung, fehler := versand.Wiederaufnehmen(ctx, stand.Vorgang,
		dasGegenueber, domain.Handelsfortschritt{Vorgang: stand.Vorgang})
	if fehler != nil {
		t.Fatalf("Wiederaufnehmen: %v", fehler)
	}

	if klaerung {
		t.Fatal("es klemmt, und das sollte es nicht")
	}

	if len(offen) != len(ausgang.zeilen) {
		t.Errorf("nachzuliefern sind %d von %d", len(offen), len(ausgang.zeilen))
	}
}

// **Jede Stelle, an der es klemmen kann, klemmt — und schickt nichts.**
//
// **Tabellengetrieben, weil es dieselbe Zusage fünfmal ist** (§3): der
// Ausgang bleibt leer, gleich woran es lag. Eine Nachricht, die nach einem
// halben Fehler dorthin fiele, ginge beim nächsten Anlauf hinaus — ohne dass
// jemand sie gewollt hätte.
func TestEinVersandOhneStandAusweisOderFolgeSchicktNichts(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		baue  func(*testing.T) (*service.Handelsversand, *ausgangsprobe)
		wobei string
	}{
		{
			name: "ohne Stand",
			baue: func(t *testing.T) (*service.Handelsversand, *ausgangsprobe) {
				t.Helper()

				return versandMit(t, standprobe{klagt: errProbe},
					&folgenprobe{}, false)
			},
		},
		{
			name: "ohne Ausweis",
			baue: func(t *testing.T) (*service.Handelsversand, *ausgangsprobe) {
				t.Helper()

				return versandMit(t, standprobe{stand: derHeutigeStand()},
					&folgenprobe{}, true)
			},
		},
		{
			name: "ohne Folge",
			baue: func(t *testing.T) (*service.Handelsversand, *ausgangsprobe) {
				t.Helper()

				return versandMit(t, standprobe{stand: derHeutigeStand()},
					&folgenprobe{klagt: errProbe}, false)
			},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			versand, ausgang := fall.baue(t)

			if _, fehler := versand.Schicken(context.Background(),
				derPostvorgang, dasGegenueber); fehler == nil {
				t.Fatal("es ging hinaus")
			}

			if len(ausgang.zeilen) != 0 {
				t.Errorf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
			}
		})
	}
}

// **Und ein Ausgang, der klagt, klagt auch beim Versand durch.**
func TestEinKlagenderAusgangKlagtBeimVersandDurch(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	stand := derHeutigeStand()
	ausgang := &ausgangsprobe{klagt: errProbe}
	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: stand},
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusgang(ausgang)

	versand := service.NeuerHandelsversand(post, standprobe{stand: stand},
		&folgenprobe{}, ausweisprobe{ich: ich})

	if _, klage := versand.Schicken(context.Background(),
		stand.Vorgang, dasGegenueber); klage == nil {
		t.Error("der klagende Ausgang kam durch")
	}
}

// **Ohne Stand sagt auch die Wiederaufnahme nichts.**
func TestOhneStandNenntDieWiederaufnahmeNichts(t *testing.T) {
	t.Parallel()

	versand, _ := versandMit(t, standprobe{klagt: errProbe}, &folgenprobe{},
		false)

	if _, _, fehler := versand.Wiederaufnehmen(context.Background(),
		derPostvorgang, dasGegenueber,
		domain.Handelsfortschritt{Vorgang: derPostvorgang}); fehler == nil {
		t.Error("ohne Stand kam eine Wiederaufnahme")
	}
}

// **Eine Bestätigung ohne Kennung endet vor der Post.**
//
// **Die Gegenprobe steht daneben**: mit Kennung geht sie durch, und der
// Ausgang ist danach leer. Ohne sie sagte das Nein oben nichts darüber, ob
// dieser Weg überhaupt einen hat.
func TestEineBestaetigungOhneKennungEndetVorDerPost(t *testing.T) {
	t.Parallel()

	stand := derHeutigeStand()
	versand, ausgang, _ := einVersand(t, stand)
	ctx := context.Background()

	if fehler := versand.Bestaetigt(ctx, ""); fehler == nil {
		t.Fatal("eine Bestätigung ohne Kennung kam durch")
	}

	nachrichtID, fehler := versand.Schicken(ctx, stand.Vorgang, dasGegenueber)
	if fehler != nil {
		t.Fatalf("Schicken: %v", fehler)
	}

	if fehler := versand.Bestaetigt(ctx, nachrichtID); fehler != nil {
		t.Fatalf("Bestätigen: %v", fehler)
	}

	if len(ausgang.zeilen) != 0 {
		t.Errorf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
	}
}

// versandMit stellt einen Versand über gewählten Doppeln auf.
func versandMit(
	t *testing.T, stand standprobe, folgen *folgenprobe, ohneAusweis bool,
) (*service.Handelsversand, *ausgangsprobe) {
	t.Helper()

	ausweis := ausweisprobe{fehler: errProbe}

	if !ohneAusweis {
		ich, fehler := peer.Erzeugen(nil)
		if fehler != nil {
			t.Fatalf("Erzeugen: %v", fehler)
		}

		ausweis = ausweisprobe{ich: ich}
	}

	ausgang := &ausgangsprobe{}
	post := service.NeueHandelspost(&briefkastenprobe{}, stand,
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusgang(ausgang)

	return service.NeuerHandelsversand(post, stand, folgen, ausweis), ausgang
}
