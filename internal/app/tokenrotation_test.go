package app_test

// **Rotation, Sammelwiderruf und Ablaufwahlen am Fenster** — Werkstück O2
// von Phase 5.2 (#1426).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieRotationGibtDenKlartextGenauEinmal(t *testing.T) {
	t.Parallel()

	fuehrer := &tokenfuehrer{klartext: "mse_7_neu"}
	antwort := mitToken(fuehrer).TokenRotieren(probekontext(t), "7")

	if antwort.Fehler != nil || antwort.Klartext != "mse_7_neu" || fuehrer.rotiert != 7 {
		t.Fatalf("Rotation: %+v — rotiert %d", antwort, fuehrer.rotiert)
	}

	if antwort.Token == nil || antwort.Token.VorgaengerBis != "2026-09-08T13:00:00Z" ||
		antwort.Token.Principal != "Baurechner" || antwort.Token.Fassung != "2" {
		t.Errorf("die Zeile trägt die Rotation nicht: %+v", antwort.Token)
	}
}

func TestDerSammelwiderrufSagtWieVieleAmFenster(t *testing.T) {
	t.Parallel()

	fuehrer := &tokenfuehrer{}
	antwort := mitToken(fuehrer).TokenAlleWiderrufen(probekontext(t), "3")

	if antwort.Fehler != nil || antwort.Gezogen != 2 || fuehrer.widerrufen != 3 {
		t.Errorf("Sammelwiderruf: %+v — Profil %d", antwort, fuehrer.widerrufen)
	}
}

func TestDieAblaufwahlenKommenAlsZeilen(t *testing.T) {
	t.Parallel()

	antwort := mitToken(&tokenfuehrer{}).Tokenablaufwahlen(probekontext(t))

	if len(antwort.Wahlen) != 2 || antwort.Wahlen[0].Name != "tag" ||
		antwort.Wahlen[0].Ablauf != "2026-09-09T12:00:00Z" || antwort.Wahlen[1].Ablauf != "" {
		t.Errorf("Ablaufwahlen: %+v", antwort.Wahlen)
	}
}

// **Jede neue Bindung meldet ihr Nein** — eine krumme Kennung und ein
// Dienst, der klemmt.
func TestDieNeuenTokenbindungenMeldenIhrNein(t *testing.T) {
	t.Parallel()

	kaputt := &tokenfuehrer{fehler: errors.New("die Ablage klemmt")}

	if antwort := mitToken(kaputt).TokenRotieren(probekontext(t), "7"); antwort.Fehler == nil {
		t.Error("die Rotation verschwieg den Fehler")
	}

	if antwort := mitToken(kaputt).TokenAlleWiderrufen(probekontext(t), "3"); antwort.Fehler == nil {
		t.Error("der Sammelwiderruf verschwieg den Fehler")
	}

	heil := &tokenfuehrer{}

	if antwort := mitToken(heil).TokenRotieren(probekontext(t), "x"); antwort.Fehler == nil ||
		antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("krumme Tokenkennung: %+v", antwort.Fehler)
	}

	if antwort := mitToken(heil).TokenAlleWiderrufen(probekontext(t), "x"); antwort.Fehler == nil ||
		antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("krumme Profilkennung: %+v", antwort.Fehler)
	}

	// **Und die Bindungen nehmen denselben Weg.**
	gebunden := app.NeueBindungen(mitToken(heil))
	if gebunden.TokenRotieren(t.Context(), "7").Klartext != "" || gebunden.TokenAlleWiderrufen(t.Context(), "3").Gezogen != 2 ||
		len(gebunden.Tokenablaufwahlen(t.Context()).Wahlen) != 2 {
		t.Error("die Bindungen reichen nicht durch")
	}
}

// **Und dieselben drei Wege über die Fläche ohne Fenster** — gegen den
// echten Bestand, denn dort entsteht das Geheimnis wirklich.
func TestRotationUndSammelwiderrufUeberDieFlaeche(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(frischerBestand(t))

	angelegt := flaeche.TokenAnlegen(probekontext(t), app.TokenangabenDTO{
		ProfilID: "1", Bezeichnung: "Nachtlauf", Principal: "Baurechner", Scope: "lesen",
	})
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	rotiert := flaeche.TokenRotieren(probekontext(t), angelegt.Token.ID)
	if rotiert.Fehler != nil || rotiert.Klartext == "" || rotiert.Klartext == angelegt.Klartext ||
		rotiert.Token.VorgaengerBis == "" {
		t.Fatalf("rotieren: %+v", rotiert)
	}

	if wahlen := flaeche.Tokenablaufwahlen(probekontext(t)); len(wahlen.Wahlen) != 4 {
		t.Errorf("Ablaufwahlen: %+v", wahlen)
	}

	if gezogen := flaeche.TokenAlleWiderrufen(probekontext(t), "1"); gezogen.Fehler != nil || gezogen.Gezogen != 1 {
		t.Errorf("Sammelwiderruf: %+v", gezogen)
	}
}
