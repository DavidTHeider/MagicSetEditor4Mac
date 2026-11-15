// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: 'Plantillas',
  erklaerung:
    'Lo que puede hacer una carta y qué aspecto tiene está en archivos de texto dentro del paquete de plantillas: campos, scripts, medidas. Aquí se pueden modificar. El estado anterior se conserva como copia de seguridad junto al paquete.',
  paket: 'Paquete',
  keinPaket: 'Todavía no se ha elegido ningún paquete.',
  waehleEines:
    'Elige arriba un paquete para ver y modificar sus archivos de texto: campos, scripts, medidas.',
  keineDateien: 'Este paquete no tiene ningún archivo que editar.',
  dateien: 'Archivos',
  datei: 'Archivo',
  inhalt: 'Contenido',
  speichern: 'Guardar',
  gespeichert: 'Guardado.',
  ungespeichert: 'Sin guardar.',
  warnung:
    'Este archivo vale para todos los sets que usan este paquete, no solo para el abierto.',
  konflikt:
    'El archivo ha cambiado desde que se abrió. Carga su versión (tu texto se pierde) o conserva la tuya: entonces los cambios mostrarán qué sobrescribes de la suya.',
  neuLaden: 'Cargar su versión',
  behalten: 'Conservar la mía',
  vorschau: 'Vista previa',
  vorschauErklaerung:
    'La primera carta del primer set que usa esta plantilla, dibujada con el texto que hay aquí. No se guarda nada en el proceso.',
  vorschauLaeuft: 'Dibujando …',
  vorschauKeineKarte:
    'Ningún set abierto usa esta plantilla. Abre en «Sets» uno que la mencione y aquí podrás ver qué efecto tiene un cambio.',

  arbeitsplatz: 'Taller de plantillas',
  baumBezeichnung: 'Paquetes y archivos',
  inspectorBezeichnung: 'Vista previa y diagnóstico',
  baumSuche: 'Buscar paquetes y archivos',
  baumSuchePlatzhalter: 'Paquete o archivo …',
  baumLeer: 'Ningún paquete ni archivo coincide con la búsqueda.',
  baumOhnePakete:
    'Todavía no hay ningún paquete que editar. Crea uno en «Paquetes» o descarga uno y aparecerá aquí.',
  nurLesbar: 'Incluido: solo lectura. Se puede editar una copia propia.',
  baumNurMitgeliefert:
    'Todos los paquetes de aquí vienen incluidos. Crea una copia propia de uno o un paquete nuevo en «Paquetes»; ambos se pueden editar y aparecen arriba.',
  dateienLaden: 'Leyendo archivos …',
  inspectorLeer:
    'Aquí aparecen la vista previa, el diagnóstico y el origen del archivo abierto.',
  waehleImBaum:
    'Elige un archivo en el árbol de la izquierda: campos, scripts y medidas están allí como texto y se pueden modificar aquí.',
  verwerfenFrage:
    'El archivo abierto tiene cambios sin guardar. ¿Descartarlos y abrir {datei}?',
  verwerfen: 'Descartar y abrir',
  hierBleiben: 'Quedarme aquí',
  testkarte: 'Carta de prueba',
  testkarteVorgabe: 'Primera carta del primer set',
  testkarteOhneNamen: 'Carta {nummer}',
  testkarteErklaerung:
    'Una plantilla rara vez falla con la primera carta. El set de ejemplo «Werkstatt-Testkarten» trae texto largo, un diseño especial, una segunda cara, una imagen que falta y caracteres especiales: ábrelo en «Sets» y cambia su estilo a este paquete.',
  veraltet:
    'Desactualizado: el archivo tiene un hallazgo en la línea {zeile}; se muestra el último estado válido.',
  wirdNeuGezeichnet: 'Volviendo a dibujar …',
  diagnose: 'Diagnóstico',
  diagnoseFrei: 'Ningún hallazgo en este archivo.',
  feldartFehlt:
    'A este campo le falta «type:». MSE2 deja de leer el paquete aquí; esta aplicación lo trata como un campo de texto.',
  feldartUnbekannt:
    'MSE2 no conoce el tipo de campo «{0}» y deja de leer el paquete aquí. Esta aplicación trata el campo como texto. Tipos conocidos: {1}.',
  diagnoseStelle: 'Línea {zeile}, columna {spalte}',
  zurStelle: 'Ir a la posición',
  herkunft: 'De dónde viene un nombre',
  herkunftKeine:
    'Coloca el cursor sobre un nombre y aquí aparecerá de dónde viene y quién lo usa.',
  herkunftUnbekannt:
    '«{name}» no está definido en este paquete: el nombre procede de la biblioteca o de un paquete incluido.',
  herkunftStelle: '{art} en {datei}, línea {zeile}',
  zurDefinition: 'Ir a la definición',
  verwendungen: 'Referencias',
  verwendungenKeine: 'Ningún otro lugar menciona este nombre.',
  verwendungLaden: 'Buscando referencias …',
  symbolarten: {
    'card field': 'Campo de carta',
    'set field': 'Campo de set',
    'styling field': 'Opción de estilo',
    'extra card field': 'Campo adicional',
    script: 'Nombre de script',
  } as Record<string, string>,
  symbolartSonst: 'Definido aquí',
  aenderungen: 'Cambios antes de guardar',
  aenderungenZeigen: 'Mostrar cambios',
  aenderungenVerbergen: 'Ocultar cambios',
  aenderungenKeine: 'Ningún cambio respecto al estado leído.',
  aenderungenZusammenfassung: '{neu} línea(s) añadida(s), {weg} eliminada(s).',
  aenderungenAbschnitt: 'Cambio {nummer}',
  zeileWeg: 'eliminada',
  zeileNeu: 'añadida',
  kopieHinweis: 'Origen del paquete',
  kopieren: 'Crear copia propia',
  kopieMitgeliefert:
    'Este paquete viene con la aplicación y no se puede modificar. Crea una copia propia: aparecerá entre tus paquetes y el original se queda como está.',
  kopieFremd:
    'Este paquete procede de fuera. Un cambio aquí cuenta como cambio local y compite con la versión nueva en la próxima descarga. Crea una copia propia si quieres que perdure.',
  kopieOriginal:
    'El original se conserva. Estás trabajando en una copia: {original} no ha cambiado.',
  dateiaktionen: 'Archivos del paquete',
  dateiNeu: 'Archivo nuevo',
  dateiName: 'Nombre del archivo nuevo',
  dateiAnlegen: 'Crear',
  dateiAbbrechen: 'Cancelar',
  dateiHinzufuegen: 'Añadir imagen o fuente …',
  dateiUmbenennen: 'Renombrar',
  dateiNeuerName: 'Nombre nuevo',
  dateiUmbenennenJa: 'Renombrar',
  dateiLoeschen: 'Eliminar …',
  dateiLoeschenFrage:
    '¿Eliminar? El archivo quedará en la copia de seguridad junto al paquete.',
  dateiLoeschenJa: 'Sí, eliminar',
  dateiErstSichern: 'Guarda o descarta primero.',
  beiwerk: 'Recurso',
  beiwerkHinweis:
    'No es texto: el paquete usa este archivo tal cual. Aquí se puede renombrar y eliminar.',
  beiwerkZeile: 'Recursos ({anzahl})',
  weitereZeigen: 'Mostrar más ({gezeigt} de {gesamt})',
} as const
