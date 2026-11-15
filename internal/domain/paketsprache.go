package domain

import "slices"

// **Das Fenster und MSE nennen dieselbe Sprache verschieden** — LOC-01
// (#1687).
//
// Das Fenster führt Sprachen nach BCP 47 (`ja`, `zh`); MSE nennt sie nach
// seinen Sprachpaketen (`jp.mse-locale`, `ch-s.mse-locale`), und unter genau
// diesem Namen stehen die Einträge eines `localized caption:`-Blocks:
//
//	String const& LocalizedString::get() const {
//	  if (translations.empty()) return default_;
//	  return get(the_locale->name());
//	}
//
// `src/data/locale.cpp`, Zeilen 138–141, Fassung f9d9356 vom 2021-07-06.
// Welche Namen es gibt, steht im Inventar `docs/SPRACHINVENTAR.md` — gezählt
// an den dreizehn `.mse-locale` des Packs, nicht erraten.
//
// **Chinesisch heißt hier vereinfacht** (`ch-s`): das Fenster führt eine
// chinesische Schrift, und die Entscheidung steht im Inventar.

// fenstersprachen sind die Sprachen, in denen das Fenster spricht — Deutsch
// als Ausgangssprache vorn, die übrigen nach ihrem Kürzel.
var fenstersprachen = []string{"de", "en", "es", "fr", "it", "ja", "ru", "zh"}

// Fenstersprachen gibt sie als Kopie: ein Aufrufer, der die Liste änderte,
// änderte die Regel.
func Fenstersprachen() []string {
	return slices.Clone(fenstersprachen)
}

// paketsprachen bildet die Fenstersprachen ab, deren MSE-Name ein anderer ist.
var paketsprachen = map[string]string{
	"ja": "jp",
	"zh": "ch-s",
}

// Paketsprache gibt den Namen, unter dem MSE eine Fenstersprache führt.
// Wo beide gleich heißen, kommt das Kürzel unverändert zurück.
func Paketsprache(sprache string) string {
	if name, da := paketsprachen[sprache]; da {
		return name
	}

	return sprache
}
