// Das Einhängen der Oberfläche — ohne React-Aufruf, damit es prüfbar bleibt.
//
// Bis zum Zweitblick am Ende von Phase 1 stand die Suche nach `#wurzel` samt
// ihrer Fehlerbehandlung in `main.tsx`, und `main.tsx` war dateiweise von der
// Abdeckung ausgenommen. Der Grund dafür — „lässt sich ohne Fenster nicht
// ausführen" — trägt für `createRoot`, nicht für ein `getElementById`: unter
// jsdom ist genau diese Verzweigung prüfbar. Eine dateiweite Ausnahme versteckte
// damit prüfbaren Produktivcode, und eine unbequeme Verzweigung ist kein
// Ausnahmegrund (AGENTS.md §3).
//
// keinfehlertext: dieser Satz erreicht keine Oberfläche — er bricht das
// Einhängen ab, und ein Fenster, in dem er erschiene, gibt es dann nicht. Er
// steht für die Person, die `index.html` verändert hat, und die liest ihn in
// der Konsole. Ein Katalogschlüssel dafür wäre ein Eintrag, den keine Sitzung
// je zeigt.
export function wurzelSuchen(dokument: Document): HTMLElement {
  const wurzel = dokument.getElementById('wurzel')

  if (wurzel === null) {
    throw new Error('Das Element #wurzel fehlt in index.html')
  }

  return wurzel
}
