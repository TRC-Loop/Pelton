// vip.ts holds the client-side mirror of the VIP-sender list (#126). The
// message list gets each row's VIP state from the backend (summary.senderVip),
// but this store lets the star update live the moment a sender is marked or
// unmarked, without waiting for a refetch, and backs the settings modal.

import { writable, get } from 'svelte/store'
import { listVIPSenders, addVIPSender, removeVIPSender } from '../lib/api'

// vipSenders is the set of bare lowercased VIP addresses.
export const vipSenders = writable<Set<string>>(new Set())

// bareAddress extracts the lowercased address from a "Name <addr>" or bare
// value, mirroring the backend's bareAddress so client and server agree.
export function bareAddress(value: string): string {
  let s = value.trim()
  const lt = s.lastIndexOf('<')
  if (lt >= 0) {
    const gt = s.indexOf('>', lt)
    if (gt >= 0) {
      s = s.slice(lt + 1, gt)
    }
  }
  return s.trim().toLowerCase()
}

// loadVIPSenders fetches the list from the backend into the store.
export async function loadVIPSenders(): Promise<void> {
  const list = await listVIPSenders()
  vipSenders.set(new Set(list))
}

// isVIPAddress reports whether a from-address is a VIP, per the current store.
export function isVIPAddress(fromAddress: string): boolean {
  return get(vipSenders).has(bareAddress(fromAddress))
}

// addVIP marks an address as VIP on the backend and updates the store.
export async function addVIP(fromAddress: string): Promise<void> {
  const addr = bareAddress(fromAddress)
  if (!addr) {
    return
  }
  await addVIPSender(addr)
  vipSenders.update((s) => new Set(s).add(addr))
}

// removeVIP drops an address from the VIP list on the backend and the store.
export async function removeVIP(fromAddress: string): Promise<void> {
  const addr = bareAddress(fromAddress)
  await removeVIPSender(addr)
  vipSenders.update((s) => {
    const next = new Set(s)
    next.delete(addr)
    return next
  })
}
