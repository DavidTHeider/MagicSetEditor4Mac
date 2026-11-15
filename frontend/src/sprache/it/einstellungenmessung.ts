// **The measurement settings** — OPS-05
// (#1722).
export const messnamen = {
  'messung.spur': 'Traccia',
  'messung.abtastung': 'Quota di chiamate registrate, in percentuale',
  'messung.spurgrenze': 'Al massimo questo numero di voci',
  'messung.aufbewahrung': 'Le voci restano per questo numero di minuti',
  'messung.profile': 'Profili delle prestazioni conservati',
  'zwischenspeicher.grenze': 'Limite per cache su disco in MB',
} as const

export const messwerte = {
  'messung.spur.aus': 'Disattivata',
  'messung.spur.an': 'Attivata',
} as const

export const messhinweise = {
  'messung.spur':
    'Una riga per ogni chiamata al servizio e al server degli strumenti — con identificativo dell’operazione, percorso, durata ed esito, senza contenuti. Vive nella memoria del processo e vale dal prossimo avvio.',
  'messung.abtastung':
    'Con 10 la traccia registra una chiamata su dieci — contate con precisione, non a caso. Vuoto significa 100.',
  'messung.spurgrenze':
    'Quando la traccia è piena, la voce più vecchia viene eliminata. Da 10 a 10000, vuoto significa 1000.',
  'messung.aufbewahrung':
    'Le voci più vecchie vengono scartate. Da 1 a 1440 minuti, vuoto significa 60.',
  'messung.profile':
    'I più recenti restano accanto al database, quelli più vecchi se ne vanno con il profilo successivo. Da 1 a 20, vuoto significa 3.',
  'zwischenspeicher.grenze':
    'Per le risposte e le immagini di Scryfall, ciascuna a sé. All’avvio vengono eliminati i più vecchi finché non rientra nel limite. Da 1 a 8192, vuoto significa 512.',
} as const
