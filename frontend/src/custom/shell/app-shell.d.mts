import type { Plugin } from 'vite'
export function earlyScript(locales?: Record<string, string>, base?: string): string
export function renderAppShell(html: string, locales?: Record<string, string>, base?: string): string
export function localeChunks(bundle?: Record<string, unknown>): Record<string, string>
export function mofaAppShell(): Plugin
