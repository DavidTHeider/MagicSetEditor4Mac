package service_test

// **Die Absprache über Prozessgrenzen** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Der Riegel im Prozess reicht nicht.** Dieses Programm hat vier Flächen,
// und drei davon sind eigene Prozesse; `mse indexaufbau` weiß nichts vom
// Fenster, und beide schreiben in denselben Bestand.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sperrdoppel gibt vor, ob ein anderer Prozess gerade baut.
type sperrdoppel struct {
	belegt bool
	fehler error

	freigegeben int
}

func (s *sperrdoppel) Nehmen() (freigeben func(), genommen bool, fehler error) {
	if s.fehler != nil {
		return nil, false, s.fehler
	}

	if s.belegt {
		return nil, false, nil
	}

	return func() { s.freigegeben++ }, true, nil
}

// mitSperre baut einen Aufbau über einem Sperrdoppel.
func mitSperre(
	schreiber *schreiberdoppel, sperre *sperrdoppel,
) *aufbaupaar {
	return &aufbaupaar{
		aufbau: service.NeuerIndexaufbau(schreiber,
			service.NeuerHaushalt(domain.Vorgabegrenzen())).
			MitProzesssperre(sperre),
		einbetter: einGutesDoppel(),
	}
}

// **Baut ein anderer Prozess, wird abgewiesen.**
func TestEinAndererBauenderProzessWeistAb(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}
	aufbau := mitSperre(schreiber, &sperrdoppel{belegt: true})

	if _, fehler := aufbau.Aufbauen(t.Context(), einAufbauwunsch(),
		einKorpus()); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if schreiber.naechste != 0 {
		t.Errorf("es wurde %d angelegt", schreiber.naechste)
	}

	// **Und der Riegel im Prozess fällt wieder** — sonst sperrte eine fremde
	// Sperre diesen Prozess für immer, ohne dass er je gebaut hätte.
	if stand := aufbau.Stand(); stand.Laeuft {
		t.Errorf("der Riegel blieb liegen: %+v", stand)
	}
}

// **Eine genommene Sperre wird auch wieder freigegeben.**
func TestEineGenommeneSperreWirdFreigegeben(t *testing.T) {
	t.Parallel()

	sperre := &sperrdoppel{}

	if _, fehler := mitSperre(&schreiberdoppel{}, sperre).Aufbauen(
		t.Context(), einAufbauwunsch(), einKorpus()); fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if sperre.freigegeben != 1 {
		t.Errorf("freigegeben wurde %dmal", sperre.freigegeben)
	}
}

// **Eine Sperre, die gar nicht antwortet, hält den Aufbau auf.**
//
// **Und sie ist ein Fehler und keine Abweisung**: „ein anderer baut" ist eine
// Auskunft, „die Sperrdatei ließ sich nicht öffnen" ist etwas anderes.
func TestEineUnlesbareSperreHaeltDenAufbauAuf(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}

	if _, fehler := mitSperre(schreiber, &sperrdoppel{
		fehler: domain.MitGrund(domain.ErrConflict, nil,
			"die Sperrdatei liegt auf einem Netzlaufwerk"),
	}).Aufbauen(t.Context(), einAufbauwunsch(), einKorpus()); fehler == nil {
		t.Fatal("die unlesbare Sperre kam durch")
	}

	if schreiber.naechste != 0 {
		t.Errorf("es wurde %d angelegt", schreiber.naechste)
	}
}

// **Ohne Sperre baut der Prozess wie bisher.**
//
// **Die Gegenprobe zur Zeile „nil heißt: keine"**: ohne sie sähe niemand,
// dass der Weg ohne Absprache noch offen ist.
func TestOhneProzesssperreBautDerProzessWieBisher(t *testing.T) {
	t.Parallel()

	if _, fehler := einAufbau(&schreiberdoppel{}, einGutesDoppel()).Aufbauen(
		t.Context(), einAufbauwunsch(), einKorpus()); fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}
}

// **Die Verwaltung reicht die Sperre an den Aufbau durch.**
//
// **Die Gegenprobe zur Durchreiche**: ohne sie stünde `MitProzesssperre` an
// der Verwaltung da und wirkte nicht — und die drei Flächen ohne Fenster
// bauten ohne Absprache.
func TestDieVerwaltungReichtDieSperreDurch(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}

	verwaltung := eineVerwaltung(&retrievalablagedoppel{}, schreiber,
		dieQuelleMit("all-minilm"), &korpusoeffnerdoppel{}).
		MitProzesssperre(&sperrdoppel{belegt: true})

	if _, fehler := verwaltung.Aufbauen(t.Context(), service.Aufbauauftrag{
		Modell: "all-minilm",
	}); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if schreiber.naechste != 0 {
		t.Errorf("es wurde %d angelegt", schreiber.naechste)
	}
}
