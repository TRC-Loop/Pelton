# Scanning links and attachments

Links and attachments in email are how most malware and phishing arrives. Pelton can check either one against [VirusTotal](https://www.virustotal.com), which aggregates around seventy antivirus engines and URL blocklists into a single verdict.

The integration is **off by default** and stays off until you supply your own VirusTotal API key. Nothing is ever checked without you turning it on.

## What this costs you in privacy

Be clear about the trade before you enable it. Asking VirusTotal about a link means telling VirusTotal that link exists and that someone is interested in it. That is a real disclosure to a third party, and it is the reason this is opt-in rather than a default.

Two things limit the exposure:

- **Attachments are never uploaded.** Pelton sends the SHA-256 hash of the file, never the file. A hash cannot be turned back into your document. If VirusTotal has never seen a file with that hash, the answer is "unknown" and it stays unknown, because Pelton will not offer to upload it to find out.
- **Nothing is submitted for analysis.** Pelton only reads existing results. A link you scan is not queued, crawled or added to anyone's dataset.

What does leave your machine is the URL itself when you scan a link, and the hash when you scan an attachment. Your mail, its contents, its sender and its recipients never do.

## Setting it up

1. Create a free account at [virustotal.com](https://www.virustotal.com/gui/join-us).
2. Open your API key at [virustotal.com/gui/my-apikey](https://www.virustotal.com/gui/my-apikey). It is on your profile menu under **API key**.
3. In Pelton, go to **Settings, External**, turn on **VirusTotal**, paste the key and press **Save key**.

The key is stored in your operating system keyring, the same place your mail passwords live, and never in the settings database or an export.

## Scanning on demand

Once a key is saved, scanning is available whenever you ask for it:

- **A link**: right-click it in the message and choose **Scan with VirusTotal**.
- **An attachment**: right-click the attachment card and choose **Scan with VirusTotal**.
- **A whole message**: the shield button in the message toolbar scans every link and every attachment in one pass.

Results appear as a small badge beside the link or attachment.

| Badge | Meaning |
| ----- | ------- |
| Check | No engine flagged it. |
| Cross | At least one engine called it malicious or suspicious. |
| Question mark | VirusTotal has never analysed it. This is not a statement that it is safe. |
| Warning triangle | The lookup itself failed, for example because the request quota ran out. |

Hover a badge for the engine counts, and click it to open the full VirusTotal report in your browser.

## Scanning automatically

Automatic scanning is a separate decision from enabling the integration, and both targets are **off by default**. Under **Settings, External** you can turn on:

- **Scan links automatically**, which scans every link in a message as you open it.
- **Scan attachments automatically**, which does the same for attachments.

Leave these off if you would rather decide message by message which links are worth disclosing. With them on, every message you open discloses its links or attachment hashes to VirusTotal, including messages you opened only to delete.

## Request quotas

A free VirusTotal key allows roughly four requests per minute and a few hundred per day. Pelton works within that:

- verdicts are cached locally for seven days, so re-opening a message costs nothing,
- one scan checks at most 25 targets, so a newsletter full of tracking links cannot drain a day's quota in one open,
- scans run one at a time.

If you do hit the limit, affected targets show the warning badge and read "rate limit reached". Waiting and scanning again is all that is needed.

## Turning it off

Switching **VirusTotal** off in **Settings, External** clears the cached verdicts as well, so the local record of which links and files were checked does not outlive the feature. Both automatic scanning toggles are reset at the same time, so turning the integration back on later never silently resumes scanning on its own. Clearing the API key has the same effect on the cache.
