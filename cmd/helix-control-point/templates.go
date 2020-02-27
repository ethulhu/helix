package main

import "html/template"

var baseTmpl = template.Must(template.New("base.html").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
	<title>Helix - {{ block "title" . }}{{ end }}</title>
	<style>
		:root {
			--background-color:  hsl(0, 0%, 100%);
			--callout-color:     hsl(0, 0%, 90%);

			--foreground-color:  hsl(0, 0%, 0%);
			--link-color:        hsl(0, 0%, 50%);
			--link-hover-color:  hsl(205, 69%, 50%);
		}
		@media (prefers-color-scheme: dark) {
			:root {
				--background-color:  hsl(0, 0%, 12%);
				--callout-color:     hsl(0, 0%, 24%);

				--foreground-color:  hsl(0, 0%, 82%);
				--link-color:        hsl(0, 0%, 65%);
				--link-hover-color:  hsl(259, 49%, 65%);
			}
		}

		body {
			background-color:  var(--background-color);
			color:             var(--foreground-color);
		}
		a, a:visited {
			color:            var(--link-color);
		}
		a:hover {
			color:            var(--link-hover-color);
		}
		input[type=text] {
			background-color:  var(--callout-color);
			color:             var(--foreground-color);

			border-radius: 5px;
			border:        solid 3px var(--callout-color);
		}
		button, select {
			background-color:  var(--link-hover-color);
			color:             var(--background-color);

			border-radius: 5px;
			border:        solid 3px var(--link-hover-color);
		}
	</style>
</head>
<body>
	{{ block "nav" . }}{{ end }}
	<h1>{{ block "title" . }}{{ end }}</h1>
	{{ block "main" . }}{{ end }}
</body>
</html>`))

var indexTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}Helix Control Point{{ end }}
{{ define "main" }}
	<section id='queues'>
		<a href='/queue'>queue</a>
	</section>
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
{{ end }}`))

var directoriesTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}directories{{ end }}
{{ define "nav" }}
	<nav>
		<ul>
			<li><a href='/'>home</a></li>
			<li><a href='/queue'>queue</a></li>
		</ul>
	</nav>
{{ end }}
{{ define "main" }}
	<ul>
	{{ range $index, $device := . }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
{{ end }}`))

var browseTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}{{ .Directory.Name }}{{ end }}
{{ define "nav" }}
	<nav>
		<ul>
			<li><a href='/'>home</a></li>
			<li><a href='/queue'>queue</a></li>
		</ul>
	</nav>
{{ end }}
{{ define "main" }}
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
{{ end }}`))

var transportsTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}transports{{ end }}
{{ define "nav" }}
	<nav>
		<ul>
			<li><a href='/'>home</a></li>
			<li><a href='/queue'>queue</a></li>
		</ul>
	</nav>
{{ end }}
{{ define "main" }}
	<ul>
	{{ range $index, $device := . }}
		<li><a href='/renderer/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
{{ end }}`))

var transportTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}{{ .Transport.Name }}{{ end }}
{{ define "nav" }}
	<nav>
		<ul>
			<li><a href='/'>home</a></li>
			<li><a href='/queue'>queue</a></li>
		</ul>
	</nav>
{{ end }}
{{ define "main" }}
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
{{ end }}`))
