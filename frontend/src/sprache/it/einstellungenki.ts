// Die Einstellungen der KI — die lokale Laufzeit, Apples Gerätemodell und
// die Adresse (A7, #1446; AI-05, #2042). **Aus `seiteneinstellungen.ts`
// herausgelöst**, als der Katalog über die Zeilengrenze aus §2 ging.

export const kinamen = {
  'ki.laufzeit': 'Runtime IA locale',
  'ki.geraetemodell': 'Modello sul dispositivo di Apple',
  'ki.vorrangmodell': 'Modello Ollama per la priorità',
  'ki.adresse': 'Indirizzo del runtime IA',
} as const

export const kihinweise = {
  'ki.laufzeit':
    'Se viene interpellato un runtime IA. **Disattivato** significa che non parte alcuna chiamata — nemmeno verso questo computer. Da qui non viene installato né avviato nulla.',
  'ki.geraetemodell':
    'Se il modello sul dispositivo di macOS viene interpellato per piccoli compiti di testo — nome della carta, testo d’atmosfera. **Automatico** significa: se è attivo su questo Mac (Apple Intelligence attiva, licenza di `fm` accettata), è incluso e interpellato per primo; i dati non lasciano il computer. **Disattivato** vale finché non lo cambi; disattivarlo per una sessione si fa in **IA**.',
  'ki.vorrangmodell':
    'Quale modello Ollama interroga un suggerimento con **priorità**, dopo il modello sul dispositivo. **Vuoto** significa il primo in grado di rispondere in forma fissa, e Ollama elenca per primo l’ultimo scaricato. Un modello scelto che Ollama non elenca più viene segnalato e non sostituito in silenzio da un altro.',
  'ki.adresse':
    'Dove il runtime resta in ascolto. Vuoto significa l’indirizzo predefinito su questo computer. Un indirizzo che **non** è su questo computer classifica ogni modello dietro di esso come non locale — con solo locale non viene interpellato affatto.',
} as const

export const kiwerte = {
  'ki.laufzeit.aus': 'Disattivato',
  'ki.laufzeit.an': 'Attivato',
  'ki.geraetemodell.automatisch': 'Automatico',
  'ki.geraetemodell.aus': 'Disattivato',
} as const
