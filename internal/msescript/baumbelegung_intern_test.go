package msescript

// **Was ein gemerkter Baum belegt, und die Grenze darüber** — PERF03
// (#2353),
// Befund A-19.
//
// **Ohne `t.Parallel()`**: der Baumspeicher gehört dem Prozess, und der Lauf
// setzt seine Grenze herab. Parallele Läufe warten, bis die übrigen fertig
// sind; sie sähen sonst einen Speicher, der ständig räumt.

import (
	"reflect"
	"strings"
	"testing"
)

func TestBaumBytesZaehltKnotenTexteUndSchluessel(t *testing.T) {
	blatt := &Knoten{Form: FormName, Text: "abc"}
	wurzel := &Knoten{Form: FormAufruf, Text: "f", Kinder: []*Knoten{blatt, nil}, Namen: []string{"", "wert"}}

	knoten := int64(reflect.TypeFor[Knoten]().Size())
	zeiger := int64(reflect.TypeFor[*Knoten]().Size())
	zeichenkette := int64(reflect.TypeFor[string]().Size())

	erwartet := int64(len("quelle")) + 2*knoten + int64(len("abc")+len("f")) +
		2*zeiger + 2*zeichenkette + int64(len("wert"))

	if bytes := baumBytes("quelle", wurzel); bytes != erwartet {
		t.Errorf("%d Bytes, erwartet %d", bytes, erwartet)
	}
}

func TestDerBaumspeicherRaeumtAnDerBytegrenze(t *testing.T) {
	vorher := baumgrenzeBytes

	t.Cleanup(func() {
		baumgrenzeBytes = vorher

		BaumspeicherLeeren()
	})

	BaumspeicherLeeren()

	baum := &Knoten{Form: FormName, Text: "abc"}
	einer := baumBytes("q1", baum)
	baumgrenzeBytes = 2*einer + einer/2

	merkeBaum("q1", baum)
	merkeBaum("q2", baum)
	merkeBaum("q2", baum)

	if zahlen := Baumzahlen(); zahlen.Eintraege != 2 || zahlen.Bytes != 2*einer || zahlen.GrenzeBytes != baumgrenzeBytes {
		t.Fatalf("zwei Bäume: %+v (einer belegt %d)", zahlen, einer)
	}

	merkeBaum("q3", baum)

	if zahlen := Baumzahlen(); zahlen.Eintraege != 1 || zahlen.Bytes != einer {
		t.Errorf("über der Grenze nicht ganz geräumt: %+v", zahlen)
	}

	merkeBaum("riesig", &Knoten{Form: FormZeichenkette, Text: strings.Repeat("x", int(baumgrenzeBytes))})

	if _, da := gemerkterBaum("riesig"); da {
		t.Error("ein Baum über der ganzen Grenze wurde gemerkt")
	}

	BaumspeicherLeeren()

	if zahlen := Baumzahlen(); zahlen.Eintraege != 0 || zahlen.Bytes != 0 {
		t.Errorf("nach dem Leeren: %+v", zahlen)
	}
}

// **Die Grenze im Betrieb** — 64 MiB, PERF03.
func TestDieBytegrenzeDesBaumspeichers(t *testing.T) {
	if GrenzeBaumbytes != 64<<20 || baumgrenzeBytes != GrenzeBaumbytes {
		t.Errorf("Grenze %d, im Betrieb %d", GrenzeBaumbytes, baumgrenzeBytes)
	}
}
