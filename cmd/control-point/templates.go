package main

import "html/template"

var browseTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
</head>
<body>
	{{ $udn := .UDN }}
	<ul>
	{{ range $index, $container := .DIDL.Containers }}
		<li><a href='/browse/{{ $udn }}/{{ $container.ID }}'>{{ $container.Title }}</a></li>
	{{ end }}
	</ul>
	<ul>
	{{ range $index, $item := .DIDL.Items }}
		<li><a href='/browse/{{ $udn }}/{{ $item.ID }}'>{{ $item.Title }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`))

var directoriesTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
</head>
<body>
	<ul>
	{{ range $index, $device := . }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`))
