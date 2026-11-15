package app

// **Die Naht der Regelfreigabe** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wogegen hier geprüft wird
//
// **Dass die Antwort auch bei einem Ja sagt, was nicht geht.** Das Issue
// verlangt es vor dem Start: „Spieler sehen vor Start, welche Regeln
// automatisch durchgesetzt werden."
//
// **Dass jeder der dreizehn Bereiche dasteht** — auch die, über die nichts zugesagt
// ist. Eine Liste, die nur die drei Zusagen zeigte, sähe aus wie eine
// vollständige.
//
// **Und dass der Weg nach der Rolle fragt.**

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Sandkasten geht, und die Antwort nennt trotzdem die Lücken.**
func TestDieFreigabeNenntDieLueckenAuchBeiEinemJa(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	antwort := anwendung.Regelfreigabe(t.Context(), "magic", "sandkasten", 2, "", nil)

	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	freigabe := antwort.Freigabe

	switch {
	case !freigabe.Erlaubt:
		t.Errorf("der Sandkasten geht nicht: %v", freigabe.Hindernisse)
	case freigabe.Regelwerk != "sandkasten":
		t.Errorf("Regelwerk %q", freigabe.Regelwerk)
	case freigabe.Fassung != domain.Grundprofilfassung:
		t.Errorf("Fassung %d", freigabe.Fassung)
	case freigabe.Wirksam != domain.Grundprofilquelle:
		t.Errorf("wirksam ab %q", freigabe.Wirksam)
	case freigabe.Hoechste != string(domain.StrengeBegleitet):
		t.Errorf("höchste Stufe %q", freigabe.Hoechste)
	// **Vier seit `M19`** (#1620):
	// der Commanderbereich ist der dreizehnte. **Drei seit `M9`**
	// (#1611):
	// die Wechselwirkungen sind dazugekommen.
	//
	// **Und zwei seit dem 2026-09-19**
	// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md)):
	// der Commander zählt im Grundprofil nicht mehr mit.
	case len(freigabe.Luecken) != 2:
		t.Errorf("%d Lücken: %v", len(freigabe.Luecken), freigabe.Luecken)
	case len(freigabe.Hindernisse) != 0:
		t.Errorf("Hindernisse: %v", freigabe.Hindernisse)
	}
}

// **Jeder der dreizehn Bereiche steht in der Matrix — auch die stummen.**
//
// **Eine Liste, die nur die drei Zusagen zeigte, sähe aus wie eine
// vollständige.** Das Issue verlangt das Gegenteil: „Unbekannt/ununterstützt
// ist ein eigener Zustand, niemals still erlaubt oder gültig."
func TestJederBereichStehtInDerMatrix(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	matrix := anwendung.Regelfreigabe(
		t.Context(), "magic", "sandkasten", 2, "", nil).Freigabe.Matrix

	if len(matrix) != len(domain.Regelbereiche()) {
		t.Fatalf("%d Einträge, erwartet %d",
			len(matrix), len(domain.Regelbereiche()))
	}

	// **In der Reihenfolge der Bereiche** — sonst liest sie sich bei jedem
	// Aufruf anders.
	for i, bereich := range domain.Regelbereiche() {
		if matrix[i].Bereich != string(bereich) {
			t.Errorf("an Stelle %d steht %q, erwartet %q",
				i, matrix[i].Bereich, bereich)
		}
	}

	zugesagt, begleitet := 0, 0

	for _, eintrag := range matrix {
		// **Eine Quelle ist ab `begleitet` Pflicht** — wer warnt, warnt
		// nach einer Regel, und [domain.Matrixeintrag.Quelle] sagt es so.
		//
		// **Hier stand bis zum 2026-09-17 „ab `automatisch`"**, und das war
		// zu eng: es war damals wahr, weil kein Bereich `begleitet` war.
		// `M2` (#1604)
		// hat `typen` dorthin gehoben, und der Lauf wurde rot mit dem Satz
		// „nennt CR 205" — als wäre die Nummer der Fehler. Ein Wächter, der
		// die halbe Zusage hält, macht die andere unsichtbar (§10).
		if eintrag.Grad == string(domain.UnterstuetzungBegleitet) {
			begleitet++
		}

		if eintrag.Grad == string(domain.UnterstuetzungAutomatisch) {
			zugesagt++
		}

		belegpflichtig := eintrag.Grad == string(domain.UnterstuetzungAutomatisch) ||
			eintrag.Grad == string(domain.UnterstuetzungBegleitet)

		if !belegpflichtig {
			// **Wo nichts zugesagt ist, steht keine Quelle** — eine Nummer
			// neben einer Nichtzusage sähe aus wie ein Beleg.
			if eintrag.Quelle != "" {
				t.Errorf("%q ist %q und nennt %q",
					eintrag.Bereich, eintrag.Grad, eintrag.Quelle)
			}

			continue
		}

		switch {
		case eintrag.Quelle == "":
			t.Errorf("%q ist %q und nennt keine Quelle",
				eintrag.Bereich, eintrag.Grad)
		case len(eintrag.Grenzen) == 0:
			t.Errorf("%q ist %q und nennt keine Grenze",
				eintrag.Bereich, eintrag.Grad)
		}
	}

	// **Fünf seit G9** (#1592) — vorher drei, **sechs seit `M5`**
	// (#1607),
	// **sieben seit `M3`**
	// (#1605):
	// der Manavorrat, und er ist zugleich der zwölfte Bereich (der
	// dreizehnte kam mit `M19` und ist im Grundprofil eine Lücke). **Acht seit
	// `M4`** (#1606):
	// der Stapel. **Neun seit `M6`**
	// (#1608):
	// der Kampf. **Zehn seit `M9`**
	// (#1611):
	// die Wechselwirkungen — eine Fähigkeit löst aus, wartet, kommt auf den
	// Stapel und wird verrechnet, ohne dass jemand etwas stellt.
	if zugesagt != 10 {
		t.Errorf("%d Zusagen, erwartet zehn", zugesagt)
	}

	// **Und eine begleitete seit M2** (#1604): die Typzeile. **Zwei seit
	// `M7`** (#1609):
	// die Handlungen mit Schlüsselwort — sie gehen, wenn jemand sie stellt,
	// und von selbst geschieht keine.
	if begleitet != 2 {
		t.Errorf("%d begleitete Bereiche, erwartet zwei", begleitet)
	}
}

// **Streng wird abgewiesen, und die Antwort trägt trotzdem alles.**
func TestStrengWirdUeberDieNahtAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	freigabe := anwendung.Regelfreigabe(
		t.Context(), "magic", "streng", 2, "", nil).Freigabe

	switch {
	case freigabe.Erlaubt:
		t.Error("streng geht")
	case len(freigabe.Hindernisse) == 0:
		t.Error("kein Grund genannt")
	case len(freigabe.Matrix) == 0:
		t.Error("die Matrix fehlt in der Absage")
	}
}

// **Eine Strenge, die es nicht gibt, ist ein Fehler — und die Listen bleiben
// leer statt `nil`.**
func TestEineErfundeneStrengeGibtEinenFehler(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	antwort := anwendung.Regelfreigabe(t.Context(), "magic", "turnier", 2, "", nil)

	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	switch {
	case antwort.Freigabe.Matrix == nil:
		t.Error("die Matrix ist `nil`")
	case antwort.Freigabe.Luecken == nil:
		t.Error("die Lücken sind `nil`")
	case antwort.Freigabe.Hindernisse == nil:
		t.Error("die Hindernisse sind `nil`")
	}
}

// **Der Weg fragt nach der Rolle.**
func TestDieFreigabeFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	wache := &zaehlendeWache{}

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	antwort := anwendung.MitWache(wache).
		Regelfreigabe(t.Context(), "magic", "sandkasten", 2, "", nil)

	switch {
	case antwort.Fehler == nil:
		t.Fatal("der Weg kam durch")
	case len(wache.gefragt) != 1 || wache.gefragt[0] != domain.WirkungLesen:
		t.Errorf("gefragt wurde nach %v", wache.gefragt)
	}
}

// **Die Fläche ohne Fenster antwortet genauso.**
//
// # Warum sie einen eigenen Lauf bekommt
//
// **Weil `mse`, `mse-mcp` und `mse-rest` über sie gehen** — und weil sie
// **mit** Bestand gebaut wird, obwohl die Rechnung keinen braucht: der
// Riegel vor jedem Weg fragt die Rolle, und die steht im Bestand. Eine
// Fläche ohne ihn ließe jeden durch.
func TestDieRegelflaecheAntwortetGenauso(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	anwendung := RegelflaecheUeber(db)
	anwendung.profil.Store(1)

	freigabe := anwendung.Regelfreigabe(
		t.Context(), "magic", "sandkasten", 2, "", nil).Freigabe

	switch {
	case !freigabe.Erlaubt:
		t.Errorf("der Sandkasten geht nicht: %v", freigabe.Hindernisse)
	case freigabe.Regelwerk != "sandkasten":
		t.Errorf("Regelwerk %q", freigabe.Regelwerk)
	case len(freigabe.Matrix) != len(domain.Regelbereiche()):
		t.Errorf("%d Einträge in der Matrix", len(freigabe.Matrix))
	}
}
