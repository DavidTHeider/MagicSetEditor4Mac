package app

// **Die Kontakte an der Naht** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Die Übersetzung und der Weg dorthin** — nicht die Regel: was ein Kontakt
// darf, entscheidet [service.Kontaktwerk], und dort steht der Lauf dazu. Hier
// steht, dass die Naht den Bestand wirklich erreicht und die Antwort trägt,
// was eine Fläche braucht — **die ganze Liste und die Zwecke**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineKontaktwelt baut eine Anwendung über einem Bestand.
func eineKontaktwelt(t *testing.T) (*App, peer.Kennung, string) {
	t.Helper()

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return PeerflaecheUeber(bestandFuerDenRiegel(t)),
		fremd.Kennung(), fremd.Fingerabdruck()
}

// **Ohne Begegnung ist die Liste leer — und die Zwecke stehen trotzdem da.**
//
// **Beides zusammen ist der Punkt.** Eine Fläche, die nur die Kontakte bekäme,
// wüsste beim ersten Gegenüber nicht, was sie ihm überhaupt freigeben kann.
func TestOhneBegegnungIstDieKontaktlisteLeer(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := eineKontaktwelt(t)

	antwort := anwendung.Peerkontakte(riegelkontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Kontakte == nil || len(antwort.Kontakte) != 0 {
		t.Fatalf("die Liste ist %+v", antwort.Kontakte)
	}

	if len(antwort.Zwecke) != len(domain.DieZwecke) {
		t.Fatalf("die Zwecke sind %v", antwort.Zwecke)
	}
}

// **Der ganze Weg: aufnehmen, bestätigen, freigeben, entziehen, sperren.**
//
// # Warum als ein Lauf und nicht als fünf
//
// **Weil jeder Schritt den vorigen voraussetzt.** Fünf Läufe bauten den
// Zustand fünfmal auf, und jeder von ihnen prüfte am Ende dasselbe: dass die
// Naht den Bestand erreicht. Was hier gemessen wird, ist die **Folge**.
func TestDerGanzeKontaktwegAnDerNaht(t *testing.T) {
	t.Parallel()

	anwendung, kennung, abdruck := eineKontaktwelt(t)
	ctx := riegelkontext(t)

	antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(kennung), "Annas Rechner")
	if antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if len(antwort.Kontakte) != 1 ||
		antwort.Kontakte[0].Stand != string(domain.KontaktUnbestaetigt) {
		t.Fatalf("nach dem Aufnehmen: %+v", antwort.Kontakte)
	}

	// **Der Abdruck kommt fertig aus Go** — eine Fläche, die ihn selbst
	// rechnen müsste, kann ihn vergessen.
	if antwort.Kontakte[0].Fingerabdruck != abdruck {
		t.Errorf("der Abdruck ist %q", antwort.Kontakte[0].Fingerabdruck)
	}

	// **Ein falscher Abdruck bestätigt nichts.**
	if fehler := anwendung.PeerkontaktBestaetigen(
		ctx, string(kennung), "0000-1111").Fehler; fehler == nil {
		t.Fatal("der falsche Abdruck kam durch")
	}

	antwort = anwendung.PeerkontaktBestaetigen(ctx, string(kennung), abdruck)
	if antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	antwort = anwendung.PeerzweckErlauben(
		ctx, string(kennung), string(domain.ZweckHandel))
	if antwort.Fehler != nil {
		t.Fatalf("Erlauben: %+v", antwort.Fehler)
	}

	if len(antwort.Kontakte[0].Rechte) != 1 {
		t.Fatalf("die Rechte sind %v", antwort.Kontakte[0].Rechte)
	}

	antwort = anwendung.PeerzweckEntziehen(
		ctx, string(kennung), string(domain.ZweckHandel))
	if antwort.Fehler != nil {
		t.Fatalf("Entziehen: %+v", antwort.Fehler)
	}

	if len(antwort.Kontakte[0].Rechte) != 0 {
		t.Fatalf("die Rechte sind %v", antwort.Kontakte[0].Rechte)
	}

	antwort = anwendung.PeerkontaktSperren(ctx, string(kennung))
	if antwort.Fehler != nil {
		t.Fatalf("Sperren: %+v", antwort.Fehler)
	}

	if antwort.Kontakte[0].Stand != string(domain.KontaktGesperrt) {
		t.Fatalf("der Stand ist %q", antwort.Kontakte[0].Stand)
	}

	antwort = anwendung.PeerkontaktEntsperren(ctx, string(kennung))
	if antwort.Fehler != nil {
		t.Fatalf("Entsperren: %+v", antwort.Fehler)
	}

	// **Auf unbestätigt und nicht auf den Stand von vorher.**
	if antwort.Kontakte[0].Stand != string(domain.KontaktUnbestaetigt) {
		t.Fatalf("der Stand ist %q", antwort.Kontakte[0].Stand)
	}

	// **Und der Zeitpunkt steht als RFC 3339 da** — dieselbe Gestalt wie
	// überall sonst an dieser Naht.
	if antwort.Kontakte[0].Gesehen == "" {
		t.Error("der Zeitpunkt fehlt")
	}
}

// **Ein Fehler kommt mit leeren Listen zurück und nicht mit `null`.**
//
// **Der Wächter `leerelistenlaufzeit_test.go` hält das für jede Bindung** —
// eine Liste, die aus der Antwort verschwindet, lässt das Fenster ohne Grund
// leer stehen (§3, „zwischen den Prüfebenen").
func TestEinKontaktfehlerKommtMitLeerenListen(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := eineKontaktwelt(t)

	antwort := anwendung.PeerkontaktBestaetigen(
		riegelkontext(t), "bniemand", "abcd")
	if antwort.Fehler == nil {
		t.Fatal("die unbekannte Kennung kam durch")
	}

	if antwort.Kontakte == nil || antwort.Zwecke == nil {
		t.Fatalf("die Antwort trägt %+v", antwort)
	}
}

// **Die Naht liest auch ein einzelnes Gegenüber.**
//
// # Warum dieser Weg heute keine Bindung hat
//
// **Weil er die Frage der Verbindung beantwortet und nicht die der Liste.**
// [service.Kontaktwerk.Darf] ist der Riegel, den N1.5 vor jede Sitzung setzt:
// „darf dieses Gegenüber diesen Zweck?". Eine Fläche dafür gibt es noch nicht,
// weil es die Verbindung noch nicht gibt — die Naht dorthin aber schon, und
// eine ungemessene Naht ist eine Behauptung (§10).
func TestDieNahtLiestAuchEinEinzelnesGegenueber(t *testing.T) {
	t.Parallel()

	anwendung, kennung, abdruck := eineKontaktwelt(t)
	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(kennung), "Wer"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	darf, fehler := anwendung.kontaktwerk().Darf(
		ctx, kennung, domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("ein unbestätigtes Gegenüber darf handeln")
	}

	if antwort := anwendung.PeerkontaktBestaetigen(
		ctx, string(kennung), abdruck); antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerzweckErlauben(
		ctx, string(kennung), string(domain.ZweckHandel)); antwort.Fehler != nil {
		t.Fatalf("Erlauben: %+v", antwort.Fehler)
	}

	darf, fehler = anwendung.kontaktwerk().Darf(
		ctx, kennung, domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if !darf {
		t.Fatal("das freigegebene Gegenüber darf nicht handeln")
	}
}

// **Ein Bestand ohne Tabelle meldet sich, statt leer zu antworten.**
func TestEinBestandOhneKontakttabelleMeldetSich(t *testing.T) {
	t.Parallel()

	bestand := bestandFuerDenRiegel(t)

	if _, err := bestand.ExecContext(
		t.Context(), `DROP TABLE peerkontakte`); err != nil {
		t.Fatalf("fallen lassen: %v", err)
	}

	antwort := PeerflaecheUeber(bestand).Peerkontakte(riegelkontext(t))
	if antwort.Fehler == nil {
		t.Fatal("ohne Tabelle kam eine Liste")
	}

	if antwort.Kontakte == nil || antwort.Zwecke == nil {
		t.Fatalf("die Antwort trägt %+v", antwort)
	}
}
