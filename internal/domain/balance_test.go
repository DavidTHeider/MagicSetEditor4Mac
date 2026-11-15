package domain_test

// **Was eine Balanceanalyse nennen muss** — Werkstück A13.2b von Phase 5.3
// (#1452).
//
// # Der Satz, um den es geht
//
// > Balanceanalyse nennt die nach nachvollziehbarer Such-/Filterstrategie
// > ausgewählten Vergleichskarten und **konkrete Unterschiede** … **kein
// > objektiver Balancewert aus Ähnlichkeitsscore**.
//
// **Die zweite Hälfte ist die schwierigere**, und sie wird nicht geprüft,
// sondern **gebaut**: es gibt kein Feld für eine Zahl. Was nicht da ist, kann
// niemand füllen — und ein Wächter darüber hielte eine Zusage, die die
// Gestalt schon hält.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiVergleichsbelege gibt zwei Karten, wie ein Kontextpaket sie vergibt.
func zweiVergleichsbelege() []domain.Beleg {
	return []domain.Beleg{
		{
			Kennung: "B1", Quelle: domain.QuelleOracle,
			Titel: "Funkenwurf", Ausschnitt: "{R}: 2 Schaden an eine Kreatur.",
		},
		{
			Kennung: "B2", Quelle: domain.QuelleOracle,
			Titel: "Blockhilfe", Ausschnitt: "{G}: +2/+2 bis zum Ende des Zuges.",
		},
	}
}

// eineBalanceantwort baut eine vollständige.
func eineBalanceantwort() domain.Balanceangaben {
	return domain.Balanceangaben{
		Vergleiche: []domain.Vergleichskarte{
			{
				Beleg: "B1", Name: "Funkenwurf",
				Unterschiede: []string{"ein Mana billiger", "trifft nur Kreaturen"},
			},
		},
		NichtBetrachtet: []string{"Zusammenspiel mit Manaflut"},
	}
}

// **Eine vollständige Balanceanalyse ist belegt.**
func TestEineVollstaendigeBalanceanalyseIstBelegt(t *testing.T) {
	t.Parallel()

	geprueft := domain.PruefeBalance(
		"Sie kostet eins mehr als [B1].", eineBalanceantwort(),
		zweiVergleichsbelege())

	if !geprueft.Belegt() {
		t.Errorf("nicht belegt: %+v", geprueft.Befunde)
	}

	if geprueft.Art != domain.EinschaetzungBalance {
		t.Errorf("Art %q", geprueft.Art)
	}
}

// **Eine Vergleichskarte, deren Beleg es nicht gibt, ist ein Befund.**
func TestEineVergleichskarteOhneBelegIstEinBefund(t *testing.T) {
	t.Parallel()

	angaben := eineBalanceantwort()
	angaben.Vergleiche[0].Beleg = "B7"

	geprueft := domain.PruefeBalance("Sie kostet eins mehr als [B1].",
		angaben, zweiVergleichsbelege())

	if geprueft.Belegt() {
		t.Fatal("die erfundene Kennung kam durch")
	}

	if len(geprueft.Befunde) != 1 ||
		geprueft.Befunde[0].Art != domain.BefundUnbekannteKennung {
		t.Errorf("Befunde %+v", geprueft.Befunde)
	}
}

// **Ein Vergleich ohne konkreten Unterschied ist keiner.**
//
// **Das ist die Zusage des Issues in ihrer schärfsten Form**: „nennt … und
// **konkrete Unterschiede**". Ein Vergleich, der nur einen Namen nennt, sagt
// „ähnlich" — und das ist genau die Aussage, die ein Ähnlichkeitsrang schon
// macht und die nichts bedeutet.
func TestEinVergleichOhneUnterschiedIstEinBefund(t *testing.T) {
	t.Parallel()

	angaben := eineBalanceantwort()
	angaben.Vergleiche[0].Unterschiede = nil

	geprueft := domain.PruefeBalance("Sie ähnelt [B1].", angaben,
		zweiVergleichsbelege())

	if geprueft.Belegt() {
		t.Fatal("der leere Vergleich kam durch")
	}

	if len(geprueft.Befunde) != 1 ||
		geprueft.Befunde[0].Art != domain.BefundOhneUnterschied {
		t.Errorf("Befunde %+v", geprueft.Befunde)
	}
}

// **Ohne einen einzigen Vergleich ist es keine Balanceanalyse.**
func TestOhneVergleichIstEsKeineBalanceanalyse(t *testing.T) {
	t.Parallel()

	geprueft := domain.PruefeBalance("Sie ist stark.",
		domain.Balanceangaben{}, zweiVergleichsbelege())

	if geprueft.Belegt() {
		t.Fatal("die Analyse ohne Vergleich galt als belegt")
	}

	arten := make([]string, 0, len(geprueft.Befunde))
	for _, befund := range geprueft.Befunde {
		arten = append(arten, befund.Art)
	}

	if !strings.Contains(strings.Join(arten, " "), domain.BefundOhneVergleich) {
		t.Errorf("Befunde %+v", geprueft.Befunde)
	}
}

// **Und es gibt keinen Balancewert.**
//
// **Geprüft an der Gestalt und nicht an einem Wert**: `Balanceangaben` hat
// kein Zahlenfeld, und `Vergleichskarte` auch nicht. Wer einen Rang in eine
// Bewertung verwandeln wollte, müsste zuerst ein Feld dafür schaffen — und
// dann steht diese Zusage im Weg, nicht nur ein Satz darüber.
func TestEineBalanceanalyseHatKeinenWert(t *testing.T) {
	t.Parallel()

	for _, feld := range domain.Balancefelder() {
		if strings.Contains(strings.ToLower(feld), "wert") ||
			strings.Contains(strings.ToLower(feld), "punkt") ||
			strings.Contains(strings.ToLower(feld), "rang") ||
			strings.Contains(strings.ToLower(feld), "score") {
			t.Errorf("es gibt ein Feld %q — ein Ähnlichkeitsrang ist keine "+
				"Bewertung (A15, 0103)", feld)
		}
	}
}
