package app_test

// **Die Rücksetzung für den Dienst tauscht nicht** — G09
// (#2273).
//
// Über eine echte Datenbank und einen echten Nachweis: die Vorschau nennt
// das Hindernis, die Ausführung hält an, und danach ist alles wie vorher —
// der Nachweis unverbraucht und die Verbindung offen. **Beiseitelegen kann
// sie gar nicht**: ihr fehlt der Weg dafür, und das ist die Zusage.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieRuecksetzungFuerDenDienstTauschtNicht(t *testing.T) {
	db := frischerBestand(t)
	ctx := probekontext(t)
	jetzt := time.Now()

	if fehler := storage.NeueNachweisablage(db).NachweisAnlegen(ctx, domain.Nachweis{
		Kennung: "probe", NutzerID: 1, ProfilID: 1,
		Zweck: domain.RuecksetzungsZweck, Ziel: string(domain.RuecksetzungAnwendung),
		Ausgestellt: jetzt.Add(-time.Minute), Ablauf: jetzt.Add(time.Hour),
	}); fehler != nil {
		t.Fatalf("den Nachweis legen: %v", fehler)
	}

	anwendung := app.RuecksetzungsAnwendungOhneTausch(db)

	vorschau := anwendung.Ruecksetzungsvorschau(ctx, "anwendung", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	if vorschau.Bereit || len(vorschau.Hindernisse) != 1 ||
		vorschau.Hindernisse[0].Art != domain.RuecksetzungOhneTausch {
		t.Errorf("die Vorschau nennt %+v", vorschau.Hindernisse)
	}

	bilanz := anwendung.Zuruecksetzen(ctx, "anwendung", "", "", "", "probe", vorschau.Revision)
	if bilanz.Fehler == nil || bilanz.Fehler.Code != app.CodeKonflikt {
		t.Fatalf("die Ausführung: %+v — erwartet ein Konflikt", bilanz.Fehler)
	}

	if bilanz.Beiseite != "" {
		t.Errorf("die Bilanz nennt einen beiseitegelegten Bestand: %q", bilanz.Beiseite)
	}

	// **Die Verbindung ist noch offen, und der Nachweis unverbraucht** — sie
	// gibt ihn her.
	nachweis, fehler := storage.NeueNachweisablage(db).Nachweis(ctx, "probe")
	if fehler != nil {
		t.Fatalf("den Nachweis lesen: %v", fehler)
	}

	if nachweis.Eingeloest {
		t.Error("der Nachweis ist verbraucht, obwohl nichts geschah")
	}

	// **Die übrigen Arten gehen** — dieselbe Wurzel wie bisher.
	if layout := anwendung.Zuruecksetzen(ctx, "layout", "", "", "", "", ""); layout.Fehler != nil {
		t.Errorf("das Layout: %+v", layout.Fehler)
	}
}
