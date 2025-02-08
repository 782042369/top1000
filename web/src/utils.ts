/* 文件大小转换为 KB */
export function convertSizeToKb(sizeStr: string) {
  const sizeUnits = {
    KB: 1,
    MB: 1024,
    GB: 1024 * 1024,
    TB: 1024 * 1024 * 1024,
  };
  const match = sizeStr.match(/([\d.]+)\s*(KB|MB|GB|TB)/i);
  return match
    ? parseFloat(match[1]) *
        sizeUnits[match[2].toUpperCase() as keyof typeof sizeUnits]
    : 0; // 优化：使用 keyof 来确保安全
}
