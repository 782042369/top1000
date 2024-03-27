/*
 * @Author: yanghongxuan
 * @Date: 2023-07-10 12:22:55
 * @LastEditors: yanghongxuan
 * @LastEditTime: 2023-07-10 12:25:54
 * @Description:
 */
declare module "mime-db" {
  interface MimeEntry {
    source: string;
    extensions?: string[];
    [key: string]: any;
  }

  const db: { [type: string]: MimeEntry };

  export = db;
}
