package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieFaktorsperreVerdoppeltSichBisZurStunde(t *testing.T) {
	for _, fall := range []struct {
		fehlversuche int
		sperre       time.Duration
	}{
		{0, 0},
		{4, 0},
		{5, 30 * time.Second},
		{6, time.Minute},
		{11, 32 * time.Minute},
		{12, time.Hour},
		{100, time.Hour},
	} {
		if sperre := domain.Faktorsperre(fall.fehlversuche); sperre != fall.sperre {
			t.Errorf("nach %d: %v statt %v", fall.fehlversuche, sperre, fall.sperre)
		}
	}
}

func TestEineFaktorsperreNenntIhrEnde(t *testing.T) {
	bis := time.Date(2026, 10, 7, 4, 0, 30, 0, time.UTC)

	var befund *domain.FieldError
	if !errors.As(domain.FaktorGesperrt(bis), &befund) {
		t.Fatal("kein Feldbefund")
	}

	if befund.Field != "faktor" || befund.Key != domain.SchluesselFaktorGesperrt ||
		befund.Werte["bis"] != "2026-10-07T04:00:30Z" {
		t.Errorf("gekommen ist %+v", befund)
	}
}
