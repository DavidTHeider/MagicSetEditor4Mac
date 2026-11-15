package service_test

// **Eine neue Karte rechnet ihre Werte** — #2516. MSE rechnet beim Anlegen
// jeden Wert der Karte in der Reihenfolge des Spiels (`AddCardAction` in
// `SetScriptManager::onAction`, `src/script/script_manager.cpp`, Zeilen
// 168–178, Fassung f9d9356 vom 2021-07-06), je einmal und ohne Nachzug.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineNeueKarteRechnetJedenWert(t *testing.T) {
	t.Parallel()

	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}).MitNachzug(stile, nil, nil)

	werte := weg.Anfangswerte(t.Context(), 1, nil)

	// **Gespeichert wird nur mit `save value`** — `anzeige` nicht.
	gelesen := make([]string, 0, len(werte))
	for _, wert := range werte {
		gelesen = append(gelesen, wert.Field+"="+wert.Text)
	}

	if want := []string{"eins=[]", "leser=[]!", "fremd=<>"}; !slices.Equal(gelesen, want) {
		t.Errorf("Anfangswerte %q, erwartet %q", gelesen, want)
	}
}

// **Was ein Kartenskript gibt, rechnet mit** — sein Wert ist der Eingang des
// Skripts, wie in MSE der Wert vor `update`.
func TestEineKarteAusEinemSkriptRechnetIhreWerte(t *testing.T) {
	t.Parallel()

	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}).MitNachzug(stile, nil, nil)

	werte := weg.Anfangswerte(t.Context(), 1, []domain.Value{{Field: "eins", Text: "Glut"}, {Field: "name", Text: "Probe"}})

	gelesen := make([]string, 0, len(werte))
	for _, wert := range werte {
		gelesen = append(gelesen, wert.Field+"="+wert.Text)
	}

	if want := []string{"eins=[Glut]", "name=Probe", "leser=[Glut]!", "fremd=<>"}; !slices.Equal(gelesen, want) {
		t.Errorf("Werte %q, erwartet %q", gelesen, want)
	}
}

// **Ohne Vorlage oder ohne Nachzug bleibt die Karte leer**, wie bis #2516.
func TestEineNeueKarteOhneVorlageBleibtLeer(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]*service.Feldweg{
		"ohne Nachzug": service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}),
		"ohne Vorlage": service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}).MitNachzug(&wegstile{}, nil, nil),
		"ohne Skript": service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}).
			MitNachzug(&wegstile{vorlage: service.Vorlage{Kartenfelder: []domain.Field{{Name: "name", Gespeichert: true}}}, gefunden: true}, nil, nil),
	} {
		if werte := weg.Anfangswerte(t.Context(), 1, nil); len(werte) != 0 {
			t.Errorf("%s: %v", name, werte)
		}
	}
}

// **Die Vorschau einer Massenänderung zeigt das Gerechnete** — der Feldweg
// sagt es ihr.
func TestDerFeldwegRechnetFuerDieVorschau(t *testing.T) {
	t.Parallel()

	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}).MitNachzug(stile, nil, nil)

	if gerechnet := weg.Gerechnet(t.Context(), nachzugkarte(), "eins", "Funke"); gerechnet != "[Funke]" {
		t.Errorf("gerechnet %q, erwartet [Funke]", gerechnet)
	}

	if roh := service.NeuerFeldweg(wegkarten{}, wegfelder{}, &wegablage{}).Gerechnet(t.Context(), nachzugkarte(), "eins", "Funke"); roh != "Funke" {
		t.Errorf("ohne Nachzug %q, erwartet Funke", roh)
	}
}
