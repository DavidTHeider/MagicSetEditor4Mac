package domain_test

// **Ein KI-Modell und was es belegt** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einModell baut ein gültiges Modell mit einer belegten Fähigkeit.
func einModell() domain.KIModell {
	return domain.KIModell{
		Alias:    "beispiel:3b",
		Anbieter: "ollama",
		Ort:      domain.KIOrtLokal,
		Laufzeit: "ollama 0.12.3",
		Geprueft: []domain.KIGeprueft{{
			Faehigkeit: domain.KIEinbettung,
			Datum:      "2026-09-11",
			Beleg:      "eigener Lauf gegen die Laufzeit",
		}},
		Herkunft: eineKette(),
	}
}

// **Ein Modell kann, was es belegt — und sonst nichts.**
//
// Das Issue: „Modelllisten garantieren keine Capability." Ein Modell, das in
// einer Liste steht, kann trotzdem nichts; gefragt wird der **Beleg**.
func TestEinModellKannNurWasEsBelegt(t *testing.T) {
	t.Parallel()

	modell := einModell()

	if !modell.Kann(domain.KIEinbettung) {
		t.Error("die belegte Fähigkeit gilt nicht")
	}

	for _, andere := range []domain.KIFaehigkeit{
		domain.KITextgenerierung, domain.KIVision, domain.KINeuordnung,
	} {
		if modell.Kann(andere) {
			t.Errorf("%q gilt ohne Beleg als vorhanden", andere)
		}
	}
}

// **Ein Modell ohne Alias, Anbieter oder Laufzeit wird abgewiesen.**
func TestEinUnvollstaendigesModellWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.KIModell){
		"ohne Alias":         func(m *domain.KIModell) { m.Alias = "" },
		"ohne Anbieter":      func(m *domain.KIModell) { m.Anbieter = "" },
		"ohne Laufzeit":      func(m *domain.KIModell) { m.Laufzeit = "" },
		"mit erfundenem Ort": func(m *domain.KIModell) { m.Ort = "irgendwo" },
		"mit einer ungeprüften Fähigkeit": func(m *domain.KIModell) {
			m.Geprueft[0].Beleg = ""
		},
		"mit einer kaputten Herkunft": func(m *domain.KIModell) {
			m.Herkunft.Glieder[0].Lizenz = ""
		},
		"mit einem Einbettungsvertrag ohne Dimension": func(m *domain.KIModell) {
			m.Einbettung = &domain.KIEinbettungsvertrag{Pooling: domain.KIPoolingMittel}
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			modell := einModell()
			kaputt(&modell)

			if befunde := modell.Validate(); len(befunde) == 0 {
				t.Errorf("%s kam durch", name)
			}
		})
	}
}

// **Ein Modell, das einbettet, braucht seinen Vertrag.**
//
// **Und eines, das nicht einbettet, darf keinen haben.** Beides zusammen ist
// die Zusage: wer einen Vektor bekommt, hat einen Vertrag, an dem er die Form
// prüfen kann — und wer keinen Vektor liefert, verspricht auch keine Form.
func TestEinEinbettungsmodellBrauchtSeinenVertrag(t *testing.T) {
	t.Parallel()

	ohne := einModell()

	if befunde := ohne.Validate(); len(befunde) == 0 {
		t.Error("ein Einbettungsmodell ohne Vertrag kam durch")
	}

	mit := einModell()
	mit.Einbettung = &domain.KIEinbettungsvertrag{
		Dimension: 768, Pooling: domain.KIPoolingMittel, Zerteiler: "x",
	}

	if befunde := mit.Validate(); len(befunde) != 0 {
		t.Errorf("ein Einbettungsmodell mit Vertrag wurde abgewiesen: %v", befunde)
	}

	// **Und umgekehrt**: ein Textmodell mit Einbettungsvertrag ist ein
	// Versprechen ohne Rumpf.
	fremd := einModell()
	fremd.Geprueft[0].Faehigkeit = domain.KITextgenerierung
	fremd.Einbettung = &domain.KIEinbettungsvertrag{
		Dimension: 768, Pooling: domain.KIPoolingMittel, Zerteiler: "x",
	}

	if befunde := fremd.Validate(); len(befunde) == 0 {
		t.Error("ein Textmodell mit Einbettungsvertrag kam durch")
	}
}

// **Die Bindung macht einen Aliaswechsel sichtbar.**
//
// # Worum es geht
//
// **Ein Alias zeigt nicht immer auf dasselbe.** `beispiel:3b` ist ein Name,
// den ein Anbieter vergibt — und hinter dem er austauschen kann, was er will.
// Das Issue verlangt: „Modellaliasänderung **sichtbar**."
//
// Die Bindung ist Alias **und** ausgeführtes Artefakt. Bleibt sie gleich, ist
// es dasselbe; ändert sie sich, hat jemand etwas getauscht.
func TestDieBindungMachtEinenAliaswechselSichtbar(t *testing.T) {
	t.Parallel()

	grund := einModell().Bindung()

	getauscht := einModell()
	getauscht.Herkunft.Glieder[3].Revision = "sha256:eeee"

	if getauscht.Bindung() == grund {
		t.Error("ein getauschtes Artefakt ändert die Bindung nicht")
	}

	umbenannt := einModell()
	umbenannt.Alias = "beispiel:7b"

	if umbenannt.Bindung() == grund {
		t.Error("ein anderer Alias ändert die Bindung nicht")
	}

	if einModell().Bindung() != grund {
		t.Error("dasselbe Modell gibt zwei Bindungen")
	}
}

// **Ohne Herkunft ist die Bindung »unbekannt« — und sagt das.**
//
// **Nicht leer und nicht geraten**: eine leere Bindung sähe aus wie „nichts
// gebunden", und eine aus dem Alias abgeleitete wäre die erfundene
// Herkunftsangabe, die 0096 und das Issue verbieten.
func TestOhneHerkunftIstDieBindungUnbekannt(t *testing.T) {
	t.Parallel()

	ohne := einModell()
	ohne.Herkunft = domain.KIHerkunft{}

	bindung := ohne.Bindung()

	if bindung == "" {
		t.Fatal("die Bindung ist leer")
	}

	if bindung == einModell().Bindung() {
		t.Error("eine unbekannte Herkunft gibt dieselbe Bindung wie eine bekannte")
	}
}
