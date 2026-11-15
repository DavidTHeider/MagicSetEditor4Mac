package app_test

// **Die Wurzel der Rücksetzung geht wirklich durch den Bestand** — Werkstück
// O4 von Phase 5.2
// (#1428).
//
// **Über eine echte Datenbank**, denn die Frage ist, ob die Naht verdrahtet
// ist (§10, funktionale Abdeckung) — ein Doppel beantwortete sie nicht.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Layoutrücksetzung nimmt genau ihre vier Schlüssel zurück** — und
// lässt die Sprache stehen.
func TestDieRuecksetzungswurzelNimmtNurDasLayoutZurueck(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)
	ctx := probekontext(t)

	for _, schluessel := range []string{
		domain.EinstellungErscheinungsbild, domain.EinstellungSprache,
	} {
		if _, fehler := db.ExecContext(ctx,
			`INSERT INTO settings (profil_id, schluessel, wert) VALUES (1, ?, 'dunkel')`,
			schluessel); fehler != nil {
			t.Fatalf("%s setzen: %v", schluessel, fehler)
		}
	}

	vorschau := anwendung.Ruecksetzungsvorschau(ctx, "layout", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	if !vorschau.Bereit || vorschau.SicherungEmpfohlen || vorschau.NachweisNoetig {
		t.Errorf("die Ansagen sind %+v", vorschau)
	}

	// **Fünf seit UX-14.AK-01** (#1686): der Fensterstart gehört zum Layout —
	// wer „das Fenster sieht falsch aus" sagt, meint auch seine Größe. **Sechs
	// seit UX-14.AK-02** mit der Lage der Leiste, **zehn seit UX-09** (#1681)
	// mit Farbwelt, Reichweite der Schrift, Fensterbild und eigenen Farben.
	if len(vorschau.Betroffen) != 1 || vorschau.Betroffen[0].Anzahl != 10 {
		t.Fatalf("die Vorschau nennt %+v", vorschau.Betroffen)
	}

	bilanz := anwendung.Zuruecksetzen(ctx, "layout", "", "", "", "", "")
	if bilanz.Fehler != nil {
		t.Fatalf("zurücksetzen: %+v", bilanz.Fehler)
	}

	if bilanz.Zurueckgesetzt != 1 {
		t.Errorf("zurückgesetzt wurde %d", bilanz.Zurueckgesetzt)
	}

	// **Die Sprache steht noch da** — was nicht in der Liste steht, wird
	// nicht angefasst.
	var uebrig int
	if fehler := db.QueryRowContext(ctx,
		`SELECT count(*) FROM settings WHERE profil_id = 1`).Scan(&uebrig); fehler != nil {
		t.Fatalf("nachzählen: %v", fehler)
	}

	if uebrig != 1 {
		t.Errorf("es stehen %d Einstellungen da", uebrig)
	}
}

// **Und das Profil zählt, was ihm gehört** — samt der Ansage, dass hier
// Daten fortgehen.
func TestDieRuecksetzungswurzelZaehltDasEigeneProfil(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)
	ctx := probekontext(t)

	vorschau := anwendung.Ruecksetzungsvorschau(ctx, "profil", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	if !vorschau.SicherungEmpfohlen {
		t.Error("die Profillöschung empfiehlt keine Sicherung")
	}

	// **Ein fremdes Profil kommt nicht durch** — und zwar ohne Zahlen.
	fremd := anwendung.Ruecksetzungsvorschau(ctx, "profil", "4711", "")
	if fremd.Fehler != nil {
		t.Fatalf("die fremde Vorschau: %+v", fremd.Fehler)
	}

	if fremd.Bereit || len(fremd.Betroffen) != 0 {
		t.Errorf("das fremde Profil ist %+v", fremd)
	}

	if fremd.Hindernisse[0].Art != domain.RuecksetzungFremdesZiel {
		t.Errorf("das Hindernis heißt %q", fremd.Hindernisse[0].Art)
	}
}

// **Die ganze Anwendung verlangt einen Nachweis** — auch über diese Wurzel.
func TestDieRuecksetzungswurzelVerlangtDenNachweis(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)
	ctx := probekontext(t)

	vorschau := anwendung.Ruecksetzungsvorschau(ctx, "anwendung", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	if !vorschau.NachweisNoetig {
		t.Error("die Vorschau sagt nicht, dass ein Nachweis nötig ist")
	}

	bilanz := anwendung.Zuruecksetzen(ctx, "anwendung", "", "", "", "",
		vorschau.Revision)
	if bilanz.Fehler == nil {
		t.Fatal("ohne Nachweis wurde der Bestand beiseitegelegt")
	}

	if bilanz.Fehler.Code != app.CodeErlaubnis {
		t.Errorf("der Fehler ist %+v", bilanz.Fehler)
	}
}

// **Und mit Nachweis legt sie den Bestand wirklich beiseite** — der ganze Weg
// über eine echte Datenbank, eine echte Datei und einen echten Nachweis.
//
// **Das ist die Probe auf den Neuanfang** (§10): jeder Teil davon war
// einzeln geprüft, und die Frage ist, ob der Weg **durch** sie trägt.
func TestDieRuecksetzungswurzelLegtDenBestandBeiseite(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	ctx := probekontext(t)
	jetzt := time.Now()

	// **Der Nachweis wird gelegt und nicht behauptet**: die Ablage ist die,
	// die der Dienst liest, und ein Feld `bestaetigt` wäre keiner.
	if err := storage.NeueNachweisablage(db).NachweisAnlegen(ctx, domain.Nachweis{
		Kennung: "probe", NutzerID: 1, ProfilID: 1,
		Zweck: domain.RuecksetzungsZweck, Ziel: string(domain.RuecksetzungAnwendung),
		Ausgestellt: jetzt.Add(-time.Minute), Ablauf: jetzt.Add(time.Hour),
	}); err != nil {
		t.Fatalf("den Nachweis legen: %v", err)
	}

	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)

	// **Erst hinsehen, dann klicken** — seit Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2 verlangt jede löschende Art die Revision aus
	// ihrer Vorschau, und dieser Weg soll gehen, was eine Bedienung geht.
	vorschau := anwendung.Ruecksetzungsvorschau(ctx, "anwendung", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	bilanz := anwendung.Zuruecksetzen(ctx, "anwendung", "", "", "", "probe",
		vorschau.Revision)
	if bilanz.Fehler != nil {
		t.Fatalf("zurücksetzen: %+v", bilanz.Fehler)
	}

	if bilanz.Beiseite == "" {
		t.Fatal("die Bilanz sagt nicht, wo der alte Bestand liegt")
	}

	if _, err := os.Stat(bilanz.Beiseite); err != nil {
		t.Errorf("der alte Bestand liegt nicht bei %q: %v", bilanz.Beiseite, err)
	}

	// **Und der Nachweis ist verbraucht** — ein zweiter Ruf geht nicht.
	zweiter := anwendung.Zuruecksetzen(ctx, "anwendung", "", "", "", "probe", "")
	if zweiter.Fehler == nil {
		t.Error("derselbe Nachweis ging ein zweites Mal durch")
	}
}

// **Die Wahl eines Menschen geht ihren eigenen Weg** — über seine Kennung.
func TestDieRuecksetzungswurzelNimmtNutzereinstellungenZurueck(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)
	ctx := probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`INSERT INTO nutzereinstellungen (nutzer_id, schluessel, wert)
		 VALUES (1, ?, 'gross')`, domain.EinstellungSchriftgroesse); fehler != nil {
		t.Fatalf("die Wahl setzen: %v", fehler)
	}

	bilanz := anwendung.Zuruecksetzen(ctx, "nutzereinstellungen", "", "1", "", "", "")
	if bilanz.Fehler != nil {
		t.Fatalf("zurücksetzen: %+v", bilanz.Fehler)
	}

	if bilanz.Zurueckgesetzt != 1 {
		t.Errorf("zurückgesetzt wurde %d", bilanz.Zurueckgesetzt)
	}
}

// **Der letzte Mensch lässt sich nicht löschen** — und die Vorschau sagt es
// mit der Zahl aus dem Bestand.
func TestDieRuecksetzungswurzelHaeltDenLetztenMenschen(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	vorschau := app.RuecksetzungsAnwendungUeber(db, ort).
		Ruecksetzungsvorschau(probekontext(t), "nutzer", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	if vorschau.Bereit ||
		vorschau.Hindernisse[0].Art != domain.RuecksetzungLetzterNutzer {
		t.Errorf("die Vorschau ist %+v", vorschau)
	}
}
