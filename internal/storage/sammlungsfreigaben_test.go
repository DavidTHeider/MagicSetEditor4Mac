package storage_test

// **Freigaben im Bestand** — Werkstück `T3` von Phase 5.6
// (#1468).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// freigabezeit ist der feste Zeitpunkt dieser Läufe.
//
// **Nicht `probezeit`** — den Namen trägt schon das Kostenjournal, und zwei
// Zeitpunkte unter einem Namen sind eine Falle für den nächsten Leser.
var freigabezeit = time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)

// eineFreigabe ist die Freigabe, um die es in diesen Läufen geht.
func eineFreigabe() domain.Sammlungsfreigabe {
	return domain.Sammlungsfreigabe{
		Empfaenger: "peer-1", Umfang: domain.UmfangSammlung,
		Zweck: domain.ZweckHandel, Erteilt: freigabezeit,
	}
}

// **Was erteilt wurde, kommt zurück — Feld für Feld.**
func TestEineFreigabeKommtZurueck(t *testing.T) {
	ablage := storage.NeueFreigabeablage(frischeDatenbank(t))
	ctx := probekontext(t)

	mit := eineFreigabe()
	mit.Umfang, mit.Auswahl = domain.UmfangAuswahl, []string{"abc", "def"}
	mit.Ablauf = freigabezeit.Add(time.Hour)

	if fehler := ablage.Erteilen(ctx, mit); fehler != nil {
		t.Fatalf("erteilen: %v", fehler)
	}

	gelesen, fehler := ablage.Fuer(ctx, "peer-1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 {
		t.Fatalf("es stehen %d Freigaben da", len(gelesen))
	}

	eine := gelesen[0]
	if eine.Umfang != domain.UmfangAuswahl || len(eine.Auswahl) != 2 ||
		!eine.Ablauf.Equal(mit.Ablauf) {
		t.Errorf("zurück kam %+v", eine)
	}
}

// **Eine erfundene Freigabe kommt gar nicht erst hinein.**
//
// **Geprüft wird vor dem Schreiben** — sonst stünde im Bestand ein Umfang,
// den kein Leser deuten kann, und `Zeigt` gäbe für ihn nichts zurück: eine
// Freigabe, die dasteht und nichts tut.
func TestEineErfundeneFreigabeWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueFreigabeablage(frischeDatenbank(t))

	kaputt := eineFreigabe()
	kaputt.Umfang = "alles"

	if fehler := ablage.Erteilen(probekontext(t), kaputt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein zweites Erteilen ersetzt die erste — es legt keine zweite an.**
//
// **Sonst stünden zwei Freigaben desselben Umfangs da**, und welche gilt,
// entschiede die Reihenfolge im Bestand. Wer eine Frist ändert, meint dieselbe
// Freigabe.
func TestEinZweitesErteilenErsetzt(t *testing.T) {
	ablage := storage.NeueFreigabeablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.Erteilen(ctx, eineFreigabe()); fehler != nil {
		t.Fatalf("erteilen: %v", fehler)
	}

	befristet := eineFreigabe()
	befristet.Ablauf = freigabezeit.Add(time.Hour)

	if fehler := ablage.Erteilen(ctx, befristet); fehler != nil {
		t.Fatalf("noch einmal erteilen: %v", fehler)
	}

	gelesen, fehler := ablage.Fuer(ctx, "peer-1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 || !gelesen[0].Ablauf.Equal(befristet.Ablauf) {
		t.Errorf("es stehen %+v da", gelesen)
	}
}

// **Ein Widerruf nimmt genau eine Freigabe weg.**
//
// **Und die Auswahl mit ihr**: an ihr hängt `ON DELETE CASCADE`; bliebe sie
// stehen, hinge sie an einer Kennung, die eine neue Freigabe wiederbekommen
// könnte.
func TestEinWiderrufNimmtGenauEineWeg(t *testing.T) {
	ablage := storage.NeueFreigabeablage(frischeDatenbank(t))
	ctx := probekontext(t)

	wuensche := eineFreigabe()
	wuensche.Umfang = domain.UmfangWuensche

	for _, freigabe := range []domain.Sammlungsfreigabe{eineFreigabe(), wuensche} {
		if fehler := ablage.Erteilen(ctx, freigabe); fehler != nil {
			t.Fatalf("erteilen: %v", fehler)
		}
	}

	if fehler := ablage.Widerrufen(
		ctx, "peer-1", domain.ZweckHandel, domain.UmfangSammlung); fehler != nil {
		t.Fatalf("widerrufen: %v", fehler)
	}

	gelesen, fehler := ablage.Fuer(ctx, "peer-1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 || gelesen[0].Umfang != domain.UmfangWuensche {
		t.Errorf("es stehen %+v da", gelesen)
	}
}

// **Ein Widerruf ins Leere ist ein Nein und kein stiller Erfolg.**
//
// **Sonst sähe jemand „widerrufen" und hätte nichts widerrufen** — etwa, weil
// er sich im Zweck vertan hat.
func TestEinWiderrufInsLeereIstEinNein(t *testing.T) {
	ablage := storage.NeueFreigabeablage(frischeDatenbank(t))

	if fehler := ablage.Widerrufen(probekontext(t), "peer-1",
		domain.ZweckHandel, domain.UmfangSammlung,
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Freigabe gehört einem Profil.**
//
// **Jede Wurzelabfrage klemmt auf eines** (0066): ohne diesen Satz sähe ein
// Gegenüber die Freigaben eines fremden Profils.
func TestEineFreigabeGehoertEinemProfil(t *testing.T) {
	ablage := storage.NeueFreigabeablage(frischeDatenbank(t))

	if fehler := ablage.Erteilen(
		t.Context(), eineFreigabe()); fehler == nil {
		t.Error("das Erteilen kam ohne Profil durch")
	}

	if _, fehler := ablage.Fuer(t.Context(), "peer-1"); fehler == nil {
		t.Error("das Lesen kam ohne Profil durch")
	}
}
