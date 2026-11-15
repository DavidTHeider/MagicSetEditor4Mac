package app

// **Die Sitzungsverwaltung an der Naht** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass der Profilwechsel wirklich schließt.** Das Issue verlangt es; die
// Regel steht in [service.Peersitzungen], und hier steht, dass die Anwendung
// sie auch ruft — an der einen Stelle, an der Profile wechseln.
//
// **Und dass es ohne Transport ein ehrliches Nein gibt** statt eines Sturzes:
// der Wähler kommt mit N2 und N3.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schleifenwaehler gibt eine Seite einer echten Schleife.
//
// **Kein Doppel mehr, seit `Oeffne` den Handschlag führt** (N2, #1455): eine
// Verbindung, die nur zählt, kommt über den ersten Schritt nicht hinaus — und
// ein Doppel, das ihn nachspielte, prüfte eine Sitzung, die es so nie gibt.
type schleifenwaehler struct{ leitung *leitung.Leitung }

func (s schleifenwaehler) Waehle(
	context.Context, peer.Kennung, string,
) (service.Verbindung, error) {
	return s.leitung, nil
}

// wechselprobe ist der kleinste Profilwechsler, den [App.profilUmstellen]
// braucht.
//
// **Nur `RolleImGezeigten` wird wirklich gerufen** — der Wechsel richtet die
// Leserverbindung, und dafür fragt er, ob dieses Profil nur lesen darf. Die
// übrigen Methoden stehen da, weil die Schnittstelle sie verlangt.
type wechselprobe struct{}

func (wechselprobe) Wechseln(
	context.Context, int64, string,
) (domain.Profil, error) {
	return domain.Profil{}, nil
}

func (wechselprobe) Gezeigt(context.Context) (domain.Profil, error) {
	return domain.Profil{}, nil
}

func (wechselprobe) Arbeitsprofil(context.Context) (domain.Profil, error) {
	return domain.Profil{}, nil
}

func (wechselprobe) WerArbeitet(context.Context) (int64, error) { return 1, nil }

func (wechselprobe) Anmeldeprofil(context.Context, int64) (int64, error) { return 1, nil }

func (wechselprobe) NutzerWechseln(context.Context, int64, string) error {
	return nil
}

func (wechselprobe) RolleImGezeigten(
	context.Context,
) (domain.Rolle, error) {
	return domain.Rolle(""), nil
}

// **Ohne Transport gibt es ein Nein und keinen Sturz.**
//
// # Drei Neins in ihrer Reihenfolge
//
// **Erst das Recht, dann der Ausweis, dann die Leitung** — und die Ordnung ist
// nach ihrem Preis gewählt:
//
//   - Die **Rechtefrage** ist ein Lesevorgang im eigenen Bestand.
//   - Der **Ausweis** geht an den Schlüsselbund, und der darf nachfragen.
//   - Die **Leitung** geht nach draußen und sagt einem Fremden, dass es
//     dieses Programm gibt.
//
// Wer eine frühere überspränge, zahlte für ein Nein, das er billiger hätte
// haben können.
func TestOhneIdentitaetUndOhneAdresseGibtEsEinNein(t *testing.T) {
	t.Parallel()

	anwendung, kennung, abdruck := eineKontaktwelt(t)
	anwendung.MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(riegelkontext(t))

	// **Ohne Freigabe endet es zuerst** — noch vor dem Schlüsselbund.
	if _, fehler := anwendung.sitzungsverwaltung().Oeffne(
		ctx, kennung, domain.ZweckHandel,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der ungefragte Zweck kam durch: %v", fehler)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(kennung), "Wer"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerkontaktBestaetigen(
		ctx, string(kennung), abdruck); antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerzweckErlauben(
		ctx, string(kennung), string(domain.ZweckTeilen),
	); antwort.Fehler != nil {
		t.Fatalf("Erlauben: %+v", antwort.Fehler)
	}

	// **Das Recht steht, die Identität fehlt.**
	_, fehler := anwendung.sitzungsverwaltung().Oeffne(
		ctx, kennung, domain.ZweckTeilen)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("ohne Identität kam etwas anderes: %v", fehler)
	}

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	// **Beides steht, die Adresse fehlt** — und genau das sagt die Antwort.
	//
	// **Hier stand ein `ErrConflict`** mit dem Satz „dieses Programm baut
	// noch keine Peerverbindungen auf". Seit N2 (#1455) gibt es einen
	// Transport, und was fehlt, ist nicht mehr das Programm, sondern die
	// **Adresse** dieses Gegenübers — ein `ErrNotFound` mit dem Namen der
	// Sache statt einer Auskunft über den Bauzustand (§8).
	_, fehler = anwendung.sitzungsverwaltung().Oeffne(
		ctx, kennung, domain.ZweckTeilen)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam etwas anderes: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "keine Adresse") {
		t.Fatalf("der Satz lautet: %v", fehler)
	}
}

// **Ein Profilwechsel schließt die offene Sitzung.**
//
// **Über den echten Weg** ([App.profilUmstellen]) und nicht über einen Ruf
// daneben: die Zusage gilt dem **Wechsel**, und ein Lauf, der die Verwaltung
// selbst pausierte, prüfte sie nicht.
//
// **Und über einen echten Handschlag**: die Gegenstelle ist eine zweite
// Anwendung mit eigenem Bestand, eigener Identität und eigenen Kontakten. Was
// hier nicht echt ist, sind die Bytes auf dem Draht.
func TestEinProfilwechselSchliesstDieOffeneSitzung(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	griff := hier.eineSitzungZu(t, dort, domain.ZweckPartie)

	if len(hier.anwendung.sitzungsverwaltung().Alle()) != 1 {
		t.Fatal("die Sitzung steht nicht in der Übersicht")
	}

	hier.anwendung.MitProfilwechsel(wechselprobe{})
	hier.anwendung.profilUmstellen(riegelkontext(t), 2)

	if len(hier.anwendung.sitzungsverwaltung().Alle()) != 0 {
		t.Error("die Sitzung steht noch in der Übersicht")
	}

	// **Und der alte Griff schreibt nicht mehr.**
	if fehler := griff.Sende(
		hier.ctx, []byte("zu spät"), service.Zuverlaessig,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("die verspätete Nachricht kam durch: %v", fehler)
	}

	// **Die Leitung ist zu** — sie gehört einem Profil, das niemand mehr
	// benutzt.
	if fehler := hier.leitung.Sende(
		hier.ctx, []byte("hallo")); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("die Leitung blieb offen: %v", fehler)
	}
}
