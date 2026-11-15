package service

import (
	"path/filepath"
	"strings"
)

// Der Name, unter dem ein Paket dasteht
// (#636).
//
// **Eine Stelle, weil es eine Frage ist.** Sie wurde zweimal beantwortet: die
// Skriptumgebung fiel auf `full name:` zurück, die Paketwahl nicht. Zwei
// Antworten auf dieselbe Fundstelle sind genau das, was §10 anstreicht.

// Kurzname gibt `short name:` — oder MSEs Rückfall auf den Dateinamen.
//
// # Die Fundstelle, samt ihrer Eigenheit
//
//	void Packaged::validate(Version) {
//	  // a default for the short name
//	  if (short_name.empty()) {
//	    if (!full_name.empty()) short_name = full_name;
//	    short_name = name();
//	  }
//
// `src/util/io/package.cpp`, Zeilen 621–626, Fassung f9d9356 vom 2021-07-06.
//
// **Die mittlere Zeile ist wirkungslos.** Sie wird von der nächsten sofort
// überschrieben; ein `full name:` ohne `short name:` erreicht `short_name`
// damit **nie**. Portiert ist, was **geschieht** (§15.4).
//
// **`name()` ist der Dateiname ohne Endung** (`package.cpp`, Zeilen 43–53):
// `magic-m15`, nicht `magic-m15.mse-style`.
//
// # Warum die Treue hier zählt
//
// `stylesheet.short_name` steht in der **Skriptumgebung**, und ein Feldskript
// vergleicht darauf. Wer dort etwas anderes einsetzt als MSE, lässt es einen
// anderen Zweig nehmen. Wer den vollen Namen sehen will, liest `full_name` —
// der steht in derselben Umgebung unter eigenem Namen.
//
// **Über `full-magic-pack` greift der Rückfall nirgends**: alle 376
// Stilpakete erklären ein `short name:`. Die Regel gilt trotzdem — für jedes
// Paket, das keines hat.
func Kurzname(kurz, dateiname string) string {
	if name := strings.TrimSpace(kurz); name != "" {
		return name
	}

	return ohneEndung(dateiname)
}

// ohneEndung schneidet die Paketendung ab — `name()` bei MSE.
//
// **Nur die letzte.** `magic-m15.mse-style` wird `magic-m15`; ein Punkt im
// Namen selbst bleibt stehen, wie `find_last_of('.')` es tut.
func ohneEndung(dateiname string) string {
	return strings.TrimSuffix(dateiname, filepath.Ext(dateiname))
}
