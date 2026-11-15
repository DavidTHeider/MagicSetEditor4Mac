package app

// **Wo Kennung und Geheimnis herkommen** — Werkstück 4 von Phase 4.12
// (#1351).
//
// **Paketintern**, weil die Quelle es ist: von außen ließe sich nur
// beobachten, dass eine Anmeldung scheitert — hier steht, **woran**.

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne eingetragene Kennung ist es ein Nein mit Begründung.**
//
// **Und die Begründung ist der Punkt** (§6): „für github ist keine Kennung
// eingetragen" sagt, was zu tun ist. „Die Anmeldung ist gescheitert" sagt es
// nicht.
func TestOhneEingetrageneKennungIstEsEinNeinMitBegruendung(t *testing.T) {
	db := probedatenbank(t)

	quelle := zugangsquelleUeber(db, probebund(t))

	_, fehler := quelle(probekontext(t), "github")
	if fehler == nil {
		t.Fatal("die fehlende Kennung kam durch")
	}

	if !strings.Contains(fehler.Error(), "registriert seine eigene Anwendung") {
		t.Errorf("die Meldung sagt nicht, was zu tun ist: %v", fehler)
	}
}

// **Eine eingetragene Kennung kommt an — und ein fehlendes Geheimnis hält
// nichts auf.**
//
// **Weil es je Anbieter verschieden ist**: Google nennt es optional, Microsoft
// verbietet es, GitHub verlangt es. Diese Stelle soll die Entscheidung nicht
// vorwegnehmen — `anmeldung.Beginne` trifft sie.
func TestEineEingetrageneKennungKommtAn(t *testing.T) {
	db := probedatenbank(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		AnbieterKennungsschluessel("github"), "meine-kennung"); fehler != nil {
		t.Fatalf("die Kennung eintragen: %v", fehler)
	}

	zugang, fehler := zugangsquelleUeber(db, probebund(t))(probekontext(t), "github")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if zugang.Kennung != "meine-kennung" {
		t.Errorf("Kennung %q", zugang.Kennung)
	}

	if zugang.Geheimnis != "" {
		t.Errorf("Geheimnis %q, erwartet war leer", zugang.Geheimnis)
	}
}

// **Ein Bestand, der nicht antwortet, reicht seinen Fehler durch.**
func TestEinStummerBestandReichtSeinenFehlerDurch(t *testing.T) {
	db := probedatenbank(t)
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := zugangsquelleUeber(db, probebund(t))(
		probekontext(t), "github"); fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// probebund gibt einen Bund, den es nicht gibt.
//
// **Und das ist der sichere Weg** (`internal/schluesselbund/kein_dialog_test.go`):
// ein Bund an einem Ort, der nicht existiert, wird **vor** dem Aufruf von
// `security` erkannt — der echte Anmeldebund bleibt unangetastet, und es
// erscheint kein Systemdialog.
func probebund(t *testing.T) *schluesselbund.Bund {
	t.Helper()

	return schluesselbund.NeuerBund(filepath.Join(t.TempDir(), "gibtsnicht.keychain"))
}

// probedatenbank gibt eine migrierte Datenbank in einem eigenen Verzeichnis.
//
// **Paketintern noch einmal**, weil `offeneProbedatenbank` zu `package
// app_test` gehört — dieselbe Lage wie bei `probekontext` daneben.
func probedatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, err := storage.Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // ein Lauf schließt selbst

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db
}

// **Und die Naht liest denselben Schlüssel, den die Domäne führt.**
//
// # Die Zusage, die gefehlt hat
//
// **Zwei Stellen haben ihn gebaut, und sie waren sich nicht einig.**
// `app.AnbieterKennungsschluessel` rechnete ihn aus, `domain` kannte ihn
// nicht — und wies ihn in `PruefeEinstellung` ab. Der Lauf darüber setzte den
// Wert **an der Prüfung vorbei** direkt in den Bestand und war deshalb grün,
// während im Betrieb niemand die Kennung eintragen konnte.
//
// **Diese Zusicherung schließt den Kreis** (§10: zwei grüne Seiten einer Naht
// sind kein Beleg für die Naht): derselbe Schlüssel, den die Naht liest, muss
// durch die Prüfung gehen.
func TestDerGelesenSchluesselKommtDurchDiePruefung(t *testing.T) {
	t.Parallel()

	for _, anbieter := range domain.Anbieter() {
		schluessel := AnbieterKennungsschluessel(anbieter)

		if fehler := domain.PruefeEinstellung(schluessel, "meine-kennung"); fehler != nil {
			t.Errorf("%s: %v", schluessel, fehler)
		}
	}
}
