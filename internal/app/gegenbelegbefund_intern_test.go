package app

// **Ein abgewiesener Gegenbeleg steht am Beleg** — an jeder Fläche, die den
// Beleg zeigt (#2500).
// Das Fenster und MCP lesen dieses DTO; Kommandozeile und REST geben es in
// ihrer eigenen Gestalt aus.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinBelegOhneBefundNenntKeinen(t *testing.T) {
	if dto := alsHandelsbelegDTO(domain.Belegstand{Vorgang: "v-1"}); dto.Befund != nil {
		t.Errorf("ohne Befund steht %+v", *dto.Befund)
	}
}

func TestEinBefundStehtMitMangelUndUhr(t *testing.T) {
	empfangen := time.Date(2026, 10, 19, 21, 30, 0, 0, time.FixedZone("MESZ", 2*3600))

	dto := alsHandelsbelegDTO(domain.Belegstand{Vorgang: "v-1", Befund: &domain.Gegenbelegbefund{
		Vorgang: "v-1", Von: "peer-bert", Abdruck: "a1", Mangel: domain.MangelZeilen,
		Grund: "zwei Zeilen statt drei", Empfangen: empfangen,
	}})

	if dto.Befund == nil {
		t.Fatal("der Befund fehlt")
	}

	erwartet := GegenbelegbefundDTO{
		Von: "peer-bert", Abdruck: "a1", Mangel: "zeilen",
		Grund: "zwei Zeilen statt drei", Empfangen: "2026-10-19T19:30:00Z",
	}
	if *dto.Befund != erwartet {
		t.Errorf("der Befund ist %+v, erwartet %+v", *dto.Befund, erwartet)
	}
}
