package app

// **Eine Wortwahl am Fenster ist so bedingt wie das Tippen** — Befund 1 des
// Nachlaufs über Phase 0.7.0-I zu U02b
// (#2179).
//
// Die Wahl aus der Auswahlliste unter einem Feld rechnet ihren Wert aus dem
// gespeicherten und schrieb ohne Bedingung. Stand ein fremder Wert im Feld,
// den das Fenster noch nicht gezeigt hatte — oder stand der Entwurf sichtbar
// im Konflikt —, änderte die Wahl ihn, bevor jemand entschieden hatte. Jetzt
// trägt sie die Werte mit, die der Entwurf kennt, wie [App.FeldSetzenBedingt].

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineWortwahlMitBekanntenWertenUebergehtKeinenFremden(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	karteID := karten.Karten[0].ID
	text := func() string {
		for _, wert := range anwendung.Karte(t.Context(), karteID).Karte.Werte {
			if wert.Feld == "text" {
				return wert.Wert
			}
		}

		return ""
	}
	gesehen := "<word-list-probe-art>Große Kreatur</word-list-probe-art>"
	fremd := "<word-list-probe-art>Kreatur</word-list-probe-art>"

	for _, wert := range []string{gesehen, fremd} {
		if antwort := anwendung.FeldSetzen(probekontext(t), geoeffnet.SetID, karteID, "text", wert); antwort.Fehler != nil {
			t.Fatalf("Feld setzen: %+v", antwort.Fehler)
		}
	}

	antwort := anwendung.WortWaehlenBedingt(
		probekontext(t), geoeffnet.SetID, karteID, "text", 0, []int{2}, []string{gesehen})

	if antwort.Fehler == nil || antwort.Fehler.Code != CodeKonflikt || antwort.Fehler.Schluessel != domain.SchluesselFremderWert {
		t.Fatalf("erwartet ein Konflikt »fremderWert«, bekommen %+v", antwort.Fehler)
	}

	if nachher := text(); nachher != fremd {
		t.Errorf("im Feld steht %q, erwartet der fremde Wert %q", nachher, fremd)
	}

	// **Und kennt der Entwurf den Wert, wählt sie wie bisher.**
	if antwort := anwendung.WortWaehlenBedingt(
		probekontext(t), geoeffnet.SetID, karteID, "text", 0, []int{2}, []string{gesehen, fremd},
	); antwort.Fehler != nil {
		t.Fatalf("wählen: %+v", antwort.Fehler)
	}

	if nachher := text(); nachher != "<word-list-probe-art>Spruch</word-list-probe-art>" {
		t.Errorf("im Feld steht %q", nachher)
	}
}
