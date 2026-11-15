// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: 'Tema colori',
  schriftbereich: 'Il carattere di visualizzazione vale per',
  hintergrund: 'Sfondo della finestra',
  gestaltung: 'Colori personalizzati',
}

export const gestaltungswerte = {
  'farbwelt.pergament': 'Pergamena — calda, con grana',
  'farbwelt.klar': 'Limpido — neutro, senza grana',
  'farbwelt.schiefer': 'Ardesia — grigio-azzurro freddo',
  'farbwelt.moos': 'Muschio — verde attenuato',
  'schriftbereich.ueberschriften': 'Solo i titoli',
  'schriftbereich.ganz': 'L’intera interfaccia',
  'hintergrund.aus': 'Disattivato',
  'hintergrund.fuellen': 'Riempi',
  'hintergrund.einpassen': 'Adatta',
  'hintergrund.kacheln': 'Affianca',
  'schriftart.buch': 'Libro — New York',
  'schriftart.rund': 'Arrotondato — SF Pro Rounded',
  'schriftart.serifenlos': 'Senza grazie — il carattere di sistema',
}

export const gestaltungshinweise = {
  farbwelt:
    'I colori dell’interfaccia, chiari e scuri, ciascuno ricalcolato. Con il contrasto elevato vale l’inchiostro, qualunque cosa sia scelta qui.',
  schriftbereich:
    'L’intera interfaccia accetta solo un carattere da testo; un carattere decorativo resta sui titoli. Le carte mantengono i caratteri del loro modello.',
  hintergrund:
    'Mette una tua immagine sotto lo sfondo, con un velo che mantiene leggibile il testo. L’immagine la scegli qui sotto; **Disattivato** la toglie subito.',
  gestaltung:
    'Colori personalizzati per le dieci proprietà dell’interfaccia, uno per riga. Solo colori esadecimali — niente indirizzi, niente caratteri. Ciò che risulterebbe illeggibile non viene applicato.',
}

export const gestaltungsbeiwerk = {
  fensterbild: 'Il tuo sfondo della finestra',
  vorschau: 'Anteprima dello sfondo della finestra',
  keinBild:
    'Ancora nessuna immagine personale. Scegline una: comparirà qui e, non appena lo sfondo della finestra non è disattivato, dietro l’interfaccia.',
  aus: 'L’immagine è pronta, ma lo sfondo della finestra è impostato su **Disattivato**.',
  waehlen: 'Scegli immagine …',
  entfernen: 'Rimuovi immagine',
  merkmale: 'Colori personalizzati',
  merkmaleHilfe:
    'Una voce per riga, ad esempio “--mse-marke: #4e2f16;”. Sono consentiti {merkmale}.',
  uebernehmen: 'Applica',
  zeilenfehler:
    'Riga {zeile}: è attesa una voce come “--mse-grund: #1e1b16;” — una delle proprietà consentite e un colore esadecimale.',
  nichtLesbar: 'Non applicato — sarebbe illeggibile:',
  paar: '{vorn} su {hinten}: {verhaeltnis} : 1, richiesto {schwelle} : 1',
  paarnamen: {
    'status:warnung': 'Avvisi',
    'status:erfolg': 'Messaggi di successo',
    'fensterbild:dunkel': '--mse-grund su uno sfondo della finestra scuro',
    'fensterbild:hell': '--mse-grund su uno sfondo della finestra chiaro',
  } as Record<string, string>,
  angewandt: 'Applicato.',
  ruhtBeiKontrast:
    'Salvato — con il contrasto elevato i colori personalizzati non valgono; tornano non appena viene disattivato.',
  schmuckNurOben:
    'Il carattere di visualizzazione scelto è decorativo — resta sui titoli.',
  zuruecksetzenGescheitert: 'Non è stato possibile ripristinare l’aspetto:',
  zurVorgabe: 'Ripristina i colori personalizzati predefiniti',
}
