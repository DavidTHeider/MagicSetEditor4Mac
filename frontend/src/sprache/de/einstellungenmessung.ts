// **Die Einstellungen der Messung** — OPS-05
// (#1722).
// Eine eigene Datei aus demselben Grund wie die des Protokolls; sie wird in
// `seiteneinstellungen.ts` und `einstellungshinweise.ts` eingesetzt.
export const messnamen = {
  'messung.spur': 'Ablaufspur',
  'messung.abtastung': 'Anteil der aufgenommenen Rufe in Prozent',
  'messung.spurgrenze': 'Höchstens so viele Einträge',
  'messung.aufbewahrung': 'Einträge bleiben so viele Minuten',
  'messung.profile': 'Leistungsprofile, die liegen bleiben',
  'zwischenspeicher.grenze': 'Grenze je Zwischenspeicher auf der Platte in MB',
} as const

export const messwerte = {
  'messung.spur.aus': 'Aus',
  'messung.spur.an': 'An',
} as const

export const messhinweise = {
  'messung.spur':
    'Eine Zeile je Ruf an Dienst und Werkzeugserver — mit Vorgangskennung, Weg, Dauer und Ausgang, ohne Inhalte. Sie liegt im Arbeitsspeicher des Prozesses und gilt ab dem nächsten Start.',
  'messung.abtastung':
    'Bei 10 nimmt die Spur jeden zehnten Ruf auf — genau gezählt, nicht gewürfelt. Leer heißt 100.',
  'messung.spurgrenze':
    'Ist die Spur voll, geht der älteste Eintrag. 10 bis 10000, leer heißt 1000.',
  'messung.aufbewahrung':
    'Ältere Einträge fallen heraus. 1 bis 1440 Minuten, leer heißt 60.',
  'messung.profile':
    'Die neuesten bleiben neben dem Bestand liegen, ältere gehen beim nächsten Profil. 1 bis 20, leer heißt 3.',
  'zwischenspeicher.grenze':
    'Für Antworten und Bilder von Scryfall, je für sich. Beim Start gehen die ältesten, bis sie passt. 1 bis 8192, leer heißt 512.',
} as const
