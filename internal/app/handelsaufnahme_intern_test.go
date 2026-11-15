package app

// **Was mit einer hereinkommenden Handelssitzung geschieht** —
// #1634.
//
// # Wogegen hier geprüft wird
//
// **Dass jeder Zweck seinen Weg findet und der unbekannte keinen bekommt.**
// Bis zu diesem Werkstück reichte der Horchposten **jede** Sitzung an die
// Partie; eine mit dem Zweck `handel` fiel still auf den Boden. Ein Zweck
// ohne Weg wird jetzt **abgewiesen** — still fallen lassen sähe für den
// Anrufer aus wie eine Leitung, die steht.

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// derSitzungsvorgang ist der Handel, der über die Sitzung kommt.
const derSitzungsvorgang = "v-sitzung"

// eineHandelsanwendung baut eine Anwendung, die ein Gegenüber handeln lässt.
//
// **Mit eingesetztem Bund** (§3): ohne ihn schriebe dieser Lauf eine
// Identität in den Anmeldebund dieses Rechners.
func eineHandelsanwendung(t *testing.T) (*App, *angenommenendoppel) {
	t.Helper()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	// **Das Profil steht, bevor etwas hereinkommt.** Beim Horchposten hat
	// [App.Starten] den Zeiger gelesen; ohne ihn liefe dieser Lauf ohne
	// Profil, und der Kontakt von oben läge außer Reichweite.
	anwendung.profil.Store(1)

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	jener, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := jener.Kennung()

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(kennung), "Der andere"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerkontaktBestaetigen(
		ctx, string(kennung), peer.Fingerabdruck(kennung),
	); antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerzweckErlauben(
		ctx, string(kennung), string(domain.ZweckHandel),
	); antwort.Fehler != nil {
		t.Fatalf("Erlauben: %+v", antwort.Fehler)
	}

	return anwendung, &angenommenendoppel{stand: service.Sitzungsstand{
		Zweck: domain.ZweckHandel, Gegenueber: kennung,
		Fingerabdruck: peer.Fingerabdruck(kennung),
	}}
}

// eineHandelsnachricht baut, was über die Leitung kommt — roh.
//
// **Die erste Revision**, denn genau die ist der Fall: wem ein Handel
// angeboten wird, der kennt ihn noch nicht.
//
// **Sie nennt dieses Profil als Partei** — sonst weist die Post sie ab
// (Befund Z3 des Zweitblicks über Phase 6.8). Hier stand `"ich"`, und das
// war genau der Fall eines Vorgangs, der dieses Profil nicht betrifft.
func eineHandelsnachricht(t *testing.T, absender peer.Kennung, ich string) []byte {
	t.Helper()

	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: derSitzungsvorgang, Revision: 1,
		Art: domain.HandelTausch, Initiator: string(absender),
		Parteien: []string{string(absender), ich},
		Zustand:  domain.HandelAngeboten,
		Positionen: []domain.Handelsposition{{
			Von: string(absender), Printing: "a1b2", Menge: 1,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	}

	roh, fehler := json.Marshal(map[string]any{
		"Umschlag": domain.Handelsumschlag{
			Fassung: domain.Handelsprotokollfassung,
			Art:     domain.NachrichtVorschlag, NachrichtID: "n-1",
			Vorgang: derSitzungsvorgang, Basisrevision: 1,
			Abdruck: angebot.Abdruck(), Folge: 1,
			Absender: string(absender), Unterschrift: "sig",
		},
		"Angebot": angebot,
	})
	if fehler != nil {
		t.Fatalf("die Nachricht bauen: %v", fehler)
	}

	return roh
}

// **Eine Sitzung, für deren Zweck es keinen Weg gibt, wird abgewiesen.**
//
// **Und nicht still fallen gelassen**: der Anrufer sähe eine Leitung, die
// steht, und wartete auf eine Antwort, die niemand schreibt (§8).
func TestEineSitzungOhneWegWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	sitzung := eineAngenommene(t, domain.ZweckTeilen)

	freigabeanwendung(t).sitzungAufnehmen(t.Context(), sitzung)

	if !sitzung.istZu() {
		t.Error("die Sitzung blieb offen, obwohl es für sie keinen Weg gibt")
	}
}

// **Eine Partiesitzung geht weiter an den Tisch.**
//
// **Gemessen daran, dass sie niedergelegt wird**: ohne offene Freigabe ist
// das die Antwort des Tischwegs, und eine Sitzung, die der Umschalter gar
// nicht weiterreichte, bliebe offen.
func TestEinePartiesitzungGehtAnDenTisch(t *testing.T) {
	t.Parallel()

	sitzung := eineAngenommene(t, domain.ZweckPartie)

	freigabeanwendung(t).sitzungAufnehmen(t.Context(), sitzung)

	if !sitzung.istZu() {
		t.Error("die Partiesitzung erreichte den Tischweg nicht")
	}
}

// **Eine Handelsnachricht kommt über die Sitzung an — und wirkt.**
//
// **Das ist der Weg, den #1634 gebaut hat**: bis dahin gab es ihn nicht, und
// `service.Handelspost.EmpfangeUeber` stand ohne Aufrufer da.
func TestEineHandelsnachrichtKommtUeberDieSitzungAn(t *testing.T) {
	t.Parallel()

	anwendung, sitzung := eineHandelsanwendung(t)

	// **Eine Nachricht, dann ein Ende** — sonst liefe die Schleife weiter
	// und der Lauf hinge bis zu seiner Frist.
	sitzung.gibt = [][]byte{
		eineHandelsnachricht(t, sitzung.stand.Gegenueber, eigeneKennungVon(t, anwendung)),
	}
	sitzung.lesefehler = errors.New("die Leitung ist zu")

	// **Mit dem Profil im Kontext**, wie der Horchposten ihn reicht: er
	// trägt das Profil, das ihn geöffnet hat (Zweitblick über Phase 6.7, Z2).
	anwendung.sitzungAufnehmen(anwendung.imProfil(t.Context()), sitzung)

	if !sitzung.istZu() {
		t.Error("die Sitzung blieb nach dem Abriss offen")
	}

	antwort := anwendung.HandelStand(t.Context(), derSitzungsvorgang)
	if antwort.Fehler != nil {
		t.Fatalf("der Vorgang kam nicht an: %+v", antwort.Fehler)
	}

	if antwort.Vorgang.Zustand != string(domain.HandelAngeboten) {
		t.Errorf("der Zustand ist %q", antwort.Vorgang.Zustand)
	}
}

// **Eine Leitung, die sich nicht schließen lässt, ändert nichts.**
//
// **Die Gegenprobe zum Satz im Kopf von [schliesseHandelssitzung]**: der
// Grund geht ins Protokoll, und der Weg endet trotzdem ordentlich.
func TestEineSitzungDieSichNichtSchliessenLaesstEndetTrotzdem(t *testing.T) {
	t.Parallel()

	sitzung := eineAngenommene(t, domain.ZweckHandel)
	sitzung.lesefehler = errors.New("die Leitung ist zu")
	sitzung.schliessfehler = errors.New("und sie klemmt")

	freigabeanwendung(t).sitzungAufnehmen(t.Context(), sitzung)

	if !sitzung.istZu() {
		t.Error("geschlossen wurde gar nicht erst versucht")
	}
}

// **Die Strecke leert nach jedem Senden.**
//
// **Sonst läge die Nachricht noch da**: `Sende` am Griff reiht ein, hinaus
// geht es mit `Leere` — und eine Strecke, die nur einreihte, sähe aus wie
// eine, die sendet.
func TestDieStreckeLeertNachJedemSenden(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		sitzung *angenommenendoppel
		klagt   bool
	}{
		{name: "beides gelingt", sitzung: &angenommenendoppel{}},
		{name: "das Senden scheitert", klagt: true, sitzung: &angenommenendoppel{
			sendefehler: errors.New("die Leitung ist voll"),
		}},
		{name: "das Leeren scheitert", klagt: true, sitzung: &angenommenendoppel{
			leerfehler: errors.New("die Leitung ist weg"),
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := handelsstrecke{griff: fall.sitzung}.Sende(
				t.Context(), []byte("hallo"))

			if (fehler != nil) != fall.klagt {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// eigeneKennungVon gibt die Peerkennung des gezeigten Profils.
func eigeneKennungVon(t *testing.T, anwendung *App) string {
	t.Helper()

	ausweis, fehler := anwendung.Ausweis(anwendung.imProfil(riegelkontext(t)))
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	return string(ausweis.Kennung())
}
