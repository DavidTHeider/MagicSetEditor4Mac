// **Was geschieht, wenn das Umschalten nicht glückt** — Werkstück 14 von Phase
// 4.12 (#1351).
//
// # Warum jeder dieser Wege einen Lauf hat
//
// **Weil auf jedem eine Datei halb dasteht.** Ein Zweig, den niemand betreten
// hat, ist eine Behauptung darüber, was im Ernstfall geschieht (§3) — und der
// Ernstfall ist hier ein Bestand, der aufgehen muss.
package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Chiffre sagt die Fläche das, statt zu raten.**
//
// Der Weg gehört zu den Flächen ohne Fenster, die eine ältere Naht bauen —
// oder zu einem Lauf, der sie nicht setzt. Beide sollen eine Auskunft
// bekommen und keinen Absturz.
func TestOhneChiffreWirdAbgewiesen(t *testing.T) {
	weg := service.NeuerBestandsweg(
		service.Sicherungsorte{Datenbank: filepath.Join(t.TempDir(), "b.db")},
		nil, nil, nil,
		func() error { return nil },
		func(context.Context) error { return nil },
		"probe",
	)

	if _, fehler := weg.Verschluesselungsstand(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war eine Abweisung", fehler)
	}

	if _, fehler := weg.VerschluesselungEinschalten(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war eine Abweisung", fehler)
	}
}

// **Ein Rest von einem abgebrochenen Lauf, der sich nicht wegräumen lässt.**
//
// `VACUUM INTO` weigert sich, in eine Datei zu schreiben, die es schon gibt —
// deshalb wird sie vorher weggeräumt. Geht das nicht, ist das die Auskunft.
func TestEinUnraeumbarerRestHaeltDasUmschaltenAuf(t *testing.T) {
	chiffre := &chiffredoppel{}
	weg, bestand := verschluesselungsweg(t, chiffre)

	// **Ein Verzeichnis an der Stelle des Umwegs** lässt sich nicht mit
	// `os.Remove` wegräumen, solange etwas darin liegt.
	umweg := filepath.Join(filepath.Dir(bestand), service.Umschreibname)
	if err := os.Mkdir(umweg, 0o700); err != nil {
		t.Fatalf("Verzeichnis anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(umweg, "drin"), nil, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
		t.Fatal("das Umschalten glückte über einen Rest hinweg")
	}

	if len(chiffre.schritte) != 0 {
		t.Errorf("es wurde gearbeitet: %v", chiffre.schritte)
	}
}

// **Der Schlüssel lässt sich nicht ablegen — dann wird auch nicht
// umgeschrieben.**
//
// Das ist die Regel aus dem Kopf von `VerschluesselungEinschalten` in ihrer
// zweiten Hälfte: kein Augenblick, in dem die Datei verschlüsselt ist und der
// Schlüssel nicht im Bund.
func TestOhneAblegbarenSchluesselWirdNichtUmgeschrieben(t *testing.T) {
	chiffre := &chiffredoppel{scheitertBei: "Schlüssel ablegen"}
	weg, _ := verschluesselungsweg(t, chiffre)

	if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
		t.Fatal("das Umschalten glückte ohne abgelegten Schlüssel")
	}

	if indexVon(chiffre.schritte, "verschlüsseln") >= 0 {
		t.Errorf("es wurde umgeschrieben: %v", chiffre.schritte)
	}
}

// **Und die drei Wege des Tauschs, auf denen die Anwendung ohne Bestand
// dastünde.**
func TestDieFehlerwegeDesTauschs(t *testing.T) {
	t.Run("das Schließen scheitert", func(t *testing.T) {
		chiffre := &chiffredoppel{}
		weg, _ := wegMitLebenszyklus(t, chiffre,
			func() error { return errors.New("die Verbindung bleibt offen") },
			func(context.Context) error { return nil },
			nil)

		if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
			t.Error("der Tausch lief über eine offene Verbindung")
		}
	})

	// **Ist jemand anders am Bestand, wird nichts angefasst — und die
	// Anwendung bekommt ihre Verbindung zurück.**
	t.Run("wir sind nicht allein", func(t *testing.T) {
		chiffre := &chiffredoppel{}
		weg, bestand := wegMitLebenszyklus(t, chiffre,
			func() error { chiffre.schritte = append(chiffre.schritte, "schließen"); return nil },
			func(context.Context) error {
				chiffre.schritte = append(chiffre.schritte, "öffnen")

				return nil
			},
			func(context.Context, string) (func() error, error) {
				return nil, errors.New("da hält noch jemand")
			})

		if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
			t.Fatal("getauscht wurde trotz eines zweiten Verbrauchers")
		}

		inhalt, err := os.ReadFile(bestand) //nolint:gosec // ein Pfad aus t.TempDir()
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		if string(inhalt) != "der bisherige Bestand" {
			t.Errorf("der Bestand wurde angefasst: %q", inhalt)
		}

		// **Und die Anwendung steht nicht ohne da** (§7, Regel 3 in ihrer
		// härtesten Form: ab dem Schließen hat sie keine Datenbank).
		if indexVon(chiffre.schritte, "öffnen") < 0 {
			t.Errorf("nicht wieder geöffnet: %v", chiffre.schritte)
		}
	})

	// **Das Umschreiben meldet Erfolg und legt nichts an.**
	//
	// Die Abwehr dagegen steht seit jeher in `bestandErsetzen`: „Die
	// eingehende Kopie zuerst ansehen, bevor irgendetwas bewegt wird. Ein
	// fehlender Quellpfad darf den Bestand nicht kosten." Ohne diesen Lauf
	// bliebe sie unbetreten — und eine Abwehr, die niemand ausgelöst hat, ist
	// eine Behauptung (§3).
	t.Run("das Umschreiben legt nichts an", func(t *testing.T) {
		chiffre := &chiffredoppel{ohneDatei: true}
		weg, bestand := verschluesselungsweg(t, chiffre)

		if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
			t.Fatal("getauscht wurde gegen eine Datei, die es nicht gibt")
		}

		inhalt, err := os.ReadFile(bestand) //nolint:gosec // ein Pfad aus t.TempDir()
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		if string(inhalt) != "der bisherige Bestand" {
			t.Errorf("der Bestand ging verloren: %q", inhalt)
		}

		if indexVon(chiffre.schritte, "öffnen") < 0 {
			t.Errorf("die Anwendung blieb ohne Bestand: %v", chiffre.schritte)
		}
	})

	// **Das Wiederöffnen scheitert — dann wird der alte Stand zurückgeholt.**
	t.Run("das Wiederöffnen scheitert", func(t *testing.T) {
		chiffre := &chiffredoppel{}
		weg, bestand := wegMitLebenszyklus(t, chiffre,
			func() error { return nil },
			func(context.Context) error { return errors.New("geht nicht mehr auf") },
			nil)

		if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
			t.Fatal("das Umschalten meldete Erfolg")
		}

		inhalt, err := os.ReadFile(bestand) //nolint:gosec // ein Pfad aus t.TempDir()
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		if string(inhalt) != "der bisherige Bestand" {
			t.Errorf("der alte Stand kam nicht zurück: %q", inhalt)
		}
	})
}

// **Und beim Ausschalten: der Bestand steht wieder im Klartext, nur der
// Schlüssel blieb liegen.**
//
// **Das ist kein halber Stand, sondern ein Rest** — und die Meldung sagt genau
// das. Ein Zurückdrehen wäre hier falsch: die Datei ist schon getauscht.
func TestEinLiegengebliebenerSchluesselWirdGemeldetUndNichtVerschwiegen(t *testing.T) {
	chiffre := &chiffredoppel{
		verschluesselt: true, schluessel: "abc",
		scheitertBei: "Schlüssel löschen",
	}
	weg, _ := verschluesselungsweg(t, chiffre)

	_, fehler := weg.VerschluesselungAusschalten(t.Context())
	if fehler == nil {
		t.Fatal("der liegengebliebene Schlüssel wurde verschwiegen")
	}

	if !strings.Contains(fehler.Error(), "Klartext") {
		t.Errorf("die Meldung sagt nicht, was gilt: %v", fehler)
	}
}

// **Scheitert auch das Nachsehen nach einem gescheiterten Tausch, gilt der
// Fehler des Tauschs.**
//
// **Er ist der, der zuerst geschah**, und der, an dem jemand etwas tut. Ein
// Lesefehler darüber wäre die Meldung über die Meldung.
func TestNachEinemGescheitertenTauschGiltDessenFehler(t *testing.T) {
	chiffre := &chiffredoppel{pruefefehler: errors.New("die Datei ist nicht anzusehen")}
	weg, _ := wegMitLebenszyklus(t, chiffre,
		func() error { return errors.New("die Verbindung bleibt offen") },
		func(context.Context) error { return nil },
		nil)

	stand, fehler := weg.VerschluesselungEinschalten(t.Context())
	if fehler == nil {
		t.Fatal("das Umschalten meldete Erfolg")
	}

	if !strings.Contains(fehler.Error(), "schließen") {
		t.Errorf("gemeldet wurde %v, erwartet war der Fehler des Tauschs", fehler)
	}

	// **Und der Stand ist der von vorher** — nicht geraten, sondern der
	// letzte, der wirklich gelesen wurde.
	if stand {
		t.Error("gemeldet wurde „verschlüsselt“, gelesen wurde nichts")
	}
}

// **Die Fehlerwege des Nutzerwechsels** — Werkstück 17 von Phase 4.12.
//
// **Jeder von ihnen lässt einen Menschen als jemand anderes zurück, als er
// glaubt** — oder gar als niemanden. Ein Zweig, den niemand betreten hat, ist
// eine Behauptung darüber, was im Ernstfall geschieht (§3).
func TestDieFehlerwegeDesNutzerwechsels(t *testing.T) {
	t.Run("der Zeiger lässt sich nicht lesen", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.nutzerzeigerfehler = errors.New("die Platte klemmt")

		if fehler := verwaltung.NutzerWechseln(t.Context(), 2, ""); fehler == nil {
			t.Error("der Wechsel glückte über einen Lesefehler hinweg")
		}
	})

	t.Run("ein Nutzer ohne Profil ist kein Ziel", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.profile[3] = nil

		// **Wer dorthin wechselte, stünde ohne Arbeitsumgebung da** — und der
		// Profilzeiger fiele auf ein fremdes zurück.
		if fehler := verwaltung.NutzerWechseln(t.Context(), 3, ""); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet war eine Abweisung", fehler)
		}
	})

	t.Run("die Übersicht klemmt", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.fehler = errors.New("die Übersicht klemmt")

		if fehler := verwaltung.NutzerWechseln(t.Context(), 2, ""); fehler == nil {
			t.Error("der Wechsel glückte über einen Lesefehler hinweg")
		}
	})

	t.Run("der Zeiger lässt sich nicht setzen", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		verwaltung.MitFaktor(&pruefdoppel{})
		ablage.setzfehler = errors.New("die Platte ist voll")

		if fehler := verwaltung.NutzerWechseln(t.Context(), 2, "x"); fehler == nil {
			t.Error("der Wechsel meldete Erfolg")
		}
	})
}

// **Und die Fehlerwege des Profilwechsels, die Werkstück 17 dazugebracht hat.**
func TestDieNeuenFehlerwegeDesProfilwechsels(t *testing.T) {
	t.Run("wer arbeitet, ist nicht zu lesen", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.nutzerzeigerfehler = errors.New("die Platte klemmt")

		if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler == nil {
			t.Error("der Wechsel glückte über einen Lesefehler hinweg")
		}
	})

	// **Und der Profilzeiger getrennt davon** — `Wechseln` liest beide, und
	// zwar den einen vor dem anderen.
	t.Run("das gezeigte Profil ist nicht zu lesen", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.zeigerfehler = errors.New("die Platte klemmt")

		if _, fehler := verwaltung.Wechseln(t.Context(), 2, ""); fehler == nil {
			t.Error("der Wechsel glückte über einen Lesefehler hinweg")
		}
	})

	// **Wer arbeitet, lässt sich nicht festschreiben** — dann findet der
	// Wechsel nicht statt. Ein Zeiger, der sich bewegt, während die Identität
	// stehenbleibt, wäre genau der Zustand, den Werkstück 17 abschafft.
	t.Run("die Identität lässt sich nicht festschreiben", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		verwaltung.MitFaktor(&pruefdoppel{})
		ablage.setzfehler = errors.New("die Platte ist voll")

		if _, fehler := verwaltung.Wechseln(t.Context(), 2, "x"); fehler == nil {
			t.Error("der Wechsel meldete Erfolg")
		}

		if ablage.zeiger != 1 {
			t.Errorf("der Zeiger steht auf %d — er hat sich bewegt", ablage.zeiger)
		}
	})

	// **Und der Zutritt allein reißt ebenso ab.**
	//
	// **Er wird nicht nur vom Wechsel gerufen**: die Deckübernahme fragt
	// dieselbe Erlaubnis (`service.Deckuebernahme`). Über `Wechseln` ist
	// dieser Zweig nicht zu erreichen — der liest den Zeiger vorher.
	t.Run("der Zutritt allein", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.nutzerzeigerfehler = errors.New("die Platte klemmt")

		fehler := verwaltung.Zutritt(t.Context(),
			domain.Profil{ID: 2, NutzerID: 2, Name: "Zweites"}, "")

		if fehler == nil {
			t.Error("der Zutritt glückte über einen Lesefehler hinweg")
		}
	})

	// **Ein Bestand, der niemanden nennt, lässt niemanden durch.**
	//
	// **Eine `0` ist keine Identität, sondern Unwissen** — und ein Weiterreichen
	// wäre eine Erlaubnis daraus (§8). Die Zusage lag bis Werkstück 17 in
	// `Gezeigt` und wäre mit dem eigenen Zeiger fast still verlorengegangen.
	t.Run("der Bestand nennt niemanden", func(t *testing.T) {
		verwaltung, ablage := dreiNutzer(t)
		ablage.profile = map[int64][]domain.Profil{}
		ablage.zeiger = 0

		if _, fehler := verwaltung.WerArbeitet(t.Context()); !errors.Is(
			fehler, domain.ErrNotFound) {
			t.Errorf("Fehler %v, erwartet war „nicht gefunden\"", fehler)
		}
	})
}
