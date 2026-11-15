// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'Al desactivarla, un área desaparece de la barra lateral y de la paleta, y sus rutas se rechazan en todas las superficies. Sus datos se conservan: al reactivarla, todo vuelve.',
    hinweis:
      'Un interruptor oculta, no protege: para proteger algo, usa roles y permisos.',
    reife: {
      stabil: 'estable',
      vorschau: 'Vista previa',
      auslaufend: 'en retirada',
    },
    saetze: {
      sammeln: 'Colección, posesiones y listas de deseos',
      spielen: 'Mazos, torneos, partidas y la mesa',
      handeln: 'Intercambios, permisos y sobres',
      ressourcen: 'Paquetes de plantillas y plantillas',
      ki: 'Sugerencias, imágenes, índice de búsqueda, cuentas y costes',
      netz: 'Identidad, contactos, invitaciones y accesibilidad',
      markt: 'Ofertas, precios y cuentas de los mercados',
      turniere: 'Torneos con rondas, emparejamientos y resultados',
      dienst:
        'REST para scripts: un servicio en marcha se detiene al desactivarlo',
      werkzeugserver:
        'MCP por HTTP: un servidor en marcha se detiene al desactivarlo',
    },
    abgeschaltet: {
      titel: '{seite} está desactivada',
      erklaerung:
        'Vuelve a activar el área en Ajustes › Áreas y funciones. Sus datos siguen ahí, sin cambios.',
      oeffnen: 'Abrir áreas y funciones',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'Coleccionar',
  'schalter.spielen': 'Jugar',
  'schalter.handeln': 'Intercambiar',
  'schalter.ressourcen': 'Recursos',
  'schalter.ki': 'IA',
  'schalter.netz': 'Red y contrapartes',
  'schalter.markt': 'Mercados y precios',
  'schalter.turniere': 'Torneos',
  'schalter.dienst': 'Servicio para scripts',
  'schalter.werkzeugserver': 'Servidor de herramientas por HTTP',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'activada',
  'schalter.sammeln.aus': 'desactivada',
  'schalter.spielen.an': 'activada',
  'schalter.spielen.aus': 'desactivada',
  'schalter.handeln.an': 'activada',
  'schalter.handeln.aus': 'desactivada',
  'schalter.ressourcen.an': 'activada',
  'schalter.ressourcen.aus': 'desactivada',
  'schalter.ki.an': 'activada',
  'schalter.ki.aus': 'desactivada',
  'schalter.netz.an': 'activada',
  'schalter.netz.aus': 'desactivada',
  'schalter.markt.an': 'activada',
  'schalter.markt.aus': 'desactivada',
  'schalter.turniere.an': 'activada',
  'schalter.turniere.aus': 'desactivada',
  'schalter.dienst.an': 'activada',
  'schalter.dienst.aus': 'desactivada',
  'schalter.werkzeugserver.an': 'activada',
  'schalter.werkzeugserver.aus': 'desactivada',
}
