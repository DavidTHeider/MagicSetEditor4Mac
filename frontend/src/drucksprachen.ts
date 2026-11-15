// **Die Sprachen eines Drucks, wie Scryfall sie führt** — UX-06.AK-01
// (#1678).
//
// **Nachgelesen am 2026-09-24** in Scryfalls Dokumentation, Abschnitt
// „Languages" (`scryfall.com/docs/api/languages`): zwanzig Codes, in der
// Reihenfolge dort. `domain.Exemplar.Sprache` ist „der Sprachcode des
// Drucks, wie Scryfall ihn führt" — deshalb diese Liste und keine nach ISO.
//
// **Vorschläge und keine Grenze** (AK-01): der Bestand nimmt jeden Code an,
// und eine Sprache, die Scryfall morgen aufnimmt, soll sich heute schon
// erfassen lassen. Zwei davon zeigen, warum die Liste nicht aus `Intl` kommt:
// `ph` ist Phyrexianisch und `dw` Zwergisch — keine Plattform kennt sie.
export const drucksprachen = [
  'en',
  'es',
  'fr',
  'de',
  'it',
  'pt',
  'ja',
  'ko',
  'ru',
  'zhs',
  'zht',
  'he',
  'la',
  'grc',
  'ar',
  'sa',
  'ph',
  'qya',
  'dw',
  'tlh',
] as const

// Drucksprache ist einer dieser Codes.
export type Drucksprache = (typeof drucksprachen)[number]
