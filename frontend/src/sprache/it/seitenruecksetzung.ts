// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'Ripristino',
  hinweis:
    'Sei tipi, dal più leggero al più pesante. Ciò che ciascuno tocca viene ' +
    'mostrato prima — con numeri e nomi.',

  artWaehlen: 'Che cosa vuoi ripristinare?',

  arten: {
    layout: 'Solo l’aspetto della finestra',
    profileinstellungen: 'Le impostazioni di questo profilo',
    nutzereinstellungen: 'Le mie impostazioni personali',
    profil: 'Questo profilo con tutto ciò che gli appartiene',
    nutzer: 'Me con tutti i miei profili',
    anwendung: 'L’intera applicazione',
  } as Record<string, string>,

  nachsehen: 'Mostra in anteprima cosa verrebbe interessato',
  nachsehenLaeuft: 'Controllo in corso …',

  nichtsBetroffen: 'Qui non c’è nulla che si possa ripristinare.',
  gruppe: '{was}: {anzahl}',

  gruppen: {
    einstellungen: 'Impostazioni',
    profile: 'Profili',
    sets: 'Set',
    karten: 'Carte',
    decks: 'Mazzi',
    exemplare: 'Esemplari nella collezione',
    token: 'Token di accesso',
    freigaben: 'Registri di autorizzazione',
    zugaenge: 'Accessi ai profili',
    bilder: 'Immagini',
  } as Record<string, string>,

  sicherung:
    'Qui i dati andranno persi. Fai prima un backup dell’archivio — nell’area qui sopra.',

  hindernisse: {
    offeneRueckkehr:
      'Le quantità in questo profilo sono incerte dopo un ripristino. ' +
      'Controllale e conferma il ripristino prima di eliminare.',
    fremdesZiel: 'Questo non è tuo.',
    letzterNutzer:
      'Questa è l’ultima persona in questo archivio — eliminarla non lascerebbe ' +
      'nessuno.',
    laufenderAuftrag: 'Qualcosa è ancora in corso. Aspetta che abbia finito.',
    nachweisFehlt:
      'Ripristinare l’intera applicazione richiede un’autorizzazione esplicita ' +
      'su questo computer.',
  } as Record<string, string>,

  nameTippen: 'Scrivi il nome per confermare',
  freigabeSatz:
    'Reimpostare l’intera applicazione richiede la tua autorizzazione su questo computer. Vale cinque minuti e per un solo nuovo inizio.',
  freigeben: 'Concedi l’autorizzazione',
  freigebenLaeuft: 'Autorizzazione in corso …',
  freigegeben: 'Autorizzato fino a {zeit}.',
  kennungFuerKommandozeile:
    'La stessa autorizzazione vale dalla riga di comando (mse zuruecksetzen anwendung):',

  tun: 'Ripristina',
  tunLaeuft: 'Ripristino in corso …',

  bilanz: '{anzahl} impostazioni ripristinate.',
  beiseite: 'L’archivio precedente si trova in {ort}.',
}
