package service_test

// **Die Herkunftsart eines Pakets, ohne die ganze Übersicht** — Zweitblick
// über Phase B, Befund 2: scheitert das Nachlesen der Übersicht nach einem
// Schreibweg, soll die Antwort trotzdem sagen, woher das Paket stammt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieHerkunftsartHaengtAmPaketUndAmVerwaltetenVerzeichnis(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerPaketverwalter(nil, "/verwaltet")

	faelle := []struct {
		name  string
		paket domain.Paket
		will  domain.Paketherkunftsart
	}{
		{"im verwalteten Verzeichnis", domain.Paket{Pfad: "/verwaltet/eigen.mse-game"}, domain.PaketEigen},
		{"mit Prüfsumme bezogen", domain.Paket{Pfad: "/verwaltet/b.mse-game", Pruefsumme: "abc"}, domain.PaketBezogen},
		{"mit Herkunft übernommen", domain.Paket{Pfad: "/verwaltet/u.mse-game", Herkunft: "/mse2"}, domain.PaketUebernommen},
		{"daneben mitgeliefert", domain.Paket{Pfad: "/buendel/saat/probe.mse-game"}, domain.PaketMitgeliefert},
	}

	for _, fall := range faelle {
		if ist := verwalter.Herkunftsart(fall.paket); ist != fall.will {
			t.Errorf("%s: %q, erwartet %q", fall.name, ist, fall.will)
		}
	}
}
