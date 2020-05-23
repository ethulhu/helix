// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

import { documentFragment, elemGenerator } from './elems.js';

const _audio  = elemGenerator( 'audio' );
const _source = elemGenerator( 'source' );
const _video  = elemGenerator( 'video' );

// <helix-media-player>
// 	<source src='…'>
// </helix-media-player>

export class HelixMediaPlayer extends HTMLElement {

	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( documentFragment(
			_audio( { id: 'player' } ),
		) );

		// TODO: maybe put observe() in connectedCallback() and disconnect() in disconnectedCallback()?
		const observer = new MutationObserver( changes => {
			// just fix up all sources with mimetypes.
			Promise.all( this.sources.map( async s => {
				if ( ! s.src ) {
					return;  // wtf?
				}
				if ( s.type ) {
					return;
				}

				const rsp = await fetch( s.src, { method: 'HEAD' } );
				s.type = rsp.headers.get( 'Content-Type' );
			} ) ).then( () => this._render() );
		} );
		observer.observe( this, { childList: true } );
	}

	_render() {
		const element = this.videoTracks.length ? _video : _audio;
		const player = element( {
				id: 'player',
			style: 'width: 100%;',
				durationchange: () => this._sendEvent( 'durationchange' ),
				timeupdate: () => this._sendEvent( 'timeupdate' ),
				ended: () => this._sendEvent( 'ended' ),
				error: e => this._sendEvent( 'error' ),
			},
			Array.from( this.children ).map( el => el.cloneNode( true ) ),
		);

		const wasPaused = this.paused;
		this.shadowRoot.innerHTML = '';
		this.shadowRoot.appendChild( player );
		if ( ! wasPaused ) {
			this.play();
		}
	}

	_sendEvent( name, payload ) {
		const e = new CustomEvent( name, { detail: payload } );
		this.dispatchEvent( e );
	}

	get sources() { return Array.from( this.getElementsByTagName( 'source' ) ); }
	set sources( srcs ) {
		const df = documentFragment( Array.from( srcs ) );
		this.innerHTML = '';
		this.appendChild( df );
	}

	get audioTracks() {
		return this.sources.filter( s => s.type.startsWith( 'audio/' ) );
	}
	get videoTracks() {
		return this.sources.filter( s => s.type.startsWith( 'video/' ) );
	}
	get textTracks() {
		return this.getElementsByTagName( 'track' );
	}
	set src( value ) {
		this.innerHTML = '';
		this.appendChild( _source( { src: value } ) );
	}


	get _playingElement() {
		return this.shadowRoot.getElementById( 'player' );
	}

	play() {
		this._playingElement.play();
	}
	pause() {
		this._playingElement.pause();
	}

	// TODO: maybe replace with an <audio> and <video> for testing?
	canPlayType( t ) {
		return this._playingElement.canPlayType( t );
	}
	get paused() {
		return this._playingElement.paused;
	}
	get duration() {
		return this._playingElement.duration;
	}
	get currentTime() {
		return this._playingElement.currentTime;
	}
	set currentTime( v ) {
		return this._playingElement.currentTime = v;
	}
}

customElements.define( 'helix-media-player', HelixMediaPlayer );
