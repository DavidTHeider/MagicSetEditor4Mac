package app

// **Die Fehlerwege der Einrichtung und der Lauschwahl** — G08
// (#2272),
// gefunden am ersten vollen Lauf der Pipeline über G: drei Wege, die kein Test
// betrat.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Zuhause richtet „Alles einrichten" nichts ein** — und sagt es, mit
// leeren Listen statt `null` (§3, die Naht zwischen den Prüfebenen).
func TestAllesEinrichtenOhneZuhauseSagtEs(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("kein Zuhause")
	ohneHeim := NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return "", kaputt },
		func() (string, error) { return "/x", nil },
		func(context.Context) string { return "" },
	)

	for name, antwort := range map[string]SchnittstellenAlleAntwort{
		"einrichten": ohneHeim.SchnittstellenAlleEinrichten(t.Context()),
		"widerrufen": ohneHeim.SchnittstellenAlleWiderrufen(t.Context()),
	} {
		if antwort.Fehler == nil || antwort.Schnittstellen == nil || antwort.Ausgaenge == nil {
			t.Errorf("%s ohne Zuhause: %+v", name, antwort)
		}
	}
}

// **Eine genannte Adresse, die keine ist, startet keinen Dienst.**
func TestEinDienstAufEinerKrummenAdresseStartetNicht(t *testing.T) {
	t.Parallel()

	if _, fehler := NeueApp("").dienstStartenUeberWahl(t.Context(), domain.DienstREST,
		[]string{"keine adresse"}, 0); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen mit %v — erwartet ErrValidation", fehler)
	}
}

// **Ohne Bindung startet das Programm auf der Vorgabe** — der Schleife, mit
// einem freien Port; hier endet der Lauf vorher am fehlenden Programm.
func TestOhneBindungGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	naehte := echteDienstnaehte()
	kaputt := errors.New("kein Programm")
	naehte.orten = func() (string, error) { return "", kaputt }

	if _, fehler := starteDienstprogrammAufAllen(t.Context(), naehte, domain.DienstREST, nil); !errors.Is(fehler, kaputt) {
		t.Errorf("gestartet mit %v — erwartet den Fehler des Ortens", fehler)
	}
}
