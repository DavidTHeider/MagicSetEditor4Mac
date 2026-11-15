package app

// **Ein Nachladefehler ist kein gescheiterter Schritt** — §7, Regel 3, an der
// Naht (#1960).
//
// # Was hier gemessen wird
//
// **Jeder Weg durch [App.amKontakt], an einem echten Bestand.** Eine Zeile
// mit krummem Zeitpunkt lässt sich schreiben und nicht lesen: jede Änderung
// an einem **anderen** Kontakt glückt, und das Nachladen der Liste danach
// scheitert an ihr. Bis #1960 kam das als `fehler` zurück — das Fenster hielt
// die Änderung für gescheitert, und eine Wiederholung scheiterte am
// geschriebenen Stand („schon angenommen").

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// krummeKontaktzeile legt einen Kontakt an, dessen Begegnung kein Zeitpunkt
// ist — **schreiben lässt er sich, lesen nicht**.
func krummeKontaktzeile(t *testing.T, bestand *sql.DB) {
	t.Helper()

	if _, fehler := bestand.ExecContext(t.Context(), `
INSERT INTO peerkontakte (profil_id, kennung, stand, gesehen)
VALUES (1, 'bkrumm', 'unbestaetigt', 'gestern')`); fehler != nil {
		t.Fatalf("die krumme Zeile: %v", fehler)
	}
}

// geschrieben prüft eine Antwort, die nach geglücktem Schreiben nicht nachladen
// konnte.
func geschrieben(t *testing.T, schritt string, antwort PeerkontakteAntwort) {
	t.Helper()

	if antwort.Fehler != nil {
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	}

	if antwort.Nachladefehler == nil {
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	}

	if antwort.Kontakte == nil || len(antwort.Zwecke) != len(domain.DieZwecke) {
		t.Fatalf("%s: die Antwort trägt %+v", schritt, antwort)
	}
}

func TestEinNachladefehlerIstKeinGescheiterterSchritt(t *testing.T) {
	t.Parallel()

	bestand := bestandFuerDenRiegel(t)
	anwendung := PeerflaecheUeber(bestand)
	ctx := riegelkontext(t)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung, abdruck := string(fremd.Kennung()), fremd.Fingerabdruck()
	ablage := storage.NeueKontaktablage(bestand)

	krummeKontaktzeile(t, bestand)

	// **In der Reihenfolge, in der jeder Schritt den vorigen braucht** —
	// erlaubt wird nur einem bestätigten Gegenüber.
	schritte := []struct {
		name      string
		weg       func(context.Context) PeerkontakteAntwort
		geschehen func(domain.Peerkontakt) bool
	}{
		{"aufnehmen", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerkontaktAufnehmen(ctx, kennung, "Annas Rechner")
		}, func(k domain.Peerkontakt) bool { return k.SelbstgewaehlterName == "Annas Rechner" }},
		{"bestätigen", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerkontaktBestaetigen(ctx, kennung, abdruck)
		}, func(k domain.Peerkontakt) bool { return k.Stand == domain.KontaktBestaetigt }},
		{"erlauben", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerzweckErlauben(ctx, kennung, string(domain.ZweckHandel))
		}, func(k domain.Peerkontakt) bool { return len(k.Rechte) == 1 }},
		{"entziehen", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerzweckEntziehen(ctx, kennung, string(domain.ZweckHandel))
		}, func(k domain.Peerkontakt) bool { return len(k.Rechte) == 0 }},
		{"Adresse setzen", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerkontaktAdresse(ctx, kennung, "192.168.1.7:41234")
		}, func(k domain.Peerkontakt) bool { return k.Adresse == "192.168.1.7:41234" }},
		{"Notiz setzen", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerkontaktNotiz(ctx, kennung, "vom Stammtisch")
		}, func(k domain.Peerkontakt) bool { return k.Notiz == "vom Stammtisch" }},
		{"sperren", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerkontaktSperren(ctx, kennung)
		}, func(k domain.Peerkontakt) bool { return k.Stand == domain.KontaktGesperrt }},
		{"entsperren", func(ctx context.Context) PeerkontakteAntwort {
			return anwendung.PeerkontaktEntsperren(ctx, kennung)
		}, func(k domain.Peerkontakt) bool { return k.Stand == domain.KontaktUnbestaetigt }},
	}

	for _, schritt := range schritte {
		geschrieben(t, schritt.name, schritt.weg(ctx))

		// **Und geschrieben ist es wirklich** — nachgelesen an der einen
		// Zeile, die sich lesen lässt.
		kontakt, lesefehler := ablage.Lesen(ctx, kennung)
		if lesefehler != nil || !schritt.geschehen(kontakt) {
			t.Fatalf("%s: nicht geschrieben: %+v (%v)", schritt.name, kontakt, lesefehler)
		}
	}

	// **Ein gescheiterter Schritt bleibt einer** — der Fehler steht in
	// `fehler` und nicht daneben, auch wenn die Liste ohnehin nicht zu lesen
	// wäre.
	antwort := anwendung.PeerkontaktBestaetigen(ctx, kennung, "0000-1111")
	if antwort.Fehler == nil || antwort.Nachladefehler != nil {
		t.Fatalf("der falsche Abdruck: %+v", antwort)
	}
}

// **Auch die Annahme einer Einladung** — der neunte Weg durch
// [App.amKontakt], und der, für den die Meldung beim Einladenden mitreist.
func TestEinNachladefehlerNachDerAnnahmeIstKeinGescheiterterSchritt(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	ctx := riegelkontext(t)

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "handel")
	if ausgestellt.Fehler != nil {
		t.Fatalf("Erstellen: %+v", ausgestellt.Fehler)
	}

	vorschau := bert.KontakteinladungPruefen(ctx, ausgestellt.Einladung.Code)
	if vorschau.Fehler != nil {
		t.Fatalf("Pruefen: %+v", vorschau.Fehler)
	}

	krummeKontaktzeile(t, bert.db)

	antwort := bert.KontakteinladungAnnehmen(
		ctx, ausgestellt.Einladung.Code, "Anna", "", false, "")
	geschrieben(t, "annehmen", antwort)

	// **Die Meldung beim Einladenden geht trotzdem mit** — sie gehört zur
	// Annahme und nicht zur Liste.
	if antwort.Gemeldet == nil {
		t.Fatalf("die Meldung fehlt: %+v", antwort)
	}

	if _, fehler := storage.NeueKontaktablage(bert.db).Lesen(
		ctx, vorschau.Vorschau.Kennung); fehler != nil {
		t.Fatalf("die Annahme ist nicht geschrieben: %v", fehler)
	}
}
