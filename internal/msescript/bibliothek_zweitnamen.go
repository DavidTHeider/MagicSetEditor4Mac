package msescript

// **Zwei Namen aus dem Nachzählen von Phase 4.9** —
// #1324.
//
// Beide sind klein, und beide würde niemand aus ihrem Namen erraten. Genau
// dafür gibt es §15: **nachlesen statt erschließen.**

import "context"

func init() {
	// **`match_text` ist `match`** — und zwar herum, wie man es nicht
	// vermutet:
	//
	//	ctx.setVariable(_("match_text"), script_match_text);
	//	ctx.setVariable(_("match"),      script_match_text); // old name
	//
	// `src/script/functions/regex.cpp`, Zeilen 266–267, Fassung f9d9356 vom
	// 2021-07-06. **`match` ist der alte Name**, `match_text` der heutige;
	// die Pakete benutzen den alten.
	//
	// Dieselbe Gestalt wie `replace_text` neben `replace` in
	// `bibliothek_text.go`.
	eintragen("match_text", passt)

	// **`sanitize` tut nichts** — im Original:
	//
	//	SCRIPT_FUNCTION(sanitize) {
	//	  SCRIPT_PARAM_C(String, input);
	//	  //TODO
	//	  SCRIPT_RETURN(input);
	//	}
	//
	// `src/script/functions/export.cpp`, Zeilen 441–445, Fassung f9d9356 vom
	// 2021-07-06. Das `//TODO` steht dort.
	//
	// **Am Orakel bestätigt** (2026-09-02): `sanitize("<b>x</b>")` gibt
	// `"<b>x</b>"` zurück, unverändert.
	//
	// # Warum das trotzdem hierher gehört
	//
	// Eine Vorlage, die `sanitize(…)` ruft, scheitert ohne den Namen — und
	// zwar an einer Stelle, an der das Original nichts täte. Der Name ist
	// hier, damit sie durchläuft; was er **nicht** tut, steht hier, damit
	// niemand ihn für eine Absicherung hält.
	//
	// **Wer ihn eines Tages füllen will**, füllt damit eine Lücke des
	// Originals, nicht dieses Projekts — und das ist eine Verhaltensänderung
	// (§15, Satz 4), keine Portierung.
	eintragen("sanitize", unveraendert)
}

// unveraendert ist `sanitize(x)` — es gibt seine Eingabe zurück.
func unveraendert(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("sanitize"); err != nil {
		return Nichts, err
	}

	return Text(ruf.Stellung[0].AlsText()), nil
}
