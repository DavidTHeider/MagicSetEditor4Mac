// **The measurement settings** — OPS-05
// (#1722).
export const messnamen = {
  'messung.spur': '运行轨迹',
  'messung.abtastung': '记录的调用所占比例（百分比）',
  'messung.spurgrenze': '最多这么多条记录',
  'messung.aufbewahrung': '记录保留这么多分钟',
  'messung.profile': '保留的性能剖析数量',
  'zwischenspeicher.grenze': '每个磁盘缓存的上限（MB）',
} as const

export const messwerte = {
  'messung.spur.aus': '关',
  'messung.spur.an': '开',
} as const

export const messhinweise = {
  'messung.spur':
    '对服务和工具服务器的每次调用记录一行——含操作标识、路径、耗时和结果，不含内容。它保存在进程的内存中，从下次启动起生效。',
  'messung.abtastung':
    '设为 10 时，轨迹每十次调用记录一次——精确计数，而非随机。留空表示 100。',
  'messung.spurgrenze':
    '轨迹记满后，最旧的一条会被移除。范围 10 到 10000，留空表示 1000。',
  'messung.aufbewahrung':
    '较旧的记录会被丢弃。范围 1 到 1440 分钟，留空表示 60。',
  'messung.profile':
    '最新的保留在资料库旁边，较旧的会在下一次剖析时删除。范围 1 到 20，留空表示 3。',
  'zwischenspeicher.grenze':
    '适用于 Scryfall 的应答和图像，二者各自独立。启动时会依次清除最旧的内容，直至符合上限。范围 1 到 8192，留空表示 512。',
} as const
