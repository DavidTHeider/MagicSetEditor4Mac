package app

// **Die Schnittstellen dieses Rechners am Fenster** — G08
// (#2272).
//
// Über eingesetzte Lesewege (C06): welche Schnittstellen es hier gibt, hängt
// an WLAN, VPN und einer laufenden VM.

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
)

func TestDieNetzschnittstellenKommenAlsListe(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("").MitNetzschnittstellen(func() ([]entdeckung.Schnittstelleneintrag, error) {
		return []entdeckung.Schnittstelleneintrag{
			{Name: "lo0", Art: entdeckung.ArtSchleife, Aktiv: true, Adressen: []string{"127.0.0.1", "::1"}},
			{Name: "en0", Art: entdeckung.ArtNetz, Aktiv: true, Adressen: []string{"192.168.178.74"}},
			{Name: "xyz9", Aktiv: true, Adressen: []string{}, Befund: "nicht lesbar"},
		}, nil
	})

	antwort := NeueBindungen(anwendung).Netzschnittstellen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Netzschnittstellen: %+v", antwort.Fehler)
	}

	erwartet := []NetzschnittstelleDTO{
		{Name: "lo0", Art: "schleife", Aktiv: true, Adressen: []string{"127.0.0.1", "::1"}, Befund: ""},
		{Name: "en0", Art: "netz", Aktiv: true, Adressen: []string{"192.168.178.74"}, Befund: ""},
		{Name: "xyz9", Art: "", Aktiv: true, Adressen: []string{}, Befund: "nicht lesbar"},
	}

	if !reflect.DeepEqual(antwort.Schnittstellen, erwartet) {
		t.Errorf("die Liste ist %+v", antwort.Schnittstellen)
	}
}

func TestOhneLesbareNetzschnittstellenEinFehler(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("").MitNetzschnittstellen(func() ([]entdeckung.Schnittstelleneintrag, error) {
		return nil, errors.New("verweigert")
	})

	antwort := anwendung.Netzschnittstellen(t.Context())
	if antwort.Fehler == nil || antwort.Schnittstellen == nil || len(antwort.Schnittstellen) != 0 {
		t.Errorf("ohne Liste: %+v", antwort)
	}
}

func TestDieAlleBindungenReichenDurch(t *testing.T) {
	t.Parallel()

	anwendung, _ := eineEinrichtende(t, "/usr/bin")
	anwendung.MitSchnittstellenbefehlen((&befehlsbuch{}).ausfuehren)
	bindungen := NeueBindungen(anwendung)

	if ausgang := ausgangVon(bindungen.SchnittstellenAlleEinrichten(t.Context()), "codex"); ausgang.Ausgang != "eingerichtet" {
		t.Errorf("alle einrichten: %+v", ausgang)
	}

	if ausgang := ausgangVon(bindungen.SchnittstellenAlleWiderrufen(t.Context()), "codex"); ausgang.Ausgang != "widerrufen" {
		t.Errorf("alle zurücknehmen: %+v", ausgang)
	}
}
