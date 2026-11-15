package service_test

// **Der Export personenbezogener Daten** — Werkstück 8 von Phase 4.12
// (#1351).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Export nennt alles, was über einen Menschen dasteht.**
func TestDerExportNenntAllesUeberDenMenschen(t *testing.T) {
	t.Parallel()

	daten, fehler := service.NeuePersonenauskunft(vollePersonenquelle()).
		MitUhr(func() time.Time { return probezeit }).
		Personendaten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("erheben: %v", fehler)
	}

	if !daten.Erhoben.Equal(probezeit) {
		t.Errorf("erhoben am %v", daten.Erhoben)
	}

	if daten.Nutzer.Vorname != "Ada" || daten.Nutzer.Nachname != "Lovelace" {
		t.Errorf("die Angaben sind %+v", daten.Nutzer)
	}

	if len(daten.Profile) != 1 || daten.Profile[0].Name != "Mein Profil" {
		t.Errorf("die Profile sind %+v", daten.Profile)
	}

	if len(daten.Konten) != 1 || daten.Konten[0].Anbieter != "github" {
		t.Errorf("die Konten sind %+v", daten.Konten)
	}

	if len(daten.Faktoren) != 1 || daten.Faktoren[0] != string(domain.FaktorTOTP) {
		t.Errorf("die Faktoren sind %+v", daten.Faktoren)
	}

	if daten.Einstellungen[domain.NutzerKontrast] != domain.KontrastHoch {
		t.Errorf("die Einstellungen sind %+v", daten.Einstellungen)
	}
}

// **Und nennt weder Kennwort noch Geheimnis.**
//
// # Warum das die wichtigste Zusage dieses Exports ist
//
// **Weil beides kein Datum über einen Menschen ist, sondern ein Mittel, ihn
// zu sein.** Eine Datei, mit der jemand anderes sich anmeldet, hilft der
// Auskunft nichts — und wer sie weitergibt, gibt sein Konto weg.
//
// **Geprüft über JSON und nicht über die Struktur** (§3): zwischen beiden
// liegen die Marken, und ein Feld ohne `json:"-"` steht in der Ausgabe, gleich
// wie es im Go-Typ heißt.
func TestDerExportNenntWederKennwortNochGeheimnis(t *testing.T) {
	t.Parallel()

	daten, fehler := service.NeuePersonenauskunft(vollePersonenquelle()).
		Personendaten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("erheben: %v", fehler)
	}

	roh, marshalfehler := json.Marshal(daten)
	if marshalfehler != nil {
		t.Fatalf("marshal: %v", marshalfehler)
	}

	// **Die Werte und nicht das Wort**: `hatKennwort` **soll** dastehen — es
	// ist die Auskunft. Verboten ist, was jemand benutzen könnte.
	for _, verboten := range []string{"argon2", "$", "GEZDGNBV"} {
		if strings.Contains(string(roh), verboten) {
			t.Errorf("%q steht im Export: %s", verboten, roh)
		}
	}

	// **Dass es ein Kennwort gibt, steht drin** — die Auskunft ist die
	// Tatsache, nicht der Wert.
	if !daten.Nutzer.HatKennwort {
		t.Error("der Export verschweigt, dass ein Kennwort gesetzt ist")
	}
}

// **Ein Nutzer, den es nicht gibt, bekommt keinen Export.**
func TestEinNutzerDenEsNichtGibtBekommtKeinenExport(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuePersonenauskunft(vollePersonenquelle()).
		Personendaten(t.Context(), 4711)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Lesefehler reißt ab und wird nicht verschluckt.**
//
// **Eine halbe Auskunft sieht aus wie eine ganze** — und genau das darf eine
// Auskunft über personenbezogene Daten nicht (§8). Jeder der fünf Wege
// einzeln, weil jeder seinen eigenen `if fehler != nil` hat.
func TestJederLesefehlerImExportReisstAb(t *testing.T) {
	t.Parallel()

	for name := range map[string]struct{}{
		"nutzer": {}, "profile": {}, "konten": {}, "faktoren": {}, "einstellungen": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle := vollePersonenquelle()
			quelle.kaputtAn = name

			_, fehler := service.NeuePersonenauskunft(quelle).
				Personendaten(t.Context(), 1)
			if !errors.Is(fehler, errStummerBestand) {
				t.Errorf("durchgelassen wurde: %v", fehler)
			}
		})
	}
}

// **Ein Mensch ohne alles bekommt Listen und keine Nullen** (§3).
func TestEinMenschOhneAllesBekommtListen(t *testing.T) {
	t.Parallel()

	daten, fehler := service.NeuePersonenauskunft(&personendoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Ich"}},
	}).Personendaten(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("erheben: %v", fehler)
	}

	if daten.Profile == nil || daten.Konten == nil || daten.Faktoren == nil {
		t.Errorf("eine Liste ist nil: %+v", daten)
	}
}

// vollePersonenquelle gibt eine Quelle mit allem.
func vollePersonenquelle() *personendoppel {
	return &personendoppel{
		nutzer: []domain.Nutzer{{
			ID: 1, Name: "Ich", Vorname: "Ada", Nachname: "Lovelace",
			Kennwort: "$argon2id$v=19$m=19456,t=2,p=1$c2FsemNoZW4$aGFzaA",
		}},
		profile:  []domain.Profil{{ID: 10, NutzerID: 1, Name: "Mein Profil"}},
		konten:   []domain.Anbieterkonto{{Anbieter: "github", Anzeigename: "ada"}},
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
		einstellungen: map[string]string{
			domain.NutzerKontrast: domain.KontrastHoch,
		},
	}
}

// personendoppel ist die Quelle im Arbeitsspeicher.
type personendoppel struct {
	nutzer        []domain.Nutzer
	profile       []domain.Profil
	konten        []domain.Anbieterkonto
	faktoren      []domain.Faktor
	einstellungen map[string]string

	// kaputtAn nennt den **einen** Weg, der scheitert.
	//
	// **Einen und nicht alle**: wer alles kaputt macht, betritt den zweiten
	// Weg nie und prüft den zweiten `if` nicht.
	kaputtAn string
}

func (p *personendoppel) Nutzer(context.Context) ([]domain.Nutzer, error) {
	if p.kaputtAn == "nutzer" {
		return nil, errStummerBestand
	}

	return p.nutzer, nil
}

func (p *personendoppel) Profile(context.Context, int64) ([]domain.Profil, error) {
	if p.kaputtAn == "profile" {
		return nil, errStummerBestand
	}

	return p.profile, nil
}

func (p *personendoppel) Konten(context.Context, int64) ([]domain.Anbieterkonto, error) {
	if p.kaputtAn == "konten" {
		return nil, errStummerBestand
	}

	return p.konten, nil
}

func (p *personendoppel) Faktoren(context.Context, int64) ([]domain.Faktor, error) {
	if p.kaputtAn == "faktoren" {
		return nil, errStummerBestand
	}

	return p.faktoren, nil
}

func (p *personendoppel) Nutzereinstellungen(context.Context, int64) (
	map[string]string, error,
) {
	if p.kaputtAn == "einstellungen" {
		return nil, errStummerBestand
	}

	return p.einstellungen, nil
}
