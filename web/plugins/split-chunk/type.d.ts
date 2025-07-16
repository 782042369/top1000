import type { BuildEnvironmentOptions } from 'vite'

type SingleArrayType<T> = T extends (infer U)[] ? U : T
export type rollupOptions = Exclude<BuildEnvironmentOptions['rollupOptions'], undefined>
export type ManualChunksOption = Exclude<SingleArrayType<Exclude<SingleArrayType<rollupOptions['output']>, undefined>['advancedChunks']['groups']>['name'], string>
export type GetModuleInfo = Parameters<ManualChunksOption>['1']['getModuleInfo']
