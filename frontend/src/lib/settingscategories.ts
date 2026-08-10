// settingscategories.ts is the shared list of settings categories: the panel's
// own navigation and the command palette's "#" mode both build from it, so a
// category added in one place is reachable from the other without being listed
// twice.

import type { ComponentType } from 'svelte'
import {
  IconPalette,
  IconBrush,
  IconMenu2,
  IconLanguage,
  IconList,
  IconEye,
  IconWriting,
  IconSignature,
  IconBell,
  IconHandMove,
  IconShieldLock,
  IconMailbox,
  IconPlugConnected,
  IconBatteryEco,
  IconKeyboard,
  IconInfoCircle,
} from '@tabler/icons-svelte'

/** One settings category: a pane of the settings panel. */
export interface SettingsCategory {
  /** The value SettingsPanel's initialCategory takes to deep-link here. */
  key: string
  /** The nav group it sits under. */
  group: string
  /** Localized display name. */
  label: string
  /** Localized synonyms, so search finds a pane by what it does. */
  keywords: string
  /** Bundled fallback icon. */
  icon: ComponentType
  /** Tabler name a theme can override the icon by. */
  iconName: string
}

/** One top-level nav group, in display order. */
export interface SettingsGroup {
  key: string
  label: string
}

/** Builds the category list with the given translator. */
export function settingsCategories(t: (key: string) => string): SettingsCategory[] {
  return [
    { key: 'appearance', group: 'appearance', label: t('settingsPanel.category.appearance'), keywords: t('settingsPanel.keywords.appearance'), icon: IconPalette, iconName: 'palette' },
    { key: 'themes', group: 'appearance', label: t('settingsPanel.category.themes'), keywords: t('settingsPanel.keywords.themes'), icon: IconBrush, iconName: 'brush' },
    { key: 'menubar', group: 'appearance', label: t('settingsPanel.category.menubar'), keywords: t('settingsPanel.keywords.menubar'), icon: IconMenu2, iconName: 'menu-2' },
    { key: 'language', group: 'appearance', label: t('settings.language'), keywords: t('settingsPanel.keywords.language'), icon: IconLanguage, iconName: 'language' },
    { key: 'list', group: 'mail', label: t('settingsPanel.category.messageList'), keywords: t('settingsPanel.keywords.list') + ' ' + t('settingsPanel.keywords.sidebar') + ' ' + t('settingsPanel.keywords.avatars'), icon: IconList, iconName: 'list' },
    { key: 'display', group: 'mail', label: t('settingsPanel.category.reading'), keywords: t('settingsPanel.keywords.display') + ' ' + t('settingsPanel.keywords.panes'), icon: IconEye, iconName: 'eye' },
    { key: 'composing', group: 'mail', label: t('settingsPanel.category.composingSending'), keywords: t('settingsPanel.keywords.composing') + ' ' + t('settingsPanel.keywords.sending'), icon: IconWriting, iconName: 'writing' },
    { key: 'signatures', group: 'mail', label: t('settingsPanel.category.signatures'), keywords: t('settingsPanel.keywords.signatures'), icon: IconSignature, iconName: 'signature' },
    { key: 'notifications', group: 'mail', label: t('settingsPanel.category.notifications'), keywords: t('settingsPanel.keywords.notifications'), icon: IconBell, iconName: 'bell' },
    { key: 'gestures', group: 'mail', label: t('settingsPanel.category.gestures'), keywords: t('settingsPanel.keywords.gestures'), icon: IconHandMove, iconName: 'hand-move' },
    { key: 'privacy', group: 'privacy', label: t('settingsPanel.category.privacyNetwork'), keywords: t('settingsPanel.keywords.privacy') + ' ' + t('settingsPanel.keywords.network'), icon: IconShieldLock, iconName: 'shield-lock' },
    { key: 'mailboxes', group: 'accounts', label: t('settingsPanel.category.accounts'), keywords: t('settingsPanel.keywords.mailboxes') + ' ' + t('settingsPanel.keywords.contacts'), icon: IconMailbox, iconName: 'mailbox' },
    { key: 'external', group: 'accounts', label: t('settingsPanel.category.integrations'), keywords: t('settingsPanel.keywords.external') + ' ' + t('settingsPanel.keywords.sync'), icon: IconPlugConnected, iconName: 'plug-connected' },
    { key: 'power', group: 'advanced', label: t('settingsPanel.category.powerSync'), keywords: t('settingsPanel.keywords.power') + ' ' + t('settingsPanel.keywords.offline'), icon: IconBatteryEco, iconName: 'battery-eco' },
    { key: 'shortcuts', group: 'advanced', label: t('settingsPanel.category.shortcuts'), keywords: t('settingsPanel.keywords.shortcuts'), icon: IconKeyboard, iconName: 'keyboard' },
    { key: 'about', group: 'about', label: t('settingsPanel.category.about'), keywords: t('settingsPanel.keywords.about'), icon: IconInfoCircle, iconName: 'info-circle' },
  ]
}

/** Builds the nav groups with the given translator. */
export function settingsGroups(t: (key: string) => string): SettingsGroup[] {
  return [
    { key: 'appearance', label: t('settingsPanel.group.appearance') },
    { key: 'mail', label: t('settingsPanel.group.mail') },
    { key: 'privacy', label: t('settingsPanel.group.privacy') },
    { key: 'accounts', label: t('settingsPanel.group.accounts') },
    { key: 'advanced', label: t('settingsPanel.group.advanced') },
    { key: 'about', label: t('settingsPanel.group.about') },
  ]
}
