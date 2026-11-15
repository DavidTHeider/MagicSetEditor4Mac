package app

// **Wer in der Nähe ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine eigene Datei**, weil `peersichtbar_intern_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, **wie man
// sich ankündigt** — hier, **wen man findet** und was aus einem Fund wird.

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Die Nachbarliste sagt, wen dieses Profil schon kennt.**
func TestDieNachbarlisteSagtWenDiesesProfilKennt(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	bekannt, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(bekannt.Kennung()), "Kennen wir"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	anwendung.MitEntdeckung(nil, func(
		context.Context, entdeckung.Frage,
	) (entdeckung.Ergebnis, error) {
		return entdeckung.Ergebnis{
			Funde: []entdeckung.Fund{
				{Kennung: fremd.Kennung(), Name: "Fremd", Wirt: "192.168.1.7", Port: 41234},
				{Kennung: bekannt.Kennung(), Name: "Bekannt", Wirt: "192.168.1.8", Port: 41234},
			},
			Verworfen: 2,
		}, nil
	})

	antwort := anwendung.Nachbarn(ctx)

	if antwort.Fehler != nil {
		t.Fatalf("Nachbarn: %+v", antwort.Fehler)
	}

	if antwort.Verworfen != 2 {
		t.Errorf("verworfen wurden %d", antwort.Verworfen)
	}

	nach := map[string]Nachbar{}
	for _, nachbar := range antwort.Nachbarn {
		nach[nachbar.Kennung] = nachbar
	}

	if nach[string(bekannt.Kennung())].Bekannt != true ||
		nach[string(bekannt.Kennung())].Stand != "unbestaetigt" {
		t.Errorf("der bekannte ist %+v", nach[string(bekannt.Kennung())])
	}

	if nach[string(fremd.Kennung())].Bekannt {
		t.Errorf("der fremde ist %+v", nach[string(fremd.Kennung())])
	}

	// **Und die Adresse ist die, die gewählt würde.**
	if nach[string(fremd.Kennung())].Adresse != "192.168.1.7:41234" {
		t.Errorf("die Adresse ist %q", nach[string(fremd.Kennung())].Adresse)
	}

	// **Der bekannte ohne gemerkte Adresse hat eine zu übernehmen**
	// (NET-02.NAT-02, #1706) — der fremde nicht, der wird aufgenommen.
	if gekannt := nach[string(bekannt.Kennung())]; gekannt.GemerkteAdresse != "" || !gekannt.AdresseGeaendert {
		t.Errorf("der bekannte ist %+v", gekannt)
	}

	if nach[string(fremd.Kennung())].AdresseGeaendert {
		t.Errorf("der fremde ist %+v", nach[string(fremd.Kennung())])
	}
}

// **Der eigene Widerhall steht nicht in der Liste.**
func TestDerEigeneWiderhallStehtNichtInDerNachbarliste(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	eigen := anwendung.Peeridentitaet(ctx)
	if eigen.Fehler != nil {
		t.Fatalf("Peeridentitaet: %+v", eigen.Fehler)
	}

	anwendung.MitEntdeckung(nil, func(
		context.Context, entdeckung.Frage,
	) (entdeckung.Ergebnis, error) {
		return entdeckung.Ergebnis{Funde: []entdeckung.Fund{{
			Kennung: peer.Kennung(eigen.Kennung), Name: "Ich",
			Wirt: "127.0.0.1", Port: 41234,
		}}}, nil
	})

	antwort := anwendung.Nachbarn(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("Nachbarn: %+v", antwort.Fehler)
	}

	if len(antwort.Nachbarn) != 0 {
		t.Fatalf("es kamen %+v", antwort.Nachbarn)
	}
}

// **Ein Fehler der Suche kommt als Fehler an — mit leerer Liste.**
func TestEinFehlerDerNachbarsucheKommtMitLeererListe(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	anwendung.MitEntdeckung(nil, func(
		context.Context, entdeckung.Frage,
	) (entdeckung.Ergebnis, error) {
		return entdeckung.Ergebnis{}, errors.New("das Netz schweigt")
	})

	antwort := anwendung.Nachbarn(ctx)
	if antwort.Fehler == nil {
		t.Fatal("die Suche kam durch")
	}

	// **Leer und nicht `nil`** — der Wächter über leere Listen hält das für
	// jede Bindung.
	if antwort.Nachbarn == nil {
		t.Fatal("die Liste ist nil")
	}
}

// **Eine krumme Adresse des Lauschers ist ein Befund über dieses Programm.**
//
// **Sie kommt vom Lauscher und nicht von einem Menschen** — deshalb ist das
// hier keine Eingabeprüfung, sondern die Stelle, an der ein Adapter auffiele,
// der etwas anderes liefert als eine Adresse.
func TestEineKrummeLauscheradresseIstEinBefund(t *testing.T) {
	t.Parallel()

	for name, adresse := range map[string]string{
		"gar keine Adresse": "krumm",
		"kein Port":         "127.0.0.1:keinPort",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung, ctx := eineHorchende(
				t, &krummerLauscher{adresse: adresse}, nil)

			t.Cleanup(anwendung.horchenBeenden)

			anwendung.MitEntdeckung(func(
				entdeckung.Angebot,
			) (Ankuendigung, error) {
				return &probeankuendigung{}, nil
			}, nil)

			if antwort := anwendung.HorchStarten(
				ctx, "127.0.0.1:0"); antwort.Fehler != nil {
				t.Fatalf("HorchStarten: %+v", antwort.Fehler)
			}

			if antwort := anwendung.SichtbarWerden(
				ctx, "Wer"); antwort.Fehler == nil {
				t.Fatal("die Ankündigung kam durch")
			}
		})
	}
}

// krummerLauscher nennt eine Adresse, die keine ist.
type krummerLauscher struct {
	probelauscher

	adresse string
}

func (k *krummerLauscher) Adresse() string { return k.adresse }

// **Über die echten Nähte, und zwar auf der Rückschleife.**
//
// # Warum dieser Lauf wirklich ins Netz geht — und warum das keines ist
//
// **Weil er sonst die echten Nähte nie beträte.** `MitEntdeckung` ist die
// Vorrichtung; was im Betrieb läuft, ist `entdeckung.Melde` und
// `entdeckung.Suche`, und ein Lauf, der beide auslässt, prüft die
// Verdrahtung nicht.
//
// **`lo0` ist kein Netz.** Es trägt Multicast (nachgemessen), und was darüber
// geht, verlässt diesen Rechner nicht — dafür gibt es
// [Entdeckungsschnittstelle].
func TestUeberDieEchtenNaehteAufDerRueckschleife(t *testing.T) {
	if _, fehler := net.InterfaceByName("lo0"); fehler != nil {
		t.Skipf("dieser Rechner hat kein lo0: %v", fehler)
	}

	t.Setenv(Entdeckungsschnittstelle, "lo0")

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SichtbarWerden(
		ctx, "Davids Rechner"); antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	// **Und die Suche findet sich selbst nicht.**
	//
	// # Warum gegen die Kennung geprüft wird und nicht gegen den Namen
	//
	// **Weil ein Name nichts trennt.** Der erste Anlauf suchte nach „Davids
	// Rechner" — und fand ihn: `internal/entdeckung` kündigt in seinem
	// eigenen Lauf denselben Namen an, und `go test` fährt beide Pakete
	// nebeneinander. Über `lo0` hört jeder jeden.
	//
	// **Die Kennung trennt.** Sie ist der öffentliche Schlüssel, sie ist je
	// Lauf eine andere, und sie ist das, wonach der Code wirklich filtert.
	eigen := anwendung.Peeridentitaet(ctx)
	if eigen.Fehler != nil {
		t.Fatalf("Peeridentitaet: %+v", eigen.Fehler)
	}

	antwort := anwendung.Nachbarn(ctx)

	// **Und `lo0` allein genügt nicht.** Die Schnittstelle trägt Multicast;
	// ob dieser **Prozess** darüber senden darf, ist eine zweite Frage — unter
	// macOS ab 15 braucht er dafür die Erlaubnis für das lokale Netzwerk. Am
	// 2026-09-20 war die Antwort auf diesem Rechner nein, und dieser Lauf war
	// rot, ohne dass etwas am Code falsch gewesen wäre.
	//
	// **Gefragt wird die Antwort des Produkts und keine eigene Netzprobe**:
	// `deuteSuchfehler` deutet `EHOSTUNREACH`, `ENETUNREACH` und `EPERM`
	// bereits als fehlende Erlaubnis, und eine zweite Antwort daneben liefe
	// mit der ersten auseinander.
	//
	// **Laut übersprungen** — was dieser Lauf sagen soll, ist an einem
	// solchen Tag nicht geprüft, und das steht in `docs/HAERTUNG.md`.
	if antwort.Fehler != nil && antwort.Fehler.Schluessel == "ohneNetzerlaubnis" {
		t.Skipf("dieser Rechner ließ die Suche nicht hinaus (%s) — unter "+
			"macOS ab 15 braucht das laufende Programm die Erlaubnis für das "+
			"lokale Netzwerk (Systemeinstellungen → Datenschutz & Sicherheit "+
			"→ Lokales Netzwerk)", antwort.Fehler.Meldung)
	}

	if antwort.Fehler != nil {
		t.Fatalf("Nachbarn: %+v", antwort.Fehler)
	}

	for _, nachbar := range antwort.Nachbarn {
		if nachbar.Kennung == eigen.Kennung {
			t.Fatalf("der eigene Widerhall steht in der Liste: %+v", nachbar)
		}
	}
}

// **Eine Schnittstelle, die es nicht gibt, hält nichts auf.**
//
// **Ein Tippfehler ist kein Grund, nicht zu suchen** — er bekommt die Vorgabe
// und einen Eintrag im Protokoll.
func TestEineUnbekannteSchnittstelleHaeltNichtsAuf(t *testing.T) {
	t.Setenv(Entdeckungsschnittstelle, "gibtesnicht0")

	if schnitt := PeerflaecheUeber(bestandFuerDenRiegel(t)).entdeckungsschnittstelle(
		riegelkontext(t)); schnitt != nil {
		t.Fatalf("es kam %+v", schnitt)
	}
}

// **Die Einstellung schlägt die Umgebung, und leer ist die Automatik** —
// NET-02.AK-03 (#1706).
// Bis hierher ließ sich die Schnittstelle nur über eine Umgebungsvariable
// wählen, und zurück ging es nur, indem man sie wieder entfernte.
func TestDieSchnittstelleKommtAusDerEinstellung(t *testing.T) {
	t.Setenv(Entdeckungsschnittstelle, "gibtesnicht0")

	anwendung := PeerflaecheUeber(bestandFuerDenRiegel(t))
	ctx := anwendung.imProfil(riegelkontext(t))

	einstellungen := storage.NeueEinstellungen(anwendung.db)
	if fehler := einstellungen.Setzen(ctx, domain.NetzSchnittstelle, "lo0"); fehler != nil {
		t.Fatalf("Setzen: %v", fehler)
	}

	if schnitt := anwendung.entdeckungsschnittstelle(ctx); schnitt == nil || schnitt.Name != "lo0" {
		t.Fatalf("es kam %+v", schnitt)
	}

	// **Zurück auf Automatik**: leer — dann gilt wieder, was die Umgebung
	// sagt, und die sagt hier einen Namen, den es nicht gibt.
	if fehler := einstellungen.Setzen(ctx, domain.NetzSchnittstelle, ""); fehler != nil {
		t.Fatalf("Zurücksetzen: %v", fehler)
	}

	if schnitt := anwendung.entdeckungsschnittstelle(ctx); schnitt != nil {
		t.Fatalf("nach dem Zurücksetzen kam %+v", schnitt)
	}
}
