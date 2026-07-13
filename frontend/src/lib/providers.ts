// providers.ts holds the wizard's provider presets. each preset seeds the server
// settings and auth method so the user usually only types an address and signs
// in or enters a password. custom uses autodiscovery instead of fixed hosts.

export type AuthKind = 'oauth' | 'password'

export interface ProviderPreset {
  id: string
  label: string
  kind: AuthKind
  // oauthProvider is the backend provider key for oauth presets.
  oauthProvider?: 'google' | 'microsoft'
  imapHost?: string
  imapPort?: number
  smtpHost?: string
  smtpPort?: number
  // note is shown under the form (e.g. app-specific password guidance).
  note?: string
  // appPasswordUrl, when set, replaces the plain note with a warning box that
  // spells out that the provider's regular password will not work and links to
  // where an app-specific password is generated.
  appPasswordUrl?: string
  // custom means "use autodiscovery and let the user edit servers".
  custom?: boolean
  // oauthOptional marks a password-first preset that can still switch to the
  // oauth sign-in form (for users who bring their own client id).
  oauthOptional?: boolean
  // allowClientSecret reveals an optional oauth client-secret field, for
  // providers whose app registration may be a confidential client.
  allowClientSecret?: boolean
}

export const providerPresets: ProviderPreset[] = [
  {
    // password-first: an app password works today on consumer accounts with
    // no Google Cloud Console involved. oauth stays reachable behind
    // oauthOptional for users who registered their own client (#56).
    id: 'gmail',
    label: 'Gmail',
    kind: 'password',
    oauthProvider: 'google',
    imapHost: 'imap.gmail.com',
    imapPort: 993,
    smtpHost: 'smtp.gmail.com',
    smtpPort: 465,
    appPasswordUrl: 'https://myaccount.google.com/apppasswords',
    oauthOptional: true,
  },
  {
    id: 'outlook',
    label: 'Outlook / Microsoft 365',
    kind: 'oauth',
    oauthProvider: 'microsoft',
    imapHost: 'outlook.office365.com',
    imapPort: 993,
    smtpHost: 'smtp.office365.com',
    smtpPort: 587,
    allowClientSecret: true,
  },
  {
    id: 'icloud',
    label: 'iCloud',
    kind: 'password',
    imapHost: 'imap.mail.me.com',
    imapPort: 993,
    smtpHost: 'smtp.mail.me.com',
    smtpPort: 587,
    appPasswordUrl: 'https://appleid.apple.com',
  },
  {
    id: 'yahoo',
    label: 'Yahoo',
    kind: 'password',
    imapHost: 'imap.mail.yahoo.com',
    imapPort: 993,
    smtpHost: 'smtp.mail.yahoo.com',
    smtpPort: 465,
    note: 'Yahoo requires an app password generated in account security.',
  },
  {
    id: 'fastmail',
    label: 'Fastmail',
    kind: 'password',
    imapHost: 'imap.fastmail.com',
    imapPort: 993,
    smtpHost: 'smtp.fastmail.com',
    smtpPort: 465,
    note: 'Fastmail requires an app password.',
  },
  {
    id: 'purelymail',
    label: 'Purelymail',
    kind: 'password',
    imapHost: 'imap.purelymail.com',
    imapPort: 993,
    smtpHost: 'smtp.purelymail.com',
    smtpPort: 465,
    note: 'Use your Purelymail password, or an app password if you enabled 2FA.',
  },
  {
    id: 'custom',
    label: 'Other (IMAP / SMTP)',
    kind: 'password',
    custom: true,
    note: 'We will try to auto-detect your server settings from your address.',
  },
]
