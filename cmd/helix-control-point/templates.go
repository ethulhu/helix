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

			font-size:    14pt;
			font-family:  sans-serif;

			min-width:    12em;
			padding:      0.5em;
		}
		@media screen and ( min-width: 40em ) {
			body {
				margin:     0 auto;
				max-width:  30em;
			}
		}


		p, li {
			line-height:  1.4;
		}
		a, a:visited {
			color:            var(--link-color);
		}
		a:hover {
			color:            var(--link-hover-color);
		}
		button, select, details {
			background-color:  var(--callout-color);
			color:             var(--foreground-color);
			font-size:    12pt;
			font-family:  sans-serif;

			border-radius:  5px;
			border:         solid 3px var(--callout-color);
			filter:         drop-shadow(1px 1px 1px black);
			margin-top:     3px;
			margin-bottom:  3px;
		}
		button:active {
			filter:     drop-shadow(1px 1px 0.5px black);
			transform:  translateY(1px);
		}

		table {
			width: 100%;
		}
		td:not(.title) {
			float: right;
		}

		#controls details {
			max-width: 100%;
		}

	</style>
</head>
<body>
	<div class='controls'>
		<form method='post' action='/queue'>
			<button name='state' value='play'>play</button>
			<button name='state' value='pause'>pause</button>
			<button name='state' value='stop'>stop</button>
		</form>

		{{- $currentUDN := .Queue.CurrentUDN }}
		<form method='post' action='/queue' id='transport-form'>
			<noscript>
				<input type='submit' value='set output'>
			</noscript>
			<select name='transport'>
				<option value='none' {{ if eq $currentUDN "none" }}selected{{ end }}r>no transport</option>
				{{- if .Queue.Transports }}
					<option disabled>────────────</option>
				{{- end }}
				{{- range $index, $device := .Queue.Transports }}
				<option value='{{ $device.UDN }}' {{ if eq $currentUDN $device.UDN }}selected{{ end }}>{{ $device.Name }}</option>
				{{- end }}
			</select>
		</form>
		<script type='module'>
			document.querySelector( 'select[name=transport]' )
			        .addEventListener( 'change', e => { e.target.parentElement.submit(); } );
		</script>

		<details>
			<summary>playlist</summary>
			<form method='post' action='/queue'>
				<input type='hidden' name='action' value='remove'>
				<button name='position' value='all'>clear playlist</button>
				<ul id='queue'>
				{{- range $index, $item := .Queue.Items }}
					<li>{{ $item.Title }}</li>
				{{- end }}
				</ul>
			</form>
		</details>
	</div>
	<section>
		{{ block "nav" . }}{{ end }}
		<h1>{{ block "title" . }}{{ end }}</h1>
		{{ block "main" . }}{{ end }}
	</section>
</body>
</html>`))

var indexTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}Helix Control Point{{ end }}
{{ define "main" }}
	<section id='directories'>
		<h2>Directories</h2>
		{{ range $index, $device := .Directories }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
		{{ end }}
	</section>
	<section id='renderers'>
		<h2>Renderers</h2>
		{{ range $index, $device := .Queue.Transports }}
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
	{{ range $index, $device := .Directories }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
{{ end }}`))

var browseTmpl = template.Must(template.Must(baseTmpl.Clone()).Parse(`
{{ define "title" }}{{ .Directory.Name }} — {{ .Container.Title }}{{ end }}
{{ define "nav" }}
	<nav>
		<ul>
			<li><a href='/'>home</a></li>
			<li><a href='/queue'>queue</a></li>
			{{ if ne .Container.ParentID "-1" }}<li><a href='/browse/{{ .Directory.UDN }}/{{ .Container.ParentID }}'>back</a></li>{{ end }}
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
	{{ $containerID := .Container.ID }}
	<form method='post' action='/queue'>
		<input type='hidden' name='action'    value='add'>
		<input type='hidden' name='position'  value='last'>
		<input type='hidden' name='directory' value='{{ $udn }}'>
		<button name='object' value='{{ $containerID }}'>+all</button>
		<table>
		{{ range $index, $item := .DIDL.Items }}
			<tr>
				<td class='title'><div>{{ $item.Title }}</div></td>
				<td><button name='object' value='{{ $item.ID }}'>+</button></td>
			</tr>
		{{ end }}
		</table>
	</form>
	{{ end }}
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
	{{ range $index, $device := .Queue.Transports }}
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
		<form method='post' action='/renderer/{{ .Transport.UDN }}'>
			<button name='action' value='play'>play</button>
			<button name='action' value='pause'>pause</button>
			<button name='action' value='stop'>stop</button>
		</form>
	</ul>
{{ end }}`))
