// **Operation and trust** — OPS-07
// (#1724).
export const betrieb = {
  ueberschrift: 'Funzionamento e fiducia',
  einleitung:
    'Cosa è installato, dove si trova, cosa è stato verificato e cosa no — letto dall’applicazione stessa.',
  fassungName: 'Versione',
  fassungSatz: '{fassung}, compilata da {stand}, stato del {gebaut}.',
  fassungVeraendert:
    'La compilazione conteneva modifiche di cui non era stato fatto il commit.',
  datenName: 'Dati',
  datenSatz: '{zahl} migrazioni applicate, l’ultima {letzte}.',
  datenOhne: 'Nessuna libreria aperta.',
  signaturName: 'Firma',
  signaturSatz:
    'Firmata ad hoc, **non notarizzata**. Al primo avvio macOS rifiuta l’applicazione; consentila una volta in Impostazioni di Sistema › Privacy e sicurezza › **Apri comunque**.',
  aktualisierungName: 'Aggiornamenti',
  aktualisierung: {
    aus: 'Disattivati — questa applicazione non cerca nuove versioni.',
    melden: 'Segnala una nuova versione da GitHub e non scarica nulla.',
    holen:
      'Scarica una nuova versione da GitHub e chiede prima di sostituire quella attuale.',
  },
  orteName: 'Dove si trova tutto',
  orte: {
    bestand: 'Libreria',
    pakete: 'Pacchetti di modelli',
    bilder: 'Immagini inserite',
    zwischenspeicher: 'Cache',
    protokoll: 'Registro',
    leistungsprofile: 'Profili delle prestazioni',
  },
  lizenzenName: 'Licenze',
  lizenzenSatz:
    'Il codice sorgente è sotto la GNU General Public License 2.0 o successiva. Le parti di terzi con le loro licenze sono elencate in `DRITTANBIETER.md` accanto al codice sorgente; un pacchetto di modelli porta con sé le proprie condizioni, e un modello di IA quelle del suo provider.',
  empfaengerName: 'Dove vanno i dati',
  empfaengerSatz:
    'Solo ai servizi che attivi — quali sono e cosa esce è indicato in **Aiuto › Cosa sa di te questa applicazione**. Non viene inviato alcun dato d’uso.',
  grenzenName: 'Confini di sicurezza',
  grenzenSatz:
    'Il servizio e il server degli strumenti ascoltano solo su `127.0.0.1`, se non diversamente impostato, e richiedono un token. Se assegni a uno di essi un indirizzo nella rete in **Impostazioni › Diagnostica**, funziona solo con TLS, e la pagina dice che esce da questo computer; tramite stdio il server degli strumenti parla attraverso i propri flussi. La libreria può essere cifrata, in **Impostazioni › Dati e backup**.',
  kontaktName: 'Quando qualcosa non funziona',
  kontaktSatz:
    'In **Impostazioni › Diagnostica** trovi la diagnosi, il pacchetto di diagnostica e i numeri di questo processo. Le segnalazioni di errori vanno a github.com/DavidTHeider/MagicSetEditor4Mac — nulla esce da solo.',
  pruefstandName: 'Cosa è verificato',
  pruefstandSatz:
    'Verificato significa: un’esecuzione lo ha dimostrato — una prova la cui documentazione, con il suo commit, si trova accanto al codice sorgente, oppure un test che ogni pipeline esegue; viene pubblicato solo ciò che lo ha superato. Presente significa: costruito, ma mai eseguito con l’originale o con la piattaforma.',
  spalteBereich: 'Area',
  spalteGegenstand: 'Oggetto',
  spalteStand: 'Stato',
  spalteBeleg: 'Prova',
  belegIssue: 'issue #{nummer}',
  bereiche: {
    plattform: 'Piattaforma',
    sprache: 'Lingue',
    barrierefreiheit: 'Accessibilità',
    regeln: 'Regole',
    integration: 'Servizi esterni',
  },
  staende: {
    geprueft: 'verificato',
    vorhanden: 'presente',
    beschrieben: 'descritto',
    blockiert: 'attende una persona',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS da 15.3 a 26',
    'macOS 15.2 auf Apple Silicon': 'macOS 15.2 su Apple Silicon',
    'macOS 27 auf Apple Silicon': 'macOS 27 su Apple Silicon',
    'Tastatur und Vorleser im WebKit':
      'Tastiera e lettore di schermo in WebKit',
  },
}
