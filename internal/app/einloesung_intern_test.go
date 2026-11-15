package app

// **Eine Einladung genügt, über eine echte Leitung** — NET-01b
// (#1954).
//
// # Was hier gemessen wird
//
// **Zwei Anwendungen, ein Code**: Anna horcht auf der Rückschleife und stellt
// aus, Bert nimmt an — und danach kennt Anna Bert, unbestätigt, mit
// demselben Vergleichscode. Dazwischen liegen TCP, TLS 1.3 festgenagelt auf
// Annas Kennung und der Handschlag; was nicht dazwischen liegt, steht in
// `peertcp_intern_test.go` (§10).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// annaHorcht startet Annas Posten und trägt ein, wo sie erreichbar ist.
func annaHorcht(t *testing.T, anna *App) {
	t.Helper()

	gestartet := anna.HorchStarten(riegelkontext(t), "127.0.0.1:0")
	if gestartet.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", gestartet.Fehler)
	}

	t.Cleanup(anna.horchenBeenden)

	// **Die Rückschleife gilt nicht als erreichbar** — ein Lauf trägt sie
	// deshalb von Hand ein, wie `TestEinladungTraegtDieAdresse`.
	anna.peerlage.horchen.erreichbar.Adressen = []string{gestartet.Stand.Adresse}
}

func TestEineEinladungGenuegt(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	ctx := riegelkontext(t)

	annaHorcht(t, anna)

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "partie")
	if ausgestellt.Fehler != nil {
		t.Fatalf("Erstellen: %+v", ausgestellt.Fehler)
	}

	if vorschau := bert.KontakteinladungPruefen(ctx, ausgestellt.Einladung.Code); !vorschau.Vorschau.Einloesbar {
		t.Errorf("die Vorschau sagt nicht, dass sie einlösbar ist: %+v", vorschau)
	}

	angenommen := bert.KontakteinladungAnnehmen(ctx, ausgestellt.Einladung.Code, "", "", false, "Bert")
	if angenommen.Fehler != nil || angenommen.Gemeldet == nil || angenommen.Gemeldet.Stand != "gemeldet" {
		t.Fatalf("Annehmen: %+v, gemeldet %+v", angenommen, angenommen.Gemeldet)
	}

	// **Anna kennt Bert jetzt** — unbestätigt, ohne Recht, mit dem Namen,
	// den er sich gab, und mit demselben Code, den Bert sieht.
	annas := anna.Peerkontakte(ctx)
	if annas.Fehler != nil || len(annas.Kontakte) != 1 {
		t.Fatalf("Annas Kontakte: %+v", annas)
	}

	bertBeiAnna := annas.Kontakte[0]
	if bertBeiAnna.Lage != "unbestaetigt" || bertBeiAnna.SelbstgewaehlterName != "Bert" ||
		len(bertBeiAnna.Rechte) != 0 ||
		bertBeiAnna.Vergleichscode != angenommen.Kontakte[0].Vergleichscode {
		t.Errorf("Bert bei Anna: %+v, Annas Code bei Bert %q",
			bertBeiAnna, angenommen.Kontakte[0].Vergleichscode)
	}

	// **Bestätigt wird weiterhin nur über den Code** (AK-01).
	if bestaetigt := anna.PeerkontaktBestaetigen(
		ctx, bertBeiAnna.Kennung, bertBeiAnna.Vergleichscode); bestaetigt.Fehler != nil {
		t.Errorf("Bestaetigen: %+v", bestaetigt.Fehler)
	}
}

// **Wer nicht horcht, wird nicht erreicht — und die Annahme bleibt** (§7,
// Regel 3). Der Satz nennt den Rückfall.
func TestOhneHorchenBleibtDieAnnahme(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	ctx := riegelkontext(t)

	annaHorcht(t, anna)

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "handel")
	anna.horchenBeenden()

	angenommen := bert.KontakteinladungAnnehmen(ctx, ausgestellt.Einladung.Code, "", "", false, "")
	if angenommen.Fehler != nil || len(angenommen.Kontakte) != 1 {
		t.Fatalf("Annehmen: %+v", angenommen)
	}

	gemeldet := angenommen.Gemeldet
	if gemeldet == nil || gemeldet.Stand != "gescheitert" ||
		gemeldet.Grund != string(domain.EinloesungUnbeantwortet) || gemeldet.Meldung == "" {
		t.Errorf("gemeldet: %+v", gemeldet)
	}
}

// **Ohne Adresse gibt es nichts einzulösen** — die Antwort sagt es, statt
// es zu versuchen.
func TestOhneAdresseWirdNichtsVersucht(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	ctx := riegelkontext(t)

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "handel")

	angenommen := bert.KontakteinladungAnnehmen(ctx, ausgestellt.Einladung.Code, "", "", false, "")
	if angenommen.Gemeldet == nil || angenommen.Gemeldet.Stand != "ohne" {
		t.Errorf("gemeldet: %+v", angenommen.Gemeldet)
	}

	// **Und eine gescheiterte Annahme meldet gar nichts.**
	if wieder := bert.KontakteinladungAnnehmen(
		ctx, ausgestellt.Einladung.Code, "", "", false, ""); wieder.Fehler == nil || wieder.Gemeldet != nil {
		t.Errorf("zweimal: %+v", wieder)
	}
}

// **Die Sitzung schließt, auch wenn nichts einzulösen war** — und was dabei
// klemmt, bleibt im Protokoll statt den Posten anzuhalten.
func TestEineEinloesesitzungSchliesstImmer(t *testing.T) {
	t.Parallel()

	anna := eineEinladungsseite(t)
	sitzung := eineAngenommene(t, domain.ZweckPaarung)
	sitzung.lesefehler = errors.New("aufgelegt")
	sitzung.schliessfehler = errors.New("klemmt")

	anna.sitzungAufnehmen(riegelkontext(t), sitzung)

	if !sitzung.istZu() {
		t.Error("die Einlösesitzung blieb offen")
	}
}
