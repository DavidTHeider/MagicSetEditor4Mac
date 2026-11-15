package service

// **Die Rückkehr und was sie hinterlässt** — Werkstück O7 von Phase 5.2
// (#1431).
//
// **Paketintern**, weil die Vorrichtung hier liegt: ein Bestandsweg über einer
// Datei, die keine Datenbank ist. Was ein Bestand hergibt, entscheidet die
// Naht — und genau die wird hier ausgetauscht.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probesicherung schreibt eine echte Sicherung neben den Bestand.
//
// **Einmal je Test.** `Sichern` überschreibt nichts; wer zweimal ruft, bekommt
// den Konflikt und nicht die Datei.
func probesicherung(t *testing.T, bau *wegbau) string {
	t.Helper()

	ziel := filepath.Join(bau.ort, "gut.msesicherung")
	if _, fehler := os.Stat(ziel); fehler == nil {
		return ziel
	}

	return gesichert(t, bau.weg, ziel)
}

// probezeit ist der Zeitpunkt, den die Uhr dieser Läufe meldet.
var probezeit = time.Date(2026, 9, 10, 14, 0, 0, 0, time.UTC)

// mitStaenden setzt die zwei Leser auf feste Antworten.
func mitStaenden(weg *Bestandsweg, buch Rueckkehrbuch, sicherung, jetzt domain.Bestandsstand) {
	weg.MitRuecksprungschau(
		func(context.Context, string) (domain.Bestandsstand, error) { return sicherung, nil },
		func(context.Context) domain.Bestandsstand { return jetzt },
		buch,
	)

	weg.MitUhr(func() time.Time { return probezeit })
}

// besitz baut einen Stand mit einer Sammlung.
func besitz(exemplare int, mengen ...domain.Sammlungsmenge) domain.Bestandsstand {
	return domain.Bestandsstand{
		AngelegtAm: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Sets:       1, Karten: 1, Decks: 0, Exemplare: exemplare,
		Sammlung: mengen, SammlungGelesen: true,
	}
}

// **Was die Sicherung mehr hat, steht nach der Rückkehr als unklar im Buch.**
//
// Das ist die Zusage des Werkstücks: ein historisches Backup kann bereits
// abgegebene Karten enthalten, und nach der Rückkehr weiß niemand, ob sie noch
// jemandem gehören.
func TestEineRueckkehrHinterlaesstIhreUnklarenMengen(t *testing.T) {
	bau := neuerWegbau(t)

	mitStaenden(bau.weg, bau.buch,
		besitz(4, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 4, Name: "Bolt"}),
		besitz(1, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1}),
	)

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau)); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	if len(bau.buch.epochen) != 1 {
		t.Fatalf("es stehen %d Epochen im Buch", len(bau.buch.epochen))
	}

	epoche := bau.buch.epochen[0]

	if !epoche.Zurueckgespielt.Equal(probezeit) {
		t.Errorf("die Rückkehr steht auf %v, erwartet war %v", epoche.Zurueckgespielt, probezeit)
	}

	if !epoche.Datenstand.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("der Datenstand ist %v", epoche.Datenstand)
	}

	if epoche.Quelle != "gut.msesicherung" {
		t.Errorf("die Quelle ist %q — erwartet war der Dateiname ohne Pfad", epoche.Quelle)
	}

	if len(epoche.Mengen) != 1 || epoche.Mengen[0].Menge != 3 {
		t.Errorf("unklar sind %+v, erwartet waren drei von scryfall:abc", epoche.Mengen)
	}
}

// **Eine Rückkehr, die nichts unklar lässt, hinterlässt keine Epoche.**
//
// Sonst stünde nach jedem Zurückspielen eine Sperre da, die jemand quittieren
// muss, damit sie verschwindet — und eine Sperre, die immer da ist,
// unterscheidet die Fälle nicht mehr (§10).
func TestEineRueckkehrOhneUnklarheitHinterlaesstNichts(t *testing.T) {
	bau := neuerWegbau(t)

	gleich := besitz(2, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 2})
	mitStaenden(bau.weg, bau.buch, gleich, gleich)

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau)); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	if len(bau.buch.epochen) != 0 {
		t.Errorf("es steht eine Epoche im Buch: %+v", bau.buch.epochen)
	}
}

// **Kann niemand nachsehen, entsteht die Epoche trotzdem** — mit ihrem Grund
// und ohne Mengen. Das Issue verlangt es: bei fehlenden Nachweisen bleibt die
// Unsicherheit sichtbar.
func TestOhneAuskunftEntstehtDieEpocheTrotzdem(t *testing.T) {
	bau := neuerWegbau(t)

	stumm := besitz(2)
	stumm.SammlungGelesen = false

	mitStaenden(bau.weg, bau.buch, stumm, besitz(2))

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau)); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	if len(bau.buch.epochen) != 1 {
		t.Fatalf("es stehen %d Epochen im Buch", len(bau.buch.epochen))
	}

	if len(bau.buch.epochen[0].Mengen) != 0 {
		t.Errorf("die Epoche nennt Mengen, obwohl niemand nachsehen konnte: %+v",
			bau.buch.epochen[0].Mengen)
	}
}

// **Ein Buch, das nicht schreibt, hält den Tausch nicht auf** (§7, Regel 3) —
// **und der Ort des alten Bestands kommt trotzdem zurück**.
//
// Getauscht ist getauscht. Wer hier den Ort verschwiege, nähme dem Aufrufer
// den einzigen Weg zurück — wegen eines Vermerks.
func TestEinBuchDasNichtSchreibtHaeltDenTauschNichtAuf(t *testing.T) {
	bau := neuerWegbau(t)

	bau.buch.fehler = errors.New("das Buch ist zu")
	mitStaenden(bau.weg, bau.buch,
		besitz(4, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 4}),
		besitz(1, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1}),
	)

	beiseite, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau))
	if fehler == nil {
		t.Fatal("der gescheiterte Vermerk wurde verschwiegen")
	}

	if beiseite == "" {
		t.Error("der Ort des alten Bestands kam nicht zurück")
	}

	// **Und der Satz sagt, was zu tun ist.**
	if !strings.Contains(fehler.Error(), "von Hand") {
		t.Errorf("der Satz nennt den Handgriff nicht: %v", fehler)
	}

	// **Der Tausch hat stattgefunden.**
	inhalt, lesefehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db"))
	if lesefehler != nil {
		t.Fatalf("den Bestand lesen: %v", lesefehler)
	}

	if string(inhalt) == "laufender bestand" {
		t.Error("der Bestand ist der alte geblieben")
	}
}

// **Ohne Buch wird gar nicht erst zurückgespielt.**
func TestOhneBuchWirdNichtZurueckgespielt(t *testing.T) {
	bau := neuerWegbau(t)

	bau.weg.MitRuecksprungschau(nil, nil, nil)

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau)); fehler == nil {
		t.Error("ein Weg ohne Buch spielte zurück")
	}
}

// **Die Lage lässt sich ansehen, ohne etwas zu ändern.**
func TestDieLageLaesstSichAnsehenOhneEtwasZuAendern(t *testing.T) {
	bau := neuerWegbau(t)

	mitStaenden(bau.weg, bau.buch,
		besitz(4, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 4}),
		besitz(1, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1}),
	)

	lage, fehler := bau.weg.Ruecksprunglage(t.Context(), probesicherung(t, bau))
	if fehler != nil {
		t.Fatalf("die Lage lesen: %v", fehler)
	}

	if len(lage.Unklar) != 1 {
		t.Errorf("die Lage nennt %d unklare Mengen", len(lage.Unklar))
	}

	if len(bau.buch.epochen) != 0 {
		t.Error("das Ansehen hat eine Epoche geschrieben")
	}

	inhalt, lesefehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db"))
	if lesefehler != nil || string(inhalt) != "laufender bestand" {
		t.Errorf("das Ansehen hat den Bestand angefasst: %q, %v", inhalt, lesefehler)
	}
}

// **Ein Leser, der scheitert, kippt die Lage — statt eine leere zu geben.**
//
// **Den Tausch hält er nicht auf, aber er verschwindet auch nicht**: die
// Rückkehr ist der Zweck, der Vergleich das Sicherheitsnetz. Wer deswegen
// abbräche, ließe jemanden mit einem kaputten Bestand sitzen, weil eine
// Auskunft darüber nicht zu bekommen war — und wer ihn stillschweigend
// überginge, gäbe „nichts unklar" aus, wo nichts gelesen wurde.
func TestEinScheiternderLeserKipptDieLage(t *testing.T) {
	bau := neuerWegbau(t)

	bau.weg.MitRuecksprungschau(
		func(context.Context, string) (domain.Bestandsstand, error) {
			return domain.Bestandsstand{}, errors.New("die Datei ist zu")
		},
		nil, bau.buch,
	)

	if _, fehler := bau.weg.Ruecksprunglage(t.Context(), probesicherung(t, bau)); fehler == nil {
		t.Error("der gescheiterte Leser kam durch")
	}

	// **Und derselbe Fehler hält das Zurückspielen nicht an** — er macht
	// daraus eine Handprüfung.
	if _, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau)); fehler != nil {
		t.Fatalf("das Zurückspielen brach am Vergleich ab: %v", fehler)
	}

	inhalt, lesefehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db"))
	if lesefehler != nil || string(inhalt) == "laufender bestand" {
		t.Errorf("der Bestand wurde nicht getauscht: %q, %v", inhalt, lesefehler)
	}

	if len(bau.buch.epochen) != 1 {
		t.Fatalf("es stehen %d Epochen im Buch", len(bau.buch.epochen))
	}

	if len(bau.buch.epochen[0].Mengen) != 0 {
		t.Errorf("die Epoche nennt Mengen, obwohl der Leser scheiterte: %+v",
			bau.buch.epochen[0].Mengen)
	}
}

// **Eine Datei, die es nicht gibt, meldet sich schon beim Auspacken.**
func TestEineLageUeberEineFehlendeDateiMeldetSich(t *testing.T) {
	bau := neuerWegbau(t)

	if _, fehler := bau.weg.Ruecksprunglage(
		t.Context(), filepath.Join(bau.ort, "gibtesnicht.msesicherung")); fehler == nil {
		t.Error("eine fehlende Datei kam durch")
	}
}

// **Quittieren nimmt die Sperre und bucht nichts.**
func TestQuittierenNimmtDieSperreUndBuchtNichts(t *testing.T) {
	bau := neuerWegbau(t)

	mitStaenden(bau.weg, bau.buch,
		besitz(4, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 4}),
		besitz(1, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1}),
	)

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), probesicherung(t, bau)); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	sperre, sperrfehler := bau.weg.Handelssperre(t.Context())
	if sperrfehler != nil {
		t.Fatalf("die Sperre lesen: %v", sperrfehler)
	}

	if !sperre.Offen() || sperre.Menge() != 3 {
		t.Errorf("die Sperre ist %v/%d, erwartet waren offen und drei", sperre.Offen(), sperre.Menge())
	}

	offene, lesefehler := bau.weg.OffeneRueckkehren(t.Context())
	if lesefehler != nil || len(offene) != 1 {
		t.Fatalf("die offenen Rückkehren sind %+v, %v", offene, lesefehler)
	}

	if fehler := bau.weg.RueckkehrQuittieren(t.Context(), offene[0].ID); fehler != nil {
		t.Fatalf("quittieren: %v", fehler)
	}

	nachher, nachfehler := bau.weg.Handelssperre(t.Context())
	if nachfehler != nil {
		t.Fatalf("die Sperre lesen: %v", nachfehler)
	}

	if nachher.Offen() {
		t.Error("die Sperre steht nach der Quittung weiter")
	}

	// **Und die Epoche ist noch da** — der Nachweis verschwindet nicht.
	alle, allefehler := bau.weg.Rueckkehren(t.Context())
	if allefehler != nil || len(alle) != 1 {
		t.Fatalf("im Buch stehen %+v, %v", alle, allefehler)
	}

	if !alle[0].GeklaertAm.Equal(probezeit) {
		t.Errorf("die Quittung steht auf %v, erwartet war %v", alle[0].GeklaertAm, probezeit)
	}
}

// **Ein Buch, das nicht liest, macht keine leere Sperre.** Eine leere Sperre
// hieße „alles frei", und das ist die eine Auskunft, die niemals falsch sein
// darf.
func TestEinBuchDasNichtLiestMachtKeineLeereSperre(t *testing.T) {
	bau := neuerWegbau(t)

	bau.buch.fehler = errors.New("das Buch ist zu")
	bau.weg.MitRuecksprungschau(nil, nil, bau.buch)

	if _, fehler := bau.weg.Handelssperre(t.Context()); fehler == nil {
		t.Error("das zugeschlagene Buch gab eine Sperre")
	}
}
