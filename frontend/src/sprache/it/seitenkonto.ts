// **The account kinds on the user page** — piece 6 of phase 4.12 (#1351).
export const konto = {
  eigenes: 'Un account personale',
  eigenesErklaerung:
    'Un indirizzo o un numero di telefono è qui un identificativo: dice chi accede. Se la casella di posta o il numero esistano davvero, il programma non lo verifica — non invia nulla e non ha bisogno di alcun servizio esterno per farlo.',

  kennung: 'Indirizzo o numero di telefono',
  kennungHinweis: 'Ad esempio nome@example.org oppure +4915112345678',

  kennwort: 'Password',
  kennwortHinweis: 'Almeno dodici caratteri',
  kennwortAnmelden: 'Password',

  anzeigename: 'Nome visualizzato (facoltativo)',
  anzeigenameHinweis: 'Se lasciato vuoto, viene ricavato dall’identificativo',

  eintragen: 'Crea account',
  eingetragen: 'Creato: {name}',

  anmelden: 'Accedi',
  angemeldet: 'Accesso eseguito come {name}',

  ohneAngaben: 'Senza dati',
  ohneAngabenErklaerung:
    'Un account anonimo riceve come nome un identificativo casuale, nessuna password e nessun indirizzo. E chi non ne vuole affatto non ne ha bisogno: l’utente del primo avvio c’è, e l’applicazione funziona.',
  anonym: 'Crea account anonimo',
  anonymAngelegt: 'Creato: {name}',

  kennwortNieZurueck:
    'La password viene salvata in forma protetta (hash) e non viene mai più mostrata. Chi la dimentica ne imposta una nuova invece di leggere la vecchia.',
} as const

// **The second factor** — piece 7 of phase 4.12 (#1351).
export const faktor = {
  titel: 'Secondo fattore',
  erklaerung:
    'Un secondo fattore viene richiesto all’accesso e al cambio di profilo. Un codice di un’app di autenticazione funziona ovunque; Touch ID funziona solo su questo Mac.',

  totp: 'Codice da un’app',
  biometrie: 'Biometria',

  einrichten: 'Configura',
  entfernen: 'Rimuovi',

  totpEinrichten: 'Configura un codice da un’app',
  totpEntfernen: 'Rimuovi il codice da un’app',
  biometrieEinrichten: 'Configura Touch ID o Face ID',
  biometrieEntfernen: 'Rimuovi Touch ID o Face ID',

  totpEntfernenFrage: 'Rimuovere il codice da un’app?',
  biometrieEntfernenFrage: 'Rimuovere Touch ID o Face ID?',
  entfernenErklaerung:
    'L’accesso e il cambio di profilo non lo richiedono più. Chi lo rivuole lo configura di nuovo — un’app con un nuovo segreto.',
  entfernenJa: 'Rimuovi',

  einmal: 'Viene mostrato una volta sola',
  einmalErklaerung:
    'Inserisci questo segreto nella tua app di autenticazione e conserva i codici di recupero. Dopo non si potranno più recuperare — chi li perde configura tutto da capo.',
  codesUeberschrift: '{zahl} codici di recupero, ciascuno valido una volta:',
  bestaetigen:
    'Inserisci un codice dalla tua app. Solo allora il secondo fattore diventa valido.',
} as const
