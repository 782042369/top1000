/*
 * @Author: yanghongxuan
 * @Date: 2025-02-08 21:31:56
 * @Description:
 * @LastEditTime: 2025-04-08 14:44:06
 * @LastEditors: yanghongxuan
 */

/** 种子详情 */
export interface DataType {
  /** 站点名称 */
  siteName: string
  /** 资源ID */
  siteid: string
  /** 重复度 */
  duplication: string
  /** 文件大小 */
  mainTitle: string
  /** 副标题 */
  subTitle: string
  /** 文件大小 */
  size: string
  /** ID */
  id: number
}
/** 接口返回 */
export interface ResDataType {
  /** 种子列表 */
  items: DataType[]
  /** 更新时间 */
  time: string
  /** 站点名称集合 */
  siteName: string[]
}
