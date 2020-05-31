package upnp

import (
	"mime"
	"path"

	"github.com/ethulhu/helix/upnp/ssdp"
)

type (
	Icon struct {
		MIMEType             string
		Width, Height, Depth int
		URL                  string
	}
)

func iconFromSSDPIcon(i ssdp.Icon) Icon {
	return Icon{
		MIMEType: i.MIMEType,
		Width:    i.Width,
		Height:   i.Height,
		Depth:    i.Depth,
		URL:      i.URL,
	}
}

func (i Icon) ssdpIcon() ssdp.Icon {
	mimetype := i.MIMEType
	if mimetype == "" {
		mimetype = mime.TypeByExtension(path.Ext(i.URL))
	}
	return ssdp.Icon{
		MIMEType: mimetype,
		Width:    i.Width,
		Height:   i.Height,
		Depth:    i.Depth,
		URL:      i.URL,
	}
}
