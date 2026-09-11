package mailview

import (
	"strings"
	"testing"
)

const videoMail = `<p>Clip:</p><video controls width="480" src="https://cdn.example.com/clip.mp4"></video>`

// A message that embeds a video is remote content like any image: nothing may
// reach the sender's server until the reader has asked for it, so the element
// goes rather than being left as a player pointing at a stripped source.
func TestRemoteVideoIsRemovedUntilRemoteContentIsAllowed(t *testing.T) {
	got := Sanitize(videoMail, false, false)

	if strings.Contains(got, "<video") {
		t.Errorf("the video survived with remote content off: %s", got)
	}
	if strings.Contains(got, "cdn.example.com") {
		t.Errorf("the remote source survived with remote content off: %s", got)
	}
	if !strings.Contains(got, "Clip:") {
		t.Errorf("the rest of the message went with it: %s", got)
	}
}

func TestRemoteVideoSurvivesOnceRemoteContentIsAllowed(t *testing.T) {
	got := Sanitize(videoMail, true, false)

	if !strings.Contains(got, "<video") {
		t.Fatalf("the video was removed with remote content on: %s", got)
	}
	if !strings.Contains(got, "https://cdn.example.com/clip.mp4") {
		t.Errorf("the source was removed with remote content on: %s", got)
	}
	if !strings.Contains(got, "controls") {
		t.Errorf("controls were stripped, leaving a player nobody can start: %s", got)
	}
}

// The url can sit on an inner source element instead, which is how mail that
// offers the same clip in several formats writes it.
func TestRemoteSourceElementIsRemovedWithItsPlayer(t *testing.T) {
	const html = `<video controls><source src="https://cdn.example.com/clip.webm" type="video/webm"><source src="https://cdn.example.com/clip.mp4" type="video/mp4"></video>`

	blocked := Sanitize(html, false, false)
	if strings.Contains(blocked, "<video") || strings.Contains(blocked, "cdn.example.com") {
		t.Errorf("a player whose sources are remote survived: %s", blocked)
	}

	allowed := Sanitize(html, true, false)
	if !strings.Contains(allowed, "clip.webm") || !strings.Contains(allowed, `type="video/mp4"`) {
		t.Errorf("the sources did not survive with remote content on: %s", allowed)
	}
}

// A cid: source is an inline part Pelton resolves itself, so it is not remote
// and has nothing to wait for.
func TestInlineMediaIsKeptEvenWithRemoteContentOff(t *testing.T) {
	const html = `<audio controls src="cid:clip@example"></audio>`

	got := Sanitize(html, false, false)
	if !strings.Contains(got, "<audio") || !strings.Contains(got, "cid:clip@example") {
		t.Errorf("an inline audio part was removed: %s", got)
	}
}

// Nothing in a message may start playing by itself: it is the reader's choice
// when a clip runs, and a message that plays on open announces that it was.
func TestAutoplayIsStripped(t *testing.T) {
	const html = `<video controls autoplay src="https://cdn.example.com/clip.mp4"></video>`

	got := Sanitize(html, true, false)
	if strings.Contains(strings.ToLower(got), "autoplay") {
		t.Errorf("autoplay survived: %s", got)
	}
}

// poster is a second url on the element, and bluemonday only scheme-checks
// src, so it would arrive unchecked. It is left out of the allowlist for that
// reason and this is what says so.
func TestPosterIsStripped(t *testing.T) {
	const html = `<video controls poster="https://cdn.example.com/thumb.jpg" src="https://cdn.example.com/clip.mp4"></video>`

	got := Sanitize(html, true, false)
	if strings.Contains(got, "poster") || strings.Contains(got, "thumb.jpg") {
		t.Errorf("poster survived: %s", got)
	}
}

// The element being allowed must not become a way in for a handler.
func TestMediaEventHandlersAreStripped(t *testing.T) {
	const html = `<video controls onerror="alert(1)" onplay="alert(2)" src="https://cdn.example.com/clip.mp4"></video>`

	got := strings.ToLower(Sanitize(html, true, false))
	for _, attr := range []string{"onerror", "onplay", "alert"} {
		if strings.Contains(got, attr) {
			t.Errorf("%s survived: %s", attr, got)
		}
	}
}

// The banner that offers to load remote content is driven by this, so a
// message whose only remote thing is a video still gets the choice.
func TestRemoteVideoCountsAsRemoteContent(t *testing.T) {
	if !HasRemoteContent(videoMail) {
		t.Error("a message whose only remote content is a video reported none")
	}
	hosts := RemoteHosts(videoMail)
	if len(hosts) != 1 || hosts[0] != "cdn.example.com" {
		t.Errorf("RemoteHosts = %v, want the video's host", hosts)
	}
}
