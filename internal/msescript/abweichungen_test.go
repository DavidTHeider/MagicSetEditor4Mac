package msescript_test

import "testing"

// **`to_int` nimmt eine breitere Zahlensyntax an als MSE** — und das ist
// gemessen folgenlos
// (#271).
//
// # Was hier festgehalten wird
//
// Nicht ein neues Verhalten, sondern ein **entschiedenes**: die Abweichung
// bleibt stehen, weil sie im Pack nicht eintritt. Ohne diesen Test wäre die
// Entscheidung ein Satz in einer Datei, und der nächste Umbau nähme sie
// stillschweigend mit.
//
// # Die Messung dahinter
//
// Am Orakel gefragt (`magicseteditor --cli … -q --raw`, gegen `White.mse-set`),
// am 2026-08-20:
//
//	Eingabe     MSE2                       hier
//	"1.5"       Fehler                     1
//	"1e3"       Fehler                     1000
//	" 12 "      Fehler                     12
//	"1_000"     Fehler                     1000
//	"Inf"       Fehler                     inf
//	"NaN"       Fehler                     nan
//	"0x1p4"     Fehler                     16
//	"0x10"      Fehler                     Fehler   — gleich
//	"12abc"     Fehler                     Fehler   — gleich
//	"+12"       12                         12       — gleich
//	"-3"        −3                         −3       — gleich
//
// **MSE benutzt `String::ToLong`** (`src/script/functions/basic.cpp`, Zeilen
// 116–124, Fassung f9d9356 vom 2021-07-06), dieses Programm
// `strconv.ParseFloat` samt `math.Trunc`. Damit
// nimmt es Exponenten, umgebende Leerzeichen, Gos Unterstriche und
// **`Inf`** an, wo MSE ablehnt.
//
// # Hier stand bis zum 2026-08-20 „nur bei einer Kommazahl"
//
// **Zu eng, und der Zweitblick über Phase K hat es angestrichen** — mit
// Recht: geprüft war ein Fall, benannt war eine Klasse. Die Klasse steht jetzt
// oben, am Orakel gefragt.
//
// **Und `"Inf"` ist der unangenehmste darin.** Eine Unendlichkeit, die in eine
// Maßangabe läuft, macht jede daraus gerechnete Stelle unbrauchbar — und die
// Meldung spräche über etwas anderes.
//
// # Was das Pack dazu sagt
//
// **Kein einziges Mal.** Über **376 Stile und 8.915 Feldstile** bekommt
// `to_int` nichts, das dieses Programm annimmt und MSE ablehnt — gemessen am
// 2026-08-20, über die **ganze** Klasse und nicht nur über `"1.5"`.
//
// Die Abweichung bleibt damit stehen, gemessen folgenlos. **`Inf`
// MSE-treu abzuweisen bliebe trotzdem richtig**, sobald jemand die Stelle
// anfasst — sie ist die einzige der fünf, die aus einem Tippfehler ein
// unbrauchbares Maß macht.
//
// **Und `5 / 2` weicht nicht ab** — beide geben 2,5. Das Issue führte es bis
// zum 2026-08-20 als zweiten Fall; er ist widerlegt.
func TestToIntNimmtEineBreitereZahlensyntaxAn(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{"eine Kommazahl wird abgeschnitten", `to_string(to_int("1.5"))`, "1"},
		{"und dieselbe eine Ebene höher", `to_string("1.5" * 2)`, "3"},
		{"ein Exponent wird gelesen", `to_string(to_int("1e3"))`, "1000"},
		{"umgebende Leerzeichen stören nicht", `to_string(to_int(" 12 "))`, "12"},
		{"Gos Unterstrich wird gelesen", `to_string(to_int("1_000"))`, "1000"},
		{"und `Inf` ergibt eine Unendlichkeit", `to_string(to_int("Inf"))`, "inf"},
		{"`NaN` ebenso", `to_string(to_int("NaN"))`, "nan"},

		// **Eine hexadezimale Fließkommazahl.** Angestrichen vom Nachlauf zum
		// Zweitblick über Phase K: `"0x10"` weist `strconv.ParseFloat` ab,
		// `"0x1p4"` nicht — die Klasse ist damit breiter, als `"0x10"` allein
		// vermuten ließ.
		{"und eine hexadezimale Fließkommazahl", `to_string(to_int("0x1p4"))`, "16"},

		// **Die Nachbarn, die nicht abweichen.** Sie stehen dabei, weil die
		// Abweichung sonst größer aussähe, als sie ist.
		{"eine Teilung gibt eine Kommazahl, wie bei MSE", `to_string(5 / 2)`, "2.5"},
		{"eine aufgehende Teilung bleibt ganz", `to_string(4 / 2)`, "2"},
		{"auch die ungerade", `to_string(7 / 2)`, "3.5"},
		{"eine ganze Zahl als Text geht", `to_string(to_int("2"))`, "2"},
		{"ein Vorzeichen auch", `to_string(to_int("+12"))`, "12"},
		{"und ein negatives", `to_string(to_int("-3"))`, "-3"},
		{"ein leerer Text ist null", `to_string(to_int(""))`, "0"},
		{"und `+` verkettet, statt zu rechnen", `to_string("2" + 1)`, "21"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ergebnis := auswerten(t, fall.quelle); ergebnis != fall.erwartet {
				t.Errorf("%s ergibt %q, erwartet %q", fall.quelle, ergebnis, fall.erwartet)
			}
		})
	}
}

// **`or else` kürzt ab, und das ist gemessen folgenlos**
// (#271).
//
// # Der Unterschied
//
// MSE wertet **beide** Seiten aus, bevor `I_OR_ELSE` läuft:
//
//	case I_OR_ELSE:
//	  if (at == SCRIPT_ERROR) a = b;
//
// `src/script/context.cpp`, Zeilen 535–537, Fassung f9d9356 vom 2021-07-06.
// Dieses Programm kürzt ab, sobald die linke Seite kein Fehlerwert ist.
// Sichtbar wird das allein, wenn die rechte Seite **wirft** —
// `1 or else gibtsnicht` ist bei MSE ein Fehler und hier eine Eins.
//
// # Die Messung
//
// Über den Packlauf (376 Stile, 8.915 Feldstile) wird die rechte Seite
// **370.807-mal** übersprungen. Wertet man sie trotzdem aus, ergibt sie
// **0-mal** einen Abbruch und 905-mal einen **Fehlerwert**.
//
// **Und ein Fehlerwert ist auch bei MSE ein Wert**, kein Abbruch:
// `delay_error` (`src/script/functions/basic.cpp`, Zeile 123) legt ihn auf den
// Stapel, und `I_OR_ELSE` sieht ihn erst dort. Die 905 sind auf beiden Seiten
// harmlos; die Zahl, die zählte, ist die **Null**.
func TestOrElseKuerztAbUndDasIstFolgenlos(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{
			"die Abweichung: eine fehlerhafte rechte Seite bleibt ungesehen",
			`to_string(1 or else gibtsnicht)`, "1",
		},

		// **Die Nachbarn, die nicht abweichen** — je am Orakel geprüft.
		{"ein heiler Wert bleibt stehen", `to_string(1 or else 2)`, "1"},
		{"`nil` ist kein Fehler und bleibt stehen", `to_string(nil or else 2)`, ""},
		{"ein leerer Text ebenso", `to_string("" or else 2)`, ""},
		{"ein fehlendes Glied gibt die rechte Seite", `to_string([a: 1].b or else 2)`, "2"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ergebnis := auswerten(t, fall.quelle); ergebnis != fall.erwartet {
				t.Errorf("%s ergibt %q, erwartet %q", fall.quelle, ergebnis, fall.erwartet)
			}
		})
	}
}
