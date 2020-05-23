// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"mime"
	"path/filepath"
	"strings"
)

type (
	// Class is a UPnP AV object class.
	// Classes are defined in ContentDirectory v1, Appendix C.
	Class string
)

var (
	AudioItem = Class("object.item.audioItem")
	ImageItem = Class("object.item.imageItem")
	VideoItem = Class("object.item.videoItem")

	AudioBook      = Class("object.item.audioItem.audioBook")
	AudioBroadcast = Class("object.item.audioItem.audioBroadcast")
	MusicTrack     = Class("object.item.audioItem.musicTrack")
	Photo          = Class("object.item.imageItem.photo")
	Movie          = Class("object.item.videoItem.movie")
	VideoBroadcast = Class("object.item.videoItem.videoBroadcast")
	MusicVideo     = Class("object.item.videoItem.musicVideoClip")
	PlaylistItem   = Class("object.item.playlistItem")
	TextItem       = Class("object.item.textItem")

	Artist        = Class("object.container.person.musicArtist")
	Playlist      = Class("object.container.playlistContainer")
	MusicAlbum    = Class("object.container.album.musicAlbum")
	PhotoAlbum    = Class("object.container.album.photoAlbum")
	MusicGenre    = Class("object.container.genre.musicGenre")
	MovieGenre    = Class("object.container.genre.movieGenre")
	StorageSystem = Class("object.container.storageSystem")
	StorageVolume = Class("object.container.storageVolume")
	StorageFolder = Class("object.container.storageFolder")
)

func ClassForURI(uri string) (Class, error) {
	mimeType := mime.TypeByExtension(filepath.Ext(uri))
	if mimeType == "" {
		return Class(""), ErrUnknownMIMEType
	}
	parts := strings.Split(mimeType, "/")
	switch parts[0] {
	case "audio":
		return AudioItem, nil
	case "image":
		return ImageItem, nil
	case "video":
		return VideoItem, nil
	default:
		return TextItem, nil
	}
}
