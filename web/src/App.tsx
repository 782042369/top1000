import type { TableColumnsType, TableProps } from 'antd';
import zhCN from 'antd/es/locale/zh_CN';
import './index.css';
type OnChange = NonNullable<TableProps<DataType>['onChange']>;
type Filters = Parameters<OnChange>[1];

type GetSingle<T> = T extends (infer U)[] ? U : never;
type Sorts = GetSingle<Parameters<OnChange>[2]>;
const ptUrlConfig = {
  hdhome: 'https://hdhome.org',
  btschool: 'https://pt.btschool.club',
  ttg: 'https://totheglory.im',
  'm-team': 'https://kp.m-team.cc',
  torrentccf: 'https://et8.org',
  audiences: 'https://audiences.me',
  greatposterwall: 'https://greatposterwall.com',
  keepfrds: 'https://pt.keepfrds.com',
  pter: 'https://pterclub.com',
  hdsky: 'https://hdsky.me',
  chdbits: 'https://chdbits.co',
  ssd: 'https://springsunday.net',
  ourbits: 'https://ourbits.club',
  ptsbao: 'https://ptsbao.club',
  nanyangpt: 'https://nanyangpt.com',
  pthome: 'https://pthome.net',
  tjupt: 'https://tjupt.org',
  upxin: 'https://pt.upxin.net',
  hd4fans: 'https://pt.hd4fans.org',
  hhanclub: 'https://hhanclub.top',
  hdtime: 'https://hdtime.org',
  hdarea: 'https://hdarea.club',
  hdzone: 'http://www.hdzone.org',
  '1ptba': 'https://1ptba.com',
  rousi: 'https://rousi.zip',
  cyanbug: 'https://cyanbug.net',
  zmpt: 'https://zmpt.club',
  hdfans: 'https://hdfans.org',
  hdatmos: 'https://hdatmos.club',
  piggo: 'https://piggo.me',
  hddolby: 'https://www.hddolby.com/',
  crabpt: 'https://crabpt.vip',
  discfan: 'https://discfan.net',
  '52pt': 'https://52pt.size',
  ubits: 'https://ubits.club',
  agsvpt: 'https://www.agsvpt.com',
  eastgame: 'https://pt.eastgame.org',
  tosky: 'https://t.tosky.club',
  icc2022: 'https://www.icc2022.com',
  carpt: 'https://carpt.net',
  qingwapt: 'https://qingwapt.com',
  oshen: 'https://www.oshen.win',
  hitpt: 'https://www.hitpt.com',
  yemapt: 'https://www.yemapt.org',
  pandapt: 'https://pandapt.net',
  monikadesign: 'https://monikadesign.uk',
  hdvideo: 'https://hdvideo.one',
  dmhy: 'https://u2.dmhy.org',
  hdcity: 'https://hdcity.city',
  dajiao: 'https://dajiao.cyou',
} as const;
interface DataType {
  siteName: keyof typeof ptUrlConfig;
  siteid: string;
  duplication: string;
  mainTitle: string;
  subTitle: string;
  size: string;
  id: number;
}
interface ResDataType {
  items: DataType[];
  time: string;
  siteName: string[];
}

function sizeToKb(sizeStr: string) {
  const units = { KB: 1, MB: 1024, GB: 1024 * 1024, TB: 1024 * 1024 * 1024 };
  const match = sizeStr.match(/([\d.]+)\s*(KB|MB|GB|TB)/i);
  if (match) {
    return parseFloat(match[1]) * units[match[2].toUpperCase() as never];
  }
  return 0;
}
const App: React.FC = () => {
  const [filteredInfo, setFilteredInfo] = useState<Filters>({});
  const [sortedInfo, setSortedInfo] = useState<Sorts>({});
  const [resData, setResData] = useState<ResDataType>({
    items: [],
    time: '',
    siteName: [],
  });
  const [siteName, setSiteName] = useState<{ text: string; value: string }[]>(
    [],
  );
  const taskDom = useRef<HTMLDivElement>(null);
  const [tableY, setTableY] = useState<number>(500);

  const handleChange: OnChange = (_pagination, filters, sorter) => {
    setFilteredInfo(filters);
    setSortedInfo(sorter as Sorts);
  };
  useEffect(() => {
    fetch('https://top1000.939593.xyz/top1000.json')
      .then(response => response.json())
      .then((json: ResDataType) => {
        if (json) {
          setResData({
            items: json.items,
            time: json.time,
            siteName: json.siteName,
          });
          setSiteName(
            json.siteName.map(item => {
              return { text: item, value: item };
            }),
          );
        }
      })
      .catch(error => {
        console.error('Error:', error);
      });
  }, []);
  const columns: TableColumnsType<DataType> = [
    {
      title: '名字',
      dataIndex: 'siteName',
      key: 'siteName',
      filters: siteName,
      filteredValue: filteredInfo.siteName || null,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      onFilter: (value: any, record) => record.siteName.includes(value),
      ellipsis: true,
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
      sorter: (a, b) => parseInt(a.duplication) - parseInt(b.duplication),
      sortOrder:
        sortedInfo.columnKey === 'duplication' ? sortedInfo.order : null,
    },
    {
      title: '文件大小',
      dataIndex: 'size',
      key: 'size',
      sorter: (a, b) => sizeToKb(a.size) - sizeToKb(b.size),
      sortOrder: sortedInfo.columnKey === 'size' ? sortedInfo.order : null,
    },
    {
      title: '操作',
      key: 'action',
      render: (_text, record) => {
        const url = ptUrlConfig[record.siteName];
        return !url ? (
          ''
        ) : (
          <a
            href={`${url}/details.php?id=${record.siteid}&hit=1`}
            target="_blank"
            rel="noreferrer"
          >
            查看
          </a>
        );
      },
    },
  ];
  useEffect(() => {
    const tableParBox = taskDom.current;
    if (tableParBox && window.innerHeight) {
      const thBox = tableParBox.querySelector('.ant-table-header');
      const { height = 0 } = thBox?.getBoundingClientRect() || {};
      const h = window.innerHeight - height;
      setTableY(h);
    }
  }, [taskDom]);
  return (
    <ConfigProvider locale={zhCN}>
      <div ref={taskDom}>
        <Table
          rowKey={'id'}
          columns={columns}
          dataSource={resData.items}
          onChange={handleChange}
          pagination={false}
          scroll={{ x: 'max-content', y: tableY }}
          virtual={true}
          style={{
            height: '100vh',
          }}
          size="small"
        />
      </div>
    </ConfigProvider>
  );
};

export default App;
