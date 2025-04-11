/*
 * @Author: yanghongxuan
 * @Date: 2025-02-08 21:16:06
 * @Description:
 * @LastEditTime: 2025-04-11 14:55:44
 * @LastEditors: yanghongxuan
 */
import type { TableColumnsType } from 'antd'

import { useEventListener } from 'ahooks'
import zhCN from 'antd/es/locale/zh_CN'

import './index.css'

import type {
  API,
  FilterParams,
  SortParams,
  TableChangeHandler,
} from './types'

import { convertSizeToKb, ptUrlConfig } from './utils'

const App: React.FC = () => {
  const [filterParams, setFilterParams] = useState<FilterParams>({})
  const [sortParams, setSortParams] = useState<SortParams>({})
  const [responseData, setResponseData] = useState<API.ResDataType>({
    items: [],
    time: '',
    siteName: [],
  })
  const [siteOptions, setSiteOptions] = useState<
    { text: string, value: string }[]
  >([])
  const taskContainerRef = useRef<HTMLDivElement>(null)
  const [tableHeight, setTableHeight] = useState<number>(500)

  /* 处理表格变化 */
  const handleTableChange: TableChangeHandler = (
    _pagination,
    filters,
    sorter,
  ) => {
    setFilterParams(filters)
    setSortParams(sorter as SortParams)
  }

  /* 获取数据 */
  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('./top1000.json')
        const json: API.ResDataType = await response.json()
        setResponseData({
          items: json.items,
          time: json.time,
          siteName: json.siteName,
        })
        setSiteOptions(
          json.siteName.map(item => ({ text: item, value: item })),
        )
      } catch (error) {
        console.error('Error:', error)
      }
    }
    fetchData()
  }, [])

  /* 表格列定义 */
  const columns: TableColumnsType<API.DataType> = [
    {
      title: '名字',
      dataIndex: 'siteName',
      key: 'siteName',
      filters: siteOptions,
      filteredValue: filterParams.siteName || null,
      onFilter: (value: any, record) => record.siteName.includes(value),
      ellipsis: true,
      filterMode: 'tree',
      filterSearch: true,
    },
    {
      title: '资源ID',
      dataIndex: 'siteid',
      key: 'siteid',
    },
    {
      title: '重复度',
      dataIndex: 'duplication',
      key: 'duplication',
      sorter: (a, b) =>
        Number.parseInt(a.duplication) - Number.parseInt(b.duplication),
      sortOrder:
        sortParams.columnKey === 'duplication' ? sortParams.order : null,
    },
    {
      title: '文件大小',
      dataIndex: 'size',
      key: 'size',
      sorter: (a, b) => convertSizeToKb(a.size) - convertSizeToKb(b.size),
      sortOrder: sortParams.columnKey === 'size' ? sortParams.order : null,
    },
    {
      title: '操作',
      key: 'action',
      render: (_text, record) => {
        const { siteName } = record
        let { siteid } = record
        const getUrl =
          ptUrlConfig[siteName === 'ptlsp' ? 'audiences' : siteName]
        if (!getUrl) {
          return null
        }
        if (siteName === 'ptlsp') {
          siteid = {
            649: '297203',
            8667: '353903',
            8765: '288867',
            default: '297203',
          }[siteid] as string
        }
        const downloadUrl = getUrl.download(siteid)
        return (
          <div>
            <a
              href={`${getUrl.details(siteid)}`}
              target="_blank"
              rel="noreferrer"
            >
              {downloadUrl ? `查看详情` : `查看详情(下载到详情页面)`}
            </a>
            {downloadUrl
              ? (
                  <a
                    style={{ marginLeft: '10px' }}
                    href={`${getUrl.download(siteid)}`}
                    target="_blank"
                    rel="noreferrer"
                  >
                    下载种子
                  </a>
                )
              : null}
          </div>
        )
      },
    },
  ]

  /* 计算表格高度 */
  useEventListener('resize', () => {
    const taskContainer = taskContainerRef.current
    if (taskContainer && window.innerHeight) {
      const headerElement = taskContainer.querySelector('.ant-table-header')
      const { height = 0 } = headerElement?.getBoundingClientRect() || {}
      setTableHeight(window.innerHeight - height)
    }
  }, {
    target: window
  })
  const x = window.innerWidth
  return (
    <ConfigProvider locale={zhCN}>
      <div ref={taskContainerRef}>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={responseData.items}
          onChange={handleTableChange}
          pagination={false}
          scroll={{ x, y: tableHeight }}
          virtual={true}
          style={{ height: '100vh' }}
          size="small"
        />
      </div>
    </ConfigProvider>
  )
}

export default App
