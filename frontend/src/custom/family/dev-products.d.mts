import type { Plugin } from 'vite'
export function familyProductEntry(url?: string, method?: string): { redirect: boolean; url: string } | null
export function familyProductDevEntries(): Plugin
