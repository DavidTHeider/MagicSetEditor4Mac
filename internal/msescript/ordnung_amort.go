package msescript

// `in_place(…)` — die Ordnung, die an Ort und Stelle sortiert.
//
// **Eigene Datei** (§2): `bibliothek_ordnung.go` beantwortet, in welcher
// Reihenfolge Zeichen stehen; hier steht, wie eine Ordnung sie **liegen** lässt
// und nur untereinander tauscht. Der Anlass war die 400-Zeilen-Grenze, der
// Schnitt folgt der Frage
// (#415).

// ordneAnOrt sortiert die genannten Zeichen untereinander und lässt sie liegen.
//
//	/// Sort things in place, keep the rest of the input
//	void in_place_sort(const String& spec, String& input, String& ret) {
//	  String result;
//	  spec_sort(spec, input, result);
//	  size_t pos_r = 0;
//	  FOR_EACH_CONST(c, input) {
//	    if (c == REMOVED) {
//	      if (pos_r < result.size()) ret += result.GetChar(pos_r++);
//	    } else {
//	      ret += c;
//	    }
//	  }
//	  input.clear(); // we ate all the input
//	}
//
// `src/util/spec_sort.cpp`, Zeilen 244–261, Fassung f9d9356 vom 2021-07-06.
//
// **Der Rest bleibt**, anders als beim gewöhnlichen `sort_text` — MSE schreibt
// es in die Kopfzeile der Funktion, und die Schleife oben gibt jedes Zeichen
// der Eingabe aus. Deshalb steht hier kein [behalteGenannte].
func (o ordnung) ordneAnOrt(eingabe string) string {
	zeichen := []rune(eingabe)

	// Die Stellen der genannten Zeichen, in Eingabereihenfolge.
	var stellen []int

	for i, einzelnes := range zeichen {
		if _, bekannt := o.rang[einzelnes]; bekannt {
			stellen = append(stellen, i)
		}
	}

	if len(stellen) == 0 {
		return eingabe
	}

	// **Über die gewöhnliche Ordnung sortiert**, damit `cycle(…)` und die
	// Klammergruppen hier genauso gelten wie sonst.
	gesammelt := make([]rune, 0, len(stellen))
	for _, wo := range stellen {
		gesammelt = append(gesammelt, zeichen[wo])
	}

	// **Genauso viele wie vorher.** [behalteGenannte] wirft nur weg, was die
	// Ordnung nicht kennt — und `gesammelt` enthält nur Genanntes. MSE führt an
	// dieser Stelle einen Wächter (`if (pos_r < result.size())`), weil sein
	// `spec_sort` auch Genanntes verlieren kann; hier wäre er ein Zweig, den
	// kein Aufruf betritt (§3).
	geordnet := []rune(o.wende(string(gesammelt)))

	for i, wo := range stellen {
		zeichen[wo] = geordnet[i]
	}

	return string(zeichen)
}
