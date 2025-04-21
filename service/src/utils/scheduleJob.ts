/*
 * @Author: 杨宏旋
 * @Date: 2020-07-04 23:15:34
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2025-04-21 09:41:04
 * @Description:
 */
import axios from 'axios'
import schedule from 'node-schedule'
import fs from 'node:fs'
import https from 'node:https'
import path from 'node:path'

import { server } from '../core'

function handleJsonData(data: string) {
  // 解析内容并创建JSON对象
  const lines = data.split('\r\n')
  const [time, _v, ...linesData] = lines
  const items: {
    siteName: string
    siteid: string
    duplication: string
    mainTitle: string
    subTitle: string
    size: string
    id: number
  }[] = []
  const siteName: string[] = []
  const regex = /站名：(.*?) 【ID：(\d+)】/
  for (let i = 0; i < linesData.length; i += 5) {
    // 假设每6行为一组数据
    const match = linesData[i].match(regex)
    const duplication = linesData[i + 1]?.split('：')[1]?.trim()
    const mainTitle = linesData[i + 2]?.split('：')[1]?.trim()
    const subTitle = linesData[i + 3]?.split('：')[1]?.trim()
    const size = linesData[i + 4]?.split('：')[1]?.trim()
    if (match?.[1]) {
      !siteName.includes(match[1]) && siteName.push(match[1])
      items.push({
        siteName: match?.[1] || '', // 网站名
        siteid: match?.[2] || '', // id
        duplication,
        mainTitle,
        subTitle,
        size,
        id: i / 5 + 1,
      })
    }
  }

  // 写入JSON文件
  const jsonFilePath = path.join(__dirname, '../../dist/top1000.json')
  fs.writeFile(
    jsonFilePath,
    JSON.stringify(
      {
        time,
        items,
        siteName,
      },
      null,
      2,
    ),
    (err) => {
      if (err) {
        server.log.error('Error writing JSON file:', err)
        return
      }
      server.log.info('JSON file was successfully created.')
    },
  )
}
function getTop1000() {
  // 创建一个新的 httpsAgent 并设置 rejectUnauthorized 为 false
  const agent = new https.Agent({
    rejectUnauthorized: false,
  })
  axios
    .get('https://api.iyuu.cn/top1000.php', { httpsAgent: agent })
    .then((res) => {
      if (res.data) {
        handleJsonData(res.data)
      }
    })
    .catch((err) => {
      server.log.error(err)
    })
}
// 定时任务
function scheduleCronstyle() {
  getTop1000()
  schedule.scheduleJob('0 09 * * *', () => {
    try {
      getTop1000()
    }
    catch (error) {
      server.log.error(error)
    }
  })
}
export default scheduleCronstyle
