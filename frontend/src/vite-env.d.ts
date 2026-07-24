/// <reference types="svelte" />
/// <reference types="vite/client" />

// The Vite-aliased Tabler icon geometry (see vite.config.ts). Each icon name
// maps to a list of svg child nodes as [tag, attributes] pairs.
declare module 'tabler-nodes-outline' {
  type IconNode = [string, Record<string, string | number>]
  const nodes: Record<string, IconNode[]>
  export default nodes
}
