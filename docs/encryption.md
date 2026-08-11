# Encryption keys

Pelton stores OpenPGP keys so it can sign the mail you send and read the encrypted mail you receive. This page covers getting keys into Pelton and looking after them.

!!! note

    Key management is the first half of PGP support. Signing, encrypting and decrypting mail build on it and land separately.

## Where keys come from

Pelton does not fetch keys from anywhere. There are no keyserver lookups and no Web Key Directory requests, because both of them tell a third party which addresses you are interested in. Keys come from files you import, and nothing else.

If you already use gpg, export what you need:

```sh
# a correspondent's public key
gpg --armor --export them@example.com > them.asc

# your own key, private half included
gpg --armor --export-secret-keys you@example.com > mine.asc
```

Thunderbird, Enigmail and most other clients can export the same thing.

## Importing

**Settings, Encryption, Import a key.** Pick the file. Pelton reads both armored (`.asc`) and binary (`.gpg`) exports, and a file holding several keys imports all of them at once.

A public key lets you encrypt to that person. A private key additionally lets you sign and decrypt, and its public half is imported too so you can encrypt to yourself.

Importing a key you already have replaces it rather than adding a duplicate, so refreshing a correspondent's key after they extend its expiry is just another import.

## Where they are kept

Two files in your Pelton data directory, under `pgp/`:

| File | Holds |
| --- | --- |
| `pubring.asc` | public keys, yours and other people's |
| `secring.asc` | your private keys |

The directory and both files are readable only by your user account. A nightly build keeps its own data directory, so its keys are separate from your normal install's.

## Passphrases

A private key exported from gpg is normally protected by a passphrase. Pelton asks for it the first time it needs the key, checks it against the key immediately, and holds it in memory until you quit.

Ticking **Remember this passphrase** stores it in your operating system keyring, the same place your mail account passwords live, so it survives a restart. That is per key and off by default. **Forget passphrase** removes it from both memory and the keyring.

A passphrase never reaches Pelton's settings database, a log, or an export.

## Backups

**Keys are deliberately excluded from Pelton's settings backup.** A backup archive is a file people email to themselves and copy between machines, and a private key does not belong in one.

That means the backup you make from Settings, Import and export will not bring your keys back. Use **Export private key** in the key list to make your own copy, and keep it somewhere you would keep a password manager export.

The consequence is worth being blunt about: delete a private key with no export, and every message encrypted to it is unreadable, permanently. Pelton warns before deleting a private key for that reason.

## Signing keys per account

With more than one key that can sign, **Signing keys** at the bottom of the pane pins which key each account uses.

Left on **Match by address**, Pelton looks for a private key whose user id carries that account's address. That is the right answer when each account has its own key. Pin one explicitly when it is not, for instance when several addresses share a single key.

Deleting a key that an account was pinned to resets that account to matching by address, rather than leaving it pointing at a key that no longer exists.

## S/MIME signatures

S/MIME is the other way mail gets signed, and the one corporate and government senders normally use. Instead of a key you exchange with someone directly, it uses a certificate issued by a certificate authority, the same kind of authority that vouches for websites.

Pelton checks those signatures on mail you receive. It cannot send S/MIME signed or encrypted mail; PGP is the path for that.

A signed message shows a badge in the reading pane, alongside the PGP one and controlled by the same setting:

| Badge | Meaning |
| --- | --- |
| **Signature verified** | The message has not been altered, the certificate is valid and trusted by your system, and it belongs to the sender. |
| **Signature unverified** | The certificate cannot be vouched for: an authority your computer does not trust, an expired certificate, or one issued to somebody other than the sender. Hover the badge for which. |
| **Signature failed** | The message was altered after it was signed. |

Mail with no signature shows no badge, which is nearly all of it.

Checking happens while the mail is downloaded, so it works offline afterwards and adds no delay when you open a message. Messages already in your cache from before this existed show no badge until they sync again.

Two limits worth knowing. Pelton does not ask the certificate authority whether a certificate has been revoked since it was issued, because that would tell the authority which mail you are reading; a revoked certificate still shows as verified. And the verdict is the one recorded when the message arrived, which is the right question for a signature: a certificate expiring later does not make an older message a forgery.
