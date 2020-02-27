package main

import "html/template"

var indexTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
	<title>Helix Control Point</title>
</head>
<body>
	<h1>Helix Control Point</h1>
	<section id='directories'>
		<h2>Directories</h2>
		{{ range $index, $device := .Directories }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
		{{ end }}
	</section>
	<section id='renderers'>
		<h2>Renderers</h2>
		{{ range $index, $device := .Transports }}
		<li><a href='/renderer/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
		{{ end }}
	</section>
	<script type='module'>
	</script>
</body>
</html>`))

var directoriesTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
	<title>Helix - directories</title>
</head>
<body>
	<h1>directories</h1>
	<ul>
	{{ range $index, $device := . }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`))

var browseTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
	<title>Helix - {{ .Directory.Name }}</title>
</head>
<body>
	<h1>{{ .Directory.Name }}</h1>
	{{ $udn := .Directory.UDN }}

	{{ if .DIDL.Containers }}
	<ul>
	{{ range $index, $container := .DIDL.Containers }}
		<li><a href='/browse/{{ $udn }}/{{ $container.ID }}'>{{ $container.Title }}</a></li>
	{{ end }}
	</ul>
	{{ end }}

	{{ if .DIDL.Items }}
	<ul>
	{{ range $index, $item := .DIDL.Items }}
		<li><button data-objectid='{{ $item.ID }}'>{{ $item.Title }}</button></li>
	{{ end }}
	</ul>
	{{ end }}
	<script type='module'>
const directoryUDN = '{{ .Directory.UDN }}';
document.querySelectorAll( 'button' ).forEach( button => {
	button.addEventListener( 'click', e => {
		const fd = new FormData();
		fd.append( 'action', 'add' );
		fd.append( 'position', 'last' );
		fd.append( 'directory', directoryUDN );
		fd.append( 'object', e.target.dataset.objectid );
		fetch( '/queue', { method: 'post', body: fd } ).then( rsp => {
			if ( ! rsp.ok ) {
				throw rsp.text();
			}
		} ).catch( console.error );
	} );
} );

	</script>
</body>
</html>`))

var transportsTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
</head>
<body>
	<h1>renderers</h1>
	<ul>
	{{ range $index, $device := . }}
		<li><a href='/renderer/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`))

var transportTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
</head>
<body>
	<h1>{{ .Transport.Name }}</h1>
	<p>state: {{ .PlaybackState }}</p>
	{{ if .DIDL }}
	<p>items:
	<ul>
	{{ range $index, $item := .DIDL.Items }}
		<li>{{ $item.Title }}</li>
	{{ end }}
	</ul>
	</p>
	{{ end }}
	<ul id='actions'>
	</ul>
	<script type='module'>
const actions = document.getElementById( 'actions' );
[ 'play', 'pause', 'stop' ].forEach( action => {
	const li = document.createElement( 'li' );
	const button = document.createElement( 'button' );
	const fd = new FormData();
	fd.append('action', action);
	button.addEventListener( 'click', e => fetch( window.location, { method: 'POST', body: fd } ).then( console.log ).catch( console.error ) );
	button.textContent = action;
	li.appendChild( button );
	actions.appendChild( li );
} );
	</script>
</body>
</html>`))
