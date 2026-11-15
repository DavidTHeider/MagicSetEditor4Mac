package storage_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Freigabe des Neuanfangs zählt nicht zu dem, was er wegnimmt** — G12
// (#2282).
//
// Gezählt, änderte das Erteilen der Freigabe die Vorschau und damit ihre
// Revision — und die Ausführung, die diese Revision vorzeigt, wies sich
// selbst ab. Gefunden im ersten Lauf am nativen Fenster (2026-10-10). **Eine
// andere Freigabe zählt weiter**: sie gehört dem Profil und geht mit ihm.
func TestDieFreigabeDesNeuanfangsZaehltNichtZumUmfang(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)
	ablage := storage.NeueNachweisablage(db)
	jetzt := time.Now()

	for _, nachweis := range []domain.Nachweis{
		{Kennung: "neuanfang", Zweck: domain.RuecksetzungsZweck, Ziel: string(domain.RuecksetzungAnwendung)},
		{Kennung: "handel", Zweck: "handel.verbuchen", Ziel: "vorgang"},
	} {
		nachweis.NutzerID, nachweis.ProfilID = 1, 1
		nachweis.Ausgestellt, nachweis.Ablauf = jetzt, jetzt.Add(time.Hour)

		if fehler := ablage.NachweisAnlegen(ctx, nachweis); fehler != nil {
			t.Fatalf("den Nachweis %q legen: %v", nachweis.Kennung, fehler)
		}
	}

	betroffen, fehler := storage.Profilumfang(ctx, db, 1)
	if fehler != nil {
		t.Fatalf("der Umfang: %v", fehler)
	}

	for _, gruppe := range betroffen {
		if gruppe.Was != domain.BetroffenFreigaben {
			continue
		}

		if gruppe.Anzahl != 1 || len(gruppe.Namen) != 1 || gruppe.Namen[0] != "handel.verbuchen" {
			t.Errorf("die Freigaben sind %+v — erwartet nur die des Handels", gruppe)
		}

		return
	}

	t.Errorf("die Freigabe des Handels steht in keiner Gruppe: %+v", betroffen)
}
