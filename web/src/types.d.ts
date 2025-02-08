/* eslint-disable @typescript-eslint/no-unused-vars */
/*
 * @Author: yanghongxuan
 * @Date: 2025-02-08 21:31:56
 * @Description:
 * @LastEditTime: 2025-02-08 21:45:54
 * @LastEditors: yanghongxuan
 */
import { ptUrlConfig } from '@/config';
import type { TableProps } from 'antd';

namespace API {
  /** 种子详情 */
  interface DataType {
    /** 站点名称 */
    siteName: keyof typeof ptUrlConfig;
    /** 资源ID */
    siteid: string;
    /** 重复度 */
    duplication: string;
    /** 文件大小 */
    mainTitle: string;
    /** 副标题 */
    subTitle: string;
    /** 文件大小 */
    size: string;
    /** ID */
    id: number;
  }
  /** 接口返回 */
  interface ResDataType {
    /** 种子列表 */
    items: DataType[];
    /** 更新时间 */
    time: string;
    /** 站点名称集合 */
    siteName: string[];
  }
}

/* 类型定义 */
type TableChangeHandler = NonNullable<TableProps<API.DataType>['onChange']>;
type FilterParams = Parameters<TableChangeHandler>[1];
type GetSingle<T> = T extends (infer U)[] ? U : never;
type SortParams = GetSingle<Parameters<TableChangeHandler>[2]>;
