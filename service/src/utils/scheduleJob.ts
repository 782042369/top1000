/*
 * @Author: 杨宏旋
 * @Date: 2020-07-04 23:15:34
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2025-05-07 11:48:36
 * @Description:
 */
import fs from 'node:fs/promises' // 使用Promise-based API
import path from 'node:path'

import { server } from '../core'

// 类型定义
interface SiteItem {
  /** 站点名称 */
  siteName: string
  /** 资源ID */
  siteid: string
  /** 重复度 */
  duplication: string
  /** 文件大小 */
  size: string
  /** ID */
  id: number
}

interface ProcessedData {
  /** 种子列表 */
  items: SiteItem[]
  /** 更新时间 */
  time: string
}

// 常量定义
const JSON_FILE_PATH = path.join(__dirname, '../../public/top1000.json')
const ONE_DAY_MS = 24 * 60 * 60 * 1000
const DATA_GROUP_SIZE = 3
const SITE_REGEX = /站名：(.*?) 【ID：(\d+)】/

/** 处理原始数据并返回结构化结果 */
function processData(rawData: string): ProcessedData {
  const lines = rawData.split(/\r?\n/) // 通用换行符处理
  const [timeLine = '', _v, ...dataLines] = lines

  const siteNames = new Set<string>()
  const items: SiteItem[] = []

  // 有效数据分组处理
  for (let i = 0; i <= dataLines.length - DATA_GROUP_SIZE; i += DATA_GROUP_SIZE) {
    const group = dataLines.slice(i, i + DATA_GROUP_SIZE)
    const [siteLine, dupLine = '', sizeLine = ''] = group

    const match = siteLine?.match(SITE_REGEX)
    if (!match)
      continue

    const [, siteName = '', siteid = ''] = match
    siteNames.add(siteName)

    items.push({
      siteName,
      siteid,
      duplication: dupLine.split('：')[1]?.trim() || '',
      size: sizeLine.split('：')[1]?.trim() || '',
      id: items.length + 1,
    })
  }

  return {
    time: parseTime(timeLine),
    items,
  }
}

/** 解析时间字符串 */
function parseTime(rawTime: string): string {
  return rawTime
    .replace('create time ', '')
    .replace(' by http://api.iyuu.cn/ptgen/', '')
}

/** 定时任务：获取并处理数据 */
export async function scheduleJob(): Promise<void> {
  try {
    const res = await fetch('https://api.iyuu.cn/top1000.php')
    const data = await res.text()
    const processed = processData(data)

    await fs.writeFile(
      JSON_FILE_PATH,
      JSON.stringify(processed, null, 2),
    )
    server.log.info('JSON file successfully updated')
  }
  catch (error: any) {
    server.log.error('Failed to update data:', error)
  }
}

/** 检查数据过期状态 */
export async function checkExpired(): Promise<void> {
  try {
    const rawData = await fs.readFile(JSON_FILE_PATH, 'utf8')
    const { time } = JSON.parse(rawData) as ProcessedData
    const dataTime = new Date(time).getTime()

    if (Date.now() - dataTime > ONE_DAY_MS) {
      await scheduleJob()
    }
  }
  catch (error: any) {
    server.log.error('Expiry check failed, triggering update:', error)
    await scheduleJob()
  }
}
