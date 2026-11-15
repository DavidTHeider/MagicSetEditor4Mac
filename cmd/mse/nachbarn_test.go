package main

// **Ein bekannter Nachbar an einer neuen Adresse, an der Kommandozeile** —
// NET-02.NAT-02 (#1706).
//
// Der Flächenlauf kann nicht steuern, wen mDNS findet; hier steht die
// Gestalt der Ausgabe gegen eine gebaute Antwort.

import (
	"bytes"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieNachbarnNennenEineAdresseZumUebernehmen(t *testing.T) {
	var raus, rfehler bytes.Buffer

	an := &schreiber{ziel: &raus, fehler: &rfehler}

	code := nachbarnSchreiben(an, app.NachbarnAntwort{Nachbarn: []app.Nachbar{
		{
			Kennung: "bumgezogen", Fingerabdruck: "AB-CD", Adresse: "10.0.0.8:41234",
			Merkadresse: "annas-neuer.local:41234", Bekannt: true, Stand: "bestaetigt",
			GemerkteAdresse: "annas-mac.local:41234", AdresseGeaendert: true,
		},
		{
			Kennung: "bohne", Fingerabdruck: "EF-GH", Adresse: "10.0.0.9:41234",
			Merkadresse: "10.0.0.9:41234", Bekannt: true, Stand: "unbestaetigt",
			AdresseGeaendert: true,
		},
		{
			Kennung: "bfremd", Fingerabdruck: "IJ-KL", Adresse: "10.0.0.10:41234",
			Merkadresse: "10.0.0.10:41234",
		},
	}, Verworfen: 1})
	if code != 0 {
		t.Fatalf("Code %d: %s", code, rfehler.String())
	}

	erwartet := "bestaetigt\tAB-CD\t10.0.0.8:41234\tannas-neuer.local:41234\t\n" +
		"\tadresse-bisher\tannas-mac.local:41234\n" +
		"unbestaetigt\tEF-GH\t10.0.0.9:41234\t-\t\n" +
		"\tadresse-bisher\t-\n" +
		"unbekannt\tIJ-KL\t10.0.0.10:41234\t-\t\n" +
		"verworfen\t1\n"

	if raus.String() != erwartet {
		t.Errorf("Ausgabe:\n%q\nerwartet:\n%q", raus.String(), erwartet)
	}
}
