package main

import "html/template"

var queueTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
	<title>Helix - Queue</title>
</head>
<body>
	<h1>Play Queue</h1>
	<label>
		Transport:
		<select id='transport' value='{{ .CurrentUDN }}'>
		{{ range $index, $device := .Transports }}
			<option value='{{ $device.UDN }}'>{{ $device.Name }}</option>
		{{ end }}
		</select>
	</label>
	<button id='play'>play</button>
	<button id='pause'>pause</button>
	<button id='stop'>stop</button>
	<ul>
	{{ range $index, $item := .Items }}
		<li>{{ $item.Title }}</li>
	{{ end }}
	</ul>
	<script type='module'>
[ 'play', 'pause', 'stop' ].forEach( action => {
	const button = document.getElementById( action );
	button.addEventListener( 'click', e => {
		const fd = new FormData();
		fd.append( 'action', action );
		fetch( '/queue', { method: 'post', body: fd } ).then( rsp => rsp.text() )
		                                               .then( console.log )
		                                               .catch( console.error );
	} );

document.getElementById( 'transport' ).addEventListener( 'change', e => {
	const fd = new FormData();
	fd.append( 'transport', e.target.value );
	fetch( '/queue', { method: 'post', body: fd } ).then( rsp => rsp.text() )
	                                               .then( console.log );
} );
} );
	</script>
</body>
</html>`))
