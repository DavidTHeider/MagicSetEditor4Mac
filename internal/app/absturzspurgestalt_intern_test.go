package app

import "testing"

func TestDieSpurFuerDasPaketSchwaerztNurDenFreienWert(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ roh, erwartet string }{
		"ein Laufzeitfehler bleibt": {
			"panic: runtime error: index out of range [5] with length 3\n\ngoroutine 1 [running]:\nmain.lauf()\n",
			"panic: runtime error: index out of range [5] with length 3\n\ngoroutine 1 [running]:\nmain.lauf()\n",
		},
		"ein Satz des Laufzeitsystems bleibt, auch wiederaufgenommen": {
			"panic: assignment to entry in nil map [recovered]\n",
			"panic: assignment to entry in nil map [recovered]\n",
		},
		"ein freier Wert geht, seine Marke bleibt": {
			"panic: das Kennwort hunter2 [recovered, repanicked]\n\ngoroutine 1 [running]:\n",
			"panic: «geschwärzt» [recovered, repanicked]\n\ngoroutine 1 [running]:\n",
		},
		"ein mehrzeiliger Wert geht ganz": {
			"panic: erste Zeile\nzweite Zeile mit hunter2\n\ngoroutine 1 [running]:\n",
			"panic: «geschwärzt»\n«geschwärzt»\n\ngoroutine 1 [running]:\n",
		},
		"eine verschachtelte Panik mit Einzug": {
			"panic: frei [recovered]\n\tpanic: runtime error: invalid memory address or nil pointer dereference\n[signal SIGSEGV: segmentation violation]\n",
			"panic: «geschwärzt» [recovered]\n\tpanic: runtime error: invalid memory address or nil pointer dereference\n[signal SIGSEGV: segmentation violation]\n",
		},
		"ein Wert direkt vor dem Signal": {
			"panic: frei\nnoch frei\n[signal SIGSEGV]\n",
			"panic: «geschwärzt»\n«geschwärzt»\n[signal SIGSEGV]\n",
		},
		// **Was Go so schreibt und sich hier nicht erzeugen lässt** — aus
		// `runtime/traceback.go` und `runtime/panic.go`, Go 1.27.2.
		"die Marken eines Kopfes gehen": {
			"goroutine 1 [running] {token: hunter2, art: \"geheim wert\"}:\nmain.lauf()\n",
			"goroutine 1 [running]:\nmain.lauf()\n",
		},
		"eine Marke an der Fortsetzung bleibt": {
			"panic: a\n\tb hunter2 [recovered]\n\tpanic: c\n",
			"panic: «geschwärzt»\n\t«geschwärzt» [recovered]\n\tpanic: «geschwärzt»\n",
		},
		"eine Typzusicherung bleibt": {
			"panic: interface conversion: interface {} is string, not int\n",
			"panic: interface conversion: interface {} is string, not int\n",
		},
		"ein Signal von außen bleibt": {
			"SIGQUIT: quit\nPC=0x104c1e114 m=0 sigcode=0\n\nr0      0x0\n",
			"SIGQUIT: quit\nPC=0x104c1e114 m=0 sigcode=0\n\nr0      0x0\n",
		},
		"eine Zeile, die Go so nicht schreibt, geht": {
			"goroutine 1 [running]:\nhunter2 steht hier\n",
			"goroutine 1 [running]:\n«geschwärzt»\n",
		},
		"fatal error bleibt": {
			"fatal error: concurrent map writes\n",
			"fatal error: concurrent map writes\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := spurFuerDasPaket(fall.roh); got != fall.erwartet {
				t.Errorf("ergibt\n%q\nerwartet\n%q", got, fall.erwartet)
			}
		})
	}
}
