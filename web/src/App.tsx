/*
 * @Author: yanghongxuan
 * @Date: 2025-02-08 21:16:06
 * @Description:
 * @LastEditTime: 2025-04-11 17:00:33
 * @LastEditors: yanghongxuan
 */
import type { ColDef, ICellEditorParams } from 'ag-grid-community'

import { AG_GRID_LOCALE_CN } from '@ag-grid-community/locale'
import { ClientSideRowModelModule, CustomFilterModule, DateFilterModule, LocaleModule, ModuleRegistry, NumberFilterModule, TextFilterModule, ValidationModule } from 'ag-grid-community'
import { GroupFilterModule, LicenseManager, MultiFilterModule, SetFilterModule } from 'ag-grid-enterprise'
import { AgGridReact } from 'ag-grid-react'
import { useMount } from 'ahooks'

import type {
  DataType,
  ResDataType,
} from './types'

import { convertSizeToKb, ptUrlConfig } from './utils'

LicenseManager.setLicenseKey(
  '[v3][RELEASE][0102]_NDg2Njc4MzY3MDgzNw==16d78ca762fb5d2ff740aed081e2af7b',
)
// https://www.ag-grid.com/vue-data-grid/modules/
ModuleRegistry.registerModules([
  ClientSideRowModelModule,
  // ag-grid-community
  TextFilterModule,
  NumberFilterModule,
  DateFilterModule,
  SetFilterModule,
  MultiFilterModule,
  GroupFilterModule,
  CustomFilterModule,
  ValidationModule,
  LocaleModule,
])

const App: React.FC = () => {
  const [rowData, setRowData] = useState<DataType[]>([])

  const colDefs = useMemo<Array<ColDef<DataType>>>(() => [
    {
      headerName: '名字',
      field: 'siteName',
      filter: true,
    },
    {
      headerName: '资源ID',
      field: 'siteid',
    },
    {
      headerName: '重复度',
      field: 'duplication',
      sortable: true,
    },
    {
      headerName: '文件大小',
      field: 'size',
      sortable: true,
      comparator: (valueA, valueB) => {
        return convertSizeToKb(valueA) - convertSizeToKb(valueB)
      },
    },
    {
      headerName: '操作',
      cellRenderer: (params: ICellEditorParams<DataType>) => {
        const { siteName } = params.data
        let { siteid } = params.data
        const getUrl
          = ptUrlConfig[siteName === 'ptlsp' ? 'audiences' : siteName]
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
  ], [])

  /* 获取数据 */
  useMount(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('./top1000.json')
        const json: ResDataType = await response.json()
        setRowData(json.items)
      }
      catch (error) {
        console.error('Error:', error)
      }
    }
    fetchData()
  })

  return (
    <div style={{ height: '100vh', width: '100vw' }}>
      <AgGridReact<DataType>
        rowData={rowData}
        columnDefs={colDefs}
        localeText={AG_GRID_LOCALE_CN}
        defaultColDef={
          {
            flex: 1,
            sortable: false,
          }
        }
        getRowId={params => `${params.data.id}`}
      />
    </div>
  )
}

export default App
