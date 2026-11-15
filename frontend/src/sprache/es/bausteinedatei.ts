export const dateiansicht = {
  ueberschrift: 'El archivo',
  laedt: 'Examinando el archivo …',
  arten: {
    paket: 'Paquete MSE',
    bild: 'Imagen',
    vektor: 'Imagen vectorial',
    pdf: 'PDF',
    text: 'Texto',
    archiv: 'Archivo comprimido',
    unbekannt: 'Tipo desconocido',
  },
  name: 'Nombre',
  art: 'Tipo',
  groesse: 'Tamaño',
  geaendert: 'Modificado',
  ort: 'Ubicación',
  masse: '{breite} × {hoehe} puntos',
  seiten: '{anzahl} páginas',
  eintraege: '{anzahl} entradas',
  gekuerzt: 'Se muestra el principio: el archivo continúa.',
  unsicher:
    '{anzahl} entradas salen del archivo o son enlaces. Se cuentan, pero nunca se accede a ellas.',
  unsicherEintrag: 'no se accede',
  verschluesselt:
    'Una copia de seguridad con contraseña. Lo que contiene solo se ve al restaurarla.',
  zuGross: 'Demasiado grande para una vista previa. Ábrelo con su aplicación.',
  unlesbar:
    'No se puede leer el contenido. El nombre, el tamaño y la hora son correctos.',
  keineVorschau:
    'Este tipo no tiene vista previa: podría ser un programa. Muéstralo en el Finder si lo necesitas.',
  imFinder: 'Mostrar en el Finder',
  oeffnen: 'Abrir',
  imFinderFuer: 'Mostrar {name} en el Finder',
  oeffnenFuer: 'Abrir {name} con su aplicación',
  ansehen: 'Ver',
  ansehenFuer: 'Ver {name}',
  zuklappen: 'Cerrar la vista',
} as const
