package domain_test

// **Eine Sicht trägt nichts Offenes** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wofür es diesen Wächter gibt
//
// **Weil `service.Partiewirt` sich darauf verlässt.** Er schickt die Sicht als
// Kontrollnachricht an den Gast und prüft dabei **keinen** Verpackungsfehler —
// weil es keinen geben kann: `encoding/json` klagt über Zeichenketten, Zahlen
// und Wahrheitswerte nie. Ein Zweig dafür wäre unerreichbar, und ein
// unerreichbarer Zweig ist eine Zusage, die niemand hält (§10).
//
// **Diese Voraussetzung steht damit hier und nicht in einem Kommentar.** Wer
// der Sicht ein `any` gibt — eine offene Abbildung, eine Schnittstelle —, hebt
// sie auf: dann kann das Verpacken scheitern, und dann braucht es den Zweig
// wieder. Dieser Lauf wird dabei rot, und zwar **bevor** ein Gast vor einem
// Tisch sitzt, der nicht nachzieht.
//
// # Was er nicht hält
//
// **Nicht, dass die Sicht das Richtige zeigt.** Das hält `partiesicht_test.go`.
// Hier geht es allein um die **Gestalt**: was für Typen darin vorkommen
// dürfen.

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was ohne Fehlerzweig verpackt wird, trägt kein offenes Feld.**
//
// **Drei Gestalten und nicht eine.** Zum Tischbild sind mit dem Gespräch am
// Tisch [domain.Chat] und [domain.Praesenz] gekommen: `service.Nebenpruefer`
// und `service.Partierunde` verpacken beide ohne Fehlerzweig, aus demselben
// Grund. Eine Aufzählung, die eine davon vergäße, wäre eine Zusage, die zur
// Hälfte hält (§10).
func TestWasOhneFehlerzweigVerpacktWirdTraegtNichtsOffenes(t *testing.T) {
	t.Parallel()

	gestalten := map[string]reflect.Type{
		"Tischbild": reflect.TypeOf(domain.Tischbild{}),
		"Chat":      reflect.TypeOf(domain.Chat{}),
		"Praesenz":  reflect.TypeOf(domain.Praesenz{}),

		// **Die Wiederaufnahme kam dazu** (G5): beide werden ohne
		// Fehlerzweig verpackt — der Wunsch beim Gast, die Antwort beim Wirt.
		"Wiederaufnahmewunsch": reflect.TypeOf(domain.Wiederaufnahmewunsch{}),
		"Rueckkehr":            reflect.TypeOf(domain.Rueckkehr{}),
	}

	for name, typ := range gestalten {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			offen := offeneFelder(typ, name, map[reflect.Type]bool{})

			if len(offen) != 0 {
				t.Errorf("offene Felder in %s: %v\n"+
					"Sie können das Verpacken scheitern lassen — dann\n"+
					"braucht die Stelle, die es verpackt, ihren\n"+
					"Fehlerzweig zurück.", name, offen)
			}
		})
	}
}

// offeneFelder sucht `any` und Schnittstellen — rekursiv, ohne sich zu
// verlaufen.
//
// **Mit einem Gedächtnis für schon gesehene Typen**: ein Verbund, der sich
// selbst enthält, ließe die Suche sonst nicht zurückkehren.
func offeneFelder(
	typ reflect.Type, pfad string, gesehen map[reflect.Type]bool,
) []string {
	if gesehen[typ] {
		return nil
	}

	gesehen[typ] = true

	switch typ.Kind() {
	case reflect.Interface:
		return []string{pfad}

	case reflect.Slice, reflect.Array, reflect.Pointer:
		return offeneFelder(typ.Elem(), pfad+"[]", gesehen)

	case reflect.Map:
		return append(
			offeneFelder(typ.Key(), pfad+"{schlüssel}", gesehen),
			offeneFelder(typ.Elem(), pfad+"{wert}", gesehen)...)

	case reflect.Struct:
		return offeneFelderImVerbund(typ, pfad, gesehen)
	}

	return nil
}

// offeneFelderImVerbund geht die Felder durch.
func offeneFelderImVerbund(
	typ reflect.Type, pfad string, gesehen map[reflect.Type]bool,
) []string {
	offen := make([]string, 0, typ.NumField())

	for i := range typ.NumField() {
		feld := typ.Field(i)
		offen = append(offen,
			offeneFelder(feld.Type, pfad+"."+feld.Name, gesehen)...)
	}

	return offen
}
