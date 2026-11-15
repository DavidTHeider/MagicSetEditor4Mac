// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'Отключённый раздел исчезает из боковой панели и палитры, а его пути отклоняются на каждой поверхности. Данные остаются — повторное включение всё восстанавливает.',
    hinweis:
      'Переключатель скрывает, но не защищает: чтобы что-то защитить, используй роли и доступы.',
    reife: {
      stabil: 'стабильно',
      vorschau: 'Предпросмотр',
      auslaufend: 'выводится из употребления',
    },
    saetze: {
      sammeln: 'Коллекция, владение и списки желаний',
      spielen: 'Колоды, турниры, партии и стол',
      handeln: 'Обмены, доступы и конверты',
      ressourcen: 'Пакеты шаблонов и шаблоны',
      ki: 'Предложения, изображения, поисковый индекс, счета и расходы',
      netz: 'Идентичность, контакты, приглашения и доступность',
      markt: 'Предложения, цены и счета торговых площадок',
      turniere: 'Турниры с раундами, парами и результатами',
      dienst:
        'REST для скриптов — работающий сервис останавливается при отключении',
      werkzeugserver:
        'MCP по HTTP — работающий сервер останавливается при отключении',
    },
    abgeschaltet: {
      titel: '{seite}: отключено',
      erklaerung:
        'Включи раздел снова в «Настройки › Разделы и функции». Его данные на месте и не изменены.',
      oeffnen: 'Открыть разделы и функции',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'Коллекция',
  'schalter.spielen': 'Игра',
  'schalter.handeln': 'Обмен',
  'schalter.ressourcen': 'Ресурсы',
  'schalter.ki': 'ИИ',
  'schalter.netz': 'Сеть и собеседники',
  'schalter.markt': 'Торговые площадки и цены',
  'schalter.turniere': 'Турниры',
  'schalter.dienst': 'Сервис для скриптов',
  'schalter.werkzeugserver': 'Сервер инструментов по HTTP',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'вкл.',
  'schalter.sammeln.aus': 'выкл.',
  'schalter.spielen.an': 'вкл.',
  'schalter.spielen.aus': 'выкл.',
  'schalter.handeln.an': 'вкл.',
  'schalter.handeln.aus': 'выкл.',
  'schalter.ressourcen.an': 'вкл.',
  'schalter.ressourcen.aus': 'выкл.',
  'schalter.ki.an': 'вкл.',
  'schalter.ki.aus': 'выкл.',
  'schalter.netz.an': 'вкл.',
  'schalter.netz.aus': 'выкл.',
  'schalter.markt.an': 'вкл.',
  'schalter.markt.aus': 'выкл.',
  'schalter.turniere.an': 'вкл.',
  'schalter.turniere.aus': 'выкл.',
  'schalter.dienst.an': 'вкл.',
  'schalter.dienst.aus': 'выкл.',
  'schalter.werkzeugserver.an': 'вкл.',
  'schalter.werkzeugserver.aus': 'выкл.',
}
