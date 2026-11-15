// **Die Eigennamen der fremden Dienste** — Phase 4.8, seit A8 auch der
// erste KI-Anbieter.
//
// **Sie werden nicht übersetzt** ([0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md),
// Entscheidung 5): „Scryfall" und „Scrydex" sind Eigennamen (§9).
//
// **Sie stehen hier und nicht im Sprachkatalog**, und das ist der Grund: ein
// Eigenname in einem Katalog lädt dazu ein, ihn zu übersetzen. Hier stehen
// sie einmal — für die Schalter **und** für die Wahl je Frage.
export const dienstnamen: Record<string, string> = {
  scryfall: 'Scryfall',
  mtgapi: 'magicthegathering.io',
  mtgjson: 'MTGJSON',
  scrydex: 'Scrydex',
  cardtrader: 'CardTrader',

  // **Anthropic ist kein Kartendienst** (A8, #1447) — es steht hier, weil
  // sein Schalter derselbe ist: ein Eigenname, der nicht übersetzt wird.
  anthropic: 'Anthropic',
  openai: 'OpenAI',
  xai: 'xAI',
}

// schalternamen sind dieselben Namen, unter ihrem Einstellungsschlüssel.
export const schalternamen: Record<string, string> = Object.fromEntries(
  Object.entries(dienstnamen).map(([kennung, name]) => [
    `dienst.${kennung}`,
    name,
  ]),
)
