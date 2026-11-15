// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: 'Combinación de colores',
  schriftbereich: 'La tipografía de títulos se aplica a',
  hintergrund: 'Fondo de la ventana',
  gestaltung: 'Colores personalizados',
}

export const gestaltungswerte = {
  'farbwelt.pergament': 'Pergamino — cálido, con fibra',
  'farbwelt.klar': 'Claro — neutro, sin fibra',
  'farbwelt.schiefer': 'Pizarra — gris azulado frío',
  'farbwelt.moos': 'Musgo — verde apagado',
  'schriftbereich.ueberschriften': 'Solo los títulos',
  'schriftbereich.ganz': 'Toda la interfaz',
  'hintergrund.aus': 'Desactivado',
  'hintergrund.fuellen': 'Rellenar',
  'hintergrund.einpassen': 'Ajustar',
  'hintergrund.kacheln': 'Mosaico',
  'schriftart.buch': 'Libro — New York',
  'schriftart.rund': 'Redondeada — SF Pro Rounded',
  'schriftart.serifenlos': 'Sin serifa — la tipografía del sistema',
}

export const gestaltungshinweise = {
  farbwelt:
    'Los colores de la interfaz, en claro y en oscuro, todos recalculados. Con contraste alto se aplica la tinta, elijas lo que elijas aquí.',
  schriftbereich:
    'Toda la interfaz solo admite una tipografía de texto; una tipografía decorativa se queda en los títulos. Las cartas conservan las tipografías de su plantilla.',
  hintergrund:
    'Coloca tu propia imagen bajo el fondo, con un velo que mantiene el texto legible. La imagen la eliges más abajo; **Desactivado** la quita al instante.',
  gestaltung:
    'Colores personalizados para las diez propiedades de la interfaz, una por línea. Solo colores en hexadecimal: sin direcciones ni tipografías. Lo que no se pudiera leer no se aplica.',
}

export const gestaltungsbeiwerk = {
  fensterbild: 'Tu fondo de ventana',
  vorschau: 'Vista previa del fondo de la ventana',
  keinBild:
    'Aún no hay ninguna imagen propia. Elige una: se mostrará aquí y, en cuanto el fondo de la ventana no esté desactivado, detrás de la interfaz.',
  aus: 'La imagen está lista, pero el fondo de la ventana está en **Desactivado**.',
  waehlen: 'Elegir imagen …',
  entfernen: 'Quitar imagen',
  merkmale: 'Colores personalizados',
  merkmaleHilfe:
    'Una entrada por línea, por ejemplo «--mse-marke: #4e2f16;». Se admiten {merkmale}.',
  uebernehmen: 'Aplicar',
  zeilenfehler:
    'Línea {zeile}: se esperaba una entrada como «--mse-grund: #1e1b16;», es decir, una de las propiedades admitidas y un color en hexadecimal.',
  nichtLesbar: 'No se aplica: no sería legible:',
  paar: '{vorn} sobre {hinten}: {verhaeltnis} : 1, se requiere {schwelle} : 1',
  paarnamen: {
    'status:warnung': 'Advertencias',
    'status:erfolg': 'Mensajes de éxito',
    'fensterbild:dunkel': '--mse-grund sobre un fondo de ventana oscuro',
    'fensterbild:hell': '--mse-grund sobre un fondo de ventana claro',
  } as Record<string, string>,
  angewandt: 'Aplicado.',
  ruhtBeiKontrast:
    'Guardado: los colores personalizados no se aplican con contraste alto; volverán en cuanto lo desactives.',
  schmuckNurOben:
    'La tipografía de títulos elegida es decorativa: se queda en los títulos.',
  zuruecksetzenGescheitert: 'No se pudo restablecer la apariencia:',
  zurVorgabe: 'Restablecer tus colores a los predeterminados',
}
