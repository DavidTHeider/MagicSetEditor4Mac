package domain_test

// **Die Mischpflicht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum sie im Zustand steht und nicht nur im Strom
//
// **Weil sie eine Tatsache über den Tisch ist und keine über die
// Vergangenheit.** Dass Anna ihre Bibliothek angesehen hat, steht im
// Ereignisstrom; **dass sie seitdem nicht gemischt hat**, müsste jeder
// Empfänger sonst aus dem Strom nachrechnen — und ein Zuschauer, der später
// dazukommt, hat den Strom gar nicht.
//
// **Erzwungen wird sie hier nicht.** Der Sandkasten gehorcht dem Menschen;
// **dass** nach einer Suche gemischt werden **muss**, ist Werkstück G4
// (#1461).
// Was G1 schuldet, ist der Platz, an dem G4 es ablesen kann.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestMischpflichtEntstehtUndVergeht(t *testing.T) {
	t.Parallel()

	partie := domain.Partie{}

	if partie.SchuldetMischen("anna", domain.ZoneBibliothek) {
		t.Fatal("eine frische Partie schuldet schon ein Mischen")
	}

	partie.MerkeMischpflicht("anna", domain.ZoneBibliothek)

	if !partie.SchuldetMischen("anna", domain.ZoneBibliothek) {
		t.Error("nach dem Ansehen steht keine Mischpflicht")
	}

	// **Sie gehört einem Sitz und einer Zone** — Annas Blick verpflichtet
	// nicht Bert, und ihre Bibliothek nicht ihren Friedhof.
	if partie.SchuldetMischen("bert", domain.ZoneBibliothek) {
		t.Error("Annas Blick verpflichtet Bert")
	}

	if partie.SchuldetMischen("anna", domain.ZoneFriedhof) {
		t.Error("ein Blick in die Bibliothek verpflichtet den Friedhof")
	}

	// **Zweimal ansehen ist eine Pflicht und nicht zwei** — sonst stünde
	// dieselbe Zone doppelt da, und ein Mischen löschte nur die Hälfte.
	partie.MerkeMischpflicht("anna", domain.ZoneBibliothek)

	if anzahl := len(partie.Mischpflichten); anzahl != 1 {
		t.Errorf("%d Mischpflichten nach zwei Blicken, erwartet eine", anzahl)
	}

	// **Und eine zweite Pflicht daneben** — sie muss stehen bleiben, wenn
	// die erste gelöst wird. Ohne sie prüfte der Lauf nur, dass die Liste
	// leer wird, und nicht, dass **die richtige** Zeile verschwindet.
	partie.MerkeMischpflicht("bert", domain.ZoneBibliothek)

	partie.LoeseMischpflicht("anna", domain.ZoneBibliothek)

	if partie.SchuldetMischen("anna", domain.ZoneBibliothek) {
		t.Error("das Mischen hat die Pflicht nicht gelöst")
	}

	if !partie.SchuldetMischen("bert", domain.ZoneBibliothek) {
		t.Error("Annas Mischen hat auch Berts Pflicht gelöst")
	}

	// **Und ein Mischen ohne Pflicht ist kein Fehler.** Wer mischt, ohne
	// gesehen zu haben, darf das — es ist die Richtung, in der ein Irrtum
	// nichts kostet.
	partie.LoeseMischpflicht("bert", domain.ZoneBibliothek)
}

// **Und die Sicht nennt sie — jedem.**
//
// **Eine Mischpflicht ist öffentlich**, und das ist ihr ganzer Zweck: sie
// sagt dem Gegenüber, dass jemand seine Bibliothek gesehen und noch nicht
// gemischt hat. Stünde sie nur beim Betroffenen, wäre sie eine Notiz an sich
// selbst.
func TestDieSichtNenntJedemDieMischpflicht(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.MerkeMischpflicht("anna", domain.ZoneBibliothek)

	for _, wer := range []string{"anna", "bert", "", "zuschauer"} {
		sicht := partie.Sicht(wer)

		if len(sicht.Mischpflichten) != 1 {
			t.Errorf("%q sieht %d Mischpflichten, erwartet eine",
				wer, len(sicht.Mischpflichten))

			continue
		}

		if sicht.Mischpflichten[0].Sitz != "anna" ||
			sicht.Mischpflichten[0].Zone != domain.ZoneBibliothek {
			t.Errorf("%q sieht %+v", wer, sicht.Mischpflichten[0])
		}
	}
}
