package domain_test

// **Was ein Tausch im Journal hinterlässt** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Dass ein empfangenes Stück den Erwerbspreis des Senders nicht erbt.** Was
// jemand anderes bezahlt hat, ist nicht, was es mich gekostet hat — und eine
// Zahl, die so entstünde, sähe später aus wie eine erfasste.
//
// **Dass ein Handelsbeleg nichts Inneres trägt.** Er geht hinaus; Bestandskennungen
// und Erwerbsdaten gehen es niemanden an.
//
// **Und dass die Vergangenheit nicht gelöscht wird.** Nach einem Vollzug gibt
// es nur eine Gegenbewegung mit Bezug — kein Wegnehmen.

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einZugang baut eine Journalzeile über ein empfangenes Stück.
func einZugang(anders func(*domain.Journalzeile)) domain.Journalzeile {
	zeile := domain.Journalzeile{
		Art:      domain.JournalZugang,
		Anlass:   domain.AnlassTausch,
		Vorgang:  "v-1",
		Printing: "a1b2c3d4",
		Menge:    2,
		Zustand:  domain.ZustandPlayed,
		Sprache:  "de",
		Finish:   domain.FinishEtched,
		ID:       4711,
		Gebucht:  time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC),
	}

	if anders != nil {
		anders(&zeile)
	}

	return zeile
}

// **Ein empfangenes Stück kommt ohne Kostenbasis.**
func TestEinZugangAusEinemTauschHatKeinenErwerbspreis(t *testing.T) {
	t.Parallel()

	zugang := einZugang(nil)

	exemplar := zugang.AlsExemplar()

	if exemplar.Erwerbspreis != 0 || exemplar.Erwerbswaehrung != "" {
		t.Errorf("das Stück kam mit %d %q",
			exemplar.Erwerbspreis, exemplar.Erwerbswaehrung)
	}

	// **Und es trägt seine Herkunft** — wer später fragt, woher es kam,
	// bekommt eine Antwort statt einer Vermutung.
	if exemplar.Herkunft.Art != domain.HerkunftEinfuhr {
		t.Errorf("die Herkunft ist %q", exemplar.Herkunft.Art)
	}

	if exemplar.Menge != 2 || exemplar.Zustand != domain.ZustandPlayed {
		t.Errorf("das Stück ist %+v", exemplar)
	}
}

// **Ein Handelsbeleg trägt nichts Inneres.**
func TestEinBelegTraegtNichtsInneres(t *testing.T) {
	t.Parallel()

	zeilen := []domain.Journalzeile{
		einZugang(nil),
		einZugang(func(z *domain.Journalzeile) {
			z.Art, z.Printing, z.Menge = domain.JournalAbgang, "b2c3d4e5", 1
		}),
	}

	beleg := domain.Handelsbeleg{
		Vorgang:    "v-1",
		Abdruck:    "abdruck-der-revision",
		Parteien:   []string{"bich", "bdu"},
		Zeilen:     zeilen,
		Gebucht:    einZugang(nil).Gebucht,
		Beidseitig: false,
	}

	text := string(beleg.Unterschriftsform())

	// **Was drinstehen muss.**
	for _, gehoertRein := range []string{"v-1", "abdruck-der-revision", "a1b2c3d4"} {
		if !strings.Contains(text, gehoertRein) {
			t.Errorf("der Handelsbeleg nennt %q nicht: %s", gehoertRein, text)
		}
	}

	// **Und was nicht.** Eine Bestandskennung, ein Lagerort, eine Notiz oder
	// ein Erwerbspreis haben in einem Handelsbeleg nichts zu suchen — er geht
	// hinaus.
	mitAllem := einZugang(func(z *domain.Journalzeile) {
		z.ExemplarID, z.Lagerort, z.Notiz = 4711, "Kiste 4", "aus dem Tausch"
	})

	beleg.Zeilen = []domain.Journalzeile{mitAllem}

	mitText := string(beleg.Unterschriftsform())

	for _, gehoertNichtRein := range []string{"4711", "Kiste 4", "aus dem Tausch"} {
		if strings.Contains(mitText, gehoertNichtRein) {
			t.Errorf("der Handelsbeleg nennt %q: %s", gehoertNichtRein, mitText)
		}
	}
}

// **Ein einseitiger Handelsbeleg sagt, dass er einer ist.**
//
// **Zwei Unterschriften vorzutäuschen wäre das Schlimmste**: der lokale
// Abschluss ist echt, der beidseitige noch nicht — und wer beides gleich
// aussehen ließe, machte aus einer offenen Sache eine erledigte.
func TestEinEinseitigerBelegSagtEsUndSiehtAndersAus(t *testing.T) {
	t.Parallel()

	einseitig := domain.Handelsbeleg{
		Vorgang: "v-1", Abdruck: "abdruck", Parteien: []string{"bich", "bdu"},
		Zeilen: []domain.Journalzeile{einZugang(nil)},
	}

	beidseitig := einseitig
	beidseitig.Beidseitig = true

	if bytes.Equal(einseitig.Unterschriftsform(), beidseitig.Unterschriftsform()) {
		t.Error("beide Formen sind gleich")
	}
}

// **Die Vergangenheit wird nicht gelöscht.**
func TestEineGegenbewegungNenntIhrenBezug(t *testing.T) {
	t.Parallel()

	gebucht := einZugang(nil)

	gegen, fehler := gebucht.Gegenbewegung("ein Irrtum")
	if fehler != nil {
		t.Fatalf("Gegenbewegung: %v", fehler)
	}

	// **Sie dreht die Richtung um** und nennt, worauf sie sich bezieht.
	if gegen.Art != domain.JournalAbgang {
		t.Errorf("die Gegenbewegung ist ein %q", gegen.Art)
	}

	if gegen.Anlass != domain.AnlassKorrektur {
		t.Errorf("der Anlass ist %q", gegen.Anlass)
	}

	if gegen.Grund != "ein Irrtum" {
		t.Errorf("der Grund ist %q", gegen.Grund)
	}

	if gegen.Menge != gebucht.Menge || gegen.Printing != gebucht.Printing {
		t.Errorf("die Gegenbewegung ist %+v", gegen)
	}

	// **Der Bezug ist die Zeile und nicht das Exemplar.** Das kann längst
	// fort sein — gerade die Buchung, die es weggegeben hat, ist die, die man
	// später berichtigt.
	if gegen.Bezug != gebucht.ID {
		t.Errorf("der Bezug ist %d statt %d", gegen.Bezug, gebucht.ID)
	}

	// **Und sie ist selbst noch keine Zeile im Journal.**
	if gegen.ID != 0 {
		t.Errorf("die Gegenbewegung trägt schon die Kennung %d", gegen.ID)
	}

	// **Und andersherum ebenso**: die Gegenbewegung zu einem Abgang ist ein
	// Zugang — sonst ließe sich ein irrtümlich gebuchter Abgang nicht
	// zurückholen.
	abgang := einZugang(func(z *domain.Journalzeile) { z.Art = domain.JournalAbgang })

	zurueck, fehler := abgang.Gegenbewegung("doch nicht weggegeben")
	if fehler != nil {
		t.Fatalf("Gegenbewegung: %v", fehler)
	}

	if zurueck.Art != domain.JournalZugang {
		t.Errorf("die Gegenbewegung zum Abgang ist ein %q", zurueck.Art)
	}

	// **Ohne Grund gibt es keine Korrektur.** Eine Buchung, die niemand
	// erklärt, ist im Journal schlimmer als keine.
	if _, fehler := gebucht.Gegenbewegung("  "); fehler == nil {
		t.Error("die grundlose Korrektur kam durch")
	}

	// **Und ohne geschriebene Zeile ebenso wenig.** Eine Korrektur auf eine
	// Buchung, die es nicht gibt, ist eine zweite Buchung.
	ungeschrieben := einZugang(func(z *domain.Journalzeile) { z.ID = 0 })

	if _, fehler := ungeschrieben.Gegenbewegung("ein Irrtum"); fehler == nil {
		t.Error("die bezuglose Korrektur kam durch")
	}
}

// **Die Anlässe sind eine Aufzählung** — was journalfähig ist, steht fest.
func TestDieAnlaesseStehenFest(t *testing.T) {
	t.Parallel()

	anlaesse := domain.Journalanlaesse()

	if len(anlaesse) != 6 {
		t.Fatalf("es sind %d Anlässe: %v", len(anlaesse), anlaesse)
	}

	for _, anlass := range anlaesse {
		if !domain.IstJournalanlass(anlass) {
			t.Errorf("%q gilt nicht als Anlass", anlass)
		}
	}

	if domain.IstJournalanlass("erfunden") {
		t.Error("ein erfundener Anlass ging durch")
	}
}

// **Eine Zeile, die nicht stimmt, kommt nicht ins Journal.** Ein Journal ist
// die eine Stelle, an der nichts nachträglich zu berichtigen ist — deshalb
// wird hier geprüft und nicht nach dem Schreiben.
func TestEineUnstimmigeJournalzeileWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gut := domain.Journalzeile{
		Art: domain.JournalZugang, Anlass: domain.AnlassKauf,
		Printing: "bolt-m10", Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil,
	}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("die gute Zeile wurde abgewiesen: %v", fehler)
	}

	faelle := map[string]func(z *domain.Journalzeile){
		"ohne Richtung":    func(z *domain.Journalzeile) { z.Art = "" },
		"fremde Richtung":  func(z *domain.Journalzeile) { z.Art = "seitwärts" },
		"ohne Anlass":      func(z *domain.Journalzeile) { z.Anlass = "" },
		"fremder Anlass":   func(z *domain.Journalzeile) { z.Anlass = "fundsache" },
		"ohne Druck":       func(z *domain.Journalzeile) { z.Printing = "" },
		"Menge null":       func(z *domain.Journalzeile) { z.Menge = 0 },
		"Menge unter null": func(z *domain.Journalzeile) { z.Menge = -1 },
		// **Eine Korrektur ohne Grund und ohne Bezug ist keine** — sie sieht
		// aus wie eine Buchung und ist eine Behauptung.
		"Korrektur ohne Grund": func(z *domain.Journalzeile) {
			z.Anlass, z.Bezug = domain.AnlassKorrektur, 17
		},
		"Korrektur ohne Bezug": func(z *domain.Journalzeile) {
			z.Anlass, z.Grund = domain.AnlassKorrektur, "verzählt"
		},
		// **Und ein Tausch ohne Vorgang ebenso** — er gehört zu einem Handel,
		// und welchem, muss dastehen.
		"Tausch ohne Vorgang": func(z *domain.Journalzeile) {
			z.Anlass = domain.AnlassTausch
		},
	}

	for name, verderben := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := gut
			verderben(&zeile)

			if fehler := zeile.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("die Zeile kam mit %v durch", fehler)
			}
		})
	}
}

// **Ein Abschluss sagt, ob er überhaupt einer ist.**
//
// Geprüft wird, **bevor** eine Transaktion beginnt: was ohne die Datenbank zu
// entscheiden ist, hält sonst eine offene Transaktion auf.
func TestEinUnstimmigerAbschlussWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "a",
		Parteien: []string{"a", "b"}, Zustand: domain.HandelLokalVerbucht,
	}

	gut := domain.Handelsabschluss{
		Angebot: angebot,
		Beleg: domain.Handelsbeleg{
			Vorgang: "v-1", Abdruck: angebot.Abdruck(),
			Zeilen: []domain.Journalzeile{{
				Art: domain.JournalAbgang, Anlass: domain.AnlassTausch,
				Vorgang: "v-1", Printing: "bolt-m10", Menge: 2,
				Zustand: domain.ZustandNearMint, Sprache: "de",
				Finish: domain.FinishNonfoil, ExemplarID: 1,
			}},
		},
		Unterschrift: "u",
	}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("der gute Abschluss wurde abgewiesen: %v", fehler)
	}

	faelle := map[string]struct {
		verderben func(*domain.Handelsabschluss)
		erwartet  error
	}{
		"fremder Vorgang": {
			func(a *domain.Handelsabschluss) { a.Beleg.Vorgang = "v-anders" },
			domain.ErrValidation,
		},
		// **Der Abdruck wird nachgerechnet** — er ist die Zusage, dass
		// unterschrieben wurde, was gebucht wird.
		"anderer Stand": {
			func(a *domain.Handelsabschluss) { a.Beleg.Abdruck = "alt" },
			domain.ErrConflict,
		},
		"ohne Unterschrift": {
			func(a *domain.Handelsabschluss) { a.Unterschrift = "" },
			domain.ErrValidation,
		},
		// **Ein unverbuchter Zustand ist kein Abschluss** — er ist das, was
		// davor kommt.
		"unverbuchter Zustand": {
			func(a *domain.Handelsabschluss) {
				a.Angebot.Zustand = domain.HandelUebergabeOffen
			},
			domain.ErrConflict,
		},
		"ohne Bewegung": {
			func(a *domain.Handelsabschluss) { a.Beleg.Zeilen = nil },
			domain.ErrValidation,
		},
		"Abgang ohne Besitzzeile": {
			func(a *domain.Handelsabschluss) { a.Beleg.Zeilen[0].ExemplarID = 0 },
			domain.ErrValidation,
		},
		"unstimmige Zeile": {
			func(a *domain.Handelsabschluss) { a.Beleg.Zeilen[0].Menge = 0 },
			domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			abschluss := gut
			abschluss.Beleg.Zeilen = slices.Clone(gut.Beleg.Zeilen)
			fall.verderben(&abschluss)

			if fehler := abschluss.Pruefe(); !errors.Is(fehler, fall.erwartet) {
				t.Errorf("der Abschluss gab %v statt %v", fehler, fall.erwartet)
			}
		})
	}
}
