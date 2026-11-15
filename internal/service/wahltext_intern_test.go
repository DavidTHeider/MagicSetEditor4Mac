package service

import "testing"

// **`capitalize_sentence`, und nur das erste Zeichen** —
// #561.
//
//	void capitalize_sentence_in_place(String& s) {
//	  if (!s.empty()) s[0] = toUpper(s[0]);
//	}
//
// `src/util/string.cpp`, Zeilen 115–119, Fassung f9d9356 vom 2021-07-06.
//
// **Ein interner Test**, weil der leere Wert über [feldebene] nicht ankommt:
// ein Feld ohne Wert zeichnet gar keine Textebene. Die Wache steht trotzdem —
// ohne sie stürzte der erste Aufruf mit einer leeren Zeichenkette ab.
func TestSatzGrossFolgtMSEsRegel(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"":               "",
		"common":         "Common",
		"Common":         "Common",
		"mythic rare":    "Mythic rare",
		"überaus selten": "Überaus selten",
		"1 of 3":         "1 of 3",
	}

	for wert, erwartet := range faelle {
		t.Run(wert, func(t *testing.T) {
			t.Parallel()

			if gab := satzGross(wert); gab != erwartet {
				t.Errorf("gab %q, erwartet %q", gab, erwartet)
			}
		})
	}
}
