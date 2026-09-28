package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Keep each guide focused on a distinct workflow. The main product page remains
// the complete reference and the canonical page for the application itself.
type kspGuide struct {
	Slug        string
	Label       string
	Title       string
	Description string
	Heading     string
	Intro       string
	Summary     string
}

func (g kspGuide) Path() string {
	return "/keyboardsoundplayer/" + g.Slug
}

func (g kspGuide) URL() string {
	return "https://jgltechnologies.com" + g.Path()
}

func (g kspGuide) TemplateName() string {
	return "ksp-" + g.Slug
}

var kspGuides = []kspGuide{
	{
		Slug:        "keyboard-soundboard",
		Label:       "Keyboard soundboard",
		Title:       "Bind Sounds to Keyboard Keys on Windows | KeyboardSoundPlayer",
		Description: "Turn your keyboard into a Windows soundboard. Bind MP3 and WAV files to keys, control playback, and choose key names with KeyboardSoundPlayer.",
		Heading:     "Turn your keyboard into a soundboard",
		Intro:       "Keep a sound effect, music clip, or spoken line one key press away. KeyboardSoundPlayer lets you bind MP3 and WAV files to keyboard keys on Windows and add playback controls to the same keyboard.",
		Summary:     "Bind MP3 and WAV files to keys, choose key names, and control sound playback.",
	},
	{
		Slug:        "audio-capture",
		Label:       "Audio capture",
		Title:       "Clip Recent Audio to a Soundboard Key | KeyboardSoundPlayer",
		Description: "Capture recent audio on Windows and bind the clip to a keyboard key. Learn capture queues, clip duration, and Voicemeeter setup in KeyboardSoundPlayer.",
		Heading:     "Clip recent audio. Replay it with a key.",
		Intro:       "Catch a moment from a game, a call, or other computer audio and add it to your soundboard. With Audio Capture enabled, KeyboardSoundPlayer clips recent audio when you press your capture key and binds it to the next key in your queue.",
		Summary:     "Capture recent audio, set the clip length, and queue keys for your next clips.",
	},
	{
		Slug:        "play-sounds-through-mic",
		Label:       "Play sounds through your mic",
		Title:       "Play Soundboard Audio Through Your Mic | KeyboardSoundPlayer",
		Description: "Play soundboard clips through a virtual microphone in Discord, games, and calls. Set up KeyboardSoundPlayer with Voicemeeter Banana on Windows.",
		Heading:     "Play soundboard audio through your microphone",
		Intro:       "Let people in a voice chat hear your soundboard as well as your voice. KeyboardSoundPlayer's Output To Mic setting works with Voicemeeter Banana to send sounds through a virtual microphone on Windows.",
		Summary:     "Set up Voicemeeter Banana for soundboard audio in Discord, games, and calls.",
	},
	{
		Slug:        "text-to-speech",
		Label:       "Text-to-speech soundboard",
		Title:       "Text-to-Speech Soundboard for Windows | KeyboardSoundPlayer",
		Description: "Turn text into soundboard audio on Windows. Bind spoken phrases to keyboard keys and adjust voice settings with KeyboardSoundPlayer's text-to-speech.",
		Heading:     "Turn typed phrases into soundboard audio",
		Intro:       "Create a spoken soundboard clip by typing a phrase. KeyboardSoundPlayer generates audio from your text and binds it to a key, so a greeting, stream cue, or recurring line is ready to replay.",
		Summary:     "Generate spoken clips from text and adjust the voice and speaking rate.",
	},
}

func registerKSPGuideRoutes(router *gin.Engine, pageCache gin.HandlerFunc) {
	for _, guide := range kspGuides {
		router.GET(guide.Path(), pageCache, func(c *gin.Context) {
			c.HTML(http.StatusOK, guide.TemplateName(), gin.H{
				"Guide":  guide,
				"Guides": kspGuides,
			})
		})
	}
}
