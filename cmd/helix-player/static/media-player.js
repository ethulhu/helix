import { documentFragment, elemGenerator } from './elems.js';

const _audio = elemGenerator( 'audio' );
const _style = elemGenerator( 'style' );
const _video = elemGenerator( 'video' );

export class HelixMediaPlayer extends HTMLElement {
	// First send to <video>.
	// If that fails, send to <audio>.
	// If that fails, report error.

	static get observedAttributes() {
		// TODO: add autoplay and controls?
		return [ 'src' ];
	}

	attributeChangedCallback( attr, oldValue, newValue ) {
		switch ( attr ) {
			case 'src':
				this._playingElement.pause();
				this._video.src = newValue;
				this._playingElement = this._video;
		}
	}

	get src() {
		return this._playingElement.src;
	}
	set src( v ) {
		this.setAttribute( 'src', v );
	}

	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( documentFragment(
			_style( `
				audio, video {
					display: none;
					width: 100%;
				}
			` ),
			_audio( {
				id: 'audio',
				durationchange: () => {
					this._audio.style.display = 'block';
					this._video.style.display = 'none';
					this._sendEvent( 'durationchange' );
				},
				timeupdate: () => this._sendEvent( 'timeupdate' ),
				ended: () => this._sendEvent( 'ended' ),
				error: e => this._sendEvent( 'error' ),
			} ),
			_video( {
				id: 'video',
				durationchange: () => {
					this._audio.style.display = 'none';
					this._video.style.display = 'block';
					this._sendEvent( 'durationchange' );
				},
				timeupdate: () => this._sendEvent( 'timeupdate' ),
				ended: () => this._sendEvent( 'ended' ),
				error: e => {
					const src = e.target.src;
					// e.target.src = '';
					this._audio.src = src;
					this._playingElement = this._audio;
				},
			} ),
		) );

		this._playingElement = this._video;
	}

	play() {
		this._playingElement.play();
	}
	pause() {
		this._playingElement.pause();
	}
	canPlayType( t ) {
		return this._video.canPlayType( t ) || this._audio.canPlayType( t );
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

	_sendEvent( name, payload ) {
		const e = new CustomEvent( name, { detail: payload } );
		this.dispatchEvent( e );
	}

	get _audio() { return this.shadowRoot.getElementById( 'audio' ); }
	get _video() { return this.shadowRoot.getElementById( 'video' ); }
}

customElements.define( 'helix-media-player', HelixMediaPlayer );
