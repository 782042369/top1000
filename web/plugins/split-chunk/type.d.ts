import type { BuildEnvironmentOptions } from 'vite'

type SingleArrayType<T> = T extends (infer U)[] ? U : T
export type rollupOptions = Exclude<BuildEnvironmentOptions['rollupOptions'], undefined>
export type ManualChunksOption = Exclude<SingleArrayType<rollupOptions['output']>, undefined>['manualChunks']
export type GetModuleInfo = Exclude<Parameters<Exclude<Exclude<ManualChunksOption, undefined>, Record<string, string[]>>>[number], string>['getModuleInfo']
