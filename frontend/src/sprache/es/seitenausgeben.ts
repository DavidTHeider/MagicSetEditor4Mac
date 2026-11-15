// The texts of the **export assistant** — where to, and what each
// destination needs.
//
// Split out of `seiten.ts` when the assistant brought the file close to the
// 400-line limit (§2, UXR-09, #1917).

export const ausgeben = {
  titel: 'Exportar',
  erklaerung:
    'Cinco formas de sacar un set de esta aplicación, y para qué sirve cada una.',
  wohin: '¿A dónde?',
  keinZiel: 'Aún no has elegido un destino',
  nichtGemerkt:
    'El destino se aplica, pero no se ha recordado: la próxima vez no estará preseleccionado:',
  keinZielErklaerung:
    'Elige en «¿A dónde?» en qué forma debe salir el set: aquí aparecerá lo que hay que ajustar.',
  ohneKarten:
    'Las imágenes y la hoja de impresión no tendrían nada que mostrar. Las cartas se crean en «Cartas»: nuevas, pegadas o importadas.',
  ziele: {
    setdatei: {
      name: 'Archivo de set',
      kurz: 'El set en sí, para volver a abrirlo — un archivo `.mse-set` que también lee Magic Set Editor 2.',
      lang: 'El atajo es ⌘S. **No incluye** mazos, ejemplares, precios, etiquetas ni el origen de cada valor — se quedan en la biblioteca, porque MSE2 descarta al guardar lo que no conoce.',
    },
    bilder: {
      kurz: 'Un archivo de imagen por carta para usar en otro sitio: un foro, una hoja de cálculo, un programa de juego.',
    },
    druckbogen: {
      kurz: 'Las cartas una junto a otra en papel, con marcas de corte.',
    },
    vorlage: {
      kurz: 'Texto en lugar de imagen: una lista, una tabla, un archivo para otro programa.',
      lang: 'Qué plantillas existen depende del paquete instalado.',
    },
    mtgjson: {
      kurz: 'El formato de cartas de mtgjson.com, para herramientas que lo leen.',
      lang: 'No sirve para seguir trabajando en esta aplicación: lo que MTG JSON no puede recoger se indica después de exportar.',
    },
  },
  vorlagenEingrenzen: 'Filtrar plantillas',
  weitereOptionen: 'Más opciones ({zahl})',
  // **„Cómo quedará"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: 'Cómo quedará',
    rechnet: 'Calculando…',
    hinweis: 'No se escribe nada hasta que pulses el botón de abajo.',
    bogen: {
      jeSeite: '{zahl} por página · {seiten}',
      masse: '{karten} · papel {breite} × {hoehe} mm · {dpi} dpi',
      letzte: 'En la última página: {karten}.',
    },
    vorlage: {
      datei: 'Archivo: {name} ({groesse})',
      karten: 'Evaluado sobre {karten}.',
      beigaben: 'Archivos junto a él: {zahl} (carpeta «{ordner}», {groesse})',
      mindestens: 'al menos {groesse}',
      befunde: 'Observado al evaluar:',
    },
  },
}
